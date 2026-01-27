package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/TangTangChu/AnzuImg/backend/internal/config"
	"github.com/TangTangChu/AnzuImg/backend/internal/logger"
	"github.com/TangTangChu/AnzuImg/backend/internal/model"
	"github.com/TangTangChu/AnzuImg/backend/internal/service"
)

type AuthHandler struct {
	cfg            *config.Config
	db             *gorm.DB
	userService    *service.UserService
	sessionService *service.SessionService
	passkeyService *service.PasskeyService
	log            *logger.Logger
}

func NewAuthHandler(cfg *config.Config, db *gorm.DB) *AuthHandler {
	log := logger.Register("auth-handler")
	passkeyService, err := service.NewPasskeyService(cfg, db)
	if err != nil {
		log.Warnf("Passkey service initialization failed: %v. Passkey authentication will be unavailable.", err)
	}

	return &AuthHandler{
		cfg:            cfg,
		db:             db,
		userService:    service.NewUserService(db),
		sessionService: service.NewSessionService(db),
		passkeyService: passkeyService,
		log:            log,
	}
}

type PasswordAuthRequest struct {
	Password string `json:"password" binding:"required"`
}

type SetupRequest struct {
	Password   string `json:"password" binding:"required,min=8"`
	SetupToken string `json:"setup_token"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

type AuthResponse struct {
	Token      string    `json:"token"`
	ExpiresAt  time.Time `json:"expires_at"`
	AuthMethod string    `json:"auth_method"`
}

// 判断有没有设置初始密码
func (h *AuthHandler) CheckInit(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"initialized": h.userService.IsInitialized(),
	})
}

// 设置初始密码
func (h *AuthHandler) Setup(c *gin.Context) {
	var req SetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid password (min 8 chars)"})
		return
	}

	// 优先从 body 读取 token，其次从 header 读取（向后兼容）
	setupToken := req.SetupToken
	if setupToken == "" {
		setupToken = c.GetHeader("X-Setup-Token")
	}

	if h.cfg.SetupToken != "" {
		if setupToken != h.cfg.SetupToken {
			c.JSON(http.StatusForbidden, gin.H{"error": "setup token required"})
			return
		}
	}
	// 如果没有设置 token，则仅允许本机初始化
	if h.cfg.SetupToken == "" {
		ip := c.ClientIP()
		if ip != "127.0.0.1" && ip != "::1" {
			c.JSON(http.StatusForbidden, gin.H{"error": "setup is only allowed from localhost"})
			return
		}
	}

	if h.userService.IsInitialized() {
		c.JSON(http.StatusForbidden, gin.H{"error": "system already initialized"})
		return
	}
	if err := h.userService.EnsureAdminExists(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to ensure admin user exists"})
		return
	}

	if err := h.userService.SetupAdmin(req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "system initialized successfully"})
}

// 密码登录
func (h *AuthHandler) AuthWithPassword(c *gin.Context) {
	var req PasswordAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	clientIP := c.ClientIP()
	if clientIP == "" {
		clientIP = "unknown"
	}

	// 检查IP是否被锁定
	locked, unlockTime := model.IsIPLocked(h.db, clientIP)
	if locked {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error":       "too many login attempts",
			"unlock_time": unlockTime.Format(time.RFC3339),
		})
		return
	}

	// 验证密码
	if !h.userService.VerifyPassword(req.Password) {
		model.RecordLoginAttempt(h.db, clientIP, "admin", false)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid password or system not initialized"})
		return
	}
	model.RecordLoginAttempt(h.db, clientIP, "admin", true)

	token, session, err := h.sessionService.CreateSession(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}

	h.sessionService.SetSessionCookie(c, token)

	c.JSON(http.StatusOK, AuthResponse{
		Token:      token,
		ExpiresAt:  session.ExpiresAt,
		AuthMethod: "password",
	})
}

func (h *AuthHandler) ValidateSession(c *gin.Context) {
	session, err := h.sessionService.ValidateSession(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":       true,
		"auth_method": "session",
		"expires_at":  session.ExpiresAt,
		"created_at":  session.CreatedAt,
		"last_used":   session.LastUsed,
	})
}

// Logout 注销当前会话
func (h *AuthHandler) Logout(c *gin.Context) {
	_ = h.sessionService.RevokeCurrentSession(c)
	h.sessionService.ClearSessionCookie(c)
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

func (h *AuthHandler) RegisterPasskeyBegin(c *gin.Context) {
	if h.passkeyService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "passkey service not available"})
		return
	}

	creation, sessionID, err := h.passkeyService.BeginRegistration()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"creation":   creation,
		"session_id": sessionID,
	})
}

type PasskeyFinishRequest struct {
	SessionID string `json:"session_id" binding:"required"`
}

func (h *AuthHandler) RegisterPasskeyFinish(c *gin.Context) {
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		// 兼容旧的 Header 名称
		sessionID = c.GetHeader("X-Session-Data")
	}

	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Session-ID header required"})
		return
	}

	if err := h.passkeyService.FinishRegistration(c.Request, sessionID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "passkey registered successfully"})
}

func (h *AuthHandler) LoginPasskeyBegin(c *gin.Context) {
	if h.passkeyService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "passkey service not available"})
		return
	}

	assertion, sessionID, err := h.passkeyService.BeginLogin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"assertion":  assertion,
		"session_id": sessionID,
	})
}

func (h *AuthHandler) LoginPasskeyFinish(c *gin.Context) {
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID = c.GetHeader("X-Session-Data")
	}

	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Session-ID header required"})
		return
	}

	if err := h.passkeyService.FinishLogin(c.Request, sessionID); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	token, session, err := h.sessionService.CreateSession(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}

	h.sessionService.SetSessionCookie(c, token)

	c.JSON(http.StatusOK, AuthResponse{
		Token:      token,
		ExpiresAt:  session.ExpiresAt,
		AuthMethod: "passkey",
	})
}

func (h *AuthHandler) DB() *gorm.DB {
	return h.db
}

// ListPasskeys 列出所有PassKey凭证
func (h *AuthHandler) ListPasskeys(c *gin.Context) {
	if h.passkeyService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "passkey service not available"})
		return
	}

	credentials, err := h.passkeyService.ListCredentials()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"credentials": credentials,
		"count":       len(credentials),
	})
}

// DeletePasskey 删除指定的PassKey凭证
func (h *AuthHandler) DeletePasskey(c *gin.Context) {
	if h.passkeyService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "passkey service not available"})
		return
	}

	credentialID := c.Param("credential_id")
	if credentialID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "credential_id parameter required"})
		return
	}

	if err := h.passkeyService.DeleteCredential(credentialID); err != nil {
		if err.Error() == "credential not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "credential not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "passkey deleted successfully"})
}

// GetPasskeyCount 获取PassKey凭证数量
func (h *AuthHandler) GetPasskeyCount(c *gin.Context) {
	if h.passkeyService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "passkey service not available"})
		return
	}

	count, err := h.passkeyService.GetCredentialCount()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count": count,
	})
}

// CheckPasskeyExists 检查是否有PassKey凭证
func (h *AuthHandler) CheckPasskeyExists(c *gin.Context) {
	if h.passkeyService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "passkey service not available"})
		return
	}

	hasPasskey, err := h.passkeyService.HasPasskey()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"has_passkey": hasPasskey,
	})
}

// ChangePassword 修改管理员密码
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.userService.ChangePassword(req.CurrentPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// 修改密码后撤销所有会话
	h.sessionService.RevokeAllSessions()

	c.JSON(http.StatusOK, gin.H{"message": "password changed successfully"})
}
