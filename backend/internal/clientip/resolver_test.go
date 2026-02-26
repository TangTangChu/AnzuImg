package clientip

import (
	"net/http/httptest"
	"testing"
)

func TestResolverResolveTrustedStrategy(t *testing.T) {
	resolver, err := NewResolver(
		[]string{"172.29.87.2/32"},
		[]string{"X-Forwarded-For"},
		"trusted",
	)
	if err != nil {
		t.Fatalf("new resolver: %v", err)
	}

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.RemoteAddr = "172.29.87.2:12345"
	req.Header.Set("X-Forwarded-For", "1.2.3.4, 5.6.7.8, 183.216.228.110")

	got := resolver.Resolve(req)
	want := "183.216.228.110"
	if got != want {
		t.Fatalf("resolve trusted mode mismatch: got %q, want %q", got, want)
	}
}

func TestResolverResolveRightMostStrategy(t *testing.T) {
	resolver, err := NewResolver(
		[]string{"172.29.87.2/32"},
		[]string{"X-Forwarded-For"},
		"rightmost",
	)
	if err != nil {
		t.Fatalf("new resolver: %v", err)
	}

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.RemoteAddr = "172.29.87.2:12345"
	req.Header.Set("X-Forwarded-For", "1.2.3.4, 5.6.7.8, 183.216.228.110")

	got := resolver.Resolve(req)
	want := "183.216.228.110"
	if got != want {
		t.Fatalf("resolve rightmost mode mismatch: got %q, want %q", got, want)
	}
}

func TestResolverStrategyDifferenceOnXFF(t *testing.T) {
	headers := []string{"X-Forwarded-For"}
	trusted := []string{"172.29.87.2/32", "183.216.228.110/32"}

	trustedResolver, err := NewResolver(trusted, headers, "trusted")
	if err != nil {
		t.Fatalf("new trusted resolver: %v", err)
	}
	rightMostResolver, err := NewResolver(trusted, headers, "rightmost")
	if err != nil {
		t.Fatalf("new rightmost resolver: %v", err)
	}

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.RemoteAddr = "172.29.87.2:12345"
	req.Header.Set("X-Forwarded-For", "1.2.3.4, 5.6.7.8, 183.216.228.110")

	trustedIP := trustedResolver.Resolve(req)
	rightMostIP := rightMostResolver.Resolve(req)

	if trustedIP != "5.6.7.8" {
		t.Fatalf("trusted mode mismatch: got %q, want %q", trustedIP, "5.6.7.8")
	}
	if rightMostIP != "183.216.228.110" {
		t.Fatalf("rightmost mode mismatch: got %q, want %q", rightMostIP, "183.216.228.110")
	}
}

func TestResolverIgnoreHeadersWhenRemoteIsUntrusted(t *testing.T) {
	resolver, err := NewResolver(
		[]string{"172.29.87.2/32"},
		[]string{"X-Forwarded-For"},
		"rightmost",
	)
	if err != nil {
		t.Fatalf("new resolver: %v", err)
	}

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.RemoteAddr = "10.10.10.10:12345"
	req.Header.Set("X-Forwarded-For", "1.2.3.4, 5.6.7.8, 183.216.228.110")

	got := resolver.Resolve(req)
	want := "10.10.10.10"
	if got != want {
		t.Fatalf("resolve with untrusted remote mismatch: got %q, want %q", got, want)
	}
}

func TestRequestContextRoundTrip(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.RemoteAddr = "172.29.87.2:12345"

	req = WithResolvedIP(req, "183.216.228.110")
	got := FromRequest(req)
	want := "183.216.228.110"
	if got != want {
		t.Fatalf("context ip mismatch: got %q, want %q", got, want)
	}
}
