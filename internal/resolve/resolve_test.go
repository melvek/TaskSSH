package resolve

import (
	"strings"
	"testing"
)

func TestReplace_Simple(t *testing.T) {
	vars := Vars{
		"app_name": "myapp",
		"version":  "1.0.0",
	}

	got, err := vars.Replace("{{app_name}}-{{version}}.jar")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "myapp-1.0.0.jar"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestReplace_WithSpaces(t *testing.T) {
	vars := Vars{"app_name": "myapp"}

	got, err := vars.Replace("{{ app_name }}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "myapp" {
		t.Errorf("got %q, want %q", got, "myapp")
	}
}

func TestReplace_Recursive(t *testing.T) {
	vars := Vars{
		"app_name": "myapp",
		"version":  "1.0.0",
		"package":  "{{app_name}}-{{version}}.jar",
	}

	got, err := vars.Replace("{{package}}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "myapp-1.0.0.jar" {
		t.Errorf("got %q, want %q", got, "myapp-1.0.0.jar")
	}
}

func TestReplace_Unresolved(t *testing.T) {
	vars := Vars{}

	_, err := vars.Replace("{{unknown}}")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "unknown") {
		t.Errorf("error should mention 'unknown', got %q", err.Error())
	}
}

func TestReplace_ShellVarUnchanged(t *testing.T) {
	vars := Vars{"app_name": "myapp"}

	// shell 变量 $HOME 和 ${HOME} 不应被处理
	got, err := vars.Replace("echo $HOME ${HOME} {{app_name}}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "echo $HOME ${HOME} myapp"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestReplace_Empty(t *testing.T) {
	vars := Vars{}

	got, err := vars.Replace("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestReplace_NoPlaceholder(t *testing.T) {
	vars := Vars{"app_name": "myapp"}

	got, err := vars.Replace("plain text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "plain text" {
		t.Errorf("got %q, want %q", got, "plain text")
	}
}
