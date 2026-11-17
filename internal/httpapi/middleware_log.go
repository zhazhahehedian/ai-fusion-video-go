package httpapi

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"
	"unicode"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/logging"
)

// AccessLog 透传或生成 X-Request-ID，把带 request_id 的 logger 放进 context，并在请求结束后记一条访问日志。
func AccessLog(base *slog.Logger) func(http.Handler) http.Handler {
	if base == nil {
		panic("httpapi.AccessLog: logger 未注入")
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := incomingRequestID(r.Header.Get("X-Request-ID"))
			if id == "" {
				generated, err := logging.NewRequestID()
				if err != nil {
					base.ErrorContext(r.Context(), "生成请求标识失败", "err", err)
					WriteErr(w, http.StatusInternalServerError, 500, "服务器内部错误")
					return
				}
				id = generated
			}
			w.Header().Set("X-Request-ID", id)
			logger := base.With(slog.String("request_id", id))
			ctx := logging.WithRequestID(logging.NewContext(r.Context(), logger), id)
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()
			next.ServeHTTP(ww, r.WithContext(ctx))
			status := ww.Status()
			if status == 0 {
				status = http.StatusOK
			}
			logging.From(ctx).InfoContext(ctx, "http access",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", status),
				slog.Int64("duration_ms", time.Since(start).Milliseconds()),
			)
		})
	}
}

func incomingRequestID(raw string) string {
	raw = trimHeader(raw)
	if raw == "" || len(raw) > 128 {
		return ""
	}
	for _, r := range raw {
		if r > unicode.MaxASCII || !isRequestIDChar(r) {
			return ""
		}
	}
	return raw
}

func isRequestIDChar(r rune) bool {
	return r == '-' || r == '_' || r == '.' || r == ':' ||
		(r >= '0' && r <= '9') || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')
}

func trimHeader(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}

// Recover 在访问日志中间件之内接住 panic，用请求 logger 记录一次。
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}
			logging.From(r.Context()).ErrorContext(r.Context(), "请求处理 panic",
				"panic", rec,
				"stack", string(debug.Stack()),
			)
			if ww, ok := w.(middleware.WrapResponseWriter); ok && ww.Status() != 0 {
				return
			}
			WriteErr(w, http.StatusInternalServerError, 500, "服务器内部错误")
		}()
		next.ServeHTTP(w, r)
	})
}
