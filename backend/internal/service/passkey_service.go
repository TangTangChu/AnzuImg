package service

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"gorm.io/gorm"

	"github.com/TangTangChu/AnzuImg/backend/internal/clientip"
	"github.com/TangTangChu/AnzuImg/backend/internal/config"
	"github.com/TangTangChu/AnzuImg/backend/internal/model"
)

type PasskeyService struct {
	db           *gorm.DB
	userID       uint64
	webAuthn     *webauthn.WebAuthn
	sessionStore sync.Map // map[string]sessionItem
}

type sessionItem struct {
	data      webauthn.SessionData
	expiresAt time.Time
}

var ErrCredentialNotFound = errors.New("credential not found")

func NewPasskeyService(cfg *config.Config, db *gorm.DB) (*PasskeyService, error) {
	wconfig := &webauthn.Config{
		RPDisplayName: cfg.PasskeyRPDisplayName,
		RPID:          cfg.PasskeyRPID,
		RPOrigins:     []string{cfg.PasskeyRPOrigin},
	}

	w, err := webauthn.New(wconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create webauthn instance: %w", err)
	}

	s := &PasskeyService{
		db:       db,
		userID:   model.DefaultUserID, // 默认单用户系统
		webAuthn: w,
	}

	return s, nil
}

func (s *PasskeyService) cleanupExpiredSessions(now time.Time) {
	s.sessionStore.Range(func(key, value interface{}) bool {
		item := value.(sessionItem)
		if now.After(item.expiresAt) {
			s.sessionStore.Delete(key)
		}
		return true
	})
}

func (s *PasskeyService) storeSession(data webauthn.SessionData) (string, error) {
	s.cleanupExpiredSessions(time.Now())

	// 生成随机 Session ID
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	sessionID := hex.EncodeToString(bytes)

	s.sessionStore.Store(sessionID, sessionItem{
		data:      data,
		expiresAt: time.Now().Add(5 * time.Minute), // 5分钟有效期
	})

	return sessionID, nil
}

func (s *PasskeyService) GetSession(sessionID string) (*webauthn.SessionData, error) {
	s.cleanupExpiredSessions(time.Now())

	value, ok := s.sessionStore.Load(sessionID)
	if !ok {
		return nil, fmt.Errorf("session not found or expired")
	}

	item := value.(sessionItem)
	if time.Now().After(item.expiresAt) {
		s.sessionStore.Delete(sessionID)
		return nil, fmt.Errorf("session expired")
	}
	s.sessionStore.Delete(sessionID)

	return &item.data, nil
}

func (s *PasskeyService) GetOrCreateUser() (*model.User, error) {
	var user model.User
	if err := s.db.Preload("Credentials").First(&user, s.userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			user = model.User{
				ID: s.userID,
			}
			if err := s.db.Create(&user).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return &user, nil
}

func (s *PasskeyService) HasPasskey() (bool, error) {
	user, err := s.GetOrCreateUser()
	if err != nil {
		return false, err
	}
	return len(user.Credentials) > 0, nil
}

// BeginRegistration 开始注册流程
func (s *PasskeyService) BeginRegistration() (*protocol.CredentialCreation, string, error) {
	user, err := s.GetOrCreateUser()
	if err != nil {
		return nil, "", err
	}
	registerOptions := func(credCreationOpts *protocol.PublicKeyCredentialCreationOptions) {
		credCreationOpts.AuthenticatorSelection.UserVerification = protocol.VerificationPreferred
	}

	creation, sessionData, err := s.webAuthn.BeginRegistration(user, registerOptions)
	if err != nil {
		return nil, "", err
	}

	sessionID, err := s.storeSession(*sessionData)
	if err != nil {
		return nil, "", err
	}

	return creation, sessionID, nil
}

// FinishRegistration 完成注册流程
func (s *PasskeyService) FinishRegistration(req *http.Request, sessionID string) error {
	user, err := s.GetOrCreateUser()
	if err != nil {
		return err
	}

	sessionData, err := s.GetSession(sessionID)
	if err != nil {
		return err
	}

	credential, err := s.webAuthn.FinishRegistration(user, *sessionData, req)
	if err != nil {
		return err
	}

	// 收集环境信息
	userAgent := req.UserAgent()
	ipAddress := clientip.FromRequest(req)
	if ipAddress == "" {
		if host, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
			if net.ParseIP(host) != nil {
				ipAddress = host
			}
		} else if net.ParseIP(req.RemoteAddr) != nil {
			ipAddress = req.RemoteAddr
		}
	}
	if ipAddress == "" {
		ipAddress = "unknown"
	}
	deviceName := parseDeviceName(userAgent)
	newCred := model.FromWebAuthnCredential(*credential, userAgent, ipAddress, deviceName)
	newCred.UserID = user.ID

	if err := s.db.Create(newCred).Error; err != nil {
		return fmt.Errorf("failed to save credential: %w", err)
	}

	return nil
}

// BeginLogin 开始登录流程
func (s *PasskeyService) BeginLogin() (*protocol.CredentialAssertion, string, error) {
	user, err := s.GetOrCreateUser()
	if err != nil {
		return nil, "", err
	}

	loginOptions := func(credAssertionOpts *protocol.PublicKeyCredentialRequestOptions) {
		credAssertionOpts.UserVerification = protocol.VerificationPreferred
	}

	assertion, sessionData, err := s.webAuthn.BeginLogin(user, loginOptions)
	if err != nil {
		return nil, "", err
	}

	sessionID, err := s.storeSession(*sessionData)
	if err != nil {
		return nil, "", err
	}

	return assertion, sessionID, nil
}

// FinishLogin 完成登录流程
func (s *PasskeyService) FinishLogin(req *http.Request, sessionID string) error {
	user, err := s.GetOrCreateUser()
	if err != nil {
		return err
	}

	sessionData, err := s.GetSession(sessionID)
	if err != nil {
		return err
	}

	credential, err := s.webAuthn.FinishLogin(user, *sessionData, req)
	if err != nil {
		return err
	}

	credentialID := base64.RawURLEncoding.EncodeToString(credential.ID)

	if err := s.db.Model(&model.PasskeyCredential{}).
		Where("credential_id = ?", credentialID).
		Update("sign_count", credential.Authenticator.SignCount).Error; err != nil {
		return fmt.Errorf("failed to update credential sign count: %w", err)
	}

	return nil
}

// ListCredentials 列出所有PassKey凭证
func (s *PasskeyService) ListCredentials() ([]model.PasskeyCredential, error) {
	user, err := s.GetOrCreateUser()
	if err != nil {
		return nil, err
	}

	var credentials []model.PasskeyCredential
	if err := s.db.Where("user_id = ?", user.ID).Order("created_at DESC").Find(&credentials).Error; err != nil {
		return nil, err
	}

	return credentials, nil
}

// DeleteCredential 删除指定的PassKey凭证
func (s *PasskeyService) DeleteCredential(credentialID string) error {
	user, err := s.GetOrCreateUser()
	if err != nil {
		return err
	}

	result := s.db.Where("user_id = ? AND credential_id = ?", user.ID, credentialID).Delete(&model.PasskeyCredential{})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrCredentialNotFound
	}

	return nil
}

// GetCredentialCount 获取凭证数量
func (s *PasskeyService) GetCredentialCount() (int64, error) {
	user, err := s.GetOrCreateUser()
	if err != nil {
		return 0, err
	}

	var count int64
	if err := s.db.Model(&model.PasskeyCredential{}).Where("user_id = ?", user.ID).Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// parseDeviceName 从User-Agent字符串中解析设备名称
func parseDeviceName(userAgent string) string {
	if userAgent == "" {
		return "Unknown Device"
	}

	switch {
	case strings.Contains(userAgent, "Chrome"):
		return "Chrome Browser"
	case strings.Contains(userAgent, "Firefox"):
		return "Firefox Browser"
	case strings.Contains(userAgent, "Safari") && !strings.Contains(userAgent, "Chrome"):
		return "Safari Browser"
	case strings.Contains(userAgent, "Edge"):
		return "Edge Browser"
	case strings.Contains(userAgent, "Opera"):
		return "Opera Browser"
	}

	switch {
	case strings.Contains(userAgent, "Windows"):
		return "Windows Device"
	case strings.Contains(userAgent, "Mac OS X") || strings.Contains(userAgent, "Macintosh"):
		return "Mac Device"
	case strings.Contains(userAgent, "Linux"):
		return "Linux Device"
	case strings.Contains(userAgent, "Android"):
		return "Android Device"
	case strings.Contains(userAgent, "iPhone") || strings.Contains(userAgent, "iPad"):
		return "iOS Device"
	}

	return "Unknown Device"
}
