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

func (a *CommandAction) Execute(ctx *ActionContext) error {
	with, err := parseWith[CommandWith](ctx.With)
	if err != nil {
		return err
	}

	if with.Command == "" {
		return fmt.Errorf("command action requires 'command' parameter")
	}

	cmd, err := ctx.Vars.Replace(with.Command)
	if err != nil {
		return err
	}

	// 在本地环境下执行当前命令，主要用于 task 的步骤中，需要处理一些本地化的内容
	if with.Local {
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

	// 根据返回的退出码对执行结果进行判断，只要 exit code 不为0，就视为失败
	if result.ExitCode != 0 {
		return fmt.Errorf("command failed with exit code %d", result.ExitCode)
	}
	return nil
}

func runLocalCommand(ctx *ActionContext, cmd string) error {

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

func init() {
	Register("command", &CommandAction{})
}
