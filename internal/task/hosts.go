package task

import (
	"fmt"
	"net"
	"path"
	"sort"
	"strings"

	"mestrap.com/taskssh/internal/config"
	"mestrap.com/taskssh/internal/resolve"
	"mestrap.com/taskssh/internal/secret"
)

// Overrides 是 CLI 覆盖参数。
type Overrides struct {
	Port         int
	Username     string
	Password     string
	IdentityFile string
}

// HostEntry 是目标主机的一项。
type HostEntry struct {
	Name string      // 显示名：组内主机为 "组名/主机名"，独立主机为输入值
	Host config.Host // Host.Host 为最终连接用的 IP
	Vars resolve.Vars
}

// ResolveHosts 解析目标主机列表。
//
// 输入分类：
//   - 以 "!" 开头：exclude，从结果中排除
//   - 其他：include，加入结果
//
// 规则：
//   - 无 include 时，默认 include 为 g:*（所有组）
//   - exclude 的任何错误都跳过，不中断
//   - 按 Name 去重和过滤
//
// 输入支持以下形式：
//
//	prod                    精确匹配组名
//	prod/web1               精确匹配 "组名/主机名"
//	web1                    精确匹配组内主机名（多组同名时报错）
//	192.168.1.10            独立主机或 IP
//	g:web*                  glob 匹配组名
//	h:web*                  glob 匹配组内主机名
//	web*                    默认：同时匹配组名和主机名，去重合并
//	!g:web*                 排除：组名匹配 web* 的组
//	!h:web1                 排除：主机名匹配 web1 的主机
func ResolveHosts(hostNames []string, inv *config.Inventory, ov Overrides) ([]HostEntry, error) {
	var includes []string
	var excludes []string

	for _, name := range hostNames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if strings.HasPrefix(name, "!") {
			excludes = append(excludes, name[1:])
		} else {
			includes = append(includes, name)
		}
	}

	// 无 include：默认 g:*
	if len(includes) == 0 {
		includes = []string{"g:*"}
	}

	var entries []HostEntry

	// 1. 处理 include
	for _, name := range includes {
		sub, err := resolveSingleHost(name, inv)
		if err != nil {
			return nil, err
		}
		entries = append(entries, sub...)
	}

	entries = dedupeEntries(entries)

	// 2. 处理 exclude
	if len(excludes) > 0 {
		excludeSet := make(map[string]bool)

		for _, name := range excludes {
			sub, err := resolveSingleHost(name, inv)
			if err != nil {
				// exclude 的任何错误都跳过
				continue
			}
			for _, e := range sub {
				excludeSet[e.Name] = true
			}
		}

		filtered := make([]HostEntry, 0, len(entries))
		for _, e := range entries {
			if !excludeSet[e.Name] {
				filtered = append(filtered, e)
			}
		}
		entries = filtered
	}

	applyOverrides(entries, ov)

	if err := decryptSecrets(entries, ov.Password); err != nil {
		return nil, err
	}

	return entries, nil
}

// resolveSingleHost 解析单个主机名。
func resolveSingleHost(name string, inv *config.Inventory) ([]HostEntry, error) {
	if strings.HasPrefix(name, "g:") {
		return resolveByGroupPattern(name[2:], inv)
	}
	if strings.HasPrefix(name, "h:") {
		return resolveByHostPattern(name[2:], inv)
	}

	return resolveDefault(name, inv)
}

// resolveByGroupPattern 只匹配组名。
func resolveByGroupPattern(pattern string, inv *config.Inventory) ([]HostEntry, error) {
	if pattern == "" {
		return nil, fmt.Errorf("empty group pattern after 'g:'")
	}

	// 精确匹配
	if group, ok := inv.Servers[pattern]; ok {
		return resolveGroup(pattern, &group, inv), nil
	}

	// glob 匹配
	if isGlobPattern(pattern) {
		if groups := matchGroups(pattern, inv); len(groups) > 0 {
			var entries []HostEntry
			for _, g := range groups {
				group := inv.Servers[g]
				entries = append(entries, resolveGroup(g, &group, inv)...)
			}
			return entries, nil
		}
	}

	return nil, fmt.Errorf("no group matched: %s", pattern)
}

// resolveByHostPattern 只匹配组内主机名。
func resolveByHostPattern(pattern string, inv *config.Inventory) ([]HostEntry, error) {
	if pattern == "" {
		return nil, fmt.Errorf("empty host pattern after 'h:'")
	}

	// 精确匹配组内主机
	entry, err := findHostInGroups(pattern, inv)
	if err != nil {
		return nil, err
	}
	if entry != nil {
		return []HostEntry{*entry}, nil
	}

	// glob 匹配
	if isGlobPattern(pattern) {
		if entries := matchHostsInGroups(pattern, inv); len(entries) > 0 {
			return entries, nil
		}
	}

	return nil, fmt.Errorf("no host matched: %s", pattern)
}

// resolveDefault 默认行为：同时匹配组名和主机名。
func resolveDefault(name string, inv *config.Inventory) ([]HostEntry, error) {
	// 1. 精确匹配组名
	if group, ok := inv.Servers[name]; ok {
		return resolveGroup(name, &group, inv), nil
	}

	// 2. 精确匹配 "组名/主机名"
	if strings.Contains(name, "/") {
		entry, err := resolveGroupHost(name, inv)
		if err != nil {
			return nil, err
		}
		if entry != nil {
			return []HostEntry{*entry}, nil
		}
	}

	// 3. 精确匹配组内主机名
	entry, err := findHostInGroups(name, inv)
	if err != nil {
		return nil, err
	}
	if entry != nil {
		return []HostEntry{*entry}, nil
	}

	// 4. glob 匹配：组名 + 主机名，累积去重
	if isGlobPattern(name) {
		var entries []HostEntry

		for _, g := range matchGroups(name, inv) {
			group := inv.Servers[g]
			entries = append(entries, resolveGroup(g, &group, inv)...)
		}

		entries = append(entries, matchHostsInGroups(name, inv)...)

		if len(entries) > 0 {
			return dedupeEntries(entries), nil
		}
	}

	// 5. 独立主机
	return []HostEntry{resolveStandalone(name, inv)}, nil
}

// isGlobPattern 判断字符串是否含 glob 元字符。
func isGlobPattern(s string) bool {
	return strings.ContainsAny(s, "*?[")
}

// matchGroups 用 glob 匹配组名。
func matchGroups(pattern string, inv *config.Inventory) []string {
	var matched []string
	for name := range inv.Servers {
		if ok, _ := path.Match(pattern, name); ok {
			matched = append(matched, name)
		}
	}
	sort.Strings(matched)
	return matched
}

// matchHostsInGroups 用 glob 匹配所有组内主机名。
func matchHostsInGroups(pattern string, inv *config.Inventory) []HostEntry {
	var entries []HostEntry

	groupNames := make([]string, 0, len(inv.Servers))
	for name := range inv.Servers {
		groupNames = append(groupNames, name)
	}
	sort.Strings(groupNames)

	for _, groupName := range groupNames {
		group := inv.Servers[groupName]

		hostNames := make([]string, 0, len(group.Hosts))
		for name := range group.Hosts {
			hostNames = append(hostNames, name)
		}
		sort.Strings(hostNames)

		for _, hostName := range hostNames {
			if ok, _ := path.Match(pattern, hostName); ok {
				entry, err := resolveGroupHost(
					fmt.Sprintf("%s/%s", groupName, hostName), inv)
				if err == nil && entry != nil {
					entries = append(entries, *entry)
				}
			}
		}
	}

	return entries
}

// dedupeEntries 按 Name 去重。
func dedupeEntries(entries []HostEntry) []HostEntry {
	seen := make(map[string]bool, len(entries))
	result := make([]HostEntry, 0, len(entries))

	for _, e := range entries {
		if seen[e.Name] {
			continue
		}
		seen[e.Name] = true
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
				"use \"group/host\" format to specify explicitly, e.g. %s\n"+
				"or use \"h:%s\" to match all hosts with this name",
			name, names, names[0], name)
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
