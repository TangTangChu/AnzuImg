package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"time"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/TangTangChu/AnzuImg/backend/internal/config"
	httpserver "github.com/TangTangChu/AnzuImg/backend/internal/http"
	"github.com/TangTangChu/AnzuImg/backend/internal/logger"
	"github.com/TangTangChu/AnzuImg/backend/internal/model"
)

func quotePostgresIdentifier(v string) string {
	return `"` + strings.ReplaceAll(v, `"`, `""`) + `"`
}

func ensureDatabase(cfg *config.Config, log *logger.Logger) error {
	adminDSN := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=postgres sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBSSL,
	)

	adminDB, err := gorm.Open(postgres.Open(adminDSN), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("connect admin db failed: %w", err)
	}

	var count int64
	if err := adminDB.
		Raw("SELECT COUNT(*) FROM pg_database WHERE datname = ?", cfg.DBName).
		Scan(&count).Error; err != nil {
		return fmt.Errorf("check database exists failed: %w", err)
	}

	if count == 0 {
		log.Infof("database %s not found, creating...", cfg.DBName)
		createDBSQL := fmt.Sprintf("CREATE DATABASE %s", quotePostgresIdentifier(cfg.DBName))
		if err := adminDB.Exec(createDBSQL).Error; err != nil {
			return fmt.Errorf("create database failed: %w", err)
		}
		log.Infof("database %s created", cfg.DBName)
	} else {
		log.Infof("database %s already exists", cfg.DBName)
	}

	return nil
}

func ensureTables(db *gorm.DB, log *logger.Logger) error {
	return db.Transaction(func(tx *gorm.DB) error {

		createImagesTable := `
CREATE TABLE IF NOT EXISTS images (
    id            BIGSERIAL PRIMARY KEY,
    hash          VARCHAR(64)  NOT NULL UNIQUE,
    file_name     VARCHAR(255) NOT NULL,
    mime_type     VARCHAR(64)  NOT NULL,
    size          BIGINT       NOT NULL,
    storage_path  VARCHAR(512) NOT NULL,
    width         INTEGER,
    height        INTEGER,
    description   TEXT,
    tags          JSONB,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_images_hash ON images(hash);
CREATE INDEX IF NOT EXISTS idx_images_created_at ON images(created_at);
CREATE INDEX IF NOT EXISTS idx_images_tags ON images USING GIN(tags);
`
		if err := tx.Exec(createImagesTable).Error; err != nil {
			return fmt.Errorf("create images table failed: %w", err)
		}

		alterImagesTable := `
ALTER TABLE images ADD COLUMN IF NOT EXISTS uploaded_by_token_id BIGINT;
ALTER TABLE images ADD COLUMN IF NOT EXISTS uploaded_by_token_name VARCHAR(255);
ALTER TABLE images ADD COLUMN IF NOT EXISTS uploaded_by_token_type VARCHAR(32);
CREATE INDEX IF NOT EXISTS idx_images_uploaded_by_token_id ON images(uploaded_by_token_id);
`
		if err := tx.Exec(alterImagesTable).Error; err != nil {
			return fmt.Errorf("alter images table failed: %w", err)
		}

		createRoutesTable := `
CREATE TABLE IF NOT EXISTS image_routes (
    id         BIGSERIAL PRIMARY KEY,
    image_id   BIGINT NOT NULL REFERENCES images(id) ON DELETE CASCADE,
    route      VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_image_routes_image_id ON image_routes(image_id);
`
		if err := tx.Exec(createRoutesTable).Error; err != nil {
			return fmt.Errorf("create image_routes table failed: %w", err)
		}

		createUsersTable := `
CREATE TABLE IF NOT EXISTS users (
    id            BIGSERIAL PRIMARY KEY,
    password_hash VARCHAR(255),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`
		if err := tx.Exec(createUsersTable).Error; err != nil {
			return fmt.Errorf("create users table failed: %w", err)
		}

		createPasskeyCredentialsTable := `
CREATE TABLE IF NOT EXISTS passkey_credentials (
    id               BIGSERIAL PRIMARY KEY,
    credential_id    VARCHAR(512) NOT NULL UNIQUE,
    public_key       BYTEA NOT NULL,
    attestation_type VARCHAR(64),
    aaguid           BYTEA,
    sign_count       INTEGER NOT NULL DEFAULT 0,
    user_id          BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_agent       TEXT,
    ip_address       VARCHAR(45),
    device_name      VARCHAR(255),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_passkey_credentials_user_id ON passkey_credentials(user_id);
CREATE INDEX IF NOT EXISTS idx_passkey_credentials_credential_id ON passkey_credentials(credential_id);
`
		if err := tx.Exec(createPasskeyCredentialsTable).Error; err != nil {
			return fmt.Errorf("create passkey_credentials table failed: %w", err)
		}

		createSystemConfigsTable := `
CREATE TABLE IF NOT EXISTS system_configs (
    key        VARCHAR(255) PRIMARY KEY,
    value      TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`
		if err := tx.Exec(createSystemConfigsTable).Error; err != nil {
			return fmt.Errorf("create system_configs table failed: %w", err)
		}

		createLoginAttemptsTable := `
CREATE TABLE IF NOT EXISTS login_attempts (
    id         BIGSERIAL PRIMARY KEY,
    ip_address VARCHAR(45) NOT NULL,
    username   VARCHAR(100) NOT NULL,
    success    BOOLEAN NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_login_attempts_ip_created ON login_attempts(ip_address, created_at);
CREATE INDEX IF NOT EXISTS idx_login_attempts_username_created ON login_attempts(username, created_at);
`
		if err := tx.Exec(createLoginAttemptsTable).Error; err != nil {
			return fmt.Errorf("create login_attempts table failed: %w", err)
		}

		createSecurityEventLogsTable := `
CREATE TABLE IF NOT EXISTS security_event_logs (
	id         BIGSERIAL PRIMARY KEY,
	category   VARCHAR(32) NOT NULL,
	level      VARCHAR(16) NOT NULL,
	action     VARCHAR(64) NOT NULL,
	message    VARCHAR(255) NOT NULL,
	method     VARCHAR(16),
	path       VARCHAR(512),
	ip_address VARCHAR(45),
	username   VARCHAR(100),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_security_event_logs_created_at ON security_event_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_security_event_logs_action ON security_event_logs(action);
CREATE INDEX IF NOT EXISTS idx_security_event_logs_ip_created ON security_event_logs(ip_address, created_at);
CREATE INDEX IF NOT EXISTS idx_security_event_logs_user_created ON security_event_logs(username, created_at);
`
		if err := tx.Exec(createSecurityEventLogsTable).Error; err != nil {
			return fmt.Errorf("create security_event_logs table failed: %w", err)
		}

		createSessionsTable := `
CREATE TABLE IF NOT EXISTS sessions (
    id         BIGSERIAL PRIMARY KEY,
    token_hash VARCHAR(128) NOT NULL UNIQUE,
    user_id    BIGINT NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    last_used  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_sessions_token_hash ON sessions(token_hash);
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
`
		if err := tx.Exec(createSessionsTable).Error; err != nil {
			return fmt.Errorf("create sessions table failed: %w", err)
		}

		createAPITokensTable := `
CREATE TABLE IF NOT EXISTS api_tokens (
    id            BIGSERIAL PRIMARY KEY,
    user_id       BIGINT NOT NULL DEFAULT 1,
    name          VARCHAR(255) NOT NULL,
	token_type    VARCHAR(32) NOT NULL DEFAULT 'full',
    token_hash    VARCHAR(128) NOT NULL UNIQUE,
    ip_allowlist  JSONB,
    last_used_at  TIMESTAMPTZ,
    last_used_ip  VARCHAR(45),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_api_tokens_token_hash ON api_tokens(token_hash);
`
		if err := tx.Exec(createAPITokensTable).Error; err != nil {
			return fmt.Errorf("create api_tokens table failed: %w", err)
		}

		alterAPITokensTable := `
ALTER TABLE api_tokens ADD COLUMN IF NOT EXISTS token_type VARCHAR(32) NOT NULL DEFAULT 'full';
`
		if err := tx.Exec(alterAPITokensTable).Error; err != nil {
			return fmt.Errorf("alter api_tokens table failed: %w", err)
		}

		createAPITokenLogsTable := `
CREATE TABLE IF NOT EXISTS api_token_logs (
	id          BIGSERIAL PRIMARY KEY,
	token_id    BIGINT NOT NULL,
	token_name  VARCHAR(255) NOT NULL,
	token_type  VARCHAR(32) NOT NULL,
	action      VARCHAR(64) NOT NULL,
	method      VARCHAR(16),
	path        VARCHAR(512),
	ip_address  VARCHAR(45),
	user_agent  VARCHAR(512),
	image_hash  VARCHAR(64),
	created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_api_token_logs_token_id ON api_token_logs(token_id);
CREATE INDEX IF NOT EXISTS idx_api_token_logs_created_at ON api_token_logs(created_at);
`
		if err := tx.Exec(createAPITokenLogsTable).Error; err != nil {
			return fmt.Errorf("create api_token_logs table failed: %w", err)
		}

		log.Infof("ensured images, image_routes, users, passkey_credentials, system_configs, login_attempts, security_event_logs, sessions, api_tokens and api_token_logs tables exist")
		return nil
	})
}

func main() {
	vips.Startup(nil)
	defer vips.Shutdown()
	cfg := config.Load()
	log := logger.Register("main")

	if err := ensureDatabase(cfg, log); err != nil {
		log.Fatalf("ensure database failed: %v", err)
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName, cfg.DBSSL,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("connect postgres failed: %v", err)
	}

	if err := ensureTables(db, log); err != nil {
		log.Fatalf("ensure tables failed: %v", err)
	}

	cleanupCtx, cleanupCancel := context.WithCancel(context.Background())
	defer cleanupCancel()

	go func(ctx context.Context) {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Infof("session cleanup worker stopped")
				return
			case <-ticker.C:
				if err := model.CleanExpiredSessions(db); err != nil {
					log.Errorf("clean expired sessions failed: %v", err)
				}
				if err := model.CleanOldLoginAttempts(db); err != nil {
					log.Errorf("clean old login attempts failed: %v", err)
				}
			}
		}
	}(cleanupCtx)

	gin.SetMode(gin.ReleaseMode)
	r, err := httpserver.NewRouter(cfg, db)
	if err != nil {
		log.Fatalf("init router failed: %v", err)
	}

	if err := r.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		log.Fatalf("set trusted proxies failed: %v", err)
	}

	httpServer := &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: r,
	}

	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- httpServer.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	log.Infof("AnzuImg backend listening on %s", cfg.ServerAddr)

	select {
	case err := <-serverErrCh:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("server run failed: %v", err)
		}
	case sig := <-sigCh:
		log.Infof("received shutdown signal: %s", sig.String())
		cleanupCancel()

		shutdownTimeoutSec := cfg.ShutdownTimeoutSec
		if shutdownTimeoutSec <= 0 {
			shutdownTimeoutSec = 10
		}

		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(shutdownTimeoutSec)*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Errorf("server graceful shutdown failed: %v", err)
			if closeErr := httpServer.Close(); closeErr != nil {
				log.Errorf("server force close failed: %v", closeErr)
			}
		}
	}
}
