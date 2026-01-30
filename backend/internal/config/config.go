package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	ServerAddr string
	DBHost     string
	DBPort     int
	DBUser     string
	DBPass     string
	DBName     string
	DBSSL      string

	StorageBase string
	StorageType string // "local" 或 "cloud"

	// CORS配置
	AllowedOrigins []string
	TrustedProxies []string
	SetupToken     string

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

func Load() *Config {
	// 解析允许的Origins
	allowedOrigins := []string{"http://localhost:9200"}
	if originsEnv := os.Getenv("ANZUIMG_ALLOWED_ORIGINS"); originsEnv != "" {
		allowedOrigins = strings.Split(originsEnv, ",")
		for i := range allowedOrigins {
			allowedOrigins[i] = strings.TrimSpace(allowedOrigins[i])
		}
	}

	trustedProxies := []string{"127.0.0.1", "::1", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}
	if proxiesEnv := os.Getenv("ANZUIMG_TRUSTED_PROXIES"); proxiesEnv != "" {
		trustedProxies = strings.Split(proxiesEnv, ",")
		for i := range trustedProxies {
			trustedProxies[i] = strings.TrimSpace(trustedProxies[i])
		}
	}

	return &Config{
		ServerAddr: getEnv("ANZUIMG_SERVER_ADDR", ":8080"),

		DBHost: getEnv("ANZUIMG_DB_HOST", "localhost"),
		DBPort: getEnvInt("ANZUIMG_DB_PORT", 5432),
		DBUser: getEnv("ANZUIMG_DB_USER", "anzuuser"),
		DBPass: getEnv("ANZUIMG_DB_PASSWORD", "anzupass"),
		DBName: getEnv("ANZUIMG_DB_NAME", "anzuimg"),
		DBSSL:  getEnv("ANZUIMG_DB_SSLMODE", "disable"),

		StorageBase: getEnv("ANZUIMG_STORAGE_BASE", "./data/images"),
		StorageType: getEnv("ANZUIMG_STORAGE_TYPE", "local"),

		// CORS配置
		AllowedOrigins: allowedOrigins,
		TrustedProxies: trustedProxies,
		SetupToken:     getEnv("ANZUIMG_SETUP_TOKEN", ""),

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
