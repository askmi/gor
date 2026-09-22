package gor

import (
	"bytes"
	"context"
	"encoding/base64"
	"net/http"
	"strings"
)

type (
	routerKey          struct{}
	securityContextKey struct{}
)

// WithSecurityContext returns a child context containing the request's security state.
func WithSecurityContext(ctx context.Context, s SecurityContext) context.Context {
	return context.WithValue(ctx, securityContextKey{}, s)
}

// PrincipalFromContext returns the authenticated principal stored in ctx when it has type P.
func PrincipalFromContext[P any](ctx context.Context) (P, bool) {
	var i P
	s, ok := SecurityFromContext(ctx)
	if !ok || !s.IsAuthenticated() {
		return i, false
	}
	i, ok = s.Identity().(P)
	if !ok {
		return i, false
	}
	return i, true
}

// SecurityFromContext returns the SecurityContext stored in ctx, if present and valid.
func SecurityFromContext(ctx context.Context) (SecurityContext, bool) {
	value := ctx.Value(securityContextKey{})
	if value == nil {
		return nil, false
	}
	s, ok := value.(SecurityContext)
	if !ok {
		return nil, false
	}
	return s, true
}

// GetSecurityFromContext returns the SecurityContext stored in ctx, if present and valid.
// Deprecated: use SecurityFromContext.
func GetSecurityFromContext(ctx context.Context) (SecurityContext, bool) {
	return SecurityFromContext(ctx)
}

// Authenticated creates an authenticated security context for principal.
// principalString is its stable textual identity for logging or display.
func Authenticated(principalString string, principal any) SecurityContext {
	return &authenticatedSecurityContext{
		principalString: principalString,
		principal:       principal,
	}
}

// Unauthenticated creates a security context containing raw credentials awaiting validation.
func Unauthenticated(raw []byte) SecurityContext {
	return &unauthenticatedSecurityContext{
		raw: raw,
	}
}

// Rejected creates an unauthenticated security context containing a rejection reason.
func Rejected(reason string) SecurityContext {
	return &rejectedSecurityContext{
		reason: reason,
	}
}

// GetStatusCode returns the status recorded by a SimpleResponseWriter.
// The boolean is false when w is not a *SimpleResponseWriter.
func GetStatusCode(w http.ResponseWriter) (int, bool) {
	sw, ok := w.(*StatusCodeResponseWriter)
	if !ok {
		return 0, false
	}
	return sw.StatusCode, true
}

// DecodeBasic decodes a base64-encoded Basic credentials payload into username and password.
// It returns false when the payload is empty, invalid, or lacks a non-empty password.
func DecodeBasic(raw []byte) ([]byte, []byte, bool) {
	if len(raw) == 0 {
		return nil, nil, false
	}

	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(raw)))
	n, err := base64.StdEncoding.Decode(decoded, raw)
	if err != nil {
		return nil, nil, false
	}
	decoded = decoded[:n]

	i := bytes.Index(decoded, []byte{':'})
	if i < 0 || i+1 >= len(decoded) {
		return nil, nil, false
	}

	return decoded[:i], decoded[i+1:], true
}

// Unwrap returns cause i from an error implementing Unwrap() []error.
// It returns nil for a non-multi-error or an out-of-range index.
func Unwrap(err error, i int) error {
	if err == nil {
		return nil
	}

	multi, ok := err.(interface {
		Unwrap() []error
	})
	if !ok {
		return nil
	}
	causes := multi.Unwrap()
	if i < 0 || i >= len(causes) {
		return nil
	}

	return causes[i]
}

// and returns independent options containing o followed by a.
func Merge[E ~[]T, T any](e ...E) E {
	cap := 0
	for i, _ := range e {
		cap += len(e[i])
	}
	opts := make([]T, 0, cap)
	for i, _ := range e {
		opts = append(opts, e[i]...)
	}
	return opts
}

func WithElement[E ~[]T, T any](a E, e T) E {
	result := make(E, len(a), len(a)+1)
	copy(result, a)
	return append(result, e)
}

//************************************************
// Private functions
//************************************************

func statusCodeOrDefault(statusCode, defaultStatusCode int) int {
	if statusCode == 0 {
		return defaultStatusCode
	}
	return statusCode
}

func mountRoutes(mux *http.ServeMux, routes []*Router) {
	for _, route := range routes {
		mountRouter(mux, route)
	}
}

func mountRouter(mux *http.ServeMux, r *Router) {
	for _, re := range r.state.entries {
		mux.Handle(re.pattern, re.h)
	}
}

func buildPattern(key, p string) string {
	if len(key) > 0 && key[len(key)-1] == '/' {
		key = key[:len(key)-1]
	}
	i := strings.IndexByte(p, ' ')
	if i < 0 {
		if len(p) > 0 && p[0] == '/' {
			return key + p
		}
		return key + "/" + p
	}

	i++
	if i == len(p) {
		if key == "" {
			return p + "/"
		}
		return p + key
	}
	if p[i] == '/' {
		return p[:i] + key + p[i:]
	}
	return p[:i] + key + "/" + p[i:]
}
