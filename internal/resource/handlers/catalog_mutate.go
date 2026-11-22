package handlers

import (
	"context"
	"database/sql"
	"errors"
	"math"

	"github.com/go-sql-driver/mysql"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/apperr"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/resource"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/sqlcdb"
)

func requireOwnedProject(ctx context.Context, q *sqlcdb.Queries, projectID, actorUserID int64) (*int64, error) {
	if projectID <= 0 {
		return nil, apperr.BadRequest("projectId 无效")
	}
	if _, err := q.LockOwnedProjectByID(ctx, sqlcdb.LockOwnedProjectByIDParams{
		ID:      projectID,
		OwnerID: actorUserID,
	}); err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.Forbidden("无权访问该项目")
		}
		return nil, err
	}
	id := projectID
	return &id, nil
}

func missing(err error, msg string) error {
	if err == sql.ErrNoRows {
		return apperr.NotFound(msg)
	}
	return err
}

func oneRow(n int64, msg string) error {
	if n != 1 {
		return apperr.Conflict(msg)
	}
	return nil
}

func duplicate(err error, msg string) error {
	var me *mysql.MySQLError
	if errors.As(err, &me) && me.Number == 1062 {
		return apperr.Conflict(msg)
	}
	return err
}

func asVersion(n int64) (int32, error) {
	if n < 0 || n > math.MaxInt32 {
		return 0, apperr.Conflict("版本号超出范围")
	}
	return int32(n), nil
}

func taskVersion(ctx context.Context, q *sqlcdb.Queries, resType resource.Type, id int64) (int32, error) {
	n, err := q.MaxResourceVersion(ctx, sqlcdb.MaxResourceVersionParams{
		ResourceType: string(resType),
		ResourceID:   id,
	})
	if err != nil {
		return 0, err
	}
	return asVersion(n)
}

func requiredProject(id *int64) (int64, error) {
	if id == nil || *id <= 0 {
		return 0, apperr.BadRequest("projectId 无效")
	}
	return *id, nil
}

func asCreate(cmd resource.Command) (resource.CreateCommand, error) {
	create, ok := cmd.(resource.CreateCommand)
	if !ok {
		return resource.CreateCommand{}, apperr.BadRequest("命令类型无效")
	}
	return create, nil
}

func asPatch(cmd resource.Command) (resource.PatchCommand, error) {
	patch, ok := cmd.(resource.PatchCommand)
	if !ok {
		return resource.PatchCommand{}, apperr.BadRequest("命令类型无效")
	}
	if patch.Patch == nil {
		patch.Patch = map[string]any{}
	}
	return patch, nil
}

func asDelete(cmd resource.Command) (resource.DeleteCommand, error) {
	del, ok := cmd.(resource.DeleteCommand)
	if !ok {
		return resource.DeleteCommand{}, apperr.BadRequest("命令类型无效")
	}
	return del, nil
}

func projectPtr(id int64) *int64 {
	if id <= 0 {
		return nil
	}
	pid := id
	return &pid
}

func created(res resource.Type, id, projectID int64, after any) (resource.Mutation, error) {
	eventID, err := resource.NewEventID()
	if err != nil {
		return resource.Mutation{}, err
	}
	return resource.Mutation{
		Result: after,
		Changes: []resource.Change{{
			EventID:       eventID,
			ResourceType:  res,
			ResourceID:    id,
			ProjectID:     projectPtr(projectID),
			Version:       0,
			Deleted:       false,
			ChangedFields: []string{"*"},
			After:         after,
		}},
	}, nil
}

func patched(res resource.Type, id, projectID int64, version int32, fields []string, before, after any) (resource.Mutation, error) {
	eventID, err := resource.NewEventID()
	if err != nil {
		return resource.Mutation{}, err
	}
	if fields == nil {
		fields = []string{}
	}
	return resource.Mutation{
		Result: after,
		Changes: []resource.Change{{
			EventID:       eventID,
			ResourceType:  res,
			ResourceID:    id,
			ProjectID:     projectPtr(projectID),
			Version:       version,
			Deleted:       false,
			ChangedFields: fields,
			Before:        before,
			After:         after,
		}},
		PreviousValues: before,
	}, nil
}

func deleted(res resource.Type, id, projectID int64, version int32, before, after any) (resource.Mutation, error) {
	eventID, err := resource.NewEventID()
	if err != nil {
		return resource.Mutation{}, err
	}
	return resource.Mutation{
		Result: true,
		Changes: []resource.Change{{
			EventID:       eventID,
			ResourceType:  res,
			ResourceID:    id,
			ProjectID:     projectPtr(projectID),
			Version:       version,
			Deleted:       true,
			ChangedFields: []string{"deleted"},
			Before:        before,
			After:         after,
		}},
		PreviousValues: before,
	}, nil
}

func markDeleted(vo map[string]any, version int32) map[string]any {
	after := cloneMap(vo)
	after["deleted"] = true
	after["version"] = version
	return after
}

func ensureScriptEpisode(ctx context.Context, q *sqlcdb.Queries, episodeID, projectID int64) error {
	ep, err := q.GetScriptEpisodeByID(ctx, episodeID)
	if err != nil {
		return missing(err, "剧本分集不存在")
	}
	script, err := q.GetScriptByID(ctx, ep.ScriptID)
	if err != nil {
		return missing(err, "剧本不存在")
	}
	if script.ProjectID != projectID {
		return apperr.BadRequest("剧本分集不属于该项目")
	}
	return nil
}

func boardProject(row sqlcdb.AfvStoryboard) (int64, error) {
	if !row.ProjectID.Valid || row.ProjectID.Int64 <= 0 {
		return 0, apperr.BadRequest("分镜缺少所属项目")
	}
	return row.ProjectID.Int64, nil
}
