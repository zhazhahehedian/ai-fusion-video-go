package generation

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/apperr"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/media"

	_ "github.com/go-sql-driver/mysql"
)

// 本机没有 Docker 时跳过。会话时区固定为 +08:00，用来确认写入用的是 UTC_TIMESTAMP 而不是 CURRENT_TIMESTAMP。
func testMySQL(t *testing.T) *sql.DB {
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
	db.SetMaxOpenConns(32)
	t.Cleanup(func() { _ = db.Close() })
	if err := db.PingContext(ctx); err != nil {
		t.Fatal(err)
	}
	var hours int
	if err := db.QueryRow("SELECT TIMESTAMPDIFF(HOUR, UTC_TIMESTAMP(), CURRENT_TIMESTAMP())").Scan(&hours); err != nil {
		t.Fatal(err)
	}
	if hours != 8 {
		t.Fatalf("会话时区应比 UTC 快 8 小时，实际差 %d 小时", hours)
	}
	for _, path := range []string{"../../db/migrations/00004_generation_jobs.sql", "../../db/migrations/00005_media_job.sql"} {
		if err := applyGooseUp(db, path); err != nil {
			t.Fatalf("迁移 %s: %v", path, err)
		}
	}
	return db
}

func applyGooseUp(db *sql.DB, path string) error {
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
	if _, err := db.Exec(text); err != nil {
		return err
	}
	return nil
}

func TestMySQLConcurrentEnqueueReplay(t *testing.T) {
	db := testMySQL(t)
	svc := NewService(db)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	const n = 16
	results := make([]EchoEnqueueResult, n)
	errs := make([]error, n)
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			results[i], errs[i] = svc.EnqueueEcho(ctx, 42, EchoRequest{RequestID: "same-req", Prompt: "同一提示"})
		}(i)
	}
	close(start)
	wg.Wait()

	var created int
	var id string
	for i := range n {
		if errs[i] != nil {
			t.Fatalf("并发入队返回错误: %v", errs[i])
		}
		if id == "" {
			id = results[i].ID
		}
		if results[i].ID != id || results[i].State != StateQueued {
			t.Fatalf("结果不一致: %+v", results[i])
		}
		if !results[i].Replay {
			created++
		}
	}
	if created != 1 {
		t.Fatalf("新建次数 = %d，想要 1", created)
	}
	assertCount(t, db, "afv_generation_submission", 1)
	assertCount(t, db, "afv_remote_generation_job", 1)
	assertCount(t, db, "afv_async_outbox", 1)

	var utcDelta, localDelta, deadlineSeconds int
	if err := db.QueryRow(`
		SELECT TIMESTAMPDIFF(SECOND, UTC_TIMESTAMP(6), s.created_at),
		       TIMESTAMPDIFF(SECOND, CURRENT_TIMESTAMP(6), s.created_at),
		       TIMESTAMPDIFF(SECOND, j.created_at, j.deadline_at)
		FROM afv_generation_submission s
		JOIN afv_remote_generation_job j ON j.submission_id = s.id`).Scan(&utcDelta, &localDelta, &deadlineSeconds); err != nil {
		t.Fatal(err)
	}
	if utcDelta < -5 || utcDelta > 5 {
		t.Fatalf("created_at 与 UTC_TIMESTAMP 相差 %d 秒", utcDelta)
	}
	if localDelta > -28790 || localDelta < -28810 {
		t.Fatalf("created_at 仍贴近会话时区 CURRENT_TIMESTAMP，相差 %d 秒", localDelta)
	}
	if deadlineSeconds < 7190 || deadlineSeconds > 7210 {
		t.Fatalf("deadline 距创建 %d 秒，想要约 2 小时", deadlineSeconds)
	}
}

func TestMySQLConcurrentEnqueueConflict(t *testing.T) {
	db := testMySQL(t)
	svc := NewService(db)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	const n = 16
	errs := make([]error, n)
	results := make([]EchoEnqueueResult, n)
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			prompt := "甲"
			if i%2 == 1 {
				prompt = "乙"
			}
			results[i], errs[i] = svc.EnqueueEcho(ctx, 7, EchoRequest{RequestID: "conflict-req", Prompt: prompt})
		}(i)
	}
	close(start)
	wg.Wait()

	var okN, conflictN int
	var id string
	for i := range n {
		if errs[i] == nil {
			okN++
			if id == "" {
				id = results[i].ID
			}
			if results[i].ID != id {
				t.Fatalf("成功结果指向了不同受理记录: %s vs %s", id, results[i].ID)
			}
			continue
		}
		var ae *apperr.Error
		if !errors.As(errs[i], &ae) || ae.Code != 409 {
			t.Fatalf("冲突应返回 409，实际 %v", errs[i])
		}
		conflictN++
	}
	if okN == 0 || conflictN == 0 {
		t.Fatalf("成功 %d，冲突 %d，两边都应该出现", okN, conflictN)
	}
	assertCount(t, db, "afv_generation_submission", 1)
	assertCount(t, db, "afv_remote_generation_job", 1)
}

func TestMySQLConcurrentMediaEnqueueReplay(t *testing.T) {
	db := testMySQL(t)
	svc := media.NewService(db)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	const n = 12
	results := make([]media.EnqueueResult, n)
	errs := make([]error, n)
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			results[i], errs[i] = svc.Enqueue(ctx, 9, media.EnqueueRequest{RequestID: "media-req", Kind: media.KindFFmpegVersion})
		}(i)
	}
	close(start)
	wg.Wait()

	var created int
	var id string
	for i := range n {
		if errs[i] != nil {
			t.Fatalf("并发媒体入队返回错误: %v", errs[i])
		}
		if id == "" {
			id = results[i].ID
		}
		if results[i].ID != id {
			t.Fatalf("媒体任务标识不一致: %s vs %s", id, results[i].ID)
		}
		if !results[i].Replay {
			created++
		}
	}
	if created != 1 {
		t.Fatalf("新建次数 = %d，想要 1", created)
	}
	assertCount(t, db, "afv_media_job", 1)

	if _, err := db.Exec("UPDATE afv_media_job SET request_hash = ? WHERE id = ?", strings.Repeat("ab", 32), id); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Enqueue(ctx, 9, media.EnqueueRequest{RequestID: "media-req", Kind: media.KindFFmpegVersion}); err == nil {
		t.Fatal("相同 requestId、不同摘要应冲突")
	} else {
		var ae *apperr.Error
		if !errors.As(err, &ae) || ae.Code != 409 {
			t.Fatalf("媒体冲突应返回 409，实际 %v", err)
		}
	}
}

func TestMySQLOutboxBackoffDoesNotBlockLaterEvents(t *testing.T) {
	db := testMySQL(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	// 白名单里的 GENERATION_ACCEPTED 在写 job 时失败。未登记类型不在领取条件里，挡不住后面的事件。
	if _, err := NewService(db).EnqueueEcho(ctx, 3, EchoRequest{RequestID: "outbox-req", Prompt: "退避"}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `
		CREATE TRIGGER afv_fail_ready BEFORE UPDATE ON afv_remote_generation_job
		FOR EACH ROW
		SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'forced delivery failure'`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO afv_async_outbox (event_key, event_type, aggregate_id, payload_json, next_attempt_at, created_at)
		VALUES ('good-key', 'GENERATION_CHANGED', 'agg-good', JSON_OBJECT(), UTC_TIMESTAMP(6), UTC_TIMESTAMP(6))`); err != nil {
		t.Fatal(err)
	}
	ob := NewOutbox(db)
	published, err := ob.publishOne(ctx)
	var failed *deliveryFailure
	if !errors.As(err, &failed) || published {
		t.Fatalf("坏事件应投递失败并完成退避，published=%v err=%v", published, err)
	}
	published, err = ob.publishOne(ctx)
	if err != nil || !published {
		t.Fatalf("后续事件应已投递，published=%v err=%v", published, err)
	}
	published, err = ob.publishOne(ctx)
	if err != nil || published {
		t.Fatalf("没有到期事件时应停止，published=%v err=%v", published, err)
	}

	var state string
	var attempts int
	var delay int
	var lastError string
	if err := db.QueryRow(`
		SELECT state, attempts, TIMESTAMPDIFF(SECOND, UTC_TIMESTAMP(6), next_attempt_at), last_error
		FROM afv_async_outbox WHERE event_type = 'GENERATION_ACCEPTED'`).Scan(&state, &attempts, &delay, &lastError); err != nil {
		t.Fatal(err)
	}
	if state != "PENDING" || attempts != 1 || delay < 1 || delay > 5 || !strings.Contains(lastError, "事件投递失败") {
		t.Fatalf("坏事件退避状态异常: state=%s attempts=%d delay=%d error=%s", state, attempts, delay, lastError)
	}
	var goodState string
	if err := db.QueryRow("SELECT state FROM afv_async_outbox WHERE event_key = 'good-key'").Scan(&goodState); err != nil {
		t.Fatal(err)
	}
	if goodState != "DELIVERED" {
		t.Fatalf("后续事件状态 = %s", goodState)
	}
}

func assertCount(t *testing.T, db *sql.DB, table string, want int) {
	t.Helper()
	var got int
	if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("%s 行数 = %d，想要 %d", table, got, want)
	}
}
