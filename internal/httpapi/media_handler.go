package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/apperr"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/auth"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/media"
)

type mediaEnqueueReq struct {
	RequestID string `json:"requestId"`
	Kind      string `json:"kind"`
}

func (d Deps) handleMediaEnqueue(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFrom(r.Context())
	if !ok {
		WriteError(w, r, apperr.Unauthorized("未登录或登录已过期"))
		return
	}
	var req mediaEnqueueReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, r, apperr.BadRequest("请求体无效"))
		return
	}
	result, err := d.Media.Enqueue(r.Context(), session.UserID, media.EnqueueRequest{
		RequestID: req.RequestID,
		Kind:      req.Kind,
	})
	if err != nil {
		WriteError(w, r, err)
		return
	}
	WriteOK(w, result)
}

func (d Deps) handleMediaJobGet(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFrom(r.Context())
	if !ok {
		WriteError(w, r, apperr.Unauthorized("未登录或登录已过期"))
		return
	}
	id := chi.URLParam(r, "id")
	if id == "" {
		WriteError(w, r, apperr.BadRequest("id 无效"))
		return
	}
	view, err := d.Media.Get(r.Context(), session.UserID, id)
	if err != nil {
		WriteError(w, r, err)
		return
	}
	WriteOK(w, view)
}
