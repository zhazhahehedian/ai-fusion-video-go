package resource

import (
	"context"
	"fmt"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/sqlcdb"
)

// Handler mirrors Java ResourceCommandHandler: authorize then apply inside one write tx.
type Handler interface {
	ResourceType() Type
	Operation() Operation
	Authorize(ctx context.Context, q *sqlcdb.Queries, cmd Command, actorUserID int64, previous *Receipt) (projectID *int64, err error)
	Apply(ctx context.Context, q *sqlcdb.Queries, cmd Command, actorUserID int64) (Mutation, error)
}

type handlerKey struct {
	resourceType Type
	operation    Operation
}

type Registry struct {
	handlers map[handlerKey]Handler
}

func NewRegistry(handlers ...Handler) (*Registry, error) {
	r := &Registry{handlers: make(map[handlerKey]Handler, len(handlers))}
	for _, h := range handlers {
		key := handlerKey{resourceType: h.ResourceType(), operation: h.Operation()}
		if _, exists := r.handlers[key]; exists {
			return nil, fmt.Errorf("业务写入处理器重复: %s/%s", key.resourceType, key.operation)
		}
		r.handlers[key] = h
	}
	return r, nil
}

func (r *Registry) Get(resourceType Type, op Operation) (Handler, bool) {
	h, ok := r.handlers[handlerKey{resourceType: resourceType, operation: op}]
	return h, ok
}
