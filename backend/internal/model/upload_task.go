package model

import (
	"time"

	"gorm.io/datatypes"
)

const (
	UploadTaskStatusPending   = "pending"
	UploadTaskStatusRunning   = "running"
	UploadTaskStatusSucceeded = "succeeded"
	UploadTaskStatusFailed    = "failed"
)

type UploadTask struct {
	ID           string         `gorm:"size:36;primaryKey" json:"id"`
	Status       string         `gorm:"size:32;index;not null" json:"status"`
	FileName     string         `gorm:"size:255" json:"file_name"`
	Result       datatypes.JSON `gorm:"type:jsonb" json:"result,omitempty"`
	ErrorCode    string         `gorm:"size:64" json:"error_code,omitempty"`
	ErrorMessage string         `json:"error_message,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	CompletedAt  *time.Time     `json:"completed_at,omitempty"`
}
