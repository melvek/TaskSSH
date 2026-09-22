package ssh

import (
	"fmt"
	"net"
	"time"

	"golang.org/x/crypto/ssh"

	"mestrap.com/taskssh/internal/config"
)

// DefaultPort 是默认 SSH 端口。
const DefaultPort = 22

// ConnectTimeout 是连接超时。
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
