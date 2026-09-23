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

	remote, err := ctx.Vars.Replace(fmt.Sprintf("%v", fileRaw))
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
	fileName := path.Base(remotePath)

	// dest 不是目录 → 视为文件路径
	if !isDirDest(dest) {
		abs, err := filepath.Abs(dest)
		if err != nil {
			return "", fmt.Errorf("resolve path %s: %w", dest, err)
		}
		return abs, nil
	}

	// dest 是目录 → dest/<主机标识>/<文件名>
	hostID := buildHostID(host)
	finalPath := filepath.Join(dest, hostID, fileName)

	abs, err := filepath.Abs(finalPath)
	if err != nil {
		return "", fmt.Errorf("resolve path %s: %w", finalPath, err)
	}
	return abs, nil
}

// isDirDest 判断 dest 是否目录。
func isDirDest(dest string) bool {
	// 以分隔符结尾
	if strings.HasSuffix(dest, "/") || strings.HasSuffix(dest, string(os.PathSeparator)) {
		return true
	}

	// 已存在且是目录
	if info, err := os.Stat(dest); err == nil && info.IsDir() {
		return true
	}

	return false
}

// buildHostID 构造主机标识，用作本地目录名。
func buildHostID(host string) string {
	if host == "" {
		return "unknown"
	}
	return strings.NewReplacer("/", "_", "\\", "_").Replace(host)
}
