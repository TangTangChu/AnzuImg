package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/TangTangChu/AnzuImg/backend/internal/config"
	httpserver "github.com/TangTangChu/AnzuImg/backend/internal/http"
	"github.com/TangTangChu/AnzuImg/backend/internal/logger"
)

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
		if err := adminDB.Exec("CREATE DATABASE " + cfg.DBName).Error; err != nil {
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
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_images_hash ON images(hash);
`
		if err := tx.Exec(createImagesTable).Error; err != nil {
			return fmt.Errorf("create images table failed: %w", err)
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

		log.Infof("ensured images and image_routes tables exist")
		return nil
	})
}

func main() {
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

	gin.SetMode(gin.ReleaseMode)
	r := httpserver.NewRouter(cfg, db)

	log.Infof("AnzuImg backend listening on %s", cfg.ServerAddr)
	if err := r.Run(cfg.ServerAddr); err != nil {
		log.Fatalf("server run failed: %v", err)
	}
}
