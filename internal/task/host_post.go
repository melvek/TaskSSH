package task

import (
	"fmt"
	"strings"

	"mestrap.com/taskssh/internal/resolve"
	"mestrap.com/taskssh/internal/secret"
)

// applyOverrides 应用 CLI 覆盖（端口、用户名）。
func applyOverrides(entries []HostEntry, ov Overrides) {
	for i := range entries {
		if ov.Port > 0 {
			entries[i].Host.Port = ov.Port
		}
		if ov.Username != "" {
			entries[i].Host.Username = ov.Username
		}
	}
}

// decryptSecrets 解密主机密码和 passphrase。
func decryptSecrets(entries []HostEntry, cliPassword string) error {
	for i := range entries {
		if cliPassword != "" {
			entries[i].Host.Password = cliPassword
		} else if entries[i].Host.Password != "" {
			plain, err := secret.Decrypt(entries[i].Host.Password)
			if err != nil {
				return fmt.Errorf(
					"host %s: decrypt password: %w (ensure the value is encrypted by 'taskssh encrypt')",
					entries[i].Name, err)
			}
			entries[i].Host.Password = plain
		}

		if entries[i].Host.Passphrase != "" {
			plain, err := secret.Decrypt(entries[i].Host.Passphrase)
			if err != nil {
				return fmt.Errorf(
					"host %s: decrypt passphrase: %w (ensure the value is encrypted by 'taskssh encrypt')",
					entries[i].Name, err)
			}
			entries[i].Host.Passphrase = plain
		}
	}
	return nil
}

// StandardVars 返回该主机的标准变量。
//
// 每次调用时生成，反映当前主机的清单信息。
// 这些变量在任务执行时注入，值因主机而异。
func (e *HostEntry) StandardVars() resolve.Vars {
	vars := make(resolve.Vars, 5)

	// host_name / group_name
	// Name 格式：组内主机为 "组名/主机名"，独立主机为 "主机名"
	if idx := strings.Index(e.Name, "/"); idx > 0 {
		vars["group_name"] = e.Name[:idx]
		vars["host_name"] = e.Name[idx+1:]
	} else {
		vars["group_name"] = ""
		vars["host_name"] = e.Name
	}

	// 连接信息
	vars["host_addr"] = e.Host.Host
	vars["host_port"] = e.Host.Port
	vars["host_user"] = e.Host.Username

	return vars
}
