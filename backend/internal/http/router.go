package http

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/TangTangChu/AnzuImg/backend/internal/clientip"
	"github.com/TangTangChu/AnzuImg/backend/internal/config"
	"github.com/TangTangChu/AnzuImg/backend/internal/http/handler"
	"github.com/TangTangChu/AnzuImg/backend/internal/http/middleware"
	"github.com/TangTangChu/AnzuImg/backend/internal/model"
)

func NewRouter(cfg *config.Config, db *gorm.DB) (*gin.Engine, error) {
	r := gin.Default()

	resolver, err := clientip.NewResolver(cfg.TrustedProxies, cfg.ClientIPHeaders, cfg.ClientIPXFFStrategy)
	if err != nil {
		return nil, fmt.Errorf("init client ip resolver failed: %w", err)
	}

	r.Use(middleware.ClientIPMiddleware(resolver))
	r.Use(middleware.RequestID())
	r.Use(middleware.SecurityHeaders())
	r.Use(func(c *gin.Context) {
		c.Set("cookie_samesite", cfg.CookieSameSite)
		c.Set("strict_session_ip", cfg.StrictSessionIP)
		c.Next()
	})
	healthH := handler.NewHealthHandler()
	imageH := handler.NewImageHandler(cfg, db)
	authH := handler.NewAuthHandler(cfg, db)
	apiTokenH := handler.NewAPITokenHandler(db)

	registerHealthRoutes(r, healthH)
	registerPublicImageRoutes(r, imageH)
	registerAuthRoutes(r, cfg, authH, apiTokenH)
	registerAPIRoutes(r, cfg, healthH, imageH, authH)

	return r, nil
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

func registerAuthRoutes(r *gin.Engine, cfg *config.Config, h *handler.AuthHandler, tokenH *handler.APITokenHandler) {
	apiPrefix := cfg.APIPrefix + "/api/v1"
	auth := r.Group(apiPrefix+"/auth", middleware.CORS(cfg.AllowedOrigins))
	{
		auth.OPTIONS("/*path", func(c *gin.Context) { c.Status(204) })
		auth.GET("/status", h.CheckInit)
		auth.POST("/setup", h.Setup)
		auth.POST("/login", h.AuthWithPassword)
		auth.POST("/logout", h.Logout)
		auth.GET("/validate", h.ValidateSession)

		// Passkey Login
		auth.GET("/passkey/login/begin", h.LoginPasskeyBegin)
		auth.POST("/passkey/login/finish", h.LoginPasskeyFinish)
	}

	// Protected auth routes
	protectedAuth := auth.Group("", middleware.Session(h.DB()), middleware.RequireSession())
	{
		// Passkey Registration
		protectedAuth.GET("/passkey/register/begin", h.RegisterPasskeyBegin)
		protectedAuth.POST("/passkey/register/finish", h.RegisterPasskeyFinish)

		// Passkey Management
		protectedAuth.GET("/passkeys", h.ListPasskeys)
		protectedAuth.GET("/passkeys/count", h.GetPasskeyCount)
		protectedAuth.GET("/passkeys/check", h.CheckPasskeyExists)
		protectedAuth.DELETE("/passkeys/:credential_id", h.DeletePasskey)
		protectedAuth.POST("/passkeys/:credential_id/delete", h.DeletePasskey)

		// Password Management
		protectedAuth.POST("/change-password", h.ChangePassword)
		protectedAuth.GET("/security/logs", h.ListSecurityLogs)

		// API Token Management
		protectedAuth.POST("/tokens", tokenH.Create)
		protectedAuth.GET("/tokens", tokenH.List)
		protectedAuth.GET("/tokens/logs", tokenH.ListLogs)
		protectedAuth.DELETE("/tokens/logs", tokenH.CleanupLogs)
		protectedAuth.POST("/tokens/logs/cleanup", tokenH.CleanupLogs)
		protectedAuth.DELETE("/tokens/:id", tokenH.Delete)
		protectedAuth.POST("/tokens/:id/delete", tokenH.Delete)
	}
}

func registerAPIRoutes(r *gin.Engine, cfg *config.Config, hh *handler.HealthHandler, ih *handler.ImageHandler, ah *handler.AuthHandler) {
	apiPrefix := cfg.APIPrefix + "/api/v1"
	api := r.Group(apiPrefix, middleware.CORS(cfg.AllowedOrigins), middleware.Session(ah.DB()))
	{
		api.GET("/ping", middleware.RequireTokenType(model.TokenTypeFull), hh.Ping)
		api.POST("/images", middleware.RequireTokenScopes(model.ScopeImagesUpload), ih.Upload)
		api.GET("/images", middleware.RequireTokenScopes(model.ScopeImagesList), ih.List)
		api.GET("/tags", middleware.RequireTokenType(model.TokenTypeFull), ih.ListTags)
		api.GET("/images/:hash/info", middleware.RequireTokenType(model.TokenTypeFull), ih.GetInfo)
		api.DELETE("/images/:hash", middleware.RequireTokenType(model.TokenTypeFull), ih.Delete)
		api.POST("/images/:hash/delete", middleware.RequireTokenType(model.TokenTypeFull), ih.Delete)
		api.PATCH("/images/:hash", middleware.RequireTokenType(model.TokenTypeFull), ih.Update)
		api.GET("/routes", middleware.RequireTokenType(model.TokenTypeFull), ih.ListRoutes)
		api.DELETE("/routes/:route", middleware.RequireTokenType(model.TokenTypeFull), ih.DeleteRoute)
		api.POST("/routes/:route/delete", middleware.RequireTokenType(model.TokenTypeFull), ih.DeleteRoute)
		api.GET("/stats", middleware.RequireTokenType(model.TokenTypeFull), ih.GetStats)

		api.OPTIONS("/ping", func(c *gin.Context) { c.Status(204) })
		api.OPTIONS("/images", func(c *gin.Context) { c.Status(204) })
		api.OPTIONS("/tags", func(c *gin.Context) { c.Status(204) })
		api.OPTIONS("/images/:hash/info", func(c *gin.Context) { c.Status(204) })
		api.OPTIONS("/images/:hash", func(c *gin.Context) { c.Status(204) })
		api.OPTIONS("/routes", func(c *gin.Context) { c.Status(204) })
		api.OPTIONS("/routes/:route", func(c *gin.Context) { c.Status(204) })
	}
}
