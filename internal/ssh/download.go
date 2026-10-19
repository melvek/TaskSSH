package ssh

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"mestrap.com/taskssh/internal/fileutil"
)

// Download 从远程下载文件到指定的本地路径。
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

	info, err := sftpClient.Lstat(remotePath)
	if err != nil {
		return fmt.Errorf("lstat remote file %s: %w", remotePath, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refusing to download remote symbolic link %s", remotePath)
	}
	if info.IsDir() {
		return fmt.Errorf("remote path is a directory: %s (please pack it first)", remotePath)
	}

	if err := fileutil.EnsureNoSymlinkPath(localPath); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return fmt.Errorf("create local dir: %w", err)
	}
	if err := fileutil.EnsureNoSymlinkPath(localPath); err != nil {
		return err
	}

	src, err := sftpClient.Open(remotePath)
	if err != nil {
		return fmt.Errorf("open remote: %w", err)
	}
	defer src.Close()

	dst, err := os.OpenFile(localPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("create local: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("copy to local: %w", err)
	}

	return nil
}
