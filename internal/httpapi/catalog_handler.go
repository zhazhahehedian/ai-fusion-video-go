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

type createOwnedReq struct {
	RequestID string         `json:"requestId"`
	ProjectID *int64         `json:"projectId"`
	Fields    map[string]any `json:"fields"`
}

func (d Deps) handleStoryboardGet(w http.ResponseWriter, r *http.Request) {
	d.readID(w, r, "id", func(actor, id int64) (any, error) {
		return d.Content.Storyboard(r.Context(), actor, id)
	})
}

func (d Deps) handleStoryboardByProject(w http.ResponseWriter, r *http.Request) {
	d.readID(w, r, "projectId", func(actor, id int64) (any, error) {
		return d.Content.StoryboardByProject(r.Context(), actor, id)
	})
}

func (d Deps) handleStoryboardCreate(w http.ResponseWriter, r *http.Request) {
	d.handleOwnedCreate(w, r, resource.TypeStoryboard, true)
}

func (d Deps) handleStoryboardUpdate(w http.ResponseWriter, r *http.Request) {
	d.handleResourcePatch(w, r, resource.TypeStoryboard)
}

func (d Deps) handleStoryboardEpisodes(w http.ResponseWriter, r *http.Request) {
	d.readID(w, r, "storyboardId", func(actor, id int64) (any, error) {
		return d.Content.StoryboardEpisodes(r.Context(), actor, id)
	})
}

func (d Deps) handleStoryboardEpisodeGet(w http.ResponseWriter, r *http.Request) {
	d.readID(w, r, "id", func(actor, id int64) (any, error) {
		return d.Content.StoryboardEpisode(r.Context(), actor, id)
	})
}

func (d Deps) handleStoryboardEpisodeCreate(w http.ResponseWriter, r *http.Request) {
	d.handleChildCreate(w, r, resource.TypeStoryboardEpisode, "storyboardId")
}

func (d Deps) handleStoryboardEpisodeUpdate(w http.ResponseWriter, r *http.Request) {
	d.handleResourcePatch(w, r, resource.TypeStoryboardEpisode)
}

func (d Deps) handleStoryboardEpisodeDelete(w http.ResponseWriter, r *http.Request) {
	d.handleResourceDelete(w, r, resource.TypeStoryboardEpisode)
}

func (d Deps) handleStoryboardBindScriptEpisode(w http.ResponseWriter, r *http.Request) {
	d.handleScopedPatch(w, r, resource.TypeStoryboardEpisode, map[string]struct{}{"scriptEpisodeId": {}})
}

func (d Deps) handleStoryboardScenesByEpisode(w http.ResponseWriter, r *http.Request) {
	d.readID(w, r, "episodeId", func(actor, id int64) (any, error) {
		return d.Content.StoryboardScenesByEpisode(r.Context(), actor, id)
	})
}

func (d Deps) handleStoryboardScenesByBoard(w http.ResponseWriter, r *http.Request) {
	d.readID(w, r, "storyboardId", func(actor, id int64) (any, error) {
		return d.Content.StoryboardScenesByBoard(r.Context(), actor, id)
	})
}

func (d Deps) handleStoryboardSceneGet(w http.ResponseWriter, r *http.Request) {
	d.readID(w, r, "id", func(actor, id int64) (any, error) {
		return d.Content.StoryboardScene(r.Context(), actor, id)
	})
}

func (d Deps) handleStoryboardSceneCreate(w http.ResponseWriter, r *http.Request) {
	d.handleChildCreate(w, r, resource.TypeStoryboardScene, "episodeId")
}

func (d Deps) handleStoryboardSceneUpdate(w http.ResponseWriter, r *http.Request) {
	d.handleResourcePatch(w, r, resource.TypeStoryboardScene)
}

func (d Deps) handleStoryboardSceneDelete(w http.ResponseWriter, r *http.Request) {
	d.handleResourceDelete(w, r, resource.TypeStoryboardScene)
}

func (d Deps) handleStoryboardItemsByBoard(w http.ResponseWriter, r *http.Request) {
	d.readID(w, r, "storyboardId", func(actor, id int64) (any, error) {
		return d.Content.StoryboardItemsByBoard(r.Context(), actor, id)
	})
}

func (d Deps) handleStoryboardItemsByScene(w http.ResponseWriter, r *http.Request) {
	d.readID(w, r, "sceneId", func(actor, id int64) (any, error) {
		return d.Content.StoryboardItemsByScene(r.Context(), actor, id)
	})
}

func (d Deps) handleStoryboardItemGet(w http.ResponseWriter, r *http.Request) {
	d.readID(w, r, "id", func(actor, id int64) (any, error) {
		return d.Content.StoryboardItem(r.Context(), actor, id)
	})
}

func (d Deps) handleStoryboardItemCreate(w http.ResponseWriter, r *http.Request) {
	d.handleChildCreate(w, r, resource.TypeShot, "storyboardSceneId")
}

func (d Deps) handleStoryboardItemUpdate(w http.ResponseWriter, r *http.Request) {
	d.handleResourcePatch(w, r, resource.TypeShot)
}

func (d Deps) handleStoryboardItemDelete(w http.ResponseWriter, r *http.Request) {
	d.handleResourceDelete(w, r, resource.TypeShot)
}

func (d Deps) handleStoryboardItemFrame(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFrom(r.Context())
	if !ok {
		WriteError(w, r, apperr.Unauthorized("未登录或登录已过期"))
		return
	}
	id, err := pathID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}
	var req struct {
		Version   int32          `json:"version"`
		RequestID string         `json:"requestId"`
		Patch     map[string]any `json:"patch"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, r, apperr.BadRequest("请求体无效"))
		return
	}
	frame, _ := req.Patch["frameType"].(string)
	if frame != "first" && frame != "last" {
		WriteError(w, r, apperr.BadRequest("frameType 仅支持 first 或 last"))
		return
	}
	if _, exists := req.Patch["imageUrl"]; !exists {
		WriteError(w, r, apperr.BadRequest("imageUrl 不能为空"))
		return
	}
	if _, exists := req.Patch["prompt"]; !exists {
		WriteError(w, r, apperr.BadRequest("prompt 不能为空"))
		return
	}
	prefix := frame + "Frame"
	fields := map[string]any{
		prefix + "ImageUrl": req.Patch["imageUrl"],
		prefix + "Prompt":   req.Patch["prompt"],
	}
	receipt, err := d.Commands.Execute(r.Context(), resource.PatchCommand{
		ReqID:   req.RequestID,
		ResType: resource.TypeShot,
		ID:      id,
		Version: req.Version,
		Patch:   fields,
	}, resource.Context{ActorUserID: session.UserID})
	if err != nil {
		WriteError(w, r, err)
		return
	}
	WriteOK(w, receipt.Result)
}

func (d Deps) handleAssetGet(w http.ResponseWriter, r *http.Request) {
	d.readID(w, r, "id", func(actor, id int64) (any, error) {
		return d.Content.Asset(r.Context(), actor, id)
	})
}

func (d Deps) handleAssetList(w http.ResponseWriter, r *http.Request) {
	d.handleAssetQuery(w, r, false)
}

func (d Deps) handleAssetListWithItems(w http.ResponseWriter, r *http.Request) {
	d.handleAssetQuery(w, r, true)
}

func (d Deps) handleAssetQuery(w http.ResponseWriter, r *http.Request, withItems bool) {
	session, ok := auth.SessionFrom(r.Context())
	if !ok {
		WriteError(w, r, apperr.Unauthorized("未登录或登录已过期"))
		return
	}
	var projectID *int64
	if raw := r.URL.Query().Get("projectId"); raw != "" {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || id <= 0 {
			WriteError(w, r, apperr.BadRequest("projectId 无效"))
			return
		}
		projectID = &id
	}
	list, err := d.Content.Assets(r.Context(), session.UserID, projectID, r.URL.Query().Get("type"), r.URL.Query().Get("keyword"), withItems)
	if err != nil {
		WriteError(w, r, err)
		return
	}
	WriteOK(w, list)
}

func (d Deps) handleAssetCreate(w http.ResponseWriter, r *http.Request) {
	d.handleOwnedCreate(w, r, resource.TypeAsset, true)
}

func (d Deps) handleAssetUpdate(w http.ResponseWriter, r *http.Request) {
	d.handleResourcePatch(w, r, resource.TypeAsset)
}

func (d Deps) handleAssetDelete(w http.ResponseWriter, r *http.Request) {
	d.handleResourceDelete(w, r, resource.TypeAsset)
}

func (d Deps) handleAssetItemGet(w http.ResponseWriter, r *http.Request) {
	d.readID(w, r, "id", func(actor, id int64) (any, error) {
		return d.Content.AssetItem(r.Context(), actor, id)
	})
}

func (d Deps) handleAssetItems(w http.ResponseWriter, r *http.Request) {
	d.readID(w, r, "assetId", func(actor, id int64) (any, error) {
		return d.Content.AssetItems(r.Context(), actor, id)
	})
}

func (d Deps) handleAssetItemCreate(w http.ResponseWriter, r *http.Request) {
	d.handleChildCreate(w, r, resource.TypeAssetItem, "assetId")
}

func (d Deps) handleAssetItemUpdate(w http.ResponseWriter, r *http.Request) {
	d.handleResourcePatch(w, r, resource.TypeAssetItem)
}

func (d Deps) handleAssetItemDelete(w http.ResponseWriter, r *http.Request) {
	d.handleResourceDelete(w, r, resource.TypeAssetItem)
}

func (d Deps) handleImageTaskCreate(w http.ResponseWriter, r *http.Request) {
	d.handleOwnedCreate(w, r, resource.TypeImageTask, false)
}

func (d Deps) handleImageTaskUpdate(w http.ResponseWriter, r *http.Request) {
	d.handleResourcePatch(w, r, resource.TypeImageTask)
}

func (d Deps) handleImageTaskDelete(w http.ResponseWriter, r *http.Request) {
	d.deleteGenerationTask(w, r, true)
}

func (d Deps) handleImageTaskGet(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFrom(r.Context())
	if !ok {
		WriteError(w, r, apperr.Unauthorized("未登录或登录已过期"))
		return
	}
	taskID := chi.URLParam(r, "taskId")
	if taskID == "" {
		WriteError(w, r, apperr.BadRequest("taskId 无效"))
		return
	}
	vo, err := d.Content.ImageTask(r.Context(), session.UserID, taskID)
	if err != nil {
		WriteError(w, r, err)
		return
	}
	WriteOK(w, vo)
}

func (d Deps) handleImageTaskPage(w http.ResponseWriter, r *http.Request) {
	d.taskPage(w, r, true)
}

func (d Deps) handleImageItems(w http.ResponseWriter, r *http.Request) {
	d.readID(w, r, "id", func(actor, id int64) (any, error) {
		return d.Content.ImageItems(r.Context(), actor, id)
	})
}

func (d Deps) handleVideoTaskCreate(w http.ResponseWriter, r *http.Request) {
	d.handleOwnedCreate(w, r, resource.TypeVideoTask, false)
}

func (d Deps) handleVideoTaskUpdate(w http.ResponseWriter, r *http.Request) {
	d.handleResourcePatch(w, r, resource.TypeVideoTask)
}

func (d Deps) handleVideoTaskDelete(w http.ResponseWriter, r *http.Request) {
	d.deleteGenerationTask(w, r, false)
}

func (d Deps) deleteGenerationTask(w http.ResponseWriter, r *http.Request, image bool) {
	session, ok := auth.SessionFrom(r.Context())
	if !ok {
		WriteError(w, r, apperr.Unauthorized("未登录或登录已过期"))
		return
	}
	id, err := pathID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}
	if image {
		err = d.Content.DeleteImageTask(r.Context(), session.UserID, id)
	} else {
		err = d.Content.DeleteVideoTask(r.Context(), session.UserID, id)
	}
	if err != nil {
		WriteError(w, r, err)
		return
	}
	WriteOK(w, true)
}

func (d Deps) handleVideoTaskGet(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.SessionFrom(r.Context())
	if !ok {
		WriteError(w, r, apperr.Unauthorized("未登录或登录已过期"))
		return
	}
	taskID := chi.URLParam(r, "taskId")
	if taskID == "" {
		WriteError(w, r, apperr.BadRequest("taskId 无效"))
		return
	}
	vo, err := d.Content.VideoTask(r.Context(), session.UserID, taskID)
	if err != nil {
		WriteError(w, r, err)
		return
	}
	WriteOK(w, vo)
}

func (d Deps) handleVideoTaskPage(w http.ResponseWriter, r *http.Request) {
	d.taskPage(w, r, false)
}

func (d Deps) handleVideoItems(w http.ResponseWriter, r *http.Request) {
	d.readID(w, r, "id", func(actor, id int64) (any, error) {
		return d.Content.VideoItems(r.Context(), actor, id)
	})
}

func (d Deps) taskPage(w http.ResponseWriter, r *http.Request, image bool) {
	session, ok := auth.SessionFrom(r.Context())
	if !ok {
		WriteError(w, r, apperr.Unauthorized("未登录或登录已过期"))
		return
	}
	pageNo := queryInt(r, "pageNo", 1)
	pageSize := queryInt(r, "pageSize", 10)
	var (
		page any
		err  error
	)
	if image {
		page, err = d.Content.ImageTaskPage(r.Context(), session.UserID, pageNo, pageSize)
	} else {
		page, err = d.Content.VideoTaskPage(r.Context(), session.UserID, pageNo, pageSize)
	}
	if err != nil {
		WriteError(w, r, err)
		return
	}
	WriteOK(w, page)
}

func (d Deps) handleOwnedCreate(w http.ResponseWriter, r *http.Request, resType resource.Type, projectRequired bool) {
	session, ok := auth.SessionFrom(r.Context())
	if !ok {
		WriteError(w, r, apperr.Unauthorized("未登录或登录已过期"))
		return
	}
	var req createOwnedReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, r, apperr.BadRequest("请求体无效"))
		return
	}
	if projectRequired && (req.ProjectID == nil || *req.ProjectID <= 0) {
		WriteError(w, r, apperr.BadRequest("projectId 无效"))
		return
	}
	if req.Fields == nil {
		req.Fields = map[string]any{}
	}
	receipt, err := d.Commands.Execute(r.Context(), resource.CreateCommand{
		ReqID:     req.RequestID,
		ResType:   resType,
		ProjectID: req.ProjectID,
		Fields:    req.Fields,
	}, resource.Context{ActorUserID: session.UserID})
	if err != nil {
		WriteError(w, r, err)
		return
	}
	WriteOK(w, receipt.Result)
}

func (d Deps) handleScopedPatch(w http.ResponseWriter, r *http.Request, resType resource.Type, allowed map[string]struct{}) {
	session, ok := auth.SessionFrom(r.Context())
	if !ok {
		WriteError(w, r, apperr.Unauthorized("未登录或登录已过期"))
		return
	}
	id, err := pathID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}
	var req patchResourceReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, r, apperr.BadRequest("请求体无效"))
		return
	}
	if len(req.Patch) != len(allowed) {
		WriteError(w, r, apperr.BadRequest("字段范围无效"))
		return
	}
	for key := range req.Patch {
		if _, ok := allowed[key]; !ok {
			WriteError(w, r, apperr.BadRequest("不允许修改字段："+key))
			return
		}
	}
	receipt, err := d.Commands.Execute(r.Context(), resource.PatchCommand{
		ReqID:   req.RequestID,
		ResType: resType,
		ID:      id,
		Version: req.Version,
		Patch:   req.Patch,
	}, resource.Context{ActorUserID: session.UserID})
	if err != nil {
		WriteError(w, r, err)
		return
	}
	WriteOK(w, receipt.Result)
}

func (d Deps) readID(w http.ResponseWriter, r *http.Request, name string, load func(actor, id int64) (any, error)) {
	session, ok := auth.SessionFrom(r.Context())
	if !ok {
		WriteError(w, r, apperr.Unauthorized("未登录或登录已过期"))
		return
	}
	id, err := pathID(r, name)
	if err != nil {
		WriteError(w, r, err)
		return
	}
	vo, err := load(session.UserID, id)
	if err != nil {
		WriteError(w, r, err)
		return
	}
	WriteOK(w, vo)
}

func (d Deps) handleStoryboardClear(w http.ResponseWriter, r *http.Request) {
	d.handleClearContent(w, r, resource.TypeStoryboard)
}

func (d Deps) handleStoryboardEpisodeClear(w http.ResponseWriter, r *http.Request) {
	d.handleClearContent(w, r, resource.TypeStoryboardEpisode)
}

func (d Deps) handleClearContent(w http.ResponseWriter, r *http.Request, resType resource.Type) {
	session, ok := auth.SessionFrom(r.Context())
	if !ok {
		WriteError(w, r, apperr.Unauthorized("未登录或登录已过期"))
		return
	}
	id, err := pathID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}
	var req struct {
		RequestID string `json:"requestId"`
		Version   int32  `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, r, apperr.BadRequest("请求体无效"))
		return
	}
	receipt, err := d.Commands.Execute(r.Context(), resource.ClearCommand{
		ReqID:   req.RequestID,
		ResType: resType,
		ID:      id,
		Version: req.Version,
	}, resource.Context{ActorUserID: session.UserID})
	if err != nil {
		WriteError(w, r, err)
		return
	}
	WriteOK(w, receipt.Result)
}

func pathID(r *http.Request, name string) (int64, error) {
	id, err := strconv.ParseInt(chi.URLParam(r, name), 10, 64)
	if err != nil || id <= 0 {
		return 0, apperr.BadRequest(name + " 无效")
	}
	return id, nil
}
