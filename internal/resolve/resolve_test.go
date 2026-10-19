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

func TestReplace_EmptyName(t *testing.T) {
	vars := Vars{}

	_, err := vars.Replace("{{ }}")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got, want := err.Error(), "unresolved variables: "; got != want {
		t.Errorf("error = %q, want %q", got, want)
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

func TestReplace_ScalarValues(t *testing.T) {
	vars := Vars{
		"port":    3306,
		"enabled": true,
	}

	got, err := vars.Replace("--port={{port}} --enabled={{enabled}}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := "--port=3306 --enabled=true"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestReplace_RejectsNonScalarValues(t *testing.T) {
	tests := []struct {
		name  string
		value any
	}{
		{name: "list", value: []string{"one", "two"}},
		{name: "map", value: map[string]string{"key": "value"}},
		{name: "null", value: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Vars{"value": tt.value}.Replace("{{value}}")
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), `variable "value" has unsupported value type`) {
				t.Errorf("error = %q, want unsupported-value error", err)
			}
		})
	}
}

func TestReplace_MaxDepth(t *testing.T) {
	vars := Vars{
		"v1": "done",
		"v2": "{{v1}}",
		"v3": "{{v2}}",
		"v4": "{{v3}}",
		"v5": "{{v4}}",
	}

	got, err := vars.Replace("{{v5}}")
	if err != nil {
		t.Fatalf("unexpected error at maximum depth: %v", err)
	}
	if got != "done" {
		t.Errorf("got %q, want %q", got, "done")
	}

	vars["v6"] = "{{v5}}"
	if _, err := vars.Replace("{{v6}}"); err == nil {
		t.Fatal("expected error beyond maximum depth, got nil")
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

func TestReplace_ShellEscape(t *testing.T) {
	vars := Vars{"app_name": "myapp'Name"}

	// shell 变量 $HOME 和 ${HOME} 不应被处理
	got, err := vars.Replace("echo {{app_name}}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "echo myapp'Name"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
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
