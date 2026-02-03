package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/TangTangChu/AnzuImg/backend/internal/model"
)

func RequireTokenScopes(scopes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authMethod, _ := c.Get("auth_method")
		if authMethod != "api_token" {
			c.Next()
			return
		}

		token, ok := c.Get("api_token")
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "api token required"})
			return
		}

		apiToken, ok := token.(*model.APIToken)
		if !ok || apiToken == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "api token required"})
			return
		}

		for _, scope := range scopes {
			if apiToken.HasScope(scope) {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient api token scope"})
	}
}

func RequireTokenType(tokenType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authMethod, _ := c.Get("auth_method")
		if authMethod != "api_token" {
			c.Next()
			return
		}

		token, ok := c.Get("api_token")
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "api token required"})
			return
		}

		apiToken, ok := token.(*model.APIToken)
		if !ok || apiToken == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "api token required"})
			return
		}

		if apiToken.NormalizedType() != tokenType {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient api token scope"})
			return
		}

		c.Next()
	}
}
