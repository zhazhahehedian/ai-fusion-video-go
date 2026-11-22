package handlers

import (
	"context"
	"database/sql"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/apperr"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/resource"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/sqlcdb"
)

type StoryboardClear struct{}

func (StoryboardClear) ResourceType() resource.Type   { return resource.TypeStoryboard }
func (StoryboardClear) Operation() resource.Operation { return resource.OpClearContent }

func (StoryboardClear) Authorize(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64, _ *resource.Receipt) (*int64, error) {
	clear, err := asClear(cmd)
	if err != nil {
		return nil, err
	}
	board, err := q.GetStoryboardByID(ctx, clear.ID)
	if err != nil {
		return nil, missing(err, "分镜不存在")
	}
	projectID, err := boardProject(board)
	if err != nil {
		return nil, err
	}
	return requireOwnedProject(ctx, q, projectID, actorUserID)
}

func (StoryboardClear) Apply(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, _ int64) (resource.Mutation, error) {
	clear, err := asClear(cmd)
	if err != nil {
		return resource.Mutation{}, err
	}
	board, err := q.LockStoryboardByID(ctx, clear.ID)
	if err != nil {
		return resource.Mutation{}, missing(err, "分镜不存在")
	}
	if board.Version != clear.Version {
		return resource.Mutation{}, apperr.Conflict("分镜版本已变化，请刷新后重试")
	}
	projectID, err := boardProject(board)
	if err != nil {
		return resource.Mutation{}, err
	}
	episodes, err := q.LockStoryboardEpisodesByBoard(ctx, board.ID)
	if err != nil {
		return resource.Mutation{}, err
	}
	scenes, err := q.LockStoryboardScenesByBoard(ctx, board.ID)
	if err != nil {
		return resource.Mutation{}, err
	}
	shots, err := q.LockStoryboardItemsByBoard(ctx, board.ID)
	if err != nil {
		return resource.Mutation{}, err
	}
	changes := make([]resource.Change, 0, len(shots)+len(scenes)+len(episodes)+1)
	for _, shot := range shots {
		change, err := softDeleteShot(ctx, q, shot, projectID)
		if err != nil {
			return resource.Mutation{}, err
		}
		changes = append(changes, change)
	}
	for _, scene := range scenes {
		change, err := softDeleteStoryboardScene(ctx, q, scene, projectID)
		if err != nil {
			return resource.Mutation{}, err
		}
		changes = append(changes, change)
	}
	for _, episode := range episodes {
		n, err := q.SoftDeleteStoryboardEpisode(ctx, sqlcdb.SoftDeleteStoryboardEpisodeParams{
			ID: episode.ID, Version: episode.Version,
		})
		if err != nil {
			return resource.Mutation{}, err
		}
		if err := oneRow(n, "分镜集版本已变化，请刷新后重试"); err != nil {
			return resource.Mutation{}, err
		}
		before := resource.StoryboardEpisodeVO(episode)
		after := markDeleted(before, episode.Version+1)
		after["deletedId"] = episode.ID
		change, err := deleteChange(resource.TypeStoryboardEpisode, episode.ID, projectID, episode.Version+1, before, after, []string{"deleted"})
		if err != nil {
			return resource.Mutation{}, err
		}
		changes = append(changes, change)
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
	saved, err := q.GetStoryboardByID(ctx, board.ID)
	if err != nil {
		return resource.Mutation{}, err
	}
	before := resource.StoryboardVO(board)
	after := resource.StoryboardVO(saved)
	parent, err := contentChange(resource.TypeStoryboard, board.ID, projectID, saved.Version, []string{"children"}, before, after)
	if err != nil {
		return resource.Mutation{}, err
	}
	changes = append(changes, parent)
	return resource.Mutation{Result: after, Changes: changes, PreviousValues: before}, nil
}

type StoryboardEpisodeClear struct{}

func (StoryboardEpisodeClear) ResourceType() resource.Type {
	return resource.TypeStoryboardEpisode
}
func (StoryboardEpisodeClear) Operation() resource.Operation { return resource.OpClearContent }

func (StoryboardEpisodeClear) Authorize(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64, _ *resource.Receipt) (*int64, error) {
	clear, err := asClear(cmd)
	if err != nil {
		return nil, err
	}
	pid, _, err := authorizeStoryboardEpisode(ctx, q, clear.ID, actorUserID)
	return pid, err
}

func (StoryboardEpisodeClear) Apply(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, _ int64) (resource.Mutation, error) {
	clear, err := asClear(cmd)
	if err != nil {
		return resource.Mutation{}, err
	}
	current, err := q.GetStoryboardEpisodeByID(ctx, clear.ID)
	if err != nil {
		return resource.Mutation{}, missing(err, "分镜集不存在")
	}
	board, err := q.LockStoryboardByID(ctx, current.StoryboardID)
	if err != nil {
		return resource.Mutation{}, missing(err, "分镜不存在")
	}
	row, err := q.LockStoryboardEpisodeByID(ctx, clear.ID)
	if err != nil {
		return resource.Mutation{}, missing(err, "分镜集不存在")
	}
	if row.Version != clear.Version {
		return resource.Mutation{}, apperr.Conflict("分镜集版本已变化，请刷新后重试")
	}
	projectID, err := boardProject(board)
	if err != nil {
		return resource.Mutation{}, err
	}
	scenes, err := q.LockStoryboardScenesByEpisode(ctx, row.ID)
	if err != nil {
		return resource.Mutation{}, err
	}
	shots, err := q.LockStoryboardItemsByEpisode(ctx, sqlcdb.LockStoryboardItemsByEpisodeParams{
		StoryboardEpisodeID: sql.NullInt64{Int64: row.ID, Valid: true},
		EpisodeID:           row.ID,
	})
	if err != nil {
		return resource.Mutation{}, err
	}
	changes := make([]resource.Change, 0, len(shots)+len(scenes)+2)
	for _, shot := range shots {
		change, err := softDeleteShot(ctx, q, shot, projectID)
		if err != nil {
			return resource.Mutation{}, err
		}
		changes = append(changes, change)
	}
	for _, scene := range scenes {
		change, err := softDeleteStoryboardScene(ctx, q, scene, projectID)
		if err != nil {
			return resource.Mutation{}, err
		}
		changes = append(changes, change)
	}
	bumped, err := q.BumpStoryboardEpisodeVersion(ctx, sqlcdb.BumpStoryboardEpisodeVersionParams{
		ID: row.ID, Version: row.Version,
	})
	if err != nil {
		return resource.Mutation{}, err
	}
	if err := oneRow(bumped, "分镜集版本已变化，请刷新后重试"); err != nil {
		return resource.Mutation{}, err
	}
	boardBumped, err := q.BumpStoryboardVersion(ctx, sqlcdb.BumpStoryboardVersionParams{
		ID: board.ID, Version: board.Version,
	})
	if err != nil {
		return resource.Mutation{}, err
	}
	if err := oneRow(boardBumped, "分镜版本已变化，请刷新后重试"); err != nil {
		return resource.Mutation{}, err
	}
	saved, err := q.GetStoryboardEpisodeByID(ctx, row.ID)
	if err != nil {
		return resource.Mutation{}, err
	}
	before := resource.StoryboardEpisodeVO(row)
	after := resource.StoryboardEpisodeVO(saved)
	parent, err := contentChange(resource.TypeStoryboardEpisode, row.ID, projectID, saved.Version, []string{"children"}, before, after)
	if err != nil {
		return resource.Mutation{}, err
	}
	changes = append(changes, parent)
	savedBoard, err := q.GetStoryboardByID(ctx, board.ID)
	if err != nil {
		return resource.Mutation{}, err
	}
	boardChange, err := contentChange(resource.TypeStoryboard, board.ID, projectID, savedBoard.Version, []string{"children"}, resource.StoryboardVO(board), resource.StoryboardVO(savedBoard))
	if err != nil {
		return resource.Mutation{}, err
	}
	changes = append(changes, boardChange)
	return resource.Mutation{Result: after, Changes: changes, PreviousValues: before}, nil
}

func asClear(cmd resource.Command) (resource.ClearCommand, error) {
	clear, ok := cmd.(resource.ClearCommand)
	if !ok {
		return resource.ClearCommand{}, apperr.BadRequest("命令类型无效")
	}
	if clear.ID <= 0 {
		return resource.ClearCommand{}, apperr.BadRequest("id 无效")
	}
	if clear.Version < 0 {
		return resource.ClearCommand{}, apperr.BadRequest("version 无效")
	}
	return clear, nil
}

func contentChange(res resource.Type, id, projectID int64, version int32, fields []string, before, after any) (resource.Change, error) {
	eventID, err := resource.NewEventID()
	if err != nil {
		return resource.Change{}, err
	}
	return resource.Change{
		EventID:       eventID,
		ResourceType:  res,
		ResourceID:    id,
		ProjectID:     projectPtr(projectID),
		Version:       version,
		ChangedFields: fields,
		Before:        before,
		After:         after,
	}, nil
}
