package model

import (
	"time"

	"gorm.io/gorm"

	"github.com/TangTangChu/AnzuImg/backend/internal/logger"
)

type SecurityEventLog struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Category  string    `gorm:"size:32;not null;index:idx_security_event_logs_created_at" json:"category"`
	Level     string    `gorm:"size:16;not null" json:"level"`
	Action    string    `gorm:"size:64;not null;index:idx_security_event_logs_action" json:"action"`
	Message   string    `gorm:"size:255;not null" json:"message"`
	Method    string    `gorm:"size:16" json:"method"`
	Path      string    `gorm:"size:512" json:"path"`
	IPAddress string    `gorm:"size:45;index:idx_security_event_logs_ip_created" json:"ip_address"`
	Username  string    `gorm:"size:100;index:idx_security_event_logs_user_created" json:"username"`
	CreatedAt time.Time `gorm:"not null;index:idx_security_event_logs_created_at" json:"created_at"`
}

// BeforeCreate 统一 level 词表，把历史写法 warning 规整为规范短名 warn，
// 使安全日志与应用日志的级别命名一致。
func (s *SecurityEventLog) BeforeCreate(tx *gorm.DB) error {
	s.Level = logger.NormalizeLevelName(s.Level)
	return nil
}
