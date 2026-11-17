package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/apperr"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/auth"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/resource"
)

type createProjectReq struct {
	RequestID string         `json:"requestId"`
	Fields    map[string]any `json:"fields"`
}

func (d Deps) handleProjectPage(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFrom(r.Context())
	if !ok {
		WriteError(w, apperr.Unauthorized("未登录或登录已过期"))
		return
	}
	pageNo := queryInt(r, "pageNo", 1)
	pageSize := queryInt(r, "pageSize", 10)
	result, err := d.Projects.Page(r.Context(), session.UserID, pageNo, pageSize)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteOK(w, result)
}

func (d Deps) handleProjectList(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFrom(r.Context())
	if !ok {
		WriteError(w, apperr.Unauthorized("未登录或登录已过期"))
		return
	}
	list, err := d.Projects.List(r.Context(), session.UserID)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteOK(w, list)
}

func (d Deps) handleProjectCreate(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFrom(r.Context())
	if !ok {
		WriteError(w, apperr.Unauthorized("未登录或登录已过期"))
		return
	}
	var req createProjectReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, apperr.BadRequest("请求体无效"))
		return
	}
	receipt, err := d.Commands.Execute(r.Context(), resource.CreateCommand{
		ReqID:   req.RequestID,
		ResType: resource.TypeProject,
		Fields:  req.Fields,
	}, resource.Context{ActorUserID: session.UserID})
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteOK(w, receipt.Result)
}

func queryInt(r *http.Request, key string, fallback int) int {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return n
}
