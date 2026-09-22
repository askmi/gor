package gor

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestEngine() Engine {
	e := NewEngine()
	e.WithLogger(slog.New(slog.NewTextHandler(io.Discard, nil)))
	e.Route(NewRouter("/"))
	return e
}

func shutdownTestEngine(t *testing.T, e Engine) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := e.StopGracefully(ctx); err != nil {
		t.Fatalf("StopGracefully() error = %v", err)
	}

	select {
	case <-e.Done():
	case <-ctx.Done():
		t.Fatalf("Done() was not closed: %v", ctx.Err())
	}
}

func TestEngineRequiresRouter(t *testing.T) {
	e := NewEngine()
	if err := e.Listen(":0"); !errors.Is(err, ErrEngineRouterIsMissing) {
		t.Fatalf("Listen() error = %v, want %v", err, ErrEngineRouterIsMissing)
	}
}

func TestEngineListen(t *testing.T) {
	e := newTestEngine()
	result := make(chan error, 1)
	go func() {
		result <- e.Listen(":0")
	}()

	deadline := time.Now().Add(2 * time.Second)
	for {
		concrete := e.(*engine)
		concrete.mu.Lock()
		started := concrete.started
		concrete.mu.Unlock()
		if started {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("Listen() did not start the engine")
		}
		time.Sleep(time.Millisecond)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := e.StopGracefully(ctx); err != nil {
		t.Fatalf("StopGracefully() error = %v", err)
	}

	select {
	case err := <-result:
		if err != nil {
			t.Fatalf("Listen() error = %v", err)
		}
	case <-ctx.Done():
		t.Fatalf("Listen() did not return: %v", ctx.Err())
	}
}

func TestEngineRequiresStartForLifecycleOperations(t *testing.T) {
	e := newTestEngine()
	if err := e.StopGracefully(context.Background()); !errors.Is(err, ErrEngineNotStarted) {
		t.Fatalf("StopGracefully() error = %v, want %v", err, ErrEngineNotStarted)
	}
}

func TestEngineMountsRouterAfterStart(t *testing.T) {
	e := newTestEngine()
	result := make(chan error, 1)
	go func() { result <- e.Listen(":0") }()
	waitForEngineStart(t, e)

	dynamic := NewRouter("/dynamic/")
	dynamic.HandleHTTP("GET /hello", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	e.Route(dynamic)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/dynamic/hello", nil)
	e.(*engine).server.Handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("dynamic route status = %d, want %d", recorder.Code, http.StatusNoContent)
	}

	shutdownTestEngine(t, e)
	if err := <-result; err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
}

func TestEngineDefaultProbes(t *testing.T) {
	e := newTestEngine().EnableProbes()
	result := make(chan error, 1)
	go func() { result <- e.Listen(":0") }()
	waitForEngineStart(t, e)

	for _, path := range []string{"/health", "/healthz", "/livez", "/startupz", "/readyz"} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, path, nil)
		e.(*engine).server.Handler.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Errorf("GET %s status = %d, want %d", path, recorder.Code, http.StatusOK)
		}
	}

	shutdownTestEngine(t, e)
	if err := <-result; err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
}

func TestEngineStartReturnsListenError(t *testing.T) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}
	e := newTestEngine()
	address := listener.Addr().String()
	if err := e.Listen(address); err == nil {
		listener.Close()
		t.Fatal("Listen() error = nil, want address-in-use error")
	}

	if err := listener.Close(); err != nil {
		t.Fatalf("listener.Close() error = %v", err)
	}
	result := make(chan error, 1)
	go func() { result <- e.Listen(address) }()
	waitForEngineStart(t, e)
	shutdownTestEngine(t, e)
	if err := <-result; err != nil {
		t.Fatalf("Listen() after releasing port error = %v", err)
	}
}

func TestEngineWithGracefulPeriod(t *testing.T) {
	e := NewEngine().(*engine)
	if e.gracefulTimeout != DefaultGracefulTimeout {
		t.Fatalf("gracefulTimeout = %v, want %v", e.gracefulTimeout, DefaultGracefulTimeout)
	}

	if e.WithGracefulPeriod(5*time.Second) != Engine(e) {
		t.Fatal("WithGracefulPeriod() did not return the engine")
	}
	if e.gracefulTimeout != 5*time.Second {
		t.Fatalf("gracefulTimeout = %v, want %v", e.gracefulTimeout, 5*time.Second)
	}
}

func TestEngineWithGracefulPeriodRejectsNonPositive(t *testing.T) {
	for _, d := range []time.Duration{0, -time.Second} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("WithGracefulPeriod(%v) did not panic", d)
				}
			}()
			NewEngine().WithGracefulPeriod(d)
		}()
	}
}

func TestEngineWithGracefulPeriodRejectsAfterStart(t *testing.T) {
	e := newTestEngine()
	result := make(chan error, 1)
	go func() { result <- e.Listen(":0") }()
	waitForEngineStart(t, e)

	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("WithGracefulPeriod() after start did not panic")
			}
		}()
		e.WithGracefulPeriod(time.Second)
	}()

	shutdownTestEngine(t, e)
	if err := <-result; err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
}

func waitForEngineStart(t *testing.T, e Engine) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		concrete := e.(*engine)
		concrete.mu.Lock()
		started := concrete.started
		concrete.mu.Unlock()
		if started {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("engine did not start")
		}
		time.Sleep(time.Millisecond)
	}
}
