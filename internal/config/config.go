package config

import (
	"fmt"
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

	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:          getenv("HTTP_ADDR", ":18080"),
		DatabaseURL:       getenv("DATABASE_URL", "fusion:123456@tcp(127.0.0.1:43306)/ai_fusion_video?parseTime=true&loc=Local"),
		RedisAddr:         getenv("REDIS_ADDR", "127.0.0.1:46379"),
		RedisPassword:     getenv("REDIS_PASSWORD", "123456"),
		RedisDB:           0,
		DBMaxOpenConns:    10,
		DBMaxIdleConns:    5,
		DBConnMaxLifetime: 30 * time.Minute,
	}
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
