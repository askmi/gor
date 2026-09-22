package gor

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type (
	// HTTPMiddleware wraps an HTTP handler and returns the handler used for the request chain.
	HTTPMiddleware func(http.Handler) http.Handler
	// Authenticator validates a security context and returns the resulting authentication state.
	Authenticator func(SecurityContext) (SecurityContext, error)
	// RouterFunc is a typed endpoint that transforms a decoded request into a response.
	RouterFunc[Req any, Resp any] func(context.Context, Req) (Resp, error)
	// ErrorHandler maps an application error to an HTTP response.
	ErrorHandler func(context.Context, error) HTTPResponse
	// ResponseWriter writes an HTTPResponse to the underlying net/http writer.
	ResponseWriter func(context.Context, HTTPResponse, http.ResponseWriter)
	// RequestHandler populates a request value from an incoming HTTP request.
	RequestHandler func(context.Context, *http.Request, any) error
	// ResponseHandler maps an application value to an HTTP response.
	ResponseHandler func(context.Context, any) (HTTPResponse, error)
	// HTTPDecoder lets a request value decode itself from an incoming HTTP request.
	HTTPDecoder interface {
		// DecodeFromHTTPRequest populates the receiver from req.
		DecodeFromHTTPRequest(*http.Request) error
	}

	// Engine owns the HTTP server lifecycle and its mounted routers.
	Engine interface {
		// Start serving synchronously.
		Listen(string) error
		// Stop gracefully stops the server within ctx's deadline.
		StopGracefully(context.Context) error
		// Done returns a channel closed when serving ends.
		Done() <-chan struct{}

		// NewRouter creates, mounts, and returns a router.
		NewRouter(string) *Router
		// Route adds a router, mounting it immediately when the engine is running.
		// It panics if the router's prefix conflicts with an existing route.
		Route(*Router) Engine
		// EnableProbes registers empty 200 OK handlers for the default probe paths,
		// or for the supplied net/http ServeMux patterns. Call it before Listen.
		EnableProbes(...string) Engine
		EnableSignals(...os.Signal) Engine
		// WithGracefulPeriod sets how long a signal-triggered shutdown may take,
		// overriding DefaultGracefulTimeout. Call it before Listen.
		// It panics if d is not positive or the engine has already started.
		WithGracefulPeriod(time.Duration) Engine
		// WithLogger configures the engine logger before Listen.
		// It panics if logger is nil or the engine has already started.
		WithLogger(*slog.Logger)
		// OnShutdown registers a cleanup func to run once serving has ended.
		// It panics if f is nil. See OnShutdownWithContext for the ordering and
		// timeout rules that apply.
		OnShutdown(func()) Engine
		// OnShutdownWithContext registers a cleanup func to run once serving has
		// ended, after active requests have drained. Hooks run in reverse
		// registration order, like deferred calls, and share the graceful period
		// with request draining: a hook is skipped once ctx is done, so long
		// cleanup must honor cancellation. It panics if f is nil.
		OnShutdownWithContext(func(context.Context)) Engine
	}

	// SecurityContext describes the authentication state and identity of a request.
	SecurityContext interface {
		// IdentityString returns a textual representation of the identity or credentials.
		IdentityString() string
		// IsAuthenticated reports whether the identity has been authenticated.
		IsAuthenticated() bool
		// Identity returns the underlying principal, credentials, or rejection value.
		Identity() any
	}

	// HTTPResponse is the transport-neutral response consumed by a Router's ResponseWriter.
	HTTPResponse interface {
		// Headers returns response headers to write before the status and body.
		Headers() map[string]string
		// StatusCode returns the HTTP status code.
		StatusCode() int
		// Content returns the response body bytes.
		Content() []byte
		// ContentType returns the media type, or an empty string when unspecified.
		ContentType() string
	}

	// ServerOpts contains HTTP server configuration.
	ServerOpts  []func(*http.Server) *http.Server
	RouteOption func(*routeOpts)
)
