package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

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
}

func (h *APITokenHandler) Create(c *gin.Context) {
	var req CreateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	rawToken, token, err := h.svc.CreateToken(req.Name, req.IPAllowlist)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
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

	if err := h.svc.DeleteToken(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
