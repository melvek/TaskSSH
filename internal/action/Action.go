package action

import (
	"context"

	"mestrap.com/taskssh/internal/config"
	"mestrap.com/taskssh/internal/console"
	"mestrap.com/taskssh/internal/resolve"
	"mestrap.com/taskssh/internal/ssh"
)

// 动作执行上下文
// 主要用于任务流模式下，多任务间的参数传递
// 后续扩展时可以将执行结果注册到此上下文中
type ActionContext struct {
	Context context.Context
	Host    *config.Host
	Vars    resolve.Vars
	With    any
	Client  *ssh.Client
	DryRun  bool
	Logger  *console.Logger
}

func (ctx *ActionContext) Log() *console.Logger {
	if ctx.Logger != nil {
		return ctx.Logger
	}
	return console.DefaultLogger()
}

type Action interface {
	Execute(ctx *ActionContext) error
}
