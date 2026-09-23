package action

import (
	"fmt"

	"mestrap.com/taskssh/internal/ssh"
)

// CommandAction 执行远程命令。
type CommandAction struct{}

func (a *CommandAction) Name() string { return "command" }

func (a *CommandAction) Execute(ctx *Context) error {
	raw, ok := ctx.With["command"]
	if !ok {
		return fmt.Errorf("command action requires 'command' parameter")
	}

	cmd := fmt.Sprintf("%v", raw)
	cmd, err := ctx.Vars.Replace(cmd)
	if err != nil {
		return err
	}

	client, err := ssh.Connect(ctx.Host)
	if err != nil {
		return err
	}
	defer client.Close()

	result, err := client.Exec(cmd)
	if err != nil {
		return err
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("command failed with exit code %d", result.ExitCode)
	}
	return nil
}
