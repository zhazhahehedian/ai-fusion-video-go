package httpapi

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/logging"
)

func TestAccessLogPropagatesRequestID(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	var seen string
	handler := AccessLog(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = logging.RequestID(r.Context())
		logging.From(r.Context()).InfoContext(r.Context(), "inside")
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/project", nil)
	req.Header.Set("X-Request-ID", "req-abc")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rr.Code)
	}
	if got := rr.Header().Get("X-Request-ID"); got != "req-abc" {
		t.Fatalf("response request id = %q", got)
	}
	if seen != "req-abc" {
		t.Fatalf("context request id = %q", seen)
	}

	lines := decodeLogs(t, buf.String())
	if len(lines) != 2 {
		t.Fatalf("log lines = %#v", lines)
	}
	if lines[0]["msg"] != "inside" || lines[0]["request_id"] != "req-abc" {
		t.Fatalf("handler log = %#v", lines[0])
	}
	access := lines[1]
	if access["msg"] != "http access" || access["method"] != "POST" || access["path"] != "/api/project" || access["request_id"] != "req-abc" {
		t.Fatalf("access log = %#v", access)
	}
	if status, _ := access["status"].(float64); status != http.StatusNoContent {
		t.Fatalf("status = %#v", access["status"])
	}
	duration, ok := access["duration_ms"].(float64)
	if !ok || duration < 0 {
		t.Fatalf("duration_ms = %#v", access["duration_ms"])
	}
}

func TestAccessLogGeneratesRequestID(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	handler := AccessLog(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if logging.RequestID(r.Context()) == "" {
			t.Fatal("未生成 request id")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("X-Request-ID", "bad id\n")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	got := rr.Header().Get("X-Request-ID")
	if got == "" || got == "bad id\n" || len(got) != 32 {
		t.Fatalf("generated request id = %q", got)
	}
}

func TestAccessLogIncludesUserIDAddedDuringRequest(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	handler := AccessLog(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logging.WithAttrs(r.Context(), "user_id", int64(9))
		logging.From(r.Context()).InfoContext(r.Context(), "authed")
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/project/list", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	var authed, access map[string]any
	for _, line := range decodeLogs(t, buf.String()) {
		switch line["msg"] {
		case "authed":
			authed = line
		case "http access":
			access = line
		}
	}
	if authed["user_id"] != float64(9) || access["user_id"] != float64(9) {
		t.Fatalf("authed=%#v access=%#v", authed, access)
	}
}

func decodeLogs(t *testing.T, text string) []map[string]any {
	t.Helper()
	var out []map[string]any
	for _, line := range bytes.Split([]byte(text), []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		var item map[string]any
		if err := json.Unmarshal(line, &item); err != nil {
			t.Fatalf("日志不是 JSON: %s (%v)", line, err)
		}
		out = append(out, item)
	}
	return out
}
