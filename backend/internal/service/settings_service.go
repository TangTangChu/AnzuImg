package service

import (
	"errors"
	"fmt"
	"net/netip"
	"strconv"
	"strings"
	"sync"

	"gorm.io/gorm"

	"github.com/TangTangChu/AnzuImg/backend/internal/config"
	"github.com/TangTangChu/AnzuImg/backend/internal/logger"
	"github.com/TangTangChu/AnzuImg/backend/internal/model"
)

// FieldType 标识 Settings schema 中字段的数据类型，前端按它决定渲染哪种控件。
type FieldType string

const (
	FieldString    FieldType = "string"
	FieldMultiline FieldType = "multiline"
	FieldInt       FieldType = "int"
	FieldInt64     FieldType = "int64"
	FieldBool      FieldType = "bool"
	FieldEnum      FieldType = "enum"
	FieldList      FieldType = "list"
)

// FieldGroup 是给前端 UI 分组用的。
type FieldGroup string

const (
	GroupUploads        FieldGroup = "uploads"
	GroupSession        FieldGroup = "session"
	GroupLoginSecurity  FieldGroup = "login_security"
	GroupPasswordPolicy FieldGroup = "password_policy"
	GroupNetwork        FieldGroup = "network"
	GroupLogs           FieldGroup = "logs"
	GroupStepUp         FieldGroup = "stepup"
)

// FieldSchema 描述一个可被 Web 修改的 effective 配置项。
type FieldSchema struct {
	Key             string     `json:"key"`
	Group           FieldGroup `json:"group"`
	Type            FieldType  `json:"type"`
	Default         any        `json:"default"`
	Min             *int64     `json:"min,omitempty"`
	Max             *int64     `json:"max,omitempty"`
	Options         []string   `json:"options,omitempty"` // 仅 enum
	Sensitive       bool       `json:"sensitive,omitempty"`
	RequiresRestart bool       `json:"requires_restart,omitempty"`
}

// FieldValue 是返回给前端的当前值与来源标记。
type FieldValue struct {
	Key             string `json:"key"`
	Value           any    `json:"value"`
	OverriddenInDB  bool   `json:"overridden_in_db"`
}

// SettingsService 负责合并 env 默认值与 system_configs 覆盖值，构造并热替换
// config.Effective；对 Web 端暴露 Schema/Get/Set/Reset 三类操作。
type SettingsService struct {
	cfg *config.Config
	db  *gorm.DB
	log *logger.Logger

	mu        sync.Mutex // 写路径串行化,读路径走 atomic
	listeners []func(*config.Effective)

	schema []FieldSchema
}

// NewSettingsService 创建配置服务并立即 reload 合并 DB 覆盖。
func NewSettingsService(cfg *config.Config, db *gorm.DB) *SettingsService {
	s := &SettingsService{
		cfg:    cfg,
		db:     db,
		log:    logger.Register("settings"),
		schema: buildSchema(),
	}
	if err := s.Reload(); err != nil {
		s.log.Warnf("initial settings reload failed (using env defaults): %v", err)
	}
	return s
}

// Snapshot 返回当前生效快照,atomic 读,无锁。
func (s *SettingsService) Snapshot() *config.Effective {
	return s.cfg.Effective()
}

// Schema 返回所有可编辑字段,前端表单渲染用。
func (s *SettingsService) Schema() []FieldSchema {
	out := make([]FieldSchema, len(s.schema))
	copy(out, s.schema)
	return out
}

// AllowsWebModify 暴露引导期开关。前端用它决定表单是否只读。
func (s *SettingsService) AllowsWebModify() bool {
	return s.cfg.AllowWebConfig
}

// CurrentValues 把当前生效值与"是否被 DB 覆盖"打包返回。
func (s *SettingsService) CurrentValues() ([]FieldValue, error) {
	overrides, err := model.LoadAllSystemConfigs(s.db)
	if err != nil {
		return nil, err
	}
	eff := s.Snapshot()
	out := make([]FieldValue, 0, len(s.schema))
	for _, f := range s.schema {
		_, in := overrides[f.Key]
		out = append(out, FieldValue{
			Key:            f.Key,
			Value:          extractEffectiveValue(eff, f.Key),
			OverriddenInDB: in,
		})
	}
	return out, nil
}

// OnReload 订阅配置重载事件,如 logger sink、CORS 列表的订阅者。
func (s *SettingsService) OnReload(fn func(*config.Effective)) {
	if fn == nil {
		return
	}
	s.mu.Lock()
	s.listeners = append(s.listeners, fn)
	s.mu.Unlock()
}

// Reload 重新拉 system_configs，构建 Effective 快照，原子替换并通知订阅者。
func (s *SettingsService) Reload() error {
	overrides, err := model.LoadAllSystemConfigs(s.db)
	if err != nil {
		return err
	}
	eff := config.DefaultEffective()
	for _, f := range s.schema {
		raw, ok := overrides[f.Key]
		if !ok {
			continue
		}
		if err := applyOverride(eff, f, raw); err != nil {
			s.log.Warnf("settings: skip invalid override %s=%q: %v", f.Key, raw, err)
		}
	}
	s.cfg.ReplaceEffective(eff)
	s.fireListeners(eff)
	return nil
}

// Set 校验并写入若干键，然后 Reload。
// keys 中若任一 key 不在 schema 中，整体回滚。
func (s *SettingsService) Set(values map[string]string) error {
	if !s.cfg.AllowWebConfig {
		return ErrWebConfigDisabled
	}
	if len(values) == 0 {
		return nil
	}
	// 先做整体校验
	for k, raw := range values {
		f, ok := s.findField(k)
		if !ok {
			return fmt.Errorf("unknown config key: %s", k)
		}
		if err := validateValue(f, raw); err != nil {
			return fmt.Errorf("%s: %w", k, err)
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	for k, raw := range values {
		if err := model.SetSystemConfig(tx, k, raw); err != nil {
			tx.Rollback()
			return err
		}
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	return s.Reload()
}

// Reset 删除若干键的 DB 覆盖,使其回到 env 默认。
func (s *SettingsService) Reset(keys []string) error {
	if !s.cfg.AllowWebConfig {
		return ErrWebConfigDisabled
	}
	if len(keys) == 0 {
		return nil
	}
	for _, k := range keys {
		if _, ok := s.findField(k); !ok {
			return fmt.Errorf("unknown config key: %s", k)
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := model.DeleteSystemConfigs(s.db, keys); err != nil {
		return err
	}
	return s.Reload()
}

func (s *SettingsService) fireListeners(eff *config.Effective) {
	s.mu.Lock()
	listeners := make([]func(*config.Effective), len(s.listeners))
	copy(listeners, s.listeners)
	s.mu.Unlock()
	for _, fn := range listeners {
		// 单个 listener 不影响其它
		func() {
			defer func() {
				if r := recover(); r != nil {
					s.log.Warnf("settings listener panic: %v", r)
				}
			}()
			fn(eff)
		}()
	}
}

func (s *SettingsService) findField(key string) (FieldSchema, bool) {
	for _, f := range s.schema {
		if f.Key == key {
			return f, true
		}
	}
	return FieldSchema{}, false
}

// ErrWebConfigDisabled 表示当前部署禁止从 Web 端写配置。
var ErrWebConfigDisabled = errors.New("web config modification is disabled")

// ----------------- Schema 定义 -----------------

func buildSchema() []FieldSchema {
	intRange := func(min, max int64) (mn, mx *int64) {
		return &min, &max
	}
	return []FieldSchema{
		// uploads
		{Key: "MAX_UPLOAD_MB", Group: GroupUploads, Type: FieldInt, Default: 110, Min: ptrInt(1), Max: ptrInt(102400)},
		{Key: "MAX_UPLOAD_FILE_MB", Group: GroupUploads, Type: FieldInt, Default: 60, Min: ptrInt(1), Max: ptrInt(102400)},
		{Key: "MAX_UPLOAD_FILES", Group: GroupUploads, Type: FieldInt, Default: 20, Min: ptrInt(1), Max: ptrInt(1000)},

		// session
		{Key: "COOKIE_SAMESITE", Group: GroupSession, Type: FieldEnum, Default: "Lax", Options: []string{"Lax", "Strict", "None"}},
		{Key: "STRICT_SESSION_IP", Group: GroupSession, Type: FieldBool, Default: false},
		{Key: "SESSION_EXPIRATION_HOURS", Group: GroupSession, Type: FieldInt, Default: 8, Min: ptrInt(1), Max: ptrInt(720)},
		{Key: "API_TOKEN_TTL_HOURS", Group: GroupSession, Type: FieldInt, Default: 0, Min: ptrInt(0), Max: ptrInt(87600)}, // 0 = never

		// login security
		{Key: "LOGIN_MAX_ATTEMPTS", Group: GroupLoginSecurity, Type: FieldInt, Default: 5, Min: ptrInt(1), Max: ptrInt(1000)},
		{Key: "LOGIN_LOCKOUT_MINUTES", Group: GroupLoginSecurity, Type: FieldInt, Default: 15, Min: ptrInt(1), Max: ptrInt(1440)},
		{Key: "BRUTEFORCE_ALERT_ATTEMPTS", Group: GroupLoginSecurity, Type: FieldInt, Default: 5, Min: ptrInt(1), Max: ptrInt(1000)},

		// password
		{Key: "PASSWORD_MIN_LENGTH", Group: GroupPasswordPolicy, Type: FieldInt, Default: 8, Min: ptrInt(8), Max: ptrInt(128)},
		{Key: "PASSWORD_REQUIRE_UPPER", Group: GroupPasswordPolicy, Type: FieldBool, Default: true},
		{Key: "PASSWORD_REQUIRE_LOWER", Group: GroupPasswordPolicy, Type: FieldBool, Default: true},
		{Key: "PASSWORD_REQUIRE_DIGIT", Group: GroupPasswordPolicy, Type: FieldBool, Default: true},
		{Key: "PASSWORD_REQUIRE_SYMBOL", Group: GroupPasswordPolicy, Type: FieldBool, Default: false},

		// network
		{Key: "ALLOWED_ORIGINS", Group: GroupNetwork, Type: FieldList, Default: []string{"http://localhost:9200"}},
		{Key: "IP_BLACKLIST", Group: GroupNetwork, Type: FieldList, Default: []string{}},
		{Key: "ADMIN_IP_ALLOWLIST", Group: GroupNetwork, Type: FieldList, Default: []string{}},
		{Key: "CSP_EXTRA", Group: GroupNetwork, Type: FieldMultiline, Default: ""},

		// logs
		{Key: "SECURITY_LOG_RETENTION_DAYS", Group: GroupLogs, Type: FieldInt, Default: 90, Min: ptrInt(1), Max: ptrInt(3650)},
		{Key: "TOKEN_LOG_RETENTION_DAYS", Group: GroupLogs, Type: FieldInt, Default: 30, Min: ptrInt(1), Max: ptrInt(3650)},
		{Key: "APP_LOG_RETENTION_DAYS", Group: GroupLogs, Type: FieldInt, Default: 14, Min: ptrInt(1), Max: ptrInt(3650)},
		{Key: "APP_LOG_STDOUT_LEVEL", Group: GroupLogs, Type: FieldEnum, Default: "info", Options: []string{"debug", "info", "warn", "error", "off"}},
		{Key: "APP_LOG_DB_LEVEL", Group: GroupLogs, Type: FieldEnum, Default: "info", Options: []string{"debug", "info", "warn", "error", "off"}},
		{Key: "APP_LOG_DB_BUFFER", Group: GroupLogs, Type: FieldInt, Default: 4096, Min: ptrInt(64), Max: ptrInt(65536)},
		{Key: "APP_LOG_FILE_ENABLED", Group: GroupLogs, Type: FieldBool, Default: true},
		{Key: "APP_LOG_FILE_LEVEL", Group: GroupLogs, Type: FieldEnum, Default: "info", Options: []string{"debug", "info", "warn", "error", "off"}},
		func() FieldSchema {
			min, max := intRange(1, 8192)
			return FieldSchema{Key: "APP_LOG_FILE_MAX_SIZE_MB", Group: GroupLogs, Type: FieldInt, Default: 50, Min: min, Max: max}
		}(),
		{Key: "APP_LOG_FILE_MAX_BACKUPS", Group: GroupLogs, Type: FieldInt, Default: 7, Min: ptrInt(0), Max: ptrInt(1000)},
		{Key: "APP_LOG_FILE_MAX_AGE_DAYS", Group: GroupLogs, Type: FieldInt, Default: 14, Min: ptrInt(0), Max: ptrInt(3650)},

		// stepup
		{Key: "STEP_UP_MAX_AGE_SEC", Group: GroupStepUp, Type: FieldInt, Default: 120, Min: ptrInt(15), Max: ptrInt(3600)},
	}
}

func ptrInt(v int64) *int64 { x := v; return &x }

// ----------------- 解析与应用 -----------------

func validateValue(f FieldSchema, raw string) error {
	switch f.Type {
	case FieldInt, FieldInt64:
		n, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
		if err != nil {
			return fmt.Errorf("expect integer, got %q", raw)
		}
		if f.Min != nil && n < *f.Min {
			return fmt.Errorf("must be >= %d", *f.Min)
		}
		if f.Max != nil && n > *f.Max {
			return fmt.Errorf("must be <= %d", *f.Max)
		}
	case FieldBool:
		if _, err := strconv.ParseBool(strings.TrimSpace(raw)); err != nil {
			return fmt.Errorf("expect boolean, got %q", raw)
		}
	case FieldEnum:
		v := strings.TrimSpace(raw)
		for _, opt := range f.Options {
			if strings.EqualFold(opt, v) {
				return nil
			}
		}
		return fmt.Errorf("must be one of %v", f.Options)
	case FieldList:
		items := model.ParseConfigStringList(raw)
		// 仅对 IP 类列表做 CIDR/IP 校验
		if strings.Contains(f.Key, "IP_") {
			for _, item := range items {
				if !isValidIPOrCIDR(item) {
					return fmt.Errorf("invalid IP or CIDR: %s", item)
				}
			}
		}
	case FieldString, FieldMultiline:
		// 任意字符串
	}
	return nil
}

func applyOverride(eff *config.Effective, f FieldSchema, raw string) error {
	switch f.Key {
	case "MAX_UPLOAD_MB":
		mb := model.ParseConfigInt64(raw, 110)
		eff.MaxUploadBytes = mb * 1024 * 1024
	case "MAX_UPLOAD_FILE_MB":
		mb := model.ParseConfigInt64(raw, 60)
		eff.MaxUploadFileBytes = mb * 1024 * 1024
	case "MAX_UPLOAD_FILES":
		eff.MaxUploadFiles = model.ParseConfigInt(raw, 20)
	case "COOKIE_SAMESITE":
		eff.CookieSameSite = strings.TrimSpace(raw)
	case "STRICT_SESSION_IP":
		eff.StrictSessionIP = model.ParseConfigBool(raw, false)
	case "SESSION_EXPIRATION_HOURS":
		eff.SessionExpirationHours = model.ParseConfigInt(raw, 8)
	case "API_TOKEN_TTL_HOURS":
		eff.APITokenTTLHours = model.ParseConfigInt(raw, 0)
	case "LOGIN_MAX_ATTEMPTS":
		eff.LoginMaxAttempts = model.ParseConfigInt(raw, 5)
	case "LOGIN_LOCKOUT_MINUTES":
		eff.LoginLockoutMinutes = model.ParseConfigInt(raw, 15)
	case "BRUTEFORCE_ALERT_ATTEMPTS":
		eff.BruteforceAlertAttempts = model.ParseConfigInt(raw, 5)
	case "PASSWORD_MIN_LENGTH":
		eff.PasswordPolicy.MinLength = model.ParseConfigInt(raw, 8)
	case "PASSWORD_REQUIRE_UPPER":
		eff.PasswordPolicy.RequireUpper = model.ParseConfigBool(raw, true)
	case "PASSWORD_REQUIRE_LOWER":
		eff.PasswordPolicy.RequireLower = model.ParseConfigBool(raw, true)
	case "PASSWORD_REQUIRE_DIGIT":
		eff.PasswordPolicy.RequireDigit = model.ParseConfigBool(raw, true)
	case "PASSWORD_REQUIRE_SYMBOL":
		eff.PasswordPolicy.RequireSymbol = model.ParseConfigBool(raw, false)
	case "ALLOWED_ORIGINS":
		eff.AllowedOrigins = model.ParseConfigStringList(raw)
	case "IP_BLACKLIST":
		eff.IPBlacklist = model.ParseConfigStringList(raw)
	case "ADMIN_IP_ALLOWLIST":
		eff.AdminIPAllowlist = model.ParseConfigStringList(raw)
	case "CSP_EXTRA":
		eff.CSPExtra = raw
	case "SECURITY_LOG_RETENTION_DAYS":
		eff.SecurityLogRetentionDays = model.ParseConfigInt(raw, 90)
	case "TOKEN_LOG_RETENTION_DAYS":
		eff.TokenLogRetentionDays = model.ParseConfigInt(raw, 30)
	case "APP_LOG_RETENTION_DAYS":
		eff.AppLogRetentionDays = model.ParseConfigInt(raw, 14)
	case "APP_LOG_STDOUT_LEVEL":
		eff.AppLogStdoutLevel = strings.ToLower(strings.TrimSpace(raw))
	case "APP_LOG_DB_LEVEL":
		eff.AppLogDBLevel = strings.ToLower(strings.TrimSpace(raw))
	case "APP_LOG_DB_BUFFER":
		eff.AppLogDBBufferSize = model.ParseConfigInt(raw, 4096)
	case "APP_LOG_FILE_ENABLED":
		eff.AppLogFileEnabled = model.ParseConfigBool(raw, true)
	case "APP_LOG_FILE_LEVEL":
		eff.AppLogFileLevel = strings.ToLower(strings.TrimSpace(raw))
	case "APP_LOG_FILE_MAX_SIZE_MB":
		eff.AppLogFileMaxSizeMB = model.ParseConfigInt(raw, 50)
	case "APP_LOG_FILE_MAX_BACKUPS":
		eff.AppLogFileMaxBackups = model.ParseConfigInt(raw, 7)
	case "APP_LOG_FILE_MAX_AGE_DAYS":
		eff.AppLogFileMaxAgeDays = model.ParseConfigInt(raw, 14)
	case "STEP_UP_MAX_AGE_SEC":
		eff.StepUpMaxAgeSeconds = model.ParseConfigInt(raw, 120)
	default:
		return fmt.Errorf("unhandled key: %s", f.Key)
	}
	return nil
}

func extractEffectiveValue(eff *config.Effective, key string) any {
	switch key {
	case "MAX_UPLOAD_MB":
		return eff.MaxUploadBytes / 1024 / 1024
	case "MAX_UPLOAD_FILE_MB":
		return eff.MaxUploadFileBytes / 1024 / 1024
	case "MAX_UPLOAD_FILES":
		return eff.MaxUploadFiles
	case "COOKIE_SAMESITE":
		return eff.CookieSameSite
	case "STRICT_SESSION_IP":
		return eff.StrictSessionIP
	case "SESSION_EXPIRATION_HOURS":
		return eff.SessionExpirationHours
	case "API_TOKEN_TTL_HOURS":
		return eff.APITokenTTLHours
	case "LOGIN_MAX_ATTEMPTS":
		return eff.LoginMaxAttempts
	case "LOGIN_LOCKOUT_MINUTES":
		return eff.LoginLockoutMinutes
	case "BRUTEFORCE_ALERT_ATTEMPTS":
		return eff.BruteforceAlertAttempts
	case "PASSWORD_MIN_LENGTH":
		return eff.PasswordPolicy.MinLength
	case "PASSWORD_REQUIRE_UPPER":
		return eff.PasswordPolicy.RequireUpper
	case "PASSWORD_REQUIRE_LOWER":
		return eff.PasswordPolicy.RequireLower
	case "PASSWORD_REQUIRE_DIGIT":
		return eff.PasswordPolicy.RequireDigit
	case "PASSWORD_REQUIRE_SYMBOL":
		return eff.PasswordPolicy.RequireSymbol
	case "ALLOWED_ORIGINS":
		return eff.AllowedOrigins
	case "IP_BLACKLIST":
		return eff.IPBlacklist
	case "ADMIN_IP_ALLOWLIST":
		return eff.AdminIPAllowlist
	case "CSP_EXTRA":
		return eff.CSPExtra
	case "SECURITY_LOG_RETENTION_DAYS":
		return eff.SecurityLogRetentionDays
	case "TOKEN_LOG_RETENTION_DAYS":
		return eff.TokenLogRetentionDays
	case "APP_LOG_RETENTION_DAYS":
		return eff.AppLogRetentionDays
	case "APP_LOG_STDOUT_LEVEL":
		return eff.AppLogStdoutLevel
	case "APP_LOG_DB_LEVEL":
		return eff.AppLogDBLevel
	case "APP_LOG_DB_BUFFER":
		return eff.AppLogDBBufferSize
	case "APP_LOG_FILE_ENABLED":
		return eff.AppLogFileEnabled
	case "APP_LOG_FILE_LEVEL":
		return eff.AppLogFileLevel
	case "APP_LOG_FILE_MAX_SIZE_MB":
		return eff.AppLogFileMaxSizeMB
	case "APP_LOG_FILE_MAX_BACKUPS":
		return eff.AppLogFileMaxBackups
	case "APP_LOG_FILE_MAX_AGE_DAYS":
		return eff.AppLogFileMaxAgeDays
	case "STEP_UP_MAX_AGE_SEC":
		return eff.StepUpMaxAgeSeconds
	}
	return nil
}

func isValidIPOrCIDR(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	if _, err := netip.ParsePrefix(s); err == nil {
		return true
	}
	if _, err := netip.ParseAddr(s); err == nil {
		return true
	}
	return false
}
