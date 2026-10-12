package action

// Registry 管理所有动作
type Registry struct {
	actions map[string]Action
}

// NewRegistry 创建注册表，自动注册内置动作
func NewRegistry() *Registry {
	r := &Registry{
		actions: make(map[string]Action),
	}
	r.Register(&CommandAction{}, "command")
	r.Register(&PushAction{}, "push")
	r.Register(&FetchAction{}, "fetch")
	r.Register(&ScriptAction{}, "srcipt")
	return r
}

// Register 注册一个动作
func (r *Registry) Register(a Action, name string) {
	r.actions[name] = a
}

// Get 按名查找动作
func (r *Registry) Get(name string) Action {
	return r.actions[name]
}
