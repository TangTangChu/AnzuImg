package model

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"time"

	"gorm.io/gorm"
)

type Session struct {
	ID        uint64    `gorm:"primaryKey"`
	TokenHash string    `gorm:"size:128;not null;uniqueIndex"`
	UserID    uint64    `gorm:"not null;index"`
	IPAddress string    `gorm:"size:45"`
	UserAgent string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"not null"`
	ExpiresAt time.Time `gorm:"not null;index"`
	LastUsed  time.Time `gorm:"not null"`
}

const (
	// 会话过期时间（小时）
	SessionExpirationHours = 8
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

// CreateSession 创建新会话
func CreateSession(db *gorm.DB, userID uint64, ipAddress, userAgent string) (string, *Session, error) {
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
		ExpiresAt: now.Add(time.Duration(SessionExpirationHours) * time.Hour),
		LastUsed:  now,
	}
	
	if err := db.Create(session).Error; err != nil {
		return "", nil, err
	}
	
	return token, session, nil
}

// ValidateSession 验证会话令牌
func ValidateSession(db *gorm.DB, token string) (*Session, error) {
	if token == "" {
		return nil, gorm.ErrRecordNotFound
	}
	
	tokenHash := HashToken(token)
	
	var session Session
	if err := db.Where("token_hash = ? AND expires_at > ?", tokenHash, time.Now()).First(&session).Error; err != nil {
		return nil, err
	}
	
	session.LastUsed = time.Now()
	if time.Until(session.ExpiresAt) < time.Duration(SessionExpirationHours/2)*time.Hour {
		session.ExpiresAt = time.Now().Add(time.Duration(SessionExpirationHours) * time.Hour)
	}

	db.Save(&session)
	
	return &session, nil
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
