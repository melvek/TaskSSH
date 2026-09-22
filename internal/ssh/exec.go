package ssh

import (
	"bytes"
	"fmt"
	"io"

	"golang.org/x/crypto/ssh"
)

// ExecResult 是命令执行结果。
type ExecResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// Exec 在远程执行命令，返回结果。
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
	session.Stdout = &stdout
	session.Stderr = &stderr

	err = session.Run(command)

	result := &ExecResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}

	if err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			result.ExitCode = exitErr.ExitStatus()
			return result, nil // 命令失败不算 Go 层的错误
		}
		return nil, fmt.Errorf("run command: %w", err)
	}

	return result, nil
}

// ExecStream 执行命令并实时输出到 writer。
func (c *Client) ExecStream(command string, out io.Writer) error {
	if c.conn == nil {
		return fmt.Errorf("client not connected")
	}

	session, err := c.conn.NewSession()
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	defer session.Close()

	session.Stdout = out
	session.Stderr = out

	return session.Run(command)
}
