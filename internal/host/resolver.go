package host

import (
	"fmt"
	"sort"
	"strings"

	"mestrap.com/taskssh/internal/config"
)

// resolver 是展开后的主机索引。
type resolver struct {
	inv         *config.Inventory
	entries     []HostEntry
	byName      map[string]int      // "组名/主机名" → entries 下标
	groups      map[string][]int    // 组名 → entries 下标列表
	groupOrder  []string            // 已排序的组名
	hostOrder   map[string][]string // 组名 → 该组已排序主机名
	hostsByName map[string][]int    // 主机名 → entries 下标列表（跨组，O(1) 精确匹配）
	hostOf      []string            // 与 entries 一一对应：该条目的主机名（不含组名前缀）
}

func newResolver(inv *config.Inventory) *resolver {
	r := &resolver{
		inv:         inv,
		byName:      make(map[string]int),
		groups:      make(map[string][]int),
		hostOrder:   make(map[string][]string),
		hostsByName: make(map[string][]int),
	}

	if inv == nil || len(inv.Servers) == 0 {
		return r
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
			r.hostsByName[hostName] = append(r.hostsByName[hostName], idx)
			r.hostOf = append(r.hostOf, hostName)
			idxs = append(idxs, idx)
		}
		r.groups[groupName] = idxs
	}

	return r
}

// resolve 解析 pattern 字符串，返回匹配的主机列表。
//
// 各子表达式只负责「匹配」；& / ! 的操作数走严格模式（匹配不到即报错，
// 不回退到独立主机），避免用户打错字时减法/交集静默失效。
func (r *resolver) resolve(pattern string, ov Overrides) ([]HostEntry, error) {
	tokens := splitPattern(pattern)

	var result []HostEntry

	if len(tokens) == 0 {
		sub, err := r.resolveSingle("g:*", true)
		if err != nil {
			return nil, err
		}
		result = sub
	} else {
		// 第一个 token 以 & 或 ! 开头 → 从 g:* 开始（严格模式）
		if strings.HasPrefix(tokens[0], "&") || strings.HasPrefix(tokens[0], "!") {
			sub, err := r.resolveSingle("g:*", true)
			if err != nil {
				return nil, err
			}
			result = sub
		}

		for _, t := range tokens {
			switch {
			case strings.HasPrefix(t, "&"):
				if t == "&" {
					return nil, fmt.Errorf("empty operand after %q in pattern %q", "&", pattern)
				}
				sub, err := r.resolveSingle(t[1:], true)
				if err != nil {
					return nil, err
				}
				result = intersectEntries(result, sub)

			case strings.HasPrefix(t, "!"):
				if t == "!" {
					return nil, fmt.Errorf("empty operand after %q in pattern %q", "!", pattern)
				}
				sub, err := r.resolveSingle(t[1:], true)
				if err != nil {
					return nil, err
				}
				result = subtractEntries(result, sub)

			default:
				sub, err := r.resolveSingle(t, false)
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
//
// strict=true 时匹配不到必须报错（用于 & / ! 操作数），不回退到独立主机。
func (r *resolver) resolveSingle(name string, strict bool) ([]HostEntry, error) {
	switch {
	case strings.HasPrefix(name, "g:"):
		return r.byGroup(name[2:], strict)
	case strings.HasPrefix(name, "h:"):
		return r.byHost(name[2:], strict)
	default:
		return r.byDefault(name, strict)
	}
}
