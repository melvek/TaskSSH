package ssh

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Download 从远程下载文件到指定的本地路径。
//
// dry-run 时跳过实际下载。
func (c *Client) Download(remotePath, localPath string) error {
	if c.conn == nil {
		return fmt.Errorf("client not connected")
	}

	if remotePath == "" {
		return fmt.Errorf("remote path is empty")
	}
	if localPath == "" {
		return fmt.Errorf("local path is empty")
	}

	sftpClient, err := c.getSFTP()
	if err != nil {
		return err
	}

	info, err := sftpClient.Stat(remotePath)
	if err != nil {
		return fmt.Errorf("remote file %s: %w", remotePath, err)
	}
	if info.IsDir() {
		return fmt.Errorf("remote path is a directory: %s (please pack it first)", remotePath)
	}

	if c.dryRun {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return fmt.Errorf("create local dir: %w", err)
	}

	src, err := sftpClient.Open(remotePath)
	if err != nil {
		return fmt.Errorf("open remote: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("create local: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("copy to local: %w", err)
	}

	return nil
}
