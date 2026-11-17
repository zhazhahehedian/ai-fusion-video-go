package resource

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// NewCommandID returns a UUID-like string without hyphens (36-char with hyphens for DB CHAR(36)).
func NewCommandID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("生成命令 ID: %w", err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	hexed := hex.EncodeToString(b[:])
	return fmt.Sprintf("%s-%s-%s-%s-%s", hexed[0:8], hexed[8:12], hexed[12:16], hexed[16:20], hexed[20:32]), nil
}

func NewEventID() (string, error) {
	return NewCommandID()
}
