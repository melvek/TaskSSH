package action

import (
	"strings"
	"testing"
)

func TestCommandErrorDiagnostic(t *testing.T) {
	tests := []struct {
		name  string
		err   CommandError
		want  string
		avoid string
	}{
		{
			name: "prefers stderr",
			err: CommandError{
				ExitCode: 7,
				Stdout:   "stdout detail",
				Stderr:   "stderr detail",
			},
			want:  "command failed with exit code 7:\n    stderr detail",
			avoid: "stdout detail",
		},
		{
			name: "falls back to stdout",
			err: CommandError{
				ExitCode: 3,
				Stdout:   "stdout detail",
			},
			want: "command failed with exit code 3:\n    stdout detail",
		},
		{
			name: "truncates diagnostics",
			err: CommandError{
				ExitCode: 1,
				Stderr:   strings.Repeat("x", maxCommandDiagnosticBytes+10),
			},
			want: "... (truncated)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if !strings.Contains(got, tt.want) {
				t.Fatalf("Error() = %q, want it to contain %q", got, tt.want)
			}
			if tt.avoid != "" && strings.Contains(got, tt.avoid) {
				t.Fatalf("Error() = %q, unexpectedly contains %q", got, tt.avoid)
			}
		})
	}
}
