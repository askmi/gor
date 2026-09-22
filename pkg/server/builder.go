package gor

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"
)

// NewServerOpts creates a configuration holding the supplied options.
func NewServerOpts(options ...func(*http.Server) *http.Server) ServerOpts {
	return append(make(ServerOpts, 0, len(options)), options...)
}

func (o ServerOpts) apply(server *http.Server) *http.Server {
	for _, option := range o {
		server = option(server)
	}
	return server
}

// WithOption appends an arbitrary option, for settings without a dedicated method.
func (o ServerOpts) WithOption(option func(*http.Server) *http.Server) ServerOpts {
	return WithElement(o, option)
}

// WithPort sets the port the server listens on.
func (o ServerOpts) WithPort(port int) ServerOpts {
	return WithElement(o, WithPort(port))
}

// WithReadHeaderTimeout sets the request-header read timeout.
func (o ServerOpts) WithReadHeaderTimeout(timeout time.Duration) ServerOpts {
	return WithElement(o, WithReadHeaderTimeout(timeout))
}

// WithReadTimeout sets the request read timeout.
func (o ServerOpts) WithReadTimeout(timeout time.Duration) ServerOpts {
	return WithElement(o, WithReadTimeout(timeout))
}

// WithWriteTimeout sets the response write timeout.
func (o ServerOpts) WithWriteTimeout(timeout time.Duration) ServerOpts {
	return WithElement(o, WithWriteTimeout(timeout))
}

// WithIdleTimeout sets the keep-alive idle timeout.
func (o ServerOpts) WithIdleTimeout(timeout time.Duration) ServerOpts {
	return WithElement(o, WithIdleTimeout(timeout))
}

// WithMaxHeaderBytes sets the maximum request-header size.
func (o ServerOpts) WithMaxHeaderBytes(bytes int) ServerOpts {
	return WithElement(o, WithMaxHeaderBytes(bytes))
}

// WithMaxHeaderValueCount sets the maximum number of header values.
func (o ServerOpts) WithMaxHeaderValueCount(count int) ServerOpts {
	return WithElement(o, WithMaxHeaderValueCount(count))
}

// WithDisableGeneralOptionsHandler controls automatic OPTIONS * handling.
func (o ServerOpts) WithDisableGeneralOptionsHandler(disable bool) ServerOpts {
	return WithElement(o, WithDisableGeneralOptionsHandler(disable))
}

// WithTLSConfig sets the server TLS configuration.
func (o ServerOpts) WithTLSConfig(config *tls.Config) ServerOpts {
	return WithElement(o, WithTLSConfig(config))
}

// WithTLSNextProto sets handlers for negotiated TLS protocols.
func (o ServerOpts) WithTLSNextProto(nextProto map[string]func(*http.Server, *tls.Conn, http.Handler)) ServerOpts {
	return WithElement(o, WithTLSNextProto(nextProto))
}

// WithConnState sets the connection-state callback.
func (o ServerOpts) WithConnState(callback func(net.Conn, http.ConnState)) ServerOpts {
	return WithElement(o, WithConnState(callback))
}

// WithErrorLog sets the server error logger.
func (o ServerOpts) WithErrorLog(logger *log.Logger) ServerOpts {
	return WithElement(o, WithErrorLog(logger))
}

// WithBaseContext sets the base context provider.
func (o ServerOpts) WithBaseContext(baseContext func(net.Listener) context.Context) ServerOpts {
	return WithElement(o, WithBaseContext(baseContext))
}

// WithConnContext sets the connection context provider.
func (o ServerOpts) WithConnContext(connContext func(context.Context, net.Conn) context.Context) ServerOpts {
	return WithElement(o, WithConnContext(connContext))
}

// WithHTTP2 sets the HTTP/2 configuration.
func (o ServerOpts) WithHTTP2(config *http.HTTP2Config) ServerOpts {
	return WithElement(o, WithHTTP2(config))
}

// WithProtocols sets the accepted HTTP protocols.
func (o ServerOpts) WithProtocols(protocols *http.Protocols) ServerOpts {
	return WithElement(o, WithProtocols(protocols))
}

// WithDisableClientPriority controls HTTP/2 client priority handling.
func (o ServerOpts) WithDisableClientPriority(disable bool) ServerOpts {
	return WithElement(o, WithDisableClientPriority(disable))
}

// WithPort sets the port the server listens on. It panics if port is outside
// the valid range; port 0 asks the operating system for an unused port.
func WithPort(port int) func(*http.Server) *http.Server {
	if port < 0 || port > 65535 {
		panic("server: invalid port " + strconv.Itoa(port))
	}

	return func(server *http.Server) *http.Server {
		server.Addr = ":" + strconv.Itoa(port)
		return server
	}
}

// WithReadHeaderTimeout sets the request-header read timeout.
func WithReadHeaderTimeout(timeout time.Duration) func(*http.Server) *http.Server {
	return func(server *http.Server) *http.Server {
		server.ReadHeaderTimeout = timeout
		return server
	}
}

// WithReadTimeout sets the request read timeout.
func WithReadTimeout(timeout time.Duration) func(*http.Server) *http.Server {
	return func(server *http.Server) *http.Server {
		server.ReadTimeout = timeout
		return server
	}
}

// WithWriteTimeout sets the response write timeout.
func WithWriteTimeout(timeout time.Duration) func(*http.Server) *http.Server {
	return func(server *http.Server) *http.Server {
		server.WriteTimeout = timeout
		return server
	}
}

// WithIdleTimeout sets the keep-alive idle timeout.
func WithIdleTimeout(timeout time.Duration) func(*http.Server) *http.Server {
	return func(server *http.Server) *http.Server {
		server.IdleTimeout = timeout
		return server
	}
}

// WithMaxHeaderBytes sets the maximum request-header size.
func WithMaxHeaderBytes(bytes int) func(*http.Server) *http.Server {
	return func(server *http.Server) *http.Server {
		server.MaxHeaderBytes = bytes
		return server
	}
}

// WithMaxHeaderValueCount sets the maximum number of header values.
func WithMaxHeaderValueCount(count int) func(*http.Server) *http.Server {
	return func(server *http.Server) *http.Server {
		server.MaxHeaderValueCount = count
		return server
	}
}

// WithDisableGeneralOptionsHandler controls automatic OPTIONS * handling.
func WithDisableGeneralOptionsHandler(disable bool) func(*http.Server) *http.Server {
	return func(server *http.Server) *http.Server {
		server.DisableGeneralOptionsHandler = disable
		return server
	}
}

// WithTLSConfig sets the server TLS configuration.
func WithTLSConfig(config *tls.Config) func(*http.Server) *http.Server {
	return func(server *http.Server) *http.Server {
		server.TLSConfig = config
		return server
	}
}

// WithTLSNextProto sets handlers for negotiated TLS protocols.
func WithTLSNextProto(nextProto map[string]func(*http.Server, *tls.Conn, http.Handler)) func(*http.Server) *http.Server {
	return func(server *http.Server) *http.Server {
		server.TLSNextProto = nextProto
		return server
	}
}

// WithConnState sets the connection-state callback.
func WithConnState(callback func(net.Conn, http.ConnState)) func(*http.Server) *http.Server {
	return func(server *http.Server) *http.Server {
		server.ConnState = callback
		return server
	}
}

// WithErrorLog sets the server error logger.
func WithErrorLog(logger *log.Logger) func(*http.Server) *http.Server {
	return func(server *http.Server) *http.Server {
		server.ErrorLog = logger
		return server
	}
}

// WithBaseContext sets the base context provider.
func WithBaseContext(baseContext func(net.Listener) context.Context) func(*http.Server) *http.Server {
	return func(server *http.Server) *http.Server {
		server.BaseContext = baseContext
		return server
	}
}

// WithConnContext sets the connection context provider.
func WithConnContext(connContext func(context.Context, net.Conn) context.Context) func(*http.Server) *http.Server {
	return func(server *http.Server) *http.Server {
		server.ConnContext = connContext
		return server
	}
}

// WithHTTP2 sets the HTTP/2 configuration.
func WithHTTP2(config *http.HTTP2Config) func(*http.Server) *http.Server {
	return func(server *http.Server) *http.Server {
		server.HTTP2 = config
		return server
	}
}

// WithProtocols sets the accepted HTTP protocols.
func WithProtocols(protocols *http.Protocols) func(*http.Server) *http.Server {
	return func(server *http.Server) *http.Server {
		server.Protocols = protocols
		return server
	}
}

// WithDisableClientPriority controls HTTP/2 client priority handling.
func WithDisableClientPriority(disable bool) func(*http.Server) *http.Server {
	return func(server *http.Server) *http.Server {
		server.DisableClientPriority = disable
		return server
	}
}
