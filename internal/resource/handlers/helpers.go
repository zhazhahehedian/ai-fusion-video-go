package handlers

import (
	"database/sql"
	"encoding/json"
	"strings"
)

func optionalNullString(v any) sql.NullString {
	if v == nil {
		return sql.NullString{}
	}
	s, ok := v.(string)
	if !ok {
		return sql.NullString{}
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func optionalNullInt32(v any) sql.NullInt32 {
	if v == nil {
		return sql.NullInt32{}
	}
	n, err := asInt64Value(v)
	if err != nil {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: int32(n), Valid: true}
}

func optionalJSON(v any) json.RawMessage {
	if v == nil {
		return json.RawMessage("{}")
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage("{}")
	}
	return raw
}

func optionalJSONArray(v any) json.RawMessage {
	if v == nil {
		return json.RawMessage("[]")
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage("[]")
	}
	return raw
}

func optionalNullInt64(v any) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	n, err := asInt64Value(v)
	if err != nil || n <= 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: n, Valid: true}
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
