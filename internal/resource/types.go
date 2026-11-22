package resource

import (
	"encoding/json"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/zhazhahehedian/ai-fusion-video-go/internal/apperr"
)

type Type string

const (
	TypeProject           Type = "project"
	TypeScript            Type = "script"
	TypeScriptEpisode     Type = "script_episode"
	TypeScriptScene       Type = "script_scene"
	TypeStoryboard        Type = "storyboard"
	TypeStoryboardEpisode Type = "storyboardEpisode"
	TypeStoryboardScene   Type = "storyboardScene"
	TypeShot              Type = "shot"
	TypeAsset             Type = "asset"
	TypeAssetItem         Type = "assetItem"
	TypeImageTask         Type = "imageTask"
	TypeVideoTask         Type = "videoTask"
)

type Operation string

const (
	OpCreate              Operation = "CREATE"
	OpPatch               Operation = "PATCH"
	OpDelete              Operation = "DELETE"
	OpReplaceScriptSource Operation = "REPLACE_SCRIPT_SOURCE"
	OpClearContent        Operation = "CLEAR_CONTENT"
)

type Origin string

const (
	OriginUser Origin = "user"
)

var requestIDPattern = regexp.MustCompile(`^[A-Za-z0-9._:-]+$`)

type Context struct {
	ActorUserID int64
	Origin      Origin
}

type Command interface {
	RequestID() string
	ResourceType() Type
	Operation() Operation
}

type CreateCommand struct {
	ReqID     string
	ResType   Type
	ProjectID *int64
	Fields    map[string]any
}

func (c CreateCommand) RequestID() string    { return c.ReqID }
func (c CreateCommand) ResourceType() Type   { return c.ResType }
func (c CreateCommand) Operation() Operation { return OpCreate }

type PatchCommand struct {
	ReqID   string
	ResType Type
	ID      int64
	Version int32
	Patch   map[string]any
	// ParentVersion 只在子资产更新时写入。空指针不进入哈希，避免改变已有命令的幂等摘要。
	ParentVersion *int32 `json:"parentVersion,omitempty"`
}

func (c PatchCommand) RequestID() string    { return c.ReqID }
func (c PatchCommand) ResourceType() Type   { return c.ResType }
func (c PatchCommand) Operation() Operation { return OpPatch }

type DeleteCommand struct {
	ReqID   string
	ResType Type
	ID      int64
	Version int32
	// ParentVersion 只在删除子资产时写入。空指针不进入哈希。
	ParentVersion *int32 `json:"parentVersion,omitempty"`
}

func (c DeleteCommand) RequestID() string    { return c.ReqID }
func (c DeleteCommand) ResourceType() Type   { return c.ResType }
func (c DeleteCommand) Operation() Operation { return OpDelete }

// SourceCommand replaces script raw content and clears derived episode/scene trees.
type SourceCommand struct {
	ReqID      string
	ResourceID int64
	Version    int32
	RawContent string
}

func (c SourceCommand) RequestID() string    { return c.ReqID }
func (c SourceCommand) ResourceType() Type   { return TypeScript }
func (c SourceCommand) Operation() Operation { return OpReplaceScriptSource }

// ClearCommand 清空分镜或分镜集的子层，保留父行。对齐 Java CLEAR。
type ClearCommand struct {
	ReqID   string
	ResType Type
	ID      int64
	Version int32
}

func (c ClearCommand) RequestID() string    { return c.ReqID }
func (c ClearCommand) ResourceType() Type   { return c.ResType }
func (c ClearCommand) Operation() Operation { return OpClearContent }

type Change struct {
	EventID       string   `json:"eventId"`
	ResourceType  Type     `json:"resourceType"`
	ResourceID    int64    `json:"resourceId"`
	ProjectID     *int64   `json:"projectId"`
	Version       int32    `json:"version"`
	Deleted       bool     `json:"deleted"`
	ChangedFields []string `json:"changedFields"`
	Before        any      `json:"-"`
	After         any      `json:"-"`
}

type Mutation struct {
	Result         any
	Changes        []Change
	PreviousValues any
}

type Receipt struct {
	CommandID string   `json:"commandId"`
	RequestID string   `json:"requestId"`
	Result    any      `json:"result"`
	Changes   []Change `json:"changes"`
}

func ValidateRequestID(requestID string) (string, error) {
	id := strings.TrimSpace(requestID)
	if id == "" {
		return "", apperr.BadRequest("requestId 不能为空")
	}
	if utf8.RuneCountInString(id) > 128 || !requestIDPattern.MatchString(id) {
		return "", apperr.BadRequest("requestId 格式无效")
	}
	return id, nil
}

func MustJSON(v any) json.RawMessage {
	if v == nil {
		return nil
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage("null")
	}
	return raw
}
