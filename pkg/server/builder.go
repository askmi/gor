package gor

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"time"
)

// NewServerOpts creates an empty HTTP server configuration.
func NewServerOpts() ServerOpts {
	return ServerOpts{}
}

func (o ServerOpts) apply(server *http.Server) *http.Server {
	for _, option := range o {
		server = option.apply(server)
	}
	return server
}

// with copies the backing array so derived configurations cannot modify each other.
func (o ServerOpts) with(option OptionFunc[*http.Server]) ServerOpts {
	opts := make([]OptionFunc[*http.Server], len(o), len(o)+1)
	copy(opts, o)
	return append(opts, option)
}

// WithReadHeaderTimeout sets the request-header read timeout.
func (o ServerOpts) WithReadHeaderTimeout(timeout time.Duration) ServerOpts {
	return o.with(func(server *http.Server) *http.Server {
		server.ReadHeaderTimeout = timeout
		return server
	})
}

// WithReadTimeout sets the request read timeout.
func (o ServerOpts) WithReadTimeout(timeout time.Duration) ServerOpts {
	return o.with(func(server *http.Server) *http.Server {
		server.ReadTimeout = timeout
		return server
	})
}

// WithWriteTimeout sets the response write timeout.
func (o ServerOpts) WithWriteTimeout(timeout time.Duration) ServerOpts {
	return o.with(func(server *http.Server) *http.Server {
		server.WriteTimeout = timeout
		return server
	})
}

// WithIdleTimeout sets the keep-alive idle timeout.
func (o ServerOpts) WithIdleTimeout(timeout time.Duration) ServerOpts {
	return o.with(func(server *http.Server) *http.Server {
		server.IdleTimeout = timeout
		return server
	})
}

// WithMaxHeaderBytes sets the maximum request-header size.
func (o ServerOpts) WithMaxHeaderBytes(bytes int) ServerOpts {
	return o.with(func(server *http.Server) *http.Server {
		server.MaxHeaderBytes = bytes
		return server
	})
}

// WithMaxHeaderValueCount sets the maximum number of header values.
func (o ServerOpts) WithMaxHeaderValueCount(count int) ServerOpts {
	return o.with(func(server *http.Server) *http.Server {
		server.MaxHeaderValueCount = count
		return server
	})
}

// WithDisableGeneralOptionsHandler controls automatic OPTIONS * handling.
func (o ServerOpts) WithDisableGeneralOptionsHandler(disable bool) ServerOpts {
	return o.with(func(server *http.Server) *http.Server {
		server.DisableGeneralOptionsHandler = disable
		return server
	})
}

// WithTLSConfig sets the server TLS configuration.
func (o ServerOpts) WithTLSConfig(config *tls.Config) ServerOpts {
	return o.with(func(server *http.Server) *http.Server {
		server.TLSConfig = config
		return server
	})
}

// WithTLSNextProto sets handlers for negotiated TLS protocols.
func (o ServerOpts) WithTLSNextProto(nextProto map[string]func(*http.Server, *tls.Conn, http.Handler)) ServerOpts {
	return o.with(func(server *http.Server) *http.Server {
		server.TLSNextProto = nextProto
		return server
	})
}

// WithConnState sets the connection-state callback.
func (o ServerOpts) WithConnState(callback func(net.Conn, http.ConnState)) ServerOpts {
	return o.with(func(server *http.Server) *http.Server {
		server.ConnState = callback
		return server
	})
}

// WithErrorLog sets the server error logger.
func (o ServerOpts) WithErrorLog(logger *log.Logger) ServerOpts {
	return o.with(func(server *http.Server) *http.Server {
		server.ErrorLog = logger
		return server
	})
}

// WithBaseContext sets the base context provider.
func (o ServerOpts) WithBaseContext(baseContext func(net.Listener) context.Context) ServerOpts {
	return o.with(func(server *http.Server) *http.Server {
		server.BaseContext = baseContext
		return server
	})
}

// WithConnContext sets the connection context provider.
func (o ServerOpts) WithConnContext(connContext func(context.Context, net.Conn) context.Context) ServerOpts {
	return o.with(func(server *http.Server) *http.Server {
		server.ConnContext = connContext
		return server
	})
}

// WithHTTP2 sets the HTTP/2 configuration.
func (o ServerOpts) WithHTTP2(config *http.HTTP2Config) ServerOpts {
	return o.with(func(server *http.Server) *http.Server {
		server.HTTP2 = config
		return server
	})
}

// WithProtocols sets the accepted HTTP protocols.
func (o ServerOpts) WithProtocols(protocols *http.Protocols) ServerOpts {
	return o.with(func(server *http.Server) *http.Server {
		server.Protocols = protocols
		return server
	})
}

// WithDisableClientPriority controls HTTP/2 client priority handling.
func (o ServerOpts) WithDisableClientPriority(disable bool) ServerOpts {
	return o.with(func(server *http.Server) *http.Server {
		server.DisableClientPriority = disable
		return server
	})
}

// WithReadHeaderTimeout sets the request-header read timeout.
func WithReadHeaderTimeout(timeout time.Duration) Option[ServerOpts] {
	return OptionFunc[ServerOpts](func(opts ServerOpts) ServerOpts {
		return opts.WithReadHeaderTimeout(timeout)
	})
}

// WithReadTimeout sets the request read timeout.
func WithReadTimeout(timeout time.Duration) Option[ServerOpts] {
	return OptionFunc[ServerOpts](func(opts ServerOpts) ServerOpts {
		return opts.WithReadTimeout(timeout)
	})
}

// WithWriteTimeout sets the response write timeout.
func WithWriteTimeout(timeout time.Duration) Option[ServerOpts] {
	return OptionFunc[ServerOpts](func(opts ServerOpts) ServerOpts {
		return opts.WithWriteTimeout(timeout)
	})
}

// WithIdleTimeout sets the keep-alive idle timeout.
func WithIdleTimeout(timeout time.Duration) Option[ServerOpts] {
	return OptionFunc[ServerOpts](func(opts ServerOpts) ServerOpts {
		return opts.WithIdleTimeout(timeout)
	})
}

// WithMaxHeaderBytes sets the maximum request-header size.
func WithMaxHeaderBytes(bytes int) Option[ServerOpts] {
	return OptionFunc[ServerOpts](func(opts ServerOpts) ServerOpts {
		return opts.WithMaxHeaderBytes(bytes)
	})
}

// WithMaxHeaderValueCount sets the maximum number of header values.
func WithMaxHeaderValueCount(count int) Option[ServerOpts] {
	return OptionFunc[ServerOpts](func(opts ServerOpts) ServerOpts {
		return opts.WithMaxHeaderValueCount(count)
	})
}

// WithDisableGeneralOptionsHandler controls automatic OPTIONS * handling.
func WithDisableGeneralOptionsHandler(disable bool) Option[ServerOpts] {
	return OptionFunc[ServerOpts](func(opts ServerOpts) ServerOpts {
		return opts.WithDisableGeneralOptionsHandler(disable)
	})
}

// WithTLSConfig sets the server TLS configuration.
func WithTLSConfig(config *tls.Config) Option[ServerOpts] {
	return OptionFunc[ServerOpts](func(opts ServerOpts) ServerOpts {
		return opts.WithTLSConfig(config)
	})
}

// WithTLSNextProto sets handlers for negotiated TLS protocols.
func WithTLSNextProto(nextProto map[string]func(*http.Server, *tls.Conn, http.Handler)) Option[ServerOpts] {
	return OptionFunc[ServerOpts](func(opts ServerOpts) ServerOpts {
		return opts.WithTLSNextProto(nextProto)
	})
}

// WithConnState sets the connection-state callback.
func WithConnState(callback func(net.Conn, http.ConnState)) Option[ServerOpts] {
	return OptionFunc[ServerOpts](func(opts ServerOpts) ServerOpts {
		return opts.WithConnState(callback)
	})
}

// WithErrorLog sets the server error logger.
func WithErrorLog(logger *log.Logger) Option[ServerOpts] {
	return OptionFunc[ServerOpts](func(opts ServerOpts) ServerOpts {
		return opts.WithErrorLog(logger)
	})
}

// WithBaseContext sets the base context provider.
func WithBaseContext(baseContext func(net.Listener) context.Context) Option[ServerOpts] {
	return OptionFunc[ServerOpts](func(opts ServerOpts) ServerOpts {
		return opts.WithBaseContext(baseContext)
	})
}

// WithConnContext sets the connection context provider.
func WithConnContext(connContext func(context.Context, net.Conn) context.Context) Option[ServerOpts] {
	return OptionFunc[ServerOpts](func(opts ServerOpts) ServerOpts {
		return opts.WithConnContext(connContext)
	})
}

// WithHTTP2 sets the HTTP/2 configuration.
func WithHTTP2(config *http.HTTP2Config) Option[ServerOpts] {
	return OptionFunc[ServerOpts](func(opts ServerOpts) ServerOpts {
		return opts.WithHTTP2(config)
	})
}

// WithProtocols sets the accepted HTTP protocols.
func WithProtocols(protocols *http.Protocols) Option[ServerOpts] {
	return OptionFunc[ServerOpts](func(opts ServerOpts) ServerOpts {
		return opts.WithProtocols(protocols)
	})
}

// WithDisableClientPriority controls HTTP/2 client priority handling.
func WithDisableClientPriority(disable bool) Option[ServerOpts] {
	return OptionFunc[ServerOpts](func(opts ServerOpts) ServerOpts {
		return opts.WithDisableClientPriority(disable)
	})
}
