package generation

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/apperr"
)

type Worker struct {
	repo     *Repository
	logger   *slog.Logger
	instance string
	slots    chan struct{}
	interval time.Duration
	hbEvery  time.Duration

	mu     sync.Mutex
	active map[Claim]struct{}
}

func NewWorker(repo *Repository, logger *slog.Logger) *Worker {
	if logger == nil {
		panic("generation.NewWorker: logger 未注入")
	}
	instance := "generation-" + newUUID()
	return &Worker{
		repo:     repo,
		logger:   logger.With(slog.String("component", "generation_worker"), slog.String("instance", instance)),
		instance: instance,
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
				if err := w.repo.Heartbeat(ctx, c); err != nil {
					w.logger.WarnContext(ctx, "生成任务续租失败",
						slog.String("job_id", c.ID),
						slog.Int64("owner_epoch", c.Epoch),
						slog.String("err", err.Error()),
					)
				}
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
					w.logger.ErrorContext(ctx, "领取生成任务失败", "err", err)
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
					jobLog := w.logger.With(
						slog.String("job_id", j.ID),
						slog.Int64("owner_epoch", c.Epoch),
						slog.String("phase", j.Phase),
						slog.String("state", j.State),
					)
					if err := w.execute(ctx, j); err != nil && !errors.Is(err, context.Canceled) {
						if !errors.Is(err, ErrLeaseLost) {
							jobLog.ErrorContext(ctx, "生成任务执行失败", "err", err)
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

	permitted, err := w.repo.ExecutionPermitted(ctx, job.SubmissionID)
	if err != nil {
		return err
	}
	now, err := w.repo.Now(ctx)
	if err != nil {
		return err
	}
	if stop, state, code, msg := gateExecution(permitted, job.CancelRequested, job.Phase, now, job.DeadlineAt); stop {
		return w.repo.Fail(ctx, claim, state, code, msg)
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

// gateExecution 对齐 Java RemoteGenerationWorker.execute 的前三段判断：
// 身份失效、准备阶段取消、以及除 INGESTING 以外的截止时间。
// 已进入供应商轮询后的身份失效在这里直接停止。Java 会先 requestStop 再靠后续轮询收尾；
// 本仓库还没有真实供应商轮询，不能把已失效的任务继续标成成功。
func gateExecution(permitted, cancelRequested bool, phase string, now, deadline time.Time) (stop bool, state, code, msg string) {
	preparing := phase == PhaseNew || phase == PhasePreparing
	if !permitted && !cancelRequested {
		if preparing {
			return true, StateCancelled, "CANCELLED_BEFORE_SUBMISSION", "已取消排队，尚未提交供应商"
		}
		return true, StateCancelled, "EXECUTION_NOT_PERMITTED", "执行身份已失效，任务已停止"
	}
	if preparing && cancelRequested {
		return true, StateCancelled, "CANCELLED_BEFORE_SUBMISSION", "已取消排队，尚未提交供应商"
	}
	if phase != PhaseIngesting && now.After(deadline) {
		if preparing {
			return true, StateFailed, "GENERATION_DEADLINE_EXCEEDED", "准备超过等待期限，未提交供应商"
		}
		return true, StateUnknownSubmission, "GENERATION_DEADLINE_EXCEEDED", "远端结果未知，已结束本地等待，原请求不会自动重发"
	}
	return false, "", "", ""
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
