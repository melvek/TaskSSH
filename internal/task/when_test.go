package task

import (
	"testing"

	"mestrap.com/taskssh/internal/resolve"
)

func TestEvalWhen_Empty(t *testing.T) {
	ok, _, err := evalWhen("", resolve.Vars{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("empty expression should return true")
	}
}

func TestEvalWhen_Eq(t *testing.T) {
	tests := []struct {
		name string
		expr string
		vars resolve.Vars
		want bool
	}{
		{"match", "{{env}} == prod", resolve.Vars{"env": "prod"}, true},
		{"no match", "{{env}} == prod", resolve.Vars{"env": "uat"}, false},
		{"with spaces", "{{ env }} == prod", resolve.Vars{"env": "prod"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := evalWhen(tt.expr, tt.vars)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvalWhen_Neq(t *testing.T) {
	tests := []struct {
		name string
		expr string
		vars resolve.Vars
		want bool
	}{
		{"match", "{{env}} != prod", resolve.Vars{"env": "uat"}, true},
		{"no match", "{{env}} != prod", resolve.Vars{"env": "prod"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := evalWhen(tt.expr, tt.vars)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvalWhen_NumericCompare(t *testing.T) {
	vars := resolve.Vars{"version": "5"}

	tests := []struct {
		name string
		expr string
		want bool
	}{
		{"gt true", "{{version}} > 3", true},
		{"gt false", "{{version}} > 10", false},
		{"ge true", "{{version}} >= 5", true},
		{"ge false", "{{version}} >= 6", false},
		{"lt true", "{{version}} < 10", true},
		{"lt false", "{{version}} < 3", false},
		{"le true", "{{version}} <= 5", true},
		{"le false", "{{version}} <= 4", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := evalWhen(tt.expr, vars)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvalWhen_StringCompare(t *testing.T) {
	vars := resolve.Vars{"name": "banana"}

	tests := []struct {
		name string
		expr string
		want bool
	}{
		{"gt", "{{name}} > apple", true},
		{"lt", "{{name}} < cherry", true},
		{"ge", "{{name}} >= banana", true},
		{"le", "{{name}} <= banana", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := evalWhen(tt.expr, vars)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvalWhen_In(t *testing.T) {
	tests := []struct {
		name string
		expr string
		vars resolve.Vars
		want bool
	}{
		{
			name: "in with parens",
			expr: "{{role}} in (web,db,cache)",
			vars: resolve.Vars{"role": "db"},
			want: true,
		},
		{
			name: "in without parens",
			expr: "{{role}} in web,db,cache",
			vars: resolve.Vars{"role": "db"},
			want: true,
		},
		{
			name: "not in",
			expr: "{{role}} in (web,db)",
			vars: resolve.Vars{"role": "cache"},
			want: false,
		},
		{
			name: "with spaces",
			expr: "{{role}} in ( web , db )",
			vars: resolve.Vars{"role": "db"},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := evalWhen(tt.expr, tt.vars)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvalWhen_And(t *testing.T) {
	vars := resolve.Vars{"env": "prod", "role": "web"}

	tests := []struct {
		expr string
		want bool
	}{
		{"{{env}} == prod && {{role}} == web", true},
		{"{{env}} == prod && {{role}} == db", false},
		{"{{env}} == uat && {{role}} == web", false},
	}
	for _, tt := range tests {
		got, _, err := evalWhen(tt.expr, vars)
		if err != nil {
			t.Fatalf("expr %q: %v", tt.expr, err)
		}
		if got != tt.want {
			t.Errorf("expr %q: got %v, want %v", tt.expr, got, tt.want)
		}
	}
}

func TestEvalWhen_Or(t *testing.T) {
	vars := resolve.Vars{"env": "uat"}

	tests := []struct {
		expr string
		want bool
	}{
		{"{{env}} == prod || {{env}} == uat", true},
		{"{{env}} == prod || {{env}} == dev", false},
	}
	for _, tt := range tests {
		got, _, err := evalWhen(tt.expr, vars)
		if err != nil {
			t.Fatalf("expr %q: %v", tt.expr, err)
		}
		if got != tt.want {
			t.Errorf("expr %q: got %v, want %v", tt.expr, got, tt.want)
		}
	}
}

func TestEvalWhen_Mixed(t *testing.T) {
	vars := resolve.Vars{"env": "prod", "role": "db"}

	// 优先级：&& 高于 ||，等价于 (env==prod && role==web) || role==db
	got, _, err := evalWhen("{{env}} == prod && {{role}} == web || {{role}} == db", vars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Error("expected true")
	}
}

func TestEvalWhen_Unresolved(t *testing.T) {
	_, _, err := evalWhen("{{unknown}} == prod", resolve.Vars{})
	if err == nil {
		t.Fatal("expected error for unresolved var, got nil")
	}
}

func TestEvalWhen_Unsupported(t *testing.T) {
	_, _, err := evalWhen("{{env}}", resolve.Vars{"env": "prod"})
	if err == nil {
		t.Fatal("expected error for unsupported expression, got nil")
	}
}

func TestEvalWhen_Complex(t *testing.T) {
	vars := resolve.Vars{
		"env":     "prod",
		"role":    "web",
		"version": "3",
	}

	tests := []struct {
		name string
		expr string
		want bool
	}{
		{
			name: "env and version",
			expr: "{{env}} == prod && {{version}} >= 2",
			want: true,
		},
		{
			name: "role in list and version",
			expr: "{{role}} in (web,db) && {{version}} > 5",
			want: false,
		},
		{
			name: "or with and",
			expr: "{{env}} == uat || {{role}} == web && {{version}} == 3",
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := evalWhen(tt.expr, vars)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
