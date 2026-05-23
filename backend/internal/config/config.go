package config

import (
	"os"
	"strconv"
	"strings"
	"sync/atomic"
)

// Config 启动期加载的"引导"配置——这些字段在进程生命周期内不变，
// 不会被 Web 端覆盖。运行期可热改的字段位于 Effective 结构。
type Config struct {
	ServerAddr         string
	ShutdownTimeoutSec int

	DBHost string
	DBPort int
	DBUser string
	DBPass string
	DBName string
	DBSSL  string

	StorageBase string
	StorageType string

	APIPrefix string

	TrustedProxies      []string
	ClientIPHeaders     []string
	ClientIPXFFStrategy string

	// 初始化保护
	SetupToken string

	// Passkey 引导,RP ID 与 Origin 域名变化会让既有凭证作废,固化
	PasskeyRPID          string
	PasskeyRPOrigin      string
	PasskeyRPDisplayName string

	// 云存储
	CloudEndpoint  string
	CloudBucket    string
	CloudRegion    string
	CloudAccessKey string
	CloudSecretKey string
	CloudUseSSL    bool

	// 是否允许 Web 端修改运行时配置；为 false 时 Settings 写接口拒绝
	AllowWebConfig bool

	// 应用日志文件输出目录,路径本身不允许 Web 改
	LogFileDir string

	// 运行期可热改的快照
	effective atomic.Pointer[Effective]
}

// Effective 是运行时可被覆盖、热生效的配置快照。
// 所有请求路径上的热配置读取都应走 cfg.Effective 的字段。
type Effective struct {
	// CORS / 会话
	AllowedOrigins  []string
	CookieSameSite  string
	StrictSessionIP bool

	// 上传限制
	MaxUploadBytes     int64
	MaxUploadFileBytes int64
	MaxUploadFiles     int

	// 会话与令牌生存期
	SessionExpirationHours int
	APITokenTTLHours       int // 0 = 永不过期

	// 登录策略
	LoginMaxAttempts        int
	LoginLockoutMinutes     int
	BruteforceAlertAttempts int

	// 密码策略
	PasswordPolicy PasswordPolicy

	// 网络访问控制
	IPBlacklist      []string // 全局黑名单
	AdminIPAllowlist []string // 管理面板白名单,空表示不限制

	// 日志保留
	SecurityLogRetentionDays int
	TokenLogRetentionDays    int
	AppLogRetentionDays      int

	// 应用日志 sink 控制
	AppLogStdoutLevel    string // debug/info/warn/error
	AppLogDBLevel        string // off/debug/info/warn/error
	AppLogDBBufferSize   int
	AppLogFileEnabled    bool
	AppLogFileLevel      string
	AppLogFileMaxSizeMB  int
	AppLogFileMaxBackups int
	AppLogFileMaxAgeDays int

	// 安全头
	CSPExtra string

	// Step-up 二次确认
	StepUpMaxAgeSeconds int

	// URL 拉取
	URLFetchTimeoutSeconds int
	URLFetchMaxBytes       int64
	URLFetchAllowPrivate   bool
}

type PasswordPolicy struct {
	MinLength     int
	RequireUpper  bool
	RequireLower  bool
	RequireDigit  bool
	RequireSymbol bool
}

// Effective 返回当前生效的快照，永远非 nil。
func (c *Config) Effective() *Effective {
	if e := c.effective.Load(); e != nil {
		return e
	}
	// 不应发生,Load 之后 Effective 必然非空
	return &Effective{}
}

// ReplaceEffective 原子替换运行时快照。仅 SettingsService 应调用。
func (c *Config) ReplaceEffective(e *Effective) {
	if e == nil {
		return
	}
	c.effective.Store(e)
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return []string{}
	}
	items := strings.Split(value, ",")
	result := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}
	return result
}

func getEnvList(def []string, keys ...string) []string {
	for _, key := range keys {
		if value, ok := os.LookupEnv(key); ok {
			return splitCSV(value)
		}
	}
	result := make([]string, len(def))
	copy(result, def)
	return result
}

func getEnvStringWithFallback(def string, keys ...string) string {
	for _, key := range keys {
		if value, ok := os.LookupEnv(key); ok {
			return strings.TrimSpace(value)
		}
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getEnvBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}

func getEnvInt64(key string, def int64) int64 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return def
}

func getEnvInt64MB(key string, defMB int64) int64 {
	mb := getEnvInt64(key, defMB)
	if mb < 0 {
		mb = 0
	}
	return mb * 1024 * 1024
}

func normalizeAPIPrefix(prefix string) string {
	trimmed := strings.TrimSpace(prefix)
	if trimmed == "" || trimmed == "/" {
		return ""
	}
	if !strings.HasPrefix(trimmed, "/") {
		trimmed = "/" + trimmed
	}
	trimmed = strings.TrimRight(trimmed, "/")
	return trimmed
}

// DefaultEffective 从环境变量构造一份"无 DB 覆盖"的运行期快照。
// SettingsService 启动后会用 system_configs 中的覆盖值再覆盖一遍。
func DefaultEffective() *Effective {
	allowedOrigins := []string{"http://localhost:9200"}
	if originsEnv := os.Getenv("ANZUIMG_ALLOWED_ORIGINS"); originsEnv != "" {
		allowedOrigins = splitCSV(originsEnv)
	}

	return &Effective{
		AllowedOrigins:  allowedOrigins,
		CookieSameSite:  getEnv("ANZUIMG_COOKIE_SAMESITE", "Lax"),
		StrictSessionIP: getEnvBool("ANZUIMG_STRICT_SESSION_IP", false),

		MaxUploadBytes:     getEnvInt64MB("ANZUIMG_MAX_UPLOAD_MB", 110),
		MaxUploadFileBytes: getEnvInt64MB("ANZUIMG_MAX_UPLOAD_FILE_MB", 60),
		MaxUploadFiles:     getEnvInt("ANZUIMG_MAX_UPLOAD_FILES", 20),

		SessionExpirationHours: getEnvInt("ANZUIMG_SESSION_EXPIRATION_HOURS", 8),
		APITokenTTLHours:       getEnvInt("ANZUIMG_API_TOKEN_TTL_HOURS", 0),

		LoginMaxAttempts:        getEnvInt("ANZUIMG_LOGIN_MAX_ATTEMPTS", 5),
		LoginLockoutMinutes:     getEnvInt("ANZUIMG_LOGIN_LOCKOUT_MINUTES", 15),
		BruteforceAlertAttempts: getEnvInt("ANZUIMG_BRUTEFORCE_ALERT_ATTEMPTS", 5),

		PasswordPolicy: PasswordPolicy{
			MinLength:     getEnvInt("ANZUIMG_PASSWORD_MIN_LENGTH", 8),
			RequireUpper:  getEnvBool("ANZUIMG_PASSWORD_REQUIRE_UPPER", true),
			RequireLower:  getEnvBool("ANZUIMG_PASSWORD_REQUIRE_LOWER", true),
			RequireDigit:  getEnvBool("ANZUIMG_PASSWORD_REQUIRE_DIGIT", true),
			RequireSymbol: getEnvBool("ANZUIMG_PASSWORD_REQUIRE_SYMBOL", false),
		},

		IPBlacklist:      getEnvList(nil, "ANZUIMG_IP_BLACKLIST"),
		AdminIPAllowlist: getEnvList(nil, "ANZUIMG_ADMIN_IP_ALLOWLIST"),

		SecurityLogRetentionDays: getEnvInt("ANZUIMG_SECURITY_LOG_RETENTION_DAYS", 90),
		TokenLogRetentionDays:    getEnvInt("ANZUIMG_TOKEN_LOG_RETENTION_DAYS", 30),
		AppLogRetentionDays:      getEnvInt("ANZUIMG_APP_LOG_RETENTION_DAYS", 14),

		AppLogStdoutLevel:    strings.ToLower(getEnv("ANZUIMG_APP_LOG_STDOUT_LEVEL", "info")),
		AppLogDBLevel:        strings.ToLower(getEnv("ANZUIMG_APP_LOG_DB_LEVEL", "info")),
		AppLogDBBufferSize:   getEnvInt("ANZUIMG_APP_LOG_DB_BUFFER", 4096),
		AppLogFileEnabled:    getEnvBool("ANZUIMG_APP_LOG_FILE_ENABLED", true),
		AppLogFileLevel:      strings.ToLower(getEnv("ANZUIMG_APP_LOG_FILE_LEVEL", "info")),
		AppLogFileMaxSizeMB:  getEnvInt("ANZUIMG_APP_LOG_FILE_MAX_SIZE_MB", 50),
		AppLogFileMaxBackups: getEnvInt("ANZUIMG_APP_LOG_FILE_MAX_BACKUPS", 7),
		AppLogFileMaxAgeDays: getEnvInt("ANZUIMG_APP_LOG_FILE_MAX_AGE_DAYS", 14),

		CSPExtra:            getEnv("ANZUIMG_CSP_EXTRA", ""),
		StepUpMaxAgeSeconds: getEnvInt("ANZUIMG_STEP_UP_MAX_AGE_SEC", 120),

		URLFetchTimeoutSeconds: getEnvInt("ANZUIMG_URL_FETCH_TIMEOUT_SEC", 15),
		URLFetchMaxBytes:       getEnvInt64MB("ANZUIMG_URL_FETCH_MAX_MB", 60),
		URLFetchAllowPrivate:   getEnvBool("ANZUIMG_URL_FETCH_ALLOW_PRIVATE", false),
	}
}

func Load() *Config {
	trustedProxies := getEnvList(
		[]string{"127.0.0.1", "::1", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"},
		"APP_TRUSTED_PROXIES",
		"ANZUIMG_TRUSTED_PROXIES",
	)
	clientIPHeaders := getEnvList(
		[]string{"X-Forwarded-For", "X-Real-IP"},
		"APP_CLIENT_IP_HEADERS",
		"ANZUIMG_CLIENT_IP_HEADERS",
	)
	clientIPXFFStrategy := getEnvStringWithFallback(
		"trusted",
		"APP_CLIENT_IP_XFF_STRATEGY",
		"ANZUIMG_CLIENT_IP_XFF_STRATEGY",
	)

	c := &Config{
		ServerAddr:         getEnv("ANZUIMG_SERVER_ADDR", ":8080"),
		ShutdownTimeoutSec: getEnvInt("ANZUIMG_SHUTDOWN_TIMEOUT_SEC", 10),

		DBHost: getEnv("ANZUIMG_DB_HOST", "localhost"),
		DBPort: getEnvInt("ANZUIMG_DB_PORT", 5432),
		DBUser: getEnv("ANZUIMG_DB_USER", "anzuuser"),
		DBPass: getEnv("ANZUIMG_DB_PASSWORD", "anzupass"),
		DBName: getEnv("ANZUIMG_DB_NAME", "anzuimg"),
		DBSSL:  getEnv("ANZUIMG_DB_SSLMODE", "disable"),

		StorageBase: getEnv("ANZUIMG_STORAGE_BASE", "./data/images"),
		StorageType: getEnv("ANZUIMG_STORAGE_TYPE", "local"),

		APIPrefix: normalizeAPIPrefix(getEnv("ANZUIMG_API_PREFIX", "")),

		TrustedProxies:      trustedProxies,
		ClientIPHeaders:     clientIPHeaders,
		ClientIPXFFStrategy: clientIPXFFStrategy,

		SetupToken: getEnv("ANZUIMG_SETUP_TOKEN", ""),

		PasskeyRPID:          getEnv("ANZUIMG_PASSKEY_RP_ID", "localhost"),
		PasskeyRPOrigin:      getEnv("ANZUIMG_PASSKEY_RP_ORIGIN", "http://localhost:8080"),
		PasskeyRPDisplayName: getEnv("ANZUIMG_PASSKEY_RP_DISPLAY_NAME", "AnzuImg"),

		CloudEndpoint:  getEnv("ANZUIMG_CLOUD_ENDPOINT", "s3.amazonaws.com"),
		CloudBucket:    getEnv("ANZUIMG_CLOUD_BUCKET", "anzuimg-bucket"),
		CloudRegion:    getEnv("ANZUIMG_CLOUD_REGION", "us-east-1"),
		CloudAccessKey: getEnv("ANZUIMG_CLOUD_ACCESS_KEY", ""),
		CloudSecretKey: getEnv("ANZUIMG_CLOUD_SECRET_KEY", ""),
		CloudUseSSL:    getEnvBool("ANZUIMG_CLOUD_USE_SSL", true),

		AllowWebConfig: getEnvBool("ANZUIMG_ALLOW_WEB_CONFIG", true),
		LogFileDir:     getEnv("ANZUIMG_LOG_FILE_DIR", "./data/logs"),
	}
	c.ReplaceEffective(DefaultEffective())
	return c
}
