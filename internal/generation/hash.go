package generation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
)

func ScopeKey(userID int64) string {
	sum := sha256.Sum256([]byte("user:" + strconv.FormatInt(userID, 10)))
	return hex.EncodeToString(sum[:])
}

func RequestHash(kind, prompt string, userID int64) string {
	payload, err := json.Marshal(map[string]any{
		"kind":   kind,
		"prompt": prompt,
		"actor":  userID,
	})
	if err != nil {
		panic(fmt.Sprintf("序列化请求摘要失败: %v", err))
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}
