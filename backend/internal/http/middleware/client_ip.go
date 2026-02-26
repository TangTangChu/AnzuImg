package middleware

import (
	"github.com/TangTangChu/AnzuImg/backend/internal/clientip"

	"github.com/gin-gonic/gin"
)

func ClientIPMiddleware(resolver *clientip.Resolver) gin.HandlerFunc {
	return func(c *gin.Context) {
		if resolver != nil {
			if ip := resolver.Resolve(c.Request); ip != "" {
				c.Request = clientip.WithResolvedIP(c.Request, ip)
			}
		}
		c.Next()
	}
}

func ClientIP(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if ip := clientip.FromRequest(c.Request); ip != "" {
		return ip
	}
	return c.ClientIP()
}
