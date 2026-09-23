package ssh

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"

	"mestrap.com/taskssh/internal/config"
	"mestrap.com/taskssh/internal/secret"
)

// 密码缓存，避免多台主机重复输入。
var (
	cachedPassword   string
	cachedPasswordMu sync.Mutex
)

// buildAuthMethods 构造认证方法列表。
//
// 优先级：
//  1. 公钥（identity_file）
//  2. keyboard-interactive（用密码应答）
//  3. password（后备）
//  4. 终端交互（未配置任何认证时）
func buildAuthMethods(host *config.Host) ([]ssh.AuthMethod, error) {
	var methods []ssh.AuthMethod

	// 1. 公钥
	if host.IdentityFile != "" {
		signer, err := loadPrivateKey(host.IdentityFile, host.Passphrase)
		if err != nil {
			return nil, fmt.Errorf("load identity %s: %w", host.IdentityFile, err)
		}
		methods = append(methods, ssh.PublicKeys(signer))
	}

	// 2/3/4. 密码相关
	if host.Password != "" {
		// 解密一次，闭包捕获值（不是指针）
		pwd := decryptPassword(host.Password)

		// keyboard-interactive：闭包捕获密码
		methods = append(methods, ssh.KeyboardInteractive(func(
			name, instruction string, questions []string, echos []bool,
		) ([]string, error) {
			if len(questions) == 0 {
				return nil, nil
			}
			answers := make([]string, len(questions))
			for i := range answers {
				answers[i] = pwd
			}
			return answers, nil
		}))

		// password 后备
		methods = append(methods, ssh.Password(pwd))
	} else if host.IdentityFile == "" {
		// 4. 终端交互
		methods = append(methods, ssh.PasswordCallback(func() (string, error) {
			return PromptPassword(), nil
		}))
	}

	if len(methods) == 0 {
		return nil, fmt.Errorf("no auth method configured for %s", host.Host)
	}

	return methods, nil
}

// loadPrivateKey 加载私钥。
func loadPrivateKey(path, encryptedPassphrase string) (ssh.Signer, error) {
	expanded := expandHome(path)

	data, err := os.ReadFile(expanded)
	if err != nil {
		return nil, fmt.Errorf("read key: %w", err)
	}

	if encryptedPassphrase == "" {
		return ssh.ParsePrivateKey(data)
	}

	passphrase := decryptPassword(encryptedPassphrase)
	return ssh.ParsePrivateKeyWithPassphrase(data, []byte(passphrase))
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

// decryptPassword 尝试解密密码。解密失败时视为明文。
func decryptPassword(encrypted string) string {
	plain, err := secret.Decrypt(encrypted)
	if err != nil {
		return encrypted
	}
	return plain
}

// PromptPassword 从终端读取密码，结果缓存。
func PromptPassword() string {
	cachedPasswordMu.Lock()
	defer cachedPasswordMu.Unlock()

	if cachedPassword != "" {
		return cachedPassword
	}

	cachedPassword = readPasswordPrompt()
	return cachedPassword
}

// ClearPasswordCache 清空密码缓存。
func ClearPasswordCache() {
	cachedPasswordMu.Lock()
	defer cachedPasswordMu.Unlock()
	cachedPassword = ""
}
