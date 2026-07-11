package model

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type LoginAttempt struct {
	ID        uint64    `gorm:"primaryKey"`
	IPAddress string    `gorm:"size:45;not null;index:idx_ip_created"`
	Username  string    `gorm:"size:100;not null;index:idx_ip_created"`
	Success   bool      `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null;index:idx_ip_created"`
}

// IsIPLocked 检查 IP 在 max/lockoutMin 策略下是否被锁定。
// 与旧版本相比，阈值与窗口都改为参数，由调用方从 effective 配置取。
func IsIPLocked(db *gorm.DB, ipAddress string, maxAttempts int, lockoutMin int) (bool, time.Time, error) {
	return IsLoginSubjectLocked(db, ipAddress, "admin", maxAttempts, lockoutMin)
}

func IsLoginSubjectLocked(db *gorm.DB, ipAddress, username string, maxAttempts int, lockoutMin int) (bool, time.Time, error) {
	if maxAttempts <= 0 || lockoutMin <= 0 {
		return false, time.Time{}, nil
	}
	var count int64
	lockoutTime := time.Now().Add(-time.Duration(lockoutMin) * time.Minute)

	if err := db.Model(&LoginAttempt{}).
		Where("ip_address = ? AND username = ? AND success = ? AND created_at > ?",
			ipAddress, username, false, lockoutTime).
		Count(&count).Error; err != nil {
		return false, time.Time{}, fmt.Errorf("count login attempts: %w", err)
	}

	if int(count) < maxAttempts {
		return false, time.Time{}, nil
	}

	var latest LoginAttempt
	if err := db.Model(&LoginAttempt{}).
		Where("ip_address = ? AND username = ? AND success = ? AND created_at > ?", ipAddress, username, false, lockoutTime).
		Order("created_at DESC").
		First(&latest).Error; err == nil {
		return true, latest.CreatedAt.Add(time.Duration(lockoutMin) * time.Minute), nil
	} else if err != gorm.ErrRecordNotFound {
		return false, time.Time{}, fmt.Errorf("load latest login attempt: %w", err)
	}

	return true, time.Now().Add(time.Duration(lockoutMin) * time.Minute), nil
}

// CountRecentFailedAttempts 给暴力破解告警使用：返回 IP 在 windowMin 分钟内的失败次数。
func CountRecentFailedAttempts(db *gorm.DB, ipAddress, username string, windowMin int) (int64, error) {
	if windowMin <= 0 {
		return 0, nil
	}
	windowStart := time.Now().Add(-time.Duration(windowMin) * time.Minute)
	var count int64
	if err := db.Model(&LoginAttempt{}).
		Where("ip_address = ? AND username = ? AND success = ? AND created_at > ?", ipAddress, username, false, windowStart).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
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

func ClearFailedLoginAttempts(db *gorm.DB, ipAddress, username string) error {
	return db.Where("ip_address = ? AND username = ? AND success = ?", ipAddress, username, false).
		Delete(&LoginAttempt{}).Error
}

// CleanOldLoginAttempts 清理旧的登录尝试记录。
// 保留窗口与最严格的锁定窗口一致,最少保留 24 小时以便审计。
func CleanOldLoginAttempts(db *gorm.DB) error {
	cutoffTime := time.Now().Add(-24 * time.Hour)
	return db.Where("created_at < ?", cutoffTime).Delete(&LoginAttempt{}).Error
}
