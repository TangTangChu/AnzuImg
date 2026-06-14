package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/TangTangChu/AnzuImg/backend/internal/http/response"
	"github.com/TangTangChu/AnzuImg/backend/internal/logger"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := strings.TrimSpace(c.GetHeader(response.HeaderRequestID))
		if requestID == "" {
			requestID = generateRequestID()
		}

		c.Set(response.CtxRequestIDKey, requestID)
		c.Header(response.HeaderRequestID, requestID)

		ctx := logger.ContextWithFields(c.Request.Context(), logger.RequestFields{
			RequestID: requestID,
			IPAddress: ClientIP(c),
		})
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

func generateRequestID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(buf)
}
