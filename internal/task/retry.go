package task

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"golang.org/x/crypto/ssh"

	"mestrap.com/taskssh/internal/console"
)

// retryDefaults 是重试的默认值。
const (
	defaultRetryTimeout = 30 // 秒
)

// shouldRetry 判断该 step 是否需要重试。
//
// 规则：interval > 0 时才启用重试。
func shouldRetry(interval int) bool {
	return interval > 0
}

// withRetry 带重试地执行 fn。
//
// 规则：
//   - fn 返回 nil：成功，立即返回
//   - fn 返回连接类错误：立即返回，不重试
//   - 距第一次执行超过 timeout：返回超时错误，带最后一次错误原因
//   - 否则：等待 interval 秒后重试
//   - ctx 取消：立即返回 ctx.Err()
//
// interval 必须 > 0，否则直接执行一次不重试。
func withRetry(ctx context.Context, fn func() error, interval, timeout int) error {
	if !shouldRetry(interval) {
		return fn()
	}

	if timeout <= 0 {
		timeout = defaultRetryTimeout
	}

	start := time.Now()
	attempt := 1

	for {
		err := fn()
		if err == nil {
			return nil
		}

		// 连接类错误，立即失败
		if isConnError(err) {
			return err
		}

		// 已超过总超时
		if time.Since(start) >= time.Duration(timeout)*time.Second {
			return fmt.Errorf("timeout after %ds (last error: %w)", timeout, err)
		}

		// Verbose 下输出重试信息
		console.Info("attempt %d failed: %v", attempt, err)

		select {
		case <-time.After(time.Duration(interval) * time.Second):
			attempt++
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// isConnError 判断错误是否为连接类错误。
//
// 连接类错误意味着连接已不可用，重试无意义，应立即失败。
// 命令非零退出（*ssh.ExitError）不属于连接错误。
func isConnError(err error) bool {
	if err == nil {
		return false
	}

	// ssh 命令非零退出：命令正常执行了，只是返回失败。不是连接错误。
	if _, ok := errors.AsType[*ssh.ExitError](err); ok {
		return false
	}

	// 服务器未返回退出状态，通常是连接异常关闭
	if _, ok := errors.AsType[*ssh.ExitMissingError](err); ok {
		return true
	}

	// io.EOF：连接被对端关闭
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}

	// net.Error：网络层错误（超时、连接重置、对端不可达等）
	if _, ok := errors.AsType[net.Error](err); ok {
		return true
	}

	return false
}
