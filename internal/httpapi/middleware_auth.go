package httpapi

import (
	"net/http"
	"strings"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/apperr"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/auth"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/logging"
)

func (d Deps) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractAccessToken(r)
		if token == "" {
			WriteError(w, r, apperr.Unauthorized("未登录或登录已过期"))
			return
		}
		session, err := d.Auth.ResolveAccessToken(r.Context(), token)
		if err != nil {
			WriteError(w, r, err)
			return
		}
		ctx := auth.WithSession(r.Context(), session)
		logging.WithAttrs(ctx, "user_id", session.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractAccessToken(r *http.Request) string {
	bearer := strings.TrimSpace(r.Header.Get("Authorization"))
	if bearer != "" {
		if strings.HasPrefix(bearer, "Bearer ") {
			return strings.TrimSpace(strings.TrimPrefix(bearer, "Bearer "))
		}
		return bearer
	}
	return strings.TrimSpace(r.URL.Query().Get("access_token"))
}
