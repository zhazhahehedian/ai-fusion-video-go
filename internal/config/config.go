package config

import (
	"os"
	"strings"
)

// Config holds process-level settings. Phase 1 only needs HTTP listen addr
// and a pointer to the Java reference tree.
type Config struct {
	HTTPAddr string
	JavaRoot string
}

func Load() Config {
	return Config{
		HTTPAddr: getenv("HTTP_ADDR", ":18080"),
		JavaRoot: getenv("JAVA_ROOT", "E:/workspace/JProject/ai-fusion-video"),
	}
}

func getenv(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}
