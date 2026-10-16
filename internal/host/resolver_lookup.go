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

// 保留 fmt 引用，避免未使用报错（若其他地方不再使用可删）
var _ = fmt.Sprintf
