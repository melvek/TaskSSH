package host

import "fmt"

// entriesAt 按索引列表拷贝出 HostEntry 切片。
func (r *resolver) entriesAt(idxs []int) []HostEntry {
	out := make([]HostEntry, len(idxs))
	for i, idx := range idxs {
		out[i] = r.entries[idx]
	}
	return out
}

// lookupGroupHost 精确查找 "组名/主机名"。
//
// 返回 nil, nil 表示格式不匹配或组不存在。
// 组存在但主机不存在时报错。
func (r *resolver) lookupGroupHost(name string) (*HostEntry, error) {
	groupName, hostName, ok := cutSlash(name)
	if !ok {
		return nil, nil
	}

	group, ok := r.inv.Servers[groupName]
	if !ok {
		return nil, nil
	}

	if _, ok := group.Hosts[hostName]; !ok {
		return nil, fmt.Errorf(
			"host %q not found in group %q", hostName, groupName)
	}

	idx, ok := r.byName[name]
	if !ok {
		return nil, fmt.Errorf("internal error: host %q not found after resolve", name)
	}
	return &r.entries[idx], nil
}

// findAllHostsByName 返回所有组内与 name 精确匹配的主机。
//
// 多组同名时全部返回，不报错。
func (r *resolver) findAllHostsByName(name string) []HostEntry {
	var all []HostEntry
	for _, groupName := range r.groupOrder {
		target := groupName + "/" + name
		if idx, ok := r.byName[target]; ok {
			all = append(all, r.entries[idx])
		}
	}
	return all
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

// cutSlash 把 "a/b" 切成 ("a", "b", true)。
//
// 空字符串、无 "/"、任一侧为空时返回 ok=false。
func cutSlash(s string) (string, string, bool) {
	i := -1
	for j := 0; j < len(s); j++ {
		if s[j] == '/' {
			i = j
			break
		}
	}
	if i <= 0 || i == len(s)-1 {
		return "", "", false
	}
	return s[:i], s[i+1:], true
}
