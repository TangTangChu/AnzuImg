package service

import (
	"fmt"

	"github.com/TangTangChu/AnzuImg/backend/internal/config"
	"github.com/TangTangChu/AnzuImg/backend/internal/logger"
)

// StorageType 存储类型枚举
type StorageType string

const (
	StorageTypeLocal StorageType = "local"
	StorageTypeCloud StorageType = "cloud"
)

// StorageFactory 存储工厂
type StorageFactory struct {
	cfg *config.Config
	log *logger.Logger
}

// NewStorageFactory 创建存储工厂
func NewStorageFactory(cfg *config.Config, log *logger.Logger) *StorageFactory {
	return &StorageFactory{cfg: cfg, log: log}
}

// CreateStorage 根据配置创建存储实例
func (f *StorageFactory) CreateStorage(storageType StorageType) (Storage, error) {
	switch storageType {
	case StorageTypeLocal:
		return NewLocalStorage(f.cfg, f.log), nil
	case StorageTypeCloud:
		return NewCloudStorage(f.cfg, f.log)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", storageType)
	}
}

// CreateDefaultStorage 创建默认存储
func (f *StorageFactory) CreateDefaultStorage() Storage {
	storageType := f.GetStorageTypeFromConfig()
	storage, err := f.CreateStorage(storageType)
	if err != nil {
		f.log.Errorf("Failed to create storage type %s: %v, falling back to local storage", storageType, err)
		return NewLocalStorage(f.cfg, f.log)
	}
	f.log.Infof("Created storage type: %s", storage.Type())
	return storage
}

// GetStorageTypeFromConfig 从配置中获取存储类型
func (f *StorageFactory) GetStorageTypeFromConfig() StorageType {
	storageType := f.cfg.StorageType
	switch storageType {
	case "local":
		return StorageTypeLocal
	case "cloud":
		return StorageTypeCloud
	default:
		f.log.Warnf("Unknown storage type: %s, using local storage", storageType)
		return StorageTypeLocal
	}
}
