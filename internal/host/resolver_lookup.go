package host

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
// 多组同名时全部返回，不报错。利用 hostsByName 索引实现 O(1) 查询，
// 返回顺序与 groupOrder 一致（索引构建时按 groupOrder 追加）。
func (r *resolver) findAllHostsByName(name string) []HostEntry {
	idxs, ok := r.hostsByName[name]
	if !ok {
		return nil
	}
	return r.entriesAt(idxs)
}

// dedupeEntries 按 Name 去重，保留首次出现的条目。
func dedupeEntries(entries []HostEntry) []HostEntry {
	seen := make(map[string]struct{}, len(entries))
	result := make([]HostEntry, 0, len(entries))
	for _, e := range entries {
		if _, ok := seen[e.Name]; ok {
			continue
		}
		seen[e.Name] = struct{}{}
		result = append(result, e)
	}
	return result
}
