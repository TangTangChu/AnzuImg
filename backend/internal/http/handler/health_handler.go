package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/TangTangChu/AnzuImg/backend/internal/service"
)

type HealthHandler struct {
	svc *service.HealthService
}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{svc: service.NewHealthService()}
}

func (h *HealthHandler) Health(c *gin.Context) {
	h.svc.HealthHandler(c)
}

func (h *HealthHandler) Ping(c *gin.Context) {
	h.svc.PingHandler(c)
}
