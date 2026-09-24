package action

import (
	"sort"
)

// Registry 管理所有动作。
type Registry struct {
	actions map[string]Action
}

// NewRegistry 创建注册表，自动注册内置动作。
func NewRegistry() *Registry {
	r := &Registry{
		actions: make(map[string]Action),
	}
	r.Register(&CommandAction{})
	r.Register(&PushAction{})
	r.Register(&FetchAction{})
	r.Register(&ScriptAction{})
	return r
}

// Register 注册一个动作。
func (r *Registry) Register(a Action) {
	r.actions[a.Name()] = a
}

// Get 按名查找动作。
func (r *Registry) Get(name string) Action {
	return r.actions[name]
}

// Names 返回所有动作名（排序）。
func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.actions))
	for name := range r.actions {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
