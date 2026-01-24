package service

import (
	"context"
)

// Storage 定义图床存储接口
type Storage interface {
	// Save 保存图片数据，返回存储路径和文件大小
	Save(ctx context.Context, hash string, data []byte, mimeType string) (path string, size int64, err error)
	// GetAbsPath 根据相对路径获取绝对路径或访问URL
	GetAbsPath(ctx context.Context, relPath string) (string, error)

	// Delete 删除指定路径的文件
	Delete(ctx context.Context, relPath string) error

	// Exists 检查文件是否存在
	Exists(ctx context.Context, relPath string) (bool, error)

	// Type 返回存储类型
	Type() string
}
