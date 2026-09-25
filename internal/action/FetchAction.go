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

	"mestrap.com/taskssh/internal/log"
)

// FetchAction 从远程下载文件或目录。
type FetchAction struct{}

func (a *FetchAction) Name() string { return "fetch" }

func (a *FetchAction) Execute(ctx *Context) error {
	// 1. 解析 file
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

	// 2. 解析 dest，默认 ./
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

	force := toBool(ctx.With["force"])
	backup := toBool(ctx.With["backup"])
	useZip := toBool(ctx.With["zip"])

	// 3. 判断远程路径类型
	info, err := ctx.Client.Stat(remote)
	if err != nil {
		return err
	}

	if info.IsDir() {
		if useZip {
			return a.fetchDirZip(ctx, remote, dest, force, backup)
		}
		return a.fetchDirRecursive(ctx, remote, dest, force, backup)
	}

	// 4. 单文件
	return a.fetchFile(ctx, remote, dest, force, backup)
}

// fetchFile 下载单个文件。
func (a *FetchAction) fetchFile(ctx *Context, remote, dest string, force, backup bool) error {
	finalPath, err := resolveFinalPath(dest, remote, ctx.Host.Host)
	if err != nil {
		return err
	}

	if err := checkOverwrite(finalPath, force, backup); err != nil {
		return err
	}

	log.Info("Download %s to %s", remote, finalPath)

	if err := ctx.Client.Download(remote, finalPath); err != nil {
		return err
	}

	log.Success("Downloaded: %s -> %s", remote, finalPath)
	return nil
}

// fetchDirRecursive 递归下载目录（不传 -z 时）。
func (a *FetchAction) fetchDirRecursive(ctx *Context, remote, dest string, force, backup bool) error {
	log.Info("Download directory %s to %s (recursive)", remote, dest)

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

		if err := checkOverwrite(localPath, force, backup); err != nil {
			return err
		}

		if err := os.MkdirAll(filepath.Dir(localPath), 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", filepath.Dir(localPath), err)
		}

		log.Info("Download %s to %s", rf, localPath)

		if err := ctx.Client.Download(rf, localPath); err != nil {
			return err
		}
	}

	log.Success("Directory downloaded: %s -> %s (%d files)", remote, rootLocal, len(files))
	return nil
}

// fetchDirZip 远程打包下载目录（传 -z 时）。
//
// 流程：
//  1. 远程临时目录：<tmpDir>/taskssh-fetch-<host>
//  2. 远程 zip -r 打包
//  3. 下载 zip
//  4. 删除远程 zip
//  5. 本地 archive/zip 解压
func (a *FetchAction) fetchDirZip(ctx *Context, remote, dest string, force, backup bool) error {
	baseName := path.Base(remote)
	remoteDir := path.Dir(remote)

	// 解析 tmp_dir，默认 /tmp
	tmpDir := "/tmp"
	if tmpRaw, ok := ctx.With["tmp_dir"]; ok {
		s := fmt.Sprintf("%v", tmpRaw)
		if s != "" {
			v, err := ctx.Vars.Replace(s)
			if err != nil {
				return err
			}
			tmpDir = v
		}
	}

	// 远程临时子目录：<tmpDir>/taskssh-fetch-<host>
	hostID := buildHostID(ctx.Host.Host)
	remoteTmpDir := path.Join(tmpDir, "taskssh-fetch-"+hostID)

	// 远程 zip 完整路径
	zipName := fmt.Sprintf("%s.%d.zip", baseName, time.Now().UnixMicro())
	remoteZip := path.Join(remoteTmpDir, zipName)

	log.Info("Zip directory %s on remote", remote)

	// 远程建临时目录并打包
	zipCmd := fmt.Sprintf(
		`which zip &>/dev/null && { mkdir -p %s && cd %s && zip -r %s %s; } || `+
			`{ echo "need install 'zip' command"; exit 1; }`,
		shellQuote(remoteTmpDir),
		shellQuote(remoteDir),
		shellQuote(remoteZip),
		shellQuote(baseName),
	)
	log.Info("Execute zip: %s", zipCmd)

	result, err := ctx.Client.Exec(zipCmd)
	if err != nil {
		return fmt.Errorf("zip: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("zip failed with exit code %d", result.ExitCode)
	}

	// 下载 zip 到本地临时文件
	localZip, err := os.CreateTemp("", "taskssh-fetch-*.zip")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	localZipPath := localZip.Name()
	localZip.Close()
	defer os.Remove(localZipPath)

	log.Info("Download zip %s to %s", remoteZip, localZipPath)

	if err := ctx.Client.Download(remoteZip, localZipPath); err != nil {
		return err
	}

	// 删除远程 zip
	rmCmd := fmt.Sprintf("rm -f %s", shellQuote(remoteZip))
	if _, err := ctx.Client.Exec(rmCmd); err != nil {
		log.Warn("Failed to remove remote zip: %v", err)
	}

	// 本地解压：dest/<主机标识>/<目录名>
	rootLocal := filepath.Join(dest, hostID, baseName)

	log.Info("Extract %s to %s", localZipPath, rootLocal)

	if err := extractZip(localZipPath, rootLocal); err != nil {
		return fmt.Errorf("extract: %w", err)
	}

	log.Success("Directory downloaded: %s -> %s", remote, rootLocal)
	return nil
}

// resolveFinalPath 计算单文件下载的最终本地路径。
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

// checkOverwrite 检查本地文件是否已存在，按 force / backup 处理。
func checkOverwrite(localPath string, force, backup bool) error {
	info, err := os.Stat(localPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if info.IsDir() {
		return nil
	}

	if !force {
		return fmt.Errorf("local file already exists: %s (use -F to overwrite, -F -B to backup)",
			localPath)
	}

	if backup {
		backupPath := fmt.Sprintf("%s.%s", localPath, time.Now().Format("20060102150405"))
		if err := os.Rename(localPath, backupPath); err != nil {
			return fmt.Errorf("backup %s: %w", localPath, err)
		}
		log.Info("Backed up: %s -> %s", localPath, backupPath)
	}

	return nil
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

// extractZip 解压 zip 到目标目录。
//
// 防止路径穿越：每个 entry 的目标路径必须落在 destDir 内。
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

		// 路径穿越检查
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
