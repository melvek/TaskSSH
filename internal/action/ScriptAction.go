package action

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"mestrap.com/taskssh/internal/log"
	"mestrap.com/taskssh/internal/resolve"
	"mestrap.com/taskssh/internal/ssh"
)

// ScriptAction 上传本地脚本并在远程执行。
type ScriptAction struct{}

func (a *ScriptAction) Name() string { return "script" }

type scriptParams struct {
	file   string
	dest   string
	remove bool
	force  bool
}

func (a *ScriptAction) Execute(ctx *Context) error {
	params, err := parseScriptWith(ctx.With, ctx.Vars)
	if err != nil {
		return err
	}

	localAbs, err := filepath.Abs(params.file)
	if err != nil {
		return fmt.Errorf("resolve path %s: %w", params.file, err)
	}

	remote := resolveScriptPath(params.dest, localAbs)

	log.Info("Upload script %s to %s", localAbs, remote)

	policy := ssh.NewPolicy(params.force, false)

	finalPath, err := ctx.Client.Upload(localAbs, remote, policy)
	if err != nil {
		return err
	}

	if _, err := ctx.Client.Exec(fmt.Sprintf("chmod +x %s", shellQuote(finalPath))); err != nil {
		return fmt.Errorf("chmod script: %w", err)
	}

	cmd := shellQuote(finalPath)
	if params.remove {
		cmd = fmt.Sprintf("%s; rc=$?; rm -f %s; exit $rc",
			cmd, shellQuote(finalPath))
	}

	log.Info("Execute script: %s", finalPath)

	result, err := ctx.Client.Exec(cmd)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("script failed with exit code %d", result.ExitCode)
	}
	return nil
}

func parseScriptWith(with map[string]any, vars resolve.Vars) (*scriptParams, error) {
	fileRaw, ok := with["file"]
	if !ok {
		return nil, fmt.Errorf("script action requires 'file' parameter")
	}
	fileStr := fmt.Sprintf("%v", fileRaw)
	if fileStr == "" {
		return nil, fmt.Errorf("script action requires non-empty 'file' parameter")
	}
	file, err := vars.Replace(fileStr)
	if err != nil {
		return nil, err
	}

	dest := "/tmp"
	if destRaw, ok := with["dest"]; ok {
		s := fmt.Sprintf("%v", destRaw)
		if s != "" {
			dest, err = vars.Replace(s)
			if err != nil {
				return nil, err
			}
		}
	}

	return &scriptParams{
		file:   file,
		dest:   dest,
		remove: toBool(with["remove"]),
		force:  toBool(with["force"]),
	}, nil
}

func resolveScriptPath(dest, localAbs string) string {
	if strings.HasSuffix(dest, "/") {
		return path.Join(dest, filepath.Base(localAbs))
	}
	return dest
}
