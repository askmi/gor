package gor

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReplayBodyMiddleware(t *testing.T) {
	const want = `{"name":"Ada"}`

	handler := ReplayBodyMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		first, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("first ReadAll() error = %v", err)
		}

		secondBody, err := r.GetBody()
		if err != nil {
			t.Fatalf("GetBody() error = %v", err)
		}
		defer secondBody.Close()
		second, err := io.ReadAll(secondBody)
		if err != nil {
			t.Fatalf("second ReadAll() error = %v", err)
		}

		if string(first) != want || string(second) != want {
			t.Fatalf("body reads = %q and %q, want %q", first, second, want)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(want))
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
}

func TestChainAcceptsStandardHTTPMiddleware(t *testing.T) {
	standard := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Middleware", "adapted")
			next.ServeHTTP(w, r)
		})
	}

	handler := Chain(standard)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
	if got := recorder.Header().Get("X-Middleware"); got != "adapted" {
		t.Fatalf("X-Middleware = %q, want %q", got, "adapted")
	}
}

func TestStatusCodeResponseWriterUnwrap(t *testing.T) {
	underlying := httptest.NewRecorder()
	w := NewResponseWriter(underlying)

	if got := w.Unwrap(); got != underlying {
		t.Fatalf("Unwrap() = %T, want underlying writer", got)
	}
}
