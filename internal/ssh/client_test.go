package ssh

import (
	"crypto/ed25519"
	"crypto/rand"
	"net"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
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
