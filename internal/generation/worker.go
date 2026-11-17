package generation

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/apperr"
)

type Worker struct {
	repo     *Repository
	instance string
	slots    chan struct{}
	interval time.Duration
	hbEvery  time.Duration

	mu     sync.Mutex
	active map[Claim]struct{}
}

func NewWorker(repo *Repository) *Worker {
	return &Worker{
		repo:     repo,
		instance: "generation-" + newUUID(),
		slots:    make(chan struct{}, 16),
		interval: 500 * time.Millisecond,
		hbEvery:  20 * time.Second,
		active:   make(map[Claim]struct{}),
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
			w.mu.Lock()
			claims := make([]Claim, 0, len(w.active))
			for c := range w.active {
				claims = append(claims, c)
			}
			w.mu.Unlock()
			for _, c := range claims {
				_ = w.repo.Heartbeat(ctx, c)
			}
		case <-dispatch.C:
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case w.slots <- struct{}{}:
				default:
					goto waitNext
				}

				job, err := w.repo.Claim(ctx, w.instance)
				if err != nil {
					<-w.slots
					log.Printf("generation claim: %v", err)
					goto waitNext
				}
				if job == nil {
					<-w.slots
					goto waitNext
				}

				claim := job.Claim()
				w.track(claim, true)
				wg.Add(1)
				go func(j Job, c Claim) {
					defer wg.Done()
					defer func() { <-w.slots }()
					defer w.track(c, false)
					if err := w.execute(ctx, j); err != nil && !errors.Is(err, context.Canceled) {
						if !errors.Is(err, ErrLeaseLost) {
							log.Printf("generation job %s: %v", j.ID, err)
						}
					}
				}(*job, claim)
			}
		waitNext:
		}
	}
}

func (w *Worker) track(claim Claim, add bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if add {
		w.active[claim] = struct{}{}
		return
	}
	delete(w.active, claim)
}

func (w *Worker) execute(ctx context.Context, initial Job) error {
	job := initial
	claim := job.Claim()

	if job.CancelRequested && (job.Phase == PhaseNew || job.Phase == PhasePreparing) {
		return w.repo.Fail(ctx, claim, StateCancelled, "CANCELLED_BEFORE_SUBMISSION", "已取消排队，尚未提交供应商")
	}
	if time.Now().UTC().After(job.DeadlineAt) {
		state := StateFailed
		if job.Phase != PhaseNew && job.Phase != PhasePreparing {
			state = StateUnknownSubmission
		}
		return w.repo.Fail(ctx, claim, state, "GENERATION_DEADLINE_EXCEEDED", "生成任务超过截止时间")
	}

	switch job.Phase {
	case PhaseNew, PhasePreparing:
		return w.submitStub(ctx, claim, job)
	case PhaseSubmitting:
		if job.ProviderTaskID == nil || *job.ProviderTaskID == "" {
			return w.repo.Fail(ctx, claim, StateUnknownSubmission, "UNKNOWN_SUBMISSION",
				"提交阶段崩溃且无供应商任务标识，禁止盲目重提")
		}
		return w.repo.WaitForNextPoll(ctx, claim, 1, PhasePolling, nil, nil)
	case PhasePolling, PhaseIngesting:
		return w.finishStub(ctx, claim, job)
	default:
		return w.repo.Fail(ctx, claim, StateFailed, "UNSUPPORTED_PHASE", "不支持的执行阶段: "+job.Phase)
	}
}

func (w *Worker) submitStub(ctx context.Context, claim Claim, job Job) error {
	attemptID := job.ActiveAttemptID
	if attemptID == nil || *attemptID == "" {
		id := newUUID()
		attemptID = &id
		req, err := json.Marshal(map[string]any{
			"provider": "echo-stub",
			"input":    json.RawMessage(job.InputJSON),
		})
		if err != nil {
			return err
		}
		if err := w.repo.PrepareAttempt(ctx, claim, id, req); err != nil {
			return err
		}
	}

	providerTaskID := "echo-" + job.ID
	if err := w.repo.BeginSubmission(ctx, claim, *attemptID, providerTaskID); err != nil {
		return err
	}
	// 释放租约，改为 WAITING；下次到点再 poll，不占线程。
	return w.repo.WaitForNextPoll(ctx, claim, 1, PhasePolling, nil, nil)
}

func (w *Worker) finishStub(ctx context.Context, claim Claim, job Job) error {
	var prompt string
	var input struct {
		Prompt string `json:"prompt"`
	}
	_ = json.Unmarshal(job.InputJSON, &input)
	prompt = input.Prompt

	result, err := json.Marshal(map[string]any{
		"state": "SUCCEEDED",
		"outputs": []map[string]any{
			{"text": "echo: " + prompt},
		},
	})
	if err != nil {
		return err
	}
	if err := w.repo.Succeed(ctx, claim, result); err != nil {
		var ae *apperr.Error
		if errors.As(err, &ae) {
			return err
		}
		return err
	}
	return nil
}
