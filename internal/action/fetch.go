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

	"mestrap.com/taskssh/internal/fileutil"
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

func (a *FetchAction) Execute(ctx *ActionContext) error {
	with, err := parseWith[FetchWith](ctx.With)
	if err != nil {
		return err
	}

	if with.File == "" {
		return fmt.Errorf("fetch action requires 'file' parameter")
	}

	remote, err := ctx.Vars.Replace(with.File)
	if err != nil {
		return err
	}
	if remote == "" {
		return fmt.Errorf("fetch action: 'file' resolved to empty string")
	}

	dest := "./"
	if with.Dest != "" {
		dest, err = ctx.Vars.Replace(with.Dest)
		if err != nil {
			return err
		}

		if dest == "" {
			return fmt.Errorf("fetch action: 'dest' resolved to empty string")
		}
	}

	if ctx.DryRun {
		ctx.Log().Info("Fetch %s -> %s (dry-run)", remote, dest)
		return nil
	}

	info, err := ctx.Client.Lstat(remote)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refusing to fetch remote symbolic link %s", remote)
	}

	if info.IsDir() {
		if with.Zip {
			tmpDir := "/tmp"
			if with.TmpDir != "" {
				tmpDir, err = ctx.Vars.Replace(with.TmpDir)
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
func (a *FetchAction) fetchFile(ctx *ActionContext, remote, dest string) error {
	finalPath, err := resolveFinalPath(dest, remote, ctx.Host.Host)
	if err != nil {
		return err
	}

	ctx.Log().Debug("Download %s to %s", remote, finalPath)

	if err := ctx.Client.Download(remote, finalPath); err != nil {
		return err
	}

	ctx.Log().Success("Downloaded: %s -> %s", remote, finalPath)
	return nil
}

// 下载文件夹
func (a *FetchAction) fetchDirRecursive(ctx *ActionContext, remote, dest string) error {
	ctx.Log().Debug("Download directory %s to %s (recursive)", remote, dest)

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
		if rel == ".." || filepath.IsAbs(rel) || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
			return fmt.Errorf("invalid remote file path: %s", rf)
		}
		rel = filepath.ToSlash(rel)

		localPath := filepath.Join(rootLocal, rel)

		ctx.Log().Debug("Download %s to %s", rf, localPath)

		if err := ctx.Client.Download(rf, localPath); err != nil {
			return err
		}
	}

	ctx.Log().Success("Directory downloaded: %s -> %s (%d files)", remote, rootLocal, len(files))
	return nil
}

// 下载压缩包
func (a *FetchAction) fetchDirZip(ctx *ActionContext, remote, dest string, tmpDir string) error {
	baseName := path.Base(remote)
	remoteDir := path.Dir(remote)

	hostID := buildHostID(ctx.Host.Host)
	remoteTmpDir := path.Join(tmpDir, "taskssh-fetch-"+hostID)

	zipName := fmt.Sprintf("%s.%d.zip", baseName, time.Now().UnixMicro())
	remoteZip := path.Join(remoteTmpDir, zipName)

	ctx.Log().Debug("Zip directory %s on remote", remote)

	zipCmd := fmt.Sprintf(
		`which zip &>/dev/null && { mkdir -p %s && cd %s && zip -r -y %s %s; } || `+
			`{ echo "need install 'zip' command"; exit 1; }`,
		shellQuote(remoteTmpDir),
		shellQuote(remoteDir),
		shellQuote(remoteZip),
		shellQuote(baseName),
	)
	ctx.Log().Debug("Execute zip files ...")

	result, err := ctx.Client.ExecContext(ctx.Context, zipCmd, ctx.Log())
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

	ctx.Log().Debug("Download zip %s to %s", remoteZip, localZipPath)

	if err := ctx.Client.Download(remoteZip, localZipPath); err != nil {
		return err
	}

	rmCmd := fmt.Sprintf("rm -f %s", shellQuote(remoteZip))
	if _, err := ctx.Client.ExecContext(ctx.Context, rmCmd, ctx.Log()); err != nil {
		ctx.Log().Warn("Failed to remove remote zip: %v", err)
	}

	rootLocal := filepath.Join(dest, hostID, baseName)

	ctx.Log().Debug("Extract %s to %s", localZipPath, rootLocal)

	if err := extractZip(localZipPath, rootLocal); err != nil {
		return fmt.Errorf("extract: %w", err)
	}

	ctx.Log().Success("Directory downloaded: %s -> %s", remote, rootLocal)
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

	if err := fileutil.EnsureNoSymlinkPath(destDir); err != nil {
		return err
	}
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", destDir, err)
	}

	destClean := filepath.Clean(destDir)

	for _, f := range r.File {
		if f.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("refusing to extract symbolic link: %s", f.Name)
		}

		entryPath := filepath.FromSlash(f.Name)
		if filepath.IsAbs(entryPath) || filepath.VolumeName(entryPath) != "" {
			return fmt.Errorf("invalid zip entry: %s", f.Name)
		}
		target := filepath.Clean(filepath.Join(destClean, entryPath))
		rel, err := filepath.Rel(destClean, target)
		if err != nil {
			return fmt.Errorf("invalid zip entry %s: %w", f.Name, err)
		}
		if rel == ".." || filepath.IsAbs(rel) || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
			return fmt.Errorf("invalid zip entry: %s", f.Name)
		}
		if err := fileutil.EnsureNoSymlinkPath(target); err != nil {
			return err
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, f.Mode()); err != nil {
				return fmt.Errorf("mkdir %s: %w", target, err)
			}
			if err := fileutil.EnsureNoSymlinkPath(target); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", filepath.Dir(target), err)
		}
		if err := fileutil.EnsureNoSymlinkPath(target); err != nil {
			return err
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

func init() {
	Register("fetch", &FetchAction{})
}
