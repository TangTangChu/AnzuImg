package model

import "time"

type SecurityEventLog struct {
	ID        uint64    `gorm:"primaryKey"`
	Category  string    `gorm:"size:32;not null;index:idx_security_event_logs_created_at"`
	Level     string    `gorm:"size:16;not null"`
	Action    string    `gorm:"size:64;not null;index:idx_security_event_logs_action"`
	Message   string    `gorm:"size:255;not null"`
	Method    string    `gorm:"size:16"`
	Path      string    `gorm:"size:512"`
	IPAddress string    `gorm:"size:45;index:idx_security_event_logs_ip_created"`
	Username  string    `gorm:"size:100;index:idx_security_event_logs_user_created"`
	CreatedAt time.Time `gorm:"not null;index:idx_security_event_logs_created_at"`
}
