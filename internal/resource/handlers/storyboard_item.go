package handlers

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/apperr"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/resource"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/sqlcdb"
)

func authorizeStoryboardScene(ctx context.Context, q *sqlcdb.Queries, sceneID, actorUserID int64) (*int64, sqlcdb.AfvStoryboardScene, error) {
	scene, err := q.GetStoryboardSceneByID(ctx, sceneID)
	if err != nil {
		return nil, sqlcdb.AfvStoryboardScene{}, missing(err, "分镜场次不存在")
	}
	board, err := q.GetStoryboardByID(ctx, scene.StoryboardID)
	if err != nil {
		return nil, sqlcdb.AfvStoryboardScene{}, missing(err, "分镜不存在")
	}
	projectID, err := boardProject(board)
	if err != nil {
		return nil, sqlcdb.AfvStoryboardScene{}, err
	}
	pid, err := requireOwnedProject(ctx, q, projectID, actorUserID)
	return pid, scene, err
}

type StoryboardItemCreate struct{}

func (StoryboardItemCreate) ResourceType() resource.Type   { return resource.TypeShot }
func (StoryboardItemCreate) Operation() resource.Operation { return resource.OpCreate }

func (StoryboardItemCreate) Authorize(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64, _ *resource.Receipt) (*int64, error) {
	create, err := asCreate(cmd)
	if err != nil {
		return nil, err
	}
	sceneID, err := requiredPositiveID(create.Fields["storyboardSceneId"], "storyboardSceneId")
	if err != nil {
		return nil, err
	}
	pid, _, err := authorizeStoryboardScene(ctx, q, sceneID, actorUserID)
	return pid, err
}

func (StoryboardItemCreate) Apply(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, _ int64) (resource.Mutation, error) {
	create, err := asCreate(cmd)
	if err != nil {
		return resource.Mutation{}, err
	}
	sceneID, err := requiredPositiveID(create.Fields["storyboardSceneId"], "storyboardSceneId")
	if err != nil {
		return resource.Mutation{}, err
	}
	parentVersion, err := requiredNonNegInt(create.Fields["parentVersion"], "parentVersion")
	if err != nil {
		return resource.Mutation{}, err
	}
	scene, err := q.LockStoryboardSceneByID(ctx, sceneID)
	if err != nil {
		return resource.Mutation{}, missing(err, "分镜场次不存在")
	}
	if scene.Version != parentVersion {
		return resource.Mutation{}, apperr.Conflict("分镜场次版本已变化，请刷新后重试")
	}
	fields := cloneMap(create.Fields)
	delete(fields, "storyboardSceneId")
	delete(fields, "parentVersion")
	count, err := q.CountStoryboardItemsByScene(ctx, sql.NullInt64{Int64: scene.ID, Valid: true})
	if err != nil {
		return resource.Mutation{}, err
	}
	row := sqlcdb.AfvStoryboardItem{
		StoryboardID:        scene.StoryboardID,
		StoryboardEpisodeID: sql.NullInt64{Int64: scene.EpisodeID, Valid: true},
		StoryboardSceneID:   sql.NullInt64{Int64: scene.ID, Valid: true},
		SortOrder:           sql.NullInt32{Int32: int32(count), Valid: true},
		AutoShotNumber:      sql.NullString{String: fmt.Sprintf("%d", count+1), Valid: true},
		Status:              sql.NullInt32{Int32: 0, Valid: true},
	}
	if _, err := applyShotFields(&row, fields, true); err != nil {
		return resource.Mutation{}, err
	}
	id, err := q.CreateStoryboardItem(ctx, shotParams(row))
	if err != nil {
		return resource.Mutation{}, err
	}
	bumped, err := q.BumpStoryboardSceneVersion(ctx, sqlcdb.BumpStoryboardSceneVersionParams{ID: scene.ID, Version: scene.Version})
	if err != nil {
		return resource.Mutation{}, err
	}
	if err := oneRow(bumped, "分镜场次版本已变化，请刷新后重试"); err != nil {
		return resource.Mutation{}, err
	}
	saved, err := q.GetStoryboardItemByID(ctx, id)
	if err != nil {
		return resource.Mutation{}, err
	}
	board, err := q.GetStoryboardByID(ctx, saved.StoryboardID)
	if err != nil {
		return resource.Mutation{}, err
	}
	projectID, err := boardProject(board)
	if err != nil {
		return resource.Mutation{}, err
	}
	return created(resource.TypeShot, id, projectID, resource.StoryboardItemVO(saved))
}

type StoryboardItemPatch struct{}

func (StoryboardItemPatch) ResourceType() resource.Type   { return resource.TypeShot }
func (StoryboardItemPatch) Operation() resource.Operation { return resource.OpPatch }

func (StoryboardItemPatch) Authorize(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64, _ *resource.Receipt) (*int64, error) {
	patch, err := asPatch(cmd)
	if err != nil {
		return nil, err
	}
	item, err := q.GetStoryboardItemByID(ctx, patch.ID)
	if err != nil {
		return nil, missing(err, "镜头不存在")
	}
	if !item.StoryboardSceneID.Valid {
		return nil, apperr.BadRequest("镜头缺少所属场次")
	}
	pid, _, err := authorizeStoryboardScene(ctx, q, item.StoryboardSceneID.Int64, actorUserID)
	return pid, err
}

func (StoryboardItemPatch) Apply(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, _ int64) (resource.Mutation, error) {
	patch, err := asPatch(cmd)
	if err != nil {
		return resource.Mutation{}, err
	}
	current, err := q.GetStoryboardItemByID(ctx, patch.ID)
	if err != nil {
		return resource.Mutation{}, missing(err, "镜头不存在")
	}
	if !current.StoryboardSceneID.Valid {
		return resource.Mutation{}, apperr.BadRequest("镜头缺少所属场次")
	}
	scene, err := q.LockStoryboardSceneByID(ctx, current.StoryboardSceneID.Int64)
	if err != nil {
		return resource.Mutation{}, missing(err, "分镜场次不存在")
	}
	row, err := q.LockStoryboardItemByID(ctx, patch.ID)
	if err != nil {
		return resource.Mutation{}, missing(err, "镜头不存在")
	}
	if row.Version != patch.Version {
		return resource.Mutation{}, apperr.Conflict("镜头版本已变化，请刷新后重试")
	}
	before := resource.StoryboardItemVO(row)
	changed, err := applyShotFields(&row, patch.Patch, false)
	if err != nil {
		return resource.Mutation{}, err
	}
	params := shotUpdate(row)
	n, err := q.UpdateStoryboardItem(ctx, params)
	if err != nil {
		return resource.Mutation{}, err
	}
	if err := oneRow(n, "镜头版本已变化，请刷新后重试"); err != nil {
		return resource.Mutation{}, err
	}
	bumped, err := q.BumpStoryboardSceneVersion(ctx, sqlcdb.BumpStoryboardSceneVersionParams{ID: scene.ID, Version: scene.Version})
	if err != nil {
		return resource.Mutation{}, err
	}
	if err := oneRow(bumped, "分镜场次版本已变化，请刷新后重试"); err != nil {
		return resource.Mutation{}, err
	}
	saved, err := q.GetStoryboardItemByID(ctx, row.ID)
	if err != nil {
		return resource.Mutation{}, err
	}
	board, err := q.GetStoryboardByID(ctx, saved.StoryboardID)
	if err != nil {
		return resource.Mutation{}, err
	}
	projectID, err := boardProject(board)
	if err != nil {
		return resource.Mutation{}, err
	}
	return patched(resource.TypeShot, saved.ID, projectID, saved.Version, changed, before, resource.StoryboardItemVO(saved))
}

type StoryboardItemDelete struct{}

func (StoryboardItemDelete) ResourceType() resource.Type   { return resource.TypeShot }
func (StoryboardItemDelete) Operation() resource.Operation { return resource.OpDelete }

func (StoryboardItemDelete) Authorize(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64, _ *resource.Receipt) (*int64, error) {
	del, err := asDelete(cmd)
	if err != nil {
		return nil, err
	}
	item, err := q.GetStoryboardItemByID(ctx, del.ID)
	if err != nil {
		return nil, missing(err, "镜头不存在")
	}
	if !item.StoryboardSceneID.Valid {
		return nil, apperr.BadRequest("镜头缺少所属场次")
	}
	pid, _, err := authorizeStoryboardScene(ctx, q, item.StoryboardSceneID.Int64, actorUserID)
	return pid, err
}

func (StoryboardItemDelete) Apply(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, _ int64) (resource.Mutation, error) {
	del, err := asDelete(cmd)
	if err != nil {
		return resource.Mutation{}, err
	}
	current, err := q.GetStoryboardItemByID(ctx, del.ID)
	if err != nil {
		return resource.Mutation{}, missing(err, "镜头不存在")
	}
	if !current.StoryboardSceneID.Valid {
		return resource.Mutation{}, apperr.BadRequest("镜头缺少所属场次")
	}
	scene, err := q.LockStoryboardSceneByID(ctx, current.StoryboardSceneID.Int64)
	if err != nil {
		return resource.Mutation{}, missing(err, "分镜场次不存在")
	}
	row, err := q.LockStoryboardItemByID(ctx, del.ID)
	if err != nil {
		return resource.Mutation{}, missing(err, "镜头不存在")
	}
	if row.Version != del.Version {
		return resource.Mutation{}, apperr.Conflict("镜头版本已变化，请刷新后重试")
	}
	n, err := q.SoftDeleteStoryboardItem(ctx, sqlcdb.SoftDeleteStoryboardItemParams{ID: row.ID, Version: row.Version})
	if err != nil {
		return resource.Mutation{}, err
	}
	if err := oneRow(n, "镜头版本已变化，请刷新后重试"); err != nil {
		return resource.Mutation{}, err
	}
	bumped, err := q.BumpStoryboardSceneVersion(ctx, sqlcdb.BumpStoryboardSceneVersionParams{ID: scene.ID, Version: scene.Version})
	if err != nil {
		return resource.Mutation{}, err
	}
	if err := oneRow(bumped, "分镜场次版本已变化，请刷新后重试"); err != nil {
		return resource.Mutation{}, err
	}
	board, err := q.GetStoryboardByID(ctx, row.StoryboardID)
	if err != nil {
		return resource.Mutation{}, err
	}
	projectID, err := boardProject(board)
	if err != nil {
		return resource.Mutation{}, err
	}
	before := resource.StoryboardItemVO(row)
	return deleted(resource.TypeShot, row.ID, projectID, row.Version+1, before, markDeleted(before, row.Version+1))
}

func shotParams(row sqlcdb.AfvStoryboardItem) sqlcdb.CreateStoryboardItemParams {
	return sqlcdb.CreateStoryboardItemParams{
		StoryboardID:        row.StoryboardID,
		StoryboardEpisodeID: row.StoryboardEpisodeID,
		StoryboardSceneID:   row.StoryboardSceneID,
		SortOrder:           row.SortOrder,
		ShotNumber:          row.ShotNumber,
		AutoShotNumber:      row.AutoShotNumber,
		ImageUrl:            row.ImageUrl,
		ReferenceImageUrl:   row.ReferenceImageUrl,
		VideoUrl:            row.VideoUrl,
		GeneratedImageUrl:   row.GeneratedImageUrl,
		FirstFrameImageUrl:  row.FirstFrameImageUrl,
		LastFrameImageUrl:   row.LastFrameImageUrl,
		FirstFramePrompt:    row.FirstFramePrompt,
		LastFramePrompt:     row.LastFramePrompt,
		GeneratedVideoUrl:   row.GeneratedVideoUrl,
		ShotType:            row.ShotType,
		Duration:            row.Duration,
		Content:             row.Content,
		SceneExpectation:    row.SceneExpectation,
		Sound:               row.Sound,
		Dialogue:            row.Dialogue,
		SoundEffect:         row.SoundEffect,
		Music:               row.Music,
		CameraMovement:      row.CameraMovement,
		CameraAngle:         row.CameraAngle,
		CameraEquipment:     row.CameraEquipment,
		FocalLength:         row.FocalLength,
		Transition:          row.Transition,
		CharacterIds:        row.CharacterIds,
		SceneAssetItemID:    row.SceneAssetItemID,
		PropIds:             row.PropIds,
		Remark:              row.Remark,
		CustomData:          row.CustomData,
		AiGenerated:         row.AiGenerated,
		Status:              row.Status,
		VideoPrompt:         row.VideoPrompt,
	}
}

func shotUpdate(row sqlcdb.AfvStoryboardItem) sqlcdb.UpdateStoryboardItemParams {
	return sqlcdb.UpdateStoryboardItemParams{
		ShotNumber:         row.ShotNumber,
		ImageUrl:           row.ImageUrl,
		ReferenceImageUrl:  row.ReferenceImageUrl,
		VideoUrl:           row.VideoUrl,
		GeneratedImageUrl:  row.GeneratedImageUrl,
		FirstFrameImageUrl: row.FirstFrameImageUrl,
		LastFrameImageUrl:  row.LastFrameImageUrl,
		FirstFramePrompt:   row.FirstFramePrompt,
		LastFramePrompt:    row.LastFramePrompt,
		GeneratedVideoUrl:  row.GeneratedVideoUrl,
		ShotType:           row.ShotType,
		Duration:           row.Duration,
		Content:            row.Content,
		SceneExpectation:   row.SceneExpectation,
		Sound:              row.Sound,
		Dialogue:           row.Dialogue,
		SoundEffect:        row.SoundEffect,
		Music:              row.Music,
		CameraMovement:     row.CameraMovement,
		CameraAngle:        row.CameraAngle,
		CameraEquipment:    row.CameraEquipment,
		FocalLength:        row.FocalLength,
		Transition:         row.Transition,
		CharacterIds:       row.CharacterIds,
		SceneAssetItemID:   row.SceneAssetItemID,
		PropIds:            row.PropIds,
		Remark:             row.Remark,
		CustomData:         row.CustomData,
		AiGenerated:        row.AiGenerated,
		Status:             row.Status,
		VideoPrompt:        row.VideoPrompt,
		ID:                 row.ID,
		Version:            row.Version,
	}
}
