package httputil

import (
	"net"
	"net/http/httptest"
	"testing"
)

func resetTrustedProxies(t *testing.T) {

	t.Helper()

	original := TrustedProxies
	t.Cleanup(func() { TrustedProxies = original })
	TrustedProxies = nil
}

func TestAddTrustedProxy(t *testing.T) {

	resetTrustedProxies(t)

	t.Run("adds valid CIDR", func(t *testing.T) {

		err := AddTrustedProxy("10.0.0.0/8")
		if err != nil {
			t.Fatalf("add: %v", err)
		}
		if len(TrustedProxies) != 1 {
			t.Errorf("want 1 proxy; got %d", len(TrustedProxies))
		}
	})

	t.Run("returns error for invalid CIDR", func(t *testing.T) {

		err := AddTrustedProxy("not-a-cidr")
		if err == nil {
			t.Error("want error for invalid CIDR")
		}
	})
}

func TestIsTrustedProxy(t *testing.T) {

	resetTrustedProxies(t)

	AddTrustedProxy("10.0.0.0/8")
	AddTrustedProxy("172.16.0.0/12")

	var tests = []struct {
		name    string
		ip      string
		trusted bool
	}{
		{"trusted 10.x", "10.1.2.3", true},
		{"trusted 172.16.x", "172.16.5.1", true},
		{"untrusted public", "8.8.8.8", false},
		{"untrusted 192.168", "192.168.1.1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ip := net.ParseIP(tt.ip)
			got := isTrustedProxy(ip)
			if got != tt.trusted {
				t.Errorf("want isTrustedProxy(%s)=%v; got %v", tt.ip, tt.trusted, got)
			}
		})
	}
}

func TestClientIP(t *testing.T) {

	t.Run("returns RemoteAddr when no trusted proxies", func(t *testing.T) {

		resetTrustedProxies(t)

		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "203.0.113.5:12345"

		got := ClientIP(req)
		if got != "203.0.113.5" {
			t.Errorf("want 203.0.113.5; got %s", got)
		}
	})

	t.Run("returns RemoteAddr when not from trusted proxy", func(t *testing.T) {

		resetTrustedProxies(t)
		AddTrustedProxy("10.0.0.0/8")

		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "203.0.113.5:12345"
		req.Header.Set("X-Forwarded-For", "1.2.3.4")

		got := ClientIP(req)
		if got != "203.0.113.5" {
			t.Errorf("want RemoteAddr 203.0.113.5; got %s", got)
		}
	})

	t.Run("parses X-Forwarded-For from trusted proxy", func(t *testing.T) {

		resetTrustedProxies(t)
		AddTrustedProxy("10.0.0.0/8")

		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "10.0.0.1:12345"
		req.Header.Set("X-Forwarded-For", "203.0.113.50, 10.0.0.2")

		got := ClientIP(req)
		if got != "203.0.113.50" {
			t.Errorf("want first untrusted IP 203.0.113.50; got %s", got)
		}
	})

	t.Run("skips trusted IPs in X-Forwarded-For chain", func(t *testing.T) {

		resetTrustedProxies(t)
		AddTrustedProxy("10.0.0.0/8")

		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "10.0.0.1:12345"
		req.Header.Set("X-Forwarded-For", "8.8.8.8, 10.0.0.5, 10.0.0.3")

		got := ClientIP(req)
		if got != "8.8.8.8" {
			t.Errorf("want first untrusted IP 8.8.8.8; got %s", got)
		}
	})

	t.Run("falls back to X-Real-IP when no XFF", func(t *testing.T) {

		resetTrustedProxies(t)
		AddTrustedProxy("10.0.0.0/8")

		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "10.0.0.1:12345"
		req.Header.Set("X-Real-IP", "203.0.113.99")

		got := ClientIP(req)
		if got != "203.0.113.99" {
			t.Errorf("want X-Real-IP 203.0.113.99; got %s", got)
		}
	})

	t.Run("handles RemoteAddr without port", func(t *testing.T) {

		resetTrustedProxies(t)

		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "203.0.113.5"

		got := ClientIP(req)
		if got != "203.0.113.5" {
			t.Errorf("want 203.0.113.5; got %s", got)
		}
	})

	t.Run("handles IPv6 RemoteAddr", func(t *testing.T) {

		resetTrustedProxies(t)

		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "[2001:db8::1]:12345"

		got := ClientIP(req)
		if got != "2001:db8::1" {
			t.Errorf("want 2001:db8::1; got %s", got)
		}
	})
}

func TestNetworkID(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		name     string
		ip       string
		expected string
	}{
		{"IPv4 unchanged", "192.168.1.1", "192.168.1.1"},
		{"IPv4 loopback", "127.0.0.1", "127.0.0.1"},
		{"IPv6 masked to /64", "2001:db8:1234:5678:abcd:ef01:2345:6789", "2001:db8:1234:5678::"},
		{"IPv6 loopback", "::1", "::"},
		{"invalid IP returned as-is", "not-an-ip", "not-an-ip"},
		{"empty string returned as-is", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := NetworkID(tt.ip)
			if got != tt.expected {
				t.Errorf("NetworkID(%q) = %q; want %q", tt.ip, got, tt.expected)
			}
		})
	}
}
