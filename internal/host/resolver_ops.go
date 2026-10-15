package host

// 集合操作。所有操作基于 HostEntry.Name 匹配。

// unionEntries 返回 a 和 b 的并集，保留 a 的顺序，去重。
func unionEntries(a, b []HostEntry) []HostEntry {
	seen := make(map[string]bool, len(a)+len(b))
	out := make([]HostEntry, 0, len(a)+len(b))

	for _, e := range a {
		if !seen[e.Name] {
			seen[e.Name] = true
			out = append(out, e)
		}
	}
	for _, e := range b {
		if !seen[e.Name] {
			seen[e.Name] = true
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

	bSet := make(map[string]bool, len(b))
	for _, e := range b {
		bSet[e.Name] = true
	}

	out := make([]HostEntry, 0, len(a))
	for _, e := range a {
		if bSet[e.Name] {
			out = append(out, e)
		}
	}
	return out
}

// subtractEntries 返回 a 中不在 b 里的条目，保留 a 的顺序。
func subtractEntries(a, b []HostEntry) []HostEntry {
	if len(a) == 0 {
		return nil
	}
	if len(b) == 0 {
		return a
	}

	bSet := make(map[string]bool, len(b))
	for _, e := range b {
		bSet[e.Name] = true
	}

	out := make([]HostEntry, 0, len(a))
	for _, e := range a {
		if !bSet[e.Name] {
			out = append(out, e)
		}
	}
	return out
}
