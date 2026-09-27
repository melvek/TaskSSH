package action

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"mestrap.com/taskssh/internal/console"
	"mestrap.com/taskssh/internal/ssh"
)

// ScriptAction 上传本地脚本并在远程执行。
type ScriptAction struct{}

var _ Action = &ScriptAction{}

func (a *ScriptAction) Name() string { return "script" }

type ScriptWith struct {
	File   string `yaml:"file"`
	Dest   string `yaml:"dest"`
	Remove bool   `yaml:"remove"`
	Force  bool   `yaml:"force"`
}

func (a *ScriptAction) Execute(ctx *Context) error {
	w, err := decodeWith[ScriptWith](ctx.With)
	if err != nil {
		return err
	}

	localAbs, err := filepath.Abs(w.File)
	if err != nil {
		return fmt.Errorf("resolve path %s: %w", w.File, err)
	}

	remote := resolveScriptPath(w.Dest, localAbs)

	console.Info("Upload script %s to %s", localAbs, remote)

	// 上传脚本到远程服务器
	policy := ssh.NewPolicy(w.Force)

	var finalPath = fmt.Sprintf("%s/%s", remote, filepath.Base(w.File))

	if !ctx.DryRun {
		f, err := ctx.Client.Upload(localAbs, remote, policy)
		if err != nil {
			return err
		}

		// 获取实际上传后的文件路径
		finalPath = f

		// 给脚本可执行权限
		if _, err := ctx.Client.Exec(fmt.Sprintf("chmod +x %s", shellQuote(finalPath))); err != nil {
			return fmt.Errorf("chmod script: %w", err)
		}
	}

	cmd := shellQuote(finalPath)
	// 如果需要删除，则在需要执行的脚本后，拼接删除脚本操作
	if w.Remove {
		console.Info("Delete remote tmp script %s", finalPath)
		if !ctx.DryRun {
			cmd = fmt.Sprintf("%s; rc=$?; rm -f %s; exit $rc", cmd, shellQuote(finalPath))
		}
	}

	console.Info("Execute script %s", finalPath)

	if !ctx.DryRun {
		// 执行脚本
		result, err := ctx.Client.Exec(cmd)
		if err != nil {
			return err
		}
		if result.ExitCode != 0 {
			return fmt.Errorf("script failed with exit code %d", result.ExitCode)
		}
	}
	return nil
}

func resolveScriptPath(dest, localAbs string) string {
	if strings.HasSuffix(dest, "/") {
		return path.Join(dest, filepath.Base(localAbs))
	}
	return dest
}
