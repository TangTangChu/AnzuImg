package handler

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/TangTangChu/AnzuImg/backend/internal/config"
	"github.com/TangTangChu/AnzuImg/backend/internal/http/middleware"
	"github.com/TangTangChu/AnzuImg/backend/internal/http/response"
	"github.com/TangTangChu/AnzuImg/backend/internal/logger"
	"github.com/TangTangChu/AnzuImg/backend/internal/model"
	"github.com/TangTangChu/AnzuImg/backend/internal/service"
)

type LogHandler struct {
	db     *gorm.DB
	q      *service.LogQueryService
	tokens *service.APITokenService
	hub    *service.LogStreamHub
	log    *logger.Logger
}

func NewLogHandler(cfg *config.Config, db *gorm.DB, hub *service.LogStreamHub) *LogHandler {
	return &LogHandler{
		db:     db,
		q:      service.NewLogQueryService(db),
		tokens: service.NewAPITokenService(cfg, db),
		hub:    hub,
		log:    logger.Register("log-handler"),
	}
}

func parsePageSize(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	return page, size
}

func parseLogFilter(c *gin.Context) service.LogFilter {
	return service.LogFilter{
		Search:    strings.TrimSpace(c.Query("search")),
		Level:     strings.TrimSpace(c.Query("level")),
		Module:    strings.TrimSpace(c.Query("module")),
		IPAddress: strings.TrimSpace(c.Query("ip")),
		Action:    strings.TrimSpace(c.Query("action")),
		StartDate: strings.TrimSpace(c.Query("start_date")),
		EndDate:   strings.TrimSpace(c.Query("end_date")),
	}
}

// ListApp GET /api/v1/logs/app
func (h *LogHandler) ListApp(c *gin.Context) {
	page, size := parsePageSize(c)
	rows, total, err := h.q.ListAppLogs(parseLogFilter(c), page, size)
	if err != nil {
		response.WriteErrorCode(c, http.StatusInternalServerError, "list_app_logs_failed", "failed to list app logs")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rows, "total": total, "page": page, "size": size})
}

// ListSecurity GET /api/v1/logs/security
func (h *LogHandler) ListSecurity(c *gin.Context) {
	page, size := parsePageSize(c)
	failedOnly, _ := strconv.ParseBool(c.DefaultQuery("failed_only", "false"))
	rows, total, err := h.q.ListSecurityLogs(parseLogFilter(c), page, size, failedOnly)
	if err != nil {
		response.WriteErrorCode(c, http.StatusInternalServerError, "list_security_logs_failed", "failed to list security logs")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rows, "total": total, "page": page, "size": size})
}

// ListToken GET /api/v1/logs/token
func (h *LogHandler) ListToken(c *gin.Context) {
	page, size := parsePageSize(c)
	f := parseLogFilter(c)
	rows, total, err := h.tokens.ListLogs(page, size, f.Search, f.StartDate, f.EndDate, f.Action)
	if err != nil {
		response.WriteErrorCode(c, http.StatusInternalServerError, "list_token_logs_failed", "failed to list token logs")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rows, "total": total, "page": page, "size": size})
}

// Cleanup DELETE /api/v1/logs/:source?days=N
func (h *LogHandler) Cleanup(c *gin.Context) {
	source := c.Param("source")
	days, _ := strconv.Atoi(c.DefaultQuery("days", "0"))
	if days <= 0 {
		response.WriteErrorCode(c, http.StatusBadRequest, "invalid_days", "invalid days")
		return
	}
	var deleted int64
	var err error
	switch source {
	case "app":
		deleted, err = h.q.CleanupAppLogs(days)
	case "security":
		deleted, err = h.q.CleanupSecurityLogs(days)
	case "token":
		cutoff := time.Now().AddDate(0, 0, -days)
		deleted, err = h.tokens.CleanupLogsBefore(cutoff)
	default:
		response.WriteErrorCode(c, http.StatusBadRequest, "invalid_source", "invalid log source")
		return
	}
	if err != nil {
		h.recordSecurityEvent(c, "warning", "log_cleanup_failed", source+": "+err.Error())
		response.WriteErrorCode(c, http.StatusInternalServerError, "cleanup_failed", "failed to cleanup logs")
		return
	}
	h.recordSecurityEvent(c, "info", "log_cleanup", source+": cleanup executed")
	c.JSON(http.StatusOK, gin.H{"deleted": deleted, "source": source})
}

// Export GET /api/v1/logs/export?source=app|security|token&format=csv|json
func (h *LogHandler) Export(c *gin.Context) {
	source := strings.TrimSpace(c.DefaultQuery("source", "app"))
	format := strings.ToLower(strings.TrimSpace(c.DefaultQuery("format", "csv")))
	if format != "csv" && format != "json" {
		response.WriteErrorCode(c, http.StatusBadRequest, "invalid_format", "format must be csv or json")
		return
	}
	if source != "app" && source != "security" && source != "token" {
		response.WriteErrorCode(c, http.StatusBadRequest, "invalid_source", "invalid log source")
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10000"))
	failedOnly, _ := strconv.ParseBool(c.DefaultQuery("failed_only", "false"))
	filter := parseLogFilter(c)

	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	filename := "logs-" + source + "-" + time.Now().UTC().Format("20060102T150405") + "." + format
	c.Writer.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	if format == "csv" {
		c.Writer.Header().Set("Content-Type", "text/csv; charset=utf-8")
	} else {
		c.Writer.Header().Set("Content-Type", "application/json")
	}

	if err := h.streamExport(c.Writer, source, format, filter, limit, failedOnly); err != nil {
		h.log.Ctx(c.Request.Context()).Warnf("export logs failed: %v", err)
	}
}

func (h *LogHandler) streamExport(w io.Writer, source, format string, filter service.LogFilter, limit int, failedOnly bool) error {
	switch source {
	case "app":
		return h.streamAppExport(w, format, filter, limit)
	case "security":
		return h.streamSecurityExport(w, format, filter, limit, failedOnly)
	case "token":
		return h.streamTokenExport(w, format, filter, limit)
	}
	return fmt.Errorf("unsupported source: %s", source)
}

func (h *LogHandler) streamAppExport(w io.Writer, format string, filter service.LogFilter, limit int) error {
	if format == "csv" {
		cw := csv.NewWriter(w)
		_ = cw.Write([]string{"created_at", "level", "module", "request_id", "ip_address", "message"})
		err := h.q.IterateAppLogs(filter, limit, func(row model.AppLog) error {
			return cw.Write([]string{row.CreatedAt.UTC().Format(time.RFC3339), row.Level, row.Module, row.RequestID, row.IPAddress, row.Message})
		})
		cw.Flush()
		return err
	}
	enc := json.NewEncoder(w)
	_, _ = w.Write([]byte("["))
	first := true
	err := h.q.IterateAppLogs(filter, limit, func(row model.AppLog) error {
		if !first {
			_, _ = w.Write([]byte(","))
		}
		first = false
		return enc.Encode(row)
	})
	_, _ = w.Write([]byte("]"))
	return err
}

func (h *LogHandler) streamSecurityExport(w io.Writer, format string, filter service.LogFilter, limit int, failedOnly bool) error {
	if format == "csv" {
		cw := csv.NewWriter(w)
		_ = cw.Write([]string{"created_at", "level", "category", "action", "message", "method", "path", "ip_address", "username"})
		err := h.q.IterateSecurityLogs(filter, failedOnly, limit, func(row model.SecurityEventLog) error {
			return cw.Write([]string{row.CreatedAt.UTC().Format(time.RFC3339), row.Level, row.Category, row.Action, row.Message, row.Method, row.Path, row.IPAddress, row.Username})
		})
		cw.Flush()
		return err
	}
	enc := json.NewEncoder(w)
	_, _ = w.Write([]byte("["))
	first := true
	err := h.q.IterateSecurityLogs(filter, failedOnly, limit, func(row model.SecurityEventLog) error {
		if !first {
			_, _ = w.Write([]byte(","))
		}
		first = false
		return enc.Encode(row)
	})
	_, _ = w.Write([]byte("]"))
	return err
}

func (h *LogHandler) streamTokenExport(w io.Writer, format string, filter service.LogFilter, limit int) error {
	rows, _, err := h.tokens.ListLogs(1, limit, filter.Search, filter.StartDate, filter.EndDate, filter.Action)
	if err != nil {
		return err
	}
	if format == "csv" {
		cw := csv.NewWriter(w)
		_ = cw.Write([]string{"created_at", "token_name", "token_type", "action", "method", "path", "ip_address", "user_agent", "image_hash"})
		for _, row := range rows {
			_ = cw.Write([]string{row.CreatedAt.UTC().Format(time.RFC3339), row.TokenName, row.TokenType, row.Action, row.Method, row.Path, row.IPAddress, row.UserAgent, row.ImageHash})
		}
		cw.Flush()
		return nil
	}
	enc := json.NewEncoder(w)
	return enc.Encode(rows)
}

// Stream GET /api/v1/logs/stream?source=app&level=info&module=...
// SSE 推送应用日志,断开则订阅自动清理。
func (h *LogHandler) Stream(c *gin.Context) {
	source := strings.TrimSpace(c.DefaultQuery("source", "app"))
	if source != "app" {
		response.WriteErrorCode(c, http.StatusBadRequest, "invalid_source", "stream only supports app logs")
		return
	}
	if h.hub == nil {
		response.WriteErrorCode(c, http.StatusServiceUnavailable, "stream_unavailable", "log stream hub is not initialized")
		return
	}
	minLevel := logger.ParseLevel(c.DefaultQuery("level", "info"))
	module := strings.TrimSpace(c.Query("module"))
	bufSize, _ := strconv.Atoi(c.DefaultQuery("buffer", "128"))

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		response.WriteErrorCode(c, http.StatusInternalServerError, "stream_unavailable", "streaming not supported")
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.WriteHeader(http.StatusOK)
	flusher.Flush()

	sub := h.hub.Subscribe(service.LogStreamFilter{MinLevel: minLevel, Module: module}, bufSize)
	defer sub.Close()

	heartbeat := time.NewTicker(20 * time.Second)
	defer heartbeat.Stop()

	ctx := c.Request.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case row, ok := <-sub.Ch:
			if !ok {
				return
			}
			payload, err := json.Marshal(row)
			if err != nil {
				continue
			}
			_, _ = fmt.Fprintf(c.Writer, "event: log\ndata: %s\n\n", payload)
			flusher.Flush()
		case <-heartbeat.C:
			_, _ = fmt.Fprint(c.Writer, ":heartbeat\n\n")
			flusher.Flush()
		}
	}
}

func (h *LogHandler) recordSecurityEvent(c *gin.Context, level, action, message string) {
	event := &model.SecurityEventLog{
		Category:  "log",
		Level:     level,
		Action:    action,
		Message:   message,
		Method:    c.Request.Method,
		Path:      c.Request.URL.Path,
		IPAddress: middleware.ClientIP(c),
		Username:  "admin",
	}
	if err := h.db.Create(event).Error; err != nil {
		h.log.Ctx(c.Request.Context()).Warnf("record security event failed: %v", err)
	}
}
