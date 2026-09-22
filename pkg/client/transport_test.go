package client

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"
)

func TestNewTransportClonesDefaults(t *testing.T) {
	transport := NewTransport()

	if transport == http.DefaultTransport {
		t.Error("NewTransport returned the shared DefaultTransport")
	}
	if transport.Proxy == nil {
		t.Error("Proxy was not inherited from DefaultTransport")
	}
	if transport.DialContext == nil {
		t.Error("DialContext was not inherited from DefaultTransport")
	}
	if transport.MaxIdleConns != 100 {
		t.Errorf("MaxIdleConns = %d, want 100", transport.MaxIdleConns)
	}
	if transport.IdleConnTimeout != 90*time.Second {
		t.Errorf("IdleConnTimeout = %v, want %v", transport.IdleConnTimeout, 90*time.Second)
	}
	if !transport.ForceAttemptHTTP2 {
		t.Error("ForceAttemptHTTP2 was not inherited from DefaultTransport")
	}
}

func TestTransportOptsMethods(t *testing.T) {
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS13}
	dial := func(context.Context, string, string) (net.Conn, error) { return nil, nil }
	dialTLS := func(context.Context, string, string) (net.Conn, error) { return nil, nil }
	proxy := func(*http.Request) (*url.URL, error) { return nil, nil }
	getProxyConnectHeader := func(context.Context, *url.URL, string) (http.Header, error) { return nil, nil }
	onProxyConnectResponse := func(context.Context, *url.URL, *http.Request, *http.Response) error { return nil }
	proxyConnectHeader := http.Header{"X-Test": []string{"1"}}
	http2Config := &http.HTTP2Config{MaxConcurrentStreams: 100}
	protocols := new(http.Protocols)
	protocols.SetHTTP1(true)
	tlsNextProto := map[string]func(string, *tls.Conn) http.RoundTripper{
		"custom": func(string, *tls.Conn) http.RoundTripper { return nil },
	}

	base := TransportOpts{}
	opts := base.
		WithMaxIdleConns(200).
		WithMaxIdleConnsPerHost(10).
		WithMaxConnsPerHost(50).
		WithIdleConnTimeout(30 * time.Second).
		WithDisableKeepAlives(true).
		WithDisableCompression(true).
		WithDialContext(dial).
		WithDialTLSContext(dialTLS).
		WithTLSClientConfig(tlsConfig).
		WithTLSHandshakeTimeout(5 * time.Second).
		WithTLSNextProto(tlsNextProto).
		WithResponseHeaderTimeout(15 * time.Second).
		WithExpectContinueTimeout(2 * time.Second).
		WithMaxResponseHeaderBytes(1 << 20).
		WithReadBufferSize(8 << 10).
		WithWriteBufferSize(16 << 10).
		WithProxy(proxy).
		WithProxyConnectHeader(proxyConnectHeader).
		WithGetProxyConnectHeader(getProxyConnectHeader).
		WithOnProxyConnectResponse(onProxyConnectResponse).
		WithForceAttemptHTTP2(true).
		WithHTTP2(http2Config).
		WithProtocols(protocols)

	transport := NewTransport(opts...)
	if transport.MaxIdleConns != 200 ||
		transport.MaxIdleConnsPerHost != 10 ||
		transport.MaxConnsPerHost != 50 ||
		transport.IdleConnTimeout != 30*time.Second ||
		!transport.DisableKeepAlives ||
		!transport.DisableCompression ||
		transport.DialContext == nil ||
		transport.DialTLSContext == nil ||
		transport.TLSClientConfig != tlsConfig ||
		transport.TLSHandshakeTimeout != 5*time.Second ||
		transport.TLSNextProto == nil ||
		transport.ResponseHeaderTimeout != 15*time.Second ||
		transport.ExpectContinueTimeout != 2*time.Second ||
		transport.MaxResponseHeaderBytes != 1<<20 ||
		transport.ReadBufferSize != 8<<10 ||
		transport.WriteBufferSize != 16<<10 ||
		transport.Proxy == nil ||
		transport.ProxyConnectHeader == nil ||
		transport.GetProxyConnectHeader == nil ||
		transport.OnProxyConnectResponse == nil ||
		!transport.ForceAttemptHTTP2 ||
		transport.HTTP2 != http2Config ||
		transport.Protocols != protocols {
		t.Fatalf("transport options were not applied: %+v", transport)
	}

	if len(base) != 0 {
		t.Fatal("configuring a copy mutated the original TransportOpts")
	}
}

func TestStandaloneTransportOptions(t *testing.T) {
	transport := NewTransport(
		WithMaxIdleConns(200),
		WithMaxIdleConnsPerHost(10),
		WithMaxConnsPerHost(50),
		WithIdleConnTimeout(30*time.Second),
	)

	if transport.MaxIdleConns != 200 ||
		transport.MaxIdleConnsPerHost != 10 ||
		transport.MaxConnsPerHost != 50 ||
		transport.IdleConnTimeout != 30*time.Second {
		t.Fatalf("standalone options were not applied: %+v", transport)
	}
}

func TestNewTransportOptsHoldsSuppliedOptions(t *testing.T) {
	opts := NewTransportOpts(WithMaxConnsPerHost(7))
	if len(opts) != 1 {
		t.Fatalf("len(opts) = %d, want 1", len(opts))
	}

	if transport := NewTransport(opts...); transport.MaxConnsPerHost != 7 {
		t.Errorf("MaxConnsPerHost = %d, want 7", transport.MaxConnsPerHost)
	}
}

func TestTransportWithOptionAppliesArbitrarySetting(t *testing.T) {
	opts := NewTransportOpts().WithOption(func(transport *http.Transport) *http.Transport {
		transport.MaxIdleConns = 5
		return transport
	})

	if transport := NewTransport(opts...); transport.MaxIdleConns != 5 {
		t.Errorf("MaxIdleConns = %d, want 5", transport.MaxIdleConns)
	}
}

func TestNewTransportAppliesLaterOptionsLast(t *testing.T) {
	transport := NewTransport(
		WithMaxIdleConnsPerHost(10),
		WithMaxIdleConnsPerHost(20),
	)

	if transport.MaxIdleConnsPerHost != 20 {
		t.Errorf("MaxIdleConnsPerHost = %d, want 20", transport.MaxIdleConnsPerHost)
	}
}

func TestWithTransportOptsBuildsPooledTransport(t *testing.T) {
	client := NewClient(
		WithTimeout(time.Second),
		WithTransportOpts(WithMaxIdleConnsPerHost(10)),
	)

	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("Transport has type %T, want *http.Transport", client.Transport)
	}
	if transport.MaxIdleConnsPerHost != 10 {
		t.Errorf("MaxIdleConnsPerHost = %d, want 10", transport.MaxIdleConnsPerHost)
	}
	// Defaults survive, since NewTransport clones http.DefaultTransport.
	if transport.MaxIdleConns != 100 {
		t.Errorf("MaxIdleConns = %d, want 100", transport.MaxIdleConns)
	}
	if client.Timeout != time.Second {
		t.Errorf("Timeout = %v, want %v", client.Timeout, time.Second)
	}
}

func TestClientOptsWithTransportOpts(t *testing.T) {
	client := NewClient(NewClientOpts().
		WithTimeout(time.Second).
		WithTransportOpts(NewTransportOpts().
			WithMaxConnsPerHost(50)...)...)

	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("Transport has type %T, want *http.Transport", client.Transport)
	}
	if transport.MaxConnsPerHost != 50 {
		t.Errorf("MaxConnsPerHost = %d, want 50", transport.MaxConnsPerHost)
	}
}

func TestPoolReusesConnection(t *testing.T) {
	var (
		mu    sync.Mutex
		conns = map[string]int{}
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	server.Config.ConnState = func(c net.Conn, state http.ConnState) {
		if state == http.StateNew {
			mu.Lock()
			conns[c.RemoteAddr().String()]++
			mu.Unlock()
		}
	}
	defer server.Close()

	client := NewClient(WithTransportOpts(WithMaxIdleConnsPerHost(1)))
	for range 3 {
		resp, err := client.Get(server.URL)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}

	mu.Lock()
	defer mu.Unlock()
	if len(conns) != 1 {
		t.Errorf("dialed %d connections, want 1", len(conns))
	}
}

func TestDisableKeepAlivesDialsEachRequest(t *testing.T) {
	var (
		mu    sync.Mutex
		conns = map[string]int{}
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	server.Config.ConnState = func(c net.Conn, state http.ConnState) {
		if state == http.StateNew {
			mu.Lock()
			conns[c.RemoteAddr().String()]++
			mu.Unlock()
		}
	}
	defer server.Close()

	client := NewClient(WithTransportOpts(WithDisableKeepAlives(true)))
	for range 3 {
		resp, err := client.Get(server.URL)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}

	mu.Lock()
	defer mu.Unlock()
	if len(conns) != 3 {
		t.Errorf("dialed %d connections, want 3", len(conns))
	}
}
