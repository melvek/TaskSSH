package fileutil

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// EnsureNoSymlinkPath rejects paths that contain an existing symbolic link.
func EnsureNoSymlinkPath(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve path %s: %w", path, err)
	}

	for current := abs; ; current = filepath.Dir(current) {
		info, err := os.Lstat(current)
		if err == nil {
			if info.Mode()&os.ModeSymlink != 0 {
				return fmt.Errorf("path contains symbolic link: %s", current)
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("inspect path %s: %w", current, err)
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
	}
	return nil
}

func ExpandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

func FileExists(path string) bool {
	info, err := os.Stat(ExpandHome(path))
	return err == nil && !info.IsDir()
}
