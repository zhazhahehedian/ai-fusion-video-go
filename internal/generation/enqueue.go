package generation

import (
	"context"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/apperr"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/sqlcdb"
)

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

type EchoRequest struct {
	RequestID string
	Prompt    string
}

func (s *Service) EnqueueEcho(ctx context.Context, actorUserID int64, req EchoRequest) (EchoEnqueueResult, error) {
	requestID := strings.TrimSpace(req.RequestID)
	prompt := strings.TrimSpace(req.Prompt)
	if requestID == "" {
		return EchoEnqueueResult{}, apperr.BadRequest("requestId 不能为空")
	}
	if prompt == "" {
		return EchoEnqueueResult{}, apperr.BadRequest("prompt 不能为空")
	}

	scope := ScopeKey(actorUserID)
	hash := RequestHash(KindEcho, prompt, actorUserID)

	var out EchoEnqueueResult
	err := withTx(ctx, s.db, func(tx *sql.Tx) error {
		q := sqlcdb.New(tx)
		existing, err := q.GetGenerationSubmissionByScopeRequest(ctx, sqlcdb.GetGenerationSubmissionByScopeRequestParams{
			ScopeKey:  scope,
			RequestID: requestID,
		})
		if err == nil {
			if existing.RequestHash != hash {
				return apperr.Conflict("同一请求 ID 不能提交不同的生成命令")
			}
			out = EchoEnqueueResult{
				ID:        existing.ID,
				RequestID: existing.RequestID,
				State:     existing.State,
				Kind:      existing.Kind,
				Replay:    true,
			}
			return nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		submissionID := newUUID()
		jobID := newUUID()
		taskID := taskIDFrom(submissionID)
		input, err := json.Marshal(map[string]any{"prompt": prompt})
		if err != nil {
			return err
		}
		contextJSON, err := json.Marshal(map[string]any{
			"origin":      "http",
			"actorUserId": actorUserID,
		})
		if err != nil {
			return err
		}
		modelJSON, err := json.Marshal(map[string]any{
			"id":   0,
			"name": "echo-stub",
		})
		if err != nil {
			return err
		}
		payload, err := json.Marshal(map[string]any{
			"submissionId": submissionID,
			"state":        StateQueued,
			"revision":     1,
		})
		if err != nil {
			return err
		}

		if err := q.InsertGenerationSubmission(ctx, sqlcdb.InsertGenerationSubmissionParams{
			ID:                submissionID,
			RequestID:         requestID,
			ScopeKey:          scope,
			RequestHash:       hash,
			Kind:              KindEcho,
			TaskID:            taskID,
			PublicTaskID:      "echo-" + submissionID,
			ActorUserID:       actorUserID,
			ActorTeamID:       sql.NullInt64{},
			ProjectID:         sql.NullInt64{},
			CanvasNodeRunID:   sql.NullString{},
			ContextJson:       contextJSON,
			ModelSnapshotJson: modelJSON,
			ApiConfigID:       0,
			ApiConfigHash:     "echo",
			State:             StateQueued,
		}); err != nil {
			return err
		}
		if err := q.InsertRemoteGenerationJob(ctx, sqlcdb.InsertRemoteGenerationJobParams{
			ID:           jobID,
			SubmissionID: submissionID,
			ItemID:       0,
			ItemIndex:    0,
			ModelID:      0,
			ActorUserID:  actorUserID,
			ProjectID:    sql.NullInt64{},
			InputJson:    input,
		}); err != nil {
			return err
		}
		if err := q.InsertAsyncOutbox(ctx, sqlcdb.InsertAsyncOutboxParams{
			EventKey:    "generation:" + submissionID + ":1",
			EventType:   EventGenerationAccepted,
			AggregateID: submissionID,
			PayloadJson: payload,
		}); err != nil {
			return err
		}

		out = EchoEnqueueResult{
			ID:        submissionID,
			RequestID: requestID,
			State:     StateQueued,
			Kind:      KindEcho,
			Replay:    false,
		}
		return nil
	})
	return out, err
}

func (s *Service) GetSubmission(ctx context.Context, actorUserID int64, id string) (SubmissionView, error) {
	q := sqlcdb.New(s.db)
	row, err := q.GetGenerationSubmissionByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SubmissionView{}, apperr.NotFound("生成受理记录不存在")
		}
		return SubmissionView{}, err
	}
	if row.ActorUserID != actorUserID {
		return SubmissionView{}, apperr.Forbidden("无权查看该生成任务")
	}
	jobs, err := q.ListJobsBySubmissionID(ctx, id)
	if err != nil {
		return SubmissionView{}, err
	}

	view := SubmissionView{
		ID:           row.ID,
		RequestID:    row.RequestID,
		Kind:         row.Kind,
		State:        row.State,
		Revision:     row.Revision,
		PublicTaskID: row.PublicTaskID,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
		Jobs:         make([]JobView, 0, len(jobs)),
	}
	if len(jobs) > 0 {
		var input struct {
			Prompt string `json:"prompt"`
		}
		_ = json.Unmarshal(jobs[0].InputJson, &input)
		view.Prompt = input.Prompt
	}
	for _, job := range jobs {
		jv := JobView{
			ID:           job.ID,
			State:        job.State,
			Phase:        job.Phase,
			PollAttempts: job.PollAttempts,
			Revision:     job.Revision,
			ResultJSON:   job.ResultJson,
		}
		if job.ProviderTaskID.Valid {
			v := job.ProviderTaskID.String
			jv.ProviderTaskID = &v
		}
		if job.ErrorCode.Valid {
			v := job.ErrorCode.String
			jv.ErrorCode = &v
		}
		if job.ErrorMessage.Valid {
			v := job.ErrorMessage.String
			jv.ErrorMessage = &v
		}
		view.Jobs = append(view.Jobs, jv)
	}
	return view, nil
}

func taskIDFrom(submissionID string) int64 {
	// 用 UUID 前 8 字节构造正数 task_id，避免依赖业务 task 表。
	raw, err := hex.DecodeString(strings.ReplaceAll(submissionID, "-", ""))
	if err != nil || len(raw) < 8 {
		return 1
	}
	v := int64(binary.BigEndian.Uint64(raw[:8]) & 0x7fffffffffffffff)
	if v == 0 {
		return 1
	}
	return v
}
