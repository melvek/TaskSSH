package action

import (
	"mestrap.com/taskssh/internal/config"
	"mestrap.com/taskssh/internal/resolve"
	"mestrap.com/taskssh/internal/ssh"
)

// 动作执行上下文
type Context struct {
	Host   *config.Host
	Vars   resolve.Vars
	With   any
	Client *ssh.Client
	DryRun bool
}

type Action interface {
	Name() string
	Execute(ctx *Context) error
}
