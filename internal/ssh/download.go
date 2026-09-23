package ssh

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/sftp"
)

// Download 从远程下载文件到指定的本地路径。
//
// localPath 必须是完整的文件路径（含文件名），调用方负责：
//   - 解析 -d 参数（目录 / 文件）
//   - 目录时拼接主机子目录
//   - 目录时拼接远程文件名
//
// 本方法只做三件事：
//   - 校验远程文件存在且不是目录
//   - 创建本地父目录
//   - 复制文件内容
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

	sftpClient, err := sftp.NewClient(c.conn)
	if err != nil {
		return fmt.Errorf("create sftp: %w", err)
	}
	defer sftpClient.Close()

	info, err := sftpClient.Stat(remotePath)
	if err != nil {
		return fmt.Errorf("remote file %s: %w", remotePath, err)
	}
	if info.IsDir() {
		return fmt.Errorf("remote path is a directory: %s (please pack it first)", remotePath)
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
