package gor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleFuncWithStatusCode(t *testing.T) {
	router := NewRouter("/")
	router.HandleFunc(
		"POST /users",
		func(context.Context, *struct{}) (struct{ ID int }, error) {
			return struct{ ID int }{ID: 1}, nil
		},
		WithStatusCode(http.StatusCreated),
	)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/users", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status code = %d, want %d", recorder.Code, http.StatusCreated)
	}
}

func TestHandleFuncWithStatusCodeForEmptyResponse(t *testing.T) {
	for _, statusCode := range []int{http.StatusOK, http.StatusNotImplemented} {
		t.Run(http.StatusText(statusCode), func(t *testing.T) {
			router := NewRouter("/")
			router.HandleFunc(
				"GET /todo",
				func(context.Context, *struct{}) (*struct{}, error) {
					return nil, nil
				},
				WithStatusCode(statusCode),
			)

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/todo", nil)
			router.ServeHTTP(recorder, request)

			if recorder.Code != statusCode {
				t.Fatalf("status code = %d, want %d", recorder.Code, statusCode)
			}
		})
	}
}

func TestHandleFuncUsesOKForNilResponseByDefault(t *testing.T) {
	router := NewRouter("/")
	router.HandleFunc(
		"GET /empty",
		func(context.Context, *struct{}) (*struct{}, error) {
			return nil, nil
		},
	)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/empty", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestHandleFuncUsesOKForResponseBodyByDefault(t *testing.T) {
	router := NewRouter("/")
	router.HandleFunc(
		"GET /value",
		func(context.Context, *struct{}) (string, error) {
			return "value", nil
		},
	)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/value", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestRouterImplementsHTTPHandlerWithPrefix(t *testing.T) {
	router := NewRouter("/api/v1/")
	router.Get("/hello", func(context.Context, *struct{}) (string, error) {
		return "hello", nil
	})

	var handler http.Handler = router
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/hello", nil)
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestRouterNormalizesPathWithoutLeadingSlash(t *testing.T) {
	router := NewRouter("/")
	router.Get("hello", func(context.Context, *struct{}) (string, error) {
		return "hello", nil
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/hello", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestRouterWithSharesRegisteredRoutes(t *testing.T) {
	router := NewRouter("/api/v1/")
	router.With(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Scoped", "true")
			next.ServeHTTP(w, r)
		})
	}).Get("/hello", func(context.Context, *struct{}) (string, error) {
		return "hello", nil
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/hello", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Header().Get("X-Scoped") != "true" {
		t.Fatal("scoped middleware was not applied")
	}
}

func TestWithStatusCodeRejectsInvalidCode(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("WithStatusCode() did not panic")
		}
	}()

	WithStatusCode(99)
}
