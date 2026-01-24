package middleware

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/TangTangChu/AnzuImg/backend/internal/service"
)

// Session 创建会话中间件
func Session(db *gorm.DB) gin.HandlerFunc {
	sessionService := service.NewSessionService(db)
	return sessionService.SessionMiddleware()
}
