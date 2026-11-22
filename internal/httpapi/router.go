package httpapi

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/auth"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/generation"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/logging"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/media"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/project"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/resource"
)

type Deps struct {
	Logger     *slog.Logger
	DB         *sql.DB
	Redis      *redis.Client
	Auth       *auth.Service
	Projects   *project.Service
	Commands   *resource.Service
	Scripts    *resource.ScriptQuery
	Content    *resource.ContentQuery
	Generation *generation.Service
	Media      *media.Service
	// MaxBodyBytes 为 0 时使用 DefaultMaxBodyBytes。
	MaxBodyBytes int64
}

func NewRouter(deps Deps) http.Handler {
	if deps.Logger == nil {
		panic("httpapi.NewRouter: logger 未注入")
	}
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(AccessLog(deps.Logger))
	r.Use(Recover)
	r.Use(LimitRequestBody(maxBodyBytes(deps)))

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

		pr.Get("/api/storyboard/project/{projectId}", deps.handleStoryboardByProject)
		pr.Get("/api/storyboard/episode/{episodeId}/scenes", deps.handleStoryboardScenesByEpisode)
		pr.Get("/api/storyboard/episode/{id}", deps.handleStoryboardEpisodeGet)
		pr.Post("/api/storyboard/episode", deps.handleStoryboardEpisodeCreate)
		pr.Put("/api/storyboard/episode", deps.handleStoryboardEpisodeUpdate)
		pr.Delete("/api/storyboard/episode/{id}", deps.handleStoryboardEpisodeDelete)
		pr.Put("/api/storyboard/episode/{id}/bindScriptEpisode", deps.handleStoryboardBindScriptEpisode)
		pr.Get("/api/storyboard/scene/{sceneId}/items", deps.handleStoryboardItemsByScene)
		pr.Get("/api/storyboard/scene/{id}", deps.handleStoryboardSceneGet)
		pr.Post("/api/storyboard/scene", deps.handleStoryboardSceneCreate)
		pr.Put("/api/storyboard/scene", deps.handleStoryboardSceneUpdate)
		pr.Delete("/api/storyboard/scene/{id}", deps.handleStoryboardSceneDelete)
		pr.Get("/api/storyboard/item/{id}", deps.handleStoryboardItemGet)
		pr.Post("/api/storyboard/item", deps.handleStoryboardItemCreate)
		pr.Put("/api/storyboard/item", deps.handleStoryboardItemUpdate)
		pr.Put("/api/storyboard/item/{id}/updateFrame", deps.handleStoryboardItemFrame)
		pr.Delete("/api/storyboard/item/{id}", deps.handleStoryboardItemDelete)
		pr.Get("/api/storyboard/{storyboardId}/episodes", deps.handleStoryboardEpisodes)
		pr.Get("/api/storyboard/{storyboardId}/scenes", deps.handleStoryboardScenesByBoard)
		pr.Get("/api/storyboard/{storyboardId}/items", deps.handleStoryboardItemsByBoard)
		pr.Get("/api/storyboard/{id}", deps.handleStoryboardGet)
		pr.Post("/api/storyboard", deps.handleStoryboardCreate)
		pr.Put("/api/storyboard", deps.handleStoryboardUpdate)
		pr.Post("/api/storyboard/{id}/clearContent", deps.handleStoryboardClear)
		pr.Post("/api/storyboard/episode/{id}/clearContent", deps.handleStoryboardEpisodeClear)

		pr.Get("/api/asset/list", deps.handleAssetList)
		pr.Get("/api/asset/list-with-items", deps.handleAssetListWithItems)
		pr.Get("/api/asset/item/{id}", deps.handleAssetItemGet)
		pr.Post("/api/asset/item", deps.handleAssetItemCreate)
		pr.Put("/api/asset/item", deps.handleAssetItemUpdate)
		pr.Delete("/api/asset/item/{id}", deps.handleAssetItemDelete)
		pr.Get("/api/asset/{assetId}/items", deps.handleAssetItems)
		pr.Get("/api/asset/{id}", deps.handleAssetGet)
		pr.Post("/api/asset", deps.handleAssetCreate)
		pr.Put("/api/asset", deps.handleAssetUpdate)
		pr.Delete("/api/asset/{id}", deps.handleAssetDelete)

		pr.Post("/api/generation/echo", deps.handleGenerationEcho)
		pr.Get("/api/generation/submissions/{id}", deps.handleGenerationSubmissionGet)
		pr.Post("/api/generation/image", deps.handleImageTaskCreate)
		pr.Put("/api/generation/image", deps.handleImageTaskUpdate)
		pr.Get("/api/generation/image/page", deps.handleImageTaskPage)
		pr.Get("/api/generation/image/{id}/items", deps.handleImageItems)
		pr.Get("/api/generation/image/{taskId}", deps.handleImageTaskGet)
		pr.Delete("/api/generation/image/{id}", deps.handleImageTaskDelete)
		pr.Post("/api/generation/video", deps.handleVideoTaskCreate)
		pr.Put("/api/generation/video", deps.handleVideoTaskUpdate)
		pr.Get("/api/generation/video/page", deps.handleVideoTaskPage)
		pr.Get("/api/generation/video/{id}/items", deps.handleVideoItems)
		pr.Get("/api/generation/video/{taskId}", deps.handleVideoTaskGet)
		pr.Delete("/api/generation/video/{id}", deps.handleVideoTaskDelete)

		pr.Post("/api/media/jobs", deps.handleMediaEnqueue)
		pr.Get("/api/media/jobs/{id}", deps.handleMediaJobGet)
	})

	return r
}

func (d Deps) handleHealthz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	logger := logging.From(r.Context())
	mysqlOK := true
	mysqlMsg := "up"
	if err := d.DB.PingContext(ctx); err != nil {
		mysqlOK = false
		mysqlMsg = err.Error()
		logger.ErrorContext(r.Context(), "MySQL 探活失败", "err", err)
	}

	redisOK := true
	redisMsg := "up"
	if err := d.Redis.Ping(ctx).Err(); err != nil {
		redisOK = false
		redisMsg = err.Error()
		logger.ErrorContext(r.Context(), "Redis 探活失败", "err", err)
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
