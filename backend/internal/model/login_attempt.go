package model

import (
	"time"

	"gorm.io/gorm"
)

type LoginAttempt struct {
	ID        uint64    `gorm:"primaryKey"`
	IPAddress string    `gorm:"size:45;not null;index:idx_ip_created"` // IPv6最大长度45
	Username  string    `gorm:"size:100;not null;index:idx_ip_created"`
	Success   bool      `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null;index:idx_ip_created"`
}

const (
	// 最大登录尝试次数
	MaxLoginAttempts = 5
	// 锁定时间（分钟）
	LockoutDuration = 15
)

// IsIPLocked 检查IP地址是否被锁定
func IsIPLocked(db *gorm.DB, ipAddress string) (bool, time.Time) {
	var count int64
	lockoutTime := time.Now().Add(-time.Duration(LockoutDuration) * time.Minute)

	db.Model(&LoginAttempt{}).
		Where("ip_address = ? AND success = ? AND created_at > ?",
			ipAddress, false, lockoutTime).
		Count(&count)

	if count < MaxLoginAttempts {
		return false, time.Time{}
	}

	var latest LoginAttempt
	if err := db.Model(&LoginAttempt{}).
		Where("ip_address = ? AND success = ? AND created_at > ?", ipAddress, false, lockoutTime).
		Order("created_at DESC").
		First(&latest).Error; err == nil {
		return true, latest.CreatedAt.Add(time.Duration(LockoutDuration) * time.Minute)
	}

	return true, time.Now().Add(time.Duration(LockoutDuration) * time.Minute)
}

// RecordLoginAttempt 记录登录尝试
func RecordLoginAttempt(db *gorm.DB, ipAddress, username string, success bool) error {
	attempt := &LoginAttempt{
		IPAddress: ipAddress,
		Username:  username,
		Success:   success,
		CreatedAt: time.Now(),
	}

	return db.Create(attempt).Error
}

// CleanOldLoginAttempts 清理旧的登录尝试记录
func CleanOldLoginAttempts(db *gorm.DB) error {
	// 保留最近24小时的记录
	cutoffTime := time.Now().Add(-24 * time.Hour)
	return db.Where("created_at < ?", cutoffTime).Delete(&LoginAttempt{}).Error
}
