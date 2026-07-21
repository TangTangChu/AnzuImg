package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/TangTangChu/AnzuImg/backend/internal/config"
	"github.com/TangTangChu/AnzuImg/backend/internal/http/middleware"
	"github.com/TangTangChu/AnzuImg/backend/internal/http/response"
	"github.com/TangTangChu/AnzuImg/backend/internal/model"
	"github.com/TangTangChu/AnzuImg/backend/internal/service"
)

type ImageHandler struct {
	svc        *service.ImageService
	urlFetcher *service.URLFetcher
}

var allowedUploadMIMETypes = map[string]struct{}{
	"image/jpeg":               {},
	"image/jpg":                {},
	"image/png":                {},
	"image/gif":                {},
	"image/webp":               {},
	"image/svg+xml":            {},
	"image/bmp":                {},
	"image/tiff":               {},
	"image/x-icon":             {},
	"image/vnd.microsoft.icon": {},
	"image/avif":               {},
	"image/jxl":                {},
	"image/heic":               {},
	"image/heif":               {},
	"video/mp4":                {},
	"video/webm":               {},
	"video/ogg":                {},
	"video/quicktime":          {},
	"video/x-matroska":         {},
}

const multipartMemoryThreshold int64 = 8 << 20

const (
	maxMediaDimension = 32768
	maxMediaPixels    = int64(100_000_000)
)

func minPositive(values ...int64) int64 {
	var result int64
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if result == 0 || value < result {
			result = value
		}
	}
	return result
}

func mediaDimensionsAllowed(width, height int) bool {
	if width <= 0 || height <= 0 {
		return true
	}
	if width > maxMediaDimension || height > maxMediaDimension {
		return false
	}
	return int64(width)*int64(height) <= maxMediaPixels
}

func detectUploadMIMEAndDimensions(buf []byte) (mimeType string, width, height int) {
	mimeType, width, height, err := service.InspectImage(bytes.NewReader(buf))
	if err == nil {
		return mimeType, width, height
	}

	detected := http.DetectContentType(buf)
	if idx := strings.Index(detected, ";"); idx >= 0 {
		detected = strings.TrimSpace(detected[:idx])
	}

	if detected == "application/octet-stream" && len(buf) > 0 {
		return "application/octet-stream", 0, 0
	}

	return detected, 0, 0
}

func NewImageHandler(cfg *config.Config, db *gorm.DB) *ImageHandler {
	return &ImageHandler{
		svc:        service.NewImageService(cfg, db),
		urlFetcher: service.NewURLFetcher(cfg),
	}
}

// POST /api/v1/images
// form-data: file=<file> (can be multiple), route=<optional>, description=<optional>, tags=<optional>
func (h *ImageHandler) Upload(c *gin.Context) {
	var uploaderToken *model.APIToken
	if v, ok := c.Get("api_token"); ok {
		if t, ok2 := v.(*model.APIToken); ok2 {
			uploaderToken = t
		}
	}
	maxTotal := int64(100 * 1024 * 1024)
	if h.svc != nil {
		if cfg := h.svc.Config(); cfg != nil {
			if v := cfg.Effective().MaxUploadBytes; v > 0 {
				maxTotal = v
			}
		}
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxTotal)

	memoryThreshold := minPositive(maxTotal, multipartMemoryThreshold)
	if err := c.Request.ParseMultipartForm(memoryThreshold); err != nil {
		response.WriteErrorCode(c, http.StatusBadRequest, "invalid_multipart_form", "file size exceeds limit or invalid form")
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		response.WriteErrorCode(c, http.StatusBadRequest, "invalid_multipart_form", "invalid multipart form")
		return
	}

	files := form.File["file"]

	// 文件数量限制
	maxFiles := 20
	if h.svc != nil {
		if cfg := h.svc.Config(); cfg != nil {
			if v := cfg.Effective().MaxUploadFiles; v > 0 {
				maxFiles = v
			}
		}
	}
	if maxFiles > 0 && len(files) > maxFiles {
		response.WriteErrorCode(c, http.StatusBadRequest, "too_many_files", "too many files")
		return
	}

	type URLSource struct {
		URL         string   `json:"url"`
		ClientIndex int      `json:"client_index"`
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
		Routes      []string `json:"routes"`
		CustomName  string   `json:"custom_name"`
	}
	var urlSources []URLSource
	if urlSourcesStr := c.PostForm("url_sources"); urlSourcesStr != "" {
		if err := json.Unmarshal([]byte(urlSourcesStr), &urlSources); err != nil {
			response.WriteErrorCode(c, http.StatusBadRequest, "invalid_url_sources", "invalid url_sources json")
			return
		}
	}
	if maxFiles > 0 && len(files)+len(urlSources) > maxFiles {
		response.WriteErrorCode(c, http.StatusBadRequest, "too_many_files", "too many files")
		return
	}
	if len(files) == 0 && len(urlSources) == 0 {
		response.WriteErrorCode(c, http.StatusBadRequest, "file_required", "file is required")
		return
	}

	routeStr := c.PostForm("route")
	description := c.PostForm("description")
	tagsStr := c.PostForm("tags")
	customName := c.PostForm("custom_name")

	convert, _ := strconv.ParseBool(c.PostForm("convert"))
	targetFormat := c.PostForm("target_format")
	quality, _ := strconv.Atoi(c.PostForm("quality"))
	effort, _ := strconv.Atoi(c.PostForm("effort"))

	// 解析标签
	var tags []string
	if tagsStr != "" {
		tagList := strings.Split(tagsStr, ",")
		for _, tag := range tagList {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				tags = append(tags, tag)
			}
		}
	}
	var routes []string
	if routeStr != "" {
		list := strings.Split(routeStr, ",")
		for _, r := range list {
			r = strings.TrimSpace(r)
			if r != "" {
				routes = append(routes, r)
			}
		}
	}
	type FileMetadata struct {
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
		Routes      []string `json:"routes"`
		CustomName  string   `json:"custom_name"`
		ClientIndex int      `json:"client_index"`
	}
	var metadataList []FileMetadata
	metadataStr := c.PostForm("metadata")
	if metadataStr != "" {
		if err := json.Unmarshal([]byte(metadataStr), &metadataList); err != nil {
			response.WriteErrorCode(c, http.StatusBadRequest, "invalid_metadata", "invalid metadata json")
			return
		}
	}

	var results []gin.H
	remainingTotal := maxTotal
	appendUploadError := func(clientIndex int, fileName, code, message string) {
		results = append(results, gin.H{
			"client_index": clientIndex,
			"success":      false,
			"file_name":    fileName,
			"code":         code,
			"message":      message,
		})
	}

	for i, fileHeader := range files {
		maxPerFile := int64(60 * 1024 * 1024)
		if h.svc != nil {
			if cfg := h.svc.Config(); cfg != nil {
				if v := cfg.Effective().MaxUploadFileBytes; v > 0 {
					maxPerFile = v
				}
			}
		}
		clientIndex := i
		if i < len(metadataList) && metadataList[i].ClientIndex >= 0 {
			clientIndex = metadataList[i].ClientIndex
		}

		if maxPerFile > 0 && fileHeader.Size > maxPerFile {
			appendUploadError(clientIndex, fileHeader.Filename, "file_too_large", "file too large")
			continue
		}

		f, err := fileHeader.Open()
		if err != nil {
			appendUploadError(clientIndex, fileHeader.Filename, "file_open_failed", "open file failed")
			continue
		}

		reader := io.Reader(f)
		if maxPerFile > 0 {
			reader = io.LimitReader(f, maxPerFile+1)
		}
		buf, err := io.ReadAll(reader)
		f.Close()
		if err != nil {
			appendUploadError(clientIndex, fileHeader.Filename, "file_read_failed", "read file failed")
			continue
		}
		if maxPerFile > 0 && int64(len(buf)) > maxPerFile {
			appendUploadError(clientIndex, fileHeader.Filename, "file_too_large", "file too large")
			continue
		}
		if remainingTotal > 0 && int64(len(buf)) > remainingTotal {
			appendUploadError(clientIndex, fileHeader.Filename, "upload_total_too_large", "total upload size exceeds limit")
			continue
		}
		remainingTotal -= int64(len(buf))

		mimeType, width, height := detectUploadMIMEAndDimensions(buf)

		if _, allowed := allowedUploadMIMETypes[mimeType]; !allowed {
			appendUploadError(clientIndex, fileHeader.Filename, "unsupported_file_type", "unsupported file type: "+mimeType)
			continue
		}
		if !mediaDimensionsAllowed(width, height) {
			appendUploadError(clientIndex, fileHeader.Filename, "media_dimensions_too_large", "media dimensions exceed limit")
			continue
		}

		// 防止路径遍历攻击
		originalFileName := fileHeader.Filename
		cleanPath := filepath.Clean(originalFileName)
		if strings.Contains(cleanPath, "..") || filepath.IsAbs(cleanPath) || strings.ContainsAny(cleanPath, "/\\") {
			originalFileName = filepath.Base(cleanPath)
		} else {
			originalFileName = cleanPath
		}

		if originalFileName == "" || originalFileName == "." || originalFileName == ".." {
			appendUploadError(clientIndex, fileHeader.Filename, "invalid_filename", "invalid filename")
			continue
		}

		currentDesc := description
		currentTags := tags
		currentRoutes := routes
		currentName := customName

		if i < len(metadataList) {
			meta := metadataList[i]
			currentDesc = meta.Description
			currentTags = meta.Tags
			currentRoutes = meta.Routes
			currentName = meta.CustomName
		}

		finalFileName := originalFileName
		if currentName != "" {
			cleanCustomName := filepath.Clean(currentName)
			if !strings.Contains(cleanCustomName, "..") && !filepath.IsAbs(cleanCustomName) && !strings.ContainsAny(cleanCustomName, "/\\") {
				cn := filepath.Base(cleanCustomName)
				if cn != "" && cn != "." && cn != ".." {
					finalFileName = cn
				}
			}
		}

		var uploadedByTokenID *uint
		var uploadedByTokenName string
		var uploadedByTokenType string
		if uploaderToken != nil {
			uploadedByTokenID = &uploaderToken.ID
			uploadedByTokenName = uploaderToken.Name
			uploadedByTokenType = uploaderToken.NormalizedType()
		}

		convertCurrent := convert && service.IsImageFile(mimeType)
		res, err := h.svc.Upload(c.Request.Context(), buf, finalFileName, currentRoutes, currentDesc, currentTags, mimeType, width, height, convertCurrent, targetFormat, quality, effort, uploadedByTokenID, uploadedByTokenName, uploadedByTokenType)
		if err != nil {
			appendUploadError(clientIndex, fileHeader.Filename, "upload_failed", "upload failed")
			continue
		}

		if uploaderToken != nil {
			tokenSvc := service.NewAPITokenService(h.svc.Config(), h.svc.DB())
			_ = tokenSvc.RecordLog(&model.APITokenLog{
				TokenID:   uploaderToken.ID,
				TokenName: uploaderToken.Name,
				TokenType: uploaderToken.NormalizedType(),
				Action:    "image_upload",
				Method:    c.Request.Method,
				Path:      c.Request.URL.Path,
				IPAddress: middleware.ClientIP(c),
				UserAgent: c.Request.UserAgent(),
				ImageHash: res.Image.Hash,
			})
		}

		results = append(results, gin.H{
			"client_index":     clientIndex,
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
			"success":          true,
		})
	}

	for i, urlSrc := range urlSources {
		clientIndex := len(files) + i
		if urlSrc.ClientIndex >= 0 {
			clientIndex = urlSrc.ClientIndex
		}

		rawURL := strings.TrimSpace(urlSrc.URL)
		if rawURL == "" {
			appendUploadError(clientIndex, "", "url_invalid", "url is required")
			continue
		}

		maxPerFile := int64(60 * 1024 * 1024)
		urlFetchMax := int64(60 * 1024 * 1024)
		if h.svc != nil {
			if cfg := h.svc.Config(); cfg != nil {
				eff := cfg.Effective()
				if eff.MaxUploadFileBytes > 0 {
					maxPerFile = eff.MaxUploadFileBytes
				}
				if eff.URLFetchMaxBytes > 0 {
					urlFetchMax = eff.URLFetchMaxBytes
				}
			}
		}
		fetchLimit := minPositive(maxPerFile, urlFetchMax, remainingTotal)
		if fetchLimit <= 0 {
			appendUploadError(clientIndex, rawURL, "upload_total_too_large", "total upload size exceeds limit")
			continue
		}

		fetchRes, err := h.urlFetcher.Fetch(c.Request.Context(), rawURL, fetchLimit)
		if err != nil {
			code, msg := classifyURLFetchError(err)
			appendUploadError(clientIndex, rawURL, code, msg)
			continue
		}

		if maxPerFile > 0 && int64(len(fetchRes.Body)) > maxPerFile {
			appendUploadError(clientIndex, rawURL, "file_too_large", "file too large")
			continue
		}
		remainingTotal -= int64(len(fetchRes.Body))

		mimeType, width, height := detectUploadMIMEAndDimensions(fetchRes.Body)
		if _, allowed := allowedUploadMIMETypes[mimeType]; !allowed {
			appendUploadError(clientIndex, rawURL, "unsupported_file_type", "unsupported file type: "+mimeType)
			continue
		}
		if !mediaDimensionsAllowed(width, height) {
			appendUploadError(clientIndex, rawURL, "media_dimensions_too_large", "media dimensions exceed limit")
			continue
		}

		candidateName := strings.TrimSpace(urlSrc.CustomName)
		if candidateName == "" {
			candidateName = fetchRes.Filename
		}
		if candidateName == "" {
			candidateName = "remote-file"
		}
		cleanPath := filepath.Clean(candidateName)
		if strings.Contains(cleanPath, "..") || filepath.IsAbs(cleanPath) || strings.ContainsAny(cleanPath, "/\\") {
			cleanPath = filepath.Base(cleanPath)
		}
		if cleanPath == "" || cleanPath == "." || cleanPath == ".." {
			appendUploadError(clientIndex, rawURL, "invalid_filename", "invalid filename")
			continue
		}
		finalFileName := cleanPath

		var uploadedByTokenID *uint
		var uploadedByTokenName string
		var uploadedByTokenType string
		if uploaderToken != nil {
			uploadedByTokenID = &uploaderToken.ID
			uploadedByTokenName = uploaderToken.Name
			uploadedByTokenType = uploaderToken.NormalizedType()
		}

		convertCurrent := convert && service.IsImageFile(mimeType)
		res, err := h.svc.Upload(c.Request.Context(), fetchRes.Body, finalFileName, urlSrc.Routes, urlSrc.Description, urlSrc.Tags, mimeType, width, height, convertCurrent, targetFormat, quality, effort, uploadedByTokenID, uploadedByTokenName, uploadedByTokenType)
		if err != nil {
			appendUploadError(clientIndex, rawURL, "upload_failed", "upload failed")
			continue
		}

		if uploaderToken != nil {
			tokenSvc := service.NewAPITokenService(h.svc.Config(), h.svc.DB())
			_ = tokenSvc.RecordLog(&model.APITokenLog{
				TokenID:   uploaderToken.ID,
				TokenName: uploaderToken.Name,
				TokenType: uploaderToken.NormalizedType(),
				Action:    "image_upload",
				Method:    c.Request.Method,
				Path:      c.Request.URL.Path,
				IPAddress: middleware.ClientIP(c),
				UserAgent: c.Request.UserAgent(),
				ImageHash: res.Image.Hash,
			})
		}

		results = append(results, gin.H{
			"client_index":     clientIndex,
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
			"source_url":       fetchRes.FinalURL,
			"success":          true,
		})
	}

	c.JSON(http.StatusOK, results)
}

// POST /api/v1/images/tasks
func (h *ImageHandler) UploadTask(c *gin.Context) {
	var uploaderToken *model.APIToken
	if v, ok := c.Get("api_token"); ok {
		if t, ok2 := v.(*model.APIToken); ok2 {
			uploaderToken = t
		}
	}

	maxTotal := int64(100 * 1024 * 1024)
	if h.svc != nil {
		if cfg := h.svc.Config(); cfg != nil {
			if v := cfg.Effective().MaxUploadBytes; v > 0 {
				maxTotal = v
			}
		}
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxTotal)

	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.WriteErrorCode(c, http.StatusBadRequest, "file_required", "file is required")
		return
	}

	maxPerFile := int64(60 * 1024 * 1024)
	if h.svc != nil {
		if cfg := h.svc.Config(); cfg != nil {
			if v := cfg.Effective().MaxUploadFileBytes; v > 0 {
				maxPerFile = v
			}
		}
	}
	if maxPerFile > 0 && fileHeader.Size > maxPerFile {
		response.WriteErrorCode(c, http.StatusBadRequest, "file_too_large", "file too large")
		return
	}

	f, err := fileHeader.Open()
	if err != nil {
		response.WriteErrorCode(c, http.StatusBadRequest, "file_open_failed", "open file failed")
		return
	}
	reader := io.Reader(f)
	if maxPerFile > 0 {
		reader = io.LimitReader(f, maxPerFile+1)
	}
	buf, err := io.ReadAll(reader)
	f.Close()
	if err != nil {
		response.WriteErrorCode(c, http.StatusBadRequest, "file_read_failed", "read file failed")
		return
	}
	if maxPerFile > 0 && int64(len(buf)) > maxPerFile {
		response.WriteErrorCode(c, http.StatusBadRequest, "file_too_large", "file too large")
		return
	}

	mimeType, width, height := detectUploadMIMEAndDimensions(buf)
	if _, allowed := allowedUploadMIMETypes[mimeType]; !allowed {
		response.WriteErrorCode(c, http.StatusBadRequest, "unsupported_file_type", "unsupported file type: "+mimeType)
		return
	}
	if !mediaDimensionsAllowed(width, height) {
		response.WriteErrorCode(c, http.StatusBadRequest, "media_dimensions_too_large", "media dimensions exceed limit")
		return
	}

	originalFileName := fileHeader.Filename
	cleanPath := filepath.Clean(originalFileName)
	if strings.Contains(cleanPath, "..") || filepath.IsAbs(cleanPath) || strings.ContainsAny(cleanPath, "/\\") {
		originalFileName = filepath.Base(cleanPath)
	} else {
		originalFileName = cleanPath
	}
	if originalFileName == "" || originalFileName == "." || originalFileName == ".." {
		response.WriteErrorCode(c, http.StatusBadRequest, "invalid_filename", "invalid filename")
		return
	}

	finalFileName := originalFileName
	if customName := c.PostForm("custom_name"); customName != "" {
		cleanCustomName := filepath.Clean(customName)
		if !strings.Contains(cleanCustomName, "..") && !filepath.IsAbs(cleanCustomName) && !strings.ContainsAny(cleanCustomName, "/\\") {
			cn := filepath.Base(cleanCustomName)
			if cn != "" && cn != "." && cn != ".." {
				finalFileName = cn
			}
		}
	}

	var tags []string
	if tagsStr := c.PostForm("tags"); tagsStr != "" {
		tagList := strings.Split(tagsStr, ",")
		for _, tag := range tagList {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				tags = append(tags, tag)
			}
		}
	}
	var routes []string
	if routeStr := c.PostForm("route"); routeStr != "" {
		list := strings.Split(routeStr, ",")
		for _, r := range list {
			r = strings.TrimSpace(r)
			if r != "" {
				routes = append(routes, r)
			}
		}
	}

	convert, _ := strconv.ParseBool(c.PostForm("convert"))
	targetFormat := c.PostForm("target_format")
	quality, _ := strconv.Atoi(c.PostForm("quality"))
	effort, _ := strconv.Atoi(c.PostForm("effort"))

	var uploadedByTokenID *uint
	var uploadedByTokenName string
	var uploadedByTokenType string
	if uploaderToken != nil {
		uploadedByTokenID = &uploaderToken.ID
		uploadedByTokenName = uploaderToken.Name
		uploadedByTokenType = uploaderToken.NormalizedType()
	}

	task, err := h.svc.EnqueueUploadTask(service.UploadTaskInput{
		Buf:                 buf,
		FileName:            finalFileName,
		Routes:              routes,
		Description:         c.PostForm("description"),
		Tags:                tags,
		MimeType:            mimeType,
		Width:               width,
		Height:              height,
		Convert:             convert && service.IsImageFile(mimeType),
		TargetFormat:        targetFormat,
		Quality:             quality,
		Effort:              effort,
		UploadedByTokenID:   uploadedByTokenID,
		UploadedByTokenName: uploadedByTokenName,
		UploadedByTokenType: uploadedByTokenType,
		RequestMethod:       c.Request.Method,
		RequestPath:         c.Request.URL.Path,
		IPAddress:           middleware.ClientIP(c),
		UserAgent:           c.Request.UserAgent(),
	})
	if err != nil {
		response.WriteErrorCode(c, http.StatusInternalServerError, "enqueue_upload_failed", "failed to enqueue upload")
		return
	}

	c.JSON(http.StatusAccepted, task)
}

// GET /api/v1/images/tasks/:id
func (h *ImageHandler) GetUploadTask(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		response.WriteErrorCode(c, http.StatusBadRequest, "task_id_required", "task id is required")
		return
	}

	task, err := h.svc.GetUploadTask(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.WriteErrorCode(c, http.StatusNotFound, "task_not_found", "task not found")
			return
		}
		response.WriteErrorCode(c, http.StatusInternalServerError, "get_upload_task_failed", "failed to get upload task")
		return
	}

	c.JSON(http.StatusOK, task)
}

func classifyURLFetchError(err error) (string, string) {
	switch {
	case errors.Is(err, service.ErrURLInvalid):
		return "url_invalid", "invalid url"
	case errors.Is(err, service.ErrURLBlocked):
		return "url_blocked", "url target not allowed"
	case errors.Is(err, service.ErrURLTooLarge):
		return "url_too_large", "url response too large"
	default:
		return "url_fetch_failed", "failed to fetch url"
	}
}

// GET /i/:hash
func (h *ImageHandler) GetByHash(c *gin.Context) {
	hashStr := c.Param("hash")

	img, absPath, err := h.svc.ResolveByHash(c.Request.Context(), hashStr)
	if err != nil {
		response.WriteErrorCode(c, http.StatusNotFound, "image_not_found", "image not found")
		return
	}

	if strings.HasPrefix(absPath, "http://") || strings.HasPrefix(absPath, "https://") {
		c.Redirect(http.StatusFound, absPath)
		return
	}
	if img.MimeType == "image/svg+xml" {
		c.Header("Content-Disposition", "attachment")
	}
	serveLocalMedia(c, absPath, img.MimeType)
}

// GET /i/:hash/thumbnail
func (h *ImageHandler) GetThumbnailByHash(c *gin.Context) {
	hashStr := c.Param("hash")

	absPath, mimeType, err := h.svc.ResolveThumbnailByHash(c.Request.Context(), hashStr)
	if err != nil {
		response.WriteErrorCode(c, http.StatusNotFound, "thumbnail_not_found", "thumbnail not found")
		return
	}

	if strings.HasPrefix(absPath, "http://") || strings.HasPrefix(absPath, "https://") {
		c.Redirect(http.StatusFound, absPath)
		return
	}
	serveLocalMedia(c, absPath, mimeType)
}

// GET /i/r/:route
func (h *ImageHandler) GetByRoute(c *gin.Context) {
	routeStr := c.Param("route")

	img, absPath, err := h.svc.ResolveByRoute(c.Request.Context(), routeStr)
	if err != nil {
		response.WriteErrorCode(c, http.StatusNotFound, "route_not_found", "route not found")
		return
	}

	if strings.HasPrefix(absPath, "http://") || strings.HasPrefix(absPath, "https://") {
		c.Redirect(http.StatusFound, absPath)
		return
	}
	if img.MimeType == "image/svg+xml" {
		c.Header("Content-Disposition", "attachment")
	}
	serveLocalMedia(c, absPath, img.MimeType)
}

func serveLocalMedia(c *gin.Context, absPath, mimeType string) {
	if mimeType != "" {
		c.Header("Content-Type", mimeType)
	}
	c.File(absPath)
}

// GET /api/v1/images
func (h *ImageHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	tag := c.Query("tag")
	fileName := c.Query("file_name")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	images, total, err := h.svc.ListImages(page, pageSize, tag, fileName)
	if err != nil {
		response.WriteErrorCode(c, http.StatusInternalServerError, "list_images_failed", "failed to list images")
		return
	}

	if v, ok := c.Get("api_token"); ok {
		if token, ok2 := v.(*model.APIToken); ok2 && token != nil {
			tokenSvc := service.NewAPITokenService(h.svc.Config(), h.svc.DB())
			_ = tokenSvc.RecordLog(&model.APITokenLog{
				TokenID:   token.ID,
				TokenName: token.Name,
				TokenType: token.NormalizedType(),
				Action:    "image_list",
				Method:    c.Request.Method,
				Path:      c.Request.URL.RequestURI(),
				IPAddress: middleware.ClientIP(c),
				UserAgent: c.Request.UserAgent(),
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  images,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

// GET /api/v1/tags
func (h *ImageHandler) ListTags(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "200"))
	if limit < 1 || limit > 1000 {
		limit = 200
	}

	tags, err := h.svc.ListTags(limit)
	if err != nil {
		response.WriteErrorCode(c, http.StatusInternalServerError, "list_tags_failed", "failed to list tags")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": tags,
	})
}

// DELETE /api/v1/images/:hash
func (h *ImageHandler) Delete(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		response.WriteErrorCode(c, http.StatusBadRequest, "hash_required", "hash is required")
		return
	}

	if err := h.svc.DeleteImage(c.Request.Context(), hash); err != nil {
		response.WriteErrorCode(c, http.StatusInternalServerError, "delete_image_failed", "failed to delete image")
		return
	}

	c.Status(http.StatusNoContent)
}

type UpdateImageRequest struct {
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	FileName    string   `json:"file_name"`
	Routes      []string `json:"routes"`
}

// PATCH /api/v1/images/:hash
func (h *ImageHandler) Update(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		response.WriteErrorCode(c, http.StatusBadRequest, "hash_required", "hash is required")
		return
	}

	var req UpdateImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.WriteErrorCode(c, http.StatusBadRequest, "invalid_update_request", "invalid request body")
		return
	}

	img, err := h.svc.UpdateImage(hash, req.Description, req.Tags, req.FileName)
	if err != nil {
		response.WriteErrorCode(c, http.StatusInternalServerError, "update_image_failed", "failed to update image")
		return
	}

	if req.Routes != nil {
		if err := h.svc.UpdateRoutes(img.ID, req.Routes); err != nil {
			status := http.StatusInternalServerError
			if containsDuplicateKey(err.Error()) {
				status = http.StatusBadRequest
			}
			if status == http.StatusBadRequest {
				response.WriteErrorCode(c, status, "route_exists", "route already exists")
			} else {
				response.WriteErrorCode(c, status, "update_routes_failed", "failed to update routes")
			}
			return
		}
	}

	c.JSON(http.StatusOK, img)
}

// GET /api/v1/routes
func (h *ImageHandler) ListRoutes(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	routes, total, err := h.svc.ListRoutes(page, pageSize)
	if err != nil {
		response.WriteErrorCode(c, http.StatusInternalServerError, "list_routes_failed", "failed to list routes")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  routes,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

// DELETE /api/v1/routes/:route
func (h *ImageHandler) DeleteRoute(c *gin.Context) {
	route := c.Param("route")
	if route == "" {
		response.WriteErrorCode(c, http.StatusBadRequest, "route_required", "route is required")
		return
	}

	if err := h.svc.DeleteRoute(route); err != nil {
		response.WriteErrorCode(c, http.StatusInternalServerError, "delete_route_failed", "failed to delete route")
		return
	}

	c.Status(http.StatusNoContent)
}

// GET /api/v1/images/:hash/info
func (h *ImageHandler) GetInfo(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		response.WriteErrorCode(c, http.StatusBadRequest, "hash_required", "hash is required")
		return
	}

	img, _, err := h.svc.ResolveByHash(c.Request.Context(), hash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.WriteErrorCode(c, http.StatusNotFound, "image_not_found", "image not found")
		} else {
			response.WriteErrorCode(c, http.StatusInternalServerError, "get_image_info_failed", "failed to get image info")
		}
		return
	}

	// 获取图片的路由映射
	var routes []string
	var imageRoutes []model.ImageRoute
	if err := h.svc.DB().Where("image_id = ?", img.ID).Find(&imageRoutes).Error; err == nil {
		for _, r := range imageRoutes {
			routes = append(routes, r.Route)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"hash":                   img.Hash,
		"file_name":              img.FileName,
		"mime_type":              img.MimeType,
		"size":                   img.Size,
		"width":                  img.Width,
		"height":                 img.Height,
		"duration_seconds":       img.DurationSeconds,
		"video_codec":            img.VideoCodec,
		"video_bitrate":          img.VideoBitrate,
		"audio_codec":            img.AudioCodec,
		"audio_bitrate":          img.AudioBitrate,
		"description":            img.Description,
		"tags":                   img.Tags,
		"uploaded_by_token_id":   img.UploadedByTokenID,
		"uploaded_by_token_name": img.UploadedByTokenName,
		"uploaded_by_token_type": img.UploadedByTokenType,
		"created_at":             img.CreatedAt,
		"updated_at":             img.UpdatedAt,
		"routes":                 routes,
	})
}

// GET /api/v1/stats
func (h *ImageHandler) GetStats(c *gin.Context) {
	stats, err := h.svc.GetStats()
	if err != nil {
		response.WriteErrorCode(c, http.StatusInternalServerError, "get_stats_failed", "failed to get system stats")
		return
	}

	c.JSON(http.StatusOK, stats)
}

// 判断唯一冲突
func containsDuplicateKey(msg string) bool {
	msg = strings.ToLower(msg)
	return strings.Contains(msg, "duplicate key") ||
		strings.Contains(msg, "unique constraint") ||
		strings.Contains(msg, "violates unique")
}
