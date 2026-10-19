package ssh

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"mestrap.com/taskssh/internal/config"
)

func TestNewHostKeyCallback(t *testing.T) {
	_, trustedPrivateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	trustedSigner, err := ssh.NewSignerFromKey(trustedPrivateKey)
	if err != nil {
		t.Fatal(err)
	}

	_, otherPrivateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	otherSigner, err := ssh.NewSignerFromKey(otherPrivateKey)
	if err != nil {
		t.Fatal(err)
	}

	homeDir := t.TempDir()
	sshDir := filepath.Join(homeDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatal(err)
	}
	knownHostsFile := filepath.Join(sshDir, "known_hosts")
	contents := knownhosts.Line([]string{"example.com", "example.com:2222"}, trustedSigner.PublicKey()) + "\n"
	if err := os.WriteFile(knownHostsFile, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}

	callback, err := newHostKeyCallback(homeDir)
	if err != nil {
		t.Fatalf("newHostKeyCallback() error = %v", err)
	}

	remote := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 22}
	for _, address := range []string{"example.com:22", "example.com:2222"} {
		if err := callback(address, remote, trustedSigner.PublicKey()); err != nil {
			t.Errorf("callback(%q) rejected trusted key: %v", address, err)
		}
		if err := callback(address, remote, otherSigner.PublicKey()); err == nil {
			t.Errorf("callback(%q) accepted untrusted key", address)
		}
	}
}

func TestNewHostKeyCallbackRequiresKnownHostsFile(t *testing.T) {
	_, err := newHostKeyCallback(t.TempDir())
	if err == nil {
		t.Fatal("newHostKeyCallback() succeeded without a known_hosts file")
	}
}

func TestConnectContextCancelsSSHHandshake(t *testing.T) {
	home := t.TempDir()
	sshDir := filepath.Join(home, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sshDir, "known_hosts"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOME", home)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	accepted := make(chan net.Conn, 1)
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			accepted <- conn
		}
	}()

	_, portText, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := make(chan error, 1)
	go func() {
		client, err := ConnectContext(ctx, &config.Host{
			Host:     "127.0.0.1",
			Port:     port,
			Username: "test",
			Password: "test",
		}, 5*time.Second)
		if client != nil {
			client.Close()
		}
		result <- err
	}()

	var serverConn net.Conn
	select {
	case serverConn = <-accepted:
	case err := <-result:
		t.Fatalf("ConnectContext returned before TCP connection was accepted: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for TCP connection")
	}
	defer serverConn.Close()

	cancel()
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("ConnectContext error = %v, want context.Canceled", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("ConnectContext did not stop after cancellation")
	}
}
