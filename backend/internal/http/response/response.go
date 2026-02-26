package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	CtxRequestIDKey = "request_id"
	HeaderRequestID = "X-Request-ID"
)

type ErrorResponse struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

func WriteError(c *gin.Context, status int, message string) {
	WriteErrorCode(c, status, defaultCode(status), message)
}

func WriteErrorCode(c *gin.Context, status int, code, message string) {
	requestID, _ := c.Get(CtxRequestIDKey)
	requestIDStr, _ := requestID.(string)
	c.JSON(status, ErrorResponse{
		Code:      code,
		Message:   message,
		RequestID: requestIDStr,
	})
}

func AbortError(c *gin.Context, status int, message string) {
	AbortErrorCode(c, status, defaultCode(status), message)
}

func AbortErrorCode(c *gin.Context, status int, code, message string) {
	requestID, _ := c.Get(CtxRequestIDKey)
	requestIDStr, _ := requestID.(string)
	c.AbortWithStatusJSON(status, ErrorResponse{
		Code:      code,
		Message:   message,
		RequestID: requestIDStr,
	})
}

func defaultCode(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "bad_request"
	case http.StatusUnauthorized:
		return "unauthorized"
	case http.StatusForbidden:
		return "forbidden"
	case http.StatusNotFound:
		return "not_found"
	case http.StatusTooManyRequests:
		return "too_many_requests"
	case http.StatusServiceUnavailable:
		return "service_unavailable"
	default:
		if status >= 500 {
			return "internal_error"
		}
		return "request_error"
	}
}
