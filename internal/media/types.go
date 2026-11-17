package media

import (
	"encoding/json"
	"time"
)

const (
	KindFFmpegVersion = "FFMPEG_VERSION"

	StatusQueued    = "QUEUED"
	StatusWaiting   = "WAITING"
	StatusRunning   = "RUNNING"
	StatusSucceeded = "SUCCEEDED"
	StatusFailed    = "FAILED"
	StatusCancelled = "CANCELLED"

	EventMediaJobAccepted = "MEDIA_JOB_ACCEPTED"
)

type Claim struct {
	ID    string
	Owner string
	Epoch int64
}

type Job struct {
	ID              string
	Kind            string
	ActorUserID     int64
	RequestID       string
	RequestHash     string
	InputJSON       json.RawMessage
	CheckpointJSON  json.RawMessage
	Status          string
	Phase           string
	ErrorMsg        *string
	OwnerInstanceID *string
	OwnerEpoch      int64
	CancelRequested bool
	Revision        int64
	DeadlineAt      time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (j Job) Claim() Claim {
	owner := ""
	if j.OwnerInstanceID != nil {
		owner = *j.OwnerInstanceID
	}
	return Claim{ID: j.ID, Owner: owner, Epoch: j.OwnerEpoch}
}

type EnqueueResult struct {
	ID        string `json:"id"`
	RequestID string `json:"requestId"`
	Kind      string `json:"kind"`
	Status    string `json:"status"`
	Replay    bool   `json:"replay"`
}

type JobView struct {
	ID             string          `json:"id"`
	Kind           string          `json:"kind"`
	Status         string          `json:"status"`
	Phase          string          `json:"phase"`
	CheckpointJSON json.RawMessage `json:"checkpointJson"`
	ErrorMsg       *string         `json:"errorMsg,omitempty"`
	Revision       int64           `json:"revision"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
}
