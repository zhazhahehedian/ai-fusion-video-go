package generation

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/sqlcdb"
)

func imageRequestHash(actorUserID int64, row sqlcdb.AfvImageTask) (string, error) {
	body := struct {
		Kind             string          `json:"kind"`
		Actor            int64           `json:"actor"`
		ProjectID        *int64          `json:"projectId"`
		Prompt           string          `json:"prompt"`
		PromptTemplateID *int64          `json:"promptTemplateId"`
		RefImageURLs     json.RawMessage `json:"refImageUrls"`
		Ratio            *string         `json:"ratio"`
		Resolution       *string         `json:"resolution"`
		AspectRatio      *string         `json:"aspectRatio"`
		Width            *int32          `json:"width"`
		Height           *int32          `json:"height"`
		Count            int32           `json:"count"`
		ModelID          *int64          `json:"modelId"`
		Category         *string         `json:"category"`
	}{
		Kind:             KindImage,
		Actor:            actorUserID,
		ProjectID:        nullInt64Ptr(row.ProjectID),
		Prompt:           nullString(row.Prompt),
		PromptTemplateID: nullInt64Ptr(row.PromptTemplateID),
		RefImageURLs:     nullJSON(row.RefImageUrls),
		Ratio:            nullStringPtr(row.Ratio),
		Resolution:       nullStringPtr(row.Resolution),
		AspectRatio:      nullStringPtr(row.AspectRatio),
		Width:            nullInt32Ptr(row.Width),
		Height:           nullInt32Ptr(row.Height),
		Count:            nullInt32(row.Count),
		ModelID:          nullInt64Ptr(row.ModelID),
		Category:         nullStringPtr(row.Category),
	}
	return hashCanon(body)
}

func videoRequestHash(actorUserID int64, row sqlcdb.AfvVideoTask) (string, error) {
	body := struct {
		Kind               string          `json:"kind"`
		Actor              int64           `json:"actor"`
		ProjectID          *int64          `json:"projectId"`
		Prompt             *string         `json:"prompt"`
		PromptTemplateID   *int64          `json:"promptTemplateId"`
		GenerateMode       *string         `json:"generateMode"`
		FirstFrameImageURL *string         `json:"firstFrameImageUrl"`
		LastFrameImageURL  *string         `json:"lastFrameImageUrl"`
		ReferenceImageURLs json.RawMessage `json:"referenceImageUrls"`
		ReferenceVideoURLs json.RawMessage `json:"referenceVideoUrls"`
		ReferenceAudioURLs json.RawMessage `json:"referenceAudioUrls"`
		Ratio              *string         `json:"ratio"`
		Resolution         *string         `json:"resolution"`
		Duration           *int32          `json:"duration"`
		Watermark          *bool           `json:"watermark"`
		GenerateAudio      *bool           `json:"generateAudio"`
		Seed               *int64          `json:"seed"`
		CameraFixed        *bool           `json:"cameraFixed"`
		Count              int32           `json:"count"`
		ModelID            *int64          `json:"modelId"`
		Category           *string         `json:"category"`
	}{
		Kind:               KindVideo,
		Actor:              actorUserID,
		ProjectID:          nullInt64Ptr(row.ProjectID),
		Prompt:             nullStringPtr(row.Prompt),
		PromptTemplateID:   nullInt64Ptr(row.PromptTemplateID),
		GenerateMode:       nullStringPtr(row.GenerateMode),
		FirstFrameImageURL: nullStringPtr(row.FirstFrameImageUrl),
		LastFrameImageURL:  nullStringPtr(row.LastFrameImageUrl),
		ReferenceImageURLs: nullJSON(row.ReferenceImageUrls),
		ReferenceVideoURLs: nullJSON(row.ReferenceVideoUrls),
		ReferenceAudioURLs: nullJSON(row.ReferenceAudioUrls),
		Ratio:              nullStringPtr(row.Ratio),
		Resolution:         nullStringPtr(row.Resolution),
		Duration:           nullInt32Ptr(row.Duration),
		Watermark:          nullBoolPtr(row.Watermark),
		GenerateAudio:      nullBoolPtr(row.GenerateAudio),
		Seed:               nullInt64Ptr(row.Seed),
		CameraFixed:        nullBoolPtr(row.CameraFixed),
		Count:              nullInt32(row.Count),
		ModelID:            nullInt64Ptr(row.ModelID),
		Category:           nullStringPtr(row.Category),
	}
	return hashCanon(body)
}

func hashCanon(body any) (string, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:]), nil
}

func nullJSON(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return json.RawMessage("null")
	}
	return raw
}

func nullString(v sql.NullString) string {
	if !v.Valid {
		return ""
	}
	return v.String
}

func nullStringPtr(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}
	s := v.String
	return &s
}

func nullInt64Ptr(v sql.NullInt64) *int64 {
	if !v.Valid {
		return nil
	}
	n := v.Int64
	return &n
}

func nullInt32Ptr(v sql.NullInt32) *int32 {
	if !v.Valid {
		return nil
	}
	n := v.Int32
	return &n
}

func nullInt32(v sql.NullInt32) int32 {
	if !v.Valid {
		return 0
	}
	return v.Int32
}

func nullBoolPtr(v sql.NullInt16) *bool {
	if !v.Valid {
		return nil
	}
	b := v.Int16 != 0
	return &b
}
