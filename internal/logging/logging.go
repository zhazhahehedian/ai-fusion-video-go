package logging

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
	"log/slog"
	"os"
	"sync"
)

type ctxKey int

const (
	boxKey ctxKey = iota
	requestIDKey
)

// loggerBox 让外层中间件和内层鉴权共享同一份 logger。
// WithAttrs 替换指针后，已经持有 context 的访问日志也能看到后加的字段。
type loggerBox struct {
	mu     sync.Mutex
	logger *slog.Logger
}

// New 按格式创建进程日志。format 只能是 text 或 json，由 config 校验。
func New(w io.Writer, level slog.Level, format string) *slog.Logger {
	if w == nil {
		w = os.Stderr
	}
	opts := &slog.HandlerOptions{Level: level}
	var handler slog.Handler
	switch format {
	case "json":
		handler = slog.NewJSONHandler(w, opts)
	case "text":
		handler = slog.NewTextHandler(w, opts)
	default:
		panic("logging.New: 未知日志格式 " + format)
	}
	return slog.New(handler)
}

// NewContext 把 logger 放进 context，供后续 WithAttrs 就地追加字段。
func NewContext(ctx context.Context, logger *slog.Logger) context.Context {
	if logger == nil {
		panic("logging.NewContext: logger 未注入")
	}
	return context.WithValue(ctx, boxKey, &loggerBox{logger: logger})
}

// From 取出当前请求的 logger。没有注入时返回进程默认 logger。
func From(ctx context.Context) *slog.Logger {
	box, ok := ctx.Value(boxKey).(*loggerBox)
	if !ok || box == nil {
		return slog.Default()
	}
	box.mu.Lock()
	logger := box.logger
	box.mu.Unlock()
	if logger == nil {
		return slog.Default()
	}
	return logger
}

// WithAttrs 给当前 context 里的 logger 追加结构化字段。
func WithAttrs(ctx context.Context, args ...any) {
	box, ok := ctx.Value(boxKey).(*loggerBox)
	if !ok || box == nil || box.logger == nil {
		panic("logging.WithAttrs: context 中没有 logger")
	}
	box.mu.Lock()
	box.logger = box.logger.With(args...)
	box.mu.Unlock()
}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

func RequestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}

func NewRequestID() (string, error) {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf[:]), nil
}
