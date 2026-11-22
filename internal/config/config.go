package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
)

// Config holds process-level settings.
type Config struct {
	HTTPAddr string

	DatabaseURL string

	RedisAddr     string
	RedisPassword string
	RedisDB       int

	FFmpegPath string

	// AppEnv 为 development 或 production。默认 development。
	AppEnv string
	// LogLevel 已解析。LOG_LEVEL 默认 info。
	LogLevel slog.Level
	// LogFormat 为 text 或 json。LOG_FORMAT 为空时，development 用 text，production 用 json。
	LogFormat string

	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr: getenv("HTTP_ADDR", ":18080"),
		// loc=UTC：驱动把 DATETIME 标成 UTC。OpenMySQL 会补上 time_zone=+00:00，
		// 让内容表的 CURRENT_TIMESTAMP 墙钟和这个标注一致。作业表仍写 UTC_TIMESTAMP。
		DatabaseURL:       getenv("DATABASE_URL", "fusion:123456@tcp(127.0.0.1:43306)/ai_fusion_video?parseTime=true&loc=UTC"),
		RedisAddr:         getenv("REDIS_ADDR", "127.0.0.1:46379"),
		RedisPassword:     getenv("REDIS_PASSWORD", "123456"),
		RedisDB:           0,
		FFmpegPath:        getenv("FFMPEG_PATH", "ffmpeg"),
		DBMaxOpenConns:    10,
		DBMaxIdleConns:    5,
		DBConnMaxLifetime: 30 * time.Minute,
	}
	env, level, format, err := resolveLog(os.Getenv("APP_ENV"), os.Getenv("LOG_LEVEL"), os.Getenv("LOG_FORMAT"))
	if err != nil {
		return Config{}, err
	}
	cfg.AppEnv = env
	cfg.LogLevel = level
	cfg.LogFormat = format
	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return Config{}, fmt.Errorf("DATABASE_URL 未配置")
	}
	if strings.TrimSpace(cfg.RedisAddr) == "" {
		return Config{}, fmt.Errorf("REDIS_ADDR 未配置")
	}
	return cfg, nil
}

func getenv(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}

func resolveLog(env, level, format string) (string, slog.Level, string, error) {
	env = strings.ToLower(strings.TrimSpace(env))
	if env == "" {
		env = "development"
	}
	switch env {
	case "development", "production":
	default:
		return "", 0, "", fmt.Errorf("APP_ENV 无效: %s", env)
	}

	level = strings.ToLower(strings.TrimSpace(level))
	if level == "" {
		level = "info"
	}
	var slogLevel slog.Level
	switch level {
	case "debug":
		slogLevel = slog.LevelDebug
	case "info":
		slogLevel = slog.LevelInfo
	case "warn", "warning":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		return "", 0, "", fmt.Errorf("LOG_LEVEL 无效: %s", level)
	}

	format = strings.ToLower(strings.TrimSpace(format))
	if format == "" {
		if env == "production" {
			format = "json"
		} else {
			format = "text"
		}
	}
	switch format {
	case "text", "json":
	default:
		return "", 0, "", fmt.Errorf("LOG_FORMAT 无效: %s", format)
	}
	return env, slogLevel, format, nil
}
