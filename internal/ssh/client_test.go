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

func TestNewHostKeyCallbackCreatesKnownHostsFile(t *testing.T) {
	homeDir := t.TempDir()
	if _, err := newHostKeyCallback(homeDir); err != nil {
		t.Fatalf("newHostKeyCallback() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(homeDir, ".ssh", "known_hosts")); err != nil {
		t.Fatalf("known_hosts was not created: %v", err)
	}
}

func TestUnknownHostKeyPromptStoresTrustedKey(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), ".ssh", "known_hosts")
	promptCalls := 0
	callback, err := newHostKeyCallbackWithPrompt(path, func(hostname string, key ssh.PublicKey) (bool, error) {
		promptCalls++
		if hostname != "new.example.com:22" {
			t.Errorf("prompt hostname = %q, want %q", hostname, "new.example.com:22")
		}
		if string(key.Marshal()) != string(signer.PublicKey().Marshal()) {
			t.Error("prompt received unexpected host key")
		}
		return true, nil
	})
	if err != nil {
		t.Fatalf("newHostKeyCallbackWithPrompt() error = %v", err)
	}

	remote := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 22}
	if err := callback("new.example.com:22", remote, signer.PublicKey()); err != nil {
		t.Fatalf("callback() error = %v", err)
	}
	if err := callback("new.example.com:22", remote, signer.PublicKey()); err != nil {
		t.Fatalf("callback() rejected stored key: %v", err)
	}
	if promptCalls != 1 {
		t.Errorf("prompt called %d times, want 1", promptCalls)
	}
}

func TestUnknownHostKeyStoresNonDefaultPort(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), "known_hosts")
	promptCalls := 0
	callback, err := newHostKeyCallbackWithPrompt(path, func(string, ssh.PublicKey) (bool, error) {
		promptCalls++
		return true, nil
	})
	if err != nil {
		t.Fatalf("newHostKeyCallbackWithPrompt() error = %v", err)
	}
	remote := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 2222}
	if err := callback("new.example.com:2222", remote, signer.PublicKey()); err != nil {
		t.Fatalf("callback() error = %v", err)
	}
	if err := callback("new.example.com:2222", remote, signer.PublicKey()); err != nil {
		t.Fatalf("callback() rejected stored key: %v", err)
	}
	if promptCalls != 1 {
		t.Errorf("prompt called %d times, want 1", promptCalls)
	}
}

func TestUnknownHostKeyCanBeRejected(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), "known_hosts")
	callback, err := newHostKeyCallbackWithPrompt(path, func(string, ssh.PublicKey) (bool, error) {
		return false, nil
	})
	if err != nil {
		t.Fatalf("newHostKeyCallbackWithPrompt() error = %v", err)
	}
	remote := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 22}
	if err := callback("new.example.com:22", remote, signer.PublicKey()); err == nil {
		t.Fatal("callback() accepted a rejected host key")
	}

	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(contents) != 0 {
		t.Fatalf("known_hosts contains an untrusted key: %q", contents)
	}
}

func TestChangedHostKeyDoesNotPromptForTrust(t *testing.T) {
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

	path := filepath.Join(t.TempDir(), "known_hosts")
	if err := os.WriteFile(path, []byte(knownhosts.Line([]string{"example.com"}, trustedSigner.PublicKey())+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	promptCalled := false
	callback, err := newHostKeyCallbackWithPrompt(path, func(string, ssh.PublicKey) (bool, error) {
		promptCalled = true
		return true, nil
	})
	if err != nil {
		t.Fatalf("newHostKeyCallbackWithPrompt() error = %v", err)
	}
	remote := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 22}
	if err := callback("example.com:22", remote, otherSigner.PublicKey()); err == nil {
		t.Fatal("callback() accepted a changed host key")
	}
	if promptCalled {
		t.Fatal("callback() prompted to trust a changed host key")
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
