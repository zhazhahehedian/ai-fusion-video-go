package generation

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

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

// Now 读取数据库 UTC 时钟，与 Java GenerationJobRepository.now 一致。
func (r *Repository) Now(ctx context.Context) (time.Time, error) {
	return sqlcdb.New(r.db).DatabaseNow(ctx)
}

func (r *Repository) Claim(ctx context.Context, owner string) (*Job, error) {
	var claimed *Job
	err := withTx(ctx, r.db, func(tx *sql.Tx) error {
		q := sqlcdb.New(tx)
		if _, err := q.LockGenerationDispatch(ctx); err != nil {
			return err
		}
		candidates, err := q.SelectClaimCandidates(ctx)
		if err != nil {
			return err
		}
		for _, c := range candidates {
			needsPermit := !c.PermitHeld && !c.CancelRequested && c.Phase != PhaseIngesting
			if needsPermit {
				held, err := q.CountPermitHeldByActor(ctx, c.ActorUserID)
				if err != nil {
					return err
				}
				if held >= MaxUserRemoteJobs {
					if err := q.DeferJobForCapacity(ctx, c.ID); err != nil {
						return err
					}
					continue
				}
			}
			permitHeld := c.PermitHeld || needsPermit
			n, err := q.ClaimRemoteGenerationJob(ctx, sqlcdb.ClaimRemoteGenerationJobParams{
				OwnerInstanceID: sql.NullString{String: owner, Valid: true},
				PermitHeld:      permitHeld,
				ID:              c.ID,
			})
			if err != nil {
				return err
			}
			if n != 1 {
				continue
			}
			row, err := q.GetRemoteGenerationJob(ctx, c.ID)
			if err != nil {
				return err
			}
			job := mapJob(row)
			claimed = &job
			return nil
		}
		return nil
	})
	return claimed, err
}

func (r *Repository) Heartbeat(ctx context.Context, claim Claim) error {
	q := sqlcdb.New(r.db)
	n, err := q.HeartbeatRemoteGenerationJob(ctx, sqlcdb.HeartbeatRemoteGenerationJobParams{
		ID:              claim.ID,
		OwnerInstanceID: sql.NullString{String: claim.Owner, Valid: true},
		OwnerEpoch:      claim.Epoch,
	})
	if err != nil {
		return fmt.Errorf("续期生成租约 job_id=%s: %w", claim.ID, err)
	}
	if n != 1 {
		return fmt.Errorf("续期生成租约未命中 job_id=%s owner_epoch=%d", claim.ID, claim.Epoch)
	}
	return nil
}

func (r *Repository) WaitForNextPoll(ctx context.Context, claim Claim, seconds int, phase string, errCode, errMsg *string) error {
	if seconds < 1 {
		seconds = 1
	}
	if seconds > 60 {
		seconds = 60
	}
	return withTx(ctx, r.db, func(tx *sql.Tx) error {
		q := sqlcdb.New(tx)
		job, err := r.leased(ctx, q, claim, true)
		if err != nil {
			return err
		}
		errorsCount := int32(0)
		if errCode != nil {
			errorsCount = job.Errors + 1
		}
		var code, msg sql.NullString
		if errCode != nil {
			code = sql.NullString{String: *errCode, Valid: true}
		}
		if errMsg != nil {
			msg = sql.NullString{String: *errMsg, Valid: true}
		}
		return q.ScheduleJobPoll(ctx, sqlcdb.ScheduleJobPollParams{
			Phase:             phase,
			DelaySeconds:      int32(seconds),
			ConsecutiveErrors: errorsCount,
			ErrorCode:         code,
			ErrorMessage:      msg,
			ID:                job.ID,
		})
	})
}

func (r *Repository) PrepareAttempt(ctx context.Context, claim Claim, attemptID string, requestJSON json.RawMessage) error {
	return withTx(ctx, r.db, func(tx *sql.Tx) error {
		q := sqlcdb.New(tx)
		job, err := r.leased(ctx, q, claim, false)
		if err != nil {
			return err
		}
		if job.ActiveAttemptID != nil && *job.ActiveAttemptID != "" {
			return nil
		}
		attemptNo := int32(1)
		if err := q.InsertGenerationAttempt(ctx, sqlcdb.InsertGenerationAttemptParams{
			ID:          attemptID,
			JobID:       job.ID,
			AttemptNo:   attemptNo,
			RequestJson: requestJSON,
		}); err != nil {
			return err
		}
		return q.PrepareJobAttempt(ctx, sqlcdb.PrepareJobAttemptParams{
			ActiveAttemptID: sql.NullString{String: attemptID, Valid: true},
			ID:              job.ID,
		})
	})
}

func (r *Repository) BeginSubmission(ctx context.Context, claim Claim, attemptID, providerTaskID string) error {
	return withTx(ctx, r.db, func(tx *sql.Tx) error {
		q := sqlcdb.New(tx)
		job, err := r.leased(ctx, q, claim, false)
		if err != nil {
			return err
		}
		if err := q.AcknowledgeGenerationAttempt(ctx, sqlcdb.AcknowledgeGenerationAttemptParams{
			ProviderTaskID: sql.NullString{String: providerTaskID, Valid: true},
			ID:             attemptID,
		}); err != nil {
			return err
		}
		if err := q.BeginJobSubmission(ctx, sqlcdb.BeginJobSubmissionParams{
			ProviderTaskID: sql.NullString{String: providerTaskID, Valid: true},
			ID:             job.ID,
		}); err != nil {
			return err
		}
		_, err = q.UpdateGenerationSubmissionState(ctx, sqlcdb.UpdateGenerationSubmissionStateParams{
			State: StateRunning,
			ID:    job.SubmissionID,
		})
		return err
	})
}

func (r *Repository) Succeed(ctx context.Context, claim Claim, result json.RawMessage) error {
	return withTx(ctx, r.db, func(tx *sql.Tx) error {
		q := sqlcdb.New(tx)
		job, err := r.leased(ctx, q, claim, true)
		if err != nil {
			return err
		}
		if err := q.FinishRemoteGenerationJob(ctx, sqlcdb.FinishRemoteGenerationJobParams{
			State:        StateSucceeded,
			ResultJson:   result,
			ErrorCode:    sql.NullString{},
			ErrorMessage: sql.NullString{},
			ID:           job.ID,
		}); err != nil {
			return err
		}
		_, err = q.UpdateGenerationSubmissionState(ctx, sqlcdb.UpdateGenerationSubmissionStateParams{
			State: StateSucceeded,
			ID:    job.SubmissionID,
		})
		return err
	})
}

func (r *Repository) Fail(ctx context.Context, claim Claim, state, code, message string) error {
	return withTx(ctx, r.db, func(tx *sql.Tx) error {
		q := sqlcdb.New(tx)
		job, err := r.leased(ctx, q, claim, true)
		if err != nil {
			return err
		}
		if err := q.FinishRemoteGenerationJob(ctx, sqlcdb.FinishRemoteGenerationJobParams{
			State:        state,
			ResultJson:   nil,
			ErrorCode:    sql.NullString{String: code, Valid: code != ""},
			ErrorMessage: sql.NullString{String: message, Valid: message != ""},
			ID:           job.ID,
		}); err != nil {
			return err
		}
		_, err = q.UpdateGenerationSubmissionState(ctx, sqlcdb.UpdateGenerationSubmissionStateParams{
			State: state,
			ID:    job.SubmissionID,
		})
		return err
	})
}

func (r *Repository) Get(ctx context.Context, id string) (Job, error) {
	row, err := sqlcdb.New(r.db).GetRemoteGenerationJob(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Job{}, apperr.NotFound("远端生成记录不存在")
		}
		return Job{}, err
	}
	return mapJob(row), nil
}

func (r *Repository) leased(ctx context.Context, q *sqlcdb.Queries, claim Claim, allowCancel bool) (Job, error) {
	current, err := q.GetRemoteGenerationJob(ctx, claim.ID)
	if err != nil {
		return Job{}, err
	}
	if _, err := q.LockGenerationSubmissionByID(ctx, current.SubmissionID); err != nil {
		return Job{}, err
	}
	row, err := q.SelectLeasedJobForUpdate(ctx, sqlcdb.SelectLeasedJobForUpdateParams{
		ID:              claim.ID,
		OwnerInstanceID: sql.NullString{String: claim.Owner, Valid: true},
		OwnerEpoch:      claim.Epoch,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Job{}, ErrLeaseLost
		}
		return Job{}, err
	}
	job := mapJob(row)
	if !allowCancel && job.CancelRequested {
		return Job{}, ErrLeaseLost
	}
	return job, nil
}

func mapJob(row sqlcdb.AfvRemoteGenerationJob) Job {
	job := Job{
		ID:              row.ID,
		SubmissionID:    row.SubmissionID,
		ItemID:          row.ItemID,
		ItemIndex:       row.ItemIndex,
		ModelID:         row.ModelID,
		ActorUserID:     row.ActorUserID,
		State:           row.State,
		Phase:           row.Phase,
		InputJSON:       row.InputJson,
		ResultJSON:      row.ResultJson,
		CancelRequested: row.CancelRequested,
		PermitHeld:      row.PermitHeld,
		OwnerEpoch:      row.OwnerEpoch,
		NextPollAt:      row.NextPollAt,
		DeadlineAt:      row.DeadlineAt,
		PollAttempts:    row.PollAttempts,
		Errors:          row.ConsecutiveErrors,
		Revision:        row.Revision,
	}
	if row.ProjectID.Valid {
		v := row.ProjectID.Int64
		job.ProjectID = &v
	}
	if row.ProviderTaskID.Valid {
		v := row.ProviderTaskID.String
		job.ProviderTaskID = &v
	}
	if row.ActiveAttemptID.Valid {
		v := row.ActiveAttemptID.String
		job.ActiveAttemptID = &v
	}
	if row.OwnerInstanceID.Valid {
		v := row.OwnerInstanceID.String
		job.OwnerInstanceID = &v
	}
	if row.LeaseUntil.Valid {
		v := row.LeaseUntil.Time
		job.LeaseUntil = &v
	}
	if row.ErrorCode.Valid {
		v := row.ErrorCode.String
		job.ErrorCode = &v
	}
	if row.ErrorMessage.Valid {
		v := row.ErrorMessage.String
		job.ErrorMessage = &v
	}
	return job
}
