package action

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"mestrap.com/taskssh/internal/console"
)

// CommandAction 执行远程命令
type CommandAction struct{}

var _ Action = &CommandAction{}

type CommandWith struct {
	Command string `yaml:"command"`
	Local   bool   `yaml:"local"`
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

	if w.Local {
		return runLocalCommand(ctx, cmd)
	}

	console.Info("Execute command: %s", cmd)

	if ctx.DryRun {
		return nil
	}

	result, err := ctx.Client.Exec(cmd)
	if err != nil {
		return err
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("command failed with exit code %d", result.ExitCode)
	}
	return nil
}

func runLocalCommand(ctx *Context, cmd string) error {

	console.Info("Execute local command: %s", cmd)

	if ctx.DryRun {
		return nil
	}

	var c *exec.Cmd
	if runtime.GOOS == "windows" {
		c = exec.Command("cmd", "/C", cmd)
	} else {
		c = exec.Command("sh", "-c", cmd)
	}

	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		return fmt.Errorf("local command failed: %w", err)
	}
	return nil
}
