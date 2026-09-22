package ssh

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/sftp"
)

// Download 从远程下载文件到本地，返回最终本地路径。
func (c *Client) Download(remotePath, localPath string) (string, error) {
	if c.conn == nil {
		return "", fmt.Errorf("client not connected")
	}

	sftpClient, err := sftp.NewClient(c.conn)
	if err != nil {
		return "", fmt.Errorf("create sftp: %w", err)
	}
	defer sftpClient.Close()

	info, err := sftpClient.Stat(remotePath)
	if err != nil {
		return "", fmt.Errorf("remote file %s: %w", remotePath, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("remote path is a directory: %s", remotePath)
	}

	finalPath := resolveLocalPath(remotePath, localPath)

	if err := os.MkdirAll(filepath.Dir(finalPath), 0755); err != nil {
		return "", fmt.Errorf("create local dir: %w", err)
	}

	src, err := sftpClient.Open(remotePath)
	if err != nil {
		return "", fmt.Errorf("open remote: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(finalPath)
	if err != nil {
		return "", fmt.Errorf("create local: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("copy to local: %w", err)
	}

	return finalPath, nil
}

// resolveLocalPath 解析最终本地路径。
func resolveLocalPath(remotePath, localPath string) string {
	fileName := path.Base(remotePath)

	if strings.HasSuffix(localPath, "/") || strings.HasSuffix(localPath, string(os.PathSeparator)) {
		return filepath.Join(localPath, fileName)
	}
	return localPath
}
