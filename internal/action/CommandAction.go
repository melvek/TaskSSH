package action

import (
	"fmt"

	"mestrap.com/taskssh/internal/console"
)

// CommandAction 执行远程命令
type CommandAction struct{}

var _ Action = &CommandAction{}

type CommandWith struct {
	Command string `yaml:"command"`
}

func (a *CommandAction) Name() string { return "command" }

func (a *CommandAction) Execute(ctx *Context) error {
	w, err := decodeWith[CommandWith](ctx.With)
	if err != nil {
		return err
	}

	if w.Command == "" {
		return fmt.Errorf("command action requires 'command' parameter")
	}

	cmd, err := ctx.Vars.Replace(w.Command)
	if err != nil {
		return err
	}

	console.Info("Execute command: %s", cmd)

	result, err := ctx.Client.Exec(cmd)
	if err != nil {
		return err
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("command failed with exit code %d", result.ExitCode)
	}
	return nil
}
