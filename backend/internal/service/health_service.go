package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthService struct{}

func NewHealthService() *HealthService { return &HealthService{} }

func (s *HealthService) HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *HealthService) PingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}
