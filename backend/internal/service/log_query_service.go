package service

import (
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/TangTangChu/AnzuImg/backend/internal/logger"
	"github.com/TangTangChu/AnzuImg/backend/internal/model"
)

// LogQueryService 给 log handler 提供分页/过滤/导出能力,
// app_logs / security_event_logs / api_token_logs 共用此实现。
type LogQueryService struct {
	db *gorm.DB
}

func NewLogQueryService(db *gorm.DB) *LogQueryService {
	return &LogQueryService{db: db}
}

type LogFilter struct {
	Search    string
	Level     string
	Module    string
	IPAddress string
	Action    string
	StartDate string
	EndDate   string
}

func clampPage(page, size int) (int, int) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 500 {
		size = 50
	}
	return page, size
}

var levelOrder = []string{"debug", "info", "warn", "error", "fatal"}

// levelsAtOrAbove 把单选的级别解释为“该级别及以上”，与实时流的 matchFilter 语义一致，
// 并带上历史别名 warning，兼容旧的安全日志数据。
func levelsAtOrAbove(name string) []string {
	target := logger.NormalizeLevelName(name)
	idx := -1
	for i, l := range levelOrder {
		if l == target {
			idx = i
			break
		}
	}
	if idx < 0 {
		return []string{strings.ToLower(strings.TrimSpace(name))}
	}
	out := make([]string, 0, len(levelOrder)+1)
	for _, l := range levelOrder[idx:] {
		out = append(out, l)
		if l == "warn" {
			out = append(out, "warning")
		}
	}
	return out
}

// escapeLike 转义 LIKE 通配符，避免用户输入的 % _ \ 被当作模式，
// 既防止意外的全表通配，也避免前导通配导致的歧义。
func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

func parseTimeBound(raw string, endOfDay bool) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t, true
	}
	if t, err := time.Parse("2006-01-02", raw); err == nil {
		if endOfDay {
			t = t.Add(24*time.Hour - time.Second)
		}
		return t, true
	}
	if n, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return time.Unix(n, 0), true
	}
	return time.Time{}, false
}

func applyTimeBounds(q *gorm.DB, f LogFilter) *gorm.DB {
	if t, ok := parseTimeBound(f.StartDate, false); ok {
		q = q.Where("created_at >= ?", t)
	}
	if t, ok := parseTimeBound(f.EndDate, true); ok {
		q = q.Where("created_at <= ?", t)
	}
	return q
}

// ListAppLogs 分页查询应用日志。
func (s *LogQueryService) ListAppLogs(filter LogFilter, page, size int) ([]model.AppLog, int64, error) {
	page, size = clampPage(page, size)
	q := s.db.Model(&model.AppLog{})
	if filter.Level != "" {
		q = q.Where("level IN ?", levelsAtOrAbove(filter.Level))
	}
	if filter.Module != "" {
		q = q.Where("module = ?", filter.Module)
	}
	if filter.IPAddress != "" {
		q = q.Where("ip_address = ?", filter.IPAddress)
	}
	if filter.Search != "" {
		like := "%" + escapeLike(filter.Search) + "%"
		q = q.Where("message LIKE ? OR module LIKE ?", like, like)
	}
	q = applyTimeBounds(q, filter)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.AppLog
	if err := q.Order("created_at DESC").Limit(size).Offset((page - 1) * size).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// ListSecurityLogs 分页查询安全事件日志。failedOnly=true 时仅返回 warning/error。
func (s *LogQueryService) ListSecurityLogs(filter LogFilter, page, size int, failedOnly bool) ([]model.SecurityEventLog, int64, error) {
	page, size = clampPage(page, size)
	q := s.db.Model(&model.SecurityEventLog{})
	if failedOnly {
		q = q.Where("level IN ?", levelsAtOrAbove("warn"))
	}
	if filter.Level != "" {
		q = q.Where("level IN ?", levelsAtOrAbove(filter.Level))
	}
	if filter.Action != "" {
		q = q.Where("action = ?", filter.Action)
	}
	if filter.IPAddress != "" {
		q = q.Where("ip_address = ?", filter.IPAddress)
	}
	if filter.Search != "" {
		like := "%" + escapeLike(filter.Search) + "%"
		q = q.Where("action LIKE ? OR message LIKE ? OR path LIKE ? OR ip_address LIKE ? OR username LIKE ?", like, like, like, like, like)
	}
	q = applyTimeBounds(q, filter)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.SecurityEventLog
	if err := q.Order("created_at DESC").Limit(size).Offset((page - 1) * size).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// CleanupAppLogs 删除指定天数前的应用日志,返回受影响行数。
func (s *LogQueryService) CleanupAppLogs(days int) (int64, error) {
	if days <= 0 {
		return 0, nil
	}
	cutoff := time.Now().AddDate(0, 0, -days)
	tx := s.db.Where("created_at < ?", cutoff).Delete(&model.AppLog{})
	return tx.RowsAffected, tx.Error
}

// CleanupSecurityLogs 删除指定天数前的安全事件日志,返回受影响行数。
func (s *LogQueryService) CleanupSecurityLogs(days int) (int64, error) {
	if days <= 0 {
		return 0, nil
	}
	cutoff := time.Now().AddDate(0, 0, -days)
	tx := s.db.Where("created_at < ?", cutoff).Delete(&model.SecurityEventLog{})
	return tx.RowsAffected, tx.Error
}

// IterateAppLogs 流式扫描应用日志,提供给 CSV/JSON 导出使用。
func (s *LogQueryService) IterateAppLogs(filter LogFilter, hardLimit int, fn func(model.AppLog) error) error {
	if hardLimit <= 0 || hardLimit > 100000 {
		hardLimit = 50000
	}
	q := s.db.Model(&model.AppLog{})
	if filter.Level != "" {
		q = q.Where("level IN ?", levelsAtOrAbove(filter.Level))
	}
	if filter.Module != "" {
		q = q.Where("module = ?", filter.Module)
	}
	if filter.IPAddress != "" {
		q = q.Where("ip_address = ?", filter.IPAddress)
	}
	if filter.Search != "" {
		like := "%" + escapeLike(filter.Search) + "%"
		q = q.Where("message LIKE ? OR module LIKE ?", like, like)
	}
	q = applyTimeBounds(q, filter)
	rows, err := q.Order("created_at DESC").Limit(hardLimit).Rows()
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var row model.AppLog
		if err := s.db.ScanRows(rows, &row); err != nil {
			return err
		}
		if err := fn(row); err != nil {
			return err
		}
	}
	return rows.Err()
}

// IterateSecurityLogs 同上,扫描安全事件日志。
func (s *LogQueryService) IterateSecurityLogs(filter LogFilter, failedOnly bool, hardLimit int, fn func(model.SecurityEventLog) error) error {
	if hardLimit <= 0 || hardLimit > 100000 {
		hardLimit = 50000
	}
	q := s.db.Model(&model.SecurityEventLog{})
	if failedOnly {
		q = q.Where("level IN ?", levelsAtOrAbove("warn"))
	}
	if filter.Action != "" {
		q = q.Where("action = ?", filter.Action)
	}
	if filter.IPAddress != "" {
		q = q.Where("ip_address = ?", filter.IPAddress)
	}
	if filter.Search != "" {
		like := "%" + escapeLike(filter.Search) + "%"
		q = q.Where("action LIKE ? OR message LIKE ? OR path LIKE ? OR ip_address LIKE ? OR username LIKE ?", like, like, like, like, like)
	}
	q = applyTimeBounds(q, filter)
	rows, err := q.Order("created_at DESC").Limit(hardLimit).Rows()
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var row model.SecurityEventLog
		if err := s.db.ScanRows(rows, &row); err != nil {
			return err
		}
		if err := fn(row); err != nil {
			return err
		}
	}
	return rows.Err()
}
