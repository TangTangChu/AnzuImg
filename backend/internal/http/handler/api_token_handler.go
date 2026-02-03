package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/TangTangChu/AnzuImg/backend/internal/model"
	"github.com/TangTangChu/AnzuImg/backend/internal/service"
)

type APITokenHandler struct {
	svc *service.APITokenService
}

func NewAPITokenHandler(db *gorm.DB) *APITokenHandler {
	return &APITokenHandler{
		svc: service.NewAPITokenService(db),
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	rawToken, token, err := h.svc.CreateToken(req.Name, req.IPAllowlist, req.TokenType)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "invalid token type" {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	_ = h.svc.RecordLog(&model.APITokenLog{
		TokenID:   token.ID,
		TokenName: token.Name,
		TokenType: token.TokenType,
		Action:    "token_create",
		Method:    c.Request.Method,
		Path:      c.Request.URL.Path,
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})

	c.JSON(http.StatusOK, gin.H{
		"token":     token,
		"raw_token": rawToken,
	})
}

func (h *APITokenHandler) List(c *gin.Context) {
	tokens, err := h.svc.ListTokens()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tokens)
}

func (h *APITokenHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	h.deleteTokenByID(c, uint(id))
}

func (h *APITokenHandler) deleteTokenByID(c *gin.Context, id uint) {
	token, _ := h.svc.GetTokenByID(id)

	if err := h.svc.DeleteToken(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
			IPAddress: c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
		})
	}

	c.Status(http.StatusNoContent)
}

func (h *APITokenHandler) ListLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	logs, total, err := h.svc.ListLogs(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
	days, err := strconv.Atoi(c.DefaultQuery("days", "0"))
	if days <= 0 {
		var req CleanupLogsRequest
		if err := c.ShouldBindJSON(&req); err == nil {
			days = req.Days
		}
	}
	if err != nil {
		days = 0
	}
	if err != nil || days <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid days"})
		return
	}

	cutoff := time.Now().AddDate(0, 0, -days)
	deleted, err := h.svc.CleanupLogsBefore(cutoff)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"deleted": deleted,
		"cutoff":  cutoff,
	})
}
