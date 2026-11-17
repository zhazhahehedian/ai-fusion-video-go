package media

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Worker struct {
	repo     *Repository
	runner   *ProcessRunner
	ffmpeg   string
	instance string
	interval time.Duration
	hbEvery  time.Duration

	busy   atomic.Bool
	active atomic.Pointer[Claim]
}

func NewWorker(repo *Repository, ffmpegPath string) *Worker {
	if strings.TrimSpace(ffmpegPath) == "" {
		ffmpegPath = "ffmpeg"
	}
	return &Worker{
		repo:     repo,
		runner:   NewProcessRunner(),
		ffmpeg:   ffmpegPath,
		instance: "media-" + newUUID(),
		interval: time.Second,
		hbEvery:  20 * time.Second,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	dispatch := time.NewTicker(w.interval)
	heartbeat := time.NewTicker(w.hbEvery)
	defer dispatch.Stop()
	defer heartbeat.Stop()

	var wg sync.WaitGroup
	defer wg.Wait()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-heartbeat.C:
			if c := w.active.Load(); c != nil {
				_ = w.repo.Heartbeat(ctx, *c)
			}
		case <-dispatch.C:
			if !w.busy.CompareAndSwap(false, true) {
				continue
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer w.busy.Store(false)
				if err := w.runOne(ctx); err != nil && !errors.Is(err, context.Canceled) {
					if !errors.Is(err, ErrLeaseLost) {
						log.Printf("media worker: %v", err)
					}
				}
			}()
		}
	}
}

func (w *Worker) runOne(ctx context.Context) error {
	job, err := w.repo.Claim(ctx, w.instance)
	if err != nil {
		return err
	}
	if job == nil {
		return nil
	}
	claim := job.Claim()
	w.active.Store(&claim)
	defer w.active.Store(nil)

	if job.CancelRequested {
		return w.repo.Cancel(ctx, claim, "已取消", "任务在执行前被取消")
	}
	if time.Now().After(job.DeadlineAt) {
		return w.repo.Fail(ctx, claim, "已超时", "媒体任务超过截止时间")
	}

	switch job.Kind {
	case KindFFmpegVersion:
		return w.runFFmpegVersion(ctx, claim)
	default:
		return w.repo.Fail(ctx, claim, "不支持", "不支持的媒体任务类型: "+job.Kind)
	}
}

func (w *Worker) runFFmpegVersion(ctx context.Context, claim Claim) error {
	workDir, err := os.MkdirTemp("", "afv-media-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(workDir)

	if _, err := w.repo.Check(ctx, claim); err != nil {
		return err
	}
	_ = w.repo.Checkpoint(ctx, claim, json.RawMessage(`{}`), "正在探测 ffmpeg")

	out, err := w.runner.Run(ctx, []string{w.ffmpeg, "-version"}, workDir, func() error {
		_, err := w.repo.Check(ctx, claim)
		return err
	})
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, ErrLeaseLost) {
			return err
		}
		msg := err.Error()
		if strings.Contains(msg, "无法启动进程") {
			msg = "无法启动 FFmpeg，请检查 FFMPEG_PATH: " + msg
		}
		return w.repo.Fail(ctx, claim, "ffmpeg 失败", trimOut(msg))
	}

	checkpoint, err := json.Marshal(map[string]any{
		"ffmpegVersion": firstLine(out),
		"workDir":       filepath.Base(workDir),
	})
	if err != nil {
		return err
	}
	return w.repo.Succeed(ctx, claim, checkpoint, "ffmpeg 探测完成")
}

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return strings.TrimSpace(s[:i])
	}
	return s
}
