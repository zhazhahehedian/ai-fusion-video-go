package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/apperr"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/auth"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/generation"
)

type echoEnqueueReq struct {
	RequestID string `json:"requestId"`
	Prompt    string `json:"prompt"`
}

func (d Deps) handleGenerationEcho(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFrom(r.Context())
	if !ok {
		WriteError(w, r, apperr.Unauthorized("未登录或登录已过期"))
		return
	}
	var req echoEnqueueReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, r, apperr.BadRequest("请求体无效"))
		return
	}
	result, err := d.Generation.EnqueueEcho(r.Context(), session.UserID, generation.EchoRequest{
		RequestID: req.RequestID,
		Prompt:    req.Prompt,
	})
	if err != nil {
		WriteError(w, r, err)
		return
	}
	WriteOK(w, result)
}

func (d Deps) handleGenerationSubmissionGet(w http.ResponseWriter, r *http.Request) {
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
	view, err := d.Generation.GetSubmission(r.Context(), session.UserID, id)
	if err != nil {
		WriteError(w, r, err)
		return
	}
	WriteOK(w, view)
}
