package resource

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/sqlcdb"
)

func ProjectVO(p sqlcdb.GetProjectByIDRow) map[string]any {
	return map[string]any{
		"id":                  p.ID,
		"name":                p.Name,
		"description":         nullString(p.Description),
		"coverUrl":            nullString(p.CoverUrl),
		"scope":               nullInt32(p.Scope),
		"ownerType":           p.OwnerType,
		"ownerId":             p.OwnerID,
		"status":              nullInt32(p.Status),
		"properties":          rawJSON(p.Properties),
		"artStyle":            nullString(p.ArtStyle),
		"artStyleDescription": nullString(p.ArtStyleDescription),
		"artStyleImagePrompt": nullString(p.ArtStyleImagePrompt),
		"artStyleImageUrl":    nullString(p.ArtStyleImageUrl),
		"version":             p.Version,
		"deleted":             p.Deleted != 0,
		"createTime":          formatTime(p.CreateTime),
		"updateTime":          formatTime(p.UpdateTime),
	}
}

func EpisodeVO(e sqlcdb.AfvScriptEpisode) map[string]any {
	return map[string]any{
		"id":               e.ID,
		"scriptId":         e.ScriptID,
		"episodeNumber":    nullInt32(e.EpisodeNumber),
		"title":            nullString(e.Title),
		"synopsis":         nullString(e.Synopsis),
		"rawContent":       nullString(e.RawContent),
		"durationEstimate": nullInt32(e.DurationEstimate),
		"totalScenes":      nullInt32(e.TotalScenes),
		"sourceType":       nullInt32(e.SourceType),
		"sortOrder":        nullInt32(e.SortOrder),
		"parsingStatus":    nullInt32(e.ParsingStatus),
		"status":           nullInt32(e.Status),
		"version":          nullInt32(e.Version),
		"deleted":          e.Deleted != 0,
		"createTime":       formatTime(e.CreateTime),
		"updateTime":       formatTime(e.UpdateTime),
	}
}

func SceneVO(s sqlcdb.AfvScriptSceneItem) map[string]any {
	return map[string]any{
		"id":                s.ID,
		"episodeId":         s.EpisodeID,
		"scriptId":          s.ScriptID,
		"sceneNumber":       nullString(s.SceneNumber),
		"sceneHeading":      nullString(s.SceneHeading),
		"location":          nullString(s.Location),
		"timeOfDay":         nullString(s.TimeOfDay),
		"intExt":            nullString(s.IntExt),
		"characters":        rawJSON(s.Characters),
		"characterAssetIds": rawJSON(s.CharacterAssetIds),
		"sceneAssetId":      nullInt64(s.SceneAssetID),
		"propAssetIds":      rawJSON(s.PropAssetIds),
		"sceneDescription":  nullString(s.SceneDescription),
		"dialogues":         rawJSON(s.Dialogues),
		"sortOrder":         nullInt32(s.SortOrder),
		"status":            nullInt32(s.Status),
		"version":           nullInt32(s.Version),
		"deleted":           s.Deleted != 0,
		"createTime":        formatTime(s.CreateTime),
		"updateTime":        formatTime(s.UpdateTime),
	}
}

func ScriptVO(s sqlcdb.AfvScript) map[string]any {
	return map[string]any{
		"id":               s.ID,
		"projectId":        s.ProjectID,
		"title":            nullString(s.Title),
		"content":          nullString(s.Content),
		"rawContent":       nullString(s.RawContent),
		"totalEpisodes":    nullInt32(s.TotalEpisodes),
		"storySynopsis":    nullString(s.StorySynopsis),
		"charactersJson":   rawJSON(s.CharactersJson),
		"sourceType":       nullInt32(s.SourceType),
		"parsingStatus":    nullInt32(s.ParsingStatus),
		"parsingProgress":  nullString(s.ParsingProgress),
		"summary":          nullString(s.Summary),
		"genre":            nullString(s.Genre),
		"targetAudience":   nullString(s.TargetAudience),
		"durationEstimate": nullInt32(s.DurationEstimate),
		"scope":            nullInt32(s.Scope),
		"ownerType":        nullInt32(s.OwnerType),
		"ownerId":          nullInt64(s.OwnerID),
		"aiGenerated":      nullInt16(s.AiGenerated),
		"version":          nullInt32(s.Version),
		"status":           nullInt32(s.Status),
		"deleted":          s.Deleted != 0,
		"createTime":       formatTime(s.CreateTime),
		"updateTime":       formatTime(s.UpdateTime),
	}
}

func nullString(v sql.NullString) any {
	if !v.Valid {
		return nil
	}
	return v.String
}

func nullInt32(v sql.NullInt32) any {
	if !v.Valid {
		return nil
	}
	return v.Int32
}

func nullInt64(v sql.NullInt64) any {
	if !v.Valid {
		return nil
	}
	return v.Int64
}

func nullInt16(v sql.NullInt16) any {
	if !v.Valid {
		return nil
	}
	return v.Int16
}

func rawJSON(v json.RawMessage) any {
	if len(v) == 0 {
		return nil
	}
	var decoded any
	if err := json.Unmarshal(v, &decoded); err != nil {
		return nil
	}
	return decoded
}

func formatTime(t time.Time) string {
	return t.Format("2006-01-02T15:04:05")
}

func formatNullTime(t sql.NullTime) any {
	if !t.Valid {
		return nil
	}
	return formatTime(t.Time)
}

func nullTiny(v sql.NullInt16) any {
	if !v.Valid {
		return nil
	}
	return int16(v.Int16)
}

func deletedTiny(v sql.NullInt16) bool {
	return v.Valid && v.Int16 != 0
}

func nullDecimal(v sql.NullString) any {
	if !v.Valid {
		return nil
	}
	return json.Number(v.String)
}

func rawTextJSON(v sql.NullString) any {
	if !v.Valid || strings.TrimSpace(v.String) == "" {
		return nil
	}
	return rawJSON(json.RawMessage(v.String))
}
