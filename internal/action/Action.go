package action

import (
	"mestrap.com/taskssh/internal/config"
	"mestrap.com/taskssh/internal/resolve"
)

// Context 是动作执行上下文。
type Context struct {
	Host *config.Host
	Vars resolve.Vars
	With map[string]any
}

// Action 是原子动作。
type Action interface {
	// Name 返回动作名，用于 step.action。
	Name() string
	// Execute 执行动作。
	Execute(ctx *Context) error
}
