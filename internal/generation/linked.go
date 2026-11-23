package generation

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/apperr"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/sqlcdb"
)

// SubmissionRef 是工作者写回业务表时需要的受理摘要。
type SubmissionRef struct {
	ID           string
	Kind         string
	TaskID       int64
	PublicTaskID string
}

type stopLinked struct {
	code string
	msg  string
}

func (e *stopLinked) Error() string { return e.msg }

func (r *Repository) LoadSubmission(ctx context.Context, id string) (SubmissionRef, error) {
	row, err := sqlcdb.New(r.db).GetGenerationSubmissionByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SubmissionRef{}, apperr.NotFound("生成受理记录不存在")
		}
		return SubmissionRef{}, err
	}
	return submissionRef(row), nil
}

func submissionRef(row sqlcdb.AfvGenerationSubmission) SubmissionRef {
	return SubmissionRef{
		ID:           row.ID,
		Kind:         row.Kind,
		TaskID:       row.TaskID,
		PublicTaskID: row.PublicTaskID,
	}
}

func linkedKind(kind string) bool {
	return kind == KindImage || kind == KindVideo
}

// MarkTaskProcessing 把仍在排队或处理中的业务任务标成处理中。
func (r *Repository) MarkTaskProcessing(ctx context.Context, kind string, taskID int64) error {
	if !linkedKind(kind) {
		return nil
	}
	return withTx(ctx, r.db, func(tx *sql.Tx) error {
		q := sqlcdb.New(tx)
		var (
			n   int64
			err error
		)
		switch kind {
		case KindImage:
			n, err = q.MarkImageTaskProcessing(ctx, taskID)
		case KindVideo:
			n, err = q.MarkVideoTaskProcessing(ctx, taskID)
		}
		if err != nil {
			return err
		}
		if n != 1 {
			return apperr.Conflict("业务任务已不在排队或处理中")
		}
		return nil
	})
}

// SucceedLinked 把 stub 结果写入图片或视频条目，并把任务标成完成。
// 条目已不能写回时，把远端作业记为失败，不留下半成功的条目。
func (r *Repository) SucceedLinked(ctx context.Context, claim Claim) error {
	var stopped *stopLinked
	err := withTx(ctx, r.db, func(tx *sql.Tx) error {
		q := sqlcdb.New(tx)
		job, err := r.leased(ctx, q, claim, true)
		if err != nil {
			return err
		}
		row, err := q.GetGenerationSubmissionByID(ctx, job.SubmissionID)
		if err != nil {
			return err
		}
		sub := submissionRef(row)
		if !linkedKind(sub.Kind) {
			return fmt.Errorf("受理类别 %s 没有业务条目可写回", sub.Kind)
		}
		input, err := parseStubInput(job.InputJSON)
		if err != nil {
			stopped = &stopLinked{code: "STUB_INPUT_INVALID", msg: "生成输入缺少结果条目，不能记为成功"}
			return stopped
		}
		platform := "stub-" + job.ID
		if job.ProviderTaskID != nil && *job.ProviderTaskID != "" {
			platform = *job.ProviderTaskID
		}
		switch sub.Kind {
		case KindImage:
			if err := r.finishImage(ctx, q, sub, input, platform); err != nil {
				return err
			}
		case KindVideo:
			if err := r.finishVideo(ctx, q, sub, input, platform); err != nil {
				return err
			}
		}
		result, err := json.Marshal(map[string]any{
			"state":    StateSucceeded,
			"provider": "stub",
			"outputs":  stubOutputs(sub.Kind, sub.PublicTaskID, input.ResultItemIDs),
		})
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
	if errors.As(err, &stopped) {
		return r.Fail(ctx, claim, StateFailed, stopped.code, stopped.msg)
	}
	return err
}

func (r *Repository) finishImage(ctx context.Context, q *sqlcdb.Queries, sub SubmissionRef, input stubJobInput, platform string) error {
	if _, err := q.GetImageTaskByID(ctx, sub.TaskID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &stopLinked{code: "TASK_DELETED", msg: "生图任务已删除，不能记为成功"}
		}
		return err
	}
	width, height := nullInt32Value(input.Width), nullInt32Value(input.Height)
	for _, itemID := range input.ResultItemIDs {
		url := stubImageURL(sub.PublicTaskID, itemID)
		n, err := q.FinishImageItemStub(ctx, sqlcdb.FinishImageItemStubParams{
			PlatformTaskID: sql.NullString{String: platform, Valid: true},
			ImageUrl:       sql.NullString{String: url, Valid: true},
			ThumbnailUrl:   sql.NullString{String: url, Valid: true},
			Width:          width,
			Height:         height,
			ID:             itemID,
			TaskID:         sub.TaskID,
		})
		if err != nil {
			return err
		}
		if n != 1 {
			return &stopLinked{code: "ITEM_NOT_OPEN", msg: "生图条目已不在生成中，不能记为成功"}
		}
	}
	n, err := q.CompleteImageTaskStub(ctx, sqlcdb.CompleteImageTaskStubParams{
		SuccessCount: sql.NullInt32{Int32: int32(len(input.ResultItemIDs)), Valid: true},
		ID:           sub.TaskID,
	})
	if err != nil {
		return err
	}
	if n != 1 {
		return &stopLinked{code: "TASK_NOT_OPEN", msg: "生图任务已不在排队或处理中，不能记为成功"}
	}
	return nil
}

func (r *Repository) finishVideo(ctx context.Context, q *sqlcdb.Queries, sub SubmissionRef, input stubJobInput, platform string) error {
	if _, err := q.GetVideoTaskByID(ctx, sub.TaskID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &stopLinked{code: "TASK_DELETED", msg: "生视频任务已删除，不能记为成功"}
		}
		return err
	}
	for _, itemID := range input.ResultItemIDs {
		n, err := q.FinishVideoItemStub(ctx, sqlcdb.FinishVideoItemStubParams{
			PlatformTaskID: sql.NullString{String: platform, Valid: true},
			VideoUrl:       sql.NullString{String: stubVideoURL(sub.PublicTaskID, itemID), Valid: true},
			CoverUrl:       sql.NullString{String: stubVideoCoverURL(sub.PublicTaskID, itemID), Valid: true},
			Duration:       nullInt32Value(input.Duration),
			ID:             itemID,
			TaskID:         sub.TaskID,
		})
		if err != nil {
			return err
		}
		if n != 1 {
			return &stopLinked{code: "ITEM_NOT_OPEN", msg: "生视频条目已不在生成中，不能记为成功"}
		}
	}
	n, err := q.CompleteVideoTaskStub(ctx, sqlcdb.CompleteVideoTaskStubParams{
		SuccessCount: sql.NullInt32{Int32: int32(len(input.ResultItemIDs)), Valid: true},
		ID:           sub.TaskID,
	})
	if err != nil {
		return err
	}
	if n != 1 {
		return &stopLinked{code: "TASK_NOT_OPEN", msg: "生视频任务已不在排队或处理中，不能记为成功"}
	}
	return nil
}

func closeLinkedTask(ctx context.Context, q *sqlcdb.Queries, submissionID, message string) error {
	row, err := q.GetGenerationSubmissionByID(ctx, submissionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	msg := sql.NullString{String: message, Valid: message != ""}
	switch row.Kind {
	case KindImage:
		if _, err := q.FailImageItemsPending(ctx, sqlcdb.FailImageItemsPendingParams{
			ErrorMsg: msg,
			TaskID:   row.TaskID,
		}); err != nil {
			return err
		}
		_, err = q.FailImageTaskOpen(ctx, sqlcdb.FailImageTaskOpenParams{
			ErrorMsg: msg,
			ID:       row.TaskID,
		})
		return err
	case KindVideo:
		if _, err := q.FailVideoItemsPending(ctx, sqlcdb.FailVideoItemsPendingParams{
			ErrorMsg: msg,
			TaskID:   row.TaskID,
		}); err != nil {
			return err
		}
		_, err = q.FailVideoTaskOpen(ctx, sqlcdb.FailVideoTaskOpenParams{
			ErrorMsg: msg,
			ID:       row.TaskID,
		})
		return err
	default:
		return nil
	}
}

func parseStubInput(raw json.RawMessage) (stubJobInput, error) {
	var input stubJobInput
	if err := json.Unmarshal(raw, &input); err != nil {
		return stubJobInput{}, err
	}
	if len(input.ResultItemIDs) == 0 {
		return stubJobInput{}, errors.New("缺少结果条目")
	}
	return input, nil
}

func stubOutputs(kind, publicID string, itemIDs []int64) []map[string]any {
	out := make([]map[string]any, 0, len(itemIDs))
	for _, id := range itemIDs {
		url := stubImageURL(publicID, id)
		if kind == KindVideo {
			url = stubVideoURL(publicID, id)
		}
		out = append(out, map[string]any{"itemId": id, "url": url})
	}
	return out
}

func stubImageURL(publicID string, itemID int64) string {
	return fmt.Sprintf("stub://image/%s/%d", publicID, itemID)
}

func stubVideoURL(publicID string, itemID int64) string {
	return fmt.Sprintf("stub://video/%s/%d", publicID, itemID)
}

func stubVideoCoverURL(publicID string, itemID int64) string {
	return fmt.Sprintf("stub://video/%s/%d/cover", publicID, itemID)
}

func nullInt32Value(v *int32) sql.NullInt32 {
	if v == nil {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: *v, Valid: true}
}

func providerTaskID(kind, jobID string) string {
	if linkedKind(kind) {
		return "stub-" + jobID
	}
	return "echo-" + jobID
}
