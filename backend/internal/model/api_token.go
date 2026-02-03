package model

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type APIToken struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	UserID      uint64         `json:"user_id" gorm:"not null"`
	Name        string         `json:"name" gorm:"not null"`
	TokenType   string         `json:"token_type" gorm:"size:32;not null;default:'full'"`
	TokenHash   string         `json:"-" gorm:"size:128;not null;uniqueIndex"` // SHA512 hash
	IPAllowlist datatypes.JSON `json:"ip_allowlist" gorm:"type:jsonb"`         // JSON string array of CIDRs
	LastUsedAt  *time.Time     `json:"last_used_at"`
	LastUsedIP  string         `json:"last_used_ip" gorm:"size:45"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

const (
	TokenTypeFull       = "full"
	TokenTypeUploadList = "upload"
	TokenTypeListOnly   = "list"
)

const (
	ScopeImagesUpload = "images:upload"
	ScopeImagesList   = "images:list"
)

func (t *APIToken) NormalizedType() string {
	if t == nil || t.TokenType == "" {
		return TokenTypeFull
	}
	return t.TokenType
}

func (t *APIToken) HasScope(scope string) bool {
	switch t.NormalizedType() {
	case TokenTypeFull:
		return true
	case TokenTypeUploadList:
		return scope == ScopeImagesUpload || scope == ScopeImagesList
	case TokenTypeListOnly:
		return scope == ScopeImagesList
	default:
		return false
	}
}

func GenerateAPIToken() (string, string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", "", err
	}
	rawToken := hex.EncodeToString(tokenBytes)
	tokenHash := HashToken(rawToken) // Defined in session.go
	return rawToken, tokenHash, nil
}

// UpdateUsage updates the last used stats
func (t *APIToken) UpdateUsage(db *gorm.DB, ip string) error {
	now := time.Now()
	return db.Model(t).Updates(map[string]interface{}{
		"last_used_at": now,
		"last_used_ip": ip,
	}).Error
}
