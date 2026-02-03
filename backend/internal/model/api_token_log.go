package model

import "time"

type APITokenLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TokenID   uint      `json:"token_id"`
	TokenName string    `gorm:"size:255" json:"token_name"`
	TokenType string    `gorm:"size:32" json:"token_type"`
	Action    string    `gorm:"size:64" json:"action"`
	Method    string    `gorm:"size:16" json:"method"`
	Path      string    `gorm:"size:512" json:"path"`
	IPAddress string    `gorm:"size:45" json:"ip_address"`
	UserAgent string    `gorm:"size:512" json:"user_agent"`
	ImageHash string    `gorm:"size:64" json:"image_hash"`
	CreatedAt time.Time `json:"created_at"`
}
