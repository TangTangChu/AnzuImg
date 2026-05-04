package http

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/TangTangChu/AnzuImg/backend/internal/clientip"
	"github.com/TangTangChu/AnzuImg/backend/internal/config"
	"github.com/TangTangChu/AnzuImg/backend/internal/http/handler"
	"github.com/TangTangChu/AnzuImg/backend/internal/http/middleware"
	"github.com/TangTangChu/AnzuImg/backend/internal/model"
	"github.com/TangTangChu/AnzuImg/backend/internal/service"
)

func NewRouter(cfg *config.Config, db *gorm.DB, settings *service.SettingsService, hub *service.LogStreamHub) (*gin.Engine, error) {
	r := gin.Default()

	resolver, err := clientip.NewResolver(cfg.TrustedProxies, cfg.ClientIPHeaders, cfg.ClientIPXFFStrategy)
	if err != nil {
		return nil, fmt.Errorf("init client ip resolver failed: %w", err)
	}

	originsFn := func() []string { return cfg.Effective().AllowedOrigins }
	cspExtraFn := func() string { return cfg.Effective().CSPExtra }
	blacklistFn := func() []string { return cfg.Effective().IPBlacklist }
	adminAllowlistFn := func() []string { return cfg.Effective().AdminIPAllowlist }
	stepUpAgeFn := func() time.Duration {
		return time.Duration(cfg.Effective().StepUpMaxAgeSeconds) * time.Second
	}

	r.Use(middleware.ClientIPMiddleware(resolver))
	r.Use(middleware.IPBlacklist(blacklistFn))
	r.Use(middleware.RequestID())
	r.Use(middleware.SecurityHeaders(cspExtraFn))

	healthH := handler.NewHealthHandler()
	imageH := handler.NewImageHandler(cfg, db)
	authH := handler.NewAuthHandler(cfg, db)
	apiTokenH := handler.NewAPITokenHandler(db)
	settingsH := handler.NewSettingsHandler(cfg, db, settings)
	logH := handler.NewLogHandler(db, hub)

	registerHealthRoutes(r, healthH)
	registerPublicImageRoutes(r, imageH)
	registerAuthRoutes(r, cfg, authH, apiTokenH, settingsH, logH, originsFn, adminAllowlistFn, stepUpAgeFn)
	registerAPIRoutes(r, cfg, healthH, imageH, authH, originsFn)

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

func registerAuthRoutes(
	r *gin.Engine,
	cfg *config.Config,
	h *handler.AuthHandler,
	tokenH *handler.APITokenHandler,
	settingsH *handler.SettingsHandler,
	logH *handler.LogHandler,
	originsFn func() []string,
	adminAllowlistFn func() []string,
	stepUpAgeFn func() time.Duration,
) {
	apiPrefix := cfg.APIPrefix + "/api/v1"
	auth := r.Group(apiPrefix+"/auth", middleware.CORS(originsFn))
	{
		auth.OPTIONS("/*path", func(c *gin.Context) { c.Status(204) })
		auth.GET("/status", h.CheckInit)
		auth.POST("/setup", h.Setup)
		auth.POST("/login", h.AuthWithPassword)
		auth.POST("/logout", h.Logout)
		auth.GET("/validate", h.ValidateSession)

		auth.GET("/passkey/login/begin", h.LoginPasskeyBegin)
		auth.POST("/passkey/login/finish", h.LoginPasskeyFinish)
	}

	protectedAuth := auth.Group("", middleware.Session(cfg, h.DB()), middleware.RequireSession(), middleware.AdminIPAllowlist(adminAllowlistFn))
	{
		protectedAuth.GET("/passkey/register/begin", h.RegisterPasskeyBegin)
		protectedAuth.POST("/passkey/register/finish", h.RegisterPasskeyFinish)

		protectedAuth.GET("/passkeys", h.ListPasskeys)
		protectedAuth.GET("/passkeys/count", h.GetPasskeyCount)
		protectedAuth.GET("/passkeys/check", h.CheckPasskeyExists)
		protectedAuth.GET("/security/logs", h.ListSecurityLogs)
		protectedAuth.GET("/tokens", tokenH.List)
		protectedAuth.GET("/tokens/logs", tokenH.ListLogs)

		protectedAuth.POST("/step-up/password", h.StepUpWithPassword)
		protectedAuth.GET("/step-up/passkey/begin", h.StepUpPasskeyBegin)
		protectedAuth.POST("/step-up/passkey/finish", h.StepUpPasskeyFinish)
	}

	stepUp := middleware.RequireStepUp(stepUpAgeFn)
	sensitiveAuth := auth.Group("",
		middleware.Session(cfg, h.DB()),
		middleware.RequireSession(),
		middleware.AdminIPAllowlist(adminAllowlistFn),
		stepUp,
	)
	{
		sensitiveAuth.DELETE("/passkeys/:credential_id", h.DeletePasskey)
		sensitiveAuth.POST("/passkeys/:credential_id/delete", h.DeletePasskey)
		sensitiveAuth.POST("/change-password", h.ChangePassword)
		sensitiveAuth.POST("/tokens", tokenH.Create)
		sensitiveAuth.DELETE("/tokens/logs", tokenH.CleanupLogs)
		sensitiveAuth.POST("/tokens/logs/cleanup", tokenH.CleanupLogs)
		sensitiveAuth.DELETE("/tokens/:id", tokenH.Delete)
		sensitiveAuth.POST("/tokens/:id/delete", tokenH.Delete)
	}

	settingsGrp := r.Group(apiPrefix+"/settings",
		middleware.CORS(originsFn),
		middleware.Session(cfg, h.DB()),
		middleware.RequireSession(),
		middleware.AdminIPAllowlist(adminAllowlistFn),
	)
	{
		settingsGrp.OPTIONS("/*path", func(c *gin.Context) { c.Status(204) })
		settingsGrp.GET("", settingsH.Get)
		settingsGrp.PATCH("", stepUp, settingsH.Patch)
		settingsGrp.POST("/reset", stepUp, settingsH.Reset)
	}

	logsGrp := r.Group(apiPrefix+"/logs",
		middleware.CORS(originsFn),
		middleware.Session(cfg, h.DB()),
		middleware.RequireSession(),
		middleware.AdminIPAllowlist(adminAllowlistFn),
	)
	{
		logsGrp.OPTIONS("/*path", func(c *gin.Context) { c.Status(204) })
		logsGrp.GET("/app", logH.ListApp)
		logsGrp.GET("/security", logH.ListSecurity)
		logsGrp.GET("/token", logH.ListToken)
		logsGrp.GET("/export", logH.Export)
		logsGrp.GET("/stream", logH.Stream)
		logsGrp.DELETE("/:source", stepUp, logH.Cleanup)
		logsGrp.POST("/:source/cleanup", stepUp, logH.Cleanup)
	}
}

func registerAPIRoutes(r *gin.Engine, cfg *config.Config, hh *handler.HealthHandler, ih *handler.ImageHandler, ah *handler.AuthHandler, originsFn func() []string) {
	apiPrefix := cfg.APIPrefix + "/api/v1"
	api := r.Group(apiPrefix, middleware.CORS(originsFn), middleware.Session(cfg, ah.DB()))
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
