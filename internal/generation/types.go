package generation

import (
	"encoding/json"
	"time"
)

const (
	KindEcho = "ECHO"

	StateQueued            = "QUEUED"
	StateReady             = "READY"
	StateRunning           = "RUNNING"
	StateWaiting           = "WAITING"
	StateSucceeded         = "SUCCEEDED"
	StateFailed            = "FAILED"
	StateCancelled         = "CANCELLED"
	StateUnknownSubmission = "UNKNOWN_SUBMISSION"
	StateNeedsReview       = "NEEDS_REVIEW"

	PhaseNew        = "NEW"
	PhasePreparing  = "PREPARING"
	PhaseSubmitting = "SUBMITTING"
	PhasePolling    = "POLLING"
	PhaseIngesting  = "INGESTING"
	PhaseFinished   = "FINISHED"

	EventGenerationAccepted = "GENERATION_ACCEPTED"

	MaxUserRemoteJobs = 8
)

type Claim struct {
	ID    string
	Owner string
	Epoch int64
}

type Job struct {
	ID              string
	SubmissionID    string
	ItemID          int64
	ItemIndex       int32
	ModelID         int64
	ActorUserID     int64
	ProjectID       *int64
	State           string
	Phase           string
	ProviderTaskID  *string
	ActiveAttemptID *string
	InputJSON       json.RawMessage
	ResultJSON      json.RawMessage
	CancelRequested bool
	PermitHeld      bool
	OwnerInstanceID *string
	OwnerEpoch      int64
	LeaseUntil      *time.Time
	NextPollAt      time.Time
	DeadlineAt      time.Time
	PollAttempts    int32
	Errors          int32
	ErrorCode       *string
	ErrorMessage    *string
	Revision        int64
}

func (j Job) Claim() Claim {
	owner := ""
	if j.OwnerInstanceID != nil {
		owner = *j.OwnerInstanceID
	}
	return Claim{ID: j.ID, Owner: owner, Epoch: j.OwnerEpoch}
}

type SubmissionView struct {
	ID           string          `json:"id"`
	RequestID    string          `json:"requestId"`
	Kind         string          `json:"kind"`
	State        string          `json:"state"`
	Revision     int64           `json:"revision"`
	PublicTaskID string          `json:"publicTaskId"`
	Prompt       string          `json:"prompt,omitempty"`
	Jobs         []JobView       `json:"jobs"`
	CreatedAt    time.Time       `json:"createdAt"`
	UpdatedAt    time.Time       `json:"updatedAt"`
	RawInput     json.RawMessage `json:"-"`
}

type JobView struct {
	ID             string          `json:"id"`
	State          string          `json:"state"`
	Phase          string          `json:"phase"`
	ProviderTaskID *string         `json:"providerTaskId,omitempty"`
	PollAttempts   int32           `json:"pollAttempts"`
	ResultJSON     json.RawMessage `json:"resultJson,omitempty"`
	ErrorCode      *string         `json:"errorCode,omitempty"`
	ErrorMessage   *string         `json:"errorMessage,omitempty"`
	Revision       int64           `json:"revision"`
}

type EchoEnqueueResult struct {
	ID        string `json:"id"`
	RequestID string `json:"requestId"`
	State     string `json:"state"`
	Kind      string `json:"kind"`
	Replay    bool   `json:"replay"`
}
