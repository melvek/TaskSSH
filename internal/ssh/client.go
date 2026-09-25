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

// DefaultPort 是默认 SSH 端口。
const DefaultPort = 22

// DefaultConnectTimeout 是默认连接超时。
const DefaultConnectTimeout = 10 * time.Second

// Client 包装 SSH 连接。
type Client struct {
	conn *ssh.Client
	sftp *sftp.Client
	host *config.Host
}

// Connect 建立 SSH 连接。
//
// timeout 为 0 时使用 DefaultConnectTimeout。
func Connect(host *config.Host, timeout time.Duration) (*Client, error) {
	if host.Host == "" {
		return nil, fmt.Errorf("host is empty")
	}

	if timeout <= 0 {
		timeout = DefaultConnectTimeout
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
		Timeout:         timeout,
	}

	addr := net.JoinHostPort(host.Host, fmt.Sprintf("%d", port))
	conn, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("connect %s: %w", addr, err)
	}

	return &Client{conn: conn, host: host}, nil
}

// Close 关闭 SFTP 会话和 SSH 连接。
func (c *Client) Close() error {
	if c.sftp != nil {
		c.sftp.Close()
		c.sftp = nil
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// getSFTP 返回复用的 SFTP 会话，首次调用时创建。
func (c *Client) getSFTP() (*sftp.Client, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("client not connected")
	}
	if c.sftp != nil {
		return c.sftp, nil
	}

	s, err := sftp.NewClient(c.conn)
	if err != nil {
		return nil, fmt.Errorf("create sftp: %w", err)
	}
	c.sftp = s
	return c.sftp, nil
}

// Stat 返回远程路径的信息。
func (c *Client) Stat(remotePath string) (os.FileInfo, error) {
	sftpClient, err := c.getSFTP()
	if err != nil {
		return nil, err
	}

	info, err := sftpClient.Stat(remotePath)
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", remotePath, err)
	}
	return info, nil
}

// ListFiles 递归列出远程目录下的所有文件。
func (c *Client) ListFiles(root string) ([]string, error) {
	sftpClient, err := c.getSFTP()
	if err != nil {
		return nil, err
	}

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
	sftpClient, err := c.getSFTP()
	if err != nil {
		return err
	}

	if err := sftpClient.MkdirAll(remotePath); err != nil {
		return fmt.Errorf("mkdir %s: %w", remotePath, err)
	}
	return nil
}
