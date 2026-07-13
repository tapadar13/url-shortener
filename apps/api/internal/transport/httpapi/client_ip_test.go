package httpapi

import (
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"
)

func TestClientIPResolver(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		remoteAddr       string
		forwardedFor     []string
		trustedProxies   []netip.Prefix
		expectedClientIP string
	}{
		{
			name:             "ignores forwarded chain from untrusted peer",
			remoteAddr:       "203.0.113.10:4321",
			forwardedFor:     []string{"198.51.100.20"},
			trustedProxies:   prefixes("10.0.0.0/8"),
			expectedClientIP: "203.0.113.10",
		},
		{
			name:             "uses forwarded client from trusted peer",
			remoteAddr:       "10.0.0.5:4321",
			forwardedFor:     []string{"198.51.100.20"},
			trustedProxies:   prefixes("10.0.0.0/8"),
			expectedClientIP: "198.51.100.20",
		},
		{
			name:             "walks trusted proxy chain from right to left",
			remoteAddr:       "10.0.0.5:4321",
			forwardedFor:     []string{"198.51.100.20, 192.168.1.8"},
			trustedProxies:   prefixes("10.0.0.0/8", "192.168.0.0/16"),
			expectedClientIP: "198.51.100.20",
		},
		{
			name:             "does not trust spoofed values left of untrusted proxy",
			remoteAddr:       "10.0.0.5:4321",
			forwardedFor:     []string{"198.51.100.20, 203.0.113.8"},
			trustedProxies:   prefixes("10.0.0.0/8"),
			expectedClientIP: "203.0.113.8",
		},
		{
			name:             "falls back to socket peer for malformed chain",
			remoteAddr:       "10.0.0.5:4321",
			forwardedFor:     []string{"198.51.100.20, invalid"},
			trustedProxies:   prefixes("10.0.0.0/8"),
			expectedClientIP: "10.0.0.5",
		},
		{
			name:             "supports multiple header lines and IPv6",
			remoteAddr:       "[2001:db8:ffff::5]:4321",
			forwardedFor:     []string{"2001:db8:1::20", "2001:db8:2::8"},
			trustedProxies:   prefixes("2001:db8:ffff::/48", "2001:db8:2::/48"),
			expectedClientIP: "2001:db8:1::20",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			request := httptest.NewRequest(http.MethodGet, "/AbC123", nil)
			request.RemoteAddr = tt.remoteAddr
			for _, value := range tt.forwardedFor {
				request.Header.Add(forwardedForHeader, value)
			}

			actual := newClientIPResolver(tt.trustedProxies).resolve(request)
			if actual != tt.expectedClientIP {
				t.Fatalf("expected client IP %q, got %q", tt.expectedClientIP, actual)
			}
		})
	}
}

func prefixes(values ...string) []netip.Prefix {
	prefixes := make([]netip.Prefix, len(values))
	for i, value := range values {
		prefixes[i] = netip.MustParsePrefix(value)
	}

	return prefixes
}
