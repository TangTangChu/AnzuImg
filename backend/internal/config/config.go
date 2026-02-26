package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	ServerAddr         string
	ShutdownTimeoutSec int
	DBHost             string
	DBPort             int
	DBUser             string
	DBPass             string
	DBName             string
	DBSSL              string

	StorageBase string
	StorageType string // "local" 或 "cloud"

	// API前缀
	APIPrefix string

	// CORS配置
	AllowedOrigins      []string
	TrustedProxies      []string
	ClientIPHeaders     []string
	ClientIPXFFStrategy string
	SetupToken          string

	// 上传限制（单位：字节）
	MaxUploadBytes     int64
	MaxUploadFileBytes int64
	MaxUploadFiles     int

	// Cookie SameSite: Lax/Strict/None
	CookieSameSite string
	// 是否启用会话严格IP绑定校验
	StrictSessionIP bool

	// Passkey配置
	PasskeyRPID          string
	PasskeyRPOrigin      string
	PasskeyRPDisplayName string

	// 云端存储配置
	CloudEndpoint  string
	CloudBucket    string
	CloudRegion    string
	CloudAccessKey string
	CloudSecretKey string
	CloudUseSSL    bool
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
	// 确保以 / 开头，移除尾部 /
	if !strings.HasPrefix(trimmed, "/") {
		trimmed = "/" + trimmed
	}
	trimmed = strings.TrimRight(trimmed, "/")
	return trimmed
}

func Load() *Config {
	// 解析允许的Origins
	allowedOrigins := []string{"http://localhost:9200"}
	if originsEnv := os.Getenv("ANZUIMG_ALLOWED_ORIGINS"); originsEnv != "" {
		allowedOrigins = strings.Split(originsEnv, ",")
		for i := range allowedOrigins {
			allowedOrigins[i] = strings.TrimSpace(allowedOrigins[i])
		}
	}

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

	return &Config{
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

		// API前缀
		APIPrefix: normalizeAPIPrefix(getEnv("ANZUIMG_API_PREFIX", "")),

		// CORS配置
		AllowedOrigins:      allowedOrigins,
		TrustedProxies:      trustedProxies,
		ClientIPHeaders:     clientIPHeaders,
		ClientIPXFFStrategy: clientIPXFFStrategy,
		SetupToken:          getEnv("ANZUIMG_SETUP_TOKEN", ""),

		MaxUploadBytes:     getEnvInt64MB("ANZUIMG_MAX_UPLOAD_MB", 110),
		MaxUploadFileBytes: getEnvInt64MB("ANZUIMG_MAX_UPLOAD_FILE_MB", 60),
		MaxUploadFiles:     getEnvInt("ANZUIMG_MAX_UPLOAD_FILES", 20),
		CookieSameSite:     getEnv("ANZUIMG_COOKIE_SAMESITE", "Lax"),
		StrictSessionIP:    getEnvBool("ANZUIMG_STRICT_SESSION_IP", false),

		// Passkey配置
		PasskeyRPID:          getEnv("ANZUIMG_PASSKEY_RP_ID", "localhost"),
		PasskeyRPOrigin:      getEnv("ANZUIMG_PASSKEY_RP_ORIGIN", "http://localhost:8080"),
		PasskeyRPDisplayName: getEnv("ANZUIMG_PASSKEY_RP_DISPLAY_NAME", "AnzuImg"),

		// 云端存储配置
		CloudEndpoint:  getEnv("ANZUIMG_CLOUD_ENDPOINT", "s3.amazonaws.com"),
		CloudBucket:    getEnv("ANZUIMG_CLOUD_BUCKET", "anzuimg-bucket"),
		CloudRegion:    getEnv("ANZUIMG_CLOUD_REGION", "us-east-1"),
		CloudAccessKey: getEnv("ANZUIMG_CLOUD_ACCESS_KEY", ""),
		CloudSecretKey: getEnv("ANZUIMG_CLOUD_SECRET_KEY", ""),
		CloudUseSSL:    getEnvBool("ANZUIMG_CLOUD_USE_SSL", true),
	}
}
