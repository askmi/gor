package client

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"time"

	gor "gor/pkg/server"
)

// https://go.dev/src/net/http/transport.go

// Transport settings are separate from client settings because they live on a
// separate standard type. *http.Client holds per-request policy, while
// *http.Transport holds the connection pool and everything about establishing a
// connection, so TransportOpts configures the second and ClientOpts the first.
//
//	*http.Transport                 option                      default
//	├─ pool
//	│  ├── MaxIdleConns             WithMaxIdleConns            100
//	│  ├── MaxIdleConnsPerHost      WithMaxIdleConnsPerHost     2
//	│  ├── MaxConnsPerHost          WithMaxConnsPerHost         0 (no limit)
//	│  ├── IdleConnTimeout          WithIdleConnTimeout         90s
//	│  └── DisableKeepAlives        WithDisableKeepAlives       false
//	├─ dial
//	│  ├── DialContext              WithDialContext             30s timeout, 30s keep-alive
//	│  ├── DialTLSContext           WithDialTLSContext          nil
//	│  ├── TLSClientConfig          WithTLSClientConfig         nil (system roots)
//	│  ├── TLSHandshakeTimeout      WithTLSHandshakeTimeout     10s
//	│  └── TLSNextProto             WithTLSNextProto            nil
//	├─ per-request timing
//	│  ├── ResponseHeaderTimeout    WithResponseHeaderTimeout   0 (no limit)
//	│  └── ExpectContinueTimeout    WithExpectContinueTimeout   1s
//	├─ proxy
//	│  ├── Proxy                    WithProxy                   ProxyFromEnvironment
//	│  ├── ProxyConnectHeader       WithProxyConnectHeader      nil
//	│  ├── GetProxyConnectHeader    WithGetProxyConnectHeader   nil
//	│  └── OnProxyConnectResponse   WithOnProxyConnectResponse  nil
//	└─ protocol and buffers
//	   ├── ForceAttemptHTTP2        WithForceAttemptHTTP2       true
//	   ├── HTTP2                    WithHTTP2                   nil
//	   ├── Protocols                WithProtocols               nil (HTTP/1.1 and HTTP/2)
//	   ├── DisableCompression       WithDisableCompression      false
//	   ├── MaxResponseHeaderBytes   WithMaxResponseHeaderBytes  0 (10 MB)
//	   ├── ReadBufferSize           WithReadBufferSize          0 (4 KB)
//	   └── WriteBufferSize          WithWriteBufferSize         0 (4 KB)
//
// The defaults are those of http.DefaultTransport, which NewTransport clones; a
// bare &http.Transport{} has every field zeroed instead, losing the proxy, the
// dialer and the 100-connection idle pool. MaxIdleConnsPerHost is the one worth
// raising: its effective default of 2 closes rather than pools any further idle
// connection, so a busy client keeps re-dialing. Connection reuse also requires
// each response body to be read to EOF and closed.
//
// The pool belongs to the transport, so connections are reused only between
// requests sharing one transport instance. Build the transport once, hand it to
// one client, and pass that client around:
//
//	transport := NewTransport(
//		WithMaxIdleConnsPerHost(100),
//		WithIdleConnTimeout(90*time.Second),
//	)
//
//	client := NewClient(
//		WithTransport(transport),
//		WithTimeout(30*time.Second),
//	)
//
// Wrapping for instrumentation takes the same transport, since the wrapper
// delegates pooling to it:
//
//	client := NewClient(
//		WithTransport(otelhttp.NewTransport(transport)),
//		WithTimeout(30*time.Second),
//	)
//
// ClientOpts.WithTransportOpts builds the transport from options in one step,
// for a client that needs no wrapping.

type (
	// TransportOpts contains HTTP transport configuration, including the
	// connection pool.
	TransportOpts []func(*http.Transport) *http.Transport
)

// NewTransport creates an HTTP transport with options applied in order.
// It starts from a clone of http.DefaultTransport, so proxy-from-environment,
// the default dialer and the default pool limits stay in place unless overridden.
func NewTransport(options ...func(*http.Transport) *http.Transport) *http.Transport {
	t := http.DefaultTransport.(*http.Transport).Clone()
	for _, option := range options {
		t = option(t)
	}
	return t
}

// NewTransportOpts creates a configuration holding the supplied options.
func NewTransportOpts(options ...func(*http.Transport) *http.Transport) TransportOpts {
	return append(make(TransportOpts, 0, len(options)), options...)
}

// WithOption appends an arbitrary option, for settings without a dedicated method.
func (o TransportOpts) WithOption(option func(*http.Transport) *http.Transport) TransportOpts {
	return gor.WithElement(o, option)
}

// WithMaxIdleConns sets the total number of idle connections kept across all
// hosts. Zero means no limit.
func (o TransportOpts) WithMaxIdleConns(n int) TransportOpts {
	return gor.WithElement(o, WithMaxIdleConns(n))
}

// WithMaxIdleConnsPerHost sets the number of idle connections kept per host.
// Zero selects http.DefaultMaxIdleConnsPerHost, which is 2.
func (o TransportOpts) WithMaxIdleConnsPerHost(n int) TransportOpts {
	return gor.WithElement(o, WithMaxIdleConnsPerHost(n))
}

// WithMaxConnsPerHost sets the total number of connections per host, counting
// dialing, active and idle ones. Zero means no limit.
func (o TransportOpts) WithMaxConnsPerHost(n int) TransportOpts {
	return gor.WithElement(o, WithMaxConnsPerHost(n))
}

// WithIdleConnTimeout sets how long an idle connection is kept before closing.
// Zero means no limit.
func (o TransportOpts) WithIdleConnTimeout(timeout time.Duration) TransportOpts {
	return gor.WithElement(o, WithIdleConnTimeout(timeout))
}

// WithDisableKeepAlives disables connection reuse, closing each connection
// after a single request.
func (o TransportOpts) WithDisableKeepAlives(disable bool) TransportOpts {
	return gor.WithElement(o, WithDisableKeepAlives(disable))
}

// WithDisableCompression stops the transport from requesting gzip encoding.
func (o TransportOpts) WithDisableCompression(disable bool) TransportOpts {
	return gor.WithElement(o, WithDisableCompression(disable))
}

// WithDialContext sets the function that establishes unencrypted connections.
func (o TransportOpts) WithDialContext(dial func(context.Context, string, string) (net.Conn, error)) TransportOpts {
	return gor.WithElement(o, WithDialContext(dial))
}

// WithDialTLSContext sets the function that establishes TLS connections,
// bypassing TLSClientConfig and TLSHandshakeTimeout when set.
func (o TransportOpts) WithDialTLSContext(dial func(context.Context, string, string) (net.Conn, error)) TransportOpts {
	return gor.WithElement(o, WithDialTLSContext(dial))
}

// WithTLSClientConfig sets the TLS configuration used for HTTPS requests.
func (o TransportOpts) WithTLSClientConfig(config *tls.Config) TransportOpts {
	return gor.WithElement(o, WithTLSClientConfig(config))
}

// WithTLSHandshakeTimeout sets the TLS handshake timeout. Zero means no limit.
func (o TransportOpts) WithTLSHandshakeTimeout(timeout time.Duration) TransportOpts {
	return gor.WithElement(o, WithTLSHandshakeTimeout(timeout))
}

// WithTLSNextProto sets the negotiated-protocol handlers for TLS connections.
// A non-nil map disables HTTP/2 unless it contains an "h2" entry.
func (o TransportOpts) WithTLSNextProto(next map[string]func(string, *tls.Conn) http.RoundTripper) TransportOpts {
	return gor.WithElement(o, WithTLSNextProto(next))
}

// WithResponseHeaderTimeout limits the wait for response headers after the
// request body is written. Zero means no limit.
func (o TransportOpts) WithResponseHeaderTimeout(timeout time.Duration) TransportOpts {
	return gor.WithElement(o, WithResponseHeaderTimeout(timeout))
}

// WithExpectContinueTimeout limits the wait for a 100-continue response when
// the request carries an Expect: 100-continue header. Zero means no limit.
func (o TransportOpts) WithExpectContinueTimeout(timeout time.Duration) TransportOpts {
	return gor.WithElement(o, WithExpectContinueTimeout(timeout))
}

// WithMaxResponseHeaderBytes sets the maximum response-header size.
// Zero selects the 10 MB default.
func (o TransportOpts) WithMaxResponseHeaderBytes(bytes int64) TransportOpts {
	return gor.WithElement(o, WithMaxResponseHeaderBytes(bytes))
}

// WithReadBufferSize sets the read buffer size used per connection.
// Zero selects the 4 KB default.
func (o TransportOpts) WithReadBufferSize(size int) TransportOpts {
	return gor.WithElement(o, WithReadBufferSize(size))
}

// WithWriteBufferSize sets the write buffer size used per connection.
// Zero selects the 4 KB default.
func (o TransportOpts) WithWriteBufferSize(size int) TransportOpts {
	return gor.WithElement(o, WithWriteBufferSize(size))
}

// WithProxy sets the function returning the proxy to use for a request.
// A nil URL means no proxy.
func (o TransportOpts) WithProxy(proxy func(*http.Request) (*url.URL, error)) TransportOpts {
	return gor.WithElement(o, WithProxy(proxy))
}

// WithProxyConnectHeader sets the headers sent to a proxy during CONNECT.
func (o TransportOpts) WithProxyConnectHeader(header http.Header) TransportOpts {
	return gor.WithElement(o, WithProxyConnectHeader(header))
}

// WithGetProxyConnectHeader sets the function returning the headers sent to a
// proxy during CONNECT, taking precedence over WithProxyConnectHeader.
func (o TransportOpts) WithGetProxyConnectHeader(get func(context.Context, *url.URL, string) (http.Header, error)) TransportOpts {
	return gor.WithElement(o, WithGetProxyConnectHeader(get))
}

// WithOnProxyConnectResponse sets the hook called with the proxy's CONNECT
// response. A non-nil error fails the request.
func (o TransportOpts) WithOnProxyConnectResponse(f func(context.Context, *url.URL, *http.Request, *http.Response) error) TransportOpts {
	return gor.WithElement(o, WithOnProxyConnectResponse(f))
}

// WithForceAttemptHTTP2 enables HTTP/2 even when a custom dialer or TLS
// configuration is set.
func (o TransportOpts) WithForceAttemptHTTP2(force bool) TransportOpts {
	return gor.WithElement(o, WithForceAttemptHTTP2(force))
}

// WithHTTP2 sets the HTTP/2 configuration.
func (o TransportOpts) WithHTTP2(config *http.HTTP2Config) TransportOpts {
	return gor.WithElement(o, WithHTTP2(config))
}

// WithProtocols sets the protocols the transport may use.
func (o TransportOpts) WithProtocols(protocols *http.Protocols) TransportOpts {
	return gor.WithElement(o, WithProtocols(protocols))
}

// WithMaxIdleConns sets the total number of idle connections kept across all
// hosts. Zero means no limit.
func WithMaxIdleConns(n int) func(*http.Transport) *http.Transport {
	return func(t *http.Transport) *http.Transport {
		t.MaxIdleConns = n
		return t
	}
}

// WithMaxIdleConnsPerHost sets the number of idle connections kept per host.
// Zero selects http.DefaultMaxIdleConnsPerHost, which is 2.
func WithMaxIdleConnsPerHost(n int) func(*http.Transport) *http.Transport {
	return func(t *http.Transport) *http.Transport {
		t.MaxIdleConnsPerHost = n
		return t
	}
}

// WithMaxConnsPerHost sets the total number of connections per host, counting
// dialing, active and idle ones. Zero means no limit.
func WithMaxConnsPerHost(n int) func(*http.Transport) *http.Transport {
	return func(t *http.Transport) *http.Transport {
		t.MaxConnsPerHost = n
		return t
	}
}

// WithIdleConnTimeout sets how long an idle connection is kept before closing.
// Zero means no limit.
func WithIdleConnTimeout(timeout time.Duration) func(*http.Transport) *http.Transport {
	return func(t *http.Transport) *http.Transport {
		t.IdleConnTimeout = timeout
		return t
	}
}

// WithDisableKeepAlives disables connection reuse, closing each connection
// after a single request.
func WithDisableKeepAlives(disable bool) func(*http.Transport) *http.Transport {
	return func(t *http.Transport) *http.Transport {
		t.DisableKeepAlives = disable
		return t
	}
}

// WithDisableCompression stops the transport from requesting gzip encoding.
func WithDisableCompression(disable bool) func(*http.Transport) *http.Transport {
	return func(t *http.Transport) *http.Transport {
		t.DisableCompression = disable
		return t
	}
}

// WithDialContext sets the function that establishes unencrypted connections.
func WithDialContext(dial func(context.Context, string, string) (net.Conn, error)) func(*http.Transport) *http.Transport {
	return func(t *http.Transport) *http.Transport {
		t.DialContext = dial
		return t
	}
}

// WithDialTLSContext sets the function that establishes TLS connections,
// bypassing TLSClientConfig and TLSHandshakeTimeout when set.
func WithDialTLSContext(dial func(context.Context, string, string) (net.Conn, error)) func(*http.Transport) *http.Transport {
	return func(t *http.Transport) *http.Transport {
		t.DialTLSContext = dial
		return t
	}
}

// WithTLSClientConfig sets the TLS configuration used for HTTPS requests.
func WithTLSClientConfig(config *tls.Config) func(*http.Transport) *http.Transport {
	return func(t *http.Transport) *http.Transport {
		t.TLSClientConfig = config
		return t
	}
}

// WithTLSHandshakeTimeout sets the TLS handshake timeout. Zero means no limit.
func WithTLSHandshakeTimeout(timeout time.Duration) func(*http.Transport) *http.Transport {
	return func(t *http.Transport) *http.Transport {
		t.TLSHandshakeTimeout = timeout
		return t
	}
}

// WithTLSNextProto sets the negotiated-protocol handlers for TLS connections.
// A non-nil map disables HTTP/2 unless it contains an "h2" entry.
func WithTLSNextProto(next map[string]func(string, *tls.Conn) http.RoundTripper) func(*http.Transport) *http.Transport {
	return func(t *http.Transport) *http.Transport {
		t.TLSNextProto = next
		return t
	}
}

// WithResponseHeaderTimeout limits the wait for response headers after the
// request body is written. Zero means no limit.
func WithResponseHeaderTimeout(timeout time.Duration) func(*http.Transport) *http.Transport {
	return func(t *http.Transport) *http.Transport {
		t.ResponseHeaderTimeout = timeout
		return t
	}
}

// WithExpectContinueTimeout limits the wait for a 100-continue response when
// the request carries an Expect: 100-continue header. Zero means no limit.
func WithExpectContinueTimeout(timeout time.Duration) func(*http.Transport) *http.Transport {
	return func(t *http.Transport) *http.Transport {
		t.ExpectContinueTimeout = timeout
		return t
	}
}

// WithMaxResponseHeaderBytes sets the maximum response-header size.
// Zero selects the 10 MB default.
func WithMaxResponseHeaderBytes(bytes int64) func(*http.Transport) *http.Transport {
	return func(t *http.Transport) *http.Transport {
		t.MaxResponseHeaderBytes = bytes
		return t
	}
}

// WithReadBufferSize sets the read buffer size used per connection.
// Zero selects the 4 KB default.
func WithReadBufferSize(size int) func(*http.Transport) *http.Transport {
	return func(t *http.Transport) *http.Transport {
		t.ReadBufferSize = size
		return t
	}
}

// WithWriteBufferSize sets the write buffer size used per connection.
// Zero selects the 4 KB default.
func WithWriteBufferSize(size int) func(*http.Transport) *http.Transport {
	return func(t *http.Transport) *http.Transport {
		t.WriteBufferSize = size
		return t
	}
}

// WithProxy sets the function returning the proxy to use for a request.
// A nil URL means no proxy.
func WithProxy(proxy func(*http.Request) (*url.URL, error)) func(*http.Transport) *http.Transport {
	return func(t *http.Transport) *http.Transport {
		t.Proxy = proxy
		return t
	}
}

// WithProxyConnectHeader sets the headers sent to a proxy during CONNECT.
func WithProxyConnectHeader(header http.Header) func(*http.Transport) *http.Transport {
	return func(t *http.Transport) *http.Transport {
		t.ProxyConnectHeader = header
		return t
	}
}

// WithGetProxyConnectHeader sets the function returning the headers sent to a
// proxy during CONNECT, taking precedence over WithProxyConnectHeader.
func WithGetProxyConnectHeader(get func(context.Context, *url.URL, string) (http.Header, error)) func(*http.Transport) *http.Transport {
	return func(t *http.Transport) *http.Transport {
		t.GetProxyConnectHeader = get
		return t
	}
}

// WithOnProxyConnectResponse sets the hook called with the proxy's CONNECT
// response. A non-nil error fails the request.
func WithOnProxyConnectResponse(f func(context.Context, *url.URL, *http.Request, *http.Response) error) func(*http.Transport) *http.Transport {
	return func(t *http.Transport) *http.Transport {
		t.OnProxyConnectResponse = f
		return t
	}
}

// WithForceAttemptHTTP2 enables HTTP/2 even when a custom dialer or TLS
// configuration is set.
func WithForceAttemptHTTP2(force bool) func(*http.Transport) *http.Transport {
	return func(t *http.Transport) *http.Transport {
		t.ForceAttemptHTTP2 = force
		return t
	}
}

// WithHTTP2 sets the HTTP/2 configuration.
func WithHTTP2(config *http.HTTP2Config) func(*http.Transport) *http.Transport {
	return func(t *http.Transport) *http.Transport {
		t.HTTP2 = config
		return t
	}
}

// WithProtocols sets the protocols the transport may use.
func WithProtocols(protocols *http.Protocols) func(*http.Transport) *http.Transport {
	return func(t *http.Transport) *http.Transport {
		t.Protocols = protocols
		return t
	}
}
