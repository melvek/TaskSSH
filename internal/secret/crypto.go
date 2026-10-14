package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/term"
	"mestrap.com/taskssh/internal/fileutil"
)

// SecretKeyFile 记录 --secret-key-file 的值，由 cmd 层设置。
var SecretKeyFile = ""

// DefaultKeyFile 是默认密钥文件路径。
const DefaultKeyFile = "~/.taskssh/vault-key"

// EnvKeyFile 是密钥文件路径的环境变量名。
const EnvKeyFile = "TASKSSH_SECRET_KEY_FILE"

// vaultKey 缓存本次会话的密钥，避免重复提示。
var vaultKey string

// SetVaultKey 设置本次会话使用的密钥。
func SetVaultKey(key string) {
	vaultKey = key
}

// ClearVaultKey 清除缓存的密钥。
func ClearVaultKey() {
	vaultKey = ""
}

// newGCM 统一创建 AES-GCM。
func newGCM() (cipher.AEAD, error) {
	raw, err := LoadVaultKey()
	if err != nil {
		return nil, err
	}

	key := sha256.Sum256([]byte(raw))

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	return gcm, nil
}

// Encrypt 加密明文，返回 Base64 编码的密文。
func Encrypt(plain string) (string, error) {
	gcm, err := newGCM()
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plain), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密 Base64 编码的密文
func Decrypt(encoded string) (string, error) {
	gcm, err := newGCM()
	if err != nil {
		return "", err
	}

	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("decode base64: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}

	return string(plain), nil
}

// 获取原始密钥字符串
func LoadVaultKey() (string, error) {
	if vaultKey != "" {
		return vaultKey, nil
	}

	if path, ok := configuredKeyFile(); ok {
		return readKeyFile(path)
	}

	return promptKeyPrompt("Secret key: ")
}

// 返回已配置的密钥文件路径
func configuredKeyFile() (string, bool) {
	if SecretKeyFile != "" {
		return SecretKeyFile, true
	}

	if path := os.Getenv(EnvKeyFile); path != "" {
		return path, true
	}

	if KeyFileExists(DefaultKeyFile) {
		return DefaultKeyFile, true
	}

	return "", false
}

// 从文件读密钥
func readKeyFile(path string) (string, error) {
	expanded, err := checkKeyFile(path)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(expanded)
	if err != nil {
		return "", fmt.Errorf("read key file %s: %w", expanded, err)
	}

	key := strings.TrimSpace(string(data))
	if key == "" {
		return "", fmt.Errorf("key file %s is empty", expanded)
	}

	return key, nil
}

func checkKeyFile(path string) (string, error) {
	expanded := fileutil.ExpandHome(path)

	info, err := os.Stat(expanded)
	if err != nil {
		return expanded, fmt.Errorf("stat key file %s: %w", expanded, err)
	}

	if info.IsDir() {
		return expanded, fmt.Errorf("key file %s is a directory", expanded)
	}

	return expanded, nil
}

func KeyFileExists(path string) bool {
	info, err := os.Stat(fileutil.ExpandHome(path))
	return err == nil && !info.IsDir()
}

func PromptKeyConfirm() (string, error) {
	first, err := promptKeyPrompt("Secret key: ")
	if err != nil {
		return "", err
	}

	second, err := promptKeyPrompt("Confirm secret key: ")
	if err != nil {
		return "", err
	}

	if first != second {
		return "", fmt.Errorf("secret keys do not match")
	}

	return first, nil
}

func promptKeyPrompt(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)

	keyBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", fmt.Errorf("read secret key: %w", err)
	}

	fmt.Fprintln(os.Stderr)

	key := strings.TrimSpace(string(keyBytes))
	if key == "" {
		return "", fmt.Errorf("secret key is empty")
	}

	return key, nil
}

// 生成 32 字节随机密钥并写入文件
func GenerateKeyFile(path string) error {
	expanded := fileutil.ExpandHome(path)

	dir := filepath.Dir(expanded)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create key directory %s: %w", dir, err)
	}

	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Errorf("generate random key: %w", err)
	}
	key := base64.StdEncoding.EncodeToString(buf)

	if err := os.WriteFile(expanded, []byte(key), 0o600); err != nil {
		return fmt.Errorf("write key file %s: %w", expanded, err)
	}

	return nil
}
