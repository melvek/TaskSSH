package host

import (
	"fmt"
	"path"
	"strings"
)

// byGroup 匹配组名。匹配不到返回 nil, nil。
//
// 空 pattern（"g:" 后无内容）视为语法错误。
// path.Match 出错视为语法错误。
func (r *resolver) byGroup(pattern string) ([]HostEntry, error) {
	if pattern == "" {
		return nil, fmt.Errorf("empty group pattern after 'g:'")
	}

	if idxs, ok := r.groups[pattern]; ok {
		return r.entriesAt(idxs), nil
	}

	if isGlobPattern(pattern) {
		return r.matchGroups(pattern)
	}

	return nil, nil
}

// byHost 匹配所有组内主机名。匹配不到返回 nil, nil。
func (r *resolver) byHost(pattern string) ([]HostEntry, error) {
	if pattern == "" {
		return nil, fmt.Errorf("empty host pattern after 'h:'")
	}

	if !isGlobPattern(pattern) {
		return r.findAllHostsByName(pattern), nil
	}

	return r.matchHosts(pattern)
}

// byDefault 无前缀的默认匹配。
//
// 匹配顺序：
//  1. 精确匹配组名
//  2. 含 "/"：组/主机组合匹配（两部分各自支持精确与 glob）
//  3. 精确匹配组内主机名（跨组返回所有同名主机）
//  4. glob 匹配组名或主机名，去重合并
//  5. 都不匹配 → 视为独立主机
func (r *resolver) byDefault(name string) ([]HostEntry, error) {
	// 1. 精确匹配组名
	if idxs, ok := r.groups[name]; ok {
		return r.entriesAt(idxs), nil
	}

	// 2. 含 "/"：组/主机组合匹配
	if strings.Contains(name, "/") {
		groupPat, hostPat, _ := strings.Cut(name, "/")
		entries, err := r.matchGroupHost(groupPat, hostPat)
		if err != nil {
			return nil, err
		}
		if len(entries) > 0 {
			return entries, nil
		}
	}

	// 3. 精确匹配组内主机名
	if all := r.findAllHostsByName(name); len(all) > 0 {
		return all, nil
	}

	// 4. glob 匹配组名 + 主机名
	if isGlobPattern(name) {
		groupHits, err := r.matchGroups(name)
		if err != nil {
			return nil, err
		}
		hostHits, err := r.matchHosts(name)
		if err != nil {
			return nil, err
		}
		all := append(groupHits, hostHits...)
		if len(all) > 0 {
			return dedupeEntries(all), nil
		}
	}

	// 5. 独立主机
	return []HostEntry{standaloneEntry(name, r.inv)}, nil
}

// matchGroups 返回所有组名匹配 pattern 的条目。
func (r *resolver) matchGroups(pattern string) ([]HostEntry, error) {
	var all []HostEntry
	for _, groupName := range r.groupOrder {
		ok, err := path.Match(pattern, groupName)
		if err != nil {
			return nil, fmt.Errorf("invalid group pattern %q: %w", pattern, err)
		}
		if ok {
			all = append(all, r.entriesAt(r.groups[groupName])...)
		}
	}
	return all, nil
}

// matchHosts 返回所有主机名匹配 pattern 的条目。
func (r *resolver) matchHosts(pattern string) ([]HostEntry, error) {
	var all []HostEntry
	for _, groupName := range r.groupOrder {
		for _, hostName := range r.hostOrder[groupName] {
			ok, err := path.Match(pattern, hostName)
			if err != nil {
				return nil, fmt.Errorf("invalid host pattern %q: %w", pattern, err)
			}
			if ok {
				if idx, ok := r.byName[groupName+"/"+hostName]; ok {
					all = append(all, r.entries[idx])
				}
			}
		}
	}
	return all, nil
}

// matchGroupHost 匹配 "组 pattern/主机 pattern"。
//
// 两部分各自支持精确和 glob：
//
//	prod/web1       精确组名 + 精确主机名
//	prod/web*       精确组名 + 主机 glob
//	prod*/web1      组 glob + 精确主机名
//	prod*/web*      组 glob + 主机 glob
//
// 匹配不到返回 nil, nil；pattern 语法错返回 error。
func (r *resolver) matchGroupHost(groupPat, hostPat string) ([]HostEntry, error) {
	var all []HostEntry
	for _, groupName := range r.groupOrder {
		ok, err := path.Match(groupPat, groupName)
		if err != nil {
			return nil, fmt.Errorf("invalid group pattern %q: %w", groupPat, err)
		}
		if !ok {
			continue
		}
		for _, hostName := range r.hostOrder[groupName] {
			ok, err := path.Match(hostPat, hostName)
			if err != nil {
				return nil, fmt.Errorf("invalid host pattern %q: %w", hostPat, err)
			}
			if ok {
				if idx, ok := r.byName[groupName+"/"+hostName]; ok {
					all = append(all, r.entries[idx])
				}
			}
		}
	}
	return all, nil
}

// isGlobPattern 判断字符串是否含 glob 元字符。
func isGlobPattern(s string) bool {
	return strings.ContainsAny(s, "*?[")
}
