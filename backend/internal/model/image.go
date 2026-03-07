package model

import (
	"time"

	"gorm.io/datatypes"
)

type Image struct {
	ID                  uint64         `gorm:"primaryKey" json:"id"`
	Hash                string         `gorm:"size:64;uniqueIndex" json:"hash"`
	FileName            string         `gorm:"size:255" json:"file_name"`
	MimeType            string         `gorm:"size:64" json:"mime_type"`
	Size                int64          `json:"size"`
	Path                string         `gorm:"column:storage_path;size:512" json:"path"`
	Width               int            `json:"width"`
	Height              int            `json:"height"`
	DurationSeconds     int            `gorm:"column:duration_seconds" json:"duration_seconds"`
	VideoCodec          string         `gorm:"column:video_codec;size:64" json:"video_codec"`
	VideoBitrate        int64          `gorm:"column:video_bitrate" json:"video_bitrate"`
	AudioCodec          string         `gorm:"column:audio_codec;size:64" json:"audio_codec"`
	AudioBitrate        int64          `gorm:"column:audio_bitrate" json:"audio_bitrate"`
	Description         string         `json:"description"`
	Tags                datatypes.JSON `gorm:"type:jsonb" json:"tags"`
	UploadedByTokenID   *uint          `gorm:"column:uploaded_by_token_id" json:"uploaded_by_token_id"`
	UploadedByTokenName string         `gorm:"size:255" json:"uploaded_by_token_name"`
	UploadedByTokenType string         `gorm:"size:32" json:"uploaded_by_token_type"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
}

// 路由映射表
type ImageRoute struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	ImageID   uint64    `gorm:"index;not null" json:"image_id"`
	Route     string    `gorm:"size:255;uniqueIndex;not null" json:"route"`
	CreatedAt time.Time `json:"created_at"`

	Image Image `gorm:"foreignKey:ImageID;constraint:OnDelete:CASCADE" json:"image"`
}
