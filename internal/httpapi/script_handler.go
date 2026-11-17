package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/apperr"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/auth"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/resource"
)

type createChildReq struct {
	ParentID      int64          `json:"parentId"`
	ParentVersion int32          `json:"parentVersion"`
	RequestID     string         `json:"requestId"`
	Fields        map[string]any `json:"fields"`
}

type patchResourceReq struct {
	ID        int64          `json:"id"`
	Version   int32          `json:"version"`
	RequestID string         `json:"requestId"`
	Patch     map[string]any `json:"patch"`
}

type replaceSourceReq struct {
	RequestID  string `json:"requestId"`
	Version    int32  `json:"version"`
	RawContent string `json:"rawContent"`
}

func (d Deps) handleScriptByProject(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFrom(r.Context())
	if !ok {
		WriteError(w, apperr.Unauthorized("未登录或登录已过期"))
		return
	}
	projectID, err := strconv.ParseInt(chi.URLParam(r, "projectId"), 10, 64)
	if err != nil || projectID <= 0 {
		WriteError(w, apperr.BadRequest("projectId 无效"))
		return
	}
	vo, err := d.Scripts.ByProject(r.Context(), session.UserID, projectID)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteOK(w, vo)
}

func (d Deps) handleScriptGet(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFrom(r.Context())
	if !ok {
		WriteError(w, apperr.Unauthorized("未登录或登录已过期"))
		return
	}
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		WriteError(w, apperr.BadRequest("id 无效"))
		return
	}
	vo, err := d.Scripts.Get(r.Context(), session.UserID, id)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteOK(w, vo)
}

func (d Deps) handleScriptUpdate(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFrom(r.Context())
	if !ok {
		WriteError(w, apperr.Unauthorized("未登录或登录已过期"))
		return
	}
	var req patchResourceReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, apperr.BadRequest("请求体无效"))
		return
	}
	receipt, err := d.Commands.Execute(r.Context(), resource.PatchCommand{
		ReqID:   req.RequestID,
		ResType: resource.TypeScript,
		ID:      req.ID,
		Version: req.Version,
		Patch:   req.Patch,
	}, resource.Context{ActorUserID: session.UserID})
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteOK(w, receipt.Result)
}

func (d Deps) handleScriptReplaceSource(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFrom(r.Context())
	if !ok {
		WriteError(w, apperr.Unauthorized("未登录或登录已过期"))
		return
	}
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		WriteError(w, apperr.BadRequest("id 无效"))
		return
	}
	var req replaceSourceReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, apperr.BadRequest("请求体无效"))
		return
	}
	receipt, err := d.Commands.Execute(r.Context(), resource.SourceCommand{
		ReqID:      req.RequestID,
		ResourceID: id,
		Version:    req.Version,
		RawContent: req.RawContent,
	}, resource.Context{ActorUserID: session.UserID})
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteOK(w, receipt.Result)
}

func (d Deps) handleScriptEpisodes(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFrom(r.Context())
	if !ok {
		WriteError(w, apperr.Unauthorized("未登录或登录已过期"))
		return
	}
	scriptID, err := strconv.ParseInt(chi.URLParam(r, "scriptId"), 10, 64)
	if err != nil || scriptID <= 0 {
		WriteError(w, apperr.BadRequest("scriptId 无效"))
		return
	}
	list, err := d.Scripts.ListEpisodes(r.Context(), session.UserID, scriptID)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteOK(w, list)
}

func (d Deps) handleEpisodeGet(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFrom(r.Context())
	if !ok {
		WriteError(w, apperr.Unauthorized("未登录或登录已过期"))
		return
	}
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		WriteError(w, apperr.BadRequest("id 无效"))
		return
	}
	vo, err := d.Scripts.GetEpisode(r.Context(), session.UserID, id)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteOK(w, vo)
}

func (d Deps) handleEpisodeCreate(w http.ResponseWriter, r *http.Request) {
	d.handleChildCreate(w, r, resource.TypeScriptEpisode, "scriptId")
}

func (d Deps) handleEpisodeUpdate(w http.ResponseWriter, r *http.Request) {
	d.handleResourcePatch(w, r, resource.TypeScriptEpisode)
}

func (d Deps) handleEpisodeDelete(w http.ResponseWriter, r *http.Request) {
	d.handleResourceDelete(w, r, resource.TypeScriptEpisode)
}

func (d Deps) handleEpisodeScenes(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFrom(r.Context())
	if !ok {
		WriteError(w, apperr.Unauthorized("未登录或登录已过期"))
		return
	}
	episodeID, err := strconv.ParseInt(chi.URLParam(r, "episodeId"), 10, 64)
	if err != nil || episodeID <= 0 {
		WriteError(w, apperr.BadRequest("episodeId 无效"))
		return
	}
	list, err := d.Scripts.ListScenes(r.Context(), session.UserID, episodeID)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteOK(w, list)
}

func (d Deps) handleSceneGet(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFrom(r.Context())
	if !ok {
		WriteError(w, apperr.Unauthorized("未登录或登录已过期"))
		return
	}
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		WriteError(w, apperr.BadRequest("id 无效"))
		return
	}
	vo, err := d.Scripts.GetScene(r.Context(), session.UserID, id)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteOK(w, vo)
}

func (d Deps) handleSceneCreate(w http.ResponseWriter, r *http.Request) {
	d.handleChildCreate(w, r, resource.TypeScriptScene, "episodeId")
}

func (d Deps) handleSceneUpdate(w http.ResponseWriter, r *http.Request) {
	d.handleResourcePatch(w, r, resource.TypeScriptScene)
}

func (d Deps) handleSceneDelete(w http.ResponseWriter, r *http.Request) {
	d.handleResourceDelete(w, r, resource.TypeScriptScene)
}

func (d Deps) handleChildCreate(w http.ResponseWriter, r *http.Request, resType resource.Type, parentField string) {
	session, ok := auth.SessionFrom(r.Context())
	if !ok {
		WriteError(w, apperr.Unauthorized("未登录或登录已过期"))
		return
	}
	var req createChildReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, apperr.BadRequest("请求体无效"))
		return
	}
	if req.ParentID <= 0 {
		WriteError(w, apperr.BadRequest("parentId 无效"))
		return
	}
	fields := req.Fields
	if fields == nil {
		fields = map[string]any{}
	}
	for _, key := range []string{"scriptId", "episodeId", "parentVersion"} {
		if _, exists := fields[key]; exists {
			WriteError(w, apperr.BadRequest("父对象字段必须在请求顶层指定"))
			return
		}
	}
	fields[parentField] = req.ParentID
	fields["parentVersion"] = req.ParentVersion
	receipt, err := d.Commands.Execute(r.Context(), resource.CreateCommand{
		ReqID:   req.RequestID,
		ResType: resType,
		Fields:  fields,
	}, resource.Context{ActorUserID: session.UserID})
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteOK(w, receipt.Result)
}

func (d Deps) handleResourcePatch(w http.ResponseWriter, r *http.Request, resType resource.Type) {
	session, ok := auth.SessionFrom(r.Context())
	if !ok {
		WriteError(w, apperr.Unauthorized("未登录或登录已过期"))
		return
	}
	var req patchResourceReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, apperr.BadRequest("请求体无效"))
		return
	}
	receipt, err := d.Commands.Execute(r.Context(), resource.PatchCommand{
		ReqID:   req.RequestID,
		ResType: resType,
		ID:      req.ID,
		Version: req.Version,
		Patch:   req.Patch,
	}, resource.Context{ActorUserID: session.UserID})
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteOK(w, receipt.Result)
}

func (d Deps) handleResourceDelete(w http.ResponseWriter, r *http.Request, resType resource.Type) {
	session, ok := auth.SessionFrom(r.Context())
	if !ok {
		WriteError(w, apperr.Unauthorized("未登录或登录已过期"))
		return
	}
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		WriteError(w, apperr.BadRequest("id 无效"))
		return
	}
	version, err := strconv.Atoi(r.URL.Query().Get("version"))
	if err != nil || version < 0 {
		WriteError(w, apperr.BadRequest("version 无效"))
		return
	}
	requestID := r.URL.Query().Get("requestId")
	_, err = d.Commands.Execute(r.Context(), resource.DeleteCommand{
		ReqID:   requestID,
		ResType: resType,
		ID:      id,
		Version: int32(version),
	}, resource.Context{ActorUserID: session.UserID})
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteOK(w, true)
}
