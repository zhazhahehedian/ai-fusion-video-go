package config

import (
	"log/slog"
	"testing"
)

func TestResolveLogDefaults(t *testing.T) {
	env, level, format, err := resolveLog("", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if env != "development" || level != slog.LevelInfo || format != "text" {
		t.Fatalf("development default = %s %v %s", env, level, format)
	}

	env, level, format, err = resolveLog("production", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if env != "production" || level != slog.LevelInfo || format != "json" {
		t.Fatalf("production default = %s %v %s", env, level, format)
	}

	_, level, format, err = resolveLog("development", "debug", "json")
	if err != nil {
		t.Fatal(err)
	}
	if level != slog.LevelDebug || format != "json" {
		t.Fatalf("override = %v %s", level, format)
	}
}

func TestResolveLogRejectsUnknown(t *testing.T) {
	if _, _, _, err := resolveLog("staging", "", ""); err == nil {
		t.Fatal("非法 APP_ENV 被接受")
	}
	if _, _, _, err := resolveLog("development", "verbose", ""); err == nil {
		t.Fatal("非法 LOG_LEVEL 被接受")
	}
	if _, _, _, err := resolveLog("production", "info", "console"); err == nil {
		t.Fatal("非法 LOG_FORMAT 被接受")
	}
}
