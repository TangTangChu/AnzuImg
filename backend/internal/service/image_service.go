package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/TangTangChu/AnzuImg/backend/internal/config"
	"github.com/TangTangChu/AnzuImg/backend/internal/logger"
	"github.com/TangTangChu/AnzuImg/backend/internal/model"
)

type ImageService struct {
	cfg *config.Config
	db  *gorm.DB
	log *logger.Logger

	storage Storage
}

type TagCount struct {
	Tag   string `json:"tag"`
	Count int64  `json:"count"`
}

func NewImageService(cfg *config.Config, db *gorm.DB) *ImageService {
	factory := NewStorageFactory(cfg, logger.Register("storage-factory"))
	storage := factory.CreateDefaultStorage()

	return &ImageService{
		cfg:     cfg,
		db:      db,
		log:     logger.Register("image"),
		storage: storage,
	}
}

// NewImageServiceWithStorage 使用指定的存储创建服务
func NewImageServiceWithStorage(cfg *config.Config, db *gorm.DB, storage Storage) *ImageService {
	return &ImageService{
		cfg:     cfg,
		db:      db,
		log:     logger.Register("image"),
		storage: storage,
	}
}

type UploadResult struct {
	Image    model.Image
	Reused   bool
	Route    string // 为空表示未映射
	HashURL  string
	RouteURL string // 为空表示未映射
}

// Upload 上传图片
// fileName 参数：显示用的文件名
// mimeType 参数：调用者提供的MIME类型
// width, height 参数：调用者提供的图片尺寸，如果是图片的话
func (s *ImageService) Upload(ctx context.Context, buf []byte, fileName string, routes []string, description string, tags []string, mimeType string, width, height int, convert bool, targetFormat string, quality int, effort int, uploadedByTokenID *uint, uploadedByTokenName string, uploadedByTokenType string) (*UploadResult, error) {
	// 如果需要转换
	if convert {
		newBuf, newMime, err := ConvertImage(bytes.NewReader(buf), targetFormat, quality, effort)
		if err != nil {
			return nil, fmt.Errorf("convert image failed: %w", err)
		}
		buf = newBuf
		mimeType = newMime

		// 更新文件名后缀
		ext := "." + targetFormat
		if idx := strings.LastIndex(fileName, "."); idx != -1 {
			fileName = fileName[:idx]
		}
		fileName = fileName + ext

		// 重新检测尺寸
		if w, h, err := DetectImageDimensions(buf); err == nil {
			width = w
			height = h
		}
	}

	tagsBytes, err := json.Marshal(tags)
	if err != nil {
		return nil, fmt.Errorf("marshal tags failed: %w", err)
	}
	tagsJSON := datatypes.JSON(tagsBytes)
	sum := sha256.Sum256(buf)
	hashStr := hex.EncodeToString(sum[:])

	// 按 hash 去重
	var existing model.Image
	if err := s.db.Where("hash = ?", hashStr).First(&existing).Error; err == nil {
		if len(routes) > 0 {
			if err := s.db.Transaction(func(tx *gorm.DB) error {
				for _, rStr := range routes {
					if rStr == "" {
						continue
					}
					r := model.ImageRoute{ImageID: existing.ID, Route: rStr}
					if err := tx.Create(&r).Error; err != nil {
						return fmt.Errorf("route insert failed: %w", err)
					}
				}
				return nil
			}); err != nil {
				return nil, err
			}
		}

		firstRoute := ""
		if len(routes) > 0 {
			firstRoute = routes[0]
		}

		return &UploadResult{
			Image:    existing,
			Reused:   true,
			Route:    firstRoute,
			HashURL:  "/i/" + existing.Hash,
			RouteURL: routeURL(firstRoute),
		}, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("db query failed: %w", err)
	}

	relPath, size, err := s.storage.Save(ctx, hashStr, buf, mimeType)
	if err != nil {
		return nil, err
	}

	if IsImageFile(mimeType) {
		if thumbData, err := GenerateThumbnail(bytes.NewReader(buf), 800, 800); err == nil {
			// LocalStorage: hash[:2]/hash_thumb.webp
			_, _, err := s.storage.Save(ctx, hashStr+"_thumb.webp", thumbData, "image/webp")
			if err != nil {
				s.log.Warnf("Failed to save thumbnail: %v", err)
			}
		} else {
			s.log.Warnf("Failed to generate thumbnail: %v", err)
		}
	}

	img := model.Image{
		Hash:                hashStr,
		FileName:            fileName,
		MimeType:            mimeType,
		Size:                size,
		Path:                relPath,
		Width:               width,
		Height:              height,
		Description:         description,
		Tags:                tagsJSON,
		UploadedByTokenID:   uploadedByTokenID,
		UploadedByTokenName: uploadedByTokenName,
		UploadedByTokenType: uploadedByTokenType,
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&img).Error; err != nil {
			return fmt.Errorf("db save image failed: %w", err)
		}

		for _, rStr := range routes {
			if rStr == "" {
				continue
			}
			r := model.ImageRoute{ImageID: img.ID, Route: rStr}
			if err := tx.Create(&r).Error; err != nil {
				return fmt.Errorf("route insert failed: %w", err)
			}
		}

		return nil
	}); err != nil {
		if delErr := s.storage.Delete(ctx, relPath); delErr != nil {
			s.log.Warnf("Failed to cleanup file after db rollback: %v", delErr)
		}
		suffixes := []string{"_thumb.webp", "_thumb.jpg", "_thumb"}
		for _, suffix := range suffixes {
			_ = s.storage.Delete(ctx, relPath+suffix)
		}
		return nil, err
	}

	firstRoute := ""
	if len(routes) > 0 {
		firstRoute = routes[0]
	}

	return &UploadResult{
		Image:    img,
		Reused:   false,
		Route:    firstRoute,
		HashURL:  "/i/" + img.Hash,
		RouteURL: routeURL(firstRoute),
	}, nil
}

func routeURL(route string) string {
	if route == "" {
		return ""
	}
	return "/i/r/" + route
}

// ResolveByHash：返回图片和绝对路径或访问URL。
func (s *ImageService) ResolveByHash(ctx context.Context, hash string) (*model.Image, string, error) {
	var img model.Image
	if err := s.db.Where("hash = ?", hash).First(&img).Error; err != nil {
		return nil, "", err
	}
	absPath, err := s.storage.GetAbsPath(ctx, img.Path)
	if err != nil {
		return nil, "", err
	}
	return &img, absPath, nil
}

// ResolveThumbnailByHash：返回图片缩略图的绝对路径或访问URL。
func (s *ImageService) ResolveThumbnailByHash(ctx context.Context, hash string) (string, error) {
	var img model.Image
	if err := s.db.Where("hash = ?", hash).First(&img).Error; err != nil {
		return "", err
	}
	suffixes := []string{"_thumb.webp", "_thumb.jpg", "_thumb"}

	for _, suffix := range suffixes {
		thumbPath := img.Path + suffix
		exists, err := s.storage.Exists(ctx, thumbPath)
		if err == nil && exists {
			return s.storage.GetAbsPath(ctx, thumbPath)
		}
	}

	// 如果缩略图都不存在，返回原图
	return s.storage.GetAbsPath(ctx, img.Path)
}

func (s *ImageService) ResolveByRoute(ctx context.Context, route string) (*model.Image, string, error) {
	var r model.ImageRoute
	if err := s.db.Where("route = ?", route).First(&r).Error; err != nil {
		return nil, "", err
	}

	var img model.Image
	if err := s.db.First(&img, r.ImageID).Error; err != nil {
		return nil, "", err
	}

	absPath, err := s.storage.GetAbsPath(ctx, img.Path)
	if err != nil {
		return nil, "", err
	}
	return &img, absPath, nil
}

// ListImages 分页获取图片列表
func (s *ImageService) ListImages(page, pageSize int, tag string, fileName string) ([]model.Image, int64, error) {
	var images []model.Image
	var total int64

	query := s.db.Model(&model.Image{})

	if tag != "" {
		// 转义JSON字符串中的特殊字符，防止JSON注入
		escapedTag := strings.ReplaceAll(tag, `\`, `\\`)
		escapedTag = strings.ReplaceAll(escapedTag, `"`, `\"`)
		query = query.Where("tags @> ?", fmt.Sprintf(`["%s"]`, escapedTag))
	}

	if fileName != "" {
		// 转义LIKE模式中的特殊字符
		escapedFileName := strings.ReplaceAll(fileName, `\`, `\\`)
		escapedFileName = strings.ReplaceAll(escapedFileName, `%`, `\%`)
		escapedFileName = strings.ReplaceAll(escapedFileName, `_`, `\_`)
		query = query.Where("file_name ILIKE ?", "%"+escapedFileName+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&images).Error; err != nil {
		return nil, 0, err
	}

	return images, total, nil
}

// ListTags 获取标签列表（按数量排序）
func (s *ImageService) ListTags(limit int) ([]TagCount, error) {
	if limit <= 0 {
		limit = 200
	}

	if s.db.Dialector.Name() == "postgres" {
		var tags []TagCount
		err := s.db.Raw(`
			SELECT tag, COUNT(*) AS count
			FROM (
				SELECT jsonb_array_elements_text(tags) AS tag
				FROM images
			) t
			GROUP BY tag
			ORDER BY count DESC, tag ASC
			LIMIT ?`, limit).Scan(&tags).Error
		return tags, err
	}

	var images []model.Image
	if err := s.db.Select("tags").Find(&images).Error; err != nil {
		return nil, err
	}

	counts := make(map[string]int64)
	for _, img := range images {
		if len(img.Tags) == 0 {
			continue
		}
		var tagList []string
		if err := json.Unmarshal(img.Tags, &tagList); err != nil {
			continue
		}
		for _, tag := range tagList {
			if strings.TrimSpace(tag) == "" {
				continue
			}
			counts[tag]++
		}
	}

	var result []TagCount
	for tag, count := range counts {
		result = append(result, TagCount{Tag: tag, Count: count})
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Count == result[j].Count {
			return result[i].Tag < result[j].Tag
		}
		return result[i].Count > result[j].Count
	})

	if len(result) > limit {
		result = result[:limit]
	}

	return result, nil
}

// ListRoutes 分页获取路由信息
func (s *ImageService) ListRoutes(page, pageSize int) ([]model.ImageRoute, int64, error) {
	var routes []model.ImageRoute
	var total int64

	query := s.db.Model(&model.ImageRoute{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("Image").Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&routes).Error; err != nil {
		return nil, 0, err
	}

	return routes, total, nil
}

// DeleteRoute 删除指定路由
func (s *ImageService) DeleteRoute(route string) error {
	return s.db.Where("route = ?", route).Delete(&model.ImageRoute{}).Error
}

// UpdateRoutes 更新图片的路由
func (s *ImageService) UpdateRoutes(imageID uint64, routes []string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除旧路由
		if err := tx.Where("image_id = ?", imageID).Delete(&model.ImageRoute{}).Error; err != nil {
			return err
		}

		// 添加新路由
		for _, r := range routes {
			if r == "" {
				continue
			}
			route := model.ImageRoute{
				ImageID: imageID,
				Route:   r,
			}
			if err := tx.Create(&route).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteImage 删除图片
func (s *ImageService) DeleteImage(ctx context.Context, hash string) error {
	var img model.Image
	if err := s.db.Where("hash = ?", hash).First(&img).Error; err != nil {
		return err
	}

	if err := s.storage.Delete(ctx, img.Path); err != nil {
		s.log.Warnf("Failed to delete file from storage: %v", err)
	}

	suffixes := []string{"_thumb.webp", "_thumb.jpg", "_thumb"}
	for _, suffix := range suffixes {
		if err := s.storage.Delete(ctx, img.Path+suffix); err != nil {
			s.log.Debugf("Failed to delete thumbnail %s: %v", suffix, err)
		}
	}

	if err := s.db.Delete(&img).Error; err != nil {
		return fmt.Errorf("failed to delete image from db: %w", err)
	}

	return nil
}

// UpdateImage 更新图片信息
func (s *ImageService) UpdateImage(hash string, description string, tags []string, fileName string) (*model.Image, error) {
	var img model.Image
	if err := s.db.Where("hash = ?", hash).First(&img).Error; err != nil {
		return nil, err
	}

	img.Description = description
	if fileName != "" {
		img.FileName = fileName
	}

	tagsBytes, err := json.Marshal(tags)
	if err != nil {
		return nil, fmt.Errorf("marshal tags failed: %w", err)
	}
	img.Tags = datatypes.JSON(tagsBytes)

	fields := []string{"Description", "Tags"}
	if fileName != "" {
		fields = append(fields, "FileName")
	}
	if err := s.db.Model(&img).Select(fields).Updates(img).Error; err != nil {
		return nil, fmt.Errorf("failed to update image: %w", err)
	}

	return &img, nil
}

func (s *ImageService) DB() *gorm.DB {
	return s.db
}

func (s *ImageService) Config() *config.Config {
	return s.cfg
}
