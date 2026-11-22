package handlers

import (
	"context"
	"database/sql"
	"strings"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/apperr"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/resource"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/sqlcdb"
)

func newPublicTaskID() (string, error) {
	id, err := resource.NewCommandID()
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(id, "-", ""), nil
}

func optionalProject(ctx context.Context, q *sqlcdb.Queries, projectID sql.NullInt64, actorUserID int64) (*int64, error) {
	if !projectID.Valid {
		return nil, nil
	}
	return requireOwnedProject(ctx, q, projectID.Int64, actorUserID)
}

type ImageTaskCreate struct{}

func (ImageTaskCreate) ResourceType() resource.Type   { return resource.TypeImageTask }
func (ImageTaskCreate) Operation() resource.Operation { return resource.OpCreate }

func (ImageTaskCreate) Authorize(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64, _ *resource.Receipt) (*int64, error) {
	create, err := asCreate(cmd)
	if err != nil {
		return nil, err
	}
	if create.ProjectID == nil {
		return nil, nil
	}
	projectID, err := requiredProject(create.ProjectID)
	if err != nil {
		return nil, err
	}
	return requireOwnedProject(ctx, q, projectID, actorUserID)
}

func (ImageTaskCreate) Apply(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64) (resource.Mutation, error) {
	create, err := asCreate(cmd)
	if err != nil {
		return resource.Mutation{}, err
	}
	row := sqlcdb.AfvImageTask{
		UserID:    actorUserID,
		OwnerType: sql.NullInt32{Int32: 1, Valid: true},
		OwnerID:   sql.NullInt64{Int64: actorUserID, Valid: true},
	}
	if create.ProjectID != nil {
		projectID, err := requiredProject(create.ProjectID)
		if err != nil {
			return resource.Mutation{}, err
		}
		row.ProjectID = sql.NullInt64{Int64: projectID, Valid: true}
	}
	if _, err := applyImageTaskFields(&row, create.Fields, true); err != nil {
		return resource.Mutation{}, err
	}
	publicID, err := newPublicTaskID()
	if err != nil {
		return resource.Mutation{}, err
	}
	id, err := q.CreateImageTask(ctx, sqlcdb.CreateImageTaskParams{
		TaskID:            publicID,
		UserID:            row.UserID,
		ProjectID:         row.ProjectID,
		Prompt:            row.Prompt,
		PromptTemplateID:  row.PromptTemplateID,
		RefImageUrls:      row.RefImageUrls,
		Ratio:             row.Ratio,
		Resolution:        row.Resolution,
		AspectRatio:       row.AspectRatio,
		Width:             row.Width,
		Height:            row.Height,
		Count:             row.Count,
		Status:            row.Status,
		ErrorMsg:          row.ErrorMsg,
		ModelID:           row.ModelID,
		WorkflowVersionID: row.WorkflowVersionID,
		Category:          row.Category,
		OwnerID:           row.OwnerID,
	})
	if err != nil {
		return resource.Mutation{}, duplicate(err, "任务标识冲突")
	}
	saved, err := q.GetImageTaskByID(ctx, id)
	if err != nil {
		return resource.Mutation{}, err
	}
	var projectID int64
	if saved.ProjectID.Valid {
		projectID = saved.ProjectID.Int64
	}
	return created(resource.TypeImageTask, id, projectID, resource.ImageTaskVO(saved, 0))
}

type ImageTaskPatch struct{}

func (ImageTaskPatch) ResourceType() resource.Type   { return resource.TypeImageTask }
func (ImageTaskPatch) Operation() resource.Operation { return resource.OpPatch }

func (ImageTaskPatch) Authorize(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64, _ *resource.Receipt) (*int64, error) {
	patch, err := asPatch(cmd)
	if err != nil {
		return nil, err
	}
	row, err := q.GetImageTaskByID(ctx, patch.ID)
	if err != nil {
		return nil, missing(err, "生图任务不存在")
	}
	if row.UserID != actorUserID {
		return nil, apperr.Forbidden("无权访问该任务")
	}
	return optionalProject(ctx, q, row.ProjectID, actorUserID)
}

func (ImageTaskPatch) Apply(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, _ int64) (resource.Mutation, error) {
	patch, err := asPatch(cmd)
	if err != nil {
		return resource.Mutation{}, err
	}
	row, err := q.LockImageTaskByID(ctx, patch.ID)
	if err != nil {
		return resource.Mutation{}, missing(err, "生图任务不存在")
	}
	version, err := taskVersion(ctx, q, resource.TypeImageTask, row.ID)
	if err != nil {
		return resource.Mutation{}, err
	}
	if version != patch.Version {
		return resource.Mutation{}, apperr.Conflict("生图任务版本已变化，请刷新后重试")
	}
	before := resource.ImageTaskVO(row, version)
	changed, err := applyImageTaskFields(&row, patch.Patch, false)
	if err != nil {
		return resource.Mutation{}, err
	}
	n, err := q.UpdateImageTask(ctx, sqlcdb.UpdateImageTaskParams{
		Prompt:            row.Prompt,
		PromptTemplateID:  row.PromptTemplateID,
		RefImageUrls:      row.RefImageUrls,
		Ratio:             row.Ratio,
		Resolution:        row.Resolution,
		AspectRatio:       row.AspectRatio,
		Width:             row.Width,
		Height:            row.Height,
		Count:             row.Count,
		Status:            row.Status,
		ErrorMsg:          row.ErrorMsg,
		ModelID:           row.ModelID,
		WorkflowVersionID: row.WorkflowVersionID,
		Category:          row.Category,
		ID:                row.ID,
	})
	if err != nil {
		return resource.Mutation{}, err
	}
	if err := oneRow(n, "生图任务版本已变化，请刷新后重试"); err != nil {
		return resource.Mutation{}, err
	}
	saved, err := q.GetImageTaskByID(ctx, row.ID)
	if err != nil {
		return resource.Mutation{}, err
	}
	var projectID int64
	if saved.ProjectID.Valid {
		projectID = saved.ProjectID.Int64
	}
	return patched(resource.TypeImageTask, saved.ID, projectID, version+1, changed, before, resource.ImageTaskVO(saved, version+1))
}

type VideoTaskCreate struct{}

func (VideoTaskCreate) ResourceType() resource.Type   { return resource.TypeVideoTask }
func (VideoTaskCreate) Operation() resource.Operation { return resource.OpCreate }

func (VideoTaskCreate) Authorize(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64, _ *resource.Receipt) (*int64, error) {
	create, err := asCreate(cmd)
	if err != nil {
		return nil, err
	}
	if create.ProjectID == nil {
		return nil, nil
	}
	projectID, err := requiredProject(create.ProjectID)
	if err != nil {
		return nil, err
	}
	return requireOwnedProject(ctx, q, projectID, actorUserID)
}

func (VideoTaskCreate) Apply(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64) (resource.Mutation, error) {
	create, err := asCreate(cmd)
	if err != nil {
		return resource.Mutation{}, err
	}
	row := sqlcdb.AfvVideoTask{
		UserID:    actorUserID,
		OwnerType: sql.NullInt32{Int32: 1, Valid: true},
		OwnerID:   sql.NullInt64{Int64: actorUserID, Valid: true},
	}
	if create.ProjectID != nil {
		projectID, err := requiredProject(create.ProjectID)
		if err != nil {
			return resource.Mutation{}, err
		}
		row.ProjectID = sql.NullInt64{Int64: projectID, Valid: true}
	}
	if _, err := applyVideoTaskFields(&row, create.Fields, true); err != nil {
		return resource.Mutation{}, err
	}
	publicID, err := newPublicTaskID()
	if err != nil {
		return resource.Mutation{}, err
	}
	id, err := q.CreateVideoTask(ctx, sqlcdb.CreateVideoTaskParams{
		TaskID:             publicID,
		UserID:             row.UserID,
		ProjectID:          row.ProjectID,
		Prompt:             row.Prompt,
		PromptTemplateID:   row.PromptTemplateID,
		GenerateMode:       row.GenerateMode,
		FirstFrameImageUrl: row.FirstFrameImageUrl,
		LastFrameImageUrl:  row.LastFrameImageUrl,
		ReferenceImageUrls: row.ReferenceImageUrls,
		ReferenceVideoUrls: row.ReferenceVideoUrls,
		ReferenceAudioUrls: row.ReferenceAudioUrls,
		Ratio:              row.Ratio,
		Resolution:         row.Resolution,
		Duration:           row.Duration,
		Watermark:          row.Watermark,
		GenerateAudio:      row.GenerateAudio,
		Seed:               row.Seed,
		CameraFixed:        row.CameraFixed,
		Count:              row.Count,
		Status:             row.Status,
		ErrorMsg:           row.ErrorMsg,
		ModelID:            row.ModelID,
		WorkflowVersionID:  row.WorkflowVersionID,
		Category:           row.Category,
		OwnerID:            row.OwnerID,
	})
	if err != nil {
		return resource.Mutation{}, duplicate(err, "任务标识冲突")
	}
	saved, err := q.GetVideoTaskByID(ctx, id)
	if err != nil {
		return resource.Mutation{}, err
	}
	var projectID int64
	if saved.ProjectID.Valid {
		projectID = saved.ProjectID.Int64
	}
	return created(resource.TypeVideoTask, id, projectID, resource.VideoTaskVO(saved, 0))
}

type VideoTaskPatch struct{}

func (VideoTaskPatch) ResourceType() resource.Type   { return resource.TypeVideoTask }
func (VideoTaskPatch) Operation() resource.Operation { return resource.OpPatch }

func (VideoTaskPatch) Authorize(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64, _ *resource.Receipt) (*int64, error) {
	patch, err := asPatch(cmd)
	if err != nil {
		return nil, err
	}
	row, err := q.GetVideoTaskByID(ctx, patch.ID)
	if err != nil {
		return nil, missing(err, "生视频任务不存在")
	}
	if row.UserID != actorUserID {
		return nil, apperr.Forbidden("无权访问该任务")
	}
	return optionalProject(ctx, q, row.ProjectID, actorUserID)
}

func (VideoTaskPatch) Apply(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, _ int64) (resource.Mutation, error) {
	patch, err := asPatch(cmd)
	if err != nil {
		return resource.Mutation{}, err
	}
	row, err := q.LockVideoTaskByID(ctx, patch.ID)
	if err != nil {
		return resource.Mutation{}, missing(err, "生视频任务不存在")
	}
	version, err := taskVersion(ctx, q, resource.TypeVideoTask, row.ID)
	if err != nil {
		return resource.Mutation{}, err
	}
	if version != patch.Version {
		return resource.Mutation{}, apperr.Conflict("生视频任务版本已变化，请刷新后重试")
	}
	before := resource.VideoTaskVO(row, version)
	changed, err := applyVideoTaskFields(&row, patch.Patch, false)
	if err != nil {
		return resource.Mutation{}, err
	}
	n, err := q.UpdateVideoTask(ctx, sqlcdb.UpdateVideoTaskParams{
		Prompt:             row.Prompt,
		PromptTemplateID:   row.PromptTemplateID,
		GenerateMode:       row.GenerateMode,
		FirstFrameImageUrl: row.FirstFrameImageUrl,
		LastFrameImageUrl:  row.LastFrameImageUrl,
		ReferenceImageUrls: row.ReferenceImageUrls,
		ReferenceVideoUrls: row.ReferenceVideoUrls,
		ReferenceAudioUrls: row.ReferenceAudioUrls,
		Ratio:              row.Ratio,
		Resolution:         row.Resolution,
		Duration:           row.Duration,
		Watermark:          row.Watermark,
		GenerateAudio:      row.GenerateAudio,
		Seed:               row.Seed,
		CameraFixed:        row.CameraFixed,
		Count:              row.Count,
		Status:             row.Status,
		ErrorMsg:           row.ErrorMsg,
		ModelID:            row.ModelID,
		WorkflowVersionID:  row.WorkflowVersionID,
		Category:           row.Category,
		ID:                 row.ID,
	})
	if err != nil {
		return resource.Mutation{}, err
	}
	if err := oneRow(n, "生视频任务版本已变化，请刷新后重试"); err != nil {
		return resource.Mutation{}, err
	}
	saved, err := q.GetVideoTaskByID(ctx, row.ID)
	if err != nil {
		return resource.Mutation{}, err
	}
	var projectID int64
	if saved.ProjectID.Valid {
		projectID = saved.ProjectID.Int64
	}
	return patched(resource.TypeVideoTask, saved.ID, projectID, version+1, changed, before, resource.VideoTaskVO(saved, version+1))
}
