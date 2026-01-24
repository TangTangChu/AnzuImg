package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	ServerAddr string
	DBHost string
	DBPort int
	DBUser string
	DBPass string
	DBName string
	DBSSL  string

	StorageBase string
	StorageType string // "local" 或 "cloud"
	
	// CORS配置
	AllowedOrigins []string
	
	// Passkey配置
	PasskeyRPID          string
	PasskeyRPOrigin      string
	PasskeyRPDisplayName string
	
	// 云端存储配置
	CloudEndpoint string
	CloudBucket   string
	CloudRegion   string
	CloudAccessKey string
	CloudSecretKey string
	CloudUseSSL   bool
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

func Load() *Config {
	// 解析允许的Origins
	allowedOrigins := []string{"http://localhost:3000"}
	if originsEnv := os.Getenv("ANZUIMG_ALLOWED_ORIGINS"); originsEnv != "" {
		allowedOrigins = strings.Split(originsEnv, ",")
		for i := range allowedOrigins {
			allowedOrigins[i] = strings.TrimSpace(allowedOrigins[i])
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
