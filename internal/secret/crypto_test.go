package secret

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// withKey 在测试期间设置固定的内存密钥，结束后清理。
func withKey(t *testing.T, key string) {
	t.Helper()
	ClearVaultKey()
	SetVaultKey(key)
	t.Cleanup(func() {
		ClearVaultKey()
	})
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	withKey(t, "test-key")

	tests := []struct {
		name  string
		plain string
	}{
		{"simple", "hello"},
		{"empty", ""},
		{"with spaces", "hello world"},
		{"with newline", "line1\nline2"},
		{"with unicode", "密码密码"},
		{"with special chars", "a!@#$%^&*()_+-={}[]|;:'\",.<>/?"},
		{"long", strings.Repeat("x", 10000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cipher, err := Encrypt(tt.plain)
			if err != nil {
				t.Fatalf("Encrypt error: %v", err)
			}
			if cipher == tt.plain {
				t.Fatal("cipher equals plain, encryption did nothing")
			}

			got, err := Decrypt(cipher)
			if err != nil {
				t.Fatalf("Decrypt error: %v", err)
			}
			if got != tt.plain {
				t.Errorf("round trip mismatch: got %q, want %q", got, tt.plain)
			}
		})
	}
}

func TestEncrypt_ProducesDifferentCiphertexts(t *testing.T) {
	withKey(t, "test-key")

	c1, err := Encrypt("same-input")
	if err != nil {
		t.Fatalf("Encrypt error: %v", err)
	}
	c2, err := Encrypt("same-input")
	if err != nil {
		t.Fatalf("Encrypt error: %v", err)
	}

	if c1 == c2 {
		t.Error("two encryptions of the same input produced identical ciphertext (nonce reuse?)")
	}
}

func TestDecrypt_WrongKeyFails(t *testing.T) {
	SetVaultKey("key-a")
	cipher, err := Encrypt("secret")
	if err != nil {
		t.Fatalf("Encrypt error: %v", err)
	}

	ClearVaultKey()
	SetVaultKey("key-b")
	defer ClearVaultKey()

	_, err = Decrypt(cipher)
	if err == nil {
		t.Fatal("expected error with wrong key, got nil")
	}
}

func TestDecrypt_InvalidBase64(t *testing.T) {
	withKey(t, "test-key")

	_, err := Decrypt("not-valid-base64!!!")
	if err == nil {
		t.Fatal("expected error for invalid base64, got nil")
	}
}

func TestDecrypt_TooShort(t *testing.T) {
	withKey(t, "test-key")

	// 短于 nonce 的合法 base64
	short := "YWJj" // "abc" in base64
	_, err := Decrypt(short)
	if err == nil {
		t.Fatal("expected error for short ciphertext, got nil")
	}
}

func TestEncrypt_NoKey(t *testing.T) {
	ClearVaultKey()
	// 清掉所有密钥来源，但 SecretKeyFile 是包级变量，需要临时置空
	oldFile := SecretKeyFile
	SecretKeyFile = ""
	defer func() { SecretKeyFile = oldFile }()

	// 无 stdin 输入时，loadRawKey 会尝试交互，无法在测试里完成
	// 这里只验证「有 key 时能工作」，不验证无 key 的行为
	t.Skip("skipping: requires interactive input")
}

func TestGenerateKeyFileDoesNotOverwriteWithoutForce(t *testing.T) {
	keyPath := filepath.Join(t.TempDir(), "vault-key")
	if err := GenerateKeyFile(keyPath, false); err != nil {
		t.Fatalf("GenerateKeyFile() error = %v", err)
	}
	original, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatal(err)
	}

	if err := GenerateKeyFile(keyPath, false); err == nil {
		t.Fatal("GenerateKeyFile() unexpectedly overwrote an existing key")
	}
	got, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(original) {
		t.Fatal("existing key contents changed without force")
	}
	assertKeyFilePermissions(t, keyPath)
}

func TestGenerateKeyFileForceReplacesAtomically(t *testing.T) {
	keyPath := filepath.Join(t.TempDir(), "vault-key")
	if err := GenerateKeyFile(keyPath, false); err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatal(err)
	}

	if err := GenerateKeyFile(keyPath, true); err != nil {
		t.Fatalf("GenerateKeyFile(force) error = %v", err)
	}
	replacement, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(replacement) == string(original) {
		t.Fatal("force did not replace the existing key")
	}
	assertKeyFilePermissions(t, keyPath)

	entries, err := os.ReadDir(filepath.Dir(keyPath))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("temporary key files remain: %v", entries)
	}
}

func TestGenerateKeyFileRejectsSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	link := filepath.Join(dir, "vault-key")
	if err := os.WriteFile(target, []byte("preserve"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("cannot create symbolic links: %v", err)
	}

	if err := GenerateKeyFile(link, true); err == nil {
		t.Fatal("GenerateKeyFile() followed a symbolic link")
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "preserve" {
		t.Fatalf("symbolic-link target changed: %q", got)
	}
}

func assertKeyFilePermissions(t *testing.T, path string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		return
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0600 {
		t.Errorf("key file permissions = %#o, want 0600", got)
	}
}
