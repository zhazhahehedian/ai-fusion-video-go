package media

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"sync"
	"time"
)

const defaultMaxOutputBytes = 1 << 20

// ProcessRunner 对齐 Java MediaProcessRunner：跑外部命令，期间可心跳/鉴租约，取消即杀进程。
type ProcessRunner struct {
	TickEvery      time.Duration
	Timeout        time.Duration
	MaxOutputBytes int
}

func NewProcessRunner() *ProcessRunner {
	return &ProcessRunner{
		TickEvery:      time.Second,
		Timeout:        30 * time.Minute,
		MaxOutputBytes: defaultMaxOutputBytes,
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
	output := newCappedBuffer(r.MaxOutputBytes)
	cmd.Stdout = output
	cmd.Stderr = output

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

	// killWait 必须等 Wait 返回后再读输出。超时后 Wait 仍可能往缓冲区写，
	// 缓冲区自己加锁，这里返回“进程仍未退出”，不再和 Wait 抢同一块裸 buffer。
	killWait := func() error {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		select {
		case err := <-done:
			return err
		case <-time.After(5 * time.Second):
			return errProcessLinger
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
				waitErr := killWait()
				if errors.Is(waitErr, errProcessLinger) {
					return "", fmt.Errorf("%w: %w", err, errProcessLinger)
				}
				return output.String(), err
			}
		case <-ctx.Done():
			waitErr := killWait()
			if errors.Is(waitErr, errProcessLinger) {
				return "", fmt.Errorf("%w: %w", ctx.Err(), errProcessLinger)
			}
			return output.String(), ctx.Err()
		}
	}
}

var errProcessLinger = errors.New("进程在终止后仍未退出")

type cappedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
	max int
}

func newCappedBuffer(max int) *cappedBuffer {
	if max <= 0 {
		max = defaultMaxOutputBytes
	}
	return &cappedBuffer{max: max}
}

func (c *cappedBuffer) Write(p []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	remain := c.max - c.buf.Len()
	if remain > 0 {
		if len(p) > remain {
			_, _ = c.buf.Write(p[:remain])
		} else {
			_, _ = c.buf.Write(p)
		}
	}
	return len(p), nil
}

func (c *cappedBuffer) String() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.buf.String()
}

func trimOut(s string) string {
	if len(s) <= 1500 {
		return s
	}
	return s[:1500]
}
