package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/TangTangChu/AnzuImg/backend/internal/http/middleware"
	"github.com/TangTangChu/AnzuImg/backend/internal/http/response"
	"github.com/TangTangChu/AnzuImg/backend/internal/logger"
	"github.com/TangTangChu/AnzuImg/backend/internal/model"
	"github.com/TangTangChu/AnzuImg/backend/internal/service"
)

type APITokenHandler struct {
	db  *gorm.DB
	svc *service.APITokenService
	log *logger.Logger
}

func NewAPITokenHandler(db *gorm.DB) *APITokenHandler {
	return &APITokenHandler{
		db:  db,
		svc: service.NewAPITokenService(db),
		log: logger.Register("api-token-handler"),
	}
}

type CreateTokenRequest struct {
	Name        string   `json:"name" binding:"required"`
	IPAllowlist []string `json:"ip_allowlist"`
	TokenType   string   `json:"token_type"`
}

type CleanupLogsRequest struct {
	Days int `json:"days"`
}

func (h *APITokenHandler) Create(c *gin.Context) {
	var req CreateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.WriteErrorCode(c, http.StatusBadRequest, "invalid_request", "invalid request")
		return
	}

	rawToken, token, err := h.svc.CreateToken(req.Name, req.IPAllowlist, req.TokenType)
	if err != nil {
		status := http.StatusInternalServerError
		message := "failed to create token"
		if errors.Is(err, service.ErrInvalidTokenType) {
			status = http.StatusBadRequest
			message = "invalid token type"
		}
		h.recordSecurityEvent(c, "warning", "token_create_failed", message)
		response.WriteError(c, status, message)
		return
	}

	_ = h.svc.RecordLog(&model.APITokenLog{
		TokenID:   token.ID,
		TokenName: token.Name,
		TokenType: token.TokenType,
		Action:    "token_create",
		Method:    c.Request.Method,
		Path:      c.Request.URL.Path,
		IPAddress: middleware.ClientIP(c),
		UserAgent: c.Request.UserAgent(),
	})
	h.recordSecurityEvent(c, "info", "token_create_success", "api token created")

	c.JSON(http.StatusOK, gin.H{
		"token":     token,
		"raw_token": rawToken,
	})
}

func (h *APITokenHandler) List(c *gin.Context) {
	tokens, err := h.svc.ListTokens()
	if err != nil {
		response.WriteErrorCode(c, http.StatusInternalServerError, "list_tokens_failed", "failed to list tokens")
		return
	}
	c.JSON(http.StatusOK, tokens)
}

func (h *APITokenHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.WriteErrorCode(c, http.StatusBadRequest, "invalid_id", "invalid id")
		return
	}

	h.deleteTokenByID(c, uint(id))
}

func (h *APITokenHandler) deleteTokenByID(c *gin.Context, id uint) {
	token, _ := h.svc.GetTokenByID(id)

	if err := h.svc.DeleteToken(id); err != nil {
		h.recordSecurityEvent(c, "warning", "token_delete_failed", "failed to delete token")
		response.WriteErrorCode(c, http.StatusInternalServerError, "delete_token_failed", "failed to delete token")
		return
	}

	if token != nil {
		_ = h.svc.RecordLog(&model.APITokenLog{
			TokenID:   token.ID,
			TokenName: token.Name,
			TokenType: token.TokenType,
			Action:    "token_delete",
			Method:    c.Request.Method,
			Path:      c.Request.URL.Path,
			IPAddress: middleware.ClientIP(c),
			UserAgent: c.Request.UserAgent(),
		})
	}
	h.recordSecurityEvent(c, "info", "token_delete_success", "api token deleted")

	c.Status(http.StatusNoContent)
}

func (h *APITokenHandler) ListLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	search := c.DefaultQuery("search", "")
	startDate := c.DefaultQuery("start_date", "")
	endDate := c.DefaultQuery("end_date", "")
	actionType := c.DefaultQuery("type", "")

	logs, total, err := h.svc.ListLogs(page, pageSize, search, startDate, endDate, actionType)
	if err != nil {
		response.WriteErrorCode(c, http.StatusInternalServerError, "list_token_logs_failed", "failed to list token logs")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  logs,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

func (h *APITokenHandler) CleanupLogs(c *gin.Context) {
	days := 0
	if q := c.Query("days"); q != "" {
		qDays, err := strconv.Atoi(q)
		if err != nil {
			response.WriteErrorCode(c, http.StatusBadRequest, "invalid_days", "invalid days")
			return
		}
		days = qDays
	}

	if days <= 0 {
		var req CleanupLogsRequest
		if err := c.ShouldBindJSON(&req); err == nil {
			days = req.Days
		}
	}

	if days <= 0 {
		response.WriteErrorCode(c, http.StatusBadRequest, "invalid_days", "invalid days")
		return
	}

	cutoff := time.Now().AddDate(0, 0, -days)
	deleted, err := h.svc.CleanupLogsBefore(cutoff)
	if err != nil {
		h.recordSecurityEvent(c, "warning", "token_logs_cleanup_failed", "failed to cleanup token logs")
		response.WriteErrorCode(c, http.StatusInternalServerError, "cleanup_token_logs_failed", "failed to cleanup token logs")
		return
	}
	h.recordSecurityEvent(c, "info", "token_logs_cleanup", "token logs cleanup executed")

	c.JSON(http.StatusOK, gin.H{
		"deleted": deleted,
		"cutoff":  cutoff,
	})
}

func (h *APITokenHandler) recordSecurityEvent(c *gin.Context, level, action, message string) {
	event := &model.SecurityEventLog{
		Category:  "auth",
		Level:     level,
		Action:    action,
		Message:   message,
		Method:    c.Request.Method,
		Path:      c.Request.URL.Path,
		IPAddress: middleware.ClientIP(c),
		Username:  "admin",
		CreatedAt: time.Now(),
	}
	if err := h.db.Create(event).Error; err != nil {
		h.log.Warnf("failed to record token security event: %v", err)
	}
}
