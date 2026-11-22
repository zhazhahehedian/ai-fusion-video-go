package resource_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/apperr"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/resource"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/resource/handlers"

	_ "github.com/go-sql-driver/mysql"
)

func TestContentCommandPipeline(t *testing.T) {
	db := openContentMySQL(t)
	for _, path := range []string{
		"../../db/migrations/00001_sys_user_and_project.sql",
		"../../db/migrations/00002_resource_command_and_script.sql",
		"../../db/migrations/00003_script_scene.sql",
		"../../db/migrations/00004_generation_jobs.sql",
		"../../db/migrations/00005_media_job.sql",
		"../../db/migrations/00006_storyboard_asset_tasks.sql",
	} {
		if err := applyContentMigration(db, path); err != nil {
			t.Fatalf("迁移 %s: %v", path, err)
		}
	}
	registry, err := resource.NewRegistry(
		handlers.ProjectCreate{},
		handlers.StoryboardCreate{},
		handlers.StoryboardEpisodeCreate{},
		handlers.StoryboardEpisodePatch{},
		handlers.StoryboardEpisodeDelete{},
		handlers.StoryboardSceneCreate{},
		handlers.StoryboardItemCreate{},
		handlers.AssetCreate{},
		handlers.AssetDelete{},
		handlers.AssetItemCreate{},
		handlers.ImageTaskCreate{},
		handlers.ImageTaskPatch{},
	)
	if err != nil {
		t.Fatal(err)
	}
	svc := resource.NewService(db, registry)
	reads := resource.NewContentQuery(db)
	ctx := context.Background()
	const actor int64 = 7

	project, err := svc.Execute(ctx, resource.CreateCommand{
		ReqID:   "proj-1",
		ResType: resource.TypeProject,
		Fields:  map[string]any{"name": "短剧"},
	}, resource.Context{ActorUserID: actor})
	if err != nil {
		t.Fatal(err)
	}
	projectID := resultID(t, project.Result)

	board, err := svc.Execute(ctx, resource.CreateCommand{
		ReqID:     "board-1",
		ResType:   resource.TypeStoryboard,
		ProjectID: &projectID,
		Fields:    map[string]any{"title": "第一稿", "description": "夜"},
	}, resource.Context{ActorUserID: actor})
	if err != nil {
		t.Fatal(err)
	}
	boardID := resultID(t, board.Result)
	episode, err := svc.Execute(ctx, resource.CreateCommand{
		ReqID:   "ep-1",
		ResType: resource.TypeStoryboardEpisode,
		Fields: map[string]any{
			"storyboardId":  boardID,
			"parentVersion": float64(0),
			"episodeNumber": float64(1),
			"title":         "开场",
		},
	}, resource.Context{ActorUserID: actor})
	if err != nil {
		t.Fatal(err)
	}
	episodeID := resultID(t, episode.Result)
	if _, err := reads.StoryboardEpisode(ctx, actor+1, episodeID); !isForbidden(err) {
		t.Fatalf("other user must be forbidden, got %v", err)
	}
	if _, err := svc.Execute(ctx, resource.PatchCommand{
		ReqID:   "ep-patch-stale",
		ResType: resource.TypeStoryboardEpisode,
		ID:      episodeID,
		Version: 3,
		Patch:   map[string]any{"title": "不该成功"},
	}, resource.Context{ActorUserID: actor}); !isConflict(err) {
		t.Fatalf("stale version must conflict, got %v", err)
	}
	renamed, err := svc.Execute(ctx, resource.PatchCommand{
		ReqID:   "ep-patch",
		ResType: resource.TypeStoryboardEpisode,
		ID:      episodeID,
		Version: 0,
		Patch:   map[string]any{"title": "开场重拍"},
	}, resource.Context{ActorUserID: actor})
	if err != nil {
		t.Fatal(err)
	}
	again, err := svc.Execute(ctx, resource.PatchCommand{
		ReqID:   "ep-patch",
		ResType: resource.TypeStoryboardEpisode,
		ID:      episodeID,
		Version: 0,
		Patch:   map[string]any{"title": "开场重拍"},
	}, resource.Context{ActorUserID: actor})
	if err != nil {
		t.Fatal(err)
	}
	if again.CommandID != renamed.CommandID {
		t.Fatalf("replay command %s != %s", again.CommandID, renamed.CommandID)
	}
	scene, err := svc.Execute(ctx, resource.CreateCommand{
		ReqID:   "scene-1",
		ResType: resource.TypeStoryboardScene,
		Fields: map[string]any{
			"episodeId":     episodeID,
			"parentVersion": float64(1),
			"sceneHeading":  "内景",
		},
	}, resource.Context{ActorUserID: actor})
	if err != nil {
		t.Fatal(err)
	}
	sceneID := resultID(t, scene.Result)
	item, err := svc.Execute(ctx, resource.CreateCommand{
		ReqID:   "shot-1",
		ResType: resource.TypeShot,
		Fields: map[string]any{
			"storyboardSceneId": sceneID,
			"parentVersion":     float64(0),
			"content":           "雨",
		},
	}, resource.Context{ActorUserID: actor})
	if err != nil {
		t.Fatal(err)
	}
	shotID := resultID(t, item.Result)
	deletedEpisode, err := svc.Execute(ctx, resource.DeleteCommand{
		ReqID:   "ep-del",
		ResType: resource.TypeStoryboardEpisode,
		ID:      episodeID,
		Version: 2,
	}, resource.Context{ActorUserID: actor})
	if err != nil {
		t.Fatal(err)
	}
	var childRevisions int
	if err := db.QueryRow(`SELECT COUNT(*) FROM afv_resource_revision WHERE command_id = ? AND resource_type IN ('shot', 'storyboardScene')`, deletedEpisode.CommandID).Scan(&childRevisions); err != nil {
		t.Fatal(err)
	}
	if childRevisions != 2 {
		t.Fatalf("episode delete should revise the shot and the scene, got %d", childRevisions)
	}
	if _, err := reads.StoryboardScene(ctx, actor, sceneID); !isNotFound(err) {
		t.Fatalf("scene should be cascade-deleted, got %v", err)
	}
	if _, err := reads.StoryboardItem(ctx, actor, shotID); !isNotFound(err) {
		t.Fatalf("shot should be cascade-deleted, got %v", err)
	}
	if _, err := reads.StoryboardEpisode(ctx, actor, episodeID); !isNotFound(err) {
		t.Fatalf("deleted episode must be hidden, got %v", err)
	}
	if _, err := reads.StoryboardEpisode(ctx, actor+1, episodeID); !isForbidden(err) && !isNotFound(err) {
		t.Fatalf("other user must not read the episode, got %v", err)
	}

	asset, err := svc.Execute(ctx, resource.CreateCommand{
		ReqID:     "asset-1",
		ResType:   resource.TypeAsset,
		ProjectID: &projectID,
		Fields:    map[string]any{"type": "character", "name": "林夏"},
	}, resource.Context{ActorUserID: actor})
	if err != nil {
		t.Fatal(err)
	}
	assetID := resultID(t, asset.Result)
	if _, err := svc.Execute(ctx, resource.CreateCommand{
		ReqID:   "item-1",
		ResType: resource.TypeAssetItem,
		Fields: map[string]any{
			"assetId":       assetID,
			"parentVersion": float64(0),
			"itemType":      "front",
			"name":          "正面",
		},
	}, resource.Context{ActorUserID: actor}); err != nil {
		t.Fatal(err)
	}
	items, err := reads.AssetItems(ctx, actor, assetID)
	if err != nil || len(items) != 1 {
		t.Fatalf("items %v err %v", items, err)
	}
	if _, err := svc.Execute(ctx, resource.DeleteCommand{
		ReqID:   "asset-del",
		ResType: resource.TypeAsset,
		ID:      assetID,
		Version: 1,
	}, resource.Context{ActorUserID: actor}); err != nil {
		t.Fatal(err)
	}
	if _, err := reads.AssetItem(ctx, actor, resultID(t, items[0])); !isNotFound(err) {
		t.Fatalf("asset item should be cascade-deleted, got %v", err)
	}

	image, err := svc.Execute(ctx, resource.CreateCommand{
		ReqID:     "img-1",
		ResType:   resource.TypeImageTask,
		ProjectID: &projectID,
		Fields:    map[string]any{"prompt": "雨夜"},
	}, resource.Context{ActorUserID: actor})
	if err != nil {
		t.Fatal(err)
	}
	imageID := resultID(t, image.Result)
	publicID, _ := image.Result.(map[string]any)["taskId"].(string)
	var skew int
	if err := db.QueryRow(`SELECT TIMESTAMPDIFF(SECOND, create_time, UTC_TIMESTAMP()) FROM afv_image_task WHERE id = ?`, imageID).Scan(&skew); err != nil {
		t.Fatal(err)
	}
	if skew < 0 || skew > 5 {
		t.Fatalf("image task create_time must follow UTC_TIMESTAMP, skew %d seconds", skew)
	}
	var boardSkew int
	if err := db.QueryRow(`SELECT TIMESTAMPDIFF(HOUR, create_time, UTC_TIMESTAMP()) FROM afv_storyboard WHERE id = ?`, boardID).Scan(&boardSkew); err != nil {
		t.Fatal(err)
	}
	if boardSkew != -8 && boardSkew != 8 && boardSkew != 0 {
		t.Fatalf("unexpected storyboard clock skew %d", boardSkew)
	}
	if boardSkew == 0 {
		t.Fatal("storyboard content clock should follow the session timezone, not UTC")
	}
	got, err := reads.ImageTask(ctx, actor, publicID)
	if err != nil {
		t.Fatal(err)
	}
	if got["prompt"] != "雨夜" || got["version"] != int32(0) {
		t.Fatalf("image read %+v", got)
	}
	if err := reads.DeleteImageTask(ctx, actor, imageID); !isConflict(err) {
		t.Fatalf("queued image task must not be deleted, got %v", err)
	}
	if _, err := db.Exec(`UPDATE afv_image_task SET status = 2 WHERE id = ?`, imageID); err != nil {
		t.Fatal(err)
	}
	if err := reads.DeleteImageTask(ctx, actor, imageID); err != nil {
		t.Fatal(err)
	}
	if _, err := reads.ImageTask(ctx, actor, publicID); !isNotFound(err) {
		t.Fatalf("deleted task must be hidden, got %v", err)
	}
}

func TestStoryboardClearKeepsBoard(t *testing.T) {
	db := openContentMySQL(t)
	for _, path := range []string{
		"../../db/migrations/00001_sys_user_and_project.sql",
		"../../db/migrations/00002_resource_command_and_script.sql",
		"../../db/migrations/00003_script_scene.sql",
		"../../db/migrations/00004_generation_jobs.sql",
		"../../db/migrations/00005_media_job.sql",
		"../../db/migrations/00006_storyboard_asset_tasks.sql",
	} {
		if err := applyContentMigration(db, path); err != nil {
			t.Fatalf("迁移 %s: %v", path, err)
		}
	}
	registry, err := resource.NewRegistry(
		handlers.ProjectCreate{},
		handlers.StoryboardCreate{},
		handlers.StoryboardClear{},
		handlers.StoryboardEpisodeCreate{},
		handlers.StoryboardSceneCreate{},
		handlers.StoryboardItemCreate{},
	)
	if err != nil {
		t.Fatal(err)
	}
	svc := resource.NewService(db, registry)
	ctx := context.Background()
	const actor int64 = 7
	project, err := svc.Execute(ctx, resource.CreateCommand{
		ReqID: "proj-clear", ResType: resource.TypeProject, Fields: map[string]any{"name": "清空"},
	}, resource.Context{ActorUserID: actor})
	if err != nil {
		t.Fatal(err)
	}
	projectID := resultID(t, project.Result)
	board, err := svc.Execute(ctx, resource.CreateCommand{
		ReqID: "board-clear", ResType: resource.TypeStoryboard, ProjectID: &projectID,
		Fields: map[string]any{"title": "稿"},
	}, resource.Context{ActorUserID: actor})
	if err != nil {
		t.Fatal(err)
	}
	boardID := resultID(t, board.Result)
	episode, err := svc.Execute(ctx, resource.CreateCommand{
		ReqID: "ep-clear", ResType: resource.TypeStoryboardEpisode,
		Fields: map[string]any{"storyboardId": boardID, "parentVersion": float64(0), "episodeNumber": float64(1), "title": "集"},
	}, resource.Context{ActorUserID: actor})
	if err != nil {
		t.Fatal(err)
	}
	episodeID := resultID(t, episode.Result)
	scene, err := svc.Execute(ctx, resource.CreateCommand{
		ReqID: "scene-clear", ResType: resource.TypeStoryboardScene,
		Fields: map[string]any{"episodeId": episodeID, "parentVersion": float64(0), "sceneHeading": "内景"},
	}, resource.Context{ActorUserID: actor})
	if err != nil {
		t.Fatal(err)
	}
	sceneID := resultID(t, scene.Result)
	if _, err := svc.Execute(ctx, resource.CreateCommand{
		ReqID: "shot-clear", ResType: resource.TypeShot,
		Fields: map[string]any{"storyboardSceneId": sceneID, "parentVersion": float64(0), "content": "雨"},
	}, resource.Context{ActorUserID: actor}); err != nil {
		t.Fatal(err)
	}
	var version int32
	if err := db.QueryRow(`SELECT version FROM afv_storyboard WHERE id = ?`, boardID).Scan(&version); err != nil {
		t.Fatal(err)
	}
	cleared, err := svc.Execute(ctx, resource.ClearCommand{
		ReqID: "clear-board", ResType: resource.TypeStoryboard, ID: boardID, Version: version,
	}, resource.Context{ActorUserID: actor})
	if err != nil {
		t.Fatal(err)
	}
	if resultID(t, cleared.Result) != boardID {
		t.Fatalf("clear should return the board, got %v", cleared.Result)
	}
	var boardDeleted, childDeleted int
	if err := db.QueryRow(`SELECT deleted FROM afv_storyboard WHERE id = ?`, boardID).Scan(&boardDeleted); err != nil {
		t.Fatal(err)
	}
	if boardDeleted != 0 {
		t.Fatal("clear must keep the storyboard row")
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM afv_storyboard_item WHERE storyboard_id = ? AND deleted = 0`, boardID).Scan(&childDeleted); err != nil {
		t.Fatal(err)
	}
	if childDeleted != 0 {
		t.Fatalf("shots still visible: %d", childDeleted)
	}
	var revisions int
	if err := db.QueryRow(`SELECT COUNT(*) FROM afv_resource_revision WHERE command_id = ? AND resource_type IN ('shot', 'storyboardScene', 'storyboardEpisode')`, cleared.CommandID).Scan(&revisions); err != nil {
		t.Fatal(err)
	}
	if revisions != 3 {
		t.Fatalf("child revisions = %d, want 3", revisions)
	}
}

func TestScriptSourceWritesChildRevisions(t *testing.T) {
	db := openContentMySQL(t)
	for _, path := range []string{
		"../../db/migrations/00001_sys_user_and_project.sql",
		"../../db/migrations/00002_resource_command_and_script.sql",
		"../../db/migrations/00003_script_scene.sql",
		"../../db/migrations/00004_generation_jobs.sql",
		"../../db/migrations/00005_media_job.sql",
		"../../db/migrations/00006_storyboard_asset_tasks.sql",
	} {
		if err := applyContentMigration(db, path); err != nil {
			t.Fatalf("迁移 %s: %v", path, err)
		}
	}
	registry, err := resource.NewRegistry(
		handlers.ProjectCreate{},
		handlers.ScriptSource{},
		handlers.ScriptEpisodeCreate{},
		handlers.ScriptSceneCreate{},
	)
	if err != nil {
		t.Fatal(err)
	}
	svc := resource.NewService(db, registry)
	ctx := context.Background()
	const actor int64 = 7
	project, err := svc.Execute(ctx, resource.CreateCommand{
		ReqID: "proj-src", ResType: resource.TypeProject, Fields: map[string]any{"name": "原文"},
	}, resource.Context{ActorUserID: actor})
	if err != nil {
		t.Fatal(err)
	}
	projectID := resultID(t, project.Result)
	var scriptID int64
	var scriptVersion int32
	if err := db.QueryRow(`SELECT id, version FROM afv_script WHERE project_id = ? AND deleted = 0`, projectID).Scan(&scriptID, &scriptVersion); err != nil {
		t.Fatal(err)
	}
	episode, err := svc.Execute(ctx, resource.CreateCommand{
		ReqID: "src-ep", ResType: resource.TypeScriptEpisode,
		Fields: map[string]any{"scriptId": scriptID, "parentVersion": float64(scriptVersion), "episodeNumber": float64(1), "title": "第一集"},
	}, resource.Context{ActorUserID: actor})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Execute(ctx, resource.CreateCommand{
		ReqID: "src-scene", ResType: resource.TypeScriptScene,
		Fields: map[string]any{"episodeId": resultID(t, episode.Result), "parentVersion": float64(0)},
	}, resource.Context{ActorUserID: actor}); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT version FROM afv_script WHERE id = ?`, scriptID).Scan(&scriptVersion); err != nil {
		t.Fatal(err)
	}
	replaced, err := svc.Execute(ctx, resource.SourceCommand{
		ReqID: "src-replace", ResourceID: scriptID, Version: scriptVersion, RawContent: "新的原文",
	}, resource.Context{ActorUserID: actor})
	if err != nil {
		t.Fatal(err)
	}
	var sceneRevisions, episodeRevisions int
	if err := db.QueryRow(`SELECT COUNT(*) FROM afv_resource_revision WHERE command_id = ? AND resource_type = 'script_scene'`, replaced.CommandID).Scan(&sceneRevisions); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM afv_resource_revision WHERE command_id = ? AND resource_type = 'script_episode'`, replaced.CommandID).Scan(&episodeRevisions); err != nil {
		t.Fatal(err)
	}
	if sceneRevisions != 1 || episodeRevisions != 1 {
		t.Fatalf("scene revisions %d, episode revisions %d", sceneRevisions, episodeRevisions)
	}
}

func openContentMySQL(t *testing.T) *sql.DB {
	t.Helper()
	adminDSN := mysqlAdminDSN()
	admin, err := sql.Open("mysql", adminDSN)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := admin.PingContext(ctx); err != nil {
		_ = admin.Close()
		t.Skipf("MySQL 不可用，跳过集成测试: %v", err)
	}
	name := "afvit_" + strings.ToLower(strings.ReplaceAll(t.Name(), "/", "_"))
	if _, err := admin.Exec("DROP DATABASE IF EXISTS `" + name + "`"); err != nil {
		t.Fatal(err)
	}
	if _, err := admin.Exec("CREATE DATABASE `" + name + "`"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = admin.Exec("DROP DATABASE `" + name + "`")
		_ = admin.Close()
	})
	dsn := mysqlSchemaDSN(adminDSN, name)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.PingContext(ctx); err != nil {
		t.Fatal(err)
	}
	return db
}

func applyContentMigration(db *sql.DB, path string) error {
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	text := string(body)
	const upMark = "-- +goose Up"
	const downMark = "-- +goose Down"
	start := strings.Index(text, upMark)
	if start < 0 {
		return fmt.Errorf("缺少 %s", upMark)
	}
	text = text[start+len(upMark):]
	if end := strings.Index(text, downMark); end >= 0 {
		text = text[:end]
	}
	_, err = db.Exec(text)
	return err
}

func resultID(t *testing.T, result any) int64 {
	t.Helper()
	m, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("result type %T", result)
	}
	switch n := m["id"].(type) {
	case int64:
		return n
	case int32:
		return int64(n)
	case float64:
		return int64(n)
	default:
		t.Fatalf("id type %T", m["id"])
		return 0
	}
}

func isConflict(err error) bool {
	var ae *apperr.Error
	return errors.As(err, &ae) && ae.HTTPStatus == 409
}

func isNotFound(err error) bool {
	var ae *apperr.Error
	return errors.As(err, &ae) && ae.HTTPStatus == 404
}

func isForbidden(err error) bool {
	var ae *apperr.Error
	return errors.As(err, &ae) && ae.HTTPStatus == 403
}

func mysqlAdminDSN() string {
	if v := strings.TrimSpace(os.Getenv("AFV_MYSQL_ADMIN_DSN")); v != "" {
		return v
	}
	return "afvtest:afvtest@tcp(127.0.0.1:3306)/?parseTime=true&loc=UTC&multiStatements=true&time_zone=%27%2B08%3A00%27"
}

func mysqlSchemaDSN(adminDSN, name string) string {
	marker := ")/"
	i := strings.Index(adminDSN, marker)
	if i < 0 {
		return adminDSN
	}
	rest := adminDSN[i+len(marker):]
	query := ""
	if j := strings.Index(rest, "?"); j >= 0 {
		query = rest[j:]
	}
	return adminDSN[:i+len(marker)] + name + query
}
