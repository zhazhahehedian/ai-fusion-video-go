package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/apperr"
)

func optionalNullString(v any) (sql.NullString, error) {
	if v == nil {
		return sql.NullString{}, nil
	}
	s, ok := v.(string)
	if !ok {
		return sql.NullString{}, apperr.BadRequest("字段类型必须是字符串")
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return sql.NullString{}, nil
	}
	return sql.NullString{String: s, Valid: true}, nil
}

func optionalNullInt32(v any) (sql.NullInt32, error) {
	if v == nil {
		return sql.NullInt32{}, nil
	}
	n, err := asInt64Value(v)
	if err != nil {
		return sql.NullInt32{}, err
	}
	if n < math.MinInt32 || n > math.MaxInt32 {
		return sql.NullInt32{}, apperr.BadRequest("数值超出范围")
	}
	return sql.NullInt32{Int32: int32(n), Valid: true}, nil
}

func optionalJSON(v any) (json.RawMessage, error) {
	if v == nil {
		return json.RawMessage("{}"), nil
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, apperr.BadRequest(fmt.Sprintf("字段不是合法 JSON: %v", err))
	}
	return raw, nil
}

func optionalJSONArray(v any) (json.RawMessage, error) {
	if v == nil {
		return json.RawMessage("[]"), nil
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, apperr.BadRequest(fmt.Sprintf("字段不是合法 JSON: %v", err))
	}
	return raw, nil
}

func optionalNullInt64(v any) (sql.NullInt64, error) {
	if v == nil {
		return sql.NullInt64{}, nil
	}
	n, err := asInt64Value(v)
	if err != nil {
		return sql.NullInt64{}, err
	}
	if n <= 0 {
		return sql.NullInt64{}, nil
	}
	return sql.NullInt64{Int64: n, Valid: true}, nil
}

func asInt64(result any, key string) (int64, bool) {
	m, ok := result.(map[string]any)
	if !ok {
		return 0, false
	}
	n, err := asInt64Value(m[key])
	if err != nil {
		return 0, false
	}
	return n, true
}
