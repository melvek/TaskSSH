package task

import (
	"strings"

	"mestrap.com/taskssh/internal/config"
	"mestrap.com/taskssh/internal/resolve"
)

// Overrides 是 CLI 覆盖参数。
type Overrides struct {
	Port         int
	Username     string
	Password     string
	IdentityFile string
}

// HostEntry 是目标主机的一项。
type HostEntry struct {
	Name string      // 显示名：组内主机为 "组名/主机名"，独立主机为输入值
	Host config.Host // Host.Host 为最终连接用的 IP
	Vars resolve.Vars
}

// ResolveHosts 解析目标主机列表。
//
// 输入分类：
//   - 以 "!" 开头：exclude，从结果中排除
//   - 其他：include，加入结果
//
// 规则：
//   - 无 include 时，默认 include 为 g:*（所有组）
//   - exclude 的任何错误都跳过，不中断
//   - 按 Name 去重和过滤
//
// 输入支持以下形式：
//
//	prod                    精确匹配组名
//	prod/web1               精确匹配 "组名/主机名"
//	web1                    精确匹配组内主机名（多组同名时报错）
//	192.168.1.10            独立主机或 IP
//	g:web*                  glob 匹配组名
//	h:web*                  glob 匹配组内主机名
//	web*                    默认：同时匹配组名和主机名，去重合并
//	!g:web*                 排除：组名匹配 web* 的组
//	!h:web1                 排除：主机名匹配 web1 的主机
func ResolveHosts(hosts string, inv *config.Inventory, ov Overrides) ([]HostEntry, error) {
	var includes []string
	var excludes []string

	// 如果是空格或逗号，则视为分隔符
	hostNames := strings.FieldsFunc(hosts, func(r rune) bool {
		return r == ' ' || r == ','
	})

	for _, name := range hostNames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if strings.HasPrefix(name, "!") {
			excludes = append(excludes, name[1:])
		} else {
			includes = append(includes, name)
		}
	}

	// 无 include：默认 g:*
	if len(includes) == 0 {
		includes = []string{"g:*"}
	}

	var entries []HostEntry

	// 1. 处理 include
	for _, name := range includes {
		sub, err := resolveSingleHost(name, inv)
		if err != nil {
			return nil, err
		}
		entries = append(entries, sub...)
	}

	entries = dedupeEntries(entries)

	// 2. 处理 exclude
	if len(excludes) > 0 {
		excludeSet := make(map[string]bool)

		for _, name := range excludes {
			sub, err := resolveSingleHost(name, inv)
			if err != nil {
				// exclude 的任何错误都跳过
				continue
			}
			for _, e := range sub {
				excludeSet[e.Name] = true
			}
		}

		filtered := make([]HostEntry, 0, len(entries))
		for _, e := range entries {
			if !excludeSet[e.Name] {
				filtered = append(filtered, e)
			}
		}
		entries = filtered
	}

	applyOverrides(entries, ov)

	if err := decryptSecrets(entries, ov.Password); err != nil {
		return nil, err
	}

	return entries, nil
}
