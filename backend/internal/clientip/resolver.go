package clientip

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/textproto"
	"strings"
)

const (
	XFFStrategyTrusted   = "trusted"
	XFFStrategyRightMost = "rightmost"
)

type resolvedIPContextKey struct{}

// Resolver resolves client IPs from request headers and trusted proxy config.
type Resolver struct {
	trustedCIDRs []*net.IPNet
	headers      []string
	xffStrategy  string
}

func NewResolver(trustedProxies []string, headers []string, xffStrategy string) (*Resolver, error) {
	trustedCIDRs, err := parseTrustedCIDRs(trustedProxies)
	if err != nil {
		return nil, err
	}

	normalizedHeaders := normalizeHeaderNames(headers)
	if len(normalizedHeaders) == 0 {
		normalizedHeaders = []string{"X-Forwarded-For", "X-Real-IP"}
	}

	return &Resolver{
		trustedCIDRs: trustedCIDRs,
		headers:      normalizedHeaders,
		xffStrategy:  normalizeXFFStrategy(xffStrategy),
	}, nil
}

func normalizeXFFStrategy(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "", "trusted", "trusted-chain", "standard", "default":
		return XFFStrategyTrusted
	case "rightmost", "right-most", "last":
		return XFFStrategyRightMost
	default:
		return XFFStrategyTrusted
	}
}

func normalizeHeaderNames(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, header := range values {
		canonical := textproto.CanonicalMIMEHeaderKey(strings.TrimSpace(header))
		if canonical == "" {
			continue
		}
		key := strings.ToLower(canonical)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, canonical)
	}
	return result
}

func parseTrustedCIDRs(trustedProxies []string) ([]*net.IPNet, error) {
	result := make([]*net.IPNet, 0, len(trustedProxies))
	for _, proxy := range trustedProxies {
		value := strings.TrimSpace(proxy)
		if value == "" {
			continue
		}
		if _, cidr, err := net.ParseCIDR(value); err == nil {
			result = append(result, cidr)
			continue
		}
		ip := parseIP(value)
		if ip == nil {
			return nil, fmt.Errorf("invalid trusted proxy: %q", value)
		}
		bits := 128
		if ip.To4() != nil {
			bits = 32
		}
		result = append(result, &net.IPNet{
			IP:   ip,
			Mask: net.CIDRMask(bits, bits),
		})
	}
	return result, nil
}

func parseIP(value string) net.IP {
	ip := net.ParseIP(strings.TrimSpace(value))
	if ip == nil {
		return nil
	}
	if ip4 := ip.To4(); ip4 != nil {
		return ip4
	}
	return ip
}

func parseRemoteIP(remoteAddr string) string {
	value := strings.TrimSpace(remoteAddr)
	if value == "" {
		return ""
	}
	host, _, err := net.SplitHostPort(value)
	if err == nil {
		if ip := parseIP(host); ip != nil {
			return ip.String()
		}
	}
	if ip := parseIP(value); ip != nil {
		return ip.String()
	}
	return value
}

func (r *Resolver) isTrustedProxy(ip net.IP) bool {
	if ip == nil || r == nil || len(r.trustedCIDRs) == 0 {
		return false
	}
	for _, cidr := range r.trustedCIDRs {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}

func parseXFFRightMost(header string) (string, bool) {
	if strings.TrimSpace(header) == "" {
		return "", false
	}
	items := strings.Split(header, ",")
	for i := len(items) - 1; i >= 0; i-- {
		ip := parseIP(items[i])
		if ip != nil {
			return ip.String(), true
		}
	}
	return "", false
}

func (r *Resolver) parseTrustedHeader(header string) (string, bool) {
	if strings.TrimSpace(header) == "" {
		return "", false
	}
	items := strings.Split(header, ",")
	for i := len(items) - 1; i >= 0; i-- {
		ip := parseIP(items[i])
		if ip == nil {
			break
		}

		// Keep Gin-compatible trusted-chain semantics.
		if i == 0 || !r.isTrustedProxy(ip) {
			return ip.String(), true
		}
	}
	return "", false
}

func (r *Resolver) resolveHeader(headerName, headerValue string) (string, bool) {
	if strings.EqualFold(headerName, "X-Forwarded-For") && r.xffStrategy == XFFStrategyRightMost {
		return parseXFFRightMost(headerValue)
	}
	return r.parseTrustedHeader(headerValue)
}

func (r *Resolver) Resolve(req *http.Request) string {
	if req == nil {
		return ""
	}
	if ip := FromRequest(req); ip != "" {
		return ip
	}

	remoteIP := parseIP(parseRemoteIP(req.RemoteAddr))
	if remoteIP == nil {
		return ""
	}
	remoteIPStr := remoteIP.String()

	if r == nil || !r.isTrustedProxy(remoteIP) {
		return remoteIPStr
	}

	for _, headerName := range r.headers {
		if ip, ok := r.resolveHeader(headerName, req.Header.Get(headerName)); ok {
			return ip
		}
	}

	return remoteIPStr
}

func WithResolvedIP(req *http.Request, ip string) *http.Request {
	if req == nil {
		return nil
	}
	parsed := parseIP(ip)
	if parsed == nil {
		return req
	}
	ctx := context.WithValue(req.Context(), resolvedIPContextKey{}, parsed.String())
	return req.WithContext(ctx)
}

func FromRequest(req *http.Request) string {
	if req == nil {
		return ""
	}
	value, _ := req.Context().Value(resolvedIPContextKey{}).(string)
	parsed := parseIP(value)
	if parsed == nil {
		return ""
	}
	return parsed.String()
}
