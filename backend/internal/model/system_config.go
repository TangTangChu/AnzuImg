package model

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// SystemConfig 是 system_configs KV 表的 ORM 模型。
// 表本身在 main.ensureTables 已创建，这里仅提供读写工具。
type SystemConfig struct {
	Key       string    `gorm:"primaryKey;size:255"`
	Value     string    `gorm:"type:text"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (SystemConfig) TableName() string { return "system_configs" }

// LoadAllSystemConfigs 一次性把表里所有键值取回，供启动期初始化 Effective 快照。
func LoadAllSystemConfigs(db *gorm.DB) (map[string]string, error) {
	var rows []SystemConfig
	if err := db.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[string]string, len(rows))
	for _, r := range rows {
		out[r.Key] = r.Value
	}
	return out, nil
}

// SetSystemConfig 写入或更新单个键。空 value 不删除,删除请用 DeleteSystemConfig。
func SetSystemConfig(db *gorm.DB, key, value string) error {
	if strings.TrimSpace(key) == "" {
		return errors.New("system config key required")
	}
	row := SystemConfig{Key: key, Value: value, UpdatedAt: time.Now()}
	return db.Save(&row).Error
}

// DeleteSystemConfig 删除单个键，使其回退到 env 默认。
func DeleteSystemConfig(db *gorm.DB, key string) error {
	return db.Where("key = ?", key).Delete(&SystemConfig{}).Error
}

// DeleteSystemConfigs 批量删除。
func DeleteSystemConfigs(db *gorm.DB, keys []string) error {
	if len(keys) == 0 {
		return nil
	}
	return db.Where("key IN ?", keys).Delete(&SystemConfig{}).Error
}

// Coercion helpers — 这些在 SettingsService 重建快照时被使用。

func ParseConfigInt(raw string, def int) int {
	if raw == "" {
		return def
	}
	if n, err := strconv.Atoi(strings.TrimSpace(raw)); err == nil {
		return n
	}
	return def
}

func ParseConfigInt64(raw string, def int64) int64 {
	if raw == "" {
		return def
	}
	if n, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64); err == nil {
		return n
	}
	return def
}

func ParseConfigBool(raw string, def bool) bool {
	if raw == "" {
		return def
	}
	if b, err := strconv.ParseBool(strings.TrimSpace(raw)); err == nil {
		return b
	}
	return def
}

func ParseConfigStringList(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	// 优先尝试 JSON 数组
	if strings.HasPrefix(raw, "[") {
		var arr []string
		if err := json.Unmarshal([]byte(raw), &arr); err == nil {
			return cleanStringList(arr)
		}
	}
	// 否则按逗号分隔
	parts := strings.Split(raw, ",")
	return cleanStringList(parts)
}

func cleanStringList(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// EncodeConfigStringList 把切片以 JSON 数组形式存,往返稳定,避免逗号歧义。
func EncodeConfigStringList(list []string) string {
	if len(list) == 0 {
		return "[]"
	}
	b, err := json.Marshal(list)
	if err != nil {
		return "[]"
	}
	return string(b)
}
