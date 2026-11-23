package generation

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"unicode/utf8"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/apperr"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/resource/handlers"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/sqlcdb"
)

// ClientScope 对齐 Java GenerationClientScope。当前会话没有团队，teamId 必须为空。
type ClientScope struct {
	UserID int64  `json:"userId"`
	TeamID *int64 `json:"teamId"`
}

// ImageSubmitRequest 对齐 Java ImageTaskSubmitReqVO 的字段。
type ImageSubmitRequest struct {
	Scope             *ClientScope     `json:"scope"`
	RequestID         string           `json:"requestId"`
	WorkflowVersionID *int64           `json:"workflowVersionId"`
	ProjectID         *int64           `json:"projectId"`
	Prompt            *string          `json:"prompt"`
	PromptTemplateID  *int64           `json:"promptTemplateId"`
	RefImageURLs      *json.RawMessage `json:"refImageUrls"`
	Ratio             *string          `json:"ratio"`
	Resolution        *string          `json:"resolution"`
	AspectRatio       *string          `json:"aspectRatio"`
	Width             *int32           `json:"width"`
	Height            *int32           `json:"height"`
	Count             *int32           `json:"count"`
	ModelID           *int64           `json:"modelId"`
	Category          *string          `json:"category"`
}

// VideoSubmitRequest 对齐 Java VideoTaskSubmitReqVO 的字段。
type VideoSubmitRequest struct {
	Scope              *ClientScope     `json:"scope"`
	RequestID          string           `json:"requestId"`
	WorkflowVersionID  *int64           `json:"workflowVersionId"`
	ProjectID          *int64           `json:"projectId"`
	Prompt             *string          `json:"prompt"`
	PromptTemplateID   *int64           `json:"promptTemplateId"`
	GenerateMode       *string          `json:"generateMode"`
	FirstFrameImageURL *string          `json:"firstFrameImageUrl"`
	LastFrameImageURL  *string          `json:"lastFrameImageUrl"`
	ReferenceImageURLs *json.RawMessage `json:"referenceImageUrls"`
	ReferenceVideoURLs *json.RawMessage `json:"referenceVideoUrls"`
	ReferenceAudioURLs *json.RawMessage `json:"referenceAudioUrls"`
	Ratio              *string          `json:"ratio"`
	Resolution         *string          `json:"resolution"`
	Duration           *int32           `json:"duration"`
	Watermark          *bool            `json:"watermark"`
	GenerateAudio      *bool            `json:"generateAudio"`
	Seed               *int64           `json:"seed"`
	CameraFixed        *bool            `json:"cameraFixed"`
	Count              *int32           `json:"count"`
	ModelID            *int64           `json:"modelId"`
	Category           *string          `json:"category"`
}

// EnqueueImage 创建生图任务、排队中的条目和受理记录。返回公开任务标识。
// 相同 requestId 与相同内容再次提交时返回原来的公开标识。
func (s *Service) EnqueueImage(ctx context.Context, actorUserID int64, req ImageSubmitRequest) (string, error) {
	requestID, err := prepareSubmit(actorUserID, req.Scope, req.RequestID)
	if err != nil {
		return "", err
	}
	row, err := normalizeImage(actorUserID, req)
	if err != nil {
		return "", err
	}
	if err := rejectWorkflow(row.WorkflowVersionID.Valid); err != nil {
		return "", err
	}
	hash, err := imageRequestHash(actorUserID, row)
	if err != nil {
		return "", err
	}
	return s.accept(ctx, acceptSpec{
		ActorUserID: actorUserID,
		Kind:        KindImage,
		RequestID:   requestID,
		Hash:        hash,
		ProjectID:   row.ProjectID,
		ModelID:     modelOrZero(row.ModelID),
		Build: func(ctx context.Context, q *sqlcdb.Queries) (createdTask, error) {
			return insertImageTask(ctx, q, row)
		},
	})
}

// EnqueueVideo 创建生视频任务、排队中的条目和受理记录。返回公开任务标识。
func (s *Service) EnqueueVideo(ctx context.Context, actorUserID int64, req VideoSubmitRequest) (string, error) {
	requestID, err := prepareSubmit(actorUserID, req.Scope, req.RequestID)
	if err != nil {
		return "", err
	}
	row, err := normalizeVideo(actorUserID, req)
	if err != nil {
		return "", err
	}
	if err := rejectWorkflow(row.WorkflowVersionID.Valid); err != nil {
		return "", err
	}
	hash, err := videoRequestHash(actorUserID, row)
	if err != nil {
		return "", err
	}
	return s.accept(ctx, acceptSpec{
		ActorUserID: actorUserID,
		Kind:        KindVideo,
		RequestID:   requestID,
		Hash:        hash,
		ProjectID:   row.ProjectID,
		ModelID:     modelOrZero(row.ModelID),
		Build: func(ctx context.Context, q *sqlcdb.Queries) (createdTask, error) {
			return insertVideoTask(ctx, q, row)
		},
	})
}

func prepareSubmit(actorUserID int64, scope *ClientScope, requestID string) (string, error) {
	if err := checkScope(actorUserID, scope); err != nil {
		return "", err
	}
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return "", apperr.BadRequest("生成请求标识不能为空")
	}
	if utf8.RuneCountInString(requestID) > 128 {
		return "", apperr.BadRequest("生成请求标识最长 128 个字符")
	}
	return requestID, nil
}

func checkScope(actorUserID int64, scope *ClientScope) error {
	if scope == nil || scope.UserID == 0 {
		return apperr.BadRequest("请先确认当前生成账号与团队")
	}
	if scope.UserID != actorUserID || scope.TeamID != nil {
		return apperr.Conflict("当前账号或团队已变化，请回到原提交范围确认任务")
	}
	return nil
}

func rejectWorkflow(set bool) error {
	if !set {
		return nil
	}
	return apperr.BadRequest("本服务尚未接入 ComfyUI，不能提交带工作流版本的生成任务")
}

func modelOrZero(id sql.NullInt64) int64 {
	if !id.Valid {
		return 0
	}
	return id.Int64
}

func normalizeImage(actorUserID int64, req ImageSubmitRequest) (sqlcdb.AfvImageTask, error) {
	fields, err := imageFields(req)
	if err != nil {
		return sqlcdb.AfvImageTask{}, err
	}
	row := sqlcdb.AfvImageTask{
		UserID:    actorUserID,
		OwnerType: sql.NullInt32{Int32: 1, Valid: true},
		OwnerID:   sql.NullInt64{Int64: actorUserID, Valid: true},
	}
	if req.ProjectID != nil {
		if *req.ProjectID <= 0 {
			return sqlcdb.AfvImageTask{}, apperr.BadRequest("projectId 必须是正整数 ID")
		}
		row.ProjectID = sql.NullInt64{Int64: *req.ProjectID, Valid: true}
	}
	if _, err := handlers.ApplyImageTaskFields(&row, fields, true); err != nil {
		return sqlcdb.AfvImageTask{}, err
	}
	return row, nil
}

func normalizeVideo(actorUserID int64, req VideoSubmitRequest) (sqlcdb.AfvVideoTask, error) {
	fields, err := videoFields(req)
	if err != nil {
		return sqlcdb.AfvVideoTask{}, err
	}
	row := sqlcdb.AfvVideoTask{
		UserID:    actorUserID,
		OwnerType: sql.NullInt32{Int32: 1, Valid: true},
		OwnerID:   sql.NullInt64{Int64: actorUserID, Valid: true},
	}
	if req.ProjectID != nil {
		if *req.ProjectID <= 0 {
			return sqlcdb.AfvVideoTask{}, apperr.BadRequest("projectId 必须是正整数 ID")
		}
		row.ProjectID = sql.NullInt64{Int64: *req.ProjectID, Valid: true}
	}
	if _, err := handlers.ApplyVideoTaskFields(&row, fields, true); err != nil {
		return sqlcdb.AfvVideoTask{}, err
	}
	return row, nil
}

func imageFields(req ImageSubmitRequest) (map[string]any, error) {
	fields := map[string]any{}
	putString(fields, "prompt", req.Prompt)
	putID(fields, "promptTemplateId", req.PromptTemplateID)
	putID(fields, "modelId", req.ModelID)
	putID(fields, "workflowVersionId", req.WorkflowVersionID)
	putString(fields, "ratio", req.Ratio)
	putString(fields, "resolution", req.Resolution)
	putString(fields, "aspectRatio", req.AspectRatio)
	putString(fields, "category", req.Category)
	putInt(fields, "width", req.Width)
	putInt(fields, "height", req.Height)
	putInt(fields, "count", req.Count)
	if err := putURLList(fields, "refImageUrls", req.RefImageURLs); err != nil {
		return nil, err
	}
	return fields, nil
}

func videoFields(req VideoSubmitRequest) (map[string]any, error) {
	fields := map[string]any{}
	putString(fields, "prompt", req.Prompt)
	putID(fields, "promptTemplateId", req.PromptTemplateID)
	putID(fields, "modelId", req.ModelID)
	putID(fields, "workflowVersionId", req.WorkflowVersionID)
	putString(fields, "generateMode", req.GenerateMode)
	putString(fields, "firstFrameImageUrl", req.FirstFrameImageURL)
	putString(fields, "lastFrameImageUrl", req.LastFrameImageURL)
	putString(fields, "ratio", req.Ratio)
	putString(fields, "resolution", req.Resolution)
	putString(fields, "category", req.Category)
	putInt(fields, "duration", req.Duration)
	putInt(fields, "count", req.Count)
	putID(fields, "seed", req.Seed)
	putBool(fields, "watermark", req.Watermark)
	putBool(fields, "generateAudio", req.GenerateAudio)
	putBool(fields, "cameraFixed", req.CameraFixed)
	if err := putURLList(fields, "referenceImageUrls", req.ReferenceImageURLs); err != nil {
		return nil, err
	}
	if err := putURLList(fields, "referenceVideoUrls", req.ReferenceVideoURLs); err != nil {
		return nil, err
	}
	if err := putURLList(fields, "referenceAudioUrls", req.ReferenceAudioURLs); err != nil {
		return nil, err
	}
	return fields, nil
}

func putString(fields map[string]any, key string, value *string) {
	if value != nil {
		fields[key] = *value
	}
}

func putID(fields map[string]any, key string, value *int64) {
	if value != nil {
		fields[key] = *value
	}
}

func putInt(fields map[string]any, key string, value *int32) {
	if value != nil {
		fields[key] = int64(*value)
	}
}

func putBool(fields map[string]any, key string, value *bool) {
	if value != nil {
		fields[key] = *value
	}
}

func putURLList(fields map[string]any, key string, raw *json.RawMessage) error {
	value, ok, err := urlListValue(raw, key)
	if err != nil || !ok {
		return err
	}
	fields[key] = value
	return nil
}
