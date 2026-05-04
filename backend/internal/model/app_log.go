package model

import "time"

// AppLog 是应用日志表的 ORM 模型，对应 app_logs。
// 用作 logger DBSink 的存储；与 security_event_logs / api_token_logs 互补：
// 它收所有级别的应用层日志,包含运行时错误、HTTP 处理过程中的提示等,
// 不局限于安全事件。
type AppLog struct {
	ID        uint64    `gorm:"primaryKey"                                                              json:"id"`
	CreatedAt time.Time `gorm:"not null;index:idx_app_logs_created_at"                                   json:"created_at"`
	Level     string    `gorm:"size:16;not null;index:idx_app_logs_level_created"                        json:"level"`
	Module    string    `gorm:"size:64;not null;index:idx_app_logs_module_created"                       json:"module"`
	Message   string    `gorm:"type:text;not null"                                                       json:"message"`
	RequestID string    `gorm:"size:64"                                                                   json:"request_id,omitempty"`
	IPAddress string    `gorm:"size:45"                                                                   json:"ip_address,omitempty"`
}

func (AppLog) TableName() string { return "app_logs" }
