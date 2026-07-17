package httpapi

import (
	"net"
	"net/http"
	"net/netip"
	"strings"
)

const forwardedForHeader = "X-Forwarded-For"

type clientIPResolver struct {
	trustedProxies []netip.Prefix
}

func newClientIPResolver(trustedProxies []netip.Prefix) clientIPResolver {
	return clientIPResolver{
		trustedProxies: append([]netip.Prefix(nil), trustedProxies...),
	}
}

func (resolver clientIPResolver) resolve(r *http.Request) string {
	clientKey, peer, ok := socketClient(r)
	if !ok || !resolver.isTrusted(peer) {
		return clientKey
	}

	forwarded, ok := forwardedFor(r)
	if !ok || len(forwarded) == 0 {
		return clientKey
	}

	for i := len(forwarded) - 1; i >= 0; i-- {
		if !resolver.isTrusted(forwarded[i]) {
			return forwarded[i].String()
		}
	}

	return forwarded[0].String()
}

func (resolver clientIPResolver) isTrusted(address netip.Addr) bool {
	for _, prefix := range resolver.trustedProxies {
		if prefix.Contains(address) {
			return true
		}
	}

	return false
}

func socketClient(r *http.Request) (string, netip.Addr, bool) {
	if r == nil {
		return "", netip.Addr{}, false
	}

	address := strings.TrimSpace(r.RemoteAddr)
	host, _, err := net.SplitHostPort(address)
	if err == nil {
		address = host
	}

	parsed, err := netip.ParseAddr(address)
	if err != nil {
		return address, netip.Addr{}, false
	}

	parsed = parsed.Unmap().WithZone("")
	return parsed.String(), parsed, true
}

func forwardedFor(r *http.Request) ([]netip.Addr, bool) {
	if r == nil {
		return nil, false
	}

	var addresses []netip.Addr
	for _, value := range r.Header.Values(forwardedForHeader) {
		for _, item := range strings.Split(value, ",") {
			parsed, err := netip.ParseAddr(strings.TrimSpace(item))
			if err != nil {
				return nil, false
			}

			addresses = append(addresses, parsed.Unmap().WithZone(""))
		}
	}

	return addresses, true
}
