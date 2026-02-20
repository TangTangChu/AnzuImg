package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/TangTangChu/AnzuImg/backend/internal/http/response"
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
			response.AbortErrorCode(c, http.StatusForbidden, "api_token_required", "api token required")
			return
		}

		apiToken, ok := token.(*model.APIToken)
		if !ok || apiToken == nil {
			response.AbortErrorCode(c, http.StatusForbidden, "api_token_required", "api token required")
			return
		}

		for _, scope := range scopes {
			if apiToken.HasScope(scope) {
				c.Next()
				return
			}
		}

		response.AbortErrorCode(c, http.StatusForbidden, "api_token_scope_denied", "insufficient api token scope")
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
			response.AbortErrorCode(c, http.StatusForbidden, "api_token_required", "api token required")
			return
		}

		apiToken, ok := token.(*model.APIToken)
		if !ok || apiToken == nil {
			response.AbortErrorCode(c, http.StatusForbidden, "api_token_required", "api token required")
			return
		}

		if apiToken.NormalizedType() != tokenType {
			response.AbortErrorCode(c, http.StatusForbidden, "api_token_scope_denied", "insufficient api token scope")
			return
		}

		c.Next()
	}
}
