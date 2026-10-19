package ssh

import (
	"bytes"
	"context"
	"fmt"
	"io"

	gossh "golang.org/x/crypto/ssh"
	"mestrap.com/taskssh/internal/console"
)

// ExecResult 是命令执行结果。
type ExecResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// Exec 执行命令，输出实时写入默认 Logger。
func (c *Client) Exec(command string) (*ExecResult, error) {
	return c.ExecContext(context.Background(), command, console.Default)
}

// ExecContext 执行命令；ctx 取消时关闭 SSH 连接以中断远程操作。
func (c *Client) ExecContext(ctx context.Context, command string, logger *console.Logger) (*ExecResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if c.conn == nil {
		return nil, fmt.Errorf("client not connected")
	}

	stopCancel := context.AfterFunc(ctx, func() {
		c.conn.Close()
	})
	defer stopCancel()

	session, err := c.conn.NewSession()
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("create session: %w", err)
	}
	defer session.Close()

	if logger == nil {
		logger = console.Default
	}
	var stdout, stderr bytes.Buffer
	session.Stdout = io.MultiWriter(logger.CommandWriter(), &stdout)
	session.Stderr = io.MultiWriter(logger.StderrWriter(), &stderr)

	err = session.Run(command)
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	result := &ExecResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}

	if err != nil {
		if exitErr, ok := err.(*gossh.ExitError); ok {
			result.ExitCode = exitErr.ExitStatus()
			return result, nil
		}
		return nil, fmt.Errorf("run command: %w", err)
	}

	return result, nil
}
