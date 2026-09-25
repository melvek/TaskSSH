package ssh

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/sftp"
)

// OverwritePolicy 是覆盖策略。
type OverwritePolicy int

const (
	// PolicyFail 目标已存在则报错（默认）。
	PolicyFail OverwritePolicy = iota
	// PolicyOverwrite 直接覆盖。
	PolicyOverwrite
)

// NewPolicy 根据 force 推导策略。
func NewPolicy(force bool) OverwritePolicy {
	if force {
		return PolicyOverwrite
	}
	return PolicyFail
}

// Upload 上传本地文件到远程，返回最终远程路径。
func (c *Client) Upload(localPath, remotePath string, policy OverwritePolicy) (string, error) {
	if c.conn == nil {
		return "", fmt.Errorf("client not connected")
	}

	info, err := os.Stat(localPath)
	if err != nil {
		return "", fmt.Errorf("local file: %w", err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("local path is a directory: %s", localPath)
	}

	sftpClient, err := c.getSFTP()
	if err != nil {
		return "", err
	}

	finalPath, err := resolveRemotePath(sftpClient, localPath, remotePath)
	if err != nil {
		return "", err
	}

	if exists(sftpClient, finalPath) {
		if err := handleExisting(finalPath, policy); err != nil {
			return "", err
		}
	}

	src, err := os.Open(localPath)
	if err != nil {
		return "", fmt.Errorf("open local: %w", err)
	}
	defer src.Close()

	dst, err := sftpClient.Create(finalPath)
	if err != nil {
		return "", fmt.Errorf("create remote %s: %w", finalPath, err)
	}
	defer dst.Close()

	if _, err := dst.ReadFrom(src); err != nil {
		return "", fmt.Errorf("copy to remote: %w", err)
	}

	return finalPath, nil
}

// handleExisting 处理已存在的远程文件。
func handleExisting(path string, policy OverwritePolicy) error {
	switch policy {
	case PolicyFail:
		return fmt.Errorf("remote file already exists: %s (use -F to overwrite)", path)
	case PolicyOverwrite:
		return nil
	}
	return nil
}

// resolveRemotePath 解析最终远程路径。
func resolveRemotePath(client *sftp.Client, localPath, remotePath string) (string, error) {
	if remotePath == "" {
		return "", fmt.Errorf("remote path is empty")
	}

	fileName := filepath.Base(localPath)

	if strings.HasSuffix(remotePath, "/") {
		return path.Join(remotePath, fileName), nil
	}

	if info, err := client.Stat(remotePath); err == nil && info.IsDir() {
		return path.Join(remotePath, fileName), nil
	}

	return remotePath, nil
}

// exists 检查远程路径是否存在。
func exists(client *sftp.Client, path string) bool {
	_, err := client.Stat(path)
	return err == nil
}
