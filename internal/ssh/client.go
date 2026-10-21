package ssh

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"golang.org/x/term"

	"mestrap.com/taskssh/internal/config"
	"mestrap.com/taskssh/internal/prompt"
)

// DefaultPort 是默认 SSH 端口。
const DefaultPort = 22

// DefaultConnectTimeout 是默认连接超时。
const DefaultConnectTimeout = 10 * time.Second

var hostKeyPromptMu sync.Mutex

// Client 包装 SSH 连接。
type Client struct {
	conn       *ssh.Client
	sftp       *sftp.Client
	host       *config.Host
	stopCancel func() bool
}

// Connect 建立 SSH 连接。
//
// timeout 为 0 时使用 DefaultConnectTimeout。
func Connect(host *config.Host, timeout time.Duration) (*Client, error) {
	return ConnectContext(context.Background(), host, timeout)
}

// ConnectContext establishes an SSH connection that is interrupted when ctx is canceled.
func ConnectContext(ctx context.Context, host *config.Host, timeout time.Duration) (*Client, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

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

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve home directory for SSH known_hosts: %w", err)
	}
	hostKeyCallback, err := newHostKeyCallback(home)
	if err != nil {
		return nil, err
	}

	sshConfig := &ssh.ClientConfig{
		User:            host.Username,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback,
		Timeout:         timeout,
	}

	addr := net.JoinHostPort(host.Host, fmt.Sprintf("%d", port))
	rawConn, err := (&net.Dialer{Timeout: timeout}).DialContext(ctx, "tcp", addr)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("connect %s: %w", addr, err)
	}

	if err := rawConn.SetDeadline(time.Now().Add(timeout)); err != nil {
		rawConn.Close()
		return nil, fmt.Errorf("set SSH handshake deadline: %w", err)
	}
	stopHandshakeCancel := context.AfterFunc(ctx, func() {
		rawConn.Close()
	})
	sshConn, chans, reqs, err := ssh.NewClientConn(rawConn, addr, sshConfig)
	stopHandshakeCancel()
	if ctx.Err() != nil {
		rawConn.Close()
		return nil, ctx.Err()
	}
	if err != nil {
		rawConn.Close()
		return nil, fmt.Errorf("connect %s: %w", addr, err)
	}
	if err := rawConn.SetDeadline(time.Time{}); err != nil {
		rawConn.Close()
		return nil, fmt.Errorf("clear SSH handshake deadline: %w", err)
	}

	client := &Client{
		conn: ssh.NewClient(sshConn, chans, reqs),
		host: host,
	}
	client.stopCancel = context.AfterFunc(ctx, func() {
		rawConn.Close()
	})
	return client, nil
}

func newHostKeyCallback(homeDir string) (ssh.HostKeyCallback, error) {
	path := filepath.Join(homeDir, ".ssh", "known_hosts")
	return newHostKeyCallbackWithPrompt(path, confirmUnknownHostKey)
}

func newHostKeyCallbackWithPrompt(path string, confirm func(string, ssh.PublicKey) (bool, error)) (ssh.HostKeyCallback, error) {
	if err := ensureKnownHostsFile(path); err != nil {
		return nil, err
	}

	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		hostKeyPromptMu.Lock()
		defer hostKeyPromptMu.Unlock()

		callback, err := knownhosts.New(path)
		if err != nil {
			return fmt.Errorf("load SSH known_hosts %s: %w", path, err)
		}
		err = callback(hostname, remote, key)
		if err == nil {
			return nil
		}

		var keyErr *knownhosts.KeyError
		if !errors.As(err, &keyErr) || len(keyErr.Want) != 0 {
			return err
		}

		trusted, err := confirm(hostname, key)
		if err != nil {
			return fmt.Errorf("confirm SSH host key for %s: %w", hostname, err)
		}
		if !trusted {
			return fmt.Errorf("SSH host key for %s was not trusted", hostname)
		}

		if err := appendKnownHost(path, hostname, key); err != nil {
			return err
		}
		return nil
	}, nil
}

func ensureKnownHostsFile(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create SSH directory %s: %w", dir, err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return fmt.Errorf("open SSH known_hosts %s: %w", path, err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close SSH known_hosts %s: %w", path, err)
	}
	return nil
}

func appendKnownHost(path, hostname string, key ssh.PublicKey) (retErr error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("open SSH known_hosts %s for writing: %w", path, err)
	}
	defer func() {
		if err := file.Close(); err != nil && retErr == nil {
			retErr = fmt.Errorf("close SSH known_hosts %s: %w", path, err)
		}
	}()

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("inspect SSH known_hosts %s: %w", path, err)
	}
	if info.Size() > 0 {
		last := make([]byte, 1)
		if _, err := file.ReadAt(last, info.Size()-1); err != nil {
			return fmt.Errorf("read SSH known_hosts %s: %w", path, err)
		}
		if last[0] != '\n' {
			if _, err := file.Write([]byte("\n")); err != nil {
				return fmt.Errorf("append to SSH known_hosts %s: %w", path, err)
			}
		}
	}

	line := knownhosts.Line([]string{knownhosts.Normalize(hostname)}, key) + "\n"
	if _, err := file.WriteString(line); err != nil {
		return fmt.Errorf("append to SSH known_hosts %s: %w", path, err)
	}
	return nil
}

func confirmUnknownHostKey(hostname string, key ssh.PublicKey) (bool, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return false, fmt.Errorf("no interactive terminal; verify the host key and add it to known_hosts")
	}

	fmt.Fprintf(os.Stderr, "The authenticity of host %s (%s) can't be established.\n", hostname, key.Type())
	fmt.Fprintf(os.Stderr, "SHA256 fingerprint: %s\n", ssh.FingerprintSHA256(key))
	fmt.Fprintln(os.Stderr, "Only continue if you trust this server and have verified its fingerprint.")
	answer, err := prompt.Line("Trust this host key and continue? (yes/no): ")
	if err != nil {
		return false, err
	}
	if strings.EqualFold(answer, "yes") {
		return true, nil
	}
	return false, nil
}

// Close 关闭 SFTP 会话和 SSH 连接。
func (c *Client) Close() error {
	if c.stopCancel != nil {
		c.stopCancel()
		c.stopCancel = nil
	}
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

// 返回远程路径的信息
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

// Lstat 返回远程路径本身的信息，不跟随符号链接。
func (c *Client) Lstat(remotePath string) (os.FileInfo, error) {
	sftpClient, err := c.getSFTP()
	if err != nil {
		return nil, err
	}

	info, err := sftpClient.Lstat(remotePath)
	if err != nil {
		return nil, fmt.Errorf("lstat %s: %w", remotePath, err)
	}
	return info, nil
}

// 递归列出远程目录下的所有文件
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
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("refusing to download remote symbolic link %s", walker.Path())
		}

		if info.IsDir() {
			continue
		}

		files = append(files, walker.Path())
	}

	return files, nil
}

// 在远程递归创建目录
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
