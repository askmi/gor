package gor

import (
	"context"
	"crypto/tls"
	"io"
	"log"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestNewEngineAppliesServerOptions(t *testing.T) {
	opts := NewServerOpts().
		WithReadHeaderTimeout(time.Second).
		WithReadTimeout(2 * time.Second).
		WithWriteTimeout(3 * time.Second).
		WithIdleTimeout(4 * time.Second).
		WithMaxHeaderBytes(1 << 20)
	e := NewEngine(opts...).(*engine)
	server := e.opts.apply(&http.Server{})

	if server.ReadHeaderTimeout != time.Second {
		t.Errorf("ReadHeaderTimeout = %v, want %v", server.ReadHeaderTimeout, time.Second)
	}
	if server.ReadTimeout != 2*time.Second {
		t.Errorf("ReadTimeout = %v, want %v", server.ReadTimeout, 2*time.Second)
	}
	if server.WriteTimeout != 3*time.Second {
		t.Errorf("WriteTimeout = %v, want %v", server.WriteTimeout, 3*time.Second)
	}
	if server.IdleTimeout != 4*time.Second {
		t.Errorf("IdleTimeout = %v, want %v", server.IdleTimeout, 4*time.Second)
	}
	if server.MaxHeaderBytes != 1<<20 {
		t.Errorf("MaxHeaderBytes = %d, want %d", server.MaxHeaderBytes, 1<<20)
	}
}

func TestServerOptsMethods(t *testing.T) {
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS13}
	errorLog := log.New(io.Discard, "server: ", 0)
	http2Config := &http.HTTP2Config{MaxConcurrentStreams: 100}
	protocols := new(http.Protocols)
	protocols.SetHTTP1(true)
	connState := func(_ net.Conn, _ http.ConnState) {}
	baseContext := func(_ net.Listener) context.Context { return context.Background() }
	connContext := func(ctx context.Context, _ net.Conn) context.Context { return ctx }
	tlsNextProto := map[string]func(*http.Server, *tls.Conn, http.Handler){
		"custom": func(*http.Server, *tls.Conn, http.Handler) {},
	}

	base := ServerOpts{}
	opts := base.
		WithReadHeaderTimeout(time.Second).
		WithReadTimeout(2 * time.Second).
		WithWriteTimeout(3 * time.Second).
		WithIdleTimeout(4 * time.Second).
		WithMaxHeaderBytes(1 << 20).
		WithMaxHeaderValueCount(100).
		WithDisableGeneralOptionsHandler(true).
		WithTLSConfig(tlsConfig).
		WithTLSNextProto(tlsNextProto).
		WithConnState(connState).
		WithErrorLog(errorLog).
		WithBaseContext(baseContext).
		WithConnContext(connContext).
		WithHTTP2(http2Config).
		WithProtocols(protocols).
		WithDisableClientPriority(true)

	server := opts.apply(&http.Server{})
	if server.ReadHeaderTimeout != time.Second ||
		server.ReadTimeout != 2*time.Second ||
		server.WriteTimeout != 3*time.Second ||
		server.IdleTimeout != 4*time.Second ||
		server.MaxHeaderBytes != 1<<20 ||
		server.MaxHeaderValueCount != 100 ||
		!server.DisableGeneralOptionsHandler ||
		server.TLSConfig != tlsConfig ||
		server.TLSNextProto == nil ||
		server.ConnState == nil ||
		server.ErrorLog != errorLog ||
		server.BaseContext == nil ||
		server.ConnContext == nil ||
		server.HTTP2 != http2Config ||
		server.Protocols != protocols ||
		!server.DisableClientPriority {
		t.Fatalf("server options were not applied: %+v", server)
	}

	baseServer := base.apply(&http.Server{})
	if baseServer.ReadHeaderTimeout != 0 || len(base) != 0 {
		t.Fatal("configuring a copy mutated the original ServerOpts")
	}
}

func TestNewEngineAppliesLaterOptionsLast(t *testing.T) {
	e := NewEngine(
		NewServerOpts().
			WithReadTimeout(time.Second).
			WithReadTimeout(2 * time.Second).
			WithWriteTimeout(3 * time.Second)...,
	).(*engine)

	server := e.opts.apply(&http.Server{})
	if server.ReadTimeout != 2*time.Second {
		t.Errorf("ReadTimeout = %v, want %v", server.ReadTimeout, 2*time.Second)
	}
	if server.WriteTimeout != 3*time.Second {
		t.Errorf("WriteTimeout = %v, want %v", server.WriteTimeout, 3*time.Second)
	}
}

func TestStandaloneServerOptions(t *testing.T) {
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS13}

	e := NewEngine(
		WithReadHeaderTimeout(time.Second),
		WithReadTimeout(2*time.Second),
		WithWriteTimeout(3*time.Second),
		WithIdleTimeout(4*time.Second),
		WithMaxHeaderBytes(1<<20),
		WithMaxHeaderValueCount(100),
		WithDisableGeneralOptionsHandler(true),
		WithTLSConfig(tlsConfig),
		WithDisableClientPriority(true),
	).(*engine)

	server := e.opts.apply(&http.Server{})
	if server.ReadHeaderTimeout != time.Second ||
		server.ReadTimeout != 2*time.Second ||
		server.WriteTimeout != 3*time.Second ||
		server.IdleTimeout != 4*time.Second ||
		server.MaxHeaderBytes != 1<<20 ||
		server.MaxHeaderValueCount != 100 ||
		!server.DisableGeneralOptionsHandler ||
		server.TLSConfig != tlsConfig ||
		!server.DisableClientPriority {
		t.Fatalf("standalone options were not applied: %+v", server)
	}
}

func TestNewServerOptsHoldsSuppliedOptions(t *testing.T) {
	opts := NewServerOpts(WithReadTimeout(time.Second))
	if len(opts) != 1 {
		t.Fatalf("len(opts) = %d, want 1", len(opts))
	}

	if server := opts.apply(&http.Server{}); server.ReadTimeout != time.Second {
		t.Errorf("ReadTimeout = %v, want %v", server.ReadTimeout, time.Second)
	}
}

func TestWithOptionAppliesArbitrarySetting(t *testing.T) {
	opts := NewServerOpts().WithOption(func(server *http.Server) *http.Server {
		server.Addr = ":9999"
		return server
	})

	if server := opts.apply(&http.Server{}); server.Addr != ":9999" {
		t.Errorf("Addr = %q, want %q", server.Addr, ":9999")
	}
}
