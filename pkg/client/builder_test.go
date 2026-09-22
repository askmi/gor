package client

import (
	"net/http"
	"net/http/cookiejar"
	"testing"
	"time"
)

func TestNewClientAppliesAllOptions(t *testing.T) {
	transport := &http.Transport{}
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookiejar.New() error = %v", err)
	}
	checkRedirect := func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}

	opts := NewClientOpts().
		WithTransport(transport).
		WithCheckRedirect(checkRedirect).
		WithCookieJar(jar).
		WithTimeout(5 * time.Second)
	client := NewClient(opts...)

	if client.Transport != transport {
		t.Error("Transport was not applied")
	}
	if client.CheckRedirect == nil {
		t.Error("CheckRedirect was not applied")
	}
	if client.Jar != jar {
		t.Error("CookieJar was not applied")
	}
	if client.Timeout != 5*time.Second {
		t.Errorf("Timeout = %v, want %v", client.Timeout, 5*time.Second)
	}
}

func TestClientOptsAreIndependent(t *testing.T) {
	base := NewClientOpts()
	fast := base.WithTimeout(time.Second)
	slow := base.WithTimeout(10 * time.Second)

	if timeout := NewClient(base...).Timeout; timeout != 0 {
		t.Errorf("base Timeout = %v, want 0", timeout)
	}
	if timeout := NewClient(fast...).Timeout; timeout != time.Second {
		t.Errorf("fast Timeout = %v, want %v", timeout, time.Second)
	}
	if timeout := NewClient(slow...).Timeout; timeout != 10*time.Second {
		t.Errorf("slow Timeout = %v, want %v", timeout, 10*time.Second)
	}
}

func TestNewClientAppliesLaterOptionsLast(t *testing.T) {
	client := NewClient(
		NewClientOpts().WithTimeout(time.Second).
			WithTimeout(2 * time.Second)...,
	)

	if client.Timeout != 2*time.Second {
		t.Errorf("Timeout = %v, want %v", client.Timeout, 2*time.Second)
	}
}

func TestNewClientOptsHoldsSuppliedOptions(t *testing.T) {
	opts := NewClientOpts(
		WithCheckRedirect(func(*http.Request, []*http.Request) error { return nil }),
	)
	if len(opts) != 1 {
		t.Fatalf("len(opts) = %d, want 1", len(opts))
	}

	if NewClient(opts...).CheckRedirect == nil {
		t.Error("CheckRedirect was not applied")
	}
}

func TestStandaloneOptionsApplyAllSettings(t *testing.T) {
	transport := &http.Transport{}
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookiejar.New() error = %v", err)
	}
	checkRedirect := func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}

	client := NewClient(
		WithTransport(transport),
		WithCheckRedirect(checkRedirect),
		WithCookieJar(jar),
		WithTimeout(5*time.Second),
	)

	if client.Transport != transport {
		t.Error("Transport was not applied")
	}
	if client.CheckRedirect == nil {
		t.Error("CheckRedirect was not applied")
	}
	if client.Jar != jar {
		t.Error("CookieJar was not applied")
	}
	if client.Timeout != 5*time.Second {
		t.Errorf("Timeout = %v, want %v", client.Timeout, 5*time.Second)
	}
}

func TestStandaloneOptionsComposeWithMethods(t *testing.T) {
	checkRedirect := func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}
	transport := &http.Transport{}

	client := NewClient(
		NewClientOpts(WithCheckRedirect(checkRedirect)).
			WithTransport(transport)...,
	)

	if client.CheckRedirect == nil {
		t.Error("standalone WithCheckRedirect was not applied")
	}
	if client.Transport != transport {
		t.Error("WithTransport was not applied")
	}
}

func TestWithOptionAppliesArbitrarySetting(t *testing.T) {
	client := NewClient(NewClientOpts().WithOption(func(c *http.Client) *http.Client {
		c.Timeout = 3 * time.Second
		return c
	})...)

	if client.Timeout != 3*time.Second {
		t.Errorf("Timeout = %v, want %v", client.Timeout, 3*time.Second)
	}
}
