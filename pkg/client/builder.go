package client

import (
	"net/http"
	"time"

	gor "gor/pkg/server"
)

// https://go.dev/src/net/http/client.go

type (
	// ClientBuilder stores reusable HTTP client options.
	ClientBuilder struct {
		opts ClientOpts
	}
	// ClientOpts contains HTTP client configuration.
	ClientOpts []gor.OptionFunc[*http.Client]
)

// NewClient creates an HTTP client with options applied in order.
func NewClient(options ...ClientOpts) *http.Client {
	c := &http.Client{}
	for _, option := range gor.Merge(options...) {
		c = option(c)
	}
	return c
}

// NewClientOpts creates an empty HTTP client configuration.
func NewClientOpts() ClientOpts {
	return ClientOpts{}
}

// WithTransport sets the client's HTTP transport.
func (o ClientOpts) WithTransport(transport http.RoundTripper) ClientOpts {
	return gor.WithOption(o, func(client *http.Client) *http.Client {
		client.Transport = transport
		return client
	})
}

// WithCheckRedirect sets the client's redirect policy.
func (o ClientOpts) WithCheckRedirect(check func(*http.Request, []*http.Request) error) ClientOpts {
	return gor.WithOption(o, func(client *http.Client) *http.Client {
		client.CheckRedirect = check
		return client
	})
}

// WithCookieJar sets the client's cookie jar.
func (o ClientOpts) WithCookieJar(jar http.CookieJar) ClientOpts {
	return gor.WithOption(o, func(client *http.Client) *http.Client {
		client.Jar = jar
		return client
	})
}

// WithTimeout sets the total request timeout.
func (o ClientOpts) WithTimeout(timeout time.Duration) ClientOpts {
	return gor.WithOption(o, func(client *http.Client) *http.Client {
		client.Timeout = timeout
		return client
	})
}
