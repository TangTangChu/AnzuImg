package model

// SystemStats 系统统计信息
type SystemStats struct {
	TotalImages       int64 `json:"total_images"`
	TotalSize         int64 `json:"total_size"`
	LoginFailures24h  int64 `json:"login_failures_24h"`
	SecurityEvents24h int64 `json:"security_events_24h"`
}
