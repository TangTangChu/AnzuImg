package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/TangTangChu/AnzuImg/backend/internal/http/response"
)

// CORS 接受一个返回当前允许 Origin 列表的回调，使其在每次请求时读取最新值。
// 这样 Web 端修改 ALLOWED_ORIGINS 后无需重启即可生效。
func CORS(originsFn func() []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		allowed := false
		var origins []string
		if originsFn != nil {
			origins = originsFn()
		}
		for _, allowedOrigin := range origins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}

		if !allowed && origin != "" {
			response.AbortErrorCode(c, http.StatusForbidden, "cors_origin_not_allowed", "origin is not allowed")
			return
		}

		if allowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Accept, Origin, Cache-Control, X-Requested-With, X-Session-Data, X-Session-ID")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func ImageCORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
