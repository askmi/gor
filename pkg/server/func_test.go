package gor

import (
	"context"
	"errors"
	"testing"
)

func TestUnwrap(t *testing.T) {
	first := errors.New("first")
	second := errors.New("second")
	joined := errors.Join(first, second)

	for _, tt := range []struct {
		err   error
		index int
		want  error
	}{
		{joined, 0, first},
		{joined, 1, second},
		{joined, -1, nil},
		{joined, 2, nil},
		{first, 0, nil},
		{nil, 0, nil},
	} {
		if got := Unwrap(tt.err, tt.index); got != tt.want {
			t.Errorf("Unwrap(%v, %d) = %v, want %v", tt.err, tt.index, got, tt.want)
		}
	}
}

func TestPrincipalFromContext(t *testing.T) {
	type principal struct{ Name string }
	want := principal{Name: "Ada"}
	ctx := WithSecurityContext(context.Background(), Authenticated(want.Name, want))

	got, ok := PrincipalFromContext[principal](ctx)
	if !ok || got != want {
		t.Fatalf("PrincipalFromContext() = %#v, %v, want %#v, true", got, ok, want)
	}

	if _, ok := PrincipalFromContext[string](ctx); ok {
		t.Fatal("PrincipalFromContext[string]() ok = true, want false")
	}

	unauthenticated := WithSecurityContext(context.Background(), Unauthenticated([]byte("token")))
	if _, ok := PrincipalFromContext[[]byte](unauthenticated); ok {
		t.Fatal("PrincipalFromContext() returned an unauthenticated identity")
	}

	if _, ok := GetSecurityFromContext(ctx); !ok {
		t.Fatal("GetSecurityFromContext() ok = false, want true")
	}
}
