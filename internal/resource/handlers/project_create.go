package handlers

import (
	"context"
	"database/sql"
	"strings"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/apperr"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/resource"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/sqlcdb"
)

type ProjectCreate struct{}

func (ProjectCreate) ResourceType() resource.Type   { return resource.TypeProject }
func (ProjectCreate) Operation() resource.Operation { return resource.OpCreate }

func (ProjectCreate) Authorize(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64, previous *resource.Receipt) (*int64, error) {
	create, ok := cmd.(resource.CreateCommand)
	if !ok {
		return nil, apperr.BadRequest("命令类型无效")
	}
	if create.ProjectID != nil {
		return nil, apperr.BadRequest("新项目不能指定所属项目")
	}
	if previous != nil {
		if id, ok := asInt64(previous.Result, "id"); ok && id > 0 {
			if _, err := q.LockOwnedProjectByID(ctx, sqlcdb.LockOwnedProjectByIDParams{
				ID:      id,
				OwnerID: actorUserID,
			}); err != nil {
				if err == sql.ErrNoRows {
					return nil, apperr.Forbidden("无权访问该项目")
				}
				return nil, err
			}
			return &id, nil
		}
	}
	return nil, nil
}

func (ProjectCreate) Apply(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64) (resource.Mutation, error) {
	create := cmd.(resource.CreateCommand)
	if create.Fields == nil {
		return resource.Mutation{}, apperr.BadRequest("fields 不能为空")
	}
	name, _ := create.Fields["name"].(string)
	name = strings.TrimSpace(name)
	if name == "" {
		return resource.Mutation{}, apperr.BadRequest("项目名称不能为空")
	}

	description, err := optionalNullString(create.Fields["description"])
	if err != nil {
		return resource.Mutation{}, err
	}
	coverURL, err := optionalNullString(create.Fields["coverUrl"])
	if err != nil {
		return resource.Mutation{}, err
	}
	properties, err := optionalJSON(create.Fields["properties"])
	if err != nil {
		return resource.Mutation{}, err
	}
	artStyle, err := optionalNullString(create.Fields["artStyle"])
	if err != nil {
		return resource.Mutation{}, err
	}
	artStyleDescription, err := optionalNullString(create.Fields["artStyleDescription"])
	if err != nil {
		return resource.Mutation{}, err
	}
	artStyleImagePrompt, err := optionalNullString(create.Fields["artStyleImagePrompt"])
	if err != nil {
		return resource.Mutation{}, err
	}
	artStyleImageURL, err := optionalNullString(create.Fields["artStyleImageUrl"])
	if err != nil {
		return resource.Mutation{}, err
	}
	projectID, err := q.CreateOwnedProject(ctx, sqlcdb.CreateOwnedProjectParams{
		Name:                name,
		Description:         description,
		CoverUrl:            coverURL,
		OwnerID:             actorUserID,
		Properties:          properties,
		ArtStyle:            artStyle,
		ArtStyleDescription: artStyleDescription,
		ArtStyleImagePrompt: artStyleImagePrompt,
		ArtStyleImageUrl:    artStyleImageURL,
	})
	if err != nil {
		return resource.Mutation{}, err
	}

	scope := sql.NullInt32{Int32: 2, Valid: true}
	_, err = q.CreateScriptForProject(ctx, sqlcdb.CreateScriptForProjectParams{
		ProjectID: projectID,
		Title:     sql.NullString{String: name, Valid: true},
		Scope:     scope,
		OwnerType: sql.NullInt32{Int32: 1, Valid: true},
		OwnerID:   sql.NullInt64{Int64: actorUserID, Valid: true},
	})
	if err != nil {
		return resource.Mutation{}, err
	}

	row, err := q.GetProjectByID(ctx, projectID)
	if err != nil {
		return resource.Mutation{}, err
	}
	vo := resource.ProjectVO(row)
	eventID, err := resource.NewEventID()
	if err != nil {
		return resource.Mutation{}, err
	}
	pid := projectID
	return resource.Mutation{
		Result: vo,
		Changes: []resource.Change{{
			EventID:       eventID,
			ResourceType:  resource.TypeProject,
			ResourceID:    projectID,
			ProjectID:     &pid,
			Version:       row.Version,
			Deleted:       false,
			ChangedFields: []string{"*"},
			Before:        nil,
			After:         vo,
		}},
		PreviousValues: map[string]any{"created": true},
	}, nil
}
