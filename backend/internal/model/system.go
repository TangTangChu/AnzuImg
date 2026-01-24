package model

import "time"

type SystemConfig struct {
	Key       string    `gorm:"primaryKey;size:255"`
	Value     string    `gorm:"type:text"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`
}
