package ssh

import (
	"path/filepath"
	"testing"

	"mestrap.com/taskssh/internal/config"
)

func TestBuildAuthMethodsUsesOnlyCLIAuthOverrides(t *testing.T) {
	methods, err := buildAuthMethods(&config.Host{
		Host:         "example.com",
		Password:     "cli-password",
		AuthOverride: true,
	})
	if err != nil {
		t.Fatalf("buildAuthMethods() error = %v", err)
	}
	if len(methods) != 2 {
		t.Fatalf("buildAuthMethods() returned %d methods, want 2 password methods", len(methods))
	}
}

func TestBuildAuthMethodsDoesNotFallbackForCLIIdentityFile(t *testing.T) {
	identityFile := filepath.Join(t.TempDir(), "missing-key")
	_, err := buildAuthMethods(&config.Host{
		Host:         "example.com",
		IdentityFile: identityFile,
		AuthOverride: true,
	})
	if err == nil {
		t.Fatal("buildAuthMethods() succeeded with a missing explicitly selected identity file")
	}
}
