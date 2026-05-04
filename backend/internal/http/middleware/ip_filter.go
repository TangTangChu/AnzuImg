package middleware

import (
	"net/http"
	"net/netip"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/TangTangChu/AnzuImg/backend/internal/http/response"
)

type ipFilterCache struct {
	mu       sync.RWMutex
	rawList  []string
	prefixes []netip.Prefix
	addrs    []netip.Addr
}

func (c *ipFilterCache) ensure(list []string) {
	c.mu.RLock()
	if equalSlice(c.rawList, list) {
		c.mu.RUnlock()
		return
	}
	c.mu.RUnlock()
	c.mu.Lock()
	defer c.mu.Unlock()
	if equalSlice(c.rawList, list) {
		return
	}
	c.rawList = append(c.rawList[:0:0], list...)
	c.prefixes = c.prefixes[:0]
	c.addrs = c.addrs[:0]
	for _, item := range list {
		entry := strings.TrimSpace(item)
		if entry == "" {
			continue
		}
		if strings.Contains(entry, "/") {
			if pfx, err := netip.ParsePrefix(entry); err == nil {
				c.prefixes = append(c.prefixes, pfx)
			}
			continue
		}
		if addr, err := netip.ParseAddr(entry); err == nil {
			c.addrs = append(c.addrs, addr)
		}
	}
}

func (c *ipFilterCache) match(ip string) bool {
	addr, err := netip.ParseAddr(strings.TrimSpace(ip))
	if err != nil {
		return false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, pfx := range c.prefixes {
		if pfx.Contains(addr) {
			return true
		}
	}
	for _, a := range c.addrs {
		if a == addr {
			return true
		}
	}
	return false
}

func equalSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// IPBlacklist 全局黑名单中间件,命中即 403。
// listFn 在每次请求时返回最新列表,允许 Web 端热修改。
func IPBlacklist(listFn func() []string) gin.HandlerFunc {
	cache := &ipFilterCache{}
	return func(c *gin.Context) {
		var list []string
		if listFn != nil {
			list = listFn()
		}
		if len(list) == 0 {
			c.Next()
			return
		}
		cache.ensure(list)
		ip := ClientIP(c)
		if ip == "" {
			c.Next()
			return
		}
		if cache.match(ip) {
			response.AbortErrorCode(c, http.StatusForbidden, "ip_blacklisted", "your IP is blocked")
			return
		}
		c.Next()
	}
}

// AdminIPAllowlist 管理面板白名单,空列表表示不限。
// 仅挂在敏感路由上以减小爆破面。
func AdminIPAllowlist(listFn func() []string) gin.HandlerFunc {
	cache := &ipFilterCache{}
	return func(c *gin.Context) {
		var list []string
		if listFn != nil {
			list = listFn()
		}
		if len(list) == 0 {
			c.Next()
			return
		}
		cache.ensure(list)
		ip := ClientIP(c)
		if ip == "" || !cache.match(ip) {
			response.AbortErrorCode(c, http.StatusForbidden, "admin_ip_denied", "admin access not allowed from this IP")
			return
		}
		c.Next()
	}
}
