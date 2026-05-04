package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Record 是单条日志的结构化形式，给 Sink 实现使用。
// Logger.log 在 fan-out 给 Sink 之前会构造 Record。
type Record struct {
	Time    time.Time
	Level   Level
	Module  string
	Message string
}

// Sink 接受结构化日志记录并按自身策略输出。Sink 实现需要并发安全。
// Write 应当尽量不阻塞调用方；DB/远程 sink 的实现应内部排队后异步处理。
type Sink interface {
	Name() string
	// MinLevel 返回该 sink 接收的最小级别；低于此级别的记录会被 Logger 跳过。
	MinLevel() Level
	Write(rec Record)
	Close() error
}

// 全局 sink 注册表。每条日志在写完本地 stdout 后也会广播给所有全局 sink，
// 例如 FileSink 与 DBSink。
var (
	globalSinksMu sync.RWMutex
	globalSinks   = map[string]Sink{}
)

// AddGlobalSink 注册或替换同名 sink,返回旧的 sink,由调用者决定是否 Close。
func AddGlobalSink(s Sink) Sink {
	if s == nil {
		return nil
	}
	globalSinksMu.Lock()
	defer globalSinksMu.Unlock()
	old := globalSinks[s.Name()]
	globalSinks[s.Name()] = s
	return old
}

// RemoveGlobalSink 取消注册并返回被移除的 sink；调用方负责 Close。
func RemoveGlobalSink(name string) Sink {
	globalSinksMu.Lock()
	defer globalSinksMu.Unlock()
	old := globalSinks[name]
	delete(globalSinks, name)
	return old
}

// GlobalSinks 返回当前所有全局 sink 的快照。
func GlobalSinks() []Sink {
	globalSinksMu.RLock()
	defer globalSinksMu.RUnlock()
	out := make([]Sink, 0, len(globalSinks))
	for _, s := range globalSinks {
		out = append(out, s)
	}
	return out
}

// CloseAllGlobalSinks 进程退出时调用，确保 sink 内部的缓冲与文件句柄被释放。
func CloseAllGlobalSinks() {
	globalSinksMu.Lock()
	defer globalSinksMu.Unlock()
	for name, s := range globalSinks {
		_ = s.Close()
		delete(globalSinks, name)
	}
}

// ParseLevel 把字符串映射为 Level；off 表示禁用。
// 解析失败返回 LevelInfo；off 返回一个超过最高级别的特殊值。
const LevelOff Level = 99

func ParseLevel(s string) Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	case "fatal":
		return LevelFatal
	case "off", "disable", "disabled", "none":
		return LevelOff
	default:
		return LevelInfo
	}
}

// LevelName 与 ParseLevel 对应的反向函数。
func LevelName(l Level) string {
	if l == LevelOff {
		return "off"
	}
	if int(l) >= 0 && int(l) < len(levelNames) {
		return strings.ToLower(levelNames[l])
	}
	return "unknown"
}

// formatPlainLine 把 Record 渲染为不含 ANSI 颜色码的单行文本，
// FileSink 与 DBSink 共用。
func formatPlainLine(rec Record) string {
	var buf bytes.Buffer
	buf.Grow(64 + len(rec.Message))
	ts := rec.Time.UTC().AppendFormat(nil, "2006-01-02 15:04:05.000")
	buf.Write(ts)
	buf.WriteString(" [")
	buf.WriteString(rec.Module)
	buf.WriteString("] [")
	if int(rec.Level) >= 0 && int(rec.Level) < len(levelNames) {
		buf.WriteString(levelNames[rec.Level])
	} else {
		buf.WriteString("?")
	}
	buf.WriteString("] ")
	buf.WriteString(rec.Message)
	buf.WriteByte('\n')
	return buf.String()
}

// FileSink 把日志写入一个按大小/时间轮转的文件。底层用 lumberjack。
type FileSink struct {
	name string
	min  Level
	w    *lumberjack.Logger
	mu   sync.Mutex
}

// FileSinkOptions 控制 FileSink 的创建。
type FileSinkOptions struct {
	Name       string // sink 注册名,默认 file
	Path       string // 完整文件路径
	MinLevel   Level  // 低于此级别的记录会丢弃
	MaxSizeMB  int
	MaxBackups int
	MaxAgeDays int
	Compress   bool
}

func NewFileSink(opts FileSinkOptions) (*FileSink, error) {
	if strings.TrimSpace(opts.Path) == "" {
		return nil, fmt.Errorf("file sink: path required")
	}
	name := opts.Name
	if name == "" {
		name = "file"
	}
	maxSize := opts.MaxSizeMB
	if maxSize <= 0 {
		maxSize = 50
	}
	w := &lumberjack.Logger{
		Filename:   opts.Path,
		MaxSize:    maxSize,
		MaxBackups: opts.MaxBackups,
		MaxAge:     opts.MaxAgeDays,
		Compress:   opts.Compress,
		LocalTime:  true,
	}
	return &FileSink{name: name, min: opts.MinLevel, w: w}, nil
}

func (s *FileSink) Name() string     { return s.name }
func (s *FileSink) MinLevel() Level  { return s.min }
func (s *FileSink) Write(rec Record) {
	if rec.Level < s.min {
		return
	}
	line := formatPlainLine(rec)
	s.mu.Lock()
	_, _ = io.WriteString(s.w, line)
	s.mu.Unlock()
}
func (s *FileSink) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.w == nil {
		return nil
	}
	return s.w.Close()
}

// 一个简易的"NopSink",仅用于测试或占位。
type nopSink struct{ name string }

func NewNopSink(name string) Sink             { return &nopSink{name: name} }
func (n *nopSink) Name() string               { return n.name }
func (n *nopSink) MinLevel() Level            { return LevelOff }
func (n *nopSink) Write(rec Record)           {}
func (n *nopSink) Close() error               { return nil }

// 兼容性：暴露内部 stdout writer 工厂，供测试或自定义 logger 输出复用。
func DefaultStdout() io.Writer { return os.Stdout }
