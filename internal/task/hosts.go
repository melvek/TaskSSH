package task

import (
	"fmt"
	"net"
	"sort"
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
	Name       string      // 显示名，组内主机是 "组名/主机名"
	Host       config.Host // Host.Host 为最终连接用的 IP
	Vars       resolve.Vars
	InputHost  string // 用户原始输入或清单里写的主机标识
	ResolvedIP string // 解析后的 IP，用于显示
}

// ResolveHosts 解析目标主机列表。
//
// 查找顺序：
//  1. 组名精确匹配 → 展开组内所有主机
//  2. "组名/主机名" 格式 → 返回该主机
//  3. 组内主机名精确匹配 → 唯一时返回该主机，多组同名时报错
//  4. 独立主机名 / IP → 按独立主机处理
func ResolveHosts(hostNames []string, inv *config.Inventory, ov Overrides) ([]HostEntry, error) {
	var entries []HostEntry

	for _, name := range hostNames {
		if group, ok := inv.Servers[name]; ok {
			groupEntries := resolveGroup(name, &group, inv)
			entries = append(entries, groupEntries...)
			continue
		}

		if strings.Contains(name, "/") {
			entry, err := resolveGroupHost(name, inv)
			if err != nil {
				return nil, err
			}
			if entry != nil {
				entries = append(entries, *entry)
				continue
			}
		}

		entry, err := findHostInGroups(name, inv)
		if err != nil {
			return nil, err
		}
		if entry != nil {
			entries = append(entries, *entry)
			continue
		}

		entries = append(entries, resolveStandalone(name, inv))
	}

	applyOverrides(entries, ov)
	return entries, nil
}

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

func findHostInGroups(name string, inv *config.Inventory) (*HostEntry, error) {
	type match struct {
		groupName string
	}

	var matches []match

	for groupName, group := range inv.Servers {
		if _, ok := group.Hosts[name]; ok {
			matches = append(matches, match{groupName})
		}
	}

	if len(matches) == 0 {
		return nil, nil
	}

	if len(matches) > 1 {
		names := make([]string, 0, len(matches))
		for _, m := range matches {
			names = append(names, fmt.Sprintf("%s/%s", m.groupName, name))
		}
		sort.Strings(names)
		return nil, fmt.Errorf(
			"ambiguous host name %q, matched multiple groups: %v\n"+
				"use \"group/host\" format to specify explicitly, e.g. %s",
			name, names, names[0])
	}

	groupName := matches[0].groupName
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

		inputHost := host.Host
		resolvedIP := resolveHostIP(inputHost)

		// Host.Host 填最终连接用的 IP
		if resolvedIP != "" {
			host.Host = resolvedIP
		}

		entries = append(entries, HostEntry{
			Name:       fmt.Sprintf("%s/%s", groupName, name),
			Host:       host,
			Vars:       vars,
			InputHost:  inputHost,
			ResolvedIP: resolvedIP,
		})
	}

	return entries
}

func resolveStandalone(hostName string, inv *config.Inventory) HostEntry {
	host := config.Host{
		Host: hostName,
	}
	host.Merge(&inv.GlobalVars)

	vars := make(resolve.Vars, len(host.Extra))
	for k, v := range host.Extra {
		vars[k] = v
	}

	resolvedIP := resolveHostIP(hostName)
	if resolvedIP != "" {
		host.Host = resolvedIP
	}

	return HostEntry{
		Name:       hostName,
		Host:       host,
		Vars:       vars,
		InputHost:  hostName,
		ResolvedIP: resolvedIP,
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
