package media

import (
	"context"
	"errors"
	"runtime"
	"testing"
	"time"
)

func TestProcessRunnerCancelKillsProcess(t *testing.T) {
	runner := NewProcessRunner()
	runner.TickEvery = 50 * time.Millisecond
	runner.Timeout = 10 * time.Second

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	argv := []string{"sleep", "30"}
	if runtime.GOOS == "windows" {
		argv = []string{"powershell", "-NoProfile", "-Command", "Start-Sleep -Seconds 30"}
	}

	started := time.Now()
	go func() {
		time.Sleep(200 * time.Millisecond)
		cancel()
	}()

	_, err := runner.Run(ctx, argv, "", nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("want context.Canceled, got %v", err)
	}
	if time.Since(started) > 5*time.Second {
		t.Fatalf("cancel did not stop process quickly: %s", time.Since(started))
	}
}

func TestProcessRunnerOnTickStopsProcess(t *testing.T) {
	runner := NewProcessRunner()
	runner.TickEvery = 50 * time.Millisecond
	runner.Timeout = 10 * time.Second

	argv := []string{"sleep", "30"}
	if runtime.GOOS == "windows" {
		argv = []string{"powershell", "-NoProfile", "-Command", "Start-Sleep -Seconds 30"}
	}

	tickErr := errors.New("lease lost")
	started := time.Now()
	_, err := runner.Run(context.Background(), argv, "", func() error {
		return tickErr
	})
	if !errors.Is(err, tickErr) {
		t.Fatalf("want tick error, got %v", err)
	}
	if time.Since(started) > 5*time.Second {
		t.Fatalf("onTick did not stop process quickly: %s", time.Since(started))
	}
}
