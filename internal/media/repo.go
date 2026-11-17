package media

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/apperr"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/sqlcdb"
)

var ErrLeaseLost = apperr.Conflict("LEASE_LOST")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Claim(ctx context.Context, owner string) (*Job, error) {
	var claimed *Job
	err := withTx(ctx, r.db, func(tx *sql.Tx) error {
		q := sqlcdb.New(tx)
		if _, err := q.LockGenerationDispatch(ctx); err != nil {
			return err
		}
		cand, err := q.SelectMediaClaimCandidate(ctx)
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		phase := "处理中"
		n, err := q.ClaimMediaJob(ctx, sqlcdb.ClaimMediaJobParams{
			Phase:           phase,
			OwnerInstanceID: sql.NullString{String: owner, Valid: true},
			ID:              cand.ID,
		})
		if err != nil {
			return err
		}
		if n != 1 {
			return nil
		}
		row, err := q.GetMediaJobByID(ctx, cand.ID)
		if err != nil {
			return err
		}
		job := mapJob(row)
		claimed = &job
		return nil
	})
	return claimed, err
}

func (r *Repository) Heartbeat(ctx context.Context, claim Claim) bool {
	n, err := sqlcdb.New(r.db).HeartbeatMediaJob(ctx, sqlcdb.HeartbeatMediaJobParams{
		ID:              claim.ID,
		OwnerInstanceID: sql.NullString{String: claim.Owner, Valid: true},
		OwnerEpoch:      claim.Epoch,
	})
	return err == nil && n == 1
}

func (r *Repository) Check(ctx context.Context, claim Claim) (Job, error) {
	var job Job
	err := withTx(ctx, r.db, func(tx *sql.Tx) error {
		q := sqlcdb.New(tx)
		row, err := q.SelectLeasedMediaJobForUpdate(ctx, sqlcdb.SelectLeasedMediaJobForUpdateParams{
			ID:              claim.ID,
			OwnerInstanceID: sql.NullString{String: claim.Owner, Valid: true},
			OwnerEpoch:      claim.Epoch,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrLeaseLost
			}
			return err
		}
		job = mapJob(row)
		if job.CancelRequested {
			return ErrLeaseLost
		}
		return nil
	})
	return job, err
}

func (r *Repository) Checkpoint(ctx context.Context, claim Claim, checkpoint json.RawMessage, phase string) error {
	return withTx(ctx, r.db, func(tx *sql.Tx) error {
		q := sqlcdb.New(tx)
		if _, err := q.SelectLeasedMediaJobForUpdate(ctx, sqlcdb.SelectLeasedMediaJobForUpdateParams{
			ID:              claim.ID,
			OwnerInstanceID: sql.NullString{String: claim.Owner, Valid: true},
			OwnerEpoch:      claim.Epoch,
		}); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrLeaseLost
			}
			return err
		}
		return q.CheckpointMediaJob(ctx, sqlcdb.CheckpointMediaJobParams{
			CheckpointJson: checkpoint,
			Phase:          phase,
			ID:             claim.ID,
		})
	})
}

func (r *Repository) Succeed(ctx context.Context, claim Claim, checkpoint json.RawMessage, phase string) error {
	return r.finish(ctx, claim, StatusSucceeded, phase, checkpoint, "")
}

func (r *Repository) Fail(ctx context.Context, claim Claim, phase, errMsg string) error {
	return r.finish(ctx, claim, StatusFailed, phase, json.RawMessage("{}"), errMsg)
}

func (r *Repository) Cancel(ctx context.Context, claim Claim, phase, errMsg string) error {
	return r.finish(ctx, claim, StatusCancelled, phase, json.RawMessage("{}"), errMsg)
}

func (r *Repository) finish(ctx context.Context, claim Claim, status, phase string, checkpoint json.RawMessage, errMsg string) error {
	return withTx(ctx, r.db, func(tx *sql.Tx) error {
		q := sqlcdb.New(tx)
		if _, err := q.SelectLeasedMediaJobForUpdate(ctx, sqlcdb.SelectLeasedMediaJobForUpdateParams{
			ID:              claim.ID,
			OwnerInstanceID: sql.NullString{String: claim.Owner, Valid: true},
			OwnerEpoch:      claim.Epoch,
		}); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrLeaseLost
			}
			return err
		}
		var msg sql.NullString
		if errMsg != "" {
			msg = sql.NullString{String: errMsg, Valid: true}
		}
		return q.FinishMediaJob(ctx, sqlcdb.FinishMediaJobParams{
			Status:         status,
			Phase:          phase,
			CheckpointJson: checkpoint,
			ErrorMsg:       msg,
			ID:             claim.ID,
		})
	})
}

func mapJob(row sqlcdb.AfvMediaJob) Job {
	job := Job{
		ID:             row.ID,
		Kind:           row.Kind,
		ActorUserID:    row.ActorUserID,
		RequestID:      row.RequestID,
		RequestHash:    row.RequestHash,
		InputJSON:      row.InputJson,
		CheckpointJSON: row.CheckpointJson,
		Status:         row.Status,
		Phase:          row.Phase,
		OwnerEpoch:     row.OwnerEpoch,
		CancelRequested: row.CancelRequested,
		Revision:       row.Revision,
		DeadlineAt:     row.DeadlineAt,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}
	if row.ErrorMsg.Valid {
		v := row.ErrorMsg.String
		job.ErrorMsg = &v
	}
	if row.OwnerInstanceID.Valid {
		v := row.OwnerInstanceID.String
		job.OwnerInstanceID = &v
	}
	return job
}
