package task

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"mestrap.com/taskssh/internal/config"
)

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
