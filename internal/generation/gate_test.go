package generation

import (
	"testing"
	"time"
)

func TestGateExecution(t *testing.T) {
	now := time.Date(2026, 11, 22, 0, 0, 0, 0, time.UTC)
	past := now.Add(-time.Minute)
	future := now.Add(time.Hour)

	tests := []struct {
		name      string
		permitted bool
		cancel    bool
		phase     string
		deadline  time.Time
		wantStop  bool
		wantState string
		wantCode  string
	}{
		{name: "echo continues", permitted: true, phase: PhasePreparing, deadline: future},
		{name: "ingesting ignores deadline", permitted: true, phase: PhaseIngesting, deadline: past},
		{name: "preparing deadline", permitted: true, phase: PhasePreparing, deadline: past, wantStop: true, wantState: StateFailed, wantCode: "GENERATION_DEADLINE_EXCEEDED"},
		{name: "polling deadline is unknown", permitted: true, phase: PhasePolling, deadline: past, wantStop: true, wantState: StateUnknownSubmission, wantCode: "GENERATION_DEADLINE_EXCEEDED"},
		{name: "cancel before submit", permitted: true, cancel: true, phase: PhaseNew, deadline: future, wantStop: true, wantState: StateCancelled, wantCode: "CANCELLED_BEFORE_SUBMISSION"},
		{name: "revoked identity before submit", permitted: false, phase: PhasePreparing, deadline: future, wantStop: true, wantState: StateCancelled, wantCode: "CANCELLED_BEFORE_SUBMISSION"},
		{name: "revoked identity while polling", permitted: false, phase: PhasePolling, deadline: future, wantStop: true, wantState: StateCancelled, wantCode: "EXECUTION_NOT_PERMITTED"},
		{name: "already cancelled polling keeps going", permitted: true, cancel: true, phase: PhasePolling, deadline: future},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stop, state, code, _ := gateExecution(tt.permitted, tt.cancel, tt.phase, now, tt.deadline)
			if stop != tt.wantStop || state != tt.wantState || code != tt.wantCode {
				t.Fatalf("got stop=%v state=%s code=%s", stop, state, code)
			}
		})
	}
}

func TestAgentRunIDFailClosed(t *testing.T) {
	if agentRunID([]byte(`{"origin":"http"}`)) != "" {
		t.Fatal("http echo context has no agent run")
	}
	if agentRunID([]byte(`{"agentRunId":" run-1 "}`)) != "run-1" {
		t.Fatal("agent run id should be trimmed")
	}
	if agentRunID([]byte(`{`)) == "" {
		t.Fatal("invalid context must not look like an ordinary user job")
	}
}
