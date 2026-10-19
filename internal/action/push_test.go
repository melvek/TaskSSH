package action

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestZipDirRejectsSymbolicLinks(t *testing.T) {
	dir := t.TempDir()
	outsideFile := filepath.Join(t.TempDir(), "outside.txt")
	if err := os.WriteFile(outsideFile, []byte("outside"), 0600); err != nil {
		t.Fatal(err)
	}

	linkPath := filepath.Join(dir, "outside.txt")
	if err := os.Symlink(outsideFile, linkPath); err != nil {
		t.Skipf("cannot create symbolic links: %v", err)
	}

	zipPath, err := zipDir(dir)
	if err == nil {
		os.Remove(zipPath)
		t.Fatal("zipDir() accepted a symbolic link")
	}
	if !strings.Contains(err.Error(), "symbolic link") {
		t.Fatalf("zipDir() error = %v, want symbolic-link error", err)
	}
}

func TestRejectSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	link := filepath.Join(dir, "link")
	if err := os.WriteFile(target, nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("cannot create symbolic links: %v", err)
	}

	info, err := os.Lstat(link)
	if err != nil {
		t.Fatal(err)
	}
	if err := rejectSymlink(link, info); err == nil {
		t.Fatal("rejectSymlink() accepted a symbolic link")
	}
}
