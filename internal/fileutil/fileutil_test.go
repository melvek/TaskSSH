package fileutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureNoSymlinkPathRejectsSymlinkAncestor(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	link := filepath.Join(root, "link")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("cannot create symbolic links: %v", err)
	}

	if err := EnsureNoSymlinkPath(filepath.Join(link, "nested", "file")); err == nil {
		t.Fatal("EnsureNoSymlinkPath() accepted a path under a symbolic link")
	}
}

func TestEnsureNoSymlinkPathAllowsRegularPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "file")
	if err := EnsureNoSymlinkPath(path); err != nil {
		t.Fatalf("EnsureNoSymlinkPath() error = %v", err)
	}
}
