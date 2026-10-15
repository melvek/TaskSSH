package host

import (
	"fmt"
	"path"
	"strings"
)

// byGroup 只匹配组名。
func (r *resolver) byGroup(pattern string) ([]HostEntry, error) {
	if pattern == "" {
		return nil, fmt.Errorf("empty group pattern after 'g:'")
	}

	// 精确匹配
	if idxs, ok := r.groups[pattern]; ok {
		return r.entriesAt(idxs), nil
	}

	// glob 匹配
	if isGlobPattern(pattern) {
		var all []HostEntry
		for _, groupName := range r.groupOrder {
			if ok, _ := path.Match(pattern, groupName); ok {
				all = append(all, r.entriesAt(r.groups[groupName])...)
			}
		}
		if len(all) > 0 {
			return all, nil
		}
	}

	return nil, fmt.Errorf("no group matched: %s", pattern)
}

// byHost 只匹配组内主机名。
//
// 跨所有组匹配，多个组有同名主机时返回全部，不报歧义。
func (r *resolver) byHost(pattern string) ([]HostEntry, error) {
	if pattern == "" {
		return nil, fmt.Errorf("empty host pattern after 'h:'")
	}

	// 精确匹配：返回所有组的同名主机
	if !isGlobPattern(pattern) {
		all := r.findAllHostsByName(pattern)
		if len(all) > 0 {
			return all, nil
		}
		return nil, fmt.Errorf("no host matched: %s", pattern)
	}

	// glob 匹配：遍历所有组的主机名
	var all []HostEntry
	for _, groupName := range r.groupOrder {
		for _, hostName := range r.hostOrder[groupName] {
			if ok, _ := path.Match(pattern, hostName); ok {
				target := groupName + "/" + hostName
				if idx, ok := r.byName[target]; ok {
					all = append(all, r.entries[idx])
				}
			}
		}
	}
	if len(all) > 0 {
		return all, nil
	}

	return nil, fmt.Errorf("no host matched: %s", pattern)
}

// byDefault 默认行为：同时匹配组名和主机名。
//
// 匹配顺序：
//  1. 精确匹配组名
//  2. 精确匹配 "组名/主机名"
//  3. 精确匹配组内主机名（跨组返回所有同名主机）
//  4. glob 匹配组名 + 主机名，去重合并
//  5. 以上都不匹配 → 视为独立主机
func (r *resolver) byDefault(name string) ([]HostEntry, error) {
	// 1. 精确匹配组名
	if idxs, ok := r.groups[name]; ok {
		return r.entriesAt(idxs), nil
	}

	// 2. 精确匹配 "组名/主机名"
	if strings.Contains(name, "/") {
		entry, err := r.lookupGroupHost(name)
		if err != nil {
			return nil, err
		}
		if entry != nil {
			return []HostEntry{*entry}, nil
		}
	}

	// 3. 精确匹配组内主机名（跨组返回所有）
	if all := r.findAllHostsByName(name); len(all) > 0 {
		return all, nil
	}

	// 4. glob 匹配：组名 + 主机名，累积去重
	if isGlobPattern(name) {
		var all []HostEntry

		for _, groupName := range r.groupOrder {
			if ok, _ := path.Match(name, groupName); ok {
				all = append(all, r.entriesAt(r.groups[groupName])...)
			}
		}

		for _, groupName := range r.groupOrder {
			for _, hostName := range r.hostOrder[groupName] {
				if ok, _ := path.Match(name, hostName); ok {
					target := groupName + "/" + hostName
					if idx, ok := r.byName[target]; ok {
						all = append(all, r.entries[idx])
					}
				}
			}
		}

		if len(all) > 0 {
			return dedupeEntries(all), nil
		}
	}

	// 5. 独立主机
	return []HostEntry{standaloneEntry(name, r.inv)}, nil
}

// isGlobPattern 判断字符串是否含 glob 元字符。
func isGlobPattern(s string) bool {
	return strings.ContainsAny(s, "*?[")
}
