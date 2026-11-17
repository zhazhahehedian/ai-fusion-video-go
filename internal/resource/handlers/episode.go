package handlers

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"strconv"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/apperr"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/resource"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/sqlcdb"
)

type ScriptEpisodeCreate struct{}

func (ScriptEpisodeCreate) ResourceType() resource.Type   { return resource.TypeScriptEpisode }
func (ScriptEpisodeCreate) Operation() resource.Operation { return resource.OpCreate }

func (ScriptEpisodeCreate) Authorize(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64, previous *resource.Receipt) (*int64, error) {
	create, ok := cmd.(resource.CreateCommand)
	if !ok {
		return nil, apperr.BadRequest("命令类型无效")
	}
	scriptID, err := requiredPositiveID(create.Fields["scriptId"], "scriptId")
	if err != nil {
		return nil, err
	}
	return authorizeScriptOwner(ctx, q, scriptID, actorUserID)
}

func (ScriptEpisodeCreate) Apply(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64) (resource.Mutation, error) {
	create := cmd.(resource.CreateCommand)
	scriptID, err := requiredPositiveID(create.Fields["scriptId"], "scriptId")
	if err != nil {
		return resource.Mutation{}, err
	}
	parentVersion, err := requiredNonNegInt(create.Fields["parentVersion"], "parentVersion")
	if err != nil {
		return resource.Mutation{}, err
	}
	script, err := q.LockScriptByID(ctx, scriptID)
	if err != nil {
		if err == sql.ErrNoRows {
			return resource.Mutation{}, apperr.NotFound("剧本不存在")
		}
		return resource.Mutation{}, err
	}
	if nullInt32Value(script.Version) != parentVersion {
		return resource.Mutation{}, apperr.Conflict("剧本版本已变化，请刷新后重试")
	}

	episodeNumber, err := requiredPositiveInt(create.Fields["episodeNumber"], "episodeNumber")
	if err != nil {
		return resource.Mutation{}, err
	}
	if _, err := q.FindEpisodeNumberConflict(ctx, sqlcdb.FindEpisodeNumberConflictParams{
		ScriptID:      scriptID,
		EpisodeNumber: sql.NullInt32{Int32: episodeNumber, Valid: true},
		ID:            0,
	}); err == nil {
		return resource.Mutation{}, apperr.Conflict("集号已存在")
	} else if err != sql.ErrNoRows {
		return resource.Mutation{}, err
	}

	sortOrder := episodeNumber
	if v, ok := create.Fields["sortOrder"]; ok {
		sortOrder, err = requiredNonNegInt(v, "sortOrder")
		if err != nil {
			return resource.Mutation{}, err
		}
	}
	sourceType := int32(1)
	if v, ok := create.Fields["sourceType"]; ok {
		sourceType, err = requiredNonNegInt(v, "sourceType")
		if err != nil {
			return resource.Mutation{}, err
		}
	}
	status := int32(0)
	if v, ok := create.Fields["status"]; ok {
		status, err = requiredNonNegInt(v, "status")
		if err != nil {
			return resource.Mutation{}, err
		}
	}

	title, err := optionalNullString(create.Fields["title"])
	if err != nil {
		return resource.Mutation{}, err
	}
	synopsis, err := optionalNullString(create.Fields["synopsis"])
	if err != nil {
		return resource.Mutation{}, err
	}
	rawContent, err := optionalNullString(create.Fields["rawContent"])
	if err != nil {
		return resource.Mutation{}, err
	}
	durationEstimate, err := optionalNullInt32(create.Fields["durationEstimate"])
	if err != nil {
		return resource.Mutation{}, err
	}
	id, err := q.CreateScriptEpisode(ctx, sqlcdb.CreateScriptEpisodeParams{
		ScriptID:         scriptID,
		EpisodeNumber:    sql.NullInt32{Int32: episodeNumber, Valid: true},
		Title:            title,
		Synopsis:         synopsis,
		RawContent:       rawContent,
		DurationEstimate: durationEstimate,
		SourceType:       sql.NullInt32{Int32: sourceType, Valid: true},
		SortOrder:        sql.NullInt32{Int32: sortOrder, Valid: true},
		Status:           sql.NullInt32{Int32: status, Valid: true},
	})
	if err != nil {
		return resource.Mutation{}, err
	}

	count, err := q.CountEpisodesByScript(ctx, scriptID)
	if err != nil {
		return resource.Mutation{}, err
	}
	updated, err := q.BumpScriptVersionAndEpisodeCount(ctx, sqlcdb.BumpScriptVersionAndEpisodeCountParams{
		TotalEpisodes: sql.NullInt32{Int32: int32(count), Valid: true},
		ID:            scriptID,
		Version:       script.Version,
	})
	if err != nil {
		return resource.Mutation{}, err
	}
	if updated != 1 {
		return resource.Mutation{}, apperr.Conflict("剧本版本已变化，请刷新后重试")
	}

	ep, err := q.GetScriptEpisodeByID(ctx, id)
	if err != nil {
		return resource.Mutation{}, err
	}
	vo := resource.EpisodeVO(ep)
	eventID, err := resource.NewEventID()
	if err != nil {
		return resource.Mutation{}, err
	}
	projectID := script.ProjectID
	return resource.Mutation{
		Result: vo,
		Changes: []resource.Change{{
			EventID:       eventID,
			ResourceType:  resource.TypeScriptEpisode,
			ResourceID:    id,
			ProjectID:     &projectID,
			Version:       0,
			Deleted:       false,
			ChangedFields: []string{"*"},
			Before:        nil,
			After:         vo,
		}},
		PreviousValues: map[string]any{"scriptVersion": parentVersion},
	}, nil
}

type ScriptEpisodePatch struct{}

func (ScriptEpisodePatch) ResourceType() resource.Type   { return resource.TypeScriptEpisode }
func (ScriptEpisodePatch) Operation() resource.Operation { return resource.OpPatch }

func (ScriptEpisodePatch) Authorize(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64, previous *resource.Receipt) (*int64, error) {
	patch, ok := cmd.(resource.PatchCommand)
	if !ok {
		return nil, apperr.BadRequest("命令类型无效")
	}
	ep, err := q.GetScriptEpisodeByID(ctx, patch.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.NotFound("分集不存在")
		}
		return nil, err
	}
	return authorizeScriptOwner(ctx, q, ep.ScriptID, actorUserID)
}

func (ScriptEpisodePatch) Apply(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64) (resource.Mutation, error) {
	patch := cmd.(resource.PatchCommand)
	if len(patch.Patch) == 0 {
		return resource.Mutation{}, apperr.BadRequest("修改字段不能为空")
	}
	ep, err := q.LockScriptEpisodeByID(ctx, patch.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return resource.Mutation{}, apperr.NotFound("分集不存在")
		}
		return resource.Mutation{}, err
	}
	if nullInt32Value(ep.Version) != patch.Version {
		return resource.Mutation{}, apperr.Conflict("分集版本已变化，请刷新后重试")
	}
	before := resource.EpisodeVO(ep)

	episodeNumber := ep.EpisodeNumber
	title := ep.Title
	synopsis := ep.Synopsis
	rawContent := ep.RawContent
	durationEstimate := ep.DurationEstimate
	sourceType := ep.SourceType
	sortOrder := ep.SortOrder
	status := ep.Status
	changed := make([]string, 0, len(patch.Patch))

	for key, value := range patch.Patch {
		changed = append(changed, key)
		switch key {
		case "episodeNumber":
			n, err := requiredPositiveInt(value, "episodeNumber")
			if err != nil {
				return resource.Mutation{}, err
			}
			if _, err := q.FindEpisodeNumberConflict(ctx, sqlcdb.FindEpisodeNumberConflictParams{
				ScriptID:      ep.ScriptID,
				EpisodeNumber: sql.NullInt32{Int32: n, Valid: true},
				ID:            ep.ID,
			}); err == nil {
				return resource.Mutation{}, apperr.Conflict("集号已存在")
			} else if err != sql.ErrNoRows {
				return resource.Mutation{}, err
			}
			episodeNumber = sql.NullInt32{Int32: n, Valid: true}
		case "title":
			title, err = optionalNullString(value)
			if err != nil {
				return resource.Mutation{}, err
			}
		case "synopsis":
			synopsis, err = optionalNullString(value)
			if err != nil {
				return resource.Mutation{}, err
			}
		case "rawContent":
			rawContent, err = optionalNullString(value)
			if err != nil {
				return resource.Mutation{}, err
			}
		case "durationEstimate":
			durationEstimate, err = optionalNullInt32(value)
			if err != nil {
				return resource.Mutation{}, err
			}
		case "sourceType":
			n, err := requiredNonNegInt(value, "sourceType")
			if err != nil {
				return resource.Mutation{}, err
			}
			sourceType = sql.NullInt32{Int32: n, Valid: true}
		case "sortOrder":
			n, err := requiredNonNegInt(value, "sortOrder")
			if err != nil {
				return resource.Mutation{}, err
			}
			sortOrder = sql.NullInt32{Int32: n, Valid: true}
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

	updated, err := q.UpdateScriptEpisode(ctx, sqlcdb.UpdateScriptEpisodeParams{
		EpisodeNumber:    episodeNumber,
		Title:            title,
		Synopsis:         synopsis,
		RawContent:       rawContent,
		DurationEstimate: durationEstimate,
		SourceType:       sourceType,
		SortOrder:        sortOrder,
		Status:           status,
		ID:               ep.ID,
		Version:          ep.Version,
	})
	if err != nil {
		return resource.Mutation{}, err
	}
	if updated != 1 {
		return resource.Mutation{}, apperr.Conflict("分集版本已变化，请刷新后重试")
	}
	afterRow, err := q.GetScriptEpisodeByID(ctx, ep.ID)
	if err != nil {
		return resource.Mutation{}, err
	}
	vo := resource.EpisodeVO(afterRow)
	eventID, err := resource.NewEventID()
	if err != nil {
		return resource.Mutation{}, err
	}
	script, err := q.GetScriptByID(ctx, ep.ScriptID)
	if err != nil {
		return resource.Mutation{}, err
	}
	projectID := script.ProjectID
	return resource.Mutation{
		Result: vo,
		Changes: []resource.Change{{
			EventID:       eventID,
			ResourceType:  resource.TypeScriptEpisode,
			ResourceID:    ep.ID,
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

type ScriptEpisodeDelete struct{}

func (ScriptEpisodeDelete) ResourceType() resource.Type   { return resource.TypeScriptEpisode }
func (ScriptEpisodeDelete) Operation() resource.Operation { return resource.OpDelete }

func (ScriptEpisodeDelete) Authorize(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64, previous *resource.Receipt) (*int64, error) {
	del, ok := cmd.(resource.DeleteCommand)
	if !ok {
		return nil, apperr.BadRequest("命令类型无效")
	}
	ep, err := q.GetScriptEpisodeByID(ctx, del.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			if previous != nil {
				return nil, nil
			}
			return nil, apperr.NotFound("分集不存在")
		}
		return nil, err
	}
	return authorizeScriptOwner(ctx, q, ep.ScriptID, actorUserID)
}

func (ScriptEpisodeDelete) Apply(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64) (resource.Mutation, error) {
	del := cmd.(resource.DeleteCommand)
	ep, err := q.LockScriptEpisodeByID(ctx, del.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return resource.Mutation{}, apperr.NotFound("分集不存在")
		}
		return resource.Mutation{}, err
	}
	if nullInt32Value(ep.Version) != del.Version {
		return resource.Mutation{}, apperr.Conflict("分集版本已变化，请刷新后重试")
	}
	before := resource.EpisodeVO(ep)
	if err := q.SoftDeleteScenesByEpisode(ctx, ep.ID); err != nil {
		return resource.Mutation{}, err
	}
	updated, err := q.SoftDeleteScriptEpisode(ctx, sqlcdb.SoftDeleteScriptEpisodeParams{
		ID:      ep.ID,
		Version: ep.Version,
	})
	if err != nil {
		return resource.Mutation{}, err
	}
	if updated != 1 {
		return resource.Mutation{}, apperr.Conflict("分集版本已变化，请刷新后重试")
	}

	script, err := q.LockScriptByID(ctx, ep.ScriptID)
	if err != nil {
		return resource.Mutation{}, err
	}
	count, err := q.CountEpisodesByScript(ctx, ep.ScriptID)
	if err != nil {
		return resource.Mutation{}, err
	}
	bumped, err := q.BumpScriptVersionAndEpisodeCount(ctx, sqlcdb.BumpScriptVersionAndEpisodeCountParams{
		TotalEpisodes: sql.NullInt32{Int32: int32(count), Valid: true},
		ID:            script.ID,
		Version:       script.Version,
	})
	if err != nil {
		return resource.Mutation{}, err
	}
	if bumped != 1 {
		return resource.Mutation{}, apperr.Conflict("剧本版本已变化，请刷新后重试")
	}

	after := cloneMap(before)
	after["deleted"] = true
	after["version"] = nullInt32Value(ep.Version) + 1
	eventID, err := resource.NewEventID()
	if err != nil {
		return resource.Mutation{}, err
	}
	projectID := script.ProjectID
	return resource.Mutation{
		Result: true,
		Changes: []resource.Change{{
			EventID:       eventID,
			ResourceType:  resource.TypeScriptEpisode,
			ResourceID:    ep.ID,
			ProjectID:     &projectID,
			Version:       nullInt32Value(ep.Version) + 1,
			Deleted:       true,
			ChangedFields: []string{"deleted"},
			Before:        before,
			After:         after,
		}},
		PreviousValues: before,
	}, nil
}

func authorizeScriptOwner(ctx context.Context, q *sqlcdb.Queries, scriptID, actorUserID int64) (*int64, error) {
	script, err := q.LockScriptByID(ctx, scriptID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.NotFound("剧本不存在")
		}
		return nil, err
	}
	if _, err := q.LockOwnedProjectByID(ctx, sqlcdb.LockOwnedProjectByIDParams{
		ID:      script.ProjectID,
		OwnerID: actorUserID,
	}); err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.Forbidden("无权访问该项目")
		}
		return nil, err
	}
	projectID := script.ProjectID
	return &projectID, nil
}

func requiredPositiveID(v any, field string) (int64, error) {
	n, err := asInt64Value(v)
	if err != nil || n <= 0 {
		return 0, apperr.BadRequest(field + " 无效")
	}
	return n, nil
}

func requiredPositiveInt(v any, field string) (int32, error) {
	n, err := asInt64Value(v)
	if err != nil || n < 1 {
		return 0, apperr.BadRequest(field + " 无效")
	}
	if n > math.MaxInt32 {
		return 0, apperr.BadRequest(field + " 超出范围")
	}
	return int32(n), nil
}

func requiredNonNegInt(v any, field string) (int32, error) {
	n, err := asInt64Value(v)
	if err != nil || n < 0 {
		return 0, apperr.BadRequest(field + " 无效")
	}
	if n > math.MaxInt32 {
		return 0, apperr.BadRequest(field + " 超出范围")
	}
	return int32(n), nil
}

func asInt64Value(v any) (int64, error) {
	switch n := v.(type) {
	case int:
		return int64(n), nil
	case int32:
		return int64(n), nil
	case int64:
		return n, nil
	case float32:
		return integralInt64(float64(n))
	case float64:
		return integralInt64(n)
	case jsonNumber:
		out, err := n.Int64()
		if err != nil {
			if errors.Is(err, strconv.ErrRange) {
				return 0, apperr.BadRequest("数值超出范围")
			}
			return 0, apperr.BadRequest("数值必须是整数")
		}
		return out, nil
	default:
		return 0, apperr.BadRequest("数值无效")
	}
}

func integralInt64(n float64) (int64, error) {
	if math.IsNaN(n) || math.IsInf(n, 0) || n != math.Trunc(n) {
		return 0, apperr.BadRequest("数值必须是整数")
	}
	// 2^63 不能放进 int64。float64 无法精确表示 MaxInt64，用上界排除溢出。
	if n < float64(math.MinInt64) || n >= 9223372036854775808 {
		return 0, apperr.BadRequest("数值超出范围")
	}
	out := int64(n)
	if float64(out) != n {
		return 0, apperr.BadRequest("数值超出范围")
	}
	return out, nil
}

// jsonNumber avoids importing encoding/json Number in signature noise.
type jsonNumber interface {
	Int64() (int64, error)
}

func nullInt32Value(v sql.NullInt32) int32 {
	if !v.Valid {
		return 0
	}
	return v.Int32
}

func cloneMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
