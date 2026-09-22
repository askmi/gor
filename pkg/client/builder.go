package client

import (
	"net/http"
	"time"

	gor "gor/pkg/server"
)

// https://go.dev/src/net/http/client.go

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
	return gor.WithElement(o, func(client *http.Client) *http.Client {
		client.Transport = transport
		return client
	})
}

// WithCheckRedirect sets the client's redirect policy.
func (o ClientOpts) WithCheckRedirect(check func(*http.Request, []*http.Request) error) ClientOpts {
	return gor.WithElement(o, func(client *http.Client) *http.Client {
		client.CheckRedirect = check
		return client
	})
}

// WithCookieJar sets the client's cookie jar.
func (o ClientOpts) WithCookieJar(jar http.CookieJar) ClientOpts {
	return gor.WithElement(o, func(client *http.Client) *http.Client {
		client.Jar = jar
		return client
	})
}

// WithTimeout sets the total request timeout.
func (o ClientOpts) WithTimeout(timeout time.Duration) ClientOpts {
	return gor.WithElement(o, func(client *http.Client) *http.Client {
		client.Timeout = timeout
		return client
	})
}

// WithCheckRedirect sets the client's redirect policy.
func WithCheckRedirect(check func(*http.Request, []*http.Request) error) func(*http.Client) *http.Client {
	return func(c *http.Client) *http.Client {
		c.CheckRedirect = check
		return c
	}
}
