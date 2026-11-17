package media

import (
	"context"
	"database/sql"
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

type EnqueueRequest struct {
	RequestID string
	Kind      string
}

func (s *Service) Enqueue(ctx context.Context, actorUserID int64, req EnqueueRequest) (EnqueueResult, error) {
	requestID := strings.TrimSpace(req.RequestID)
	kind := strings.TrimSpace(req.Kind)
	if requestID == "" {
		return EnqueueResult{}, apperr.BadRequest("requestId 不能为空")
	}
	if kind == "" {
		kind = KindFFmpegVersion
	}
	if kind != KindFFmpegVersion {
		return EnqueueResult{}, apperr.BadRequest("暂不支持的 media kind: " + kind)
	}

	input := map[string]any{"action": "version"}
	hash := requestHash(kind, actorUserID, input)

	var out EnqueueResult
	err := withTx(ctx, s.db, func(tx *sql.Tx) error {
		q := sqlcdb.New(tx)
		existing, err := q.GetMediaJobByActorRequest(ctx, sqlcdb.GetMediaJobByActorRequestParams{
			ActorUserID: actorUserID,
			RequestID:   requestID,
		})
		if err == nil {
			if existing.RequestHash != hash {
				return apperr.Conflict("同一请求 ID 不能提交不同的媒体命令")
			}
			out = EnqueueResult{
				ID:        existing.ID,
				RequestID: existing.RequestID,
				Kind:      existing.Kind,
				Status:    existing.Status,
				Replay:    true,
			}
			return nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		id := newUUID()
		inputJSON, err := json.Marshal(input)
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
		payload, err := json.Marshal(map[string]any{
			"jobId":  id,
			"status": StatusQueued,
		})
		if err != nil {
			return err
		}

		if err := q.InsertMediaJob(ctx, sqlcdb.InsertMediaJobParams{
			ID:             id,
			Kind:           kind,
			ProjectID:      sql.NullInt64{},
			TargetID:       sql.NullInt64{},
			TargetVersion:  sql.NullInt32{},
			ActorUserID:    actorUserID,
			RequestID:      requestID,
			RequestHash:    hash,
			ContextJson:    contextJSON,
			InputJson:      inputJSON,
			CheckpointJson: json.RawMessage("{}"),
			Phase:          "等待处理",
		}); err != nil {
			return err
		}
		if err := q.InsertAsyncOutbox(ctx, sqlcdb.InsertAsyncOutboxParams{
			EventKey:    "media:" + id + ":1",
			EventType:   EventMediaJobAccepted,
			AggregateID: id,
			PayloadJson: payload,
		}); err != nil {
			return err
		}

		out = EnqueueResult{
			ID:        id,
			RequestID: requestID,
			Kind:      kind,
			Status:    StatusQueued,
			Replay:    false,
		}
		return nil
	})
	return out, err
}

func (s *Service) Get(ctx context.Context, actorUserID int64, id string) (JobView, error) {
	row, err := sqlcdb.New(s.db).GetMediaJobByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return JobView{}, apperr.NotFound("媒体任务不存在")
		}
		return JobView{}, err
	}
	if row.ActorUserID != actorUserID {
		return JobView{}, apperr.Forbidden("无权查看该媒体任务")
	}
	view := JobView{
		ID:             row.ID,
		Kind:           row.Kind,
		Status:         row.Status,
		Phase:          row.Phase,
		CheckpointJSON: row.CheckpointJson,
		Revision:       row.Revision,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}
	if row.ErrorMsg.Valid {
		v := row.ErrorMsg.String
		view.ErrorMsg = &v
	}
	return view, nil
}
