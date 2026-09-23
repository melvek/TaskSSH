package action

import (
	"fmt"
	"os"
	"path/filepath"

	"mestrap.com/taskssh/internal/log"
	"mestrap.com/taskssh/internal/ssh"
)

// PushAction 上传文件到远程。
type PushAction struct{}

func (a *PushAction) Name() string { return "push" }

func (a *PushAction) Execute(ctx *Context) error {
	fileRaw, ok := ctx.With["file"]
	if !ok {
		return fmt.Errorf("push action requires 'file' parameter")
	}

	local, err := ctx.Vars.Replace(fmt.Sprintf("%v", fileRaw))
	if err != nil {
		return err
	}

	info, err := os.Stat(local)
	if err != nil {
		return fmt.Errorf("local file: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("directory upload not supported: %s", local)
	}

	dest, err := resolveDest(ctx)
	if err != nil {
		return err
	}

	force := toBool(ctx.With["force"])
	backup := toBool(ctx.With["backup"])
	policy := ssh.NewPolicy(force, backup)

	localAbs, err := filepath.Abs(local)
	if err != nil {
		localAbs = local
	}

	log.Info("Upload %s to %s", localAbs, dest)

	client, err := ssh.Connect(ctx.Host)
	if err != nil {
		return err
	}
	defer client.Close()

	finalPath, err := client.Upload(localAbs, dest, policy)
	if err != nil {
		return err
	}

	log.Success("Uploaded: %s -> %s", localAbs, finalPath)
	return nil
}

// resolveDest 解析远程目标路径。
//
// 优先级：
//  1. with 里的 dest（非空）
//  2. 变量池的 service_path
func resolveDest(ctx *Context) (string, error) {
	if destRaw, ok := ctx.With["dest"]; ok {
		s := fmt.Sprintf("%v", destRaw)
		if s != "" {
			return ctx.Vars.Replace(s)
		}
	}

	if sp, ok := ctx.Vars["service_path"]; ok {
		s := fmt.Sprintf("%v", sp)
		if s != "" {
			return s, nil
		}
	}

	return "", fmt.Errorf("push action requires 'dest' parameter or 'service_path' variable")
}

// toBool 把 any 转为 bool。
func toBool(v any) bool {
	switch x := v.(type) {
	case bool:
		return x
	case string:
		return x == "true" || x == "1" || x == "yes"
	}
	return false
}
