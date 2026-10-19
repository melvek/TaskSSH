package action

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createZip(t *testing.T, name string, mode os.FileMode, contents string) string {
	t.Helper()

	zipPath := filepath.Join(t.TempDir(), "archive.zip")
	file, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}

	writer := zip.NewWriter(file)
	header := &zip.FileHeader{Name: name}
	header.SetMode(mode)
	entry, err := writer.CreateHeader(header)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte(contents)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	return zipPath
}

func TestExtractZipRejectsExistingSymlinkAncestor(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "dest")
	if err := os.MkdirAll(dest, 0755); err != nil {
		t.Fatal(err)
	}
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(dest, "link")); err != nil {
		t.Skipf("cannot create symbolic links: %v", err)
	}

	zipPath := createZip(t, "link/file.txt", 0644, "outside")
	if err := extractZip(zipPath, dest); err == nil {
		t.Fatal("extractZip() wrote through an existing symbolic link")
	}
	if _, err := os.Stat(filepath.Join(outside, "file.txt")); !os.IsNotExist(err) {
		t.Fatalf("outside file unexpectedly created, stat error = %v", err)
	}
}

func TestExtractZipRejectsTraversal(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "dest")
	zipPath := createZip(t, ".."+string(filepath.Separator)+"escaped.txt", 0644, "outside")

	if err := extractZip(zipPath, dest); err == nil || !strings.Contains(err.Error(), "invalid zip entry") {
		t.Fatalf("extractZip() error = %v, want invalid-entry error", err)
	}
}

func TestExtractZipRejectsSymlinkEntry(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "dest")
	zipPath := createZip(t, "link", os.ModeSymlink|0777, "../outside")

	if err := extractZip(zipPath, dest); err == nil || !strings.Contains(err.Error(), "symbolic link") {
		t.Fatalf("extractZip() error = %v, want symbolic-link error", err)
	}
}
