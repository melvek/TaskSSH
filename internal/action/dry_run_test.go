package action

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mestrap.com/taskssh/internal/console"
	"mestrap.com/taskssh/internal/resolve"
)

func TestFetchDryRunDoesNotInspectRemote(t *testing.T) {
	var output strings.Builder
	logger := console.NewLogger(&output)

	ctx := &ActionContext{
		With: map[string]any{
			"file": "/remote/data",
			"dest": "./downloads/",
		},
		Vars:   resolve.Vars{},
		DryRun: true,
		Logger: logger,
	}
	if err := (&FetchAction{}).Execute(ctx); err != nil {
		t.Fatalf("FetchAction.Execute() error = %v", err)
	}
	if got := output.String(); !strings.Contains(got, "/remote/data -> ./downloads/") {
		t.Fatalf("dry-run output %q does not show source and destination", got)
	}
}

func TestPushZipDryRunDoesNotCreateArchiveOrUseClient(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("content"), 0600); err != nil {
		t.Fatal(err)
	}

	var output strings.Builder
	ctx := &ActionContext{
		With: map[string]any{
			"file": dir,
			"dest": "/remote/dest",
			"zip":  true,
		},
		Vars:   resolve.Vars{},
		DryRun: true,
		Logger: console.NewLogger(&output),
	}
	if err := (&PushAction{}).Execute(ctx); err != nil {
		t.Fatalf("PushAction.Execute() error = %v", err)
	}
	if got := output.String(); !strings.Contains(got, dir+" -> /remote/dest") {
		t.Fatalf("dry-run output %q does not show source and destination", got)
	}
}
