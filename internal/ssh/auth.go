package ssh

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"
	"mestrap.com/taskssh/internal/config"
	"mestrap.com/taskssh/internal/prompt"
)

// 密码缓存，避免多台主机重复输入。
var (
	cachedPassword   string
	cachedPasswordMu sync.Mutex
)

// 构造认证方法列表。
func buildAuthMethods(host *config.Host) ([]ssh.AuthMethod, error) {
	var methods []ssh.AuthMethod

	// SSH Agent
	if !host.AuthOverride {
		if ag := getAgent(); ag != nil {
			methods = append(methods, ssh.PublicKeysCallback(ag.Signers))
		}
	}

	// 公钥
	if host.IdentityFile != "" {
		signer, err := loadPrivateKey(host.IdentityFile, host.Passphrase)
		if err != nil {
			// agent 可用时，公钥加载失败不中断
			if host.AuthOverride || len(methods) == 0 {
				return nil, fmt.Errorf("load identity %s: %w", host.IdentityFile, err)
			}
		} else {
			methods = append(methods, ssh.PublicKeys(signer))
		}
	}

	// 密码相关
	if host.Password != "" {
		pwd := host.Password

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

		methods = append(methods, ssh.Password(pwd))
	} else if host.IdentityFile == "" && len(methods) == 0 && !host.AuthOverride {
		methods = append(methods, ssh.PasswordCallback(func() (string, error) {
			return PromptPassword(), nil
		}))
	}

	if len(methods) == 0 {
		return nil, fmt.Errorf("no auth method configured for %s", host.Host)
	}

	return methods, nil
}

// 加载私钥
func loadPrivateKey(path, passphrase string) (ssh.Signer, error) {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(home, path[2:])
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read key: %w", err)
	}

	if passphrase == "" {
		return ssh.ParsePrivateKey(data)
	}

	return ssh.ParsePrivateKeyWithPassphrase(data, []byte(passphrase))
}

// PromptPassword 从终端读取密码，结果缓存。
func PromptPassword() string {
	cachedPasswordMu.Lock()
	defer cachedPasswordMu.Unlock()

	if cachedPassword != "" {
		return cachedPassword
	}

	cachedPassword, err := prompt.Hidden("Password: ")
	if err != nil {
		return ""
	}

	return cachedPassword
}

// ClearPasswordCache 清空密码缓存。
func ClearPasswordCache() {
	cachedPasswordMu.Lock()
	defer cachedPasswordMu.Unlock()
	cachedPassword = ""
}
