package resource

import (
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/sqlcdb"
)

func StoryboardVO(row sqlcdb.AfvStoryboard) map[string]any {
	return map[string]any{
		"id":            row.ID,
		"projectId":     nullInt64(row.ProjectID),
		"scriptId":      nullInt64(row.ScriptID),
		"title":         nullString(row.Title),
		"description":   nullString(row.Description),
		"customColumns": rawJSON(row.CustomColumns),
		"scope":         nullInt32(row.Scope),
		"ownerType":     nullInt32(row.OwnerType),
		"ownerId":       nullInt64(row.OwnerID),
		"totalDuration": nullInt32(row.TotalDuration),
		"status":        nullInt32(row.Status),
		"deleted":       row.Deleted != 0,
		"createTime":    formatTime(row.CreateTime),
		"updateTime":    formatTime(row.UpdateTime),
		"version":       row.Version,
	}
}

func StoryboardEpisodeVO(row sqlcdb.AfvStoryboardEpisode) map[string]any {
	return map[string]any{
		"id":               row.ID,
		"storyboardId":     row.StoryboardID,
		"scriptEpisodeId":  nullInt64(row.ScriptEpisodeID),
		"episodeNumber":    nullInt32(row.EpisodeNumber),
		"title":            nullString(row.Title),
		"synopsis":         nullString(row.Synopsis),
		"sortOrder":        nullInt32(row.SortOrder),
		"status":           nullInt32(row.Status),
		"deleted":          deletedTiny(row.Deleted),
		"deletedId":        row.DeletedID,
		"createTime":       formatNullTime(row.CreateTime),
		"updateTime":       formatNullTime(row.UpdateTime),
		"composedVideoUrl": nullString(row.ComposedVideoUrl),
		"composeStatus":    row.ComposeStatus,
		"composeErrorMsg":  nullString(row.ComposeErrorMsg),
		"composedAt":       formatNullTime(row.ComposedAt),
		"bgmTaskId":        nullString(row.BgmTaskID),
		"bgmAudioUrl":      nullString(row.BgmAudioUrl),
		"bgmSfxAudioUrl":   nullString(row.BgmSfxAudioUrl),
		"bgmErrorMsg":      nullString(row.BgmErrorMsg),
		"version":          row.Version,
		"composeJobId":     nullString(row.ComposeJobID),
	}
}

func StoryboardSceneVO(row sqlcdb.AfvStoryboardScene) map[string]any {
	return map[string]any{
		"id":           row.ID,
		"episodeId":    row.EpisodeID,
		"storyboardId": row.StoryboardID,
		"sceneNumber":  nullString(row.SceneNumber),
		"sceneHeading": nullString(row.SceneHeading),
		"location":     nullString(row.Location),
		"timeOfDay":    nullString(row.TimeOfDay),
		"intExt":       nullString(row.IntExt),
		"sortOrder":    nullInt32(row.SortOrder),
		"status":       nullInt32(row.Status),
		"deleted":      deletedTiny(row.Deleted),
		"createTime":   formatNullTime(row.CreateTime),
		"updateTime":   formatNullTime(row.UpdateTime),
		"version":      row.Version,
	}
}

func StoryboardItemVO(row sqlcdb.AfvStoryboardItem) map[string]any {
	return map[string]any{
		"id":                  row.ID,
		"storyboardId":        row.StoryboardID,
		"storyboardEpisodeId": nullInt64(row.StoryboardEpisodeID),
		"storyboardSceneId":   nullInt64(row.StoryboardSceneID),
		"sortOrder":           nullInt32(row.SortOrder),
		"shotNumber":          nullString(row.ShotNumber),
		"autoShotNumber":      nullString(row.AutoShotNumber),
		"imageUrl":            nullString(row.ImageUrl),
		"referenceImageUrl":   nullString(row.ReferenceImageUrl),
		"videoUrl":            nullString(row.VideoUrl),
		"generatedImageUrl":   nullString(row.GeneratedImageUrl),
		"firstFrameImageUrl":  nullString(row.FirstFrameImageUrl),
		"lastFrameImageUrl":   nullString(row.LastFrameImageUrl),
		"firstFramePrompt":    nullString(row.FirstFramePrompt),
		"lastFramePrompt":     nullString(row.LastFramePrompt),
		"generatedVideoUrl":   nullString(row.GeneratedVideoUrl),
		"shotType":            nullString(row.ShotType),
		"duration":            nullDecimal(row.Duration),
		"content":             nullString(row.Content),
		"sceneExpectation":    nullString(row.SceneExpectation),
		"sound":               nullString(row.Sound),
		"dialogue":            nullString(row.Dialogue),
		"soundEffect":         nullString(row.SoundEffect),
		"music":               nullString(row.Music),
		"cameraMovement":      nullString(row.CameraMovement),
		"cameraAngle":         nullString(row.CameraAngle),
		"cameraEquipment":     nullString(row.CameraEquipment),
		"focalLength":         nullString(row.FocalLength),
		"transition":          nullString(row.Transition),
		"characterIds":        rawJSON(row.CharacterIds),
		"sceneAssetItemId":    nullInt64(row.SceneAssetItemID),
		"propIds":             rawTextJSON(row.PropIds),
		"remark":              nullString(row.Remark),
		"customData":          rawJSON(row.CustomData),
		"aiGenerated":         nullTiny(row.AiGenerated),
		"status":              nullInt32(row.Status),
		"deleted":             row.Deleted != 0,
		"createTime":          formatTime(row.CreateTime),
		"updateTime":          formatTime(row.UpdateTime),
		"videoPrompt":         nullString(row.VideoPrompt),
		"version":             row.Version,
	}
}

func AssetVO(row sqlcdb.AfvAsset) map[string]any {
	return map[string]any{
		"id":          row.ID,
		"userId":      nullInt64(row.UserID),
		"projectId":   nullInt64(row.ProjectID),
		"type":        nullString(row.Type),
		"name":        row.Name,
		"description": nullString(row.Description),
		"coverUrl":    nullString(row.CoverUrl),
		"properties":  rawJSON(row.Properties),
		"tags":        rawJSON(row.Tags),
		"sourceType":  nullInt32(row.SourceType),
		"aiPrompt":    nullString(row.AiPrompt),
		"ownerType":   nullInt32(row.OwnerType),
		"ownerId":     nullInt64(row.OwnerID),
		"status":      nullInt32(row.Status),
		"deleted":     row.Deleted != 0,
		"createTime":  formatTime(row.CreateTime),
		"updateTime":  formatTime(row.UpdateTime),
		"version":     row.Version,
	}
}

func AssetItemVO(row sqlcdb.AfvAssetItem) map[string]any {
	return map[string]any{
		"id":           row.ID,
		"assetId":      row.AssetID,
		"itemType":     nullString(row.ItemType),
		"name":         nullString(row.Name),
		"imageUrl":     nullString(row.ImageUrl),
		"thumbnailUrl": nullString(row.ThumbnailUrl),
		"properties":   rawJSON(row.Properties),
		"sortOrder":    nullInt32(row.SortOrder),
		"sourceType":   nullInt32(row.SourceType),
		"aiPrompt":     nullString(row.AiPrompt),
		"deleted":      row.Deleted != 0,
		"createTime":   formatTime(row.CreateTime),
		"updateTime":   formatTime(row.UpdateTime),
		"version":      row.Version,
	}
}

func ImageTaskVO(row sqlcdb.AfvImageTask, version int32) map[string]any {
	return map[string]any{
		"id":                row.ID,
		"taskId":            row.TaskID,
		"userId":            row.UserID,
		"projectId":         nullInt64(row.ProjectID),
		"prompt":            nullString(row.Prompt),
		"promptTemplateId":  nullInt64(row.PromptTemplateID),
		"refImageUrls":      rawJSON(row.RefImageUrls),
		"ratio":             nullString(row.Ratio),
		"resolution":        nullString(row.Resolution),
		"aspectRatio":       nullString(row.AspectRatio),
		"width":             nullInt32(row.Width),
		"height":            nullInt32(row.Height),
		"count":             nullInt32(row.Count),
		"successCount":      nullInt32(row.SuccessCount),
		"status":            nullInt32(row.Status),
		"errorMsg":          nullString(row.ErrorMsg),
		"modelId":           nullInt64(row.ModelID),
		"workflowVersionId": nullInt64(row.WorkflowVersionID),
		"category":          nullString(row.Category),
		"ownerType":         nullInt32(row.OwnerType),
		"ownerId":           nullInt64(row.OwnerID),
		"deleted":           row.Deleted != 0,
		"createTime":        formatTime(row.CreateTime),
		"updateTime":        formatTime(row.UpdateTime),
		"version":           version,
	}
}

func ImageItemVO(row sqlcdb.AfvImageItem) map[string]any {
	return map[string]any{
		"id":             row.ID,
		"taskId":         row.TaskID,
		"platformTaskId": nullString(row.PlatformTaskID),
		"imageUrl":       nullString(row.ImageUrl),
		"thumbnailUrl":   nullString(row.ThumbnailUrl),
		"width":          nullInt32(row.Width),
		"height":         nullInt32(row.Height),
		"fileSize":       nullInt64(row.FileSize),
		"status":         nullInt32(row.Status),
		"errorMsg":       nullString(row.ErrorMsg),
		"deleted":        row.Deleted != 0,
		"createTime":     formatTime(row.CreateTime),
		"updateTime":     formatTime(row.UpdateTime),
	}
}

func VideoTaskVO(row sqlcdb.AfvVideoTask, version int32) map[string]any {
	return map[string]any{
		"id":                 row.ID,
		"taskId":             row.TaskID,
		"userId":             row.UserID,
		"projectId":          nullInt64(row.ProjectID),
		"prompt":             nullString(row.Prompt),
		"promptTemplateId":   nullInt64(row.PromptTemplateID),
		"generateMode":       nullString(row.GenerateMode),
		"firstFrameImageUrl": nullString(row.FirstFrameImageUrl),
		"lastFrameImageUrl":  nullString(row.LastFrameImageUrl),
		"referenceImageUrls": rawJSON(row.ReferenceImageUrls),
		"referenceVideoUrls": rawJSON(row.ReferenceVideoUrls),
		"referenceAudioUrls": rawJSON(row.ReferenceAudioUrls),
		"ratio":              nullString(row.Ratio),
		"resolution":         nullString(row.Resolution),
		"duration":           nullInt32(row.Duration),
		"watermark":          nullTiny(row.Watermark),
		"generateAudio":      nullTiny(row.GenerateAudio),
		"seed":               nullInt64(row.Seed),
		"cameraFixed":        nullTiny(row.CameraFixed),
		"count":              nullInt32(row.Count),
		"successCount":       nullInt32(row.SuccessCount),
		"status":             nullInt32(row.Status),
		"errorMsg":           nullString(row.ErrorMsg),
		"modelId":            nullInt64(row.ModelID),
		"workflowVersionId":  nullInt64(row.WorkflowVersionID),
		"category":           nullString(row.Category),
		"ownerType":          nullInt32(row.OwnerType),
		"ownerId":            nullInt64(row.OwnerID),
		"deleted":            row.Deleted != 0,
		"createTime":         formatTime(row.CreateTime),
		"updateTime":         formatTime(row.UpdateTime),
		"version":            version,
	}
}

func VideoItemVO(row sqlcdb.AfvVideoItem) map[string]any {
	return map[string]any{
		"id":             row.ID,
		"taskId":         row.TaskID,
		"platformTaskId": nullString(row.PlatformTaskID),
		"videoUrl":       nullString(row.VideoUrl),
		"coverUrl":       nullString(row.CoverUrl),
		"duration":       nullInt32(row.Duration),
		"fileSize":       nullInt64(row.FileSize),
		"status":         nullInt32(row.Status),
		"errorMsg":       nullString(row.ErrorMsg),
		"firstFrameUrl":  nullString(row.FirstFrameUrl),
		"lastFrameUrl":   nullString(row.LastFrameUrl),
		"deleted":        row.Deleted != 0,
		"createTime":     formatTime(row.CreateTime),
		"updateTime":     formatTime(row.UpdateTime),
	}
}
