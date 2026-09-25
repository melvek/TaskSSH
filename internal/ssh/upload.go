package ssh

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/sftp"

	"mestrap.com/taskssh/internal/log"
)

// OverwritePolicy 是覆盖策略。
type OverwritePolicy int

const (
	// PolicyFail 目标已存在则报错（默认）。
	PolicyFail OverwritePolicy = iota
	// PolicyOverwrite 直接覆盖。
	PolicyOverwrite
	// PolicyBackup 备份后覆盖。
	PolicyBackup
)

// NewPolicy 根据 force / backup 推导策略。
func NewPolicy(force, backup bool) OverwritePolicy {
	if !force {
		return PolicyFail
	}
	if backup {
		return PolicyBackup
	}
	return PolicyOverwrite
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
		if err := handleExisting(sftpClient, finalPath, policy); err != nil {
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
func handleExisting(client *sftp.Client, path string, policy OverwritePolicy) error {
	switch policy {
	case PolicyFail:
		return fmt.Errorf("remote file already exists: %s (use -F to overwrite, -F -B to backup)",
			path)
	case PolicyOverwrite:
		return nil
	case PolicyBackup:
		backupPath, err := backupRemote(client, path)
		if err != nil {
			return fmt.Errorf("backup %s: %w", path, err)
		}
		log.Info("Backed up: %s -> %s", path, backupPath)
		return nil
	}
	return nil
}

// backupRemote 备份远程文件，追加时间戳。
func backupRemote(client *sftp.Client, path string) (string, error) {
	dir := remoteDir(path)
	name := remoteBase(path)

	base, ext := splitExt(name)
	timestamp := time.Now().Format("20060102150405")
	backupName := fmt.Sprintf("%s.%s%s", base, timestamp, ext)
	backupPath := path2(dir, backupName)

	if err := client.Rename(path, backupPath); err != nil {
		return "", err
	}
	return backupPath, nil
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

// remoteDir 取远程路径的目录部分。
func remoteDir(p string) string {
	idx := strings.LastIndex(p, "/")
	if idx <= 0 {
		return "/"
	}
	return p[:idx]
}

// remoteBase 取远程路径的文件名部分。
func remoteBase(p string) string {
	idx := strings.LastIndex(p, "/")
	if idx < 0 {
		return p
	}
	return p[idx+1:]
}

// splitExt 分离文件名和扩展名。
func splitExt(name string) (string, string) {
	idx := strings.LastIndex(name, ".")
	if idx <= 0 {
		return name, ""
	}
	return name[:idx], name[idx:]
}

// path2 拼接远程路径。
func path2(dir, name string) string {
	if strings.HasSuffix(dir, "/") {
		return dir + name
	}
	return dir + "/" + name
}
