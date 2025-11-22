package http

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/TangTangChu/AnzuImg/backend/internal/config"
	"github.com/TangTangChu/AnzuImg/backend/internal/http/handler"
	"github.com/TangTangChu/AnzuImg/backend/internal/http/middleware"
)

func NewRouter(cfg *config.Config, db *gorm.DB) *gin.Engine {
	r := gin.Default()

	healthH := handler.NewHealthHandler()
	imageH := handler.NewImageHandler(cfg, db)

	registerHealthRoutes(r, healthH)
	registerPublicImageRoutes(r, imageH)
	registerAPIRoutes(r, cfg, healthH, imageH)

	return r
}

func registerHealthRoutes(r *gin.Engine, h *handler.HealthHandler) {
	r.GET("/health", h.Health)
}

func registerPublicImageRoutes(r *gin.Engine, h *handler.ImageHandler) {
	r.GET("/i/:hash", h.GetByHash)
	r.GET("/i/r/:route", h.GetByRoute)
}

func registerAPIRoutes(r *gin.Engine, cfg *config.Config, hh *handler.HealthHandler, ih *handler.ImageHandler) {
	api := r.Group("/api/v1", middleware.APIKey(cfg))
	{
		api.GET("/ping", hh.Ping)
		api.POST("/images", ih.Upload)
	}
}
