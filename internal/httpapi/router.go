package httpapi

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/auth"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/generation"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/media"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/project"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/resource"
)

type Deps struct {
	DB         *sql.DB
	Redis      *redis.Client
	Auth       *auth.Service
	Projects   *project.Service
	Commands   *resource.Service
	Scripts    *resource.ScriptQuery
	Generation *generation.Service
	Media      *media.Service
}

func NewRouter(deps Deps) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", deps.handleHealthz)

	r.Route("/api/auth", func(ar chi.Router) {
		ar.Post("/login", deps.handleLogin)
		ar.Post("/register", deps.handleRegister)
		ar.Post("/refresh", deps.handleRefresh)
		ar.Group(func(pr chi.Router) {
			pr.Use(deps.RequireAuth)
			pr.Post("/logout", deps.handleLogout)
			pr.Get("/user-info", deps.handleUserInfo)
		})
	})

	r.Group(func(pr chi.Router) {
		pr.Use(deps.RequireAuth)
		pr.Get("/api/project/page", deps.handleProjectPage)
		pr.Get("/api/project/list", deps.handleProjectList)
		pr.Post("/api/project", deps.handleProjectCreate)

		// 具体前缀先于 /api/script/{id}，避免 episode/scene/project 被吞掉
		pr.Get("/api/script/project/{projectId}", deps.handleScriptByProject)
		pr.Get("/api/script/episode/{episodeId}/scenes", deps.handleEpisodeScenes)
		pr.Get("/api/script/episode/{id}", deps.handleEpisodeGet)
		pr.Post("/api/script/episode", deps.handleEpisodeCreate)
		pr.Put("/api/script/episode", deps.handleEpisodeUpdate)
		pr.Delete("/api/script/episode/{id}", deps.handleEpisodeDelete)
		pr.Get("/api/script/scene/{id}", deps.handleSceneGet)
		pr.Post("/api/script/scene", deps.handleSceneCreate)
		pr.Put("/api/script/scene", deps.handleSceneUpdate)
		pr.Delete("/api/script/scene/{id}", deps.handleSceneDelete)
		pr.Get("/api/script/{scriptId}/episodes", deps.handleScriptEpisodes)
		pr.Put("/api/script/{id}/source", deps.handleScriptReplaceSource)
		pr.Get("/api/script/{id}", deps.handleScriptGet)
		pr.Put("/api/script", deps.handleScriptUpdate)

		pr.Post("/api/generation/echo", deps.handleGenerationEcho)
		pr.Get("/api/generation/submissions/{id}", deps.handleGenerationSubmissionGet)

		pr.Post("/api/media/jobs", deps.handleMediaEnqueue)
		pr.Get("/api/media/jobs/{id}", deps.handleMediaJobGet)
	})

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
