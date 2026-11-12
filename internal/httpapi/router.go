package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"origin.cursor.com/yuigakki-w/ai-fusion-video-study/internal/config"
)

func NewRouter(cfg config.Config) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		WriteOK(w, map[string]any{
			"status":   "up",
			"phase":    1,
			"javaRoot": cfg.JavaRoot,
		})
	})

	return r
}
