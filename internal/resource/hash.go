package resource

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// HashCommand builds a stable SHA-256 digest of operation + command + actor scope.
// Matches Java admit hash intent: same request_id with different body → conflict.
func HashCommand(cmd Command, actorUserID int64, origin Origin) (string, error) {
	payload := map[string]any{
		"operation": cmd.Operation(),
		"command":   cmd,
		"context": map[string]any{
			"actorUserId": actorUserID,
			"origin":      origin,
		},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}
