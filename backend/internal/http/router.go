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
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.CORS(cfg.AllowedOrigins))
	healthH := handler.NewHealthHandler()
	imageH := handler.NewImageHandler(cfg, db)
	authH := handler.NewAuthHandler(cfg, db)
	apiTokenH := handler.NewAPITokenHandler(db)

	registerHealthRoutes(r, healthH)
	registerPublicImageRoutes(r, imageH)
	registerAuthRoutes(r, authH, apiTokenH)
	registerAPIRoutes(r, healthH, imageH, authH)

	return r
}

func registerHealthRoutes(r *gin.Engine, h *handler.HealthHandler) {
	r.GET("/health", h.Health)
}

func registerPublicImageRoutes(r *gin.Engine, h *handler.ImageHandler) {
	imageRoutes := r.Group("/i")
	imageRoutes.Use(middleware.ImageCORS())
	imageRoutes.Use(middleware.ImageSecurityHeaders())
	{
		imageRoutes.GET("/:hash", h.GetByHash)
		imageRoutes.GET("/:hash/thumbnail", h.GetThumbnailByHash)
		imageRoutes.GET("/r/:route", h.GetByRoute)
	}
}

func registerAuthRoutes(r *gin.Engine, h *handler.AuthHandler, tokenH *handler.APITokenHandler) {
	auth := r.Group("/api/v1/auth")
	{
		auth.GET("/status", h.CheckInit)
		auth.POST("/setup", h.Setup)
		auth.POST("/login", h.AuthWithPassword)
		auth.GET("/validate", h.ValidateSession)

		// Passkey Login
		auth.GET("/passkey/login/begin", h.LoginPasskeyBegin)
		auth.POST("/passkey/login/finish", h.LoginPasskeyFinish)
	}

	// Protected auth routes
	protectedAuth := auth.Group("", middleware.Session(h.DB()))
	{
		// Passkey Registration
		protectedAuth.GET("/passkey/register/begin", h.RegisterPasskeyBegin)
		protectedAuth.POST("/passkey/register/finish", h.RegisterPasskeyFinish)

		// Passkey Management
		protectedAuth.GET("/passkeys", h.ListPasskeys)
		protectedAuth.GET("/passkeys/count", h.GetPasskeyCount)
		protectedAuth.GET("/passkeys/check", h.CheckPasskeyExists)
		protectedAuth.DELETE("/passkeys/:credential_id", h.DeletePasskey)

		// Password Management
		protectedAuth.POST("/change-password", h.ChangePassword)

		// API Token Management
		protectedAuth.POST("/tokens", tokenH.Create)
		protectedAuth.GET("/tokens", tokenH.List)
		protectedAuth.DELETE("/tokens/:id", tokenH.Delete)
	}
}

func registerAPIRoutes(r *gin.Engine, hh *handler.HealthHandler, ih *handler.ImageHandler, ah *handler.AuthHandler) {
	api := r.Group("/api/v1", middleware.Session(ah.DB()))
	{
		api.GET("/ping", hh.Ping)
		api.POST("/images", ih.Upload)
		api.GET("/images", ih.List)
		api.GET("/images/:hash/info", ih.GetInfo)
		api.DELETE("/images/:hash", ih.Delete)
		api.PATCH("/images/:hash", ih.Update)
		api.GET("/routes", ih.ListRoutes)
		api.DELETE("/routes/:route", ih.DeleteRoute)
	}
}
