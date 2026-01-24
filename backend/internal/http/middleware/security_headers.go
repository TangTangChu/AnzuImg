package middleware

import "github.com/gin-gonic/gin"

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		
		csp := "default-src 'self'; img-src 'self' data: blob:; style-src 'self' 'unsafe-inline'; script-src 'self';"
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
