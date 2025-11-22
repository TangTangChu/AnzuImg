package model

import "time"

type Image struct {
	ID        uint64    `gorm:"primaryKey"`
	Hash      string    `gorm:"size:64;uniqueIndex"`
	FileName  string    `gorm:"size:255"`
	MimeType  string    `gorm:"size:64"`
	Size      int64
	Path      string    `gorm:"size:512"` // 存储路径
	CreatedAt time.Time
	UpdatedAt time.Time
}

// 路由映射表
type ImageRoute struct {
	ID        uint64    `gorm:"primaryKey"`
	ImageID   uint64    `gorm:"index;not null"`
	Route     string    `gorm:"size:255;uniqueIndex;not null"` 
	CreatedAt time.Time

	Image Image `gorm:"foreignKey:ImageID;constraint:OnDelete:CASCADE"`
}
