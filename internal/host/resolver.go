package host

import (
	"sort"
	"strings"

	"mestrap.com/taskssh/internal/config"
)

// resolver 是展开后的主机索引。
//
// 构造时一次性把所有组和主机展开成 []HostEntry，
// 后续所有 pattern 查找走 map，避免重复展开。
type resolver struct {
	inv        *config.Inventory
	entries    []HostEntry         // 所有主机，按组名 / 主机名字典序
	byName     map[string]int      // "group/host" → entries 索引
	groups     map[string][]int    // group → entries 索引列表
	groupOrder []string            // 组名（已排序）
	hostOrder  map[string][]string // group → 主机名（已排序）
}

// newResolver 展开清单，构建主机索引。
func newResolver(inv *config.Inventory) *resolver {
	r := &resolver{
		inv:       inv,
		byName:    make(map[string]int),
		groups:    make(map[string][]int),
		hostOrder: make(map[string][]string),
	}

	// 排序组名
	r.groupOrder = make([]string, 0, len(inv.Servers))
	for name := range inv.Servers {
		r.groupOrder = append(r.groupOrder, name)
	}
	sort.Strings(r.groupOrder)

	// 逐个组展开
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
// pattern 用空格或逗号分隔多个子表达式，从左到右依次应用：
//
//	prefix        操作
//	（无）        并集：加到当前结果
//	!             差集：从当前结果移除
//	&             交集：在当前结果中过滤
//
// 初始集：
//   - 第一个子表达式以 & 或 ! 开头 → 从 g:* 开始
//   - 否则 → 从空集开始，第一个子表达式作为并集加入
//
// 无任何子表达式时默认 g:*。! 匹配错误容忍，& 匹配错误时结果为空并继续。
func (r *resolver) resolve(pattern string, ov Overrides) ([]HostEntry, error) {
	tokens := splitPattern(pattern)

	var result []HostEntry

	if len(tokens) == 0 {
		// 空 pattern → 默认 g:*
		sub, err := r.resolveSingle("g:*")
		if err != nil {
			return nil, err
		}
		result = sub
	} else {
		// 第一个 token 是 & 或 ! 开头 → 从 g:* 开始
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
					// & 过滤：匹配不到时结果为空，不中断，后续 token 继续处理
					result = nil
					continue
				}
				result = intersectEntries(result, sub)

			case strings.HasPrefix(t, "!"):
				sub, err := r.resolveSingle(t[1:])
				if err != nil {
					// ! 排除：匹配不到时跳过，不中断
					continue
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
