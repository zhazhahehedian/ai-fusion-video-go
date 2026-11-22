package handlers

import (
	"context"
	"database/sql"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/apperr"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/resource"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/sqlcdb"
)

func authorizeStoryboardEpisode(ctx context.Context, q *sqlcdb.Queries, episodeID, actorUserID int64) (*int64, sqlcdb.AfvStoryboardEpisode, error) {
	ep, err := q.GetStoryboardEpisodeByID(ctx, episodeID)
	if err != nil {
		return nil, sqlcdb.AfvStoryboardEpisode{}, missing(err, "分镜集不存在")
	}
	board, err := q.GetStoryboardByID(ctx, ep.StoryboardID)
	if err != nil {
		return nil, sqlcdb.AfvStoryboardEpisode{}, missing(err, "分镜不存在")
	}
	projectID, err := boardProject(board)
	if err != nil {
		return nil, sqlcdb.AfvStoryboardEpisode{}, err
	}
	pid, err := requireOwnedProject(ctx, q, projectID, actorUserID)
	return pid, ep, err
}

type StoryboardSceneCreate struct{}

func (StoryboardSceneCreate) ResourceType() resource.Type   { return resource.TypeStoryboardScene }
func (StoryboardSceneCreate) Operation() resource.Operation { return resource.OpCreate }

func (StoryboardSceneCreate) Authorize(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64, _ *resource.Receipt) (*int64, error) {
	create, err := asCreate(cmd)
	if err != nil {
		return nil, err
	}
	episodeID, err := requiredPositiveID(create.Fields["episodeId"], "episodeId")
	if err != nil {
		return nil, err
	}
	pid, _, err := authorizeStoryboardEpisode(ctx, q, episodeID, actorUserID)
	return pid, err
}

func (StoryboardSceneCreate) Apply(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, _ int64) (resource.Mutation, error) {
	create, err := asCreate(cmd)
	if err != nil {
		return resource.Mutation{}, err
	}
	episodeID, err := requiredPositiveID(create.Fields["episodeId"], "episodeId")
	if err != nil {
		return resource.Mutation{}, err
	}
	parentVersion, err := requiredNonNegInt(create.Fields["parentVersion"], "parentVersion")
	if err != nil {
		return resource.Mutation{}, err
	}
	ep, err := q.LockStoryboardEpisodeByID(ctx, episodeID)
	if err != nil {
		return resource.Mutation{}, missing(err, "分镜集不存在")
	}
	if ep.Version != parentVersion {
		return resource.Mutation{}, apperr.Conflict("分镜集版本已变化，请刷新后重试")
	}
	fields := cloneMap(create.Fields)
	delete(fields, "episodeId")
	delete(fields, "parentVersion")
	count, err := q.CountStoryboardScenesByEpisode(ctx, ep.ID)
	if err != nil {
		return resource.Mutation{}, err
	}
	row := sqlcdb.AfvStoryboardScene{
		EpisodeID:    ep.ID,
		StoryboardID: ep.StoryboardID,
		SortOrder:    sql.NullInt32{Int32: int32(count), Valid: true},
		Status:       sql.NullInt32{Int32: 0, Valid: true},
	}
	if _, err := applySceneFields(&row, fields, true); err != nil {
		return resource.Mutation{}, err
	}
	id, err := q.CreateStoryboardScene(ctx, sqlcdb.CreateStoryboardSceneParams{
		EpisodeID:    row.EpisodeID,
		StoryboardID: row.StoryboardID,
		SceneNumber:  row.SceneNumber,
		SceneHeading: row.SceneHeading,
		Location:     row.Location,
		TimeOfDay:    row.TimeOfDay,
		IntExt:       row.IntExt,
		SortOrder:    row.SortOrder,
		Status:       row.Status,
	})
	if err != nil {
		return resource.Mutation{}, err
	}
	bumped, err := q.BumpStoryboardEpisodeVersion(ctx, sqlcdb.BumpStoryboardEpisodeVersionParams{ID: ep.ID, Version: ep.Version})
	if err != nil {
		return resource.Mutation{}, err
	}
	if err := oneRow(bumped, "分镜集版本已变化，请刷新后重试"); err != nil {
		return resource.Mutation{}, err
	}
	saved, err := q.GetStoryboardSceneByID(ctx, id)
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
	return created(resource.TypeStoryboardScene, id, projectID, resource.StoryboardSceneVO(saved))
}

type StoryboardScenePatch struct{}

func (StoryboardScenePatch) ResourceType() resource.Type   { return resource.TypeStoryboardScene }
func (StoryboardScenePatch) Operation() resource.Operation { return resource.OpPatch }

func (StoryboardScenePatch) Authorize(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64, _ *resource.Receipt) (*int64, error) {
	patch, err := asPatch(cmd)
	if err != nil {
		return nil, err
	}
	scene, err := q.GetStoryboardSceneByID(ctx, patch.ID)
	if err != nil {
		return nil, missing(err, "分镜场次不存在")
	}
	pid, _, err := authorizeStoryboardEpisode(ctx, q, scene.EpisodeID, actorUserID)
	return pid, err
}

func (StoryboardScenePatch) Apply(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, _ int64) (resource.Mutation, error) {
	patch, err := asPatch(cmd)
	if err != nil {
		return resource.Mutation{}, err
	}
	current, err := q.GetStoryboardSceneByID(ctx, patch.ID)
	if err != nil {
		return resource.Mutation{}, missing(err, "分镜场次不存在")
	}
	ep, err := q.LockStoryboardEpisodeByID(ctx, current.EpisodeID)
	if err != nil {
		return resource.Mutation{}, missing(err, "分镜集不存在")
	}
	row, err := q.LockStoryboardSceneByID(ctx, patch.ID)
	if err != nil {
		return resource.Mutation{}, missing(err, "分镜场次不存在")
	}
	if row.Version != patch.Version {
		return resource.Mutation{}, apperr.Conflict("分镜场次版本已变化，请刷新后重试")
	}
	before := resource.StoryboardSceneVO(row)
	changed, err := applySceneFields(&row, patch.Patch, false)
	if err != nil {
		return resource.Mutation{}, err
	}
	n, err := q.UpdateStoryboardScene(ctx, sqlcdb.UpdateStoryboardSceneParams{
		SceneNumber:  row.SceneNumber,
		SceneHeading: row.SceneHeading,
		Location:     row.Location,
		TimeOfDay:    row.TimeOfDay,
		IntExt:       row.IntExt,
		Status:       row.Status,
		ID:           row.ID,
		Version:      row.Version,
	})
	if err != nil {
		return resource.Mutation{}, err
	}
	if err := oneRow(n, "分镜场次版本已变化，请刷新后重试"); err != nil {
		return resource.Mutation{}, err
	}
	bumped, err := q.BumpStoryboardEpisodeVersion(ctx, sqlcdb.BumpStoryboardEpisodeVersionParams{ID: ep.ID, Version: ep.Version})
	if err != nil {
		return resource.Mutation{}, err
	}
	if err := oneRow(bumped, "分镜集版本已变化，请刷新后重试"); err != nil {
		return resource.Mutation{}, err
	}
	saved, err := q.GetStoryboardSceneByID(ctx, row.ID)
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
	return patched(resource.TypeStoryboardScene, saved.ID, projectID, saved.Version, changed, before, resource.StoryboardSceneVO(saved))
}

type StoryboardSceneDelete struct{}

func (StoryboardSceneDelete) ResourceType() resource.Type   { return resource.TypeStoryboardScene }
func (StoryboardSceneDelete) Operation() resource.Operation { return resource.OpDelete }

func (StoryboardSceneDelete) Authorize(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64, _ *resource.Receipt) (*int64, error) {
	del, err := asDelete(cmd)
	if err != nil {
		return nil, err
	}
	scene, err := q.GetStoryboardSceneByID(ctx, del.ID)
	if err != nil {
		return nil, missing(err, "分镜场次不存在")
	}
	pid, _, err := authorizeStoryboardEpisode(ctx, q, scene.EpisodeID, actorUserID)
	return pid, err
}

func (StoryboardSceneDelete) Apply(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, _ int64) (resource.Mutation, error) {
	del, err := asDelete(cmd)
	if err != nil {
		return resource.Mutation{}, err
	}
	current, err := q.GetStoryboardSceneByID(ctx, del.ID)
	if err != nil {
		return resource.Mutation{}, missing(err, "分镜场次不存在")
	}
	ep, err := q.LockStoryboardEpisodeByID(ctx, current.EpisodeID)
	if err != nil {
		return resource.Mutation{}, missing(err, "分镜集不存在")
	}
	row, err := q.LockStoryboardSceneByID(ctx, del.ID)
	if err != nil {
		return resource.Mutation{}, missing(err, "分镜场次不存在")
	}
	if row.Version != del.Version {
		return resource.Mutation{}, apperr.Conflict("分镜场次版本已变化，请刷新后重试")
	}
	items, err := q.CountStoryboardItemsByScene(ctx, sql.NullInt64{Int64: row.ID, Valid: true})
	if err != nil {
		return resource.Mutation{}, err
	}
	if items > 0 {
		return resource.Mutation{}, apperr.Conflict("请先删除场次下的镜头")
	}
	n, err := q.SoftDeleteStoryboardScene(ctx, sqlcdb.SoftDeleteStoryboardSceneParams{ID: row.ID, Version: row.Version})
	if err != nil {
		return resource.Mutation{}, err
	}
	if err := oneRow(n, "分镜场次版本已变化，请刷新后重试"); err != nil {
		return resource.Mutation{}, err
	}
	bumped, err := q.BumpStoryboardEpisodeVersion(ctx, sqlcdb.BumpStoryboardEpisodeVersionParams{ID: ep.ID, Version: ep.Version})
	if err != nil {
		return resource.Mutation{}, err
	}
	if err := oneRow(bumped, "分镜集版本已变化，请刷新后重试"); err != nil {
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
	before := resource.StoryboardSceneVO(row)
	return deleted(resource.TypeStoryboardScene, row.ID, projectID, row.Version+1, before, markDeleted(before, row.Version+1))
}
