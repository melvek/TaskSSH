package action

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"mestrap.com/taskssh/internal/log"
	"mestrap.com/taskssh/internal/ssh"
)

// FetchAction 从远程下载文件。
type FetchAction struct{}

func (a *FetchAction) Name() string { return "fetch" }

func (a *FetchAction) Execute(ctx *Context) error {
	fileRaw, ok := ctx.With["file"]
	if !ok {
		return fmt.Errorf("fetch action requires 'file' parameter")
	}

	fileStr := fmt.Sprintf("%v", fileRaw)
	if fileStr == "" {
		return fmt.Errorf("fetch action requires non-empty 'file' parameter")
	}

	remote, err := ctx.Vars.Replace(fileStr)
	if err != nil {
		return err
	}

	// 解析 dest，默认当前目录
	dest := "./"
	if destRaw, ok := ctx.With["dest"]; ok {
		s := fmt.Sprintf("%v", destRaw)
		if s != "" {
			dest, err = ctx.Vars.Replace(s)
			if err != nil {
				return err
			}
		}
	}

	// 计算最终本地路径
	finalPath, err := resolveFinalPath(dest, remote, ctx.Host.Host)
	if err != nil {
		return err
	}

	log.Info("Download %s to %s", remote, finalPath)

	client, err := ssh.Connect(ctx.Host)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.Download(remote, finalPath); err != nil {
		return err
	}

	log.Success("Downloaded: %s -> %s", remote, finalPath)
	return nil
}

// resolveFinalPath 计算最终本地路径。
//
// 规则：
//  1. dest 是文件路径（非目录）→ 直接用 dest
//  2. dest 是目录 → dest/<主机标识>/<文件名>
func resolveFinalPath(dest, remotePath, host string) (string, error) {
	if remotePath == "" {
		return "", fmt.Errorf("remote path is empty")
	}

	fileName := path.Base(remotePath)

	// 远程路径异常（如以 / 结尾）
	if fileName == "." || fileName == "/" || fileName == "" {
		return "", fmt.Errorf("remote path is invalid: %s", remotePath)
	}

	// dest 不是目录 → 视为文件路径
	if !isDirDest(dest) {
		return absPath(dest)
	}

	// dest 是目录 → dest/<主机标识>/<文件名>
	hostID := buildHostID(host)
	return absPath(filepath.Join(dest, hostID, fileName))
}

// absPath 取绝对路径。
func absPath(p string) (string, error) {
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", fmt.Errorf("resolve path %s: %w", p, err)
	}
	return abs, nil
}

// isDirDest 判断 dest 是否目录。
//
// 规则：
//  1. 以 / 或 \ 结尾 → 目录
//  2. 已存在且是目录 → 目录
func isDirDest(dest string) bool {
	if strings.HasSuffix(dest, "/") || strings.HasSuffix(dest, string(os.PathSeparator)) {
		return true
	}

	if info, err := os.Stat(dest); err == nil && info.IsDir() {
		return true
	}

	return false
}

// buildHostID 构造主机标识，用作本地目录名。
//
// 主机标识中的 / 和 \ 替换为 _。
func buildHostID(host string) string {
	if host == "" {
		return "unknown"
	}
	return strings.NewReplacer("/", "_", "\\", "_").Replace(host)
}
