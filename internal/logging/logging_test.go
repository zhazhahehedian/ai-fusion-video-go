package logging

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestFromReturnsDefaultWithoutContext(t *testing.T) {
	if From(context.Background()) != slog.Default() {
		t.Fatal("未注入时应返回 slog.Default")
	}
}

func TestWithAttrsVisibleToExistingContext(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	ctx := NewContext(context.Background(), logger)

	From(ctx).InfoContext(ctx, "before")
	WithAttrs(ctx, "request_id", "req-7", "user_id", int64(42))
	From(ctx).InfoContext(ctx, "after")

	text := buf.String()
	before, after, ok := strings.Cut(text, "\n")
	if !ok || !strings.Contains(before, "msg=before") || strings.Contains(before, "user_id=42") {
		t.Fatalf("追加前的日志不应带 user_id: %s", before)
	}
	if !strings.Contains(after, "msg=after") || !strings.Contains(after, "request_id=req-7") || !strings.Contains(after, "user_id=42") {
		t.Fatalf("追加后的日志应带 request_id 和 user_id: %s", after)
	}
}

func TestNewRequestID(t *testing.T) {
	id, err := NewRequestID()
	if err != nil {
		t.Fatal(err)
	}
	if len(id) != 32 {
		t.Fatalf("id = %q", id)
	}
	other, err := NewRequestID()
	if err != nil {
		t.Fatal(err)
	}
	if id == other {
		t.Fatal("两次生成了相同的 request id")
	}
}
