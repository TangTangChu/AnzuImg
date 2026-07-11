package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
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

	storage        Storage
	uploadQueue    chan uploadTaskJob
	thumbnailQueue chan thumbnailJob
	uploadLocks    [256]sync.Mutex
}

type uploadTaskJob struct {
	TaskID string
	Input  UploadTaskInput
}

type UploadTaskInput struct {
	Buf                 []byte
	TempPath            string
	FileName            string
	Routes              []string
	Description         string
	Tags                []string
	MimeType            string
	Width               int
	Height              int
	Convert             bool
	TargetFormat        string
	Quality             int
	Effort              int
	UploadedByTokenID   *uint
	UploadedByTokenName string
	UploadedByTokenType string
	RequestMethod       string
	RequestPath         string
	IPAddress           string
	UserAgent           string
}

type thumbnailJob struct {
	Hash     string
	TempPath string
	MIMEType string
}

const (
	maxProcessedMediaDimension = 32768
	maxProcessedMediaPixels    = int64(100_000_000)
)

func processedMediaDimensionsAllowed(width, height int) bool {
	if width <= 0 || height <= 0 {
		return true
	}
	if width > maxProcessedMediaDimension || height > maxProcessedMediaDimension {
		return false
	}
	return int64(width)*int64(height) <= maxProcessedMediaPixels
}

type TagCount struct {
	Tag   string `json:"tag"`
	Count int64  `json:"count"`
}

func NewImageService(cfg *config.Config, db *gorm.DB) *ImageService {
	factory := NewStorageFactory(cfg, logger.Register("storage-factory"))
	storage := factory.CreateDefaultStorage()

	svc := &ImageService{
		cfg:     cfg,
		db:      db,
		log:     logger.Register("image"),
		storage: storage,
	}
	svc.startUploadWorkers(2, 8)
	svc.startThumbnailWorkers(2, 4)
	return svc
}

// NewImageServiceWithStorage 使用指定的存储创建服务
func NewImageServiceWithStorage(cfg *config.Config, db *gorm.DB, storage Storage) *ImageService {
	svc := &ImageService{
		cfg:     cfg,
		db:      db,
		log:     logger.Register("image"),
		storage: storage,
	}
	svc.startUploadWorkers(2, 8)
	svc.startThumbnailWorkers(2, 4)
	return svc
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
	if convert && IsImageFile(mimeType) {
		newBuf, newMime, err := ConvertImage(ctx, buf, mimeType, targetFormat, quality, effort)
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

	durationSeconds := 0
	videoCodec := ""
	videoBitrate := int64(0)
	audioCodec := ""
	audioBitrate := int64(0)
	if IsVideoFile(mimeType) {
		if info, err := ProbeVideoInfo(ctx, buf); err == nil {
			if width <= 0 {
				width = info.Width
			}
			if height <= 0 {
				height = info.Height
			}
			durationSeconds = info.DurationSeconds
			videoCodec = info.VideoCodec
			videoBitrate = info.VideoBitrate
			audioCodec = info.AudioCodec
			audioBitrate = info.AudioBitrate
		} else {
			s.log.Ctx(ctx).Warnf("Failed to probe video metadata: %v", err)
		}
	}
	if !processedMediaDimensionsAllowed(width, height) {
		return nil, errors.New("media dimensions exceed limit")
	}

	tagsBytes, err := json.Marshal(tags)
	if err != nil {
		return nil, fmt.Errorf("marshal tags failed: %w", err)
	}
	tagsJSON := datatypes.JSON(tagsBytes)
	sum := sha256.Sum256(buf)
	hashStr := hex.EncodeToString(sum[:])
	contentLock := &s.uploadLocks[int(sum[0])]
	contentLock.Lock()
	defer contentLock.Unlock()

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

	img := model.Image{
		Hash:                hashStr,
		FileName:            fileName,
		MimeType:            mimeType,
		Size:                size,
		Path:                relPath,
		Width:               width,
		Height:              height,
		DurationSeconds:     durationSeconds,
		VideoCodec:          videoCodec,
		VideoBitrate:        videoBitrate,
		AudioCodec:          audioCodec,
		AudioBitrate:        audioBitrate,
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
			s.log.Ctx(ctx).Warnf("Failed to cleanup file after db rollback: %v", delErr)
		}
		suffixes := []string{"_thumb.webp", "_thumb.jpg", "_thumb"}
		for _, suffix := range suffixes {
			_ = s.storage.Delete(ctx, relPath+suffix)
		}
		return nil, err
	}

	s.enqueueThumbnail(hashStr, buf, mimeType)

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

func writeTempData(pattern string, data []byte) (string, error) {
	tmp, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", err
	}
	path := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(path)
		return "", err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(path)
		return "", err
	}
	return path, nil
}

func (s *ImageService) enqueueThumbnail(hashStr string, buf []byte, mimeType string) {
	if !IsImageFile(mimeType) && !IsVideoFile(mimeType) {
		return
	}
	tempPath, err := writeTempData("anzuimg-thumbnail-*", buf)
	if err != nil {
		s.log.Warnf("Failed to stage thumbnail input: %v", err)
		return
	}
	job := thumbnailJob{Hash: hashStr, TempPath: tempPath, MIMEType: mimeType}
	select {
	case s.thumbnailQueue <- job:
	default:
		_ = os.Remove(tempPath)
		s.log.Warnf("Thumbnail queue full; skipped thumbnail for %s", hashStr)
	}
}

func (s *ImageService) startThumbnailWorkers(workerCount, queueSize int) {
	if workerCount <= 0 {
		workerCount = 1
	}
	if queueSize <= 0 {
		queueSize = 4
	}
	s.thumbnailQueue = make(chan thumbnailJob, queueSize)
	for i := 0; i < workerCount; i++ {
		go func() {
			for job := range s.thumbnailQueue {
				s.runThumbnailJob(job)
			}
		}()
	}
}

func (s *ImageService) runThumbnailJob(job thumbnailJob) {
	defer func() { _ = os.Remove(job.TempPath) }()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	data, err := os.ReadFile(job.TempPath)
	if err != nil {
		s.log.Ctx(ctx).Warnf("Failed to read thumbnail input: %v", err)
		return
	}
	if IsImageFile(job.MIMEType) {
		if thumbData, err := GenerateThumbnail(bytes.NewReader(data), 800, 800); err == nil {
			if _, _, err := s.storage.Save(ctx, job.Hash+"_thumb.webp", thumbData, "image/webp"); err != nil {
				s.log.Ctx(ctx).Warnf("Failed to save thumbnail: %v", err)
			}
		} else {
			s.log.Ctx(ctx).Warnf("Failed to generate thumbnail: %v", err)
		}
		return
	}
	if thumbData, err := GenerateVideoThumbnail(ctx, data, 800, 800); err == nil {
		if _, _, err := s.storage.Save(ctx, job.Hash+"_thumb.jpg", thumbData, "image/jpeg"); err != nil {
			s.log.Ctx(ctx).Warnf("Failed to save video thumbnail: %v", err)
		}
	} else {
		s.log.Ctx(ctx).Warnf("Failed to generate video thumbnail: %v", err)
	}
}

func (s *ImageService) startUploadWorkers(workerCount int, queueSize int) {
	if workerCount <= 0 {
		workerCount = 1
	}
	if queueSize <= 0 {
		queueSize = 16
	}

	s.uploadQueue = make(chan uploadTaskJob, queueSize)
	for i := 0; i < workerCount; i++ {
		go func() {
			for job := range s.uploadQueue {
				s.runUploadTask(job)
			}
		}()
	}
}

func (s *ImageService) EnqueueUploadTask(input UploadTaskInput) (*model.UploadTask, error) {
	if s.uploadQueue == nil {
		s.startUploadWorkers(2, 8)
	}

	tempPath, err := writeTempData("anzuimg-upload-task-*", input.Buf)
	if err != nil {
		return nil, fmt.Errorf("stage upload task failed: %w", err)
	}
	input.Buf = nil
	input.TempPath = tempPath
	task := model.UploadTask{
		ID:       uuid.NewString(),
		Status:   model.UploadTaskStatusPending,
		FileName: input.FileName,
	}
	if err := s.db.Create(&task).Error; err != nil {
		_ = os.Remove(tempPath)
		return nil, fmt.Errorf("create upload task failed: %w", err)
	}

	select {
	case s.uploadQueue <- uploadTaskJob{TaskID: task.ID, Input: input}:
		return &task, nil
	default:
		_ = os.Remove(tempPath)
		now := time.Now()
		updates := map[string]interface{}{
			"status":        model.UploadTaskStatusFailed,
			"error_code":    "queue_full",
			"error_message": "upload queue is full",
			"completed_at":  &now,
		}
		_ = s.db.Model(&model.UploadTask{}).Where("id = ?", task.ID).Updates(updates).Error
		task.Status = model.UploadTaskStatusFailed
		task.ErrorCode = "queue_full"
		task.ErrorMessage = "upload queue is full"
		task.CompletedAt = &now
		return &task, nil
	}
}

func (s *ImageService) GetUploadTask(id string) (*model.UploadTask, error) {
	var task model.UploadTask
	if err := s.db.Where("id = ?", id).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *ImageService) runUploadTask(job uploadTaskJob) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	_ = s.db.Model(&model.UploadTask{}).Where("id = ?", job.TaskID).Updates(map[string]interface{}{
		"status": model.UploadTaskStatusRunning,
	}).Error

	input := job.Input
	defer func() { _ = os.Remove(input.TempPath) }()
	buf, readErr := os.ReadFile(input.TempPath)
	if readErr != nil {
		now := time.Now()
		_ = s.db.Model(&model.UploadTask{}).Where("id = ?", job.TaskID).Updates(map[string]interface{}{
			"status":        model.UploadTaskStatusFailed,
			"error_code":    "upload_input_unavailable",
			"error_message": "upload input unavailable",
			"completed_at":  &now,
		}).Error
		s.log.Ctx(ctx).Warnf("Failed to read upload task input %s: %v", job.TaskID, readErr)
		return
	}
	res, err := s.Upload(
		ctx,
		buf,
		input.FileName,
		input.Routes,
		input.Description,
		input.Tags,
		input.MimeType,
		input.Width,
		input.Height,
		input.Convert,
		input.TargetFormat,
		input.Quality,
		input.Effort,
		input.UploadedByTokenID,
		input.UploadedByTokenName,
		input.UploadedByTokenType,
	)

	now := time.Now()
	if err != nil {
		if updateErr := s.db.Model(&model.UploadTask{}).Where("id = ?", job.TaskID).Updates(map[string]interface{}{
			"status":        model.UploadTaskStatusFailed,
			"error_code":    "upload_failed",
			"error_message": "upload processing failed",
			"completed_at":  &now,
		}).Error; updateErr != nil {
			s.log.Ctx(ctx).Warnf("Failed to update failed upload task %s: %v", job.TaskID, updateErr)
		}
		s.log.Ctx(ctx).Warnf("Upload task %s failed: %v", job.TaskID, err)
		return
	}

	if input.UploadedByTokenID != nil {
		tokenSvc := NewAPITokenService(s.cfg, s.db)
		_ = tokenSvc.RecordLog(&model.APITokenLog{
			TokenID:   *input.UploadedByTokenID,
			TokenName: input.UploadedByTokenName,
			TokenType: input.UploadedByTokenType,
			Action:    "image_upload",
			Method:    input.RequestMethod,
			Path:      input.RequestPath,
			IPAddress: input.IPAddress,
			UserAgent: input.UserAgent,
			ImageHash: res.Image.Hash,
		})
	}

	result := map[string]interface{}{
		"success":          true,
		"hash":             res.Image.Hash,
		"file_name":        res.Image.FileName,
		"size":             res.Image.Size,
		"mime":             res.Image.MimeType,
		"path":             res.Image.Path,
		"width":            res.Image.Width,
		"height":           res.Image.Height,
		"duration_seconds": res.Image.DurationSeconds,
		"video_codec":      res.Image.VideoCodec,
		"video_bitrate":    res.Image.VideoBitrate,
		"audio_codec":      res.Image.AudioCodec,
		"audio_bitrate":    res.Image.AudioBitrate,
		"description":      res.Image.Description,
		"tags":             res.Image.Tags,
		"created_at":       res.Image.CreatedAt,
		"updated_at":       res.Image.UpdatedAt,
		"reused":           res.Reused,
		"url":              res.HashURL,
		"route":            res.Route,
		"route_url":        res.RouteURL,
	}
	resultBytes, err := json.Marshal(result)
	if err != nil {
		if updateErr := s.db.Model(&model.UploadTask{}).Where("id = ?", job.TaskID).Updates(map[string]interface{}{
			"status":        model.UploadTaskStatusFailed,
			"error_code":    "result_encode_failed",
			"error_message": "failed to encode upload result",
			"completed_at":  &now,
		}).Error; updateErr != nil {
			s.log.Ctx(ctx).Warnf("Failed to update result encode error for upload task %s: %v", job.TaskID, updateErr)
		}
		return
	}

	if updateErr := s.db.Model(&model.UploadTask{}).Where("id = ?", job.TaskID).Updates(map[string]interface{}{
		"status":       model.UploadTaskStatusSucceeded,
		"result":       datatypes.JSON(resultBytes),
		"completed_at": &now,
	}).Error; updateErr != nil {
		s.log.Ctx(ctx).Warnf("Failed to update succeeded upload task %s: %v", job.TaskID, updateErr)
	}
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
		s.log.Ctx(ctx).Warnf("Failed to delete file from storage: %v", err)
	}

	suffixes := []string{"_thumb.webp", "_thumb.jpg", "_thumb"}
	for _, suffix := range suffixes {
		if err := s.storage.Delete(ctx, img.Path+suffix); err != nil {
			s.log.Ctx(ctx).Debugf("Failed to delete thumbnail %s: %v", suffix, err)
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

// GetStats 获取系统统计信息
func (s *ImageService) GetStats() (*model.SystemStats, error) {
	var stats model.SystemStats
	var totalSize *int64 // 使用指针来处理 NULL 值

	// 统计图片总数
	if err := s.db.Model(&model.Image{}).Count(&stats.TotalImages).Error; err != nil {
		return nil, err
	}

	// 统计总大小
	if err := s.db.Model(&model.Image{}).Select("SUM(size)").Scan(&totalSize).Error; err != nil {
		return nil, err
	}

	if totalSize != nil {
		stats.TotalSize = *totalSize
	} else {
		stats.TotalSize = 0
	}

	// 统计过去24小时登录失败次数
	yesterday := time.Now().Add(-24 * time.Hour)
	if err := s.db.Model(&model.LoginAttempt{}).
		Where("created_at > ? AND success = ?", yesterday, false).
		Count(&stats.LoginFailures24h).Error; err != nil {
		return nil, err
	}

	// 统计过去24小时安全事件数
	if err := s.db.Model(&model.SecurityEventLog{}).
		Where("created_at > ?", yesterday).
		Count(&stats.SecurityEvents24h).Error; err != nil {
		return nil, err
	}

	return &stats, nil
}
