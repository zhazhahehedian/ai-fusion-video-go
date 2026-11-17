package handlers

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/apperr"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/resource"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/sqlcdb"
)

type ScriptPatch struct{}

func (ScriptPatch) ResourceType() resource.Type   { return resource.TypeScript }
func (ScriptPatch) Operation() resource.Operation { return resource.OpPatch }

func (ScriptPatch) Authorize(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64, previous *resource.Receipt) (*int64, error) {
	patch, ok := cmd.(resource.PatchCommand)
	if !ok {
		return nil, apperr.BadRequest("命令类型无效")
	}
	return authorizeScriptOwner(ctx, q, patch.ID, actorUserID)
}

func (ScriptPatch) Apply(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64) (resource.Mutation, error) {
	patch := cmd.(resource.PatchCommand)
	if len(patch.Patch) == 0 {
		return resource.Mutation{}, apperr.BadRequest("修改字段不能为空")
	}
	script, err := q.LockScriptByID(ctx, patch.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return resource.Mutation{}, apperr.NotFound("剧本不存在")
		}
		return resource.Mutation{}, err
	}
	if nullInt32Value(script.Version) != patch.Version {
		return resource.Mutation{}, apperr.Conflict("剧本版本已变化，请刷新后重试")
	}
	before := resource.ScriptVO(script)

	content := script.Content
	rawContent := script.RawContent
	storySynopsis := script.StorySynopsis
	charactersJSON := script.CharactersJson
	sourceType := script.SourceType
	parsingStatus := script.ParsingStatus
	parsingProgress := script.ParsingProgress
	summary := script.Summary
	genre := script.Genre
	targetAudience := script.TargetAudience
	durationEstimate := script.DurationEstimate
	status := script.Status
	totalEpisodes := script.TotalEpisodes
	changed := make([]string, 0, len(patch.Patch))

	for key, value := range patch.Patch {
		changed = append(changed, key)
		switch key {
		case "content":
			content = optionalNullString(value)
		case "rawContent":
			rawContent = optionalNullString(value)
		case "storySynopsis":
			storySynopsis = optionalNullString(value)
		case "charactersJson":
			charactersJSON = optionalJSONArray(value)
		case "sourceType":
			n, err := requiredNonNegInt(value, "sourceType")
			if err != nil {
				return resource.Mutation{}, err
			}
			sourceType = sql.NullInt32{Int32: n, Valid: true}
		case "parsingStatus":
			n, err := requiredNonNegInt(value, "parsingStatus")
			if err != nil {
				return resource.Mutation{}, err
			}
			parsingStatus = sql.NullInt32{Int32: n, Valid: true}
		case "parsingProgress":
			parsingProgress = optionalNullString(value)
		case "summary":
			summary = optionalNullString(value)
		case "genre":
			genre = optionalNullString(value)
		case "targetAudience":
			targetAudience = optionalNullString(value)
		case "durationEstimate":
			durationEstimate = optionalNullInt32(value)
		case "status":
			n, err := requiredNonNegInt(value, "status")
			if err != nil {
				return resource.Mutation{}, err
			}
			status = sql.NullInt32{Int32: n, Valid: true}
		default:
			return resource.Mutation{}, apperr.BadRequest("不允许修改字段：" + key)
		}
	}

	updated, err := q.UpdateScript(ctx, sqlcdb.UpdateScriptParams{
		Content:          content,
		RawContent:       rawContent,
		StorySynopsis:    storySynopsis,
		CharactersJson:   charactersJSON,
		SourceType:       sourceType,
		ParsingStatus:    parsingStatus,
		ParsingProgress:  parsingProgress,
		Summary:          summary,
		Genre:            genre,
		TargetAudience:   targetAudience,
		DurationEstimate: durationEstimate,
		Status:           status,
		TotalEpisodes:    totalEpisodes,
		ID:               script.ID,
		Version:          script.Version,
	})
	if err != nil {
		return resource.Mutation{}, err
	}
	if updated != 1 {
		return resource.Mutation{}, apperr.Conflict("剧本版本已变化，请刷新后重试")
	}
	afterRow, err := q.GetScriptByID(ctx, script.ID)
	if err != nil {
		return resource.Mutation{}, err
	}
	vo := resource.ScriptVO(afterRow)
	eventID, err := resource.NewEventID()
	if err != nil {
		return resource.Mutation{}, err
	}
	projectID := script.ProjectID
	return resource.Mutation{
		Result: vo,
		Changes: []resource.Change{{
			EventID:       eventID,
			ResourceType:  resource.TypeScript,
			ResourceID:    script.ID,
			ProjectID:     &projectID,
			Version:       nullInt32Value(afterRow.Version),
			Deleted:       false,
			ChangedFields: changed,
			Before:        before,
			After:         vo,
		}},
		PreviousValues: before,
	}, nil
}

type ScriptSource struct{}

func (ScriptSource) ResourceType() resource.Type   { return resource.TypeScript }
func (ScriptSource) Operation() resource.Operation { return resource.OpReplaceScriptSource }

func (ScriptSource) Authorize(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64, previous *resource.Receipt) (*int64, error) {
	source, ok := cmd.(resource.SourceCommand)
	if !ok {
		return nil, apperr.BadRequest("命令类型无效")
	}
	return authorizeScriptOwner(ctx, q, source.ResourceID, actorUserID)
}

func (ScriptSource) Apply(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64) (resource.Mutation, error) {
	source := cmd.(resource.SourceCommand)
	script, err := q.LockScriptByID(ctx, source.ResourceID)
	if err != nil {
		if err == sql.ErrNoRows {
			return resource.Mutation{}, apperr.NotFound("剧本不存在")
		}
		return resource.Mutation{}, err
	}
	if nullInt32Value(script.Version) != source.Version {
		return resource.Mutation{}, apperr.Conflict("剧本版本已变化，请刷新后重试")
	}
	before := resource.ScriptVO(script)

	if err := q.SoftDeleteScenesByScript(ctx, script.ID); err != nil {
		return resource.Mutation{}, err
	}
	if err := q.SoftDeleteEpisodesByScript(ctx, script.ID); err != nil {
		return resource.Mutation{}, err
	}

	updated, err := q.UpdateScript(ctx, sqlcdb.UpdateScriptParams{
		Content:          sql.NullString{},
		RawContent:       sql.NullString{String: source.RawContent, Valid: true},
		StorySynopsis:    sql.NullString{},
		CharactersJson:   json.RawMessage("[]"),
		SourceType:       script.SourceType,
		ParsingStatus:    sql.NullInt32{Int32: 0, Valid: true},
		ParsingProgress:  sql.NullString{},
		Summary:          sql.NullString{},
		Genre:            sql.NullString{},
		TargetAudience:   sql.NullString{},
		DurationEstimate: sql.NullInt32{},
		Status:           script.Status,
		TotalEpisodes:    sql.NullInt32{Int32: 0, Valid: true},
		ID:               script.ID,
		Version:          script.Version,
	})
	if err != nil {
		return resource.Mutation{}, err
	}
	if updated != 1 {
		return resource.Mutation{}, apperr.Conflict("剧本版本已变化，请刷新后重试")
	}
	afterRow, err := q.GetScriptByID(ctx, script.ID)
	if err != nil {
		return resource.Mutation{}, err
	}
	vo := resource.ScriptVO(afterRow)
	eventID, err := resource.NewEventID()
	if err != nil {
		return resource.Mutation{}, err
	}
	projectID := script.ProjectID
	return resource.Mutation{
		Result: vo,
		Changes: []resource.Change{{
			EventID:       eventID,
			ResourceType:  resource.TypeScript,
			ResourceID:    script.ID,
			ProjectID:     &projectID,
			Version:       nullInt32Value(afterRow.Version),
			Deleted:       false,
			ChangedFields: []string{"rawContent", "content", "storySynopsis", "charactersJson", "parsingStatus", "totalEpisodes"},
			Before:        before,
			After:         vo,
		}},
		PreviousValues: before,
	}, nil
}
