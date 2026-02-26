package service

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/TangTangChu/AnzuImg/backend/internal/clientip"
	"github.com/TangTangChu/AnzuImg/backend/internal/http/response"
	"github.com/TangTangChu/AnzuImg/backend/internal/model"
)

const (
	SessionCookieName = "anzuimg_session"
	SessionHeaderName = "X-Session-Token"
)

type SessionService struct {
	db *gorm.DB
}

func NewSessionService(db *gorm.DB) *SessionService {
	return &SessionService{db: db}
}

func requestClientIP(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if ip := clientip.FromRequest(c.Request); ip != "" {
		return ip
	}
	return c.ClientIP()
}

// CreateSession 创建新会话，添加会话固定攻击防护
func (s *SessionService) CreateSession(c *gin.Context) (string, *model.Session, error) {
	clientIP := requestClientIP(c)
	if clientIP == "" {
		clientIP = "unknown"
	}

	userAgent := c.Request.UserAgent()
	if err := model.RevokeAllUserSessions(s.db, model.DefaultUserID); err != nil {
		return "", nil, err
	}

	token, session, err := model.CreateSession(s.db, model.DefaultUserID, clientIP, userAgent)
	if err != nil {
		return "", nil, err
	}

	return token, session, nil
}

// ValidateSession 验证会话令牌
func (s *SessionService) ValidateSession(c *gin.Context) (*model.Session, error) {
	token := s.extractToken(c)
	if token == "" {
		return nil, gorm.ErrRecordNotFound
	}

	session, err := model.ValidateSession(s.db, token)
	if err != nil {
		return nil, err
	}

	// 严格的IP地址验证
	strictIP := false
	if v, ok := c.Get("strict_session_ip"); ok {
		if b, ok2 := v.(bool); ok2 {
			strictIP = b
		}
	}
	if strictIP {
		clientIP := requestClientIP(c)
		if clientIP != "" && clientIP != "unknown" && session.IPAddress != "" && session.IPAddress != "unknown" {
			if clientIP != session.IPAddress {
				// IP地址不匹配，撤销会话并返回错误
				_ = model.RevokeSession(s.db, model.HashToken(token))
				return nil, gorm.ErrRecordNotFound
			}
		}
	}

	return session, nil
}

// extractToken 从请求中提取令牌
func (s *SessionService) extractToken(c *gin.Context) string {
	// 优先从Cookie获取
	if cookie, err := c.Cookie(SessionCookieName); err == nil && cookie != "" {
		return cookie
	}

	// 从Authorization Header获取
	authHeader := c.GetHeader("Authorization")
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}

	// 从自定义Header获取
	if header := c.GetHeader(SessionHeaderName); header != "" {
		return header
	}

	return ""
}

// SetSessionCookie 设置会话Cookie
func (s *SessionService) SetSessionCookie(c *gin.Context, token string) {
	secure := false
	if c.Request.TLS != nil || c.Request.Header.Get("X-Forwarded-Proto") == "https" {
		secure = true
	}

	maxAge := int(model.SessionExpirationHours * 60 * 60)
	ss := "Lax"
	if v, ok := c.Get("cookie_samesite"); ok {
		if ssv, ok2 := v.(string); ok2 && ssv != "" {
			ss = ssv
		}
	}
	ssLower := strings.ToLower(ss)
	if ssLower != "lax" && ssLower != "strict" && ssLower != "none" {
		ss = "Lax"
	}

	cookie := &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	}
	switch strings.ToLower(ss) {
	case "strict":
		cookie.SameSite = http.SameSiteStrictMode
	case "none":
		cookie.SameSite = http.SameSiteNoneMode
	}

	if cookie.SameSite == http.SameSiteNoneMode {
		cookie.Secure = true
	}

	http.SetCookie(c.Writer, cookie)
}

// ClearSessionCookie 清除会话Cookie
func (s *SessionService) ClearSessionCookie(c *gin.Context) {
	secure := false
	if c.Request.TLS != nil || c.Request.Header.Get("X-Forwarded-Proto") == "https" {
		secure = true
	}

	ss := "Lax"
	if v, ok := c.Get("cookie_samesite"); ok {
		if ssv, ok2 := v.(string); ok2 && ssv != "" {
			ss = ssv
		}
	}
	ssLower := strings.ToLower(ss)
	if ssLower != "lax" && ssLower != "strict" && ssLower != "none" {
		ss = "Lax"
	}

	cookie := &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	}
	switch strings.ToLower(ss) {
	case "strict":
		cookie.SameSite = http.SameSiteStrictMode
	case "none":
		cookie.SameSite = http.SameSiteNoneMode
	}
	if cookie.SameSite == http.SameSiteNoneMode {
		cookie.Secure = true
	}

	http.SetCookie(c.Writer, cookie)
}

// RevokeCurrentSession 撤销当前会话
func (s *SessionService) RevokeCurrentSession(c *gin.Context) error {
	token := s.extractToken(c)
	if token == "" {
		return nil
	}

	tokenHash := model.HashToken(token)
	return model.RevokeSession(s.db, tokenHash)
}

// RevokeAllSessions 撤销所有会话
func (s *SessionService) RevokeAllSessions() error {
	return model.RevokeAllUserSessions(s.db, model.DefaultUserID)
}

// CleanExpiredSessions 清理过期会话
func (s *SessionService) CleanExpiredSessions() error {
	return model.CleanExpiredSessions(s.db)
}

// SessionMiddleware 会话中间件
func (s *SessionService) SessionMiddleware() gin.HandlerFunc {
	apiTokenService := NewAPITokenService(s.db)

	return func(c *gin.Context) {
		session, err := s.ValidateSession(c)
		if err == nil {
			c.Set("session", session)
			c.Set("user_id", session.UserID)
			c.Set("auth_method", "session")
			c.Next()
			return
		}
		token := s.extractToken(c)
		if token != "" {
			clientIP := requestClientIP(c)
			if clientIP == "" {
				clientIP = "unknown"
			}

			if apiToken, err := apiTokenService.ValidateToken(token, clientIP); err == nil {
				c.Set("api_token", apiToken)
				c.Set("user_id", uint64(model.DefaultUserID))
				c.Set("auth_method", "api_token")
				c.Next()
				return
			}
		}

		response.AbortErrorCode(c, http.StatusUnauthorized, "session_or_token_invalid", "invalid or expired session/token")
	}
}
