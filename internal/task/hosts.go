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

	// 应用 CLI 覆盖
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
func resolveGroup(groupName string, group *config.Group, inv *config.Inventory) []HostEntry {
	var entries []HostEntry

	names := make([]string, 0, len(group.Hosts))
	for name := range group.Hosts {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		host := group.Hosts[name]
		host.Merge(&inv.GlobalVars)
		host.Merge(&group.Vars)

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
