package service

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/TangTangChu/AnzuImg/backend/internal/clientip"
	"github.com/TangTangChu/AnzuImg/backend/internal/config"
	"github.com/TangTangChu/AnzuImg/backend/internal/http/response"
	"github.com/TangTangChu/AnzuImg/backend/internal/model"
)

const (
	SessionCookieName = "anzuimg_session"
	SessionHeaderName = "X-Session-Token"
)

type SessionService struct {
	cfg *config.Config
	db  *gorm.DB
}

func NewSessionService(cfg *config.Config, db *gorm.DB) *SessionService {
	return &SessionService{cfg: cfg, db: db}
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

func (s *SessionService) sessionTTL() int {
	if s.cfg == nil {
		return model.DefaultSessionExpirationHours
	}
	if h := s.cfg.Effective().SessionExpirationHours; h > 0 {
		return h
	}
	return model.DefaultSessionExpirationHours
}

// CreateSession 创建新会话,登录前会先撤销同用户旧会话以防 session fixation。
func (s *SessionService) CreateSession(c *gin.Context) (string, *model.Session, error) {
	clientIP := requestClientIP(c)
	if clientIP == "" {
		clientIP = "unknown"
	}

	userAgent := c.Request.UserAgent()
	if err := model.RevokeAllUserSessions(s.db, model.DefaultUserID); err != nil {
		return "", nil, err
	}

	token, session, err := model.CreateSession(s.db, model.DefaultUserID, clientIP, userAgent, s.sessionTTL())
	if err != nil {
		return "", nil, err
	}

	return token, session, nil
}

func (s *SessionService) ValidateSession(c *gin.Context) (*model.Session, error) {
	token := s.extractToken(c)
	if token == "" {
		return nil, gorm.ErrRecordNotFound
	}

	session, err := model.ValidateSession(s.db, token, s.sessionTTL())
	if err != nil {
		return nil, err
	}

	strictIP := false
	if s.cfg != nil {
		strictIP = s.cfg.Effective().StrictSessionIP
	}
	if strictIP {
		clientIP := requestClientIP(c)
		if clientIP != "" && clientIP != "unknown" && session.IPAddress != "" && session.IPAddress != "unknown" {
			if clientIP != session.IPAddress {
				_ = model.RevokeSession(s.db, model.HashToken(token))
				return nil, gorm.ErrRecordNotFound
			}
		}
	}

	return session, nil
}

func (s *SessionService) extractToken(c *gin.Context) string {
	if cookie, err := c.Cookie(SessionCookieName); err == nil && cookie != "" {
		return cookie
	}

	authHeader := c.GetHeader("Authorization")
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}

	if header := c.GetHeader(SessionHeaderName); header != "" {
		return header
	}

	return ""
}

// SetSessionCookie 写会话 Cookie,SameSite 与 TTL 来自 effective 快照。
func (s *SessionService) SetSessionCookie(c *gin.Context, token string) {
	secure := false
	if c.Request.TLS != nil || c.Request.Header.Get("X-Forwarded-Proto") == "https" {
		secure = true
	}

	maxAge := s.sessionTTL() * 60 * 60
	ss := "Lax"
	if s.cfg != nil {
		if v := strings.TrimSpace(s.cfg.Effective().CookieSameSite); v != "" {
			ss = v
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

func (s *SessionService) ClearSessionCookie(c *gin.Context) {
	secure := false
	if c.Request.TLS != nil || c.Request.Header.Get("X-Forwarded-Proto") == "https" {
		secure = true
	}

	ss := "Lax"
	if s.cfg != nil {
		if v := strings.TrimSpace(s.cfg.Effective().CookieSameSite); v != "" {
			ss = v
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

func (s *SessionService) RevokeCurrentSession(c *gin.Context) error {
	token := s.extractToken(c)
	if token == "" {
		return nil
	}

	tokenHash := model.HashToken(token)
	return model.RevokeSession(s.db, tokenHash)
}

func (s *SessionService) RevokeAllSessions() error {
	return model.RevokeAllUserSessions(s.db, model.DefaultUserID)
}

func (s *SessionService) CleanExpiredSessions() error {
	return model.CleanExpiredSessions(s.db)
}

// MarkStepUp 把当前会话标记为已通过 step-up,记录在 sessions.step_up_at。
func (s *SessionService) MarkStepUp(c *gin.Context) error {
	token := s.extractToken(c)
	if token == "" {
		return gorm.ErrRecordNotFound
	}
	return model.MarkSessionStepUp(s.db, model.HashToken(token))
}

func (s *SessionService) SessionMiddleware() gin.HandlerFunc {
	apiTokenService := NewAPITokenService(s.cfg, s.db)

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
