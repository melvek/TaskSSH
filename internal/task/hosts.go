package task

import (
	"fmt"
	"sort"

	"mestrap.com/taskssh/internal/config"
	"mestrap.com/taskssh/internal/resolve"
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
// hostNames 可以是组名，也可以是独立主机名。
// ov 中的非零值会覆盖所有主机的对应字段。
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
	return entries, nil
}

// applyOverrides 应用 CLI 覆盖。
func applyOverrides(entries []HostEntry, ov Overrides) {
	for i := range entries {
		if ov.Port > 0 {
			entries[i].Host.Port = ov.Port
		}
		if ov.Username != "" {
			entries[i].Host.Username = ov.Username
		}
		if ov.Password != "" {
			entries[i].Host.Password = ov.Password
		}
	}
}

// resolveGroup 展开服务器组。
//
// 合并顺序（优先级从低到高）：
//  1. global_vars
//  2. group.vars
//  3. host（主机级，本来就有的值）
//
// Merge 的语义是"目标已有值则跳过"，所以：
//   - 先把 global 合并到 groupVars，组的值优先
//   - 再把合并后的 groupVars 合并到 host，主机的值优先
func resolveGroup(groupName string, group *config.Group, inv *config.Inventory) []HostEntry {
	var entries []HostEntry

	names := make([]string, 0, len(group.Hosts))
	for name := range group.Hosts {
		names = append(names, name)
	}
	sort.Strings(names)

	// 全局合并到组（组优先）
	groupVars := group.Vars
	groupVars.Merge(&inv.GlobalVars)

	for _, name := range names {
		host := group.Hosts[name]
		// 组合并到主机（主机优先）
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
