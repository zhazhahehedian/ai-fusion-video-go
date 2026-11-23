package generation

import (
	"encoding/json"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/apperr"
)

// urlListValue 接受两种请求体：JSON 数组，或 Java 提交体里的 JSON 数组字符串。
// 其它形状直接拒绝，不把单个地址悄悄包成数组。
func urlListValue(raw *json.RawMessage, key string) (any, bool, error) {
	if raw == nil {
		return nil, false, nil
	}
	if string(*raw) == "null" {
		return nil, true, nil
	}
	var arr []any
	if err := json.Unmarshal(*raw, &arr); err == nil {
		return arr, true, nil
	}
	var text string
	if err := json.Unmarshal(*raw, &text); err != nil {
		return nil, false, apperr.BadRequest(key + " 必须为地址数组或其 JSON 字符串")
	}
	if text == "" {
		return nil, true, nil
	}
	if err := json.Unmarshal([]byte(text), &arr); err != nil {
		return nil, false, apperr.BadRequest(key + " 必须为地址数组或其 JSON 字符串")
	}
	return arr, true, nil
}
