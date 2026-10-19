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
		return "", err
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

	if fileutil.FileExists(DefaultKeyFile) {
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

// GenerateKeyFile writes a newly generated 32-byte key with owner-only permissions.
func GenerateKeyFile(path string, force bool) (retErr error) {
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

	if !force {
		file, err := os.OpenFile(expanded, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if err != nil {
			if os.IsExist(err) {
				return fmt.Errorf("vault key already exists: %s; use -F/--force to overwrite", expanded)
			}
			return fmt.Errorf("create key file %s: %w", expanded, err)
		}
		keepFile := false
		defer func() {
			if !keepFile {
				if err := os.Remove(expanded); err != nil && !os.IsNotExist(err) {
					retErr = errors.Join(retErr, fmt.Errorf("remove incomplete key file %s: %w", expanded, err))
				}
			}
		}()
		if err := writeKeyContents(file, expanded, key); err != nil {
			return err
		}
		keepFile = true
		return nil
	}

	if err := rejectKeyTarget(expanded); err != nil {
		return err
	}

	temp, err := os.CreateTemp(dir, ".taskssh-vault-key-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary key file in %s: %w", dir, err)
	}
	tempPath := temp.Name()
	defer func() {
		if err := os.Remove(tempPath); err != nil && !os.IsNotExist(err) {
			retErr = errors.Join(retErr, fmt.Errorf("remove temporary key file %s: %w", tempPath, err))
		}
	}()
	if err := temp.Chmod(0o600); err != nil {
		closeErr := temp.Close()
		return errors.Join(
			fmt.Errorf("set key file permissions %s: %w", tempPath, err),
			wrapFileError("close key file", tempPath, closeErr),
		)
	}
	if err := writeKeyContents(temp, tempPath, key); err != nil {
		return err
	}
	if err := rejectKeyTarget(expanded); err != nil {
		return err
	}
	if err := os.Rename(tempPath, expanded); err != nil {
		return fmt.Errorf("replace key file %s: %w", expanded, err)
	}

	return nil
}

func rejectKeyTarget(path string) error {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect key file %s: %w", path, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refusing to write through symbolic link %s", path)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("key file target is not a regular file: %s", path)
	}
	return nil
}

func writeKeyContents(file *os.File, path, key string) error {
	data := []byte(key)
	for len(data) > 0 {
		n, err := file.Write(data)
		if err != nil {
			closeErr := file.Close()
			return errors.Join(fmt.Errorf("write key file %s: %w", path, err),
				wrapFileError("close key file", path, closeErr))
		}
		if n == 0 {
			closeErr := file.Close()
			return errors.Join(fmt.Errorf("write key file %s: %w", path, io.ErrShortWrite),
				wrapFileError("close key file", path, closeErr))
		}
		data = data[n:]
	}
	if err := file.Sync(); err != nil {
		closeErr := file.Close()
		return errors.Join(fmt.Errorf("sync key file %s: %w", path, err),
			wrapFileError("close key file", path, closeErr))
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close key file %s: %w", path, err)
	}
	return nil
}

func wrapFileError(operation, path string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s %s: %w", operation, path, err)
}
