package task

import (
	"fmt"
	"sort"

	"mestrap.com/taskssh/internal/config"
	"mestrap.com/taskssh/internal/resolve"
	"mestrap.com/taskssh/internal/secret"
)

// Overrides 是 CLI 覆盖参数。
type Overrides struct {
	Port     int
	Username string
	Password string
}

// HostEntry 是目标主机的一项。
type HostEntry struct {
	Name string
	Host config.Host
	Vars resolve.Vars
}

// ResolveHosts 解析目标主机列表。
//
// 合并顺序：global -> group -> host，host 优先。
// 合并后解密密码和 passphrase，使 Host 中的敏感字段为明文。
func ResolveHosts(hostNames []string, inv *config.Inventory, ov Overrides) ([]HostEntry, error) {
	var entries []HostEntry

	for _, name := range hostNames {
		if group, ok := inv.Servers[name]; ok {
			groupEntries := resolveGroup(name, &group, inv)
			entries = append(entries, groupEntries...)
			continue
		}

		entry := resolveStandalone(name, inv)
		entries = append(entries, entry)
	}

	applyOverrides(entries, ov)

	if err := decryptSecrets(entries, ov.Password); err != nil {
		return nil, err
	}

	return entries, nil
}

// applyOverrides 应用 CLI 覆盖（不含密码）。
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
//
// 规则：
//   - CLI 密码优先：非空时直接写入，视为明文，不解密
//   - 清单密码：必须能解密，失败即报错
//   - passphrase：只能来自清单，必须能解密
func decryptSecrets(entries []HostEntry, cliPassword string) error {
	for i := range entries {
		// 密码
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

		// passphrase
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

// resolveGroup 展开服务器组。
func resolveGroup(groupName string, group *config.Group, inv *config.Inventory) []HostEntry {
	var entries []HostEntry

	names := make([]string, 0, len(group.Hosts))
	for name := range group.Hosts {
		names = append(names, name)
	}
	sort.Strings(names)

	groupVars := group.Vars
	groupVars.Merge(&inv.GlobalVars)

	for _, name := range names {
		host := group.Hosts[name]
		host.Merge(&groupVars)

		vars := make(resolve.Vars, len(host.Extra))
		for k, v := range host.Extra {
			vars[k] = v
		}

		entries = append(entries, HostEntry{
			Name: fmt.Sprintf("%s/%s", groupName, name),
			Host: host,
			Vars: vars,
		})
	}

	return entries
}

// resolveStandalone 解析独立主机。
func resolveStandalone(hostName string, inv *config.Inventory) HostEntry {
	host := config.Host{
		Host: hostName,
	}
	host.Merge(&inv.GlobalVars)

	vars := make(resolve.Vars, len(host.Extra))
	for k, v := range host.Extra {
		vars[k] = v
	}

	return HostEntry{
		Name: hostName,
		Host: host,
		Vars: vars,
	}
}
