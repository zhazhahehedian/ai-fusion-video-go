package resource

import (
	"context"
	"database/sql"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/apperr"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/sqlcdb"
)

// ScriptQuery provides read APIs for script / episodes / scenes (outside the command pipeline).
type ScriptQuery struct {
	db *sql.DB
}

func NewScriptQuery(db *sql.DB) *ScriptQuery {
	return &ScriptQuery{db: db}
}

func (s *ScriptQuery) ByProject(ctx context.Context, actorUserID, projectID int64) (map[string]any, error) {
	q := sqlcdb.New(s.db)
	if _, err := q.GetOwnedProjectByID(ctx, sqlcdb.GetOwnedProjectByIDParams{
		ID:      projectID,
		OwnerID: actorUserID,
	}); err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.Forbidden("无权访问该项目")
		}
		return nil, err
	}
	script, err := q.GetScriptByProjectID(ctx, projectID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.NotFound("项目尚未初始化剧本")
		}
		return nil, err
	}
	return ScriptVO(script), nil
}

func (s *ScriptQuery) Get(ctx context.Context, actorUserID, scriptID int64) (map[string]any, error) {
	q := sqlcdb.New(s.db)
	script, err := q.GetScriptByID(ctx, scriptID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.NotFound("剧本不存在")
		}
		return nil, err
	}
	if _, err := q.GetOwnedProjectByID(ctx, sqlcdb.GetOwnedProjectByIDParams{
		ID:      script.ProjectID,
		OwnerID: actorUserID,
	}); err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.Forbidden("无权访问该项目")
		}
		return nil, err
	}
	return ScriptVO(script), nil
}

func (s *ScriptQuery) ListEpisodes(ctx context.Context, actorUserID, scriptID int64) ([]map[string]any, error) {
	q := sqlcdb.New(s.db)
	if _, err := requireScriptOwner(ctx, q, scriptID, actorUserID); err != nil {
		return nil, err
	}
	rows, err := q.ListEpisodesByScript(ctx, scriptID)
	if err != nil {
		return nil, err
	}
	list := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		list = append(list, EpisodeVO(row))
	}
	return list, nil
}

func (s *ScriptQuery) GetEpisode(ctx context.Context, actorUserID, episodeID int64) (map[string]any, error) {
	q := sqlcdb.New(s.db)
	ep, err := q.GetScriptEpisodeByID(ctx, episodeID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.NotFound("分集不存在")
		}
		return nil, err
	}
	if _, err := requireScriptOwner(ctx, q, ep.ScriptID, actorUserID); err != nil {
		return nil, err
	}
	return EpisodeVO(ep), nil
}

func (s *ScriptQuery) ListScenes(ctx context.Context, actorUserID, episodeID int64) ([]map[string]any, error) {
	q := sqlcdb.New(s.db)
	ep, err := q.GetScriptEpisodeByID(ctx, episodeID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.NotFound("分集不存在")
		}
		return nil, err
	}
	if _, err := requireScriptOwner(ctx, q, ep.ScriptID, actorUserID); err != nil {
		return nil, err
	}
	rows, err := q.ListScenesByEpisode(ctx, episodeID)
	if err != nil {
		return nil, err
	}
	list := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		list = append(list, SceneVO(row))
	}
	return list, nil
}

func (s *ScriptQuery) GetScene(ctx context.Context, actorUserID, sceneID int64) (map[string]any, error) {
	q := sqlcdb.New(s.db)
	scene, err := q.GetScriptSceneByID(ctx, sceneID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.NotFound("场次不存在")
		}
		return nil, err
	}
	if _, err := requireScriptOwner(ctx, q, scene.ScriptID, actorUserID); err != nil {
		return nil, err
	}
	return SceneVO(scene), nil
}

func requireScriptOwner(ctx context.Context, q *sqlcdb.Queries, scriptID, actorUserID int64) (int64, error) {
	script, err := q.GetScriptByID(ctx, scriptID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, apperr.NotFound("剧本不存在")
		}
		return 0, err
	}
	if _, err := q.GetOwnedProjectByID(ctx, sqlcdb.GetOwnedProjectByIDParams{
		ID:      script.ProjectID,
		OwnerID: actorUserID,
	}); err != nil {
		if err == sql.ErrNoRows {
			return 0, apperr.Forbidden("无权访问该项目")
		}
		return 0, err
	}
	return script.ProjectID, nil
}
