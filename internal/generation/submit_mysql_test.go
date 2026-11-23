package generation

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/apperr"
)

func TestMySQLImageSubmitReplayAndStubResult(t *testing.T) {
	db := testMySQL(t)
	if err := applyGooseUp(db, "../../db/migrations/00006_storyboard_asset_tasks.sql"); err != nil {
		t.Fatal(err)
	}
	svc := NewService(db)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	count := int32(2)
	req := imageSubmit(7, "img-req", "雨夜的街道", count)
	first, err := svc.EnqueueImage(ctx, 7, req)
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.EnqueueImage(ctx, 7, req)
	if err != nil {
		t.Fatal(err)
	}
	if first == "" || first != second {
		t.Fatalf("公开任务标识 = %q / %q", first, second)
	}
	assertCount(t, db, "afv_image_task", 1)
	assertCount(t, db, "afv_image_item", 2)
	assertCount(t, db, "afv_generation_submission", 1)
	assertCount(t, db, "afv_remote_generation_job", 1)

	changed := req
	prompt := "另一条提示"
	changed.Prompt = &prompt
	if _, err := svc.EnqueueImage(ctx, 7, changed); err == nil {
		t.Fatal("相同 requestId、不同提示应冲突")
	} else {
		var ae *apperr.Error
		if !errors.As(err, &ae) || ae.Code != 409 {
			t.Fatalf("冲突应返回 409，实际 %v", err)
		}
	}
	assertCount(t, db, "afv_image_task", 1)

	repo := NewRepository(db)
	worker := NewWorker(repo, slog.New(slog.DiscardHandler))
	if err := driveGeneration(ctx, t, db, worker); err != nil {
		t.Fatal(err)
	}

	var status, success int
	if err := db.QueryRow(`SELECT status, success_count FROM afv_image_task WHERE task_id = ?`, first).Scan(&status, &success); err != nil {
		t.Fatal(err)
	}
	if status != 2 || success != 2 {
		t.Fatalf("任务状态 = %d，成功数 = %d", status, success)
	}
	var itemStatus int
	var imageURL string
	if err := db.QueryRow(`SELECT status, image_url FROM afv_image_item ORDER BY id LIMIT 1`).Scan(&itemStatus, &imageURL); err != nil {
		t.Fatal(err)
	}
	if itemStatus != 1 || len(imageURL) < len("stub://image/") || imageURL[:len("stub://image/")] != "stub://image/" {
		t.Fatalf("条目 status=%d url=%s", itemStatus, imageURL)
	}
	var submissionState string
	if err := db.QueryRow(`SELECT state FROM afv_generation_submission`).Scan(&submissionState); err != nil {
		t.Fatal(err)
	}
	if submissionState != StateSucceeded {
		t.Fatalf("受理状态 = %s", submissionState)
	}
}

func TestMySQLDeletedImageTaskDoesNotSucceed(t *testing.T) {
	db := testMySQL(t)
	if err := applyGooseUp(db, "../../db/migrations/00006_storyboard_asset_tasks.sql"); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if _, err := NewService(db).EnqueueImage(ctx, 7, imageSubmit(7, "gone", "已删除", 1)); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE afv_image_task SET deleted = 1`); err != nil {
		t.Fatal(err)
	}
	repo := NewRepository(db)
	worker := NewWorker(repo, slog.New(slog.DiscardHandler))
	ob := NewOutbox(db, slog.New(slog.DiscardHandler))
	published, err := ob.publishOne(ctx)
	if err != nil || !published {
		t.Fatalf("投递受理事件 published=%v err=%v", published, err)
	}
	job, err := repo.Claim(ctx, worker.instance)
	if err != nil || job == nil {
		t.Fatalf("领取失败 job=%v err=%v", job, err)
	}
	if err := worker.execute(ctx, *job); err != nil {
		t.Fatal(err)
	}
	var state string
	if err := db.QueryRow(`SELECT state FROM afv_generation_submission`).Scan(&state); err != nil {
		t.Fatal(err)
	}
	if state != StateFailed {
		t.Fatalf("已删除任务的受理状态 = %s", state)
	}
	var itemStatus int
	if err := db.QueryRow(`SELECT status FROM afv_image_item`).Scan(&itemStatus); err != nil {
		t.Fatal(err)
	}
	if itemStatus != 2 {
		t.Fatalf("条目 status=%d，想要失败", itemStatus)
	}
}

func TestMySQLVideoSubmitFailureClosesTask(t *testing.T) {
	db := testMySQL(t)
	if err := applyGooseUp(db, "../../db/migrations/00006_storyboard_asset_tasks.sql"); err != nil {
		t.Fatal(err)
	}
	svc := NewService(db)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	frame := "https://example.com/first.png"
	publicID, err := svc.EnqueueVideo(ctx, 3, VideoSubmitRequest{
		Scope:              &ClientScope{UserID: 3},
		RequestID:          "vid-req",
		FirstFrameImageURL: &frame,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE afv_remote_generation_job SET deadline_at = TIMESTAMPADD(HOUR, -3, UTC_TIMESTAMP(6))`); err != nil {
		t.Fatal(err)
	}
	repo := NewRepository(db)
	worker := NewWorker(repo, slog.New(slog.DiscardHandler))
	ob := NewOutbox(db, slog.New(slog.DiscardHandler))
	if _, err := ob.publishOne(ctx); err != nil {
		t.Fatal(err)
	}
	job, err := repo.Claim(ctx, worker.instance)
	if err != nil || job == nil {
		t.Fatalf("领取失败 job=%v err=%v", job, err)
	}
	if err := worker.execute(ctx, *job); err != nil {
		t.Fatal(err)
	}
	var status int
	var errMsg sql.NullString
	if err := db.QueryRow(`SELECT status, error_msg FROM afv_video_task WHERE task_id = ?`, publicID).Scan(&status, &errMsg); err != nil {
		t.Fatal(err)
	}
	if status != 3 || !errMsg.Valid || errMsg.String == "" {
		t.Fatalf("失败任务 status=%d msg=%v", status, errMsg)
	}
	var itemStatus int
	if err := db.QueryRow(`SELECT status FROM afv_video_item`).Scan(&itemStatus); err != nil {
		t.Fatal(err)
	}
	if itemStatus != 2 {
		t.Fatalf("失败条目 status=%d", itemStatus)
	}
}

func TestMySQLImageSubmitRejectsForeignProject(t *testing.T) {
	db := testMySQL(t)
	for _, path := range []string{
		"../../db/migrations/00001_sys_user_and_project.sql",
		"../../db/migrations/00002_resource_command_and_script.sql",
		"../../db/migrations/00006_storyboard_asset_tasks.sql",
	} {
		if err := applyGooseUp(db, path); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`INSERT INTO afv_project (name, owner_type, owner_id, deleted, create_time, update_time)
		VALUES ('别人的项目', 1, 99, 0, UTC_TIMESTAMP(), UTC_TIMESTAMP())`); err != nil {
		t.Fatal(err)
	}
	var projectID int64
	if err := db.QueryRow(`SELECT id FROM afv_project`).Scan(&projectID); err != nil {
		t.Fatal(err)
	}
	prompt := "一只猫"
	_, err := NewService(db).EnqueueImage(context.Background(), 7, ImageSubmitRequest{
		Scope:     &ClientScope{UserID: 7},
		RequestID: "proj-req",
		ProjectID: &projectID,
		Prompt:    &prompt,
	})
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.Code != 403 {
		t.Fatalf("应拒绝非拥有者，实际 %v", err)
	}
	assertCount(t, db, "afv_image_task", 0)
}

func TestMySQLConcurrentImageSubmitReplay(t *testing.T) {
	db := testMySQL(t)
	if err := applyGooseUp(db, "../../db/migrations/00006_storyboard_asset_tasks.sql"); err != nil {
		t.Fatal(err)
	}
	svc := NewService(db)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	const n = 12
	ids := make([]string, n)
	errs := make([]error, n)
	var wg sync.WaitGroup
	start := make(chan struct{})
	req := imageSubmit(11, "same-image", "同一张图", 1)
	for i := range n {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			ids[i], errs[i] = svc.EnqueueImage(ctx, 11, req)
		}(i)
	}
	close(start)
	wg.Wait()
	var id string
	for i := range n {
		if errs[i] != nil {
			t.Fatalf("并发提交返回错误: %v", errs[i])
		}
		if id == "" {
			id = ids[i]
		}
		if ids[i] != id {
			t.Fatalf("公开任务标识不一致: %s vs %s", id, ids[i])
		}
	}
	assertCount(t, db, "afv_image_task", 1)
	assertCount(t, db, "afv_image_item", 1)
	assertCount(t, db, "afv_generation_submission", 1)
}

func imageSubmit(actor int64, requestID, prompt string, count int32) ImageSubmitRequest {
	return ImageSubmitRequest{
		Scope:     &ClientScope{UserID: actor},
		RequestID: requestID,
		Prompt:    &prompt,
		Count:     &count,
	}
}

func driveGeneration(ctx context.Context, t *testing.T, db *sql.DB, worker *Worker) error {
	t.Helper()
	ob := NewOutbox(db, slog.New(slog.DiscardHandler))
	published, err := ob.publishOne(ctx)
	if err != nil {
		return err
	}
	if !published {
		return errors.New("受理事件没有投递")
	}
	for step := 0; step < 2; step++ {
		if step == 1 {
			if _, err := db.ExecContext(ctx, `UPDATE afv_remote_generation_job SET next_poll_at = UTC_TIMESTAMP(6)`); err != nil {
				return err
			}
		}
		job, err := worker.repo.Claim(ctx, worker.instance)
		if err != nil {
			return err
		}
		if job == nil {
			t.Fatalf("第 %d 次没有领到作业", step+1)
		}
		if err := worker.execute(ctx, *job); err != nil {
			return err
		}
	}
	return nil
}
