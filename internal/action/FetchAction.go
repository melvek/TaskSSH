package action

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"mestrap.com/taskssh/internal/console"
)

// FetchAction 从远程下载文件或目录。
type FetchAction struct{}

var _ Action = &FetchAction{}

type FetchWith struct {
	File   string `yaml:"file"`
	Dest   string `yaml:"dest"`
	Zip    bool   `yaml:"zip"`
	TmpDir string `yaml:"tmp_dir"`
}

func (a *FetchAction) Name() string { return "fetch" }

func (a *FetchAction) Execute(ctx *Context) error {
	w, err := decodeWith[FetchWith](ctx.With)
	if err != nil {
		return err
	}

	if w.File == "" {
		return fmt.Errorf("fetch action requires 'file' parameter")
	}
	fileStr := fmt.Sprintf("%v", w.File)
	if fileStr == "" {
		return fmt.Errorf("fetch action requires non-empty 'file' parameter")
	}
	remote, err := ctx.Vars.Replace(fileStr)
	if err != nil {
		return err
	}

	dest := "./"
	if w.Dest != "" {
		dest, err = ctx.Vars.Replace(w.Dest)
		if err != nil {
			return err
		}
	}

	info, err := ctx.Client.Stat(remote)
	if err != nil {
		return err
	}

	if info.IsDir() {
		if w.Zip {
			tmpDir := "/tmp"
			if w.TmpDir != "" {
				tmpDir, err = ctx.Vars.Replace(w.TmpDir)
				if err != nil {
					return err
				}
			}
			return a.fetchDirZip(ctx, remote, dest, tmpDir)
		}
		return a.fetchDirRecursive(ctx, remote, dest)
	}

	return a.fetchFile(ctx, remote, dest)
}

// 下载单个文件
func (a *FetchAction) fetchFile(ctx *Context, remote, dest string) error {
	finalPath, err := resolveFinalPath(dest, remote, ctx.Host.Host)
	if err != nil {
		return err
	}

	console.Info("Download %s to %s", remote, finalPath)

	if err := ctx.Client.Download(remote, finalPath); err != nil {
		return err
	}

	console.Success("Downloaded: %s -> %s", remote, finalPath)
	return nil
}

// 下载文件夹
func (a *FetchAction) fetchDirRecursive(ctx *Context, remote, dest string) error {
	console.Info("Download directory %s to %s (recursive)", remote, dest)

	baseName := path.Base(remote)
	hostID := buildHostID(ctx.Host.Host)
	rootLocal := filepath.Join(dest, hostID, baseName)

	files, err := ctx.Client.ListFiles(remote)
	if err != nil {
		return err
	}

	for _, rf := range files {
		rel, err := filepath.Rel(remote, rf)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		localPath := filepath.Join(rootLocal, rel)

		if err := os.MkdirAll(filepath.Dir(localPath), 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", filepath.Dir(localPath), err)
		}

		console.Info("Download %s to %s", rf, localPath)

		if err := ctx.Client.Download(rf, localPath); err != nil {
			return err
		}
	}

	console.Success("Directory downloaded: %s -> %s (%d files)", remote, rootLocal, len(files))
	return nil
}

// 下载压缩包
func (a *FetchAction) fetchDirZip(ctx *Context, remote, dest string, tmpDir string) error {
	baseName := path.Base(remote)
	remoteDir := path.Dir(remote)

	hostID := buildHostID(ctx.Host.Host)
	remoteTmpDir := path.Join(tmpDir, "taskssh-fetch-"+hostID)

	zipName := fmt.Sprintf("%s.%d.zip", baseName, time.Now().UnixMicro())
	remoteZip := path.Join(remoteTmpDir, zipName)

	console.Info("Zip directory %s on remote", remote)

	zipCmd := fmt.Sprintf(
		`which zip &>/dev/null && { mkdir -p %s && cd %s && zip -r %s %s; } || `+
			`{ echo "need install 'zip' command"; exit 1; }`,
		shellQuote(remoteTmpDir),
		shellQuote(remoteDir),
		shellQuote(remoteZip),
		shellQuote(baseName),
	)
	console.Info("Execute zip files ...")

	result, err := ctx.Client.Exec(zipCmd)
	if err != nil {
		return fmt.Errorf("zip: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("zip failed with exit code %d", result.ExitCode)
	}

	localZip, err := os.CreateTemp("", "taskssh-fetch-*.zip")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	localZipPath := localZip.Name()
	localZip.Close()
	defer os.Remove(localZipPath)

	console.Info("Download zip %s to %s", remoteZip, localZipPath)

	if err := ctx.Client.Download(remoteZip, localZipPath); err != nil {
		return err
	}

	rmCmd := fmt.Sprintf("rm -f %s", shellQuote(remoteZip))
	if _, err := ctx.Client.Exec(rmCmd); err != nil {
		console.Warn("Failed to remove remote zip: %v", err)
	}

	rootLocal := filepath.Join(dest, hostID, baseName)

	console.Info("Extract %s to %s", localZipPath, rootLocal)

	if err := extractZip(localZipPath, rootLocal); err != nil {
		return fmt.Errorf("extract: %w", err)
	}

	console.Success("Directory downloaded: %s -> %s", remote, rootLocal)
	return nil
}

// 解析完整远程路径
func resolveFinalPath(dest, remotePath, host string) (string, error) {
	if remotePath == "" {
		return "", fmt.Errorf("remote path is empty")
	}

	fileName := path.Base(remotePath)

	if fileName == "." || fileName == "/" || fileName == "" {
		return "", fmt.Errorf("remote path is invalid: %s", remotePath)
	}

	if !isDirDest(dest) {
		return absPath(dest)
	}

	hostID := buildHostID(host)
	return absPath(filepath.Join(dest, hostID, fileName))
}

func absPath(p string) (string, error) {
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", fmt.Errorf("resolve path %s: %w", p, err)
	}
	return abs, nil
}

func isDirDest(dest string) bool {
	if strings.HasSuffix(dest, "/") || strings.HasSuffix(dest, string(os.PathSeparator)) {
		return true
	}

	if info, err := os.Stat(dest); err == nil && info.IsDir() {
		return true
	}

	return false
}

func buildHostID(host string) string {
	if host == "" {
		return "unknown"
	}
	return strings.NewReplacer("/", "_", "\\", "_").Replace(host)
}

func extractZip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}
	defer r.Close()

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", destDir, err)
	}

	destClean := filepath.Clean(destDir)

	for _, f := range r.File {
		target := filepath.Join(destDir, f.Name)

		if !strings.HasPrefix(filepath.Clean(target), destClean+string(os.PathSeparator)) &&
			filepath.Clean(target) != destClean {
			return fmt.Errorf("invalid zip entry: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, f.Mode()); err != nil {
				return fmt.Errorf("mkdir %s: %w", target, err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", filepath.Dir(target), err)
		}

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("open zip entry %s: %w", f.Name, err)
		}

		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return fmt.Errorf("create %s: %w", target, err)
		}

		if _, err := io.Copy(out, rc); err != nil {
			out.Close()
			rc.Close()
			return fmt.Errorf("extract %s: %w", f.Name, err)
		}

		out.Close()
		rc.Close()
	}

	return nil
}
