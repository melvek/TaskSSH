package ssh

import (
	"bytes"
	"fmt"
	"io"

	"golang.org/x/crypto/ssh"
	"mestrap.com/taskssh/internal/console"
)

// ExecResult 是命令执行结果。
type ExecResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// Exec 执行命令，输出实时写入当前 goroutine 的输出目标
func (c *Client) Exec(command string) (*ExecResult, error) {

	if c.conn == nil {
		return nil, fmt.Errorf("client not connected")
	}

	session, err := c.conn.NewSession()
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = io.MultiWriter(console.CommandWriter(), &stdout)
	session.Stderr = io.MultiWriter(console.StderrWriter(), &stderr)

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
