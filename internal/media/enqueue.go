package media

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
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

		// 先插入，唯一键冲突时保持原行。并发的同一 requestId 在锁读时汇合，不再把 1062 变成 500。
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
		existing, err := q.LockMediaJobByActorRequest(ctx, sqlcdb.LockMediaJobByActorRequestParams{
			ActorUserID: actorUserID,
			RequestID:   requestID,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("媒体任务未写入，requestId=%s", requestID)
			}
			return err
		}
		if existing.RequestHash != hash {
			return apperr.Conflict("同一请求 ID 不能提交不同的媒体命令")
		}
		if existing.ID != id {
			out = EnqueueResult{
				ID:        existing.ID,
				RequestID: existing.RequestID,
				Kind:      existing.Kind,
				Status:    existing.Status,
				Replay:    true,
			}
			return nil
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
