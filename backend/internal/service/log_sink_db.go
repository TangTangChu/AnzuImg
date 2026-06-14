package service

import (
	"sync"
	"sync/atomic"
	"time"

	"gorm.io/gorm"

	"github.com/TangTangChu/AnzuImg/backend/internal/logger"
	"github.com/TangTangChu/AnzuImg/backend/internal/model"
)

// LogStreamHub 把异步写入 DB 的应用日志同时 fan-out 给所有 SSE 订阅者。
// 每个订阅者有独立缓冲 channel,满了直接丢弃,以防慢客户端拖累其他订阅者。
type LogStreamHub struct {
	mu          sync.RWMutex
	subscribers map[*logSubscription]struct{}
	dropped     atomic.Int64
}

type LogSubscriber struct {
	ID    int64
	Ch    <-chan model.AppLog
	close func()
}

func (s *LogSubscriber) Close() {
	if s.close != nil {
		s.close()
	}
}

type logSubscription struct {
	ch     chan model.AppLog
	filter LogStreamFilter
}

type LogStreamFilter struct {
	MinLevel logger.Level
	Module   string
}

func NewLogStreamHub() *LogStreamHub {
	return &LogStreamHub{subscribers: map[*logSubscription]struct{}{}}
}

func (h *LogStreamHub) Subscribe(filter LogStreamFilter, buffer int) *LogSubscriber {
	if buffer <= 0 {
		buffer = 64
	}
	sub := &logSubscription{ch: make(chan model.AppLog, buffer), filter: filter}
	h.mu.Lock()
	h.subscribers[sub] = struct{}{}
	h.mu.Unlock()
	return &LogSubscriber{
		Ch: sub.ch,
		close: func() {
			h.mu.Lock()
			if _, ok := h.subscribers[sub]; ok {
				delete(h.subscribers, sub)
				close(sub.ch)
			}
			h.mu.Unlock()
		},
	}
}

func (h *LogStreamHub) Publish(rec model.AppLog) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for sub := range h.subscribers {
		if !matchFilter(sub.filter, rec) {
			continue
		}
		select {
		case sub.ch <- rec:
		default:
			// 慢消费者: 直接丢弃这一条避免阻塞其他订阅者
			h.dropped.Add(1)
		}
	}
}

// DroppedCount 返回因订阅者缓冲满而丢弃的累计条数,用于可观测。
func (h *LogStreamHub) DroppedCount() int64 { return h.dropped.Load() }

func matchFilter(f LogStreamFilter, rec model.AppLog) bool {
	if f.Module != "" && rec.Module != f.Module {
		return false
	}
	if f.MinLevel == logger.LevelOff {
		return false
	}
	return logger.ParseLevel(rec.Level) >= f.MinLevel
}

// LogDBSink 把日志异步批量写入 app_logs 表。
// 通道缓冲满了直接丢弃,保证业务路径不被阻塞;丢弃量通过 dropped 计数暴露。
type LogDBSink struct {
	name   string
	min    logger.Level
	db     *gorm.DB
	bufLog *logger.Logger

	in       chan model.AppLog
	dropped  atomic.Int64
	stopOnce sync.Once
	stopped  chan struct{}
	done     chan struct{}
}

type LogDBSinkOptions struct {
	Name       string
	MinLevel   logger.Level
	BufferSize int
	BatchSize  int
	FlushEvery time.Duration
}

func NewLogDBSink(db *gorm.DB, opts LogDBSinkOptions) *LogDBSink {
	name := opts.Name
	if name == "" {
		name = "db"
	}
	bufSize := opts.BufferSize
	if bufSize <= 0 {
		bufSize = 4096
	}
	if opts.BatchSize <= 0 {
		opts.BatchSize = 64
	}
	if opts.FlushEvery <= 0 {
		opts.FlushEvery = 2 * time.Second
	}
	s := &LogDBSink{
		name:    name,
		min:     opts.MinLevel,
		db:      db,
		bufLog:  logger.Register("log-db-sink"),
		in:      make(chan model.AppLog, bufSize),
		stopped: make(chan struct{}),
		done:    make(chan struct{}),
	}
	go s.run(opts.BatchSize, opts.FlushEvery)
	return s
}

func (s *LogDBSink) Name() string           { return s.name }
func (s *LogDBSink) MinLevel() logger.Level { return s.min }
func (s *LogDBSink) Write(rec logger.Record) {
	if rec.Level < s.min || s.min == logger.LevelOff {
		return
	}
	row := model.AppLog{
		CreatedAt: rec.Time,
		Level:     logger.LevelName(rec.Level),
		Module:    rec.Module,
		Message:   rec.Message,
		RequestID: rec.RequestID,
		IPAddress: rec.IPAddress,
	}
	select {
	case <-s.stopped:
		return
	default:
	}
	select {
	case s.in <- row:
	default:
		// 缓冲满直接丢弃,避免反压拖慢业务
		s.dropped.Add(1)
	}
}

func (s *LogDBSink) Close() error {
	s.stopOnce.Do(func() { close(s.stopped) })
	<-s.done
	return nil
}

func (s *LogDBSink) run(batchSize int, flushEvery time.Duration) {
	defer close(s.done)
	ticker := time.NewTicker(flushEvery)
	defer ticker.Stop()
	batch := make([]model.AppLog, 0, batchSize)
	flush := func() {
		if len(batch) == 0 {
			return
		}
		if err := s.db.Create(&batch).Error; err != nil {
			// 不通过自身 logger 防止反复递归: 用 stdout 的 default logger 即可
			s.bufLog.Warnf("flush app logs failed: %v", err)
		}
		if d := s.dropped.Swap(0); d > 0 {
			s.bufLog.Warnf("dropped %d app log(s): db sink buffer full", d)
		}
		batch = batch[:0]
	}
	for {
		select {
		case <-s.stopped:
			// drain remaining
			for {
				select {
				case rec := <-s.in:
					batch = append(batch, rec)
					if len(batch) >= batchSize {
						flush()
					}
				default:
					flush()
					return
				}
			}
		case rec := <-s.in:
			batch = append(batch, rec)
			if len(batch) >= batchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}

// LogStreamSink 把每条日志直接 fan-out 给 LogStreamHub,与 DB 持久化解耦:
// 即使 DB sink 被关闭或攒批延迟,SSE 实时流也始终实时。
// 自身只负责广播,过滤交给各订阅者的 filter,故 MinLevel 取最低。
type LogStreamSink struct {
	name string
	min  logger.Level
	hub  *LogStreamHub
}

func NewLogStreamSink(hub *LogStreamHub, name string, min logger.Level) *LogStreamSink {
	if name == "" {
		name = "stream"
	}
	return &LogStreamSink{name: name, min: min, hub: hub}
}

func (s *LogStreamSink) Name() string           { return s.name }
func (s *LogStreamSink) MinLevel() logger.Level { return s.min }

func (s *LogStreamSink) Write(rec logger.Record) {
	if s.hub == nil || rec.Level < s.min || s.min == logger.LevelOff {
		return
	}
	s.hub.Publish(model.AppLog{
		CreatedAt: rec.Time,
		Level:     logger.LevelName(rec.Level),
		Module:    rec.Module,
		Message:   rec.Message,
		RequestID: rec.RequestID,
		IPAddress: rec.IPAddress,
	})
}

func (s *LogStreamSink) Close() error { return nil }
