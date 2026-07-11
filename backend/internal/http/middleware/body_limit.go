package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const DefaultJSONBodyLimit int64 = 1 << 20

// JSONBodyLimit caps JSON request bodies before handlers attempt to decode them.
// Multipart uploads have their own larger, explicit limits in ImageHandler.
func JSONBodyLimit(maxBytes int64) gin.HandlerFunc {
	if maxBytes <= 0 {
		maxBytes = DefaultJSONBodyLimit
	}
	return func(c *gin.Context) {
		contentType := strings.ToLower(c.GetHeader("Content-Type"))
		if strings.Contains(contentType, "json") && c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}
		c.Next()
	}
}
