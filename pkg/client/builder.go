package client

import (
	"net/http"
	"time"

	gor "gor/pkg/server"
)

// https://go.dev/src/net/http/client.go
// https://go.dev/src/net/http/transport.go

// Client settings live on two standard types. *http.Client holds per-request
// policy, and is what ClientOpts configures. *http.Transport holds the connection
// pool and everything about establishing a connection: see TransportOpts in
// transport.go.
//
//	*http.Client        option              default
//	├── Timeout         WithTimeout         0 (no limit)
//	├── CheckRedirect   WithCheckRedirect   follow up to 10 hops
//	├── Jar             WithCookieJar       nil (no cookies)
//	└── Transport       WithTransport       http.DefaultTransport
//	                    WithTransportOpts
//
// Two ways to configure:
//
//	client := NewClient(
//		WithTimeout(30*time.Second),
//		WithTransportOpts(WithMaxIdleConnsPerHost(100)),
//	)
//
//	opts := NewClientOpts().
//		WithTimeout(30 * time.Second).
//		WithTransportOpts(NewTransportOpts().
//			WithMaxIdleConnsPerHost(100)...)
//
//	client := NewClient(opts...)
//
// The pool belongs to the transport, so connections are reused only between
// requests sharing one transport instance: build the client once and hand it
// around, never per request. Pass a transport built elsewhere with WithTransport,
// which is how instrumentation composes:
//
//	client := NewClient(
//		WithTransport(otelhttp.NewTransport(transport)),
//		WithTimeout(30*time.Second),
//	)

type (
	// ClientOpts contains HTTP client configuration.
	ClientOpts []func(*http.Client) *http.Client
)

// NewClient creates an HTTP client with options applied in order.
func NewClient(options ...func(*http.Client) *http.Client) *http.Client {
	c := &http.Client{}
	for _, option := range options {
		c = option(c)
	}
	return c
}

// NewClientOpts creates a configuration holding the supplied options.
func NewClientOpts(options ...func(*http.Client) *http.Client) ClientOpts {
	return append(make(ClientOpts, 0, len(options)), options...)
}

// WithOption appends an arbitrary option, for settings without a dedicated method.
func (o ClientOpts) WithOption(option func(*http.Client) *http.Client) ClientOpts {
	return gor.WithElement(o, option)
}

// WithTransport sets the client's HTTP transport.
func (o ClientOpts) WithTransport(transport http.RoundTripper) ClientOpts {
	return gor.WithElement(o, WithTransport(transport))
}

// WithTransportOpts sets the client's HTTP transport, built from the supplied
// transport options over a clone of http.DefaultTransport. Use WithTransport
// instead to supply a transport built elsewhere, such as an instrumented one.
func (o ClientOpts) WithTransportOpts(options ...func(*http.Transport) *http.Transport) ClientOpts {
	return gor.WithElement(o, WithTransportOpts(options...))
}

// WithCheckRedirect sets the client's redirect policy.
func (o ClientOpts) WithCheckRedirect(check func(*http.Request, []*http.Request) error) ClientOpts {
	return gor.WithElement(o, WithCheckRedirect(check))
}

// WithCookieJar sets the client's cookie jar.
func (o ClientOpts) WithCookieJar(jar http.CookieJar) ClientOpts {
	return gor.WithElement(o, WithCookieJar(jar))
}

// WithTimeout sets the total request timeout.
func (o ClientOpts) WithTimeout(timeout time.Duration) ClientOpts {
	return gor.WithElement(o, WithTimeout(timeout))
}

// WithTransport sets the client's HTTP transport.
func WithTransport(transport http.RoundTripper) func(*http.Client) *http.Client {
	return func(client *http.Client) *http.Client {
		client.Transport = transport
		return client
	}
}

// WithTransportOpts sets the client's HTTP transport, built from the supplied
// transport options over a clone of http.DefaultTransport. Use WithTransport
// instead to supply a transport built elsewhere, such as an instrumented one.
func WithTransportOpts(options ...func(*http.Transport) *http.Transport) func(*http.Client) *http.Client {
	return func(client *http.Client) *http.Client {
		client.Transport = NewTransport(options...)
		return client
	}
}

// WithCheckRedirect sets the client's redirect policy.
func WithCheckRedirect(check func(*http.Request, []*http.Request) error) func(*http.Client) *http.Client {
	return func(client *http.Client) *http.Client {
		client.CheckRedirect = check
		return client
	}
}

// WithCookieJar sets the client's cookie jar.
func WithCookieJar(jar http.CookieJar) func(*http.Client) *http.Client {
	return func(client *http.Client) *http.Client {
		client.Jar = jar
		return client
	}
}

// WithTimeout sets the total request timeout.
func WithTimeout(timeout time.Duration) func(*http.Client) *http.Client {
	return func(client *http.Client) *http.Client {
		client.Timeout = timeout
		return client
	}
}
