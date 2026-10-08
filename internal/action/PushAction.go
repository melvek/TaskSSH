package action

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"mestrap.com/taskssh/internal/console"
	"mestrap.com/taskssh/internal/ssh"
)

// PushAction 上传文件或目录到远程。
type PushAction struct{}

var _ Action = &PushAction{}

func (a *PushAction) Name() string { return "push" }

type PushWith struct {
	File  string `yaml:"file"`
	Dest  string `yaml:"dest"`
	Force bool   `yaml:"force"`
	Zip   bool   `yaml:"zip"`
}

func (a *PushAction) Execute(ctx *Context) error {
	w, err := decodeWith[PushWith](ctx.With)
	if err != nil {
		return err
	}
	if w.File == "" {
		return fmt.Errorf("push action requires 'file' parameter")
	}

	local, err := ctx.Vars.Replace(w.File)
	if err != nil {
		return err
	}

	localAbs, err := filepath.Abs(local)
	if err != nil {
		return fmt.Errorf("resolve path %s: %w", local, err)
	}

	info, err := os.Stat(localAbs)
	if err != nil {
		return fmt.Errorf("local file: %w", err)
	}

	if w.Dest == "" {
		return fmt.Errorf("push action requires 'dest' parameter")
	}

	remote, err := ctx.Vars.Replace(w.Dest)
	if err != nil {
		return err
	}

	policy := ssh.NewPolicy(w.Force)

	if info.IsDir() {
		if w.Zip {
			return a.pushDirZip(ctx, localAbs, remote, policy)
		}
		return a.pushDirRecursive(ctx, localAbs, remote, policy)
	}

	return a.pushFile(ctx, localAbs, remote, policy)
}

func (a *PushAction) pushFile(ctx *Context, localAbs, dest string, policy ssh.OverwritePolicy) error {
	console.Info("Upload %s to %s", localAbs, dest)

	if ctx.DryRun {
		return nil
	}

	finalPath, err := ctx.Client.Upload(localAbs, dest, policy)
	if err != nil {
		return err
	}
	console.Info("Uploaded: %s -> %s", localAbs, finalPath)
	return nil
}

func (a *PushAction) pushDirRecursive(ctx *Context, localAbs, dest string, policy ssh.OverwritePolicy) error {
	console.Info("Upload directory %s to %s (recursive)", localAbs, dest)

	if ctx.DryRun {
		return nil
	}

	base := filepath.Dir(localAbs)
	rootRemote := path.Join(dest, filepath.Base(localAbs))
	if err := ctx.Client.MkdirAll(rootRemote); err != nil {
		return err
	}

	fileCount := 0
	err := filepath.Walk(localAbs, func(localPath string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		rel, err := filepath.Rel(base, localPath)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		remotePath := path.Join(dest, rel)

		if info.IsDir() {
			return ctx.Client.MkdirAll(remotePath)
		}

		if _, err := ctx.Client.Upload(localPath, remotePath, policy); err != nil {
			return fmt.Errorf("upload %s: %w", localPath, err)
		}
		fileCount++
		return nil
	})
	if err != nil {
		return err
	}

	console.Info("Directory uploaded: %s -> %s (%d files)", localAbs, dest, fileCount)
	return nil
}

func (a *PushAction) pushDirZip(ctx *Context, localAbs, dest string, policy ssh.OverwritePolicy) error {
	console.Info("Zip directory %s", localAbs)

	zipPath, err := zipDir(localAbs)
	if err != nil {
		return fmt.Errorf("zip directory: %w", err)
	}
	defer os.Remove(zipPath)

	console.Info("Upload zip %s to %s", zipPath, dest)

	finalZip, err := ctx.Client.Upload(zipPath, dest, policy)
	if err != nil {
		return err
	}
	console.Success("Uploaded zip: %s -> %s", zipPath, finalZip)

	unzipCmd := fmt.Sprintf(
		`which unzip &>/dev/null && { cd %s && unzip -o %s && rm -f %s; } || `+
			`{ echo "need install 'unzip' command"; rm -f %s; exit 1; }`,
		shellQuote(dest),
		shellQuote(finalZip),
		shellQuote(finalZip),
		shellQuote(finalZip),
	)
	console.Info("Execute unzip: %s", unzipCmd)

	result, err := ctx.Client.Exec(unzipCmd)
	if err != nil {
		return fmt.Errorf("unzip: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("unzip failed with exit code %d", result.ExitCode)
	}

	console.Info("Directory uploaded: %s -> %s", localAbs, dest)
	return nil
}

func zipDir(dir string) (string, error) {
	tmpFile, err := os.CreateTemp("", "taskssh-push-*.zip")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	defer tmpFile.Close()

	zw := zip.NewWriter(tmpFile)

	base := filepath.Dir(dir)

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(base, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		w, err := zw.Create(rel)
		if err != nil {
			return err
		}

		src, err := os.Open(path)
		if err != nil {
			return err
		}
		defer src.Close()

		if _, err := io.Copy(w, src); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		zw.Close()
		return "", err
	}

	if err := zw.Close(); err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}
