package ssh

import (
	"bytes"
	"fmt"
	"io"

	"golang.org/x/crypto/ssh"

	"mestrap.com/taskssh/internal/log"
)

// ExecResult 是命令执行结果。
type ExecResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// Exec 执行命令，输出实时写入当前 goroutine 的输出目标。
func (c *Client) Exec(command string) (*ExecResult, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("client not connected")
	}

	session, err := c.conn.NewSession()
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	defer session.Close()

	// 当前 goroutine 的输出目标
	out := log.GetOutput()

	var stdout, stderr bytes.Buffer

	// 同时写到输出目标和缓冲
	session.Stdout = io.MultiWriter(out, &stdout)
	session.Stderr = io.MultiWriter(out, &stderr)

	err = session.Run(command)

	result := &ExecResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}

	if err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			result.ExitCode = exitErr.ExitStatus()
			return result, nil
		}
		return nil, fmt.Errorf("run command: %w", err)
	}

	return result, nil
}
