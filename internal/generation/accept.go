package generation

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/apperr"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/sqlcdb"
)

type createdTask struct {
	DBID     int64
	PublicID string
	ItemIDs  []int64
	Input    json.RawMessage
}

type acceptSpec struct {
	ActorUserID int64
	Kind        string
	RequestID   string
	Hash        string
	ProjectID   sql.NullInt64
	ModelID     int64
	Build       func(ctx context.Context, q *sqlcdb.Queries) (createdTask, error)
}

type replayHit struct {
	publicID string
}

func (e *replayHit) Error() string { return "generation replay" }

// accept 在一个事务里写入业务任务、条目、受理和远端作业。
// 同一 scope 与 requestId 已存在时回滚这次新建的任务行，返回原来的公开标识。
func (s *Service) accept(ctx context.Context, spec acceptSpec) (string, error) {
	var publicID string
	err := withTx(ctx, s.db, func(tx *sql.Tx) error {
		q := sqlcdb.New(tx)
		if err := requireOwned(ctx, q, spec.ProjectID, spec.ActorUserID); err != nil {
			return err
		}
		created, err := spec.Build(ctx, q)
		if err != nil {
			return err
		}
		submissionID := newUUID()
		if err := insertSubmission(ctx, q, spec, created, submissionID); err != nil {
			return err
		}
		existing, err := q.LockGenerationSubmissionByScopeRequest(ctx, sqlcdb.LockGenerationSubmissionByScopeRequestParams{
			ScopeKey:  ScopeKey(spec.ActorUserID),
			RequestID: spec.RequestID,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errors.New("生成受理未写入")
			}
			return err
		}
		if existing.RequestHash != spec.Hash || existing.Kind != spec.Kind {
			return apperr.Conflict("此请求标识已用于不同的生成内容")
		}
		if existing.ID != submissionID {
			if err := requireOwned(ctx, q, existing.ProjectID, spec.ActorUserID); err != nil {
				return err
			}
			return &replayHit{publicID: existing.PublicTaskID}
		}
		if len(created.ItemIDs) == 0 {
			return errors.New("生成任务没有结果条目")
		}
		if err := q.InsertRemoteGenerationJob(ctx, sqlcdb.InsertRemoteGenerationJobParams{
			ID:           newUUID(),
			SubmissionID: submissionID,
			ItemID:       created.ItemIDs[0],
			ItemIndex:    0,
			ModelID:      spec.ModelID,
			ActorUserID:  spec.ActorUserID,
			ProjectID:    spec.ProjectID,
			InputJson:    created.Input,
		}); err != nil {
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
		if err := q.InsertAsyncOutbox(ctx, sqlcdb.InsertAsyncOutboxParams{
			EventKey:    "generation:" + submissionID + ":1",
			EventType:   EventGenerationAccepted,
			AggregateID: submissionID,
			PayloadJson: payload,
		}); err != nil {
			return err
		}
		publicID = created.PublicID
		return nil
	})
	var replay *replayHit
	if errors.As(err, &replay) {
		return replay.publicID, nil
	}
	return publicID, err
}

func insertSubmission(ctx context.Context, q *sqlcdb.Queries, spec acceptSpec, created createdTask, submissionID string) error {
	contextJSON, err := json.Marshal(map[string]any{
		"origin":      "http",
		"actorUserId": spec.ActorUserID,
	})
	if err != nil {
		return err
	}
	modelJSON, err := json.Marshal(map[string]any{
		"id":       spec.ModelID,
		"name":     "stub",
		"provider": "stub",
	})
	if err != nil {
		return err
	}
	return q.InsertGenerationSubmission(ctx, sqlcdb.InsertGenerationSubmissionParams{
		ID:                submissionID,
		RequestID:         spec.RequestID,
		ScopeKey:          ScopeKey(spec.ActorUserID),
		RequestHash:       spec.Hash,
		Kind:              spec.Kind,
		TaskID:            created.DBID,
		PublicTaskID:      created.PublicID,
		ActorUserID:       spec.ActorUserID,
		ActorTeamID:       sql.NullInt64{},
		ProjectID:         spec.ProjectID,
		CanvasNodeRunID:   sql.NullString{},
		ContextJson:       contextJSON,
		ModelSnapshotJson: modelJSON,
		ApiConfigID:       0,
		ApiConfigHash:     "stub",
		State:             StateQueued,
	})
}

func requireOwned(ctx context.Context, q *sqlcdb.Queries, projectID sql.NullInt64, actorUserID int64) error {
	if !projectID.Valid {
		return nil
	}
	_, err := q.GetOwnedProjectByID(ctx, sqlcdb.GetOwnedProjectByIDParams{
		ID:      projectID.Int64,
		OwnerID: actorUserID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return apperr.Forbidden("无权访问该项目")
	}
	return err
}

func insertImageTask(ctx context.Context, q *sqlcdb.Queries, row sqlcdb.AfvImageTask) (createdTask, error) {
	publicID := newUUID()
	id, err := q.CreateImageTask(ctx, sqlcdb.CreateImageTaskParams{
		TaskID:            publicID,
		UserID:            row.UserID,
		ProjectID:         row.ProjectID,
		Prompt:            row.Prompt,
		PromptTemplateID:  row.PromptTemplateID,
		RefImageUrls:      row.RefImageUrls,
		Ratio:             row.Ratio,
		Resolution:        row.Resolution,
		AspectRatio:       row.AspectRatio,
		Width:             row.Width,
		Height:            row.Height,
		Count:             row.Count,
		Status:            row.Status,
		ErrorMsg:          row.ErrorMsg,
		ModelID:           row.ModelID,
		WorkflowVersionID: row.WorkflowVersionID,
		Category:          row.Category,
		OwnerID:           row.OwnerID,
	})
	if err != nil {
		return createdTask{}, err
	}
	itemIDs, err := insertPendingItems(ctx, row.Count.Int32, func(ctx context.Context) (int64, error) {
		return q.InsertImageItemPending(ctx, id)
	})
	if err != nil {
		return createdTask{}, err
	}
	input, err := json.Marshal(stubJobInput{
		Prompt:        nullString(row.Prompt),
		Count:         row.Count.Int32,
		ResultItemIDs: itemIDs,
		Kind:          KindImage,
		Width:         nullInt32Ptr(row.Width),
		Height:        nullInt32Ptr(row.Height),
	})
	if err != nil {
		return createdTask{}, err
	}
	return createdTask{DBID: id, PublicID: publicID, ItemIDs: itemIDs, Input: input}, nil
}

func insertVideoTask(ctx context.Context, q *sqlcdb.Queries, row sqlcdb.AfvVideoTask) (createdTask, error) {
	publicID := newUUID()
	id, err := q.CreateVideoTask(ctx, sqlcdb.CreateVideoTaskParams{
		TaskID:             publicID,
		UserID:             row.UserID,
		ProjectID:          row.ProjectID,
		Prompt:             row.Prompt,
		PromptTemplateID:   row.PromptTemplateID,
		GenerateMode:       row.GenerateMode,
		FirstFrameImageUrl: row.FirstFrameImageUrl,
		LastFrameImageUrl:  row.LastFrameImageUrl,
		ReferenceImageUrls: row.ReferenceImageUrls,
		ReferenceVideoUrls: row.ReferenceVideoUrls,
		ReferenceAudioUrls: row.ReferenceAudioUrls,
		Ratio:              row.Ratio,
		Resolution:         row.Resolution,
		Duration:           row.Duration,
		Watermark:          row.Watermark,
		GenerateAudio:      row.GenerateAudio,
		Seed:               row.Seed,
		CameraFixed:        row.CameraFixed,
		Count:              row.Count,
		Status:             row.Status,
		ErrorMsg:           row.ErrorMsg,
		ModelID:            row.ModelID,
		WorkflowVersionID:  row.WorkflowVersionID,
		Category:           row.Category,
		OwnerID:            row.OwnerID,
	})
	if err != nil {
		return createdTask{}, err
	}
	itemIDs, err := insertPendingItems(ctx, row.Count.Int32, func(ctx context.Context) (int64, error) {
		return q.InsertVideoItemPending(ctx, id)
	})
	if err != nil {
		return createdTask{}, err
	}
	input, err := json.Marshal(stubJobInput{
		Prompt:        nullString(row.Prompt),
		Count:         row.Count.Int32,
		ResultItemIDs: itemIDs,
		Kind:          KindVideo,
		Duration:      nullInt32Ptr(row.Duration),
	})
	if err != nil {
		return createdTask{}, err
	}
	return createdTask{DBID: id, PublicID: publicID, ItemIDs: itemIDs, Input: input}, nil
}

func insertPendingItems(ctx context.Context, count int32, insert func(context.Context) (int64, error)) ([]int64, error) {
	if count < 1 {
		return nil, apperr.BadRequest("count 整数范围无效")
	}
	ids := make([]int64, 0, count)
	for range count {
		id, err := insert(ctx)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

type stubJobInput struct {
	Prompt        string  `json:"prompt"`
	Count         int32   `json:"count"`
	ResultItemIDs []int64 `json:"_resultItemIds"`
	Kind          string  `json:"kind"`
	Width         *int32  `json:"width,omitempty"`
	Height        *int32  `json:"height,omitempty"`
	Duration      *int32  `json:"duration,omitempty"`
}
