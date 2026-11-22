package handlers

import (
	"context"
	"database/sql"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/apperr"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/resource"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/sqlcdb"
)

type StoryboardEpisodeCreate struct{}

func (StoryboardEpisodeCreate) ResourceType() resource.Type   { return resource.TypeStoryboardEpisode }
func (StoryboardEpisodeCreate) Operation() resource.Operation { return resource.OpCreate }

func (StoryboardEpisodeCreate) Authorize(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64, _ *resource.Receipt) (*int64, error) {
	create, err := asCreate(cmd)
	if err != nil {
		return nil, err
	}
	boardID, err := requiredPositiveID(create.Fields["storyboardId"], "storyboardId")
	if err != nil {
		return nil, err
	}
	board, err := q.GetStoryboardByID(ctx, boardID)
	if err != nil {
		return nil, missing(err, "分镜不存在")
	}
	projectID, err := boardProject(board)
	if err != nil {
		return nil, err
	}
	return requireOwnedProject(ctx, q, projectID, actorUserID)
}

func (StoryboardEpisodeCreate) Apply(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, _ int64) (resource.Mutation, error) {
	create, err := asCreate(cmd)
	if err != nil {
		return resource.Mutation{}, err
	}
	boardID, err := requiredPositiveID(create.Fields["storyboardId"], "storyboardId")
	if err != nil {
		return resource.Mutation{}, err
	}
	parentVersion, err := requiredNonNegInt(create.Fields["parentVersion"], "parentVersion")
	if err != nil {
		return resource.Mutation{}, err
	}
	board, err := q.LockStoryboardByID(ctx, boardID)
	if err != nil {
		return resource.Mutation{}, missing(err, "分镜不存在")
	}
	if board.Version != parentVersion {
		return resource.Mutation{}, apperr.Conflict("分镜版本已变化，请刷新后重试")
	}
	projectID, err := boardProject(board)
	if err != nil {
		return resource.Mutation{}, err
	}
	fields := cloneMap(create.Fields)
	delete(fields, "storyboardId")
	delete(fields, "parentVersion")
	count, err := q.CountStoryboardEpisodes(ctx, board.ID)
	if err != nil {
		return resource.Mutation{}, err
	}
	row := sqlcdb.AfvStoryboardEpisode{
		StoryboardID: board.ID,
		SortOrder:    sql.NullInt32{Int32: int32(count), Valid: true},
		Status:       sql.NullInt32{Int32: 0, Valid: true},
	}
	if _, err := applyEpisodeFields(&row, fields, true); err != nil {
		return resource.Mutation{}, err
	}
	if row.ScriptEpisodeID.Valid {
		if err := ensureScriptEpisode(ctx, q, row.ScriptEpisodeID.Int64, projectID); err != nil {
			return resource.Mutation{}, err
		}
	}
	id, err := q.CreateStoryboardEpisode(ctx, sqlcdb.CreateStoryboardEpisodeParams{
		StoryboardID:    row.StoryboardID,
		ScriptEpisodeID: row.ScriptEpisodeID,
		EpisodeNumber:   row.EpisodeNumber,
		Title:           row.Title,
		Synopsis:        row.Synopsis,
		SortOrder:       row.SortOrder,
		Status:          row.Status,
	})
	if err != nil {
		return resource.Mutation{}, duplicate(err, "该剧本分集已关联到此分镜")
	}
	bumped, err := q.BumpStoryboardVersion(ctx, sqlcdb.BumpStoryboardVersionParams{
		ID: board.ID, Version: board.Version,
	})
	if err != nil {
		return resource.Mutation{}, err
	}
	if err := oneRow(bumped, "分镜版本已变化，请刷新后重试"); err != nil {
		return resource.Mutation{}, err
	}
	saved, err := q.GetStoryboardEpisodeByID(ctx, id)
	if err != nil {
		return resource.Mutation{}, err
	}
	return created(resource.TypeStoryboardEpisode, id, projectID, resource.StoryboardEpisodeVO(saved))
}

type StoryboardEpisodePatch struct{}

func (StoryboardEpisodePatch) ResourceType() resource.Type {
	return resource.TypeStoryboardEpisode
}
func (StoryboardEpisodePatch) Operation() resource.Operation { return resource.OpPatch }

func (StoryboardEpisodePatch) Authorize(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64, _ *resource.Receipt) (*int64, error) {
	patch, err := asPatch(cmd)
	if err != nil {
		return nil, err
	}
	row, err := q.GetStoryboardEpisodeByID(ctx, patch.ID)
	if err != nil {
		return nil, missing(err, "分镜集不存在")
	}
	board, err := q.GetStoryboardByID(ctx, row.StoryboardID)
	if err != nil {
		return nil, missing(err, "分镜不存在")
	}
	projectID, err := boardProject(board)
	if err != nil {
		return nil, err
	}
	return requireOwnedProject(ctx, q, projectID, actorUserID)
}

func (StoryboardEpisodePatch) Apply(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, _ int64) (resource.Mutation, error) {
	patch, err := asPatch(cmd)
	if err != nil {
		return resource.Mutation{}, err
	}
	current, err := q.GetStoryboardEpisodeByID(ctx, patch.ID)
	if err != nil {
		return resource.Mutation{}, missing(err, "分镜集不存在")
	}
	board, err := q.LockStoryboardByID(ctx, current.StoryboardID)
	if err != nil {
		return resource.Mutation{}, missing(err, "分镜不存在")
	}
	row, err := q.LockStoryboardEpisodeByID(ctx, patch.ID)
	if err != nil {
		return resource.Mutation{}, missing(err, "分镜集不存在")
	}
	if row.Version != patch.Version {
		return resource.Mutation{}, apperr.Conflict("分镜集版本已变化，请刷新后重试")
	}
	projectID, err := boardProject(board)
	if err != nil {
		return resource.Mutation{}, err
	}
	before := resource.StoryboardEpisodeVO(row)
	changed, err := applyEpisodeFields(&row, patch.Patch, false)
	if err != nil {
		return resource.Mutation{}, err
	}
	if row.ScriptEpisodeID.Valid {
		if err := ensureScriptEpisode(ctx, q, row.ScriptEpisodeID.Int64, projectID); err != nil {
			return resource.Mutation{}, err
		}
	}
	n, err := q.UpdateStoryboardEpisode(ctx, sqlcdb.UpdateStoryboardEpisodeParams{
		ScriptEpisodeID: row.ScriptEpisodeID,
		EpisodeNumber:   row.EpisodeNumber,
		Title:           row.Title,
		Synopsis:        row.Synopsis,
		Status:          row.Status,
		ID:              row.ID,
		Version:         row.Version,
	})
	if err != nil {
		return resource.Mutation{}, duplicate(err, "该剧本分集已关联到此分镜")
	}
	if err := oneRow(n, "分镜集版本已变化，请刷新后重试"); err != nil {
		return resource.Mutation{}, err
	}
	bumped, err := q.BumpStoryboardVersion(ctx, sqlcdb.BumpStoryboardVersionParams{
		ID: board.ID, Version: board.Version,
	})
	if err != nil {
		return resource.Mutation{}, err
	}
	if err := oneRow(bumped, "分镜版本已变化，请刷新后重试"); err != nil {
		return resource.Mutation{}, err
	}
	saved, err := q.GetStoryboardEpisodeByID(ctx, row.ID)
	if err != nil {
		return resource.Mutation{}, err
	}
	return patched(resource.TypeStoryboardEpisode, saved.ID, projectID, saved.Version, changed, before, resource.StoryboardEpisodeVO(saved))
}

type StoryboardEpisodeDelete struct{}

func (StoryboardEpisodeDelete) ResourceType() resource.Type {
	return resource.TypeStoryboardEpisode
}
func (StoryboardEpisodeDelete) Operation() resource.Operation { return resource.OpDelete }

func (StoryboardEpisodeDelete) Authorize(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64, _ *resource.Receipt) (*int64, error) {
	del, err := asDelete(cmd)
	if err != nil {
		return nil, err
	}
	row, err := q.GetStoryboardEpisodeByID(ctx, del.ID)
	if err != nil {
		return nil, missing(err, "分镜集不存在")
	}
	board, err := q.GetStoryboardByID(ctx, row.StoryboardID)
	if err != nil {
		return nil, missing(err, "分镜不存在")
	}
	projectID, err := boardProject(board)
	if err != nil {
		return nil, err
	}
	return requireOwnedProject(ctx, q, projectID, actorUserID)
}

func (StoryboardEpisodeDelete) Apply(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, _ int64) (resource.Mutation, error) {
	del, err := asDelete(cmd)
	if err != nil {
		return resource.Mutation{}, err
	}
	current, err := q.GetStoryboardEpisodeByID(ctx, del.ID)
	if err != nil {
		return resource.Mutation{}, missing(err, "分镜集不存在")
	}
	board, err := q.LockStoryboardByID(ctx, current.StoryboardID)
	if err != nil {
		return resource.Mutation{}, missing(err, "分镜不存在")
	}
	row, err := q.LockStoryboardEpisodeByID(ctx, del.ID)
	if err != nil {
		return resource.Mutation{}, missing(err, "分镜集不存在")
	}
	if row.Version != del.Version {
		return resource.Mutation{}, apperr.Conflict("分镜集版本已变化，请刷新后重试")
	}
	scenes, err := q.CountStoryboardScenesByEpisode(ctx, row.ID)
	if err != nil {
		return resource.Mutation{}, err
	}
	if scenes > 0 {
		return resource.Mutation{}, apperr.Conflict("请先删除分集下的场次")
	}
	n, err := q.SoftDeleteStoryboardEpisode(ctx, sqlcdb.SoftDeleteStoryboardEpisodeParams{ID: row.ID, Version: row.Version})
	if err != nil {
		return resource.Mutation{}, err
	}
	if err := oneRow(n, "分镜集版本已变化，请刷新后重试"); err != nil {
		return resource.Mutation{}, err
	}
	bumped, err := q.BumpStoryboardVersion(ctx, sqlcdb.BumpStoryboardVersionParams{
		ID: board.ID, Version: board.Version,
	})
	if err != nil {
		return resource.Mutation{}, err
	}
	if err := oneRow(bumped, "分镜版本已变化，请刷新后重试"); err != nil {
		return resource.Mutation{}, err
	}
	projectID, err := boardProject(board)
	if err != nil {
		return resource.Mutation{}, err
	}
	before := resource.StoryboardEpisodeVO(row)
	after := markDeleted(before, row.Version+1)
	after["deletedId"] = row.ID
	return deleted(resource.TypeStoryboardEpisode, row.ID, projectID, row.Version+1, before, after)
}
