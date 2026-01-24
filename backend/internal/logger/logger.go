package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

var levelNames = [...]string{
	"DEBUG",
	"INFO",
	"WARN",
	"ERROR",
	"FATAL",
}

const (
	colorReset = "\033[0m"

	colorWhite = "\033[37m"
	colorDebug = "\033[36m" // 青色
	colorWarn  = "\033[33m" // 黄色
	colorError = "\033[31m" // 红色
	colorFatal = "\033[35m" // 品红
)

type Logger struct {
	module  string
	out     io.Writer
	bufPool *sync.Pool
}

var (
	defaultLogger = New("default")

	registryMu sync.RWMutex
	registry   = make(map[string]*Logger)
)

type Option func(*Logger)

func WithOutput(w io.Writer) Option {
	return func(l *Logger) {
		l.out = w
	}
}

// New 创建一个带模块名的 logger
func New(module string, opts ...Option) *Logger {
	l := &Logger{
		module: module,
		out:    os.Stdout,
		bufPool: &sync.Pool{
			New: func() any {
				return new(bytes.Buffer)
			},
		},
	}

	for _, opt := range opts {
		opt(l)
	}

	return l
}

// Register 注册一个模块 logger，如果重复注册同一个模块，返回已存在的
func Register(module string, opts ...Option) *Logger {
	registryMu.Lock()
	defer registryMu.Unlock()

	if existing, ok := registry[module]; ok {
		return existing
	}

	l := New(module, opts...)
	registry[module] = l
	return l
}

// Get 获取某个模块的 logger
func Get(module string) *Logger {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return registry[module]
}

func Default() *Logger {
	return defaultLogger
}

func Debug(v ...any)            { defaultLogger.Debug(v...) }
func Info(v ...any)             { defaultLogger.Info(v...) }
func Warn(v ...any)             { defaultLogger.Warn(v...) }
func Error(v ...any)            { defaultLogger.Error(v...) }
func Fatal(v ...any)            { defaultLogger.Fatal(v...) }
func Debugf(f string, a ...any) { defaultLogger.Debugf(f, a...) }
func Infof(f string, a ...any)  { defaultLogger.Infof(f, a...) }
func Warnf(f string, a ...any)  { defaultLogger.Warnf(f, a...) }
func Errorf(f string, a ...any) { defaultLogger.Errorf(f, a...) }
func Fatalf(f string, a ...any) { defaultLogger.Fatalf(f, a...) }

func (l *Logger) Debug(v ...any) { l.logPlain(LevelDebug, v...) }
func (l *Logger) Info(v ...any)  { l.logPlain(LevelInfo, v...) }
func (l *Logger) Warn(v ...any)  { l.logPlain(LevelWarn, v...) }
func (l *Logger) Error(v ...any) { l.logPlain(LevelError, v...) }
func (l *Logger) Fatal(v ...any) {
	l.logPlain(LevelFatal, v...)
	os.Exit(1)
}

func (l *Logger) Debugf(format string, args ...any) { l.logf(LevelDebug, format, args...) }
func (l *Logger) Infof(format string, args ...any)  { l.logf(LevelInfo, format, args...) }
func (l *Logger) Warnf(format string, args ...any)  { l.logf(LevelWarn, format, args...) }
func (l *Logger) Errorf(format string, args ...any) { l.logf(LevelError, format, args...) }
func (l *Logger) Fatalf(format string, args ...any) {
	l.logf(LevelFatal, format, args...)
	os.Exit(1)
}

func (l *Logger) logPlain(level Level, v ...any) {
	msg := fmt.Sprint(v...)
	l.log(level, msg)
}

func (l *Logger) logf(level Level, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	l.log(level, msg)
}

// 格式：时间(UTC) [模块] [类型] 输出文本
func (l *Logger) log(level Level, msg string) {
	buf := l.bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer l.bufPool.Put(buf)

	now := time.Now().UTC()
	ts := now.AppendFormat(nil, "2006-01-02 15:04:05.000")

	prefixColor := colorWhite
	levelName := levelNames[level]

	switch level {
	case LevelDebug:
		prefixColor = colorDebug
	case LevelWarn:
		prefixColor = colorWarn
	case LevelError:
		prefixColor = colorError
	case LevelFatal:
		prefixColor = colorFatal
	case LevelInfo:
		prefixColor = colorWhite
	}
	buf.WriteString(prefixColor)
	buf.Write(ts)
	buf.WriteString(" [")
	buf.WriteString(l.module)
	buf.WriteString("]")
	buf.WriteString(colorReset)
	buf.WriteByte(' ')
	buf.WriteString(colorWhite)
	buf.WriteByte('[')
	buf.WriteString(levelName)
	buf.WriteString("] ")
	buf.WriteString(msg)
	buf.WriteString(colorReset)
	buf.WriteByte('\n')
	_, _ = l.out.Write(buf.Bytes())
}
