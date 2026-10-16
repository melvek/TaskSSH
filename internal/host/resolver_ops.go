package host

// 集合操作。所有操作基于 HostEntry.Name 匹配。
//
// 所有返回的切片均为全新分配、不与入参共享底层数组，调用方可安全修改返回值。

// unionEntries 返回 a 和 b 的并集，保留 a 的顺序，去重。
func unionEntries(a, b []HostEntry) []HostEntry {
	seen := make(map[string]struct{}, len(a)+len(b))
	out := make([]HostEntry, 0, len(a)+len(b))

	for _, e := range a {
		if _, ok := seen[e.Name]; !ok {
			seen[e.Name] = struct{}{}
			out = append(out, e)
		}
	}
	for _, e := range b {
		if _, ok := seen[e.Name]; !ok {
			seen[e.Name] = struct{}{}
			out = append(out, e)
		}
	}
	return out
}

// intersectEntries 返回 a 和 b 的交集，保留 a 的顺序。
func intersectEntries(a, b []HostEntry) []HostEntry {
	if len(a) == 0 || len(b) == 0 {
		return nil
	}

	bSet := make(map[string]struct{}, len(b))
	for _, e := range b {
		bSet[e.Name] = struct{}{}
	}

	out := make([]HostEntry, 0, len(a))
	for _, e := range a {
		if _, ok := bSet[e.Name]; ok {
			out = append(out, e)
		}
	}
	return out
}

// subtractEntries 返回 a 中不在 b 里的条目，保留 a 的顺序。
//
// 即使 len(b)==0 也会拷贝一份，避免返回值与入参共享底层数组导致调用方
// 修改返回值时意外污染上游切片。
func subtractEntries(a, b []HostEntry) []HostEntry {
	if len(a) == 0 {
		return nil
	}

	out := make([]HostEntry, 0, len(a))

	if len(b) == 0 {
		// 没有排除项 → 全部拷贝
		out = append(out, a...)
		return out
	}

	bSet := make(map[string]struct{}, len(b))
	for _, e := range b {
		bSet[e.Name] = struct{}{}
	}

	for _, e := range a {
		if _, ok := bSet[e.Name]; !ok {
			out = append(out, e)
		}
	}
	return out
}
