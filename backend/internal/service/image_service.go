package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/TangTangChu/AnzuImg/backend/internal/config"
	"github.com/TangTangChu/AnzuImg/backend/internal/logger"
	"github.com/TangTangChu/AnzuImg/backend/internal/model"
)

type ImageService struct {
	cfg *config.Config
	db  *gorm.DB
	log *logger.Logger

	storage *ImageStorage
}

func NewImageService(cfg *config.Config, db *gorm.DB) *ImageService {
	return &ImageService{
		cfg:     cfg,
		db:      db,
		log:     logger.Register("image"),
		storage: NewImageStorage(cfg, logger.Register("image-storage")),
	}
}

type UploadResult struct {
	Image    model.Image
	Reused   bool
	Route    string // 为空表示未映射
	HashURL  string
	RouteURL string // 为空表示未映射
}

// Upload
func (s *ImageService) Upload(buf []byte, fileName string, route string) (*UploadResult, error) {
	sum := sha256.Sum256(buf)
	hashStr := hex.EncodeToString(sum[:])

	// 按 hash 去重
	var existing model.Image
	if err := s.db.Where("hash = ?", hashStr).First(&existing).Error; err == nil {
		if route != "" {
			r := model.ImageRoute{ImageID: existing.ID, Route: route}
			if err := s.db.Create(&r).Error; err != nil {
				return nil, fmt.Errorf("route insert failed: %w", err)
			}
		}

		return &UploadResult{
			Image:    existing,
			Reused:   true,
			Route:    route,
			HashURL:  "/i/" + existing.Hash,
			RouteURL: routeURL(route),
		}, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("db query failed: %w", err)
	}

	relPath, mimeType, size, err := s.storage.SaveByHash(hashStr, buf)
	if err != nil {
		return nil, err
	}

	img := model.Image{
		Hash:     hashStr,
		FileName: fileName,
		MimeType: mimeType,
		Size:     size,
		Path:     relPath,
	}

	if err := s.db.Create(&img).Error; err != nil {
		return nil, fmt.Errorf("db save image failed: %w", err)
	}

	if route != "" {
		r := model.ImageRoute{ImageID: img.ID, Route: route}
		if err := s.db.Create(&r).Error; err != nil {
			return nil, fmt.Errorf("route insert failed: %w", err)
		}
	}

	return &UploadResult{
		Image:    img,
		Reused:   false,
		Route:    route,
		HashURL:  "/i/" + img.Hash,
		RouteURL: routeURL(route),
	}, nil
}

func routeURL(route string) string {
	if route == "" {
		return ""
	}
	return "/i/r/" + route
}

// ResolveByHash：返回图片和绝对路径。
func (s *ImageService) ResolveByHash(hash string) (*model.Image, string, error) {
	var img model.Image
	if err := s.db.Where("hash = ?", hash).First(&img).Error; err != nil {
		return nil, "", err
	}
	absPath := s.storage.AbsPath(img.Path)
	return &img, absPath, nil
}

// ResolveByRoute：route 必须存在，否则报错；不 fallback。
func (s *ImageService) ResolveByRoute(route string) (*model.Image, string, error) {
	var r model.ImageRoute
	if err := s.db.Where("route = ?", route).First(&r).Error; err != nil {
		return nil, "", err
	}

	var img model.Image
	if err := s.db.First(&img, r.ImageID).Error; err != nil {
		return nil, "", err
	}

	absPath := s.storage.AbsPath(img.Path)
	return &img, absPath, nil
}
