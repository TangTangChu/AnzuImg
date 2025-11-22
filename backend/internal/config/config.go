package config

import (
	"os"
	"strconv"
)

type Config struct {
	ServerAddr string
	DBHost string
	DBPort int
	DBUser string
	DBPass string
	DBName string
	DBSSL  string

	APIKey      string
	StorageBase string
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

func Load() *Config {
	return &Config{
		ServerAddr: getEnv("ANZUIMG_SERVER_ADDR", ":8080"),

		DBHost: getEnv("ANZUIMG_DB_HOST", "localhost"),
		DBPort: getEnvInt("ANZUIMG_DB_PORT", 5432),
		DBUser: getEnv("ANZUIMG_DB_USER", "anzuuser"),
		DBPass: getEnv("ANZUIMG_DB_PASSWORD", "anzupass"),
		DBName: getEnv("ANZUIMG_DB_NAME", "anzuimg"),
		DBSSL:  getEnv("ANZUIMG_DB_SSLMODE", "disable"),

		APIKey:      getEnv("ANZUIMG_API_KEY", "AnzuChan_Kawaii"),
		StorageBase: getEnv("ANZUIMG_STORAGE_BASE", "./data/images"),
	}
}
