package action

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

const maxCommandDiagnosticBytes = 4096

// CommandError 描述远程命令的非零退出及其诊断输出。
type CommandError struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

func (e *CommandError) Error() string {
	message := fmt.Sprintf("command failed with exit code %d", e.ExitCode)
	diagnostic := strings.TrimSpace(e.Stderr)
	if diagnostic == "" {
		diagnostic = strings.TrimSpace(e.Stdout)
	}
	if diagnostic == "" {
		return message
	}
	if len(diagnostic) > maxCommandDiagnosticBytes {
		diagnostic = strings.ToValidUTF8(diagnostic[:maxCommandDiagnosticBytes], "?") + "... (truncated)"
	}
	diagnostic = strings.ReplaceAll(diagnostic, "\r\n", "\n")
	diagnostic = strings.ReplaceAll(diagnostic, "\r", "")
	diagnostic = strings.ReplaceAll(diagnostic, "\n", "\n    ")
	return message + ":\n    " + diagnostic
}

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

	ctx.Log().Debug("Execute command: %s", cmd)

	if ctx.DryRun {
		return nil
	}

	result, err := ctx.Client.ExecContext(ctx.Context, cmd, ctx.Log())
	if err != nil {
		return err
	}

	// 根据返回的退出码对执行结果进行判断，只要 exit code 不为0，就视为失败
	if result.ExitCode != 0 {
		return &CommandError{
			ExitCode: result.ExitCode,
			Stdout:   result.Stdout,
			Stderr:   result.Stderr,
		}
	}
	return nil
}

func runLocalCommand(ctx *ActionContext, cmd string) error {

	ctx.Log().Debug("Execute local command: %s", cmd)

	if ctx.DryRun {
		return nil
	}

	var c *exec.Cmd
	if runtime.GOOS == "windows" {
		c = exec.Command("cmd", "/C", cmd)
	} else {
		c = exec.Command("sh", "-c", cmd)
	}

	c.Stdout = ctx.Log().CommandWriter()
	c.Stderr = ctx.Log().StderrWriter()

	if err := c.Run(); err != nil {
		return fmt.Errorf("local command failed: %w", err)
	}
	return nil
}

func init() {
	Register("command", &CommandAction{})
}
