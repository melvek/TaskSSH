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
)

// SecretKeyFile 记录 --secret-key-file 的值，由 cmd 层设置。
var SecretKeyFile = ""

// rawKey 缓存本次会话的密钥，避免重复提示。
var rawKey string

// ErrNoKey 表示未设置密钥。
var ErrNoKey = errors.New("no encryption key available")

// SetRawKey 设置本次会话使用的密钥。
func SetRawKey(key string) {
	rawKey = key
}

// ClearRawKey 清除缓存的密钥。
func ClearRawKey() {
	rawKey = ""
}

// Encrypt 加密明文，返回 Base64 编码的密文。
func Encrypt(plain string) (string, error) {
	key, err := loadKey()
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plain), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密 Base64 编码的密文。
func Decrypt(encoded string) (string, error) {
	key, err := loadKey()
	if err != nil {
		return "", err
	}

	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("decode base64: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create GCM: %w", err)
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

// loadKey 派生 32 字节密钥。
func loadKey() ([]byte, error) {
	raw, err := loadRawKey()
	if err != nil {
		return nil, err
	}

	h := sha256.Sum256([]byte(raw))
	return h[:], nil
}

// loadRawKey 获取原始密钥字符串。
func loadRawKey() (string, error) {
	if rawKey != "" {
		return rawKey, nil
	}
	if SecretKeyFile != "" {
		return readKeyFile(SecretKeyFile)
	}
	return promptKey()
}

// readKeyFile 从文件读密钥。
func readKeyFile(path string) (string, error) {
	expanded := expandHome(path)

	info, err := os.Stat(expanded)
	if err != nil {
		return "", fmt.Errorf("stat key file %s: %w", expanded, err)
	}

	if info.Mode().Perm()&0o111 != 0 {
		return runKeyScript(expanded)
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

// expandHome 展开 ~ 为当前用户 home 目录。
func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

// promptKey 从终端读取密钥，无回显。
func promptKey() (string, error) {
	return promptKeyPrompt("Secret key: ")
}

// PromptKeyConfirm 从终端读取密钥并要求确认（用于加密）。
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

// promptKeyPrompt 读取密钥，可指定提示语。
func promptKeyPrompt(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)

	var keyBytes []byte
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
