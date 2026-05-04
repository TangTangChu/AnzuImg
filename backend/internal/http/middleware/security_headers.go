package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// SecurityHeaders 在每次请求时根据当前 effective 配置追加 CSP 片段。
// cspExtraFn 返回额外的 CSP directive,用分号分隔,允许 Web 端调整。
func SecurityHeaders(cspExtraFn func() string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		csp := "default-src 'self'; img-src 'self' data: blob:; style-src 'self' 'unsafe-inline'; script-src 'self';"
		if cspExtraFn != nil {
			extra := strings.TrimSpace(cspExtraFn())
			if extra != "" {
				if !strings.HasSuffix(csp, ";") {
					csp += ";"
				}
				csp += " " + extra
			}
		}
		c.Header("Content-Security-Policy", csp)

		c.Next()
	}
}

func ImageSecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")

		csp := "default-src 'none'; img-src * data: blob:; style-src 'unsafe-inline'; frame-ancestors 'none';"
		c.Header("Content-Security-Policy", csp)

		c.Next()
	}
}
