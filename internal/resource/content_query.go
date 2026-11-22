package resource

import (
	"context"
	"database/sql"
	"strings"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/apperr"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/sqlcdb"
)

// ContentQuery 读取分镜、资产和图片/视频任务。写路径走命令管道。
type ContentQuery struct {
	db *sql.DB
}

func NewContentQuery(db *sql.DB) *ContentQuery {
	return &ContentQuery{db: db}
}

func (s *ContentQuery) Storyboard(ctx context.Context, actorUserID, id int64) (map[string]any, error) {
	q := sqlcdb.New(s.db)
	row, err := q.GetStoryboardByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.NotFound("分镜不存在")
		}
		return nil, err
	}
	if err := s.ownProject(ctx, q, row.ProjectID, actorUserID); err != nil {
		return nil, err
	}
	return StoryboardVO(row), nil
}

func (s *ContentQuery) StoryboardByProject(ctx context.Context, actorUserID, projectID int64) (map[string]any, error) {
	q := sqlcdb.New(s.db)
	if err := s.ownProject(ctx, q, sql.NullInt64{Int64: projectID, Valid: projectID > 0}, actorUserID); err != nil {
		return nil, err
	}
	row, err := q.GetLatestStoryboardByProject(ctx, sql.NullInt64{Int64: projectID, Valid: true})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.NotFound("项目尚未创建分镜")
		}
		return nil, err
	}
	return StoryboardVO(row), nil
}

func (s *ContentQuery) StoryboardEpisodes(ctx context.Context, actorUserID, storyboardID int64) ([]map[string]any, error) {
	if _, err := s.Storyboard(ctx, actorUserID, storyboardID); err != nil {
		return nil, err
	}
	rows, err := sqlcdb.New(s.db).ListStoryboardEpisodes(ctx, storyboardID)
	if err != nil {
		return nil, err
	}
	return mapRows(rows, StoryboardEpisodeVO), nil
}

func (s *ContentQuery) StoryboardEpisode(ctx context.Context, actorUserID, id int64) (map[string]any, error) {
	q := sqlcdb.New(s.db)
	row, err := q.GetStoryboardEpisodeByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.NotFound("分镜集不存在")
		}
		return nil, err
	}
	if _, err := s.Storyboard(ctx, actorUserID, row.StoryboardID); err != nil {
		return nil, err
	}
	return StoryboardEpisodeVO(row), nil
}

func (s *ContentQuery) StoryboardScenesByEpisode(ctx context.Context, actorUserID, episodeID int64) ([]map[string]any, error) {
	if _, err := s.StoryboardEpisode(ctx, actorUserID, episodeID); err != nil {
		return nil, err
	}
	rows, err := sqlcdb.New(s.db).ListStoryboardScenesByEpisode(ctx, episodeID)
	if err != nil {
		return nil, err
	}
	return mapRows(rows, StoryboardSceneVO), nil
}

func (s *ContentQuery) StoryboardScenesByBoard(ctx context.Context, actorUserID, storyboardID int64) ([]map[string]any, error) {
	if _, err := s.Storyboard(ctx, actorUserID, storyboardID); err != nil {
		return nil, err
	}
	rows, err := sqlcdb.New(s.db).ListStoryboardScenesByBoard(ctx, storyboardID)
	if err != nil {
		return nil, err
	}
	return mapRows(rows, StoryboardSceneVO), nil
}

func (s *ContentQuery) StoryboardScene(ctx context.Context, actorUserID, id int64) (map[string]any, error) {
	q := sqlcdb.New(s.db)
	row, err := q.GetStoryboardSceneByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.NotFound("分镜场次不存在")
		}
		return nil, err
	}
	if _, err := s.Storyboard(ctx, actorUserID, row.StoryboardID); err != nil {
		return nil, err
	}
	return StoryboardSceneVO(row), nil
}

func (s *ContentQuery) StoryboardItemsByBoard(ctx context.Context, actorUserID, storyboardID int64) ([]map[string]any, error) {
	if _, err := s.Storyboard(ctx, actorUserID, storyboardID); err != nil {
		return nil, err
	}
	rows, err := sqlcdb.New(s.db).ListStoryboardItemsByBoard(ctx, storyboardID)
	if err != nil {
		return nil, err
	}
	return mapRows(rows, StoryboardItemVO), nil
}

func (s *ContentQuery) StoryboardItemsByScene(ctx context.Context, actorUserID, sceneID int64) ([]map[string]any, error) {
	if _, err := s.StoryboardScene(ctx, actorUserID, sceneID); err != nil {
		return nil, err
	}
	rows, err := sqlcdb.New(s.db).ListStoryboardItemsByScene(ctx, sql.NullInt64{Int64: sceneID, Valid: true})
	if err != nil {
		return nil, err
	}
	return mapRows(rows, StoryboardItemVO), nil
}

func (s *ContentQuery) StoryboardItem(ctx context.Context, actorUserID, id int64) (map[string]any, error) {
	q := sqlcdb.New(s.db)
	row, err := q.GetStoryboardItemByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.NotFound("镜头不存在")
		}
		return nil, err
	}
	if _, err := s.Storyboard(ctx, actorUserID, row.StoryboardID); err != nil {
		return nil, err
	}
	return StoryboardItemVO(row), nil
}

func (s *ContentQuery) Asset(ctx context.Context, actorUserID, id int64) (map[string]any, error) {
	q := sqlcdb.New(s.db)
	row, err := q.GetAssetByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.NotFound("资产不存在")
		}
		return nil, err
	}
	if err := s.ownProject(ctx, q, row.ProjectID, actorUserID); err != nil {
		return nil, err
	}
	return AssetVO(row), nil
}

func (s *ContentQuery) Assets(ctx context.Context, actorUserID int64, projectID *int64, assetType, keyword string, withItems bool) ([]map[string]any, error) {
	q := sqlcdb.New(s.db)
	var project sql.NullInt64
	if projectID != nil {
		if *projectID <= 0 {
			return nil, apperr.BadRequest("projectId 无效")
		}
		if err := s.ownProject(ctx, q, sql.NullInt64{Int64: *projectID, Valid: true}, actorUserID); err != nil {
			return nil, err
		}
		project = sql.NullInt64{Int64: *projectID, Valid: true}
	}
	rows, err := q.ListAssetsByOwner(ctx, sqlcdb.ListAssetsByOwnerParams{
		OwnerID:        actorUserID,
		ProjectID:      project,
		AssetType:      sql.NullString{String: assetType, Valid: true},
		Keyword:        keyword,
		KeywordPattern: likeContains(keyword),
	})
	if err != nil {
		return nil, err
	}
	list := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		vo := AssetVO(row)
		if withItems {
			items, err := s.assetItems(ctx, row.ID)
			if err != nil {
				return nil, err
			}
			vo["items"] = items
		}
		list = append(list, vo)
	}
	return list, nil
}

func (s *ContentQuery) AssetItem(ctx context.Context, actorUserID, id int64) (map[string]any, error) {
	q := sqlcdb.New(s.db)
	row, err := q.GetAssetItemByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.NotFound("子资产不存在")
		}
		return nil, err
	}
	if _, err := s.Asset(ctx, actorUserID, row.AssetID); err != nil {
		return nil, err
	}
	return AssetItemVO(row), nil
}

func (s *ContentQuery) AssetItems(ctx context.Context, actorUserID, assetID int64) ([]map[string]any, error) {
	if _, err := s.Asset(ctx, actorUserID, assetID); err != nil {
		return nil, err
	}
	return s.assetItems(ctx, assetID)
}

func (s *ContentQuery) assetItems(ctx context.Context, assetID int64) ([]map[string]any, error) {
	rows, err := sqlcdb.New(s.db).ListAssetItems(ctx, assetID)
	if err != nil {
		return nil, err
	}
	return mapRows(rows, AssetItemVO), nil
}

func (s *ContentQuery) ImageTask(ctx context.Context, actorUserID int64, publicID string) (map[string]any, error) {
	q := sqlcdb.New(s.db)
	row, err := q.GetImageTaskByPublicID(ctx, publicID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.NotFound("生图任务不存在")
		}
		return nil, err
	}
	if err := s.canReadTask(ctx, q, row.UserID, row.ProjectID, actorUserID); err != nil {
		return nil, err
	}
	version, err := maxVersion(ctx, q, TypeImageTask, row.ID)
	if err != nil {
		return nil, err
	}
	return ImageTaskVO(row, version), nil
}

func (s *ContentQuery) ImageTaskPage(ctx context.Context, actorUserID int64, pageNo, pageSize int) (map[string]any, error) {
	pageNo, pageSize, err := normalizePage(pageNo, pageSize)
	if err != nil {
		return nil, err
	}
	q := sqlcdb.New(s.db)
	total, err := q.CountImageTasksByUser(ctx, actorUserID)
	if err != nil {
		return nil, err
	}
	rows, err := q.PageImageTasksByUser(ctx, sqlcdb.PageImageTasksByUserParams{
		UserID: actorUserID,
		Limit:  int32(pageSize),
		Offset: int32((pageNo - 1) * pageSize),
	})
	if err != nil {
		return nil, err
	}
	return taskPage(ctx, q, TypeImageTask, rows, total, func(row sqlcdb.AfvImageTask) int64 { return row.ID }, func(row sqlcdb.AfvImageTask, version int32) map[string]any {
		return ImageTaskVO(row, version)
	})
}

func (s *ContentQuery) ImageItems(ctx context.Context, actorUserID, taskID int64) ([]map[string]any, error) {
	q := sqlcdb.New(s.db)
	task, err := q.GetImageTaskByID(ctx, taskID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.NotFound("生图任务不存在")
		}
		return nil, err
	}
	if err := s.canReadTask(ctx, q, task.UserID, task.ProjectID, actorUserID); err != nil {
		return nil, err
	}
	rows, err := q.ListImageItemsByTask(ctx, taskID)
	if err != nil {
		return nil, err
	}
	return mapRows(rows, ImageItemVO), nil
}

func (s *ContentQuery) VideoTask(ctx context.Context, actorUserID int64, publicID string) (map[string]any, error) {
	q := sqlcdb.New(s.db)
	row, err := q.GetVideoTaskByPublicID(ctx, publicID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.NotFound("生视频任务不存在")
		}
		return nil, err
	}
	if err := s.canReadTask(ctx, q, row.UserID, row.ProjectID, actorUserID); err != nil {
		return nil, err
	}
	version, err := maxVersion(ctx, q, TypeVideoTask, row.ID)
	if err != nil {
		return nil, err
	}
	return VideoTaskVO(row, version), nil
}

func (s *ContentQuery) VideoTaskPage(ctx context.Context, actorUserID int64, pageNo, pageSize int) (map[string]any, error) {
	pageNo, pageSize, err := normalizePage(pageNo, pageSize)
	if err != nil {
		return nil, err
	}
	q := sqlcdb.New(s.db)
	total, err := q.CountVideoTasksByUser(ctx, actorUserID)
	if err != nil {
		return nil, err
	}
	rows, err := q.PageVideoTasksByUser(ctx, sqlcdb.PageVideoTasksByUserParams{
		UserID: actorUserID,
		Limit:  int32(pageSize),
		Offset: int32((pageNo - 1) * pageSize),
	})
	if err != nil {
		return nil, err
	}
	return taskPage(ctx, q, TypeVideoTask, rows, total, func(row sqlcdb.AfvVideoTask) int64 { return row.ID }, func(row sqlcdb.AfvVideoTask, version int32) map[string]any {
		return VideoTaskVO(row, version)
	})
}

func (s *ContentQuery) DeleteImageTask(ctx context.Context, actorUserID, id int64) error {
	return s.deleteGenerationTask(ctx, actorUserID, id, true)
}

func (s *ContentQuery) DeleteVideoTask(ctx context.Context, actorUserID, id int64) error {
	return s.deleteGenerationTask(ctx, actorUserID, id, false)
}

func (s *ContentQuery) deleteGenerationTask(ctx context.Context, actorUserID, id int64, image bool) error {
	return withTx(ctx, s.db, func(tx *sql.Tx) error {
		q := sqlcdb.New(tx)
		if image {
			row, err := q.LockImageTaskByID(ctx, id)
			if err != nil {
				if err == sql.ErrNoRows {
					return apperr.NotFound("图片任务不存在")
				}
				return err
			}
			if row.UserID != actorUserID {
				return apperr.Forbidden("无权访问该任务")
			}
			if !DeletableTaskStatus(row.Status) {
				return apperr.Conflict("请等待图片生成结束后再删除记录")
			}
			if err := q.SoftDeleteImageItemsByTask(ctx, row.ID); err != nil {
				return err
			}
			n, err := q.SoftDeleteImageTask(ctx, row.ID)
			if err != nil {
				return err
			}
			if n != 1 {
				return apperr.NotFound("图片任务不存在")
			}
			return nil
		}
		row, err := q.LockVideoTaskByID(ctx, id)
		if err != nil {
			if err == sql.ErrNoRows {
				return apperr.NotFound("视频任务不存在")
			}
			return err
		}
		if row.UserID != actorUserID {
			return apperr.Forbidden("无权访问该任务")
		}
		if !DeletableTaskStatus(row.Status) {
			return apperr.Conflict("请等待视频生成结束后再删除记录")
		}
		if err := q.SoftDeleteVideoItemsByTask(ctx, row.ID); err != nil {
			return err
		}
		n, err := q.SoftDeleteVideoTask(ctx, row.ID)
		if err != nil {
			return err
		}
		if n != 1 {
			return apperr.NotFound("视频任务不存在")
		}
		return nil
	})
}

// DeletableTaskStatus 对齐 Java：只有已结束的生成记录可以删除。
func DeletableTaskStatus(status sql.NullInt32) bool {
	if !status.Valid {
		return false
	}
	switch status.Int32 {
	case 2, 3, 4, 6:
		return true
	default:
		return false
	}
}

func (s *ContentQuery) VideoItems(ctx context.Context, actorUserID, taskID int64) ([]map[string]any, error) {
	q := sqlcdb.New(s.db)
	task, err := q.GetVideoTaskByID(ctx, taskID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.NotFound("生视频任务不存在")
		}
		return nil, err
	}
	if err := s.canReadTask(ctx, q, task.UserID, task.ProjectID, actorUserID); err != nil {
		return nil, err
	}
	rows, err := q.ListVideoItemsByTask(ctx, taskID)
	if err != nil {
		return nil, err
	}
	return mapRows(rows, VideoItemVO), nil
}

func (s *ContentQuery) ownProject(ctx context.Context, q *sqlcdb.Queries, projectID sql.NullInt64, actorUserID int64) error {
	if !projectID.Valid || projectID.Int64 <= 0 {
		return apperr.Forbidden("无权访问该项目")
	}
	_, err := q.GetOwnedProjectByID(ctx, sqlcdb.GetOwnedProjectByIDParams{
		ID:      projectID.Int64,
		OwnerID: actorUserID,
	})
	if err == sql.ErrNoRows {
		return apperr.Forbidden("无权访问该项目")
	}
	return err
}

func (s *ContentQuery) canReadTask(ctx context.Context, q *sqlcdb.Queries, userID int64, projectID sql.NullInt64, actorUserID int64) error {
	if userID == actorUserID {
		return nil
	}
	if projectID.Valid {
		return s.ownProject(ctx, q, projectID, actorUserID)
	}
	return apperr.Forbidden("无权访问该任务")
}

func maxVersion(ctx context.Context, q *sqlcdb.Queries, resType Type, id int64) (int32, error) {
	n, err := q.MaxResourceVersion(ctx, sqlcdb.MaxResourceVersionParams{
		ResourceType: string(resType),
		ResourceID:   id,
	})
	if err != nil {
		return 0, err
	}
	if n < 0 || n > 1<<31-1 {
		return 0, apperr.Conflict("版本号超出范围")
	}
	return int32(n), nil
}

func taskPage[T any](ctx context.Context, q *sqlcdb.Queries, resType Type, rows []T, total int64, idOf func(T) int64, vo func(T, int32) map[string]any) (map[string]any, error) {
	versions := map[int64]int32{}
	if len(rows) > 0 {
		ids := make([]int64, 0, len(rows))
		for _, row := range rows {
			ids = append(ids, idOf(row))
		}
		found, err := q.ListResourceVersions(ctx, sqlcdb.ListResourceVersionsParams{
			ResourceType: string(resType),
			Ids:          ids,
		})
		if err != nil {
			return nil, err
		}
		for _, item := range found {
			if item.Version < 0 || item.Version > 1<<31-1 {
				return nil, apperr.Conflict("版本号超出范围")
			}
			versions[item.ResourceID] = int32(item.Version)
		}
	}
	list := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		list = append(list, vo(row, versions[idOf(row)]))
	}
	return map[string]any{"list": list, "total": total}, nil
}

func mapRows[T any](rows []T, vo func(T) map[string]any) []map[string]any {
	list := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		list = append(list, vo(row))
	}
	return list
}

func likeContains(keyword string) string {
	if keyword == "" {
		return ""
	}
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return "%" + replacer.Replace(keyword) + "%"
}

func normalizePage(pageNo, pageSize int) (int, int, error) {
	if pageNo < 1 {
		return 0, 0, apperr.BadRequest("pageNo 无效")
	}
	if pageSize < 1 {
		return 0, 0, apperr.BadRequest("pageSize 无效")
	}
	if pageSize > 100 {
		return 0, 0, apperr.BadRequest("pageSize 不能超过 100")
	}
	return pageNo, pageSize, nil
}
