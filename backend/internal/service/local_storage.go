package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/TangTangChu/AnzuImg/backend/internal/config"
	"github.com/TangTangChu/AnzuImg/backend/internal/logger"
)

// LocalStorage 本地文件系统存储
type LocalStorage struct {
	cfg *config.Config
	log *logger.Logger
}

// NewLocalStorage 创建本地存储实例
func NewLocalStorage(cfg *config.Config, log *logger.Logger) *LocalStorage {
	return &LocalStorage{cfg: cfg, log: log}
}

// Save 保存图片到本地文件系统
func (s *LocalStorage) Save(ctx context.Context, hash string, data []byte, mimeType string) (string, int64, error) {
	if err := os.MkdirAll(s.cfg.StorageBase, 0o755); err != nil {
		return "", 0, fmt.Errorf("mkdir storage base failed: %w", err)
	}

	subdir := filepath.Join(s.cfg.StorageBase, hash[:2])
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		return "", 0, fmt.Errorf("mkdir subdir failed: %w", err)
	}

	absPath := filepath.Join(subdir, hash)
	relPath := filepath.Join(hash[:2], hash)

	if err := os.WriteFile(absPath, data, 0o644); err != nil {
		return "", 0, fmt.Errorf("write file failed: %w", err)
	}

	return relPath, int64(len(data)), nil
}

// GetAbsPath 根据相对路径获取绝对路径
func (s *LocalStorage) GetAbsPath(ctx context.Context, relPath string) (string, error) {
	return filepath.Join(s.cfg.StorageBase, relPath), nil
}

// Delete 删除文件
func (s *LocalStorage) Delete(ctx context.Context, relPath string) error {
	absPath, err := s.GetAbsPath(ctx, relPath)
	if err != nil {
		return err
	}

	if err := os.Remove(absPath); err != nil {
		if os.IsNotExist(err) {
			return nil // 文件不存在，视为删除成功
		}
		return fmt.Errorf("delete file failed: %w", err)
	}

	return nil
}

// Exists 检查文件是否存在
func (s *LocalStorage) Exists(ctx context.Context, relPath string) (bool, error) {
	absPath, err := s.GetAbsPath(ctx, relPath)
	if err != nil {
		return false, err
	}

	_, err = os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("check file exists failed: %w", err)
	}

	return true, nil
}

// Type 返回存储类型
func (s *LocalStorage) Type() string {
	return "local"
}
