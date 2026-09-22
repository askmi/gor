package gor

import "net/http"

type (
	simpleHTTPResponse struct {
		statusCode  int
		content     []byte
		contentType string
	}

	rejectedSecurityContext struct {
		reason string
	}

	unauthenticatedSecurityContext struct {
		raw []byte
	}

	authenticatedSecurityContext struct {
		principalString string
		principal       any
	}

	// StatusCodeResponseWriter wraps an HTTP response writer and records the latest status code.
	StatusCodeResponseWriter struct {
		http.ResponseWriter
		// StatusCode is the most recent code passed to WriteHeader, initially 200.
		StatusCode int
	}
)

var (
	// HTTPResponse200 is an empty HTTP 200 OK response.
	HTTPResponse200 = simpleHTTPResponse{statusCode: 200}
	// HTTPResponse201 is an empty HTTP 201 Created response.
	HTTPResponse201 = simpleHTTPResponse{statusCode: 201}
	// HTTPResponse202 is an empty HTTP 202 Accepted response.
	HTTPResponse202 = simpleHTTPResponse{statusCode: 202}
	// HTTPResponse204 is an empty HTTP 204 No Content response.
	HTTPResponse204 = simpleHTTPResponse{statusCode: 204}
)

func (e simpleHTTPResponse) Headers() map[string]string { return nil }
func (e simpleHTTPResponse) StatusCode() int            { return e.statusCode }
func (e simpleHTTPResponse) Content() []byte            { return e.content }
func (e simpleHTTPResponse) ContentType() string        { return e.contentType }

func (s *unauthenticatedSecurityContext) IsAuthenticated() bool  { return false }
func (s *unauthenticatedSecurityContext) IdentityString() string { return string(s.raw) }
func (s *unauthenticatedSecurityContext) Identity() any          { return s.raw }

func (s *rejectedSecurityContext) IsAuthenticated() bool  { return false }
func (s *rejectedSecurityContext) IdentityString() string { return s.reason }
func (s *rejectedSecurityContext) Identity() any          { return s.reason }

func (s *authenticatedSecurityContext) IsAuthenticated() bool  { return true }
func (s *authenticatedSecurityContext) IdentityString() string { return s.principalString }
func (s *authenticatedSecurityContext) Identity() any          { return s.principal }

// NewResponseWriter wraps w and initializes its recorded status to HTTP 200 OK.
func NewResponseWriter(w http.ResponseWriter) *StatusCodeResponseWriter {
	return &StatusCodeResponseWriter{w, 200}
}

// WriteHeader records code and forwards it to the wrapped response writer.
func (w *StatusCodeResponseWriter) WriteHeader(code int) {
	w.StatusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// Unwrap returns the underlying writer for optional interfaces such as http.Hijacker.
func (w *StatusCodeResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// OkJSON creates an HTTP 200 response containing the supplied JSON bytes.
func OkJSON(b []byte) HTTPResponse {
	return &simpleHTTPResponse{
		statusCode:  http.StatusOK,
		content:     b,
		contentType: "application/json",
	}
}

// NewHTTPResponse creates a response with the supplied status, media type, and body.
func NewHTTPResponse(statusCode int, contentType, content string) HTTPResponse {
	return &simpleHTTPResponse{
		statusCode:  statusCode,
		content:     []byte(content),
		contentType: contentType,
	}
}

// NewJSONResponse creates a JSON response with the supplied status and body.
func NewJSONResponse(statusCode int, content string) HTTPResponse {
	return NewHTTPResponse(statusCode, "application/json", content)
}
