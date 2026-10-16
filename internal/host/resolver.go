package host

import (
	"fmt"
	"sort"
	"strings"

	"mestrap.com/taskssh/internal/config"
)

// resolver 是展开后的主机索引。
type resolver struct {
	inv        *config.Inventory
	entries    []HostEntry
	byName     map[string]int
	groups     map[string][]int
	groupOrder []string
	hostOrder  map[string][]string
}

func newResolver(inv *config.Inventory) *resolver {
	r := &resolver{
		inv:       inv,
		byName:    make(map[string]int),
		groups:    make(map[string][]int),
		hostOrder: make(map[string][]string),
	}

	r.groupOrder = make([]string, 0, len(inv.Servers))
	for name := range inv.Servers {
		r.groupOrder = append(r.groupOrder, name)
	}
	sort.Strings(r.groupOrder)

	for _, groupName := range r.groupOrder {
		group := inv.Servers[groupName]

		hostNames := make([]string, 0, len(group.Hosts))
		for name := range group.Hosts {
			hostNames = append(hostNames, name)
		}
		sort.Strings(hostNames)
		r.hostOrder[groupName] = hostNames

		idxs := make([]int, 0, len(hostNames))
		for _, hostName := range hostNames {
			entry := buildEntry(groupName, hostName, &group, inv)
			idx := len(r.entries)
			r.entries = append(r.entries, entry)
			r.byName[entry.Name] = idx
			idxs = append(idxs, idx)
		}
		r.groups[groupName] = idxs
	}

	return r
}

// resolve 解析 pattern 字符串，返回匹配的主机列表。
//
// 各子表达式只负责「匹配」，匹配不到返回空结果。
// 最终结果为空时报错。
func (r *resolver) resolve(pattern string, ov Overrides) ([]HostEntry, error) {
	tokens := splitPattern(pattern)

	var result []HostEntry

	if len(tokens) == 0 {
		sub, err := r.resolveSingle("g:*")
		if err != nil {
			return nil, err
		}
		result = sub
	} else {
		// 第一个 token 以 & 或 ! 开头 → 从 g:* 开始
		if strings.HasPrefix(tokens[0], "&") || strings.HasPrefix(tokens[0], "!") {
			sub, err := r.resolveSingle("g:*")
			if err != nil {
				return nil, err
			}
			result = sub
		}

		for _, t := range tokens {
			switch {
			case strings.HasPrefix(t, "&"):
				sub, err := r.resolveSingle(t[1:])
				if err != nil {
					return nil, err
				}
				result = intersectEntries(result, sub)

			case strings.HasPrefix(t, "!"):
				sub, err := r.resolveSingle(t[1:])
				if err != nil {
					return nil, err
				}
				result = subtractEntries(result, sub)

			default:
				sub, err := r.resolveSingle(t)
				if err != nil {
					return nil, err
				}
				result = unionEntries(result, sub)
			}
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no hosts matched: %q", pattern)
	}

	applyOverrides(result, ov)

	if err := decryptSecrets(result, ov.Password); err != nil {
		return nil, err
	}

	return result, nil
}

// splitPattern 按空格或逗号拆分 pattern 字符串。
func splitPattern(s string) []string {
	parts := strings.FieldsFunc(s, func(c rune) bool {
		return c == ' ' || c == ','
	})
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// resolveSingle 分派单个 pattern。
func (r *resolver) resolveSingle(name string) ([]HostEntry, error) {
	switch {
	case strings.HasPrefix(name, "g:"):
		return r.byGroup(name[2:])
	case strings.HasPrefix(name, "h:"):
		return r.byHost(name[2:])
	default:
		return r.byDefault(name)
	}
}
