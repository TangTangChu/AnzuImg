package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/TangTangChu/AnzuImg/backend/internal/http/response"
	"github.com/TangTangChu/AnzuImg/backend/internal/service"
)

// Session 创建会话中间件
func Session(db *gorm.DB) gin.HandlerFunc {
	sessionService := service.NewSessionService(db)
	return sessionService.SessionMiddleware()
}

// RequireSession 仅允许 session 认证通过的请求访问
func RequireSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		authMethod, _ := c.Get("auth_method")
		if authMethod != "session" {
			response.AbortErrorCode(c, http.StatusForbidden, "session_required", "session authentication required")
			return
		}
		c.Next()
	}
}
