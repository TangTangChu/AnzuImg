package model

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"time"

	"gorm.io/gorm"
)

type Session struct {
	ID        uint64     `gorm:"primaryKey"`
	TokenHash string     `gorm:"size:128;not null;uniqueIndex"`
	UserID    uint64     `gorm:"not null;index"`
	IPAddress string     `gorm:"size:45"`
	UserAgent string     `gorm:"type:text"`
	CreatedAt time.Time  `gorm:"not null"`
	ExpiresAt time.Time  `gorm:"not null;index"`
	LastUsed  time.Time  `gorm:"not null"`
	StepUpAt  *time.Time `gorm:"index:idx_sessions_step_up_at"`
}

const (
	// 默认会话过期小时数；当调用方未提供 effective TTL 时回落用。
	DefaultSessionExpirationHours = 8
	// 令牌字节长度
	TokenBytes = 32
)

// GenerateToken 生成随机令牌
func GenerateToken() (string, string, error) {
	tokenBytes := make([]byte, TokenBytes)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", "", err
	}
	token := hex.EncodeToString(tokenBytes)
	tokenHash := HashToken(token)

	return token, tokenHash, nil
}

// HashToken 计算令牌哈希值
func HashToken(token string) string {
	hash := sha512.Sum512([]byte(token))
	return hex.EncodeToString(hash[:])
}

// CreateSession 创建新会话,TTL 来自 effective.SessionExpirationHours。
func CreateSession(db *gorm.DB, userID uint64, ipAddress, userAgent string, expirationHours int) (string, *Session, error) {
	if expirationHours <= 0 {
		expirationHours = DefaultSessionExpirationHours
	}
	token, tokenHash, err := GenerateToken()
	if err != nil {
		return "", nil, err
	}

	now := time.Now()
	session := &Session{
		TokenHash: tokenHash,
		UserID:    userID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		CreatedAt: now,
		ExpiresAt: now.Add(time.Duration(expirationHours) * time.Hour),
		LastUsed:  now,
	}

	if err := db.Create(session).Error; err != nil {
		return "", nil, err
	}

	return token, session, nil
}

// ValidateSession 校验 token 并按 sliding window 续期。
// expirationHours 用于决定续期阈值与新过期时间。
func ValidateSession(db *gorm.DB, token string, expirationHours int) (*Session, error) {
	if token == "" {
		return nil, gorm.ErrRecordNotFound
	}
	if expirationHours <= 0 {
		expirationHours = DefaultSessionExpirationHours
	}

	tokenHash := HashToken(token)

	var session Session
	if err := db.Where("token_hash = ? AND expires_at > ?", tokenHash, time.Now()).First(&session).Error; err != nil {
		return nil, err
	}

	session.LastUsed = time.Now()
	if time.Until(session.ExpiresAt) < time.Duration(expirationHours/2)*time.Hour {
		session.ExpiresAt = time.Now().Add(time.Duration(expirationHours) * time.Hour)
	}

	if err := db.Save(&session).Error; err != nil {
		return nil, err
	}

	return &session, nil
}

// MarkSessionStepUp 把指定会话的 step_up_at 置为当前时间。
func MarkSessionStepUp(db *gorm.DB, tokenHash string) error {
	now := time.Now()
	return db.Model(&Session{}).
		Where("token_hash = ?", tokenHash).
		Update("step_up_at", now).Error
}

// ClearSessionStepUp 把 step_up_at 清空,极少需要,如管理员强制再确认。
func ClearSessionStepUp(db *gorm.DB, tokenHash string) error {
	return db.Model(&Session{}).
		Where("token_hash = ?", tokenHash).
		Update("step_up_at", nil).Error
}

// RevokeSession 撤销会话
func RevokeSession(db *gorm.DB, tokenHash string) error {
	return db.Where("token_hash = ?", tokenHash).Delete(&Session{}).Error
}

// RevokeAllUserSessions 撤销用户所有会话
func RevokeAllUserSessions(db *gorm.DB, userID uint64) error {
	return db.Where("user_id = ?", userID).Delete(&Session{}).Error
}

// CleanExpiredSessions 清理过期会话
func CleanExpiredSessions(db *gorm.DB) error {
	return db.Where("expires_at < ?", time.Now()).Delete(&Session{}).Error
}
