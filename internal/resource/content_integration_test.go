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
		handlers.AssetCreate{},
		handlers.AssetItemCreate{},
		handlers.ImageTaskCreate{},
		handlers.ImageTaskPatch{},
		handlers.ImageTaskDelete{},
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
	if _, err := svc.Execute(ctx, resource.DeleteCommand{
		ReqID:   "ep-del",
		ResType: resource.TypeStoryboardEpisode,
		ID:      episodeID,
		Version: 1,
	}, resource.Context{ActorUserID: actor}); err != nil {
		t.Fatal(err)
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
	if _, err := svc.Execute(ctx, resource.DeleteCommand{
		ReqID:   "img-del",
		ResType: resource.TypeImageTask,
		ID:      imageID,
		Version: 0,
	}, resource.Context{ActorUserID: actor}); err != nil {
		t.Fatal(err)
	}
	if _, err := reads.ImageTask(ctx, actor, publicID); !isNotFound(err) {
		t.Fatalf("deleted task must be hidden, got %v", err)
	}
}

func openContentMySQL(t *testing.T) *sql.DB {
	t.Helper()
	adminDSN := "afvtest:afvtest@tcp(127.0.0.1:3306)/?parseTime=true&loc=UTC&multiStatements=true&time_zone=%27%2B08%3A00%27"
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
	dsn := "afvtest:afvtest@tcp(127.0.0.1:3306)/" + name + "?parseTime=true&loc=UTC&multiStatements=true&time_zone=%27%2B08%3A00%27"
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
