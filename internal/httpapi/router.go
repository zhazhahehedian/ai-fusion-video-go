package httpapi

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
)

type Deps struct {
	DB    *sql.DB
	Redis *redis.Client
}

func NewRouter(deps Deps) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", deps.handleHealthz)
	return r
}

func (d Deps) handleHealthz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	mysqlOK := true
	mysqlMsg := "up"
	if err := d.DB.PingContext(ctx); err != nil {
		mysqlOK = false
		mysqlMsg = err.Error()
	}

	redisOK := true
	redisMsg := "up"
	if err := d.Redis.Ping(ctx).Err(); err != nil {
		redisOK = false
		redisMsg = err.Error()
	}

	payload := map[string]any{
		"status": "up",
		"mysql":  map[string]any{"ok": mysqlOK, "detail": mysqlMsg},
		"redis":  map[string]any{"ok": redisOK, "detail": redisMsg},
	}

	if !mysqlOK || !redisOK {
		payload["status"] = "degraded"
		write(w, http.StatusServiceUnavailable, Result{
			Code: 50300,
			Msg:  "依赖探活失败",
			Data: payload,
		})
		return
	}

	WriteOK(w, payload)
}
