package action

import (
	"fmt"
	"path/filepath"

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

	// dest 可选，默认当前目录
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

	client, err := ssh.Connect(ctx.Host)
	if err != nil {
		return err
	}
	defer client.Close()

	finalPath, err := client.Download(remote, dest)
	if err != nil {
		return err
	}

	localAbs, err := filepath.Abs(finalPath)
	if err != nil {
		localAbs = finalPath
	}

	fmt.Printf("Downloaded: %s -> %s\n", remote, localAbs)
	return nil
}
