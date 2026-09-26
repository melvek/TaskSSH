package task

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"

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
	Name string      // 显示名：组内主机为 "组名/主机名"，独立主机为输入值
	Host config.Host // Host.Host 为最终连接用的 IP
	Vars resolve.Vars
}

// ResolveHosts 解析目标主机列表。
//
// 步骤：
//  1. 对每个输入做主机模式展开
//  2. 对展开后的每个主机名做查找：
//     组名 → 组名/主机名 → 组内主机名 → 独立主机
//  3. 去重
func ResolveHosts(hostNames []string, inv *config.Inventory, ov Overrides) ([]HostEntry, error) {
	var entries []HostEntry

	for _, name := range hostNames {
		expanded, err := ExpandHostPattern(name)
		if err != nil {
			return nil, err
		}

		for _, item := range expanded {
			sub, err := resolveSingleHost(item, inv)
			if err != nil {
				return nil, err
			}
			entries = append(entries, sub...)
		}
	}

	entries = dedupeEntries(entries)

	applyOverrides(entries, ov)
	return entries, nil
}

// resolveSingleHost 解析单个主机名（不含模式）。
func resolveSingleHost(name string, inv *config.Inventory) ([]HostEntry, error) {
	// 1. 组名
	if group, ok := inv.Servers[name]; ok {
		return resolveGroup(name, &group, inv), nil
	}

	// 2. "组名/主机名"
	if strings.Contains(name, "/") {
		entry, err := resolveGroupHost(name, inv)
		if err != nil {
			return nil, err
		}
		if entry != nil {
			return []HostEntry{*entry}, nil
		}
	}

	// 3. 组内主机名
	entry, err := findHostInGroups(name, inv)
	if err != nil {
		return nil, err
	}
	if entry != nil {
		return []HostEntry{*entry}, nil
	}

	// 4. 独立主机
	return []HostEntry{resolveStandalone(name, inv)}, nil
}

// dedupeEntries 按 host:port 去重。
func dedupeEntries(entries []HostEntry) []HostEntry {
	seen := make(map[string]bool, len(entries))
	result := make([]HostEntry, 0, len(entries))

	for _, e := range entries {
		key := e.Host.Host + ":" + strconv.Itoa(e.Host.Port)
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, e)
	}
	return result
}

// resolveGroupHost 解析 "组名/主机名" 格式。
func resolveGroupHost(name string, inv *config.Inventory) (*HostEntry, error) {
	parts := strings.SplitN(name, "/", 2)
	if len(parts) != 2 {
		return nil, nil
	}

	groupName := parts[0]
	hostName := parts[1]

	if groupName == "" || hostName == "" {
		return nil, nil
	}

	group, ok := inv.Servers[groupName]
	if !ok {
		return nil, nil
	}

	if _, ok := group.Hosts[hostName]; !ok {
		return nil, fmt.Errorf(
			"host %q not found in group %q", hostName, groupName)
	}

	groupEntries := resolveGroup(groupName, &group, inv)
	target := fmt.Sprintf("%s/%s", groupName, hostName)

	for i := range groupEntries {
		if groupEntries[i].Name == target {
			return &groupEntries[i], nil
		}
	}

	return nil, fmt.Errorf("internal error: host %q not found after resolve", target)
}

// findHostInGroups 在所有组内查找同名主机。
func findHostInGroups(name string, inv *config.Inventory) (*HostEntry, error) {
	var matchedGroups []string

	for groupName, group := range inv.Servers {
		if _, ok := group.Hosts[name]; ok {
			matchedGroups = append(matchedGroups, groupName)
		}
	}

	if len(matchedGroups) == 0 {
		return nil, nil
	}

	if len(matchedGroups) > 1 {
		sort.Strings(matchedGroups)
		names := make([]string, 0, len(matchedGroups))
		for _, g := range matchedGroups {
			names = append(names, fmt.Sprintf("%s/%s", g, name))
		}
		return nil, fmt.Errorf(
			"ambiguous host name %q, matched multiple groups: %v\n"+
				"use \"group/host\" format to specify explicitly, e.g. %s",
			name, names, names[0])
	}

	groupName := matchedGroups[0]
	group := inv.Servers[groupName]
	groupEntries := resolveGroup(groupName, &group, inv)

	target := fmt.Sprintf("%s/%s", groupName, name)

	for i := range groupEntries {
		if groupEntries[i].Name == target {
			return &groupEntries[i], nil
		}
	}

	return nil, nil
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

	groupVars := group.Vars
	groupVars.Merge(&inv.GlobalVars)

	for _, name := range names {
		host := group.Hosts[name]
		host.Merge(&groupVars)

		vars := make(resolve.Vars, len(host.Extra))
		for k, v := range host.Extra {
			vars[k] = v
		}

		if ip := resolveHostIP(host.Host); ip != "" {
			host.Host = ip
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

	if ip := resolveHostIP(hostName); ip != "" {
		host.Host = ip
	}

	return HostEntry{
		Name: hostName,
		Host: host,
		Vars: vars,
	}
}

// resolveHostIP 把主机名解析为 IP。
//
//   - 已经是 IP：返回自身
//   - 是主机名：解析为第一个 IP
//   - 解析失败：返回空字符串
func resolveHostIP(host string) string {
	if host == "" {
		return ""
	}

	if net.ParseIP(host) != nil {
		return host
	}

	ips, err := net.LookupHost(host)
	if err != nil || len(ips) == 0 {
		return ""
	}

	return ips[0]
}
