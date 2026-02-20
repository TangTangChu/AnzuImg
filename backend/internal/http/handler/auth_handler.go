package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/TangTangChu/AnzuImg/backend/internal/config"
	"github.com/TangTangChu/AnzuImg/backend/internal/http/response"
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

type SecurityLogItem struct {
	ID        uint64    `json:"id"`
	Category  string    `json:"category"`
	Level     string    `json:"level"`
	Action    string    `json:"action"`
	Message   string    `json:"message"`
	Method    string    `json:"method,omitempty"`
	Path      string    `json:"path,omitempty"`
	IPAddress string    `json:"ip_address"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
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
		response.WriteErrorCode(c, http.StatusBadRequest, "invalid_setup_request", "invalid password (min 8 chars)")
		return
	}

	// 优先从 body 读取 token，其次从 header 读取（向后兼容）
	setupToken := req.SetupToken
	if setupToken == "" {
		setupToken = c.GetHeader("X-Setup-Token")
	}

	if h.cfg.SetupToken != "" {
		if setupToken != h.cfg.SetupToken {
			response.WriteErrorCode(c, http.StatusForbidden, "setup_token_required", "setup token required")
			return

		}
	}
	// 如果没有设置 token，则仅允许本机初始化
	if h.cfg.SetupToken == "" {
		ip := c.ClientIP()
		if ip != "127.0.0.1" && ip != "::1" {
			response.WriteErrorCode(c, http.StatusForbidden, "setup_localhost_only", "setup is only allowed from localhost")
			return
		}
	}

	if h.userService.IsInitialized() {
		response.WriteErrorCode(c, http.StatusForbidden, "system_already_initialized", "system already initialized")
		return
	}
	if err := h.userService.EnsureAdminExists(); err != nil {
		response.WriteErrorCode(c, http.StatusInternalServerError, "ensure_admin_failed", "failed to ensure admin user exists")
		return
	}

	if err := h.userService.SetupAdmin(req.Password); err != nil {
		response.WriteErrorCode(c, http.StatusInternalServerError, "setup_password_failed", "failed to set password")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "system initialized successfully"})
}

// 密码登录
func (h *AuthHandler) AuthWithPassword(c *gin.Context) {
	var req PasswordAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.WriteErrorCode(c, http.StatusBadRequest, "invalid_login_request", "invalid request")
		return
	}
	clientIP := c.ClientIP()
	if clientIP == "" {
		clientIP = "unknown"
	}

	// 检查IP是否被锁定
	locked, unlockTime := model.IsIPLocked(h.db, clientIP)
	if locked {
		h.recordSecurityEventWithDedup(c, "warning", "login_rate_limited", "too many login attempts", time.Duration(model.LockoutDuration)*time.Minute)
		requestID, _ := c.Get(response.CtxRequestIDKey)
		requestIDStr, _ := requestID.(string)
		c.JSON(http.StatusTooManyRequests, gin.H{
			"code":        "too_many_login_attempts",
			"message":     "too many login attempts",
			"request_id":  requestIDStr,
			"unlock_time": unlockTime.Format(time.RFC3339),
		})
		return
	}

	// 验证密码
	if !h.userService.VerifyPassword(req.Password) {
		model.RecordLoginAttempt(h.db, clientIP, "admin", false)
		h.recordSecurityEvent(c, "warning", "login_failed", "failed login attempt")
		h.recordBruteforceAlertIfNeeded(c, clientIP)
		response.WriteErrorCode(c, http.StatusUnauthorized, "invalid_credentials", "invalid password or system not initialized")
		return
	}
	model.RecordLoginAttempt(h.db, clientIP, "admin", true)
	h.recordSecurityEvent(c, "info", "login_success", "successful login")

	token, session, err := h.sessionService.CreateSession(c)
	if err != nil {
		response.WriteErrorCode(c, http.StatusInternalServerError, "create_session_failed", "failed to create session")
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
		response.WriteErrorCode(c, http.StatusUnauthorized, "session_invalid", "invalid or expired session")
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
	h.recordSecurityEvent(c, "info", "logout", "session logged out")
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

func (h *AuthHandler) RegisterPasskeyBegin(c *gin.Context) {
	if h.passkeyService == nil {
		response.WriteErrorCode(c, http.StatusServiceUnavailable, "passkey_unavailable", "passkey service not available")
		return
	}

	creation, sessionID, err := h.passkeyService.BeginRegistration()
	if err != nil {
		response.WriteErrorCode(c, http.StatusInternalServerError, "passkey_register_begin_failed", "failed to begin passkey registration")
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
	if h.passkeyService == nil {
		response.WriteErrorCode(c, http.StatusServiceUnavailable, "passkey_unavailable", "passkey service not available")
		return
	}

	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		// 兼容旧的 Header 名称
		sessionID = c.GetHeader("X-Session-Data")
	}

	if sessionID == "" {
		response.WriteErrorCode(c, http.StatusBadRequest, "passkey_session_id_required", "X-Session-ID header required")
		return
	}

	if err := h.passkeyService.FinishRegistration(c.Request, sessionID); err != nil {
		h.recordSecurityEvent(c, "warning", "passkey_register_failed", "passkey registration failed")
		response.WriteErrorCode(c, http.StatusBadRequest, "passkey_register_finish_failed", "invalid passkey registration response")
		return
	}
	h.recordSecurityEvent(c, "info", "passkey_register_success", "passkey registered successfully")

	c.JSON(http.StatusOK, gin.H{"message": "passkey registered successfully"})
}

func (h *AuthHandler) LoginPasskeyBegin(c *gin.Context) {
	if h.passkeyService == nil {
		response.WriteErrorCode(c, http.StatusServiceUnavailable, "passkey_unavailable", "passkey service not available")
		return
	}

	assertion, sessionID, err := h.passkeyService.BeginLogin()
	if err != nil {
		response.WriteErrorCode(c, http.StatusInternalServerError, "passkey_login_begin_failed", "failed to begin passkey login")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"assertion":  assertion,
		"session_id": sessionID,
	})
}

func (h *AuthHandler) LoginPasskeyFinish(c *gin.Context) {
	if h.passkeyService == nil {
		response.WriteErrorCode(c, http.StatusServiceUnavailable, "passkey_unavailable", "passkey service not available")
		return
	}

	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID = c.GetHeader("X-Session-Data")
	}

	if sessionID == "" {
		response.WriteErrorCode(c, http.StatusBadRequest, "passkey_session_id_required", "X-Session-ID header required")
		return
	}

	if err := h.passkeyService.FinishLogin(c.Request, sessionID); err != nil {
		h.recordSecurityEvent(c, "warning", "passkey_login_failed", "failed passkey login attempt")
		response.WriteErrorCode(c, http.StatusUnauthorized, "passkey_login_failed", "invalid passkey login response")
		return
	}

	token, session, err := h.sessionService.CreateSession(c)
	if err != nil {
		response.WriteErrorCode(c, http.StatusInternalServerError, "create_session_failed", "failed to create session")
		return
	}

	h.sessionService.SetSessionCookie(c, token)
	h.recordSecurityEvent(c, "info", "passkey_login_success", "successful passkey login")

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
		response.WriteErrorCode(c, http.StatusServiceUnavailable, "passkey_unavailable", "passkey service not available")
		return
	}

	credentials, err := h.passkeyService.ListCredentials()
	if err != nil {
		response.WriteErrorCode(c, http.StatusInternalServerError, "list_passkeys_failed", "failed to list passkeys")
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
		response.WriteErrorCode(c, http.StatusServiceUnavailable, "passkey_unavailable", "passkey service not available")
		return
	}

	credentialID := c.Param("credential_id")
	if credentialID == "" {
		response.WriteErrorCode(c, http.StatusBadRequest, "credential_id_required", "credential_id parameter required")
		return
	}

	if err := h.passkeyService.DeleteCredential(credentialID); err != nil {
		if errors.Is(err, service.ErrCredentialNotFound) {
			h.recordSecurityEvent(c, "warning", "passkey_delete_failed", "passkey not found")
			response.WriteErrorCode(c, http.StatusNotFound, "credential_not_found", "credential not found")
		} else {
			h.recordSecurityEvent(c, "warning", "passkey_delete_failed", "failed to delete passkey")
			response.WriteErrorCode(c, http.StatusInternalServerError, "delete_passkey_failed", "failed to delete passkey")
		}
		return
	}
	h.recordSecurityEvent(c, "info", "passkey_delete_success", "passkey deleted successfully")

	c.JSON(http.StatusOK, gin.H{"message": "passkey deleted successfully"})
}

// GetPasskeyCount 获取PassKey凭证数量
func (h *AuthHandler) GetPasskeyCount(c *gin.Context) {
	if h.passkeyService == nil {
		response.WriteErrorCode(c, http.StatusServiceUnavailable, "passkey_unavailable", "passkey service not available")
		return
	}

	count, err := h.passkeyService.GetCredentialCount()
	if err != nil {
		response.WriteErrorCode(c, http.StatusInternalServerError, "passkey_count_failed", "failed to get passkey count")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count": count,
	})
}

// CheckPasskeyExists 检查是否有PassKey凭证
func (h *AuthHandler) CheckPasskeyExists(c *gin.Context) {
	if h.passkeyService == nil {
		response.WriteErrorCode(c, http.StatusServiceUnavailable, "passkey_unavailable", "passkey service not available")
		return
	}

	hasPasskey, err := h.passkeyService.HasPasskey()
	if err != nil {
		response.WriteErrorCode(c, http.StatusInternalServerError, "passkey_status_failed", "failed to check passkey status")
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
		response.WriteErrorCode(c, http.StatusBadRequest, "invalid_change_password_request", "invalid request")
		return
	}

	if err := h.userService.ChangePassword(req.CurrentPassword, req.NewPassword); err != nil {
		if errors.Is(err, service.ErrCurrentPasswordIncorrect) {
			h.recordSecurityEvent(c, "warning", "password_change_failed", "password change failed")
			response.WriteErrorCode(c, http.StatusUnauthorized, "password_update_failed", "password update failed")
			return
		}
		h.recordSecurityEvent(c, "warning", "password_change_failed", "password change failed")
		response.WriteErrorCode(c, http.StatusBadRequest, "password_update_failed", "password update failed")
		return
	}

	// 修改密码后撤销所有会话
	h.sessionService.RevokeAllSessions()
	h.recordSecurityEvent(c, "info", "password_changed", "password changed successfully")

	c.JSON(http.StatusOK, gin.H{"message": "password changed successfully"})
}

// ListSecurityLogs 列出安全与操作日志
func (h *AuthHandler) ListSecurityLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	failedOnly, _ := strconv.ParseBool(c.DefaultQuery("failed_only", "true"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}

	query := h.db.Model(&model.SecurityEventLog{})
	if failedOnly {
		query = query.Where("level = ? OR level = ?", "warning", "error")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.WriteErrorCode(c, http.StatusInternalServerError, "list_security_logs_failed", "failed to list security logs")
		return
	}

	offset := (page - 1) * pageSize
	var events []model.SecurityEventLog
	if err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&events).Error; err != nil {
		response.WriteErrorCode(c, http.StatusInternalServerError, "list_security_logs_failed", "failed to list security logs")
		return
	}

	items := make([]SecurityLogItem, 0, len(events))
	for _, event := range events {
		items = append(items, SecurityLogItem{
			ID:        event.ID,
			Category:  event.Category,
			Level:     event.Level,
			Action:    event.Action,
			Message:   event.Message,
			Method:    event.Method,
			Path:      event.Path,
			IPAddress: event.IPAddress,
			Username:  event.Username,
			CreatedAt: event.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  items,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

func (h *AuthHandler) recordSecurityEvent(c *gin.Context, level, action, message string) {
	log := &model.SecurityEventLog{
		Category:  "auth",
		Level:     level,
		Action:    action,
		Message:   message,
		Method:    c.Request.Method,
		Path:      c.Request.URL.Path,
		IPAddress: c.ClientIP(),
		Username:  "admin",
		CreatedAt: time.Now(),
	}
	if err := h.db.Create(log).Error; err != nil {
		h.log.Warnf("failed to record security event: %v", err)
	}
}

func (h *AuthHandler) recordSecurityEventWithDedup(c *gin.Context, level, action, message string, dedupWindow time.Duration) {
	cutoff := time.Now().Add(-dedupWindow)
	var exists int64
	err := h.db.Model(&model.SecurityEventLog{}).
		Where("action = ? AND ip_address = ? AND created_at > ?", action, c.ClientIP(), cutoff).
		Count(&exists).Error
	if err != nil {
		h.log.Warnf("failed to check security event dedup: %v", err)
	}
	if exists > 0 {
		return
	}
	h.recordSecurityEvent(c, level, action, message)
}

func (h *AuthHandler) recordBruteforceAlertIfNeeded(c *gin.Context, clientIP string) {
	windowStart := time.Now().Add(-time.Duration(model.LockoutDuration) * time.Minute)
	var failedCount int64
	err := h.db.Model(&model.LoginAttempt{}).
		Where("ip_address = ? AND success = ? AND created_at > ?", clientIP, false, windowStart).
		Count(&failedCount).Error
	if err != nil {
		h.log.Warnf("failed to count login attempts for alert: %v", err)
		return
	}

	if failedCount < model.MaxLoginAttempts {
		return
	}

	h.recordSecurityEventWithDedup(
		c,
		"error",
		"login_bruteforce_alert",
		"high-frequency failed login attempts detected",
		time.Duration(model.LockoutDuration)*time.Minute,
	)
}
