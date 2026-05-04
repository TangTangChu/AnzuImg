package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/TangTangChu/AnzuImg/backend/internal/config"
	"github.com/TangTangChu/AnzuImg/backend/internal/http/response"
	"github.com/TangTangChu/AnzuImg/backend/internal/service"
)

func Session(cfg *config.Config, db *gorm.DB) gin.HandlerFunc {
	sessionService := service.NewSessionService(cfg, db)
	return sessionService.SessionMiddleware()
}

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
