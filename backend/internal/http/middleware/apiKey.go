package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/TangTangChu/AnzuImg/backend/internal/config"
)

func APIKey(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-API-Key")
		if key == "" || key != cfg.APIKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid api key",
			})
			return
		}
		c.Next()
	}
}
