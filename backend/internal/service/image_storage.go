package service

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/TangTangChu/AnzuImg/backend/internal/config"
	"github.com/TangTangChu/AnzuImg/backend/internal/logger"
)

type ImageStorage struct {
	cfg *config.Config
	log *logger.Logger
}

func NewImageStorage(cfg *config.Config, log *logger.Logger) *ImageStorage {
	return &ImageStorage{cfg: cfg, log: log}
}

// SaveByHash：按 hash 前两位分目录存储。
// 返回 relPath/mime/size
func (st *ImageStorage) SaveByHash(hash string, buf []byte) (string, string, int64, error) {
	if err := os.MkdirAll(st.cfg.StorageBase, 0o755); err != nil {
		return "", "", 0, fmt.Errorf("mkdir storage base failed: %w", err)
	}

	subdir := filepath.Join(st.cfg.StorageBase, hash[:2])
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		return "", "", 0, fmt.Errorf("mkdir subdir failed: %w", err)
	}

	absPath := filepath.Join(subdir, hash)
	relPath := filepath.Join(hash[:2], hash)

	if err := os.WriteFile(absPath, buf, 0o644); err != nil {
		return "", "", 0, fmt.Errorf("write file failed: %w", err)
	}

	mimeType := http.DetectContentType(buf)
	return relPath, mimeType, int64(len(buf)), nil
}

func (st *ImageStorage) AbsPath(rel string) string {
	return filepath.Join(st.cfg.StorageBase, rel)
}
