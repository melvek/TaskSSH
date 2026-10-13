package action

// Registry 管理所有Action
type Registry struct {
	actions map[string]Action
}

// NewRegistry 注册内置Action
func NewRegistry() *Registry {
	r := &Registry{
		actions: make(map[string]Action),
	}
	r.Register("command", &CommandAction{})
	r.Register("push", &PushAction{})
	r.Register("fetch", &FetchAction{})
	r.Register("srcipt", &ScriptAction{})
	return r
}

func (r *Registry) Register(name string, a Action) {
	r.actions[name] = a
}

func (r *Registry) Get(name string) Action {
	return r.actions[name]
}
