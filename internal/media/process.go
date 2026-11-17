package media

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"
)

// ProcessRunner 对齐 Java MediaProcessRunner：跑外部命令，期间可心跳/鉴租约，取消即杀进程。
type ProcessRunner struct {
	TickEvery time.Duration
	Timeout   time.Duration
}

func NewProcessRunner() *ProcessRunner {
	return &ProcessRunner{
		TickEvery: time.Second,
		Timeout:   30 * time.Minute,
	}
}

// Run 在 workDir 下执行 argv；onTick 每秒调用（用于 lease check）。ctx 取消或 onTick 失败会 Kill 进程。
func (r *ProcessRunner) Run(ctx context.Context, argv []string, workDir string, onTick func() error) (string, error) {
	if len(argv) == 0 {
		return "", fmt.Errorf("命令不能为空")
	}
	if r.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.Timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Dir = workDir
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("无法启动进程 %q: %w", argv[0], err)
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	tickEvery := r.TickEvery
	if tickEvery <= 0 {
		tickEvery = time.Second
	}
	ticker := time.NewTicker(tickEvery)
	defer ticker.Stop()

	killWait := func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		select {
		case <-done:
		case <-time.After(5 * time.Second):
		}
	}

	for {
		select {
		case err := <-done:
			out := output.String()
			if err != nil {
				if errors.Is(ctx.Err(), context.Canceled) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
					return out, ctx.Err()
				}
				return out, fmt.Errorf("进程退出失败: %w; 输出: %s", err, trimOut(out))
			}
			return out, nil
		case <-ticker.C:
			if onTick == nil {
				continue
			}
			if err := onTick(); err != nil {
				killWait()
				return output.String(), err
			}
		case <-ctx.Done():
			killWait()
			return output.String(), ctx.Err()
		}
	}
}

func trimOut(s string) string {
	if len(s) <= 1500 {
		return s
	}
	return s[:1500]
}
