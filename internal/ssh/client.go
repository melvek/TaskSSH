package ssh

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	"mestrap.com/taskssh/internal/config"
)

const DefaultPort = 22

const ConnectTimeout = 30 * time.Second

// Client 包装 SSH 连接。
type Client struct {
	conn *ssh.Client
	host *config.Host
}

// Connect 建立 SSH 连接。
func Connect(host *config.Host) (*Client, error) {
	if host.Host == "" {
		return nil, fmt.Errorf("host is empty")
	}

	authMethods, err := buildAuthMethods(host)
	if err != nil {
		return nil, err
	}

	port := host.Port
	if port == 0 {
		port = DefaultPort
	}

	sshConfig := &ssh.ClientConfig{
		User:            host.Username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         ConnectTimeout,
	}

	addr := net.JoinHostPort(host.Host, fmt.Sprintf("%d", port))
	conn, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("connect %s: %w", addr, err)
	}

	return &Client{conn: conn, host: host}, nil
}

// Close 关闭连接。
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Stat 返回远程路径的信息。
func (c *Client) Stat(remotePath string) (os.FileInfo, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("client not connected")
	}

	sftpClient, err := sftp.NewClient(c.conn)
	if err != nil {
		return nil, fmt.Errorf("create sftp: %w", err)
	}
	defer sftpClient.Close()

	info, err := sftpClient.Stat(remotePath)
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", remotePath, err)
	}
	return info, nil
}

// ListFiles 递归列出远程目录下的所有文件。
//
// 返回完整远程路径，不含目录项。
func (c *Client) ListFiles(root string) ([]string, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("client not connected")
	}

	sftpClient, err := sftp.NewClient(c.conn)
	if err != nil {
		return nil, fmt.Errorf("create sftp: %w", err)
	}
	defer sftpClient.Close()

	walker := sftpClient.Walk(root)

	var files []string
	for walker.Step() {
		if err := walker.Err(); err != nil {
			return nil, fmt.Errorf("walk %s: %w", root, err)
		}

		info := walker.Stat()
		if info == nil {
			continue
		}

		if info.IsDir() {
			continue
		}

		files = append(files, walker.Path())
	}

	return files, nil
}

// MkdirAll 在远程递归创建目录。
func (c *Client) MkdirAll(remotePath string) error {
	if c.conn == nil {
		return fmt.Errorf("client not connected")
	}

	sftpClient, err := sftp.NewClient(c.conn)
	if err != nil {
		return fmt.Errorf("create sftp: %w", err)
	}
	defer sftpClient.Close()

	if err := sftpClient.MkdirAll(remotePath); err != nil {
		return fmt.Errorf("mkdir %s: %w", remotePath, err)
	}
	return nil
}
