package handlers

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/apperr"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/resource"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/sqlcdb"
)

type ScriptSceneCreate struct{}

func (ScriptSceneCreate) ResourceType() resource.Type   { return resource.TypeScriptScene }
func (ScriptSceneCreate) Operation() resource.Operation { return resource.OpCreate }

func (ScriptSceneCreate) Authorize(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64, previous *resource.Receipt) (*int64, error) {
	create, ok := cmd.(resource.CreateCommand)
	if !ok {
		return nil, apperr.BadRequest("命令类型无效")
	}
	episodeID, err := requiredPositiveID(create.Fields["episodeId"], "episodeId")
	if err != nil {
		return nil, err
	}
	ep, err := q.GetScriptEpisodeByID(ctx, episodeID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.NotFound("分集不存在")
		}
		return nil, err
	}
	return authorizeScriptOwner(ctx, q, ep.ScriptID, actorUserID)
}

func (ScriptSceneCreate) Apply(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64) (resource.Mutation, error) {
	create := cmd.(resource.CreateCommand)
	episodeID, err := requiredPositiveID(create.Fields["episodeId"], "episodeId")
	if err != nil {
		return resource.Mutation{}, err
	}
	parentVersion, err := requiredNonNegInt(create.Fields["parentVersion"], "parentVersion")
	if err != nil {
		return resource.Mutation{}, err
	}
	ep, err := q.LockScriptEpisodeByID(ctx, episodeID)
	if err != nil {
		if err == sql.ErrNoRows {
			return resource.Mutation{}, apperr.NotFound("分集不存在")
		}
		return resource.Mutation{}, err
	}
	if nullInt32Value(ep.Version) != parentVersion {
		return resource.Mutation{}, apperr.Conflict("分集版本已变化，请刷新后重试")
	}

	existing, err := q.CountScenesByEpisode(ctx, episodeID)
	if err != nil {
		return resource.Mutation{}, err
	}
	sortOrder := int32(existing)
	if v, ok := create.Fields["sortOrder"]; ok {
		sortOrder, err = requiredNonNegInt(v, "sortOrder")
		if err != nil {
			return resource.Mutation{}, err
		}
		if int64(sortOrder) > existing {
			return resource.Mutation{}, apperr.BadRequest("插入位置超出当前场次列表")
		}
	}

	sceneNumber := optionalNullString(create.Fields["sceneNumber"])
	if !sceneNumber.Valid {
		epNum := nullInt32Value(ep.EpisodeNumber)
		sceneNumber = sql.NullString{
			String: fmt.Sprintf("%d-%d", epNum, sortOrder+1),
			Valid:  true,
		}
	}

	status := int32(0)
	if v, ok := create.Fields["status"]; ok {
		status, err = requiredNonNegInt(v, "status")
		if err != nil {
			return resource.Mutation{}, err
		}
	}

	id, err := q.CreateScriptScene(ctx, sqlcdb.CreateScriptSceneParams{
		EpisodeID:         episodeID,
		ScriptID:          ep.ScriptID,
		SceneNumber:       sceneNumber,
		SceneHeading:      optionalNullString(create.Fields["sceneHeading"]),
		Location:          optionalNullString(create.Fields["location"]),
		TimeOfDay:         optionalNullString(create.Fields["timeOfDay"]),
		IntExt:            optionalNullString(create.Fields["intExt"]),
		Characters:        optionalJSONArray(create.Fields["characters"]),
		CharacterAssetIds: optionalJSONArray(create.Fields["characterAssetIds"]),
		SceneAssetID:      optionalNullInt64(create.Fields["sceneAssetId"]),
		PropAssetIds:      optionalJSONArray(create.Fields["propAssetIds"]),
		SceneDescription:  optionalNullString(create.Fields["sceneDescription"]),
		Dialogues:         optionalJSONArray(create.Fields["dialogues"]),
		SortOrder:         sql.NullInt32{Int32: sortOrder, Valid: true},
		Status:            sql.NullInt32{Int32: status, Valid: true},
	})
	if err != nil {
		return resource.Mutation{}, err
	}

	count, err := q.CountScenesByEpisode(ctx, episodeID)
	if err != nil {
		return resource.Mutation{}, err
	}
	bumped, err := q.BumpEpisodeSceneCount(ctx, sqlcdb.BumpEpisodeSceneCountParams{
		TotalScenes: sql.NullInt32{Int32: int32(count), Valid: true},
		ID:          episodeID,
		Version:     ep.Version,
	})
	if err != nil {
		return resource.Mutation{}, err
	}
	if bumped != 1 {
		return resource.Mutation{}, apperr.Conflict("分集版本已变化，请刷新后重试")
	}

	script, err := q.LockScriptByID(ctx, ep.ScriptID)
	if err != nil {
		return resource.Mutation{}, err
	}
	if _, err := q.BumpScriptVersionOnly(ctx, sqlcdb.BumpScriptVersionOnlyParams{
		ID:      script.ID,
		Version: script.Version,
	}); err != nil {
		return resource.Mutation{}, err
	}

	row, err := q.GetScriptSceneByID(ctx, id)
	if err != nil {
		return resource.Mutation{}, err
	}
	vo := resource.SceneVO(row)
	eventID, err := resource.NewEventID()
	if err != nil {
		return resource.Mutation{}, err
	}
	projectID := script.ProjectID
	return resource.Mutation{
		Result: vo,
		Changes: []resource.Change{{
			EventID:       eventID,
			ResourceType:  resource.TypeScriptScene,
			ResourceID:    id,
			ProjectID:     &projectID,
			Version:       0,
			Deleted:       false,
			ChangedFields: []string{"*"},
			Before:        nil,
			After:         vo,
		}},
		PreviousValues: map[string]any{"episodeVersion": parentVersion},
	}, nil
}

type ScriptScenePatch struct{}

func (ScriptScenePatch) ResourceType() resource.Type   { return resource.TypeScriptScene }
func (ScriptScenePatch) Operation() resource.Operation { return resource.OpPatch }

func (ScriptScenePatch) Authorize(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64, previous *resource.Receipt) (*int64, error) {
	patch, ok := cmd.(resource.PatchCommand)
	if !ok {
		return nil, apperr.BadRequest("命令类型无效")
	}
	scene, err := q.GetScriptSceneByID(ctx, patch.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.NotFound("场次不存在")
		}
		return nil, err
	}
	return authorizeScriptOwner(ctx, q, scene.ScriptID, actorUserID)
}

func (ScriptScenePatch) Apply(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64) (resource.Mutation, error) {
	patch := cmd.(resource.PatchCommand)
	if len(patch.Patch) == 0 {
		return resource.Mutation{}, apperr.BadRequest("修改字段不能为空")
	}
	scene, err := q.LockScriptSceneByID(ctx, patch.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return resource.Mutation{}, apperr.NotFound("场次不存在")
		}
		return resource.Mutation{}, err
	}
	if nullInt32Value(scene.Version) != patch.Version {
		return resource.Mutation{}, apperr.Conflict("场次版本已变化，请刷新后重试")
	}
	before := resource.SceneVO(scene)

	sceneNumber := scene.SceneNumber
	sceneHeading := scene.SceneHeading
	location := scene.Location
	timeOfDay := scene.TimeOfDay
	intExt := scene.IntExt
	characters := scene.Characters
	characterAssetIds := scene.CharacterAssetIds
	sceneAssetID := scene.SceneAssetID
	propAssetIds := scene.PropAssetIds
	sceneDescription := scene.SceneDescription
	dialogues := scene.Dialogues
	sortOrder := scene.SortOrder
	status := scene.Status
	changed := make([]string, 0, len(patch.Patch))

	for key, value := range patch.Patch {
		changed = append(changed, key)
		switch key {
		case "sceneNumber":
			sceneNumber = optionalNullString(value)
		case "sceneHeading":
			sceneHeading = optionalNullString(value)
		case "location":
			location = optionalNullString(value)
		case "timeOfDay":
			timeOfDay = optionalNullString(value)
		case "intExt":
			intExt = optionalNullString(value)
		case "characters":
			characters = optionalJSONArray(value)
		case "characterAssetIds":
			characterAssetIds = optionalJSONArray(value)
		case "sceneAssetId":
			sceneAssetID = optionalNullInt64(value)
		case "propAssetIds":
			propAssetIds = optionalJSONArray(value)
		case "sceneDescription":
			sceneDescription = optionalNullString(value)
		case "dialogues":
			dialogues = optionalJSONArray(value)
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

	updated, err := q.UpdateScriptScene(ctx, sqlcdb.UpdateScriptSceneParams{
		SceneNumber:       sceneNumber,
		SceneHeading:      sceneHeading,
		Location:          location,
		TimeOfDay:         timeOfDay,
		IntExt:            intExt,
		Characters:        characters,
		CharacterAssetIds: characterAssetIds,
		SceneAssetID:      sceneAssetID,
		PropAssetIds:      propAssetIds,
		SceneDescription:  sceneDescription,
		Dialogues:         dialogues,
		SortOrder:         sortOrder,
		Status:            status,
		ID:                scene.ID,
		Version:           scene.Version,
	})
	if err != nil {
		return resource.Mutation{}, err
	}
	if updated != 1 {
		return resource.Mutation{}, apperr.Conflict("场次版本已变化，请刷新后重试")
	}

	ep, err := q.LockScriptEpisodeByID(ctx, scene.EpisodeID)
	if err != nil {
		return resource.Mutation{}, err
	}
	if _, err := q.BumpEpisodeSceneCount(ctx, sqlcdb.BumpEpisodeSceneCountParams{
		TotalScenes: ep.TotalScenes,
		ID:          ep.ID,
		Version:     ep.Version,
	}); err != nil {
		return resource.Mutation{}, err
	}
	script, err := q.LockScriptByID(ctx, scene.ScriptID)
	if err != nil {
		return resource.Mutation{}, err
	}
	if _, err := q.BumpScriptVersionOnly(ctx, sqlcdb.BumpScriptVersionOnlyParams{
		ID:      script.ID,
		Version: script.Version,
	}); err != nil {
		return resource.Mutation{}, err
	}

	afterRow, err := q.GetScriptSceneByID(ctx, scene.ID)
	if err != nil {
		return resource.Mutation{}, err
	}
	vo := resource.SceneVO(afterRow)
	eventID, err := resource.NewEventID()
	if err != nil {
		return resource.Mutation{}, err
	}
	projectID := script.ProjectID
	return resource.Mutation{
		Result: vo,
		Changes: []resource.Change{{
			EventID:       eventID,
			ResourceType:  resource.TypeScriptScene,
			ResourceID:    scene.ID,
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

type ScriptSceneDelete struct{}

func (ScriptSceneDelete) ResourceType() resource.Type   { return resource.TypeScriptScene }
func (ScriptSceneDelete) Operation() resource.Operation { return resource.OpDelete }

func (ScriptSceneDelete) Authorize(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64, previous *resource.Receipt) (*int64, error) {
	del, ok := cmd.(resource.DeleteCommand)
	if !ok {
		return nil, apperr.BadRequest("命令类型无效")
	}
	scene, err := q.GetScriptSceneByID(ctx, del.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			if previous != nil {
				return nil, nil
			}
			return nil, apperr.NotFound("场次不存在")
		}
		return nil, err
	}
	return authorizeScriptOwner(ctx, q, scene.ScriptID, actorUserID)
}

func (ScriptSceneDelete) Apply(ctx context.Context, q *sqlcdb.Queries, cmd resource.Command, actorUserID int64) (resource.Mutation, error) {
	del := cmd.(resource.DeleteCommand)
	scene, err := q.LockScriptSceneByID(ctx, del.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return resource.Mutation{}, apperr.NotFound("场次不存在")
		}
		return resource.Mutation{}, err
	}
	if nullInt32Value(scene.Version) != del.Version {
		return resource.Mutation{}, apperr.Conflict("场次版本已变化，请刷新后重试")
	}
	before := resource.SceneVO(scene)
	updated, err := q.SoftDeleteScriptScene(ctx, sqlcdb.SoftDeleteScriptSceneParams{
		ID:      scene.ID,
		Version: scene.Version,
	})
	if err != nil {
		return resource.Mutation{}, err
	}
	if updated != 1 {
		return resource.Mutation{}, apperr.Conflict("场次版本已变化，请刷新后重试")
	}

	ep, err := q.LockScriptEpisodeByID(ctx, scene.EpisodeID)
	if err != nil {
		return resource.Mutation{}, err
	}
	count, err := q.CountScenesByEpisode(ctx, scene.EpisodeID)
	if err != nil {
		return resource.Mutation{}, err
	}
	if _, err := q.BumpEpisodeSceneCount(ctx, sqlcdb.BumpEpisodeSceneCountParams{
		TotalScenes: sql.NullInt32{Int32: int32(count), Valid: true},
		ID:          ep.ID,
		Version:     ep.Version,
	}); err != nil {
		return resource.Mutation{}, err
	}
	script, err := q.LockScriptByID(ctx, scene.ScriptID)
	if err != nil {
		return resource.Mutation{}, err
	}
	if _, err := q.BumpScriptVersionOnly(ctx, sqlcdb.BumpScriptVersionOnlyParams{
		ID:      script.ID,
		Version: script.Version,
	}); err != nil {
		return resource.Mutation{}, err
	}

	after := cloneMap(before)
	after["deleted"] = true
	after["version"] = nullInt32Value(scene.Version) + 1
	eventID, err := resource.NewEventID()
	if err != nil {
		return resource.Mutation{}, err
	}
	projectID := script.ProjectID
	return resource.Mutation{
		Result: true,
		Changes: []resource.Change{{
			EventID:       eventID,
			ResourceType:  resource.TypeScriptScene,
			ResourceID:    scene.ID,
			ProjectID:     &projectID,
			Version:       nullInt32Value(scene.Version) + 1,
			Deleted:       true,
			ChangedFields: []string{"deleted"},
			Before:        before,
			After:         after,
		}},
		PreviousValues: before,
	}, nil
}
