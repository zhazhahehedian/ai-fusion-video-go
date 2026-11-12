package config

import (
	"os"
	"strings"
)

// Config holds process-level settings.
type Config struct {
	HTTPAddr string
}

func Load() Config {
	return Config{
		HTTPAddr: getenv("HTTP_ADDR", ":18080"),
	}
}

func getenv(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}
