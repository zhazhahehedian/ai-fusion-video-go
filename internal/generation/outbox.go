package generation

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/sqlcdb"
)

type Outbox struct {
	db       *sql.DB
	interval time.Duration
}

func NewOutbox(db *sql.DB) *Outbox {
	return &Outbox{db: db, interval: 500 * time.Millisecond}
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
					log.Printf("generation outbox: %v", err)
					break
				}
				if !published {
					break
				}
			}
		}
	}
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
			return q.BackoffOutbox(ctx, sqlcdb.BackoffOutboxParams{
				LastError: sql.NullString{String: "未登记的事件类型: " + row.EventType, Valid: true},
				ID:        row.ID,
			})
		}
		if err := q.MarkOutboxDelivered(ctx, row.ID); err != nil {
			return err
		}
		published = true
		return nil
	})
	return published, err
}
