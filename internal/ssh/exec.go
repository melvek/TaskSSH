package ssh

import (
	"bytes"
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"
)

// ExecResult 是命令执行结果。
type ExecResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// Exec 在远程执行命令，输出实时打印到终端。
func (c *Client) Exec(command string) (*ExecResult, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("client not connected")
	}

	session, err := c.conn.NewSession()
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	defer session.Close()

	// 实时输出到终端
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	err = session.Run(command)

	result := &ExecResult{}
	if err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			result.ExitCode = exitErr.ExitStatus()
			return result, nil
		}
		return nil, fmt.Errorf("run command: %w", err)
	}

	return result, nil
}

// ExecBuffered 执行命令，收集输出后返回。
func (c *Client) ExecBuffered(command string) (*ExecResult, error) {
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
			return result, nil
		}
		return nil, fmt.Errorf("run command: %w", err)
	}

	return result, nil
}
