package action

import (
	"sort"
)

// actions 是全局 action 实例表，由各 action 的 init() 填充。
var actions = map[string]Action{}

// Register 注册一个 action 实例。
// 由各 action 文件的 init() 调用。
func Register(name string, a Action) {
	actions[name] = a
}

// Get 按名称获取 action。未注册时返回 nil。
func Get(name string) Action {
	return actions[name]
}

// Names 返回所有已注册 action 的名称（用于错误提示）。
func Names() []string {
	out := make([]string, 0, len(actions))
	for n := range actions {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}
