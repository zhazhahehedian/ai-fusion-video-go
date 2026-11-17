package generation

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/sqlcdb"
)

// deliveryFailure 表示本条事件投递失败。发布事务已回滚，需要在独立事务里退避。
type deliveryFailure struct {
	id          int64
	eventType   string
	aggregateID string
	err         error
}

func (e *deliveryFailure) Error() string {
	return fmt.Sprintf("事件投递失败 [outboxId=%d]: %v", e.id, e.err)
}

func (e *deliveryFailure) Unwrap() error { return e.err }

// backoffFailure 是独立事务里的退避失败。投递错误仍挂在 delivery 上，循环应停下来。
type backoffFailure struct {
	delivery *deliveryFailure
	err      error
}

func (e *backoffFailure) Error() string {
	return fmt.Sprintf("outbox %d 退避失败: %v（投递错误: %v）", e.delivery.id, e.err, e.delivery.err)
}

func (e *backoffFailure) Unwrap() error { return e.err }

type Outbox struct {
	db       *sql.DB
	logger   *slog.Logger
	interval time.Duration
}

func NewOutbox(db *sql.DB, logger *slog.Logger) *Outbox {
	if logger == nil {
		panic("generation.NewOutbox: logger 未注入")
	}
	return &Outbox{
		db:       db,
		logger:   logger.With(slog.String("component", "generation_outbox")),
		interval: 500 * time.Millisecond,
	}
}

func (o *Outbox) Run(ctx context.Context) error {
	ticker := time.NewTicker(o.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			for i := 0; i < 32; i++ {
				published, err := o.publishOne(ctx)
				if err != nil {
					o.logFailure(ctx, err)
					var stopped *backoffFailure
					if errors.As(err, &stopped) {
						break
					}
					var failed *deliveryFailure
					if errors.As(err, &failed) {
						// 退避已在独立事务提交，本条 next_attempt_at 推后，后续事件可以继续。
						continue
					}
					break
				}
				if !published {
					break
				}
			}
		}
	}
}

func (o *Outbox) logFailure(ctx context.Context, err error) {
	var stopped *backoffFailure
	if errors.As(err, &stopped) {
		o.logger.ErrorContext(ctx, "发件箱退避失败",
			slog.Int64("outbox_id", stopped.delivery.id),
			slog.String("job_id", stopped.delivery.aggregateID),
			slog.String("event_type", stopped.delivery.eventType),
			slog.String("aggregate_id", stopped.delivery.aggregateID),
			"err", stopped.err,
			"delivery_err", stopped.delivery.err,
		)
		return
	}
	var failed *deliveryFailure
	if errors.As(err, &failed) {
		o.logger.ErrorContext(ctx, "发件箱投递失败",
			slog.Int64("outbox_id", failed.id),
			slog.String("job_id", failed.aggregateID),
			slog.String("event_type", failed.eventType),
			slog.String("aggregate_id", failed.aggregateID),
			"err", failed.err,
		)
		return
	}
	o.logger.ErrorContext(ctx, "发件箱投递失败", "err", err)
}

func (o *Outbox) publishOne(ctx context.Context) (bool, error) {
	var published bool
	err := withTx(ctx, o.db, func(tx *sql.Tx) error {
		q := sqlcdb.New(tx)
		if _, err := q.LockGenerationDispatch(ctx); err != nil {
			return err
		}
		row, err := q.NextPendingGenerationOutbox(ctx)
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		if err := deliverOutbox(ctx, q, row); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return err
			}
			return &deliveryFailure{
				id:          row.ID,
				eventType:   row.EventType,
				aggregateID: row.AggregateID,
				err:         err,
			}
		}
		published = true
		return nil
	})
	var failed *deliveryFailure
	if errors.As(err, &failed) {
		if berr := o.backoff(ctx, failed); berr != nil {
			return false, &backoffFailure{delivery: failed, err: berr}
		}
		return false, err
	}
	return published, err
}

func deliverOutbox(ctx context.Context, q *sqlcdb.Queries, row sqlcdb.NextPendingGenerationOutboxRow) error {
	switch row.EventType {
	case EventGenerationAccepted:
		if err := q.MarkJobsReadyBySubmission(ctx, row.AggregateID); err != nil {
			return err
		}
	case "MEDIA_JOB_ACCEPTED":
		if err := q.MarkMediaJobWaiting(ctx, sqlcdb.MarkMediaJobWaitingParams{
			Phase: "等待处理",
			ID:    row.AggregateID,
		}); err != nil {
			return err
		}
	case "GENERATION_CHANGED":
		// 本切片不写 event 流，仅确认投递。
	default:
		return fmt.Errorf("发件箱出现未登记的事件类型: %s", row.EventType)
	}
	return q.MarkOutboxDelivered(ctx, row.ID)
}

func (o *Outbox) backoff(ctx context.Context, failed *deliveryFailure) error {
	return withTx(ctx, o.db, func(tx *sql.Tx) error {
		q := sqlcdb.New(tx)
		if _, err := q.LockGenerationDispatch(ctx); err != nil {
			return err
		}
		n, err := q.BackoffOutbox(ctx, sqlcdb.BackoffOutboxParams{
			LastError: sql.NullString{String: outboxErrorText(failed.err), Valid: true},
			ID:        failed.id,
		})
		if err != nil {
			return err
		}
		if n != 1 {
			return fmt.Errorf("待投递记录 %d 未更新（影响行数 %d）", failed.id, n)
		}
		return nil
	})
}

func outboxErrorText(err error) string {
	msg := "事件投递失败：" + err.Error()
	runes := []rune(msg)
	if len(runes) > 500 {
		return string(runes[:500])
	}
	return msg
}
