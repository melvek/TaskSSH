package action

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"mestrap.com/taskssh/internal/log"
	"mestrap.com/taskssh/internal/ssh"
)

// ScriptAction 上传本地脚本并在远程执行。
type ScriptAction struct{}

func (a *ScriptAction) Name() string { return "script" }

func (a *ScriptAction) Execute(ctx *Context) error {
	// 1. 解析 file
	fileRaw, ok := ctx.With["file"]
	if !ok {
		return fmt.Errorf("script action requires 'file' parameter")
	}
	fileStr := fmt.Sprintf("%v", fileRaw)
	if fileStr == "" {
		return fmt.Errorf("script action requires non-empty 'file' parameter")
	}
	local, err := ctx.Vars.Replace(fileStr)
	if err != nil {
		return err
	}

	localAbs, err := filepath.Abs(local)
	if err != nil {
		return fmt.Errorf("resolve path %s: %w", local, err)
	}

	// 2. 解析 dest，默认 /tmp
	dest := "/tmp"
	if destRaw, ok := ctx.With["dest"]; ok {
		s := fmt.Sprintf("%v", destRaw)
		if s != "" {
			dest, err = ctx.Vars.Replace(s)
			if err != nil {
				return err
			}
		}
	}

	// 3. 解析 remove / force
	remove := toBool(ctx.With["remove"])
	force := toBool(ctx.With["force"])

	policy := ssh.NewPolicy(force, false)

	// 4. 计算远程最终路径：dest 以 / 结尾视为目录
	remote := dest
	if strings.HasSuffix(dest, "/") {
		remote = path.Join(dest, filepath.Base(localAbs))
	}

	log.Info("Upload script %s to %s", localAbs, remote)

	client, err := ssh.Connect(ctx.Host)
	if err != nil {
		return err
	}
	defer client.Close()

	finalPath, err := client.Upload(localAbs, remote, policy)
	if err != nil {
		return err
	}

	// 5. 赋予执行权限
	if _, err := client.Exec(fmt.Sprintf("chmod +x %s", shellQuote(finalPath))); err != nil {
		return fmt.Errorf("chmod script: %w", err)
	}

	// 6. 执行脚本
	cmd := shellQuote(finalPath)
	if remove {
		// 执行后删除，不论成功失败
		cmd = fmt.Sprintf("%s; rc=$?; rm -f %s; exit $rc", cmd, shellQuote(finalPath))
	}

	log.Info("Execute script: %s", finalPath)

	result, err := client.Exec(cmd)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("script failed with exit code %d", result.ExitCode)
	}
	return nil
}

// shellQuote 对路径做最小限度的 shell 转义。
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
