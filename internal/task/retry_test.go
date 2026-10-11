package task

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"

	"mestrap.com/taskssh/internal/console"
)

func TestMain(m *testing.M) {
	// 测试期间静默 console，避免 withRetry 的输出污染测试日志
	console.SetLevel(console.LevelQuiet)
	m.Run()
}

// ==================== shouldRetry ====================

func TestShouldRetry(t *testing.T) {
	tests := []struct {
		name     string
		interval int
		want     bool
	}{
		{"zero", 0, false},
		{"negative", -1, false},
		{"positive", 1, true},
		{"large", 60, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldRetry(tt.interval); got != tt.want {
				t.Errorf("shouldRetry(%d) = %v, want %v", tt.interval, got, tt.want)
			}
		})
	}
}

// ==================== isConnError ====================

func TestIsConnError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil",
			err:  nil,
			want: false,
		},
		{
			name: "plain error",
			err:  errors.New("something failed"),
			want: false,
		},
		{
			name: "ssh exit error (command failed, not conn)",
			err:  &ssh.ExitError{},
			want: false,
		},
		{
			name: "wrapped ssh exit error",
			err:  wrap(&ssh.ExitError{}, "step failed"),
			want: false,
		},
		{
			name: "ssh exit missing error (conn closed)",
			err:  &ssh.ExitMissingError{},
			want: true,
		},
		{
			name: "io.EOF",
			err:  io.EOF,
			want: true,
		},
		{
			name: "io.ErrUnexpectedEOF",
			err:  io.ErrUnexpectedEOF,
			want: true,
		},
		{
			name: "wrapped io.EOF",
			err:  wrap(io.EOF, "read failed"),
			want: true,
		},
		{
			name: "net.OpError",
			err:  &net.OpError{Op: "dial", Err: errors.New("refused")},
			want: true,
		},
		{
			name: "wrapped net.OpError",
			err:  wrap(&net.OpError{Op: "dial", Err: errors.New("refused")}, "connect"),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isConnError(tt.err); got != tt.want {
				t.Errorf("isConnError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

// wrap 用 %w 包装错误，模拟实际代码里的错误链。
func wrap(err error, msg string) error {
	return fmt.Errorf("%s: %w", msg, err)
}

// ==================== withRetry ====================

// noopCtx 返回一个不会取消的 ctx。
func noopCtx() context.Context {
	return context.Background()
}

func TestWithRetry_NoRetryWhenIntervalZero(t *testing.T) {
	calls := 0
	fn := func() error {
		calls++
		return errors.New("fail")
	}

	err := withRetry(noopCtx(), fn, 0, 10)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if calls != 1 {
		t.Errorf("fn called %d times, want 1", calls)
	}
}

func TestWithRetry_SuccessOnFirstTry(t *testing.T) {
	calls := 0
	fn := func() error {
		calls++
		return nil
	}

	err := withRetry(noopCtx(), fn, 1, 10)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Errorf("fn called %d times, want 1", calls)
	}
}

func TestWithRetry_SuccessAfterRetries(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping: requires ~2s of real waiting")
	}

	calls := 0
	fn := func() error {
		calls++
		if calls < 3 {
			return errors.New("transient failure")
		}
		return nil
	}

	start := time.Now()
	err := withRetry(noopCtx(), fn, 1, 10)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if calls != 3 {
		t.Errorf("fn called %d times, want 3", calls)
	}
	// 两次 1 秒等待，应至少 2 秒
	if elapsed < 1900*time.Millisecond {
		t.Errorf("elapsed %v, want >= ~2s", elapsed)
	}
}

func TestWithRetry_ConnErrorReturnsImmediately(t *testing.T) {
	calls := 0
	fn := func() error {
		calls++
		return io.EOF
	}

	start := time.Now()
	err := withRetry(noopCtx(), fn, 1, 10)
	elapsed := time.Since(start)

	if !errors.Is(err, io.EOF) {
		t.Errorf("err = %v, want io.EOF", err)
	}
	if calls != 1 {
		t.Errorf("fn called %d times, want 1 (no retry on conn error)", calls)
	}
	if elapsed > 500*time.Millisecond {
		t.Errorf("elapsed %v, want < 500ms (should not wait)", elapsed)
	}
}

func TestWithRetry_Timeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping: requires ~1s of real waiting")
	}

	calls := 0
	fn := func() error {
		calls++
		return errors.New("always fails")
	}

	start := time.Now()
	err := withRetry(noopCtx(), fn, 1, 1)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	// 至少执行了一次，等待期间超时
	if calls < 2 {
		t.Errorf("fn called %d times, want >= 2", calls)
	}
	if elapsed < 900*time.Millisecond {
		t.Errorf("elapsed %v, want >= ~1s", elapsed)
	}
}

func TestWithRetry_ContextCancelDuringWait(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// 100ms 后取消
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	calls := 0
	fn := func() error {
		calls++
		return errors.New("fail")
	}

	start := time.Now()
	err := withRetry(ctx, fn, 1, 10) // interval=1s，会在等待期间被取消
	elapsed := time.Since(start)

	if !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
	if calls != 1 {
		t.Errorf("fn called %d times, want 1", calls)
	}
	if elapsed > 500*time.Millisecond {
		t.Errorf("elapsed %v, want < 500ms (cancel should interrupt wait)", elapsed)
	}
}

func TestWithRetry_TimeoutDefaultWhenZero(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping: requires real waiting")
	}

	// timeout=0 应使用默认 30 秒
	// 只验证「不会在 1 秒内超时」即可，避免等 30 秒
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()

	fn := func() error {
		return errors.New("fail")
	}

	err := withRetry(ctx, fn, 1, 0)

	// ctx 超时，而不是 retry timeout
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("err = %v, want context.DeadlineExceeded (ctx cancel before retry timeout)", err)
	}
}
