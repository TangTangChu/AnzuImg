package middleware

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/TangTangChu/AnzuImg/backend/internal/clientip"
)

// AccessLogger records request metadata without the query string. Query values
// frequently contain one-time codes or tokens and must never enter access logs.
func AccessLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(formatAccessLog)
}

func formatAccessLog(p gin.LogFormatterParams) string {
	path := "/"
	clientIP := p.ClientIP
	if p.Request != nil {
		if ip := clientip.FromRequest(p.Request); ip != "" {
			clientIP = ip
		}
		if p.Request.URL != nil && p.Request.URL.EscapedPath() != "" {
			path = p.Request.URL.EscapedPath()
		}
	}

	return fmt.Sprintf("[GIN] %s | %3d | %13v | %15s | %-7s %s\n",
		p.TimeStamp.Format(time.RFC3339),
		p.StatusCode,
		p.Latency,
		clientIP,
		p.Method,
		strconv.Quote(path),
	)
}
