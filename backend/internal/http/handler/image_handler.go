package handler

import (
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/TangTangChu/AnzuImg/backend/internal/config"
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
// form-data: file=<file>, route=<optional>
func (h *ImageHandler) Upload(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	routeStr := c.PostForm("route")

	f, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "open file failed"})
		return
	}
	defer f.Close()

	buf, err := io.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "read file failed"})
		return
	}

	res, err := h.svc.Upload(buf, fileHeader.Filename, routeStr)
	if err != nil {
		// route 冲突 / 非法属于 400，其余 500
		status := http.StatusInternalServerError
		if containsDuplicateKey(err.Error()) {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"hash":      res.Image.Hash,
		"file_name": res.Image.FileName,
		"size":      res.Image.Size,
		"mime":      res.Image.MimeType,
		"path":      res.Image.Path,
		"reused":    res.Reused,
		"url":       res.HashURL,
		"route":     res.Route,
		"route_url": res.RouteURL,
	})
}

// GET /i/:hash
func (h *ImageHandler) GetByHash(c *gin.Context) {
	hashStr := c.Param("hash")

	_, absPath, err := h.svc.ResolveByHash(hashStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "image not found"})
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

	c.File(absPath)
}

// 判断唯一冲突
func containsDuplicateKey(msg string) bool {
	msg = strings.ToLower(msg)
	return strings.Contains(msg, "duplicate key") ||
		strings.Contains(msg, "unique constraint") ||
		strings.Contains(msg, "violates unique")
}
