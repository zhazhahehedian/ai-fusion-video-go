package handlers

import (
	"context"
	"database/sql"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/apperr"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/resource"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/sqlcdb"
)

type StoryboardCreate struct{}

func (StoryboardCreate) ResourceType() resource.Type   { return resource.TypeStoryboard }
func (StoryboardCreate) Operation() resource.Operation { return resource.OpCreate }

func (StoryboardCreate) Authorize(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64, _ *resource.Receipt) (*int64, error) {
	create, err := asCreate(cmd)
	if err != nil {
		return nil, err
	}
	projectID, err := requiredProject(create.ProjectID)
	if err != nil {
		return nil, err
	}
	return requireOwnedProject(ctx, q, projectID, actorUserID)
}

func (StoryboardCreate) Apply(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64) (resource.Mutation, error) {
	create, err := asCreate(cmd)
	if err != nil {
		return resource.Mutation{}, err
	}
	projectID, err := requiredProject(create.ProjectID)
	if err != nil {
		return resource.Mutation{}, err
	}
	row := sqlcdb.AfvStoryboard{
		ProjectID: sql.NullInt64{Int64: projectID, Valid: true},
		Scope:     sql.NullInt32{Int32: 3, Valid: true},
		OwnerType: sql.NullInt32{Int32: 1, Valid: true},
		OwnerID:   sql.NullInt64{Int64: actorUserID, Valid: true},
		Status:    sql.NullInt32{Int32: 0, Valid: true},
	}
	if _, err := applyStoryboardFields(&row, create.Fields, true); err != nil {
		return resource.Mutation{}, err
	}
	if row.ScriptID.Valid {
		script, err := q.GetScriptByID(ctx, row.ScriptID.Int64)
		if err != nil {
			return resource.Mutation{}, missing(err, "剧本不存在")
		}
		if script.ProjectID != projectID {
			return resource.Mutation{}, apperr.BadRequest("剧本不属于该项目")
		}
	}
	id, err := q.CreateStoryboard(ctx, sqlcdb.CreateStoryboardParams{
		ProjectID:     row.ProjectID,
		ScriptID:      row.ScriptID,
		Title:         row.Title,
		Description:   row.Description,
		CustomColumns: row.CustomColumns,
		Scope:         row.Scope,
		OwnerID:       row.OwnerID,
		Status:        row.Status,
	})
	if err != nil {
		return resource.Mutation{}, err
	}
	saved, err := q.GetStoryboardByID(ctx, id)
	if err != nil {
		return resource.Mutation{}, err
	}
	return created(resource.TypeStoryboard, id, projectID, resource.StoryboardVO(saved))
}

type StoryboardPatch struct{}

func (StoryboardPatch) ResourceType() resource.Type   { return resource.TypeStoryboard }
func (StoryboardPatch) Operation() resource.Operation { return resource.OpPatch }

func (StoryboardPatch) Authorize(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64, _ *resource.Receipt) (*int64, error) {
	patch, err := asPatch(cmd)
	if err != nil {
		return nil, err
	}
	row, err := q.GetStoryboardByID(ctx, patch.ID)
	if err != nil {
		return nil, missing(err, "分镜不存在")
	}
	projectID, err := boardProject(row)
	if err != nil {
		return nil, err
	}
	return requireOwnedProject(ctx, q, projectID, actorUserID)
}

func (StoryboardPatch) Apply(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, _ int64) (resource.Mutation, error) {
	patch, err := asPatch(cmd)
	if err != nil {
		return resource.Mutation{}, err
	}
	row, err := q.LockStoryboardByID(ctx, patch.ID)
	if err != nil {
		return resource.Mutation{}, missing(err, "分镜不存在")
	}
	if row.Version != patch.Version {
		return resource.Mutation{}, apperr.Conflict("分镜版本已变化，请刷新后重试")
	}
	before := resource.StoryboardVO(row)
	changed, err := applyStoryboardFields(&row, patch.Patch, false)
	if err != nil {
		return resource.Mutation{}, err
	}
	n, err := q.UpdateStoryboard(ctx, sqlcdb.UpdateStoryboardParams{
		Description:   row.Description,
		CustomColumns: row.CustomColumns,
		Status:        row.Status,
		ID:            row.ID,
		Version:       row.Version,
	})
	if err != nil {
		return resource.Mutation{}, err
	}
	if err := oneRow(n, "分镜版本已变化，请刷新后重试"); err != nil {
		return resource.Mutation{}, err
	}
	saved, err := q.GetStoryboardByID(ctx, row.ID)
	if err != nil {
		return resource.Mutation{}, err
	}
	projectID, err := boardProject(saved)
	if err != nil {
		return resource.Mutation{}, err
	}
	return patched(resource.TypeStoryboard, saved.ID, projectID, saved.Version, changed, before, resource.StoryboardVO(saved))
}
