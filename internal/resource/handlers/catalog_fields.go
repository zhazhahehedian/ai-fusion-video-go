package handlers

import (
	"database/sql"
	"encoding/json"
	"math"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/apperr"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/sqlcdb"
)

var assetTypes = map[string]struct{}{
	"character": {}, "scene": {}, "prop": {}, "vehicle": {}, "building": {},
	"costume": {}, "effect": {}, "material": {}, "image": {},
}

var assetItemTypes = map[string]struct{}{
	"front": {}, "side": {}, "back": {}, "detail": {}, "expression": {},
	"pose": {}, "variant": {}, "original": {},
}

var videoModes = map[string]struct{}{
	"text2video": {}, "image2video": {},
}

func rejectUnknown(fields map[string]any, allowed []string) error {
	ok := make(map[string]struct{}, len(allowed))
	for _, key := range allowed {
		ok[key] = struct{}{}
	}
	for key := range fields {
		if _, exists := ok[key]; !exists {
			return apperr.BadRequest("不允许修改字段：" + key)
		}
	}
	return nil
}

func present(fields map[string]any, key string) (any, bool) {
	v, ok := fields[key]
	return v, ok
}

func checkText(s, key string, maxRunes int) error {
	if utf8.RuneCountInString(s) > maxRunes || len(s) > 65535 {
		return apperr.BadRequest(key + " 文本格式或长度无效")
	}
	return nil
}

func readText(v any, key string, maxRunes int, asURL bool) (sql.NullString, error) {
	if v == nil {
		return sql.NullString{}, nil
	}
	s, ok := v.(string)
	if !ok {
		return sql.NullString{}, apperr.BadRequest(key + " 必须是字符串")
	}
	if err := checkText(s, key, maxRunes); err != nil {
		return sql.NullString{}, err
	}
	if strings.TrimSpace(s) == "" {
		return sql.NullString{}, nil
	}
	if asURL && !validMediaURL(s) {
		return sql.NullString{}, apperr.BadRequest(key + " 必须是已上传路径或 HTTP(S) 地址")
	}
	return sql.NullString{String: s, Valid: true}, nil
}

func validMediaURL(s string) bool {
	if strings.ContainsAny(s, "\r\n\t") {
		return false
	}
	for _, r := range s {
		if r < 0x20 || r == 0x7f {
			return false
		}
	}
	if strings.HasPrefix(s, "/") && !strings.HasPrefix(s, "//") {
		return !strings.Contains(s, "://")
	}
	u, err := url.Parse(s)
	if err != nil || u.User != nil || u.Host == "" {
		return false
	}
	scheme := strings.ToLower(u.Scheme)
	return scheme == "http" || scheme == "https"
}

func readStatus(v any, key string, max int32) (sql.NullInt32, error) {
	if v == nil {
		return sql.NullInt32{}, apperr.BadRequest(key + " 不能清空")
	}
	n, err := requiredNonNegInt(v, key)
	if err != nil {
		return sql.NullInt32{}, err
	}
	if n > max {
		return sql.NullInt32{}, apperr.BadRequest(key + " 整数范围无效")
	}
	return sql.NullInt32{Int32: n, Valid: true}, nil
}

func readRange(v any, key string, min, max int32) (sql.NullInt32, error) {
	if v == nil {
		return sql.NullInt32{}, nil
	}
	n, err := asInt64Value(v)
	if err != nil || n < int64(min) || n > int64(max) {
		return sql.NullInt32{}, apperr.BadRequest(key + " 整数范围无效")
	}
	return sql.NullInt32{Int32: int32(n), Valid: true}, nil
}

func readOptionalID(v any, key string) (sql.NullInt64, error) {
	if v == nil {
		return sql.NullInt64{}, nil
	}
	n, err := asInt64Value(v)
	if err != nil || n <= 0 {
		return sql.NullInt64{}, apperr.BadRequest(key + " 必须是正整数 ID")
	}
	return sql.NullInt64{Int64: n, Valid: true}, nil
}

func readBoolTiny(v any, key string) (sql.NullInt16, error) {
	if v == nil {
		return sql.NullInt16{}, nil
	}
	b, ok := v.(bool)
	if !ok {
		return sql.NullInt16{}, apperr.BadRequest(key + " 必须为布尔值")
	}
	if b {
		return sql.NullInt16{Int16: 1, Valid: true}, nil
	}
	return sql.NullInt16{Int16: 0, Valid: true}, nil
}

func readDuration(v any) (sql.NullString, error) {
	if v == nil {
		return sql.NullString{}, nil
	}
	var n float64
	switch x := v.(type) {
	case float64:
		n = x
	case float32:
		n = float64(x)
	case int:
		n = float64(x)
	case int32:
		n = float64(x)
	case int64:
		n = float64(x)
	case json.Number:
		f, err := x.Float64()
		if err != nil {
			return sql.NullString{}, apperr.BadRequest("duration 必须为数字")
		}
		n = f
	default:
		return sql.NullString{}, apperr.BadRequest("duration 必须为数字")
	}
	if math.IsNaN(n) || math.IsInf(n, 0) || n < 0 || n > 86400 {
		return sql.NullString{}, apperr.BadRequest("duration 范围为 0 至 86400 秒，最多两位小数")
	}
	scaled := math.Round(n * 100)
	if math.Abs(n*100-scaled) > 1e-6 {
		return sql.NullString{}, apperr.BadRequest("duration 范围为 0 至 86400 秒，最多两位小数")
	}
	return sql.NullString{String: strconv.FormatFloat(scaled/100, 'f', 2, 64), Valid: true}, nil
}

func readIDList(v any, key string) (json.RawMessage, error) {
	if v == nil {
		return nil, apperr.BadRequest(key + " 不能为 null，清空关联请使用空数组")
	}
	arr, ok := v.([]any)
	if !ok || len(arr) > 500 {
		return nil, apperr.BadRequest(key + " 必须为最多 500 项的 ID 数组")
	}
	ids := make([]int64, 0, len(arr))
	seen := make(map[int64]struct{}, len(arr))
	for _, item := range arr {
		id, err := asInt64Value(item)
		if err != nil || id <= 0 {
			return nil, apperr.BadRequest(key + " 必须是正整数 ID")
		}
		if _, exists := seen[id]; exists {
			return nil, apperr.BadRequest(key + " 不能重复")
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	raw, err := json.Marshal(ids)
	if err != nil {
		return nil, apperr.BadRequest(key + " 不是合法 JSON")
	}
	return raw, nil
}

func readJSONValue(v any, key string, wantArray bool) (json.RawMessage, error) {
	if v == nil {
		return nil, nil
	}
	raw, err := json.Marshal(v)
	if err != nil || len(raw) > 262144 {
		return nil, apperr.BadRequest(key + " JSON 格式或长度无效")
	}
	if wantArray {
		var arr []any
		if err := json.Unmarshal(raw, &arr); err != nil || len(arr) > 200 {
			return nil, apperr.BadRequest(key + " JSON 格式或长度无效")
		}
		return raw, nil
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil || len(obj) > 200 {
		return nil, apperr.BadRequest(key + " JSON 格式或长度无效")
	}
	return raw, nil
}

func readURLList(v any, key string) (json.RawMessage, error) {
	if v == nil {
		return nil, nil
	}
	arr, ok := v.([]any)
	if !ok || len(arr) > 20 {
		return nil, apperr.BadRequest(key + " 必须为最多 20 项的地址数组")
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		s, ok := item.(string)
		if !ok || !validMediaURL(s) || utf8.RuneCountInString(s) > 1024 {
			return nil, apperr.BadRequest(key + " 必须是已上传路径或 HTTP(S) 地址")
		}
		out = append(out, s)
	}
	raw, err := json.Marshal(out)
	if err != nil {
		return nil, apperr.BadRequest(key + " 不是合法 JSON")
	}
	return raw, nil
}

func assignText(fields map[string]any, key string, max int, asURL bool, dst *sql.NullString, changed *[]string) error {
	v, ok := present(fields, key)
	if !ok {
		return nil
	}
	got, err := readText(v, key, max, asURL)
	if err != nil {
		return err
	}
	*dst = got
	*changed = append(*changed, key)
	return nil
}

func applyStoryboardFields(row *sqlcdb.AfvStoryboard, fields map[string]any, creating bool) ([]string, error) {
	allowed := []string{"description", "customColumns", "status"}
	if creating {
		allowed = append(allowed, "title", "scope", "scriptId")
	}
	if err := rejectUnknown(fields, allowed); err != nil {
		return nil, err
	}
	var changed []string
	if err := assignText(fields, "title", 256, false, &row.Title, &changed); err != nil {
		return nil, err
	}
	if err := assignText(fields, "description", 65535, false, &row.Description, &changed); err != nil {
		return nil, err
	}
	if v, ok := present(fields, "customColumns"); ok {
		raw, err := readJSONValue(v, "customColumns", true)
		if err != nil {
			return nil, err
		}
		row.CustomColumns = raw
		changed = append(changed, "customColumns")
	}
	if v, ok := present(fields, "status"); ok {
		status, err := readStatus(v, "status", 1)
		if err != nil {
			return nil, err
		}
		row.Status = status
		changed = append(changed, "status")
	}
	if v, ok := present(fields, "scope"); ok {
		scope, err := readRange(v, "scope", 1, 3)
		if err != nil || !scope.Valid {
			return nil, apperr.BadRequest("scope 整数范围无效")
		}
		row.Scope = scope
		changed = append(changed, "scope")
	}
	if v, ok := present(fields, "scriptId"); ok {
		id, err := readOptionalID(v, "scriptId")
		if err != nil {
			return nil, err
		}
		row.ScriptID = id
		changed = append(changed, "scriptId")
	}
	if !creating && len(changed) == 0 {
		return nil, apperr.BadRequest("没有可更新的字段")
	}
	return changed, nil
}

func applyEpisodeFields(row *sqlcdb.AfvStoryboardEpisode, fields map[string]any, creating bool) ([]string, error) {
	allowed := []string{"scriptEpisodeId", "episodeNumber", "title", "synopsis", "status"}
	if creating {
		allowed = append(allowed, "sortOrder")
	}
	if err := rejectUnknown(fields, allowed); err != nil {
		return nil, err
	}
	var changed []string
	if v, ok := present(fields, "scriptEpisodeId"); ok {
		id, err := readOptionalID(v, "scriptEpisodeId")
		if err != nil {
			return nil, err
		}
		row.ScriptEpisodeID = id
		changed = append(changed, "scriptEpisodeId")
	}
	if v, ok := present(fields, "episodeNumber"); ok {
		n, err := readRange(v, "episodeNumber", 1, 1000000)
		if err != nil || !n.Valid {
			return nil, apperr.BadRequest("episodeNumber 整数范围无效")
		}
		row.EpisodeNumber = n
		changed = append(changed, "episodeNumber")
	} else if creating {
		return nil, apperr.BadRequest("创建分镜集必须提供 episodeNumber")
	}
	if err := assignText(fields, "title", 200, false, &row.Title, &changed); err != nil {
		return nil, err
	}
	if err := assignText(fields, "synopsis", 65535, false, &row.Synopsis, &changed); err != nil {
		return nil, err
	}
	if v, ok := present(fields, "sortOrder"); ok {
		n, err := readStatus(v, "sortOrder", 1000000)
		if err != nil {
			return nil, err
		}
		row.SortOrder = n
		changed = append(changed, "sortOrder")
	}
	if v, ok := present(fields, "status"); ok {
		n, err := readStatus(v, "status", 1)
		if err != nil {
			return nil, err
		}
		row.Status = n
		changed = append(changed, "status")
	}
	if !creating && len(changed) == 0 {
		return nil, apperr.BadRequest("没有可更新的字段")
	}
	return changed, nil
}

func applySceneFields(row *sqlcdb.AfvStoryboardScene, fields map[string]any, creating bool) ([]string, error) {
	allowed := []string{"sceneNumber", "sceneHeading", "location", "timeOfDay", "intExt", "status"}
	if creating {
		allowed = append(allowed, "sortOrder")
	}
	if err := rejectUnknown(fields, allowed); err != nil {
		return nil, err
	}
	var changed []string
	if err := assignText(fields, "sceneNumber", 20, false, &row.SceneNumber, &changed); err != nil {
		return nil, err
	}
	if err := assignText(fields, "sceneHeading", 200, false, &row.SceneHeading, &changed); err != nil {
		return nil, err
	}
	if err := assignText(fields, "location", 200, false, &row.Location, &changed); err != nil {
		return nil, err
	}
	if err := assignText(fields, "timeOfDay", 20, false, &row.TimeOfDay, &changed); err != nil {
		return nil, err
	}
	if err := assignText(fields, "intExt", 20, false, &row.IntExt, &changed); err != nil {
		return nil, err
	}
	if v, ok := present(fields, "sortOrder"); ok {
		n, err := readStatus(v, "sortOrder", 1000000)
		if err != nil {
			return nil, err
		}
		row.SortOrder = n
		changed = append(changed, "sortOrder")
	}
	if v, ok := present(fields, "status"); ok {
		n, err := readStatus(v, "status", 1)
		if err != nil {
			return nil, err
		}
		row.Status = n
		changed = append(changed, "status")
	}
	if !creating && len(changed) == 0 {
		return nil, apperr.BadRequest("没有可更新的字段")
	}
	return changed, nil
}

func applyShotFields(row *sqlcdb.AfvStoryboardItem, fields map[string]any, creating bool) ([]string, error) {
	allowed := []string{
		"shotNumber", "shotType", "duration", "content", "sceneExpectation", "sound", "dialogue",
		"soundEffect", "music", "cameraMovement", "cameraAngle", "cameraEquipment", "focalLength", "transition",
		"characterIds", "sceneAssetItemId", "propIds", "remark", "customData", "aiGenerated", "status",
		"imageUrl", "referenceImageUrl", "videoUrl", "generatedImageUrl", "firstFrameImageUrl", "lastFrameImageUrl",
		"firstFramePrompt", "lastFramePrompt", "generatedVideoUrl", "videoPrompt",
	}
	if creating {
		allowed = append(allowed, "sortOrder")
	}
	if err := rejectUnknown(fields, allowed); err != nil {
		return nil, err
	}
	var changed []string
	texts := []struct {
		key string
		max int
		url bool
		dst *sql.NullString
	}{
		{"shotNumber", 32, false, &row.ShotNumber},
		{"shotType", 32, false, &row.ShotType},
		{"content", 65535, false, &row.Content},
		{"sceneExpectation", 65535, false, &row.SceneExpectation},
		{"sound", 65535, false, &row.Sound},
		{"dialogue", 65535, false, &row.Dialogue},
		{"soundEffect", 512, false, &row.SoundEffect},
		{"music", 512, false, &row.Music},
		{"cameraMovement", 64, false, &row.CameraMovement},
		{"cameraAngle", 64, false, &row.CameraAngle},
		{"cameraEquipment", 64, false, &row.CameraEquipment},
		{"focalLength", 64, false, &row.FocalLength},
		{"transition", 64, false, &row.Transition},
		{"remark", 65535, false, &row.Remark},
		{"imageUrl", 1024, true, &row.ImageUrl},
		{"referenceImageUrl", 1024, true, &row.ReferenceImageUrl},
		{"videoUrl", 1024, true, &row.VideoUrl},
		{"generatedImageUrl", 1024, true, &row.GeneratedImageUrl},
		{"firstFrameImageUrl", 1024, true, &row.FirstFrameImageUrl},
		{"lastFrameImageUrl", 1024, true, &row.LastFrameImageUrl},
		{"generatedVideoUrl", 1024, true, &row.GeneratedVideoUrl},
		{"firstFramePrompt", 65535, false, &row.FirstFramePrompt},
		{"lastFramePrompt", 65535, false, &row.LastFramePrompt},
		{"videoPrompt", 65535, false, &row.VideoPrompt},
	}
	for _, spec := range texts {
		if err := assignText(fields, spec.key, spec.max, spec.url, spec.dst, &changed); err != nil {
			return nil, err
		}
	}
	if v, ok := present(fields, "duration"); ok {
		got, err := readDuration(v)
		if err != nil {
			return nil, err
		}
		row.Duration = got
		changed = append(changed, "duration")
	}
	if v, ok := present(fields, "characterIds"); ok {
		raw, err := readIDList(v, "characterIds")
		if err != nil {
			return nil, err
		}
		row.CharacterIds = raw
		changed = append(changed, "characterIds")
	}
	if v, ok := present(fields, "propIds"); ok {
		raw, err := readIDList(v, "propIds")
		if err != nil {
			return nil, err
		}
		row.PropIds = sql.NullString{String: string(raw), Valid: true}
		changed = append(changed, "propIds")
	}
	if v, ok := present(fields, "sceneAssetItemId"); ok {
		id, err := readOptionalID(v, "sceneAssetItemId")
		if err != nil {
			return nil, err
		}
		row.SceneAssetItemID = id
		changed = append(changed, "sceneAssetItemId")
	}
	if v, ok := present(fields, "customData"); ok {
		raw, err := readJSONValue(v, "customData", false)
		if err != nil {
			return nil, err
		}
		row.CustomData = raw
		changed = append(changed, "customData")
	}
	if v, ok := present(fields, "aiGenerated"); ok {
		flag, err := readBoolTiny(v, "aiGenerated")
		if err != nil {
			return nil, err
		}
		row.AiGenerated = flag
		changed = append(changed, "aiGenerated")
	}
	if v, ok := present(fields, "status"); ok {
		n, err := readStatus(v, "status", 1)
		if err != nil {
			return nil, err
		}
		row.Status = n
		changed = append(changed, "status")
	}
	if v, ok := present(fields, "sortOrder"); ok {
		n, err := readStatus(v, "sortOrder", 1000000)
		if err != nil {
			return nil, err
		}
		row.SortOrder = n
		changed = append(changed, "sortOrder")
	}
	if !creating && len(changed) == 0 {
		return nil, apperr.BadRequest("没有可更新的字段")
	}
	return changed, nil
}

func applyAssetFields(row *sqlcdb.AfvAsset, fields map[string]any, creating bool) ([]string, error) {
	allowed := []string{"type", "name", "description", "coverUrl", "properties", "tags", "sourceType", "aiPrompt", "status"}
	if err := rejectUnknown(fields, allowed); err != nil {
		return nil, err
	}
	var changed []string
	if v, ok := present(fields, "type"); ok {
		s, ok := v.(string)
		if !ok {
			return nil, apperr.BadRequest("type 必须是字符串")
		}
		if _, exists := assetTypes[s]; !exists {
			return nil, apperr.BadRequest("不支持的资产类型")
		}
		row.Type = sql.NullString{String: s, Valid: true}
		changed = append(changed, "type")
	} else if creating {
		return nil, apperr.BadRequest("创建资产必须提供 type 和 name")
	}
	if v, ok := present(fields, "name"); ok {
		s, ok := v.(string)
		if !ok {
			return nil, apperr.BadRequest("name 必须是字符串")
		}
		s = strings.TrimSpace(s)
		if s == "" {
			return nil, apperr.BadRequest("name 不能清空")
		}
		if err := checkText(s, "name", 128); err != nil {
			return nil, err
		}
		row.Name = s
		changed = append(changed, "name")
	} else if creating {
		return nil, apperr.BadRequest("创建资产必须提供 type 和 name")
	}
	if err := assignText(fields, "description", 65535, false, &row.Description, &changed); err != nil {
		return nil, err
	}
	if err := assignText(fields, "coverUrl", 1024, true, &row.CoverUrl, &changed); err != nil {
		return nil, err
	}
	if err := assignText(fields, "aiPrompt", 65535, false, &row.AiPrompt, &changed); err != nil {
		return nil, err
	}
	if v, ok := present(fields, "properties"); ok {
		raw, err := readJSONValue(v, "properties", false)
		if err != nil {
			return nil, err
		}
		row.Properties = raw
		changed = append(changed, "properties")
	}
	if v, ok := present(fields, "tags"); ok {
		raw, err := readJSONValue(v, "tags", true)
		if err != nil {
			return nil, err
		}
		row.Tags = raw
		changed = append(changed, "tags")
	}
	if v, ok := present(fields, "sourceType"); ok {
		n, err := readRange(v, "sourceType", 1, 2)
		if err != nil || !n.Valid {
			return nil, apperr.BadRequest("sourceType 整数范围无效")
		}
		row.SourceType = n
		changed = append(changed, "sourceType")
	}
	if v, ok := present(fields, "status"); ok {
		n, err := readStatus(v, "status", 1)
		if err != nil {
			return nil, err
		}
		row.Status = n
		changed = append(changed, "status")
	}
	if !creating && len(changed) == 0 {
		return nil, apperr.BadRequest("没有可更新的字段")
	}
	return changed, nil
}

func applyAssetItemFields(row *sqlcdb.AfvAssetItem, fields map[string]any, creating bool) ([]string, error) {
	allowed := []string{"itemType", "name", "imageUrl", "thumbnailUrl", "properties", "sortOrder", "sourceType", "aiPrompt"}
	if err := rejectUnknown(fields, allowed); err != nil {
		return nil, err
	}
	var changed []string
	if v, ok := present(fields, "itemType"); ok {
		if v == nil {
			return nil, apperr.BadRequest("itemType 不能清空")
		}
		s, ok := v.(string)
		if !ok {
			return nil, apperr.BadRequest("itemType 必须是字符串")
		}
		if _, exists := assetItemTypes[s]; !exists {
			return nil, apperr.BadRequest("不支持的子资产类型")
		}
		row.ItemType = sql.NullString{String: s, Valid: true}
		changed = append(changed, "itemType")
	}
	if err := assignText(fields, "name", 128, false, &row.Name, &changed); err != nil {
		return nil, err
	}
	if err := assignText(fields, "imageUrl", 1024, true, &row.ImageUrl, &changed); err != nil {
		return nil, err
	}
	if err := assignText(fields, "thumbnailUrl", 1024, true, &row.ThumbnailUrl, &changed); err != nil {
		return nil, err
	}
	if err := assignText(fields, "aiPrompt", 65535, false, &row.AiPrompt, &changed); err != nil {
		return nil, err
	}
	if v, ok := present(fields, "properties"); ok {
		raw, err := readJSONValue(v, "properties", false)
		if err != nil {
			return nil, err
		}
		row.Properties = raw
		changed = append(changed, "properties")
	}
	if v, ok := present(fields, "sortOrder"); ok {
		n, err := readStatus(v, "sortOrder", 1000000)
		if err != nil {
			return nil, err
		}
		row.SortOrder = n
		changed = append(changed, "sortOrder")
	}
	if v, ok := present(fields, "sourceType"); ok {
		n, err := readRange(v, "sourceType", 1, 2)
		if err != nil || !n.Valid {
			return nil, apperr.BadRequest("sourceType 整数范围无效")
		}
		row.SourceType = n
		changed = append(changed, "sourceType")
	}
	if !creating && len(changed) == 0 {
		return nil, apperr.BadRequest("没有可更新的字段")
	}
	return changed, nil
}

// ApplyImageTaskFields 校验并写入生图任务字段。资源命令和生成提交共用这套规则。
func ApplyImageTaskFields(row *sqlcdb.AfvImageTask, fields map[string]any, creating bool) ([]string, error) {
	return applyImageTaskFields(row, fields, creating)
}

func applyImageTaskFields(row *sqlcdb.AfvImageTask, fields map[string]any, creating bool) ([]string, error) {
	allowed := []string{
		"prompt", "promptTemplateId", "refImageUrls", "ratio", "resolution", "aspectRatio",
		"width", "height", "count", "modelId", "workflowVersionId", "category", "status", "errorMsg",
	}
	if err := rejectUnknown(fields, allowed); err != nil {
		return nil, err
	}
	var changed []string
	if err := assignText(fields, "prompt", 65535, false, &row.Prompt, &changed); err != nil {
		return nil, err
	}
	if creating && (!row.Prompt.Valid || strings.TrimSpace(row.Prompt.String) == "") {
		return nil, apperr.BadRequest("提示词不能为空")
	}
	if v, ok := present(fields, "promptTemplateId"); ok {
		id, err := readOptionalID(v, "promptTemplateId")
		if err != nil {
			return nil, err
		}
		row.PromptTemplateID = id
		changed = append(changed, "promptTemplateId")
	}
	if v, ok := present(fields, "refImageUrls"); ok {
		raw, err := readURLList(v, "refImageUrls")
		if err != nil {
			return nil, err
		}
		row.RefImageUrls = raw
		changed = append(changed, "refImageUrls")
	}
	if err := assignText(fields, "ratio", 32, false, &row.Ratio, &changed); err != nil {
		return nil, err
	}
	if err := assignText(fields, "resolution", 16, false, &row.Resolution, &changed); err != nil {
		return nil, err
	}
	if err := assignText(fields, "aspectRatio", 16, false, &row.AspectRatio, &changed); err != nil {
		return nil, err
	}
	if err := assignText(fields, "category", 64, false, &row.Category, &changed); err != nil {
		return nil, err
	}
	if err := assignText(fields, "errorMsg", 65535, false, &row.ErrorMsg, &changed); err != nil {
		return nil, err
	}
	if v, ok := present(fields, "width"); ok {
		n, err := readRange(v, "width", 1, 8192)
		if err != nil {
			return nil, err
		}
		row.Width = n
		changed = append(changed, "width")
	}
	if v, ok := present(fields, "height"); ok {
		n, err := readRange(v, "height", 1, 8192)
		if err != nil {
			return nil, err
		}
		row.Height = n
		changed = append(changed, "height")
	}
	if v, ok := present(fields, "count"); ok {
		n, err := readRange(v, "count", 1, 16)
		if err != nil || !n.Valid {
			return nil, apperr.BadRequest("count 整数范围无效")
		}
		row.Count = n
		changed = append(changed, "count")
	} else if creating {
		row.Count = sql.NullInt32{Int32: 1, Valid: true}
	}
	if v, ok := present(fields, "modelId"); ok {
		id, err := readOptionalID(v, "modelId")
		if err != nil {
			return nil, err
		}
		row.ModelID = id
		changed = append(changed, "modelId")
	}
	if v, ok := present(fields, "workflowVersionId"); ok {
		id, err := readOptionalID(v, "workflowVersionId")
		if err != nil {
			return nil, err
		}
		row.WorkflowVersionID = id
		changed = append(changed, "workflowVersionId")
	}
	if v, ok := present(fields, "status"); ok {
		if creating {
			return nil, apperr.BadRequest("新建任务状态固定为排队中")
		}
		n, err := readStatus(v, "status", 3)
		if err != nil {
			return nil, err
		}
		row.Status = n
		changed = append(changed, "status")
	} else if creating {
		row.Status = sql.NullInt32{Int32: 0, Valid: true}
	}
	if !creating && len(changed) == 0 {
		return nil, apperr.BadRequest("没有可更新的字段")
	}
	return changed, nil
}

// ApplyVideoTaskFields 校验并写入生视频任务字段。资源命令和生成提交共用这套规则。
func ApplyVideoTaskFields(row *sqlcdb.AfvVideoTask, fields map[string]any, creating bool) ([]string, error) {
	return applyVideoTaskFields(row, fields, creating)
}

func applyVideoTaskFields(row *sqlcdb.AfvVideoTask, fields map[string]any, creating bool) ([]string, error) {
	allowed := []string{
		"prompt", "promptTemplateId", "generateMode", "firstFrameImageUrl", "lastFrameImageUrl",
		"referenceImageUrls", "referenceVideoUrls", "referenceAudioUrls", "ratio", "resolution",
		"duration", "watermark", "generateAudio", "seed", "cameraFixed", "count", "modelId",
		"workflowVersionId", "category", "status", "errorMsg",
	}
	if err := rejectUnknown(fields, allowed); err != nil {
		return nil, err
	}
	var changed []string
	if err := assignText(fields, "prompt", 65535, false, &row.Prompt, &changed); err != nil {
		return nil, err
	}
	if err := assignText(fields, "firstFrameImageUrl", 1024, true, &row.FirstFrameImageUrl, &changed); err != nil {
		return nil, err
	}
	if err := assignText(fields, "lastFrameImageUrl", 1024, true, &row.LastFrameImageUrl, &changed); err != nil {
		return nil, err
	}
	if creating && !row.Prompt.Valid && !row.FirstFrameImageUrl.Valid {
		return nil, apperr.BadRequest("生视频任务需要提示词或首帧图片")
	}
	if v, ok := present(fields, "generateMode"); ok {
		if v == nil {
			row.GenerateMode = sql.NullString{}
		} else {
			s, ok := v.(string)
			if !ok {
				return nil, apperr.BadRequest("generateMode 必须是字符串")
			}
			if _, exists := videoModes[s]; !exists {
				return nil, apperr.BadRequest("generateMode 仅支持 text2video 或 image2video")
			}
			row.GenerateMode = sql.NullString{String: s, Valid: true}
		}
		changed = append(changed, "generateMode")
	}
	for _, key := range []string{"referenceImageUrls", "referenceVideoUrls", "referenceAudioUrls"} {
		v, ok := present(fields, key)
		if !ok {
			continue
		}
		raw, err := readURLList(v, key)
		if err != nil {
			return nil, err
		}
		switch key {
		case "referenceImageUrls":
			row.ReferenceImageUrls = raw
		case "referenceVideoUrls":
			row.ReferenceVideoUrls = raw
		default:
			row.ReferenceAudioUrls = raw
		}
		changed = append(changed, key)
	}
	if v, ok := present(fields, "promptTemplateId"); ok {
		id, err := readOptionalID(v, "promptTemplateId")
		if err != nil {
			return nil, err
		}
		row.PromptTemplateID = id
		changed = append(changed, "promptTemplateId")
	}
	if err := assignText(fields, "ratio", 32, false, &row.Ratio, &changed); err != nil {
		return nil, err
	}
	if err := assignText(fields, "resolution", 16, false, &row.Resolution, &changed); err != nil {
		return nil, err
	}
	if err := assignText(fields, "category", 64, false, &row.Category, &changed); err != nil {
		return nil, err
	}
	if err := assignText(fields, "errorMsg", 65535, false, &row.ErrorMsg, &changed); err != nil {
		return nil, err
	}
	if v, ok := present(fields, "duration"); ok {
		n, err := readRange(v, "duration", 1, 600)
		if err != nil {
			return nil, err
		}
		row.Duration = n
		changed = append(changed, "duration")
	}
	if v, ok := present(fields, "watermark"); ok {
		flag, err := readBoolTiny(v, "watermark")
		if err != nil {
			return nil, err
		}
		row.Watermark = flag
		changed = append(changed, "watermark")
	}
	if v, ok := present(fields, "generateAudio"); ok {
		flag, err := readBoolTiny(v, "generateAudio")
		if err != nil {
			return nil, err
		}
		row.GenerateAudio = flag
		changed = append(changed, "generateAudio")
	}
	if v, ok := present(fields, "cameraFixed"); ok {
		flag, err := readBoolTiny(v, "cameraFixed")
		if err != nil {
			return nil, err
		}
		row.CameraFixed = flag
		changed = append(changed, "cameraFixed")
	}
	if v, ok := present(fields, "seed"); ok {
		if v == nil {
			row.Seed = sql.NullInt64{}
		} else {
			n, err := asInt64Value(v)
			if err != nil {
				return nil, apperr.BadRequest("seed 必须是整数")
			}
			row.Seed = sql.NullInt64{Int64: n, Valid: true}
		}
		changed = append(changed, "seed")
	}
	if v, ok := present(fields, "count"); ok {
		n, err := readRange(v, "count", 1, 16)
		if err != nil || !n.Valid {
			return nil, apperr.BadRequest("count 整数范围无效")
		}
		row.Count = n
		changed = append(changed, "count")
	} else if creating {
		row.Count = sql.NullInt32{Int32: 1, Valid: true}
	}
	if v, ok := present(fields, "modelId"); ok {
		id, err := readOptionalID(v, "modelId")
		if err != nil {
			return nil, err
		}
		row.ModelID = id
		changed = append(changed, "modelId")
	}
	if v, ok := present(fields, "workflowVersionId"); ok {
		id, err := readOptionalID(v, "workflowVersionId")
		if err != nil {
			return nil, err
		}
		row.WorkflowVersionID = id
		changed = append(changed, "workflowVersionId")
	}
	if v, ok := present(fields, "status"); ok {
		if creating {
			return nil, apperr.BadRequest("新建任务状态固定为排队中")
		}
		n, err := readStatus(v, "status", 3)
		if err != nil {
			return nil, err
		}
		row.Status = n
		changed = append(changed, "status")
	} else if creating {
		row.Status = sql.NullInt32{Int32: 0, Valid: true}
	}
	if !creating && len(changed) == 0 {
		return nil, apperr.BadRequest("没有可更新的字段")
	}
	return changed, nil
}
