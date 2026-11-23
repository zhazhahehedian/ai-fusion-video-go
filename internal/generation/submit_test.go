package generation

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/apperr"
)

func TestSubmitRejectsScopePromptAndWorkflow(t *testing.T) {
	svc := NewService(nil)
	ctx := t.Context()
	actor := int64(7)
	scope := &ClientScope{UserID: actor}
	prompt := "一只猫"
	workflow := int64(3)
	count := int32(0)
	team := int64(9)

	cases := []struct {
		name string
		req  ImageSubmitRequest
		code int
		msg  string
	}{
		{name: "missing scope", req: ImageSubmitRequest{RequestID: "r", Prompt: &prompt}, code: 400, msg: "请先确认当前生成账号与团队"},
		{name: "other user", req: ImageSubmitRequest{Scope: &ClientScope{UserID: 8}, RequestID: "r", Prompt: &prompt}, code: 409, msg: "当前账号或团队已变化"},
		{name: "team set", req: ImageSubmitRequest{Scope: &ClientScope{UserID: actor, TeamID: &team}, RequestID: "r", Prompt: &prompt}, code: 409, msg: "当前账号或团队已变化"},
		{name: "blank request", req: ImageSubmitRequest{Scope: scope, Prompt: &prompt}, code: 400, msg: "生成请求标识不能为空"},
		{name: "blank prompt", req: ImageSubmitRequest{Scope: scope, RequestID: "r"}, code: 400, msg: "提示词不能为空"},
		{name: "count", req: ImageSubmitRequest{Scope: scope, RequestID: "r", Prompt: &prompt, Count: &count}, code: 400, msg: "count"},
		{name: "workflow", req: ImageSubmitRequest{Scope: scope, RequestID: "r", Prompt: &prompt, WorkflowVersionID: &workflow}, code: 400, msg: "ComfyUI"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.EnqueueImage(ctx, actor, tt.req)
			var ae *apperr.Error
			if !errors.As(err, &ae) || ae.Code != tt.code || !strings.Contains(ae.Msg, tt.msg) {
				t.Fatalf("got %v", err)
			}
		})
	}
}

func TestVideoSubmitNeedsPromptOrFrame(t *testing.T) {
	actor := int64(7)
	_, err := normalizeVideo(actor, VideoSubmitRequest{})
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.Code != 400 || !strings.Contains(ae.Msg, "提示词或首帧") {
		t.Fatalf("got %v", err)
	}
	frame := "https://example.com/first.png"
	row, err := normalizeVideo(actor, VideoSubmitRequest{FirstFrameImageURL: &frame})
	if err != nil {
		t.Fatal(err)
	}
	if !row.FirstFrameImageUrl.Valid || row.Count.Int32 != 1 || row.Status.Int32 != 0 {
		t.Fatalf("row = %+v", row)
	}
}

func TestImageRefURLAcceptsJSONString(t *testing.T) {
	raw := json.RawMessage(`"[\"https://example.com/a.png\"]"`)
	prompt := "一只猫"
	row, err := normalizeImage(4, ImageSubmitRequest{Prompt: &prompt, RefImageURLs: &raw})
	if err != nil {
		t.Fatal(err)
	}
	if string(row.RefImageUrls) != `["https://example.com/a.png"]` {
		t.Fatalf("ref = %s", row.RefImageUrls)
	}
	plain := json.RawMessage(`"https://example.com/a.png"`)
	if _, err := normalizeImage(4, ImageSubmitRequest{Prompt: &prompt, RefImageURLs: &plain}); err == nil {
		t.Fatal("单个地址字符串应被拒绝")
	}
}

func TestSubmitHashTracksPrompt(t *testing.T) {
	actor := int64(4)
	prompt := "一只猫"
	other := "一只狗"
	left, err := normalizeImage(actor, ImageSubmitRequest{Prompt: &prompt})
	if err != nil {
		t.Fatal(err)
	}
	right, err := normalizeImage(actor, ImageSubmitRequest{Prompt: &prompt})
	if err != nil {
		t.Fatal(err)
	}
	changed, err := normalizeImage(actor, ImageSubmitRequest{Prompt: &other})
	if err != nil {
		t.Fatal(err)
	}
	h1, err := imageRequestHash(actor, left)
	if err != nil {
		t.Fatal(err)
	}
	h2, err := imageRequestHash(actor, right)
	if err != nil {
		t.Fatal(err)
	}
	h3, err := imageRequestHash(actor, changed)
	if err != nil {
		t.Fatal(err)
	}
	if h1 != h2 || h1 == h3 || len(h1) != 64 {
		t.Fatalf("hash left=%s right=%s changed=%s", h1, h2, h3)
	}
}
