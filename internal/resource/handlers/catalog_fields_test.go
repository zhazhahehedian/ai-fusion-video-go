package handlers

import (
	"errors"
	"testing"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/apperr"
	"github.com/zhazhahehedian/ai-fusion-video-go/internal/sqlcdb"
)

func TestStoryboardPatchRejectsUnknownAndSort(t *testing.T) {
	row := sqlcdb.AfvStoryboard{}
	_, err := applyStoryboardFields(&row, map[string]any{"title": "改标题"}, false)
	if !isBadRequest(err) {
		t.Fatalf("title is create-only, got %v", err)
	}
	_, err = applyStoryboardFields(&row, map[string]any{"sortOrder": 1}, false)
	if !isBadRequest(err) {
		t.Fatalf("sortOrder is not a storyboard field, got %v", err)
	}
	changed, err := applyStoryboardFields(&row, map[string]any{"description": "夜戏", "status": float64(1)}, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(changed) != 2 || !row.Description.Valid || row.Description.String != "夜戏" || row.Status.Int32 != 1 {
		t.Fatalf("patch = %+v changed %v", row, changed)
	}
}

func TestEpisodeCreateRequiresNumber(t *testing.T) {
	row := sqlcdb.AfvStoryboardEpisode{}
	if _, err := applyEpisodeFields(&row, map[string]any{"title": "第一集"}, true); !isBadRequest(err) {
		t.Fatalf("missing episodeNumber, got %v", err)
	}
	if _, err := applyEpisodeFields(&row, map[string]any{"episodeNumber": float64(1), "sortOrder": float64(0)}, false); !isBadRequest(err) {
		t.Fatalf("sortOrder on patch must fail, got %v", err)
	}
	changed, err := applyEpisodeFields(&row, map[string]any{"episodeNumber": float64(2), "sortOrder": float64(3)}, true)
	if err != nil {
		t.Fatal(err)
	}
	if row.EpisodeNumber.Int32 != 2 || row.SortOrder.Int32 != 3 || len(changed) != 2 {
		t.Fatalf("row %+v changed %v", row, changed)
	}
}

func TestShotFieldsDurationURLAndIDs(t *testing.T) {
	row := sqlcdb.AfvStoryboardItem{}
	if _, err := applyShotFields(&row, map[string]any{"duration": 1.234}, true); !isBadRequest(err) {
		t.Fatalf("three decimals must fail, got %v", err)
	}
	if _, err := applyShotFields(&row, map[string]any{"imageUrl": "ftp://files.example/a.png"}, true); !isBadRequest(err) {
		t.Fatalf("ftp url must fail, got %v", err)
	}
	if _, err := applyShotFields(&row, map[string]any{"characterIds": []any{float64(1), float64(1)}}, true); !isBadRequest(err) {
		t.Fatalf("duplicate ids must fail, got %v", err)
	}
	if _, err := applyShotFields(&row, map[string]any{"sortOrder": float64(1)}, false); !isBadRequest(err) {
		t.Fatalf("sortOrder patch must fail, got %v", err)
	}
	changed, err := applyShotFields(&row, map[string]any{
		"duration":     1.5,
		"imageUrl":     "https://cdn.example/a.png",
		"characterIds": []any{float64(7), float64(8)},
		"propIds":      []any{},
		"aiGenerated":  true,
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	if row.Duration.String != "1.50" || row.ImageUrl.String != "https://cdn.example/a.png" || string(row.CharacterIds) != "[7,8]" || row.PropIds.String != "[]" || row.AiGenerated.Int16 != 1 {
		t.Fatalf("shot %+v", row)
	}
	if len(changed) != 5 {
		t.Fatalf("changed %v", changed)
	}
}

func TestAssetFieldsRequireTypeAndName(t *testing.T) {
	row := sqlcdb.AfvAsset{}
	if _, err := applyAssetFields(&row, map[string]any{"name": "甲"}, true); !isBadRequest(err) {
		t.Fatalf("missing type, got %v", err)
	}
	if _, err := applyAssetFields(&row, map[string]any{"type": "robot", "name": "甲"}, true); !isBadRequest(err) {
		t.Fatalf("unknown type, got %v", err)
	}
	if _, err := applyAssetFields(&row, map[string]any{"name": "  "}, false); !isBadRequest(err) {
		t.Fatalf("blank name must fail, got %v", err)
	}
	_, err := applyAssetFields(&row, map[string]any{"type": "character", "name": " 林夏 "}, true)
	if err != nil {
		t.Fatal(err)
	}
	if row.Type.String != "character" || row.Name != "林夏" {
		t.Fatalf("asset %+v", row)
	}
}

func TestImageTaskCreateRules(t *testing.T) {
	row := sqlcdb.AfvImageTask{}
	if _, err := applyImageTaskFields(&row, map[string]any{}, true); !isBadRequest(err) {
		t.Fatalf("empty prompt must fail, got %v", err)
	}
	if _, err := applyImageTaskFields(&row, map[string]any{"prompt": "一只猫", "status": float64(2)}, true); !isBadRequest(err) {
		t.Fatalf("create must not set status, got %v", err)
	}
	if _, err := applyImageTaskFields(&row, map[string]any{"prompt": "一只猫", "count": float64(0)}, true); !isBadRequest(err) {
		t.Fatalf("count 0 must fail, got %v", err)
	}
	_, err := applyImageTaskFields(&row, map[string]any{
		"prompt":       "一只猫",
		"refImageUrls": []any{"/upload/a.png"},
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	if row.Status.Int32 != 0 || row.Count.Int32 != 1 || string(row.RefImageUrls) != `["/upload/a.png"]` {
		t.Fatalf("image task %+v", row)
	}
}

func TestVideoTaskNeedsPromptOrFrame(t *testing.T) {
	row := sqlcdb.AfvVideoTask{}
	if _, err := applyVideoTaskFields(&row, map[string]any{"generateMode": "text2video"}, true); !isBadRequest(err) {
		t.Fatalf("empty video task must fail, got %v", err)
	}
	if _, err := applyVideoTaskFields(&row, map[string]any{"prompt": "走", "generateMode": "pan"}, true); !isBadRequest(err) {
		t.Fatalf("bad mode must fail, got %v", err)
	}
	_, err := applyVideoTaskFields(&row, map[string]any{
		"firstFrameImageUrl": "https://cdn.example/frame.png",
		"watermark":          false,
		"duration":           float64(4),
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	if !row.FirstFrameImageUrl.Valid || row.Watermark.Int16 != 0 || row.Duration.Int32 != 4 || row.Status.Int32 != 0 {
		t.Fatalf("video task %+v", row)
	}
}

func isBadRequest(err error) bool {
	var ae *apperr.Error
	return errors.As(err, &ae) && ae.HTTPStatus == 400
}
