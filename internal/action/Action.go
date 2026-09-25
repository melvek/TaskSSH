package action

import (
	"mestrap.com/taskssh/internal/config"
	"mestrap.com/taskssh/internal/resolve"
	"mestrap.com/taskssh/internal/ssh"
)

// Context 是动作执行上下文。
type Context struct {
	Host   *config.Host
	Vars   resolve.Vars
	With   map[string]any
	Client *ssh.Client // 新增：由 executor 注入，action 不再自己连接
}

// Action 是原子动作。
type Action interface {
	// Name 返回动作名，用于 step.action。
	Name() string
	// Execute 执行动作。
	Execute(ctx *Context) error
}
