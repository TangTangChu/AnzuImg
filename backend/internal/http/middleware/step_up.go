package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/TangTangChu/AnzuImg/backend/internal/http/response"
	"github.com/TangTangChu/AnzuImg/backend/internal/model"
)

// RequireStepUp 要求当前会话在 maxAgeFn 返回的时间窗口内通过过 step-up,
// 否则返回 403 step_up_required,前端拿到后弹模态让用户重输密码或 Passkey。
// maxAgeFn 在每次请求时调用,允许 Web 端调 STEP_UP_MAX_AGE_SEC 后立即生效。
func RequireStepUp(maxAgeFn func() time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, ok := c.Get("session")
		if !ok {
			response.AbortErrorCode(c, http.StatusForbidden, "session_required", "session authentication required")
			return
		}
		session, ok := raw.(*model.Session)
		if !ok || session == nil {
			response.AbortErrorCode(c, http.StatusForbidden, "session_required", "session authentication required")
			return
		}
		var maxAge time.Duration
		if maxAgeFn != nil {
			maxAge = maxAgeFn()
		}
		if maxAge <= 0 {
			maxAge = 2 * time.Minute
		}
		if session.StepUpAt == nil || time.Since(*session.StepUpAt) > maxAge {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":              "step_up_required",
				"message":           "step-up authentication required",
				"available_methods": []string{"password", "passkey"},
				"max_age_seconds":   int(maxAge / time.Second),
			})
			return
		}
		c.Next()
	}
}
