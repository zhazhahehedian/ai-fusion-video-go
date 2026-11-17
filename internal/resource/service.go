package resource

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/apperr"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/sqlcdb"
)

type Service struct {
	db       *sql.DB
	registry *Registry
}

func NewService(db *sql.DB, registry *Registry) *Service {
	return &Service{db: db, registry: registry}
}

func (s *Service) Execute(ctx context.Context, cmd Command, rc Context) (Receipt, error) {
	requestID, err := ValidateRequestID(cmd.RequestID())
	if err != nil {
		return Receipt{}, err
	}
	handler, ok := s.registry.Get(cmd.ResourceType(), cmd.Operation())
	if !ok {
		return Receipt{}, apperr.BadRequest("此业务类型未登记该命令")
	}
	if rc.Origin == "" {
		rc.Origin = OriginUser
	}
	hash, err := HashCommand(cmd, rc.ActorUserID, rc.Origin)
	if err != nil {
		return Receipt{}, err
	}
	requestJSON, err := json.Marshal(map[string]any{
		"schemaVersion": 1,
		"resourceType":  cmd.ResourceType(),
		"operation":     cmd.Operation(),
		"input":         cmd,
	})
	if err != nil {
		return Receipt{}, err
	}

	commandID, err := NewCommandID()
	if err != nil {
		return Receipt{}, err
	}

	var out Receipt
	err = withTx(ctx, s.db, func(tx *sql.Tx) error {
		q := sqlcdb.New(tx)
		if err := q.AdmitResourceCommand(ctx, sqlcdb.AdmitResourceCommandParams{
			ID:          commandID,
			ActorUserID: rc.ActorUserID,
			RequestID:   requestID,
			RequestHash: hash,
			Origin:      string(rc.Origin),
			RequestJson: requestJSON,
		}); err != nil {
			return err
		}
		row, err := q.LockResourceCommandByRequest(ctx, sqlcdb.LockResourceCommandByRequestParams{
			ActorUserID: rc.ActorUserID,
			RequestID:   requestID,
		})
		if err != nil {
			return err
		}
		matching := row.RequestHash == hash
		var previous *Receipt
		if raw := asJSONBytes(row.ResultJson); len(raw) > 0 {
			var receipt Receipt
			if err := json.Unmarshal(raw, &receipt); err != nil {
				return err
			}
			previous = &receipt
		}
		if _, err := handler.Authorize(ctx, q, cmd, rc.ActorUserID, previous); err != nil {
			return err
		}
		if !matching {
			return apperr.Conflict("同一请求 ID 不能提交不同的业务命令")
		}
		if previous != nil {
			out = *previous
			return nil
		}
		if row.ID != commandID {
			return apperr.Conflict("业务命令存在未提交的持久化记录")
		}
		mutation, err := handler.Apply(ctx, q, cmd, rc.ActorUserID)
		if err != nil {
			return err
		}
		receipt := Receipt{
			CommandID: commandID,
			RequestID: requestID,
			Result:    mutation.Result,
			Changes:   mutation.Changes,
		}
		if err := commit(ctx, q, receipt, mutation); err != nil {
			return err
		}
		out = receipt
		return nil
	})
	return out, err
}

func asJSONBytes(v any) []byte {
	switch x := v.(type) {
	case nil:
		return nil
	case []byte:
		if len(x) == 0 {
			return nil
		}
		return x
	case string:
		if x == "" {
			return nil
		}
		return []byte(x)
	default:
		return nil
	}
}

func commit(ctx context.Context, q *sqlcdb.Queries, receipt Receipt, mutation Mutation) error {
	updated, err := q.CommitResourceCommand(ctx, sqlcdb.CommitResourceCommandParams{
		ResultJson: MustJSON(receipt),
		UndoJson:   MustJSON(mutation.PreviousValues),
		ID:         receipt.CommandID,
	})
	if err != nil {
		return err
	}
	if updated != 1 {
		return apperr.Conflict("业务命令重复提交")
	}
	for _, change := range mutation.Changes {
		fields := change.ChangedFields
		if fields == nil {
			fields = []string{}
		}
		var projectID sql.NullInt64
		if change.ProjectID != nil {
			projectID = sql.NullInt64{Int64: *change.ProjectID, Valid: true}
		}
		afterJSON := MustJSON(change.After)
		if afterJSON == nil {
			afterJSON = json.RawMessage("{}")
		}
		if err := q.InsertResourceRevision(ctx, sqlcdb.InsertResourceRevisionParams{
			CommandID:       receipt.CommandID,
			ResourceType:    string(change.ResourceType),
			ResourceID:      change.ResourceID,
			ProjectID:       projectID,
			ResourceVersion: change.Version,
			BeforeJson:      MustJSON(change.Before),
			AfterJson:       afterJSON,
			ChangedFields:   MustJSON(fields),
		}); err != nil {
			return err
		}
	}
	return nil
}
