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
	"github.com/TangTangChu/AnzuImg/backend/internal/model"
	"github.com/TangTangChu/AnzuImg/backend/internal/service"
)

type ImageHandler struct {
	svc *service.ImageService
}

func NewImageHandler(cfg *config.Config, db *gorm.DB) *ImageHandler {
	return &ImageHandler{
		svc: service.NewImageService(cfg, db),
	}
}

// POST /api/v1/images
// form-data: file=<file> (can be multiple), route=<optional>, description=<optional>, tags=<optional>
func (h *ImageHandler) Upload(c *gin.Context) {
	maxTotal := int64(100 * 1024 * 1024)
	if h.svc != nil {
		if cfg := h.svc.Config(); cfg != nil && cfg.MaxUploadBytes > 0 {
			maxTotal = cfg.MaxUploadBytes
		}
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxTotal)

	if err := c.Request.ParseMultipartForm(maxTotal); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file size exceeds limit or invalid form"})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid multipart form"})
		return
	}

	files := form.File["file"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	// 文件数量限制
	maxFiles := 20
	if h.svc != nil {
		if cfg := h.svc.Config(); cfg != nil && cfg.MaxUploadFiles > 0 {
			maxFiles = cfg.MaxUploadFiles
		}
	}
	if maxFiles > 0 && len(files) > maxFiles {
		c.JSON(http.StatusBadRequest, gin.H{"error": "too many files"})
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
	}
	var metadataList []FileMetadata
	metadataStr := c.PostForm("metadata")
	if metadataStr != "" {
		_ = json.Unmarshal([]byte(metadataStr), &metadataList)
	}

	var results []gin.H

	for i, fileHeader := range files {
		maxPerFile := int64(60 * 1024 * 1024)
		if h.svc != nil {
			if cfg := h.svc.Config(); cfg != nil && cfg.MaxUploadFileBytes > 0 {
				maxPerFile = cfg.MaxUploadFileBytes
			}
		}
		if maxPerFile > 0 && fileHeader.Size > maxPerFile {
			results = append(results, gin.H{"file_name": fileHeader.Filename, "error": "file too large"})
			continue
		}

		f, err := fileHeader.Open()
		if err != nil {
			results = append(results, gin.H{"file_name": fileHeader.Filename, "error": "open file failed: " + err.Error()})
			continue
		}

		reader := io.Reader(f)
		if maxPerFile > 0 {
			reader = io.LimitReader(f, maxPerFile+1)
		}
		buf, err := io.ReadAll(reader)
		f.Close()
		if err != nil {
			results = append(results, gin.H{"file_name": fileHeader.Filename, "error": "read file failed: " + err.Error()})
			continue
		}
		if maxPerFile > 0 && int64(len(buf)) > maxPerFile {
			results = append(results, gin.H{"file_name": fileHeader.Filename, "error": "file too large"})
			continue
		}

		// MIME类型验证
		allowedMIMETypes := map[string]struct{}{
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
		}

		mimeType, width, height, err := service.InspectImage(bytes.NewReader(buf))
		if err != nil {
			mimeType = "application/octet-stream"
			width = 0
			height = 0
		}

		if _, allowed := allowedMIMETypes[mimeType]; !allowed {
			results = append(results, gin.H{
				"file_name": fileHeader.Filename,
				"error":     "unsupported file type: " + mimeType,
			})
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
			results = append(results, gin.H{"file_name": fileHeader.Filename, "error": "invalid filename"})
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

		res, err := h.svc.Upload(buf, finalFileName, currentRoutes, currentDesc, currentTags, mimeType, width, height, convert, targetFormat, quality, effort)
		if err != nil {
			results = append(results, gin.H{"file_name": fileHeader.Filename, "error": err.Error()})
			continue
		}

		results = append(results, gin.H{
			"hash":        res.Image.Hash,
			"file_name":   res.Image.FileName,
			"size":        res.Image.Size,
			"mime":        res.Image.MimeType,
			"path":        res.Image.Path,
			"width":       res.Image.Width,
			"height":      res.Image.Height,
			"description": res.Image.Description,
			"tags":        res.Image.Tags,
			"created_at":  res.Image.CreatedAt,
			"updated_at":  res.Image.UpdatedAt,
			"reused":      res.Reused,
			"url":         res.HashURL,
			"route":       res.Route,
			"route_url":   res.RouteURL,
			"success":     true,
		})
	}

	c.JSON(http.StatusOK, results)
}

// GET /i/:hash
func (h *ImageHandler) GetByHash(c *gin.Context) {
	hashStr := c.Param("hash")

	_, absPath, err := h.svc.ResolveByHash(hashStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "image not found"})
		return
	}

	if strings.HasPrefix(absPath, "http://") || strings.HasPrefix(absPath, "https://") {
		c.Redirect(http.StatusFound, absPath)
		return
	}
	c.File(absPath)
}

// GET /i/:hash/thumbnail
func (h *ImageHandler) GetThumbnailByHash(c *gin.Context) {
	hashStr := c.Param("hash")

	absPath, err := h.svc.ResolveThumbnailByHash(hashStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "thumbnail not found"})
		return
	}

	if strings.HasPrefix(absPath, "http://") || strings.HasPrefix(absPath, "https://") {
		c.Redirect(http.StatusFound, absPath)
		return
	}
	c.File(absPath)
}

// GET /i/r/:route
func (h *ImageHandler) GetByRoute(c *gin.Context) {
	routeStr := c.Param("route")

	_, absPath, err := h.svc.ResolveByRoute(routeStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
		return
	}

	if strings.HasPrefix(absPath, "http://") || strings.HasPrefix(absPath, "https://") {
		c.Redirect(http.StatusFound, absPath)
		return
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "hash is required"})
		return
	}

	if err := h.svc.DeleteImage(hash); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "hash is required"})
		return
	}

	var req UpdateImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	img, err := h.svc.UpdateImage(hash, req.Description, req.Tags, req.FileName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if req.Routes != nil {
		if err := h.svc.UpdateRoutes(img.ID, req.Routes); err != nil {
			status := http.StatusInternalServerError
			if containsDuplicateKey(err.Error()) {
				status = http.StatusBadRequest
			}
			c.JSON(status, gin.H{"error": "failed to update routes: " + err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "route is required"})
		return
	}

	if err := h.svc.DeleteRoute(route); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// GET /api/v1/images/:hash/info
func (h *ImageHandler) GetInfo(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "hash is required"})
		return
	}

	img, _, err := h.svc.ResolveByHash(hash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "image not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		"hash":        img.Hash,
		"file_name":   img.FileName,
		"mime_type":   img.MimeType,
		"size":        img.Size,
		"width":       img.Width,
		"height":      img.Height,
		"description": img.Description,
		"tags":        img.Tags,
		"created_at":  img.CreatedAt,
		"updated_at":  img.UpdatedAt,
		"routes":      routes,
	})
}

// 判断唯一冲突
func containsDuplicateKey(msg string) bool {
	msg = strings.ToLower(msg)
	return strings.Contains(msg, "duplicate key") ||
		strings.Contains(msg, "unique constraint") ||
		strings.Contains(msg, "violates unique")
}
