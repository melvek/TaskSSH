package cmd

import (
	"testing"

	"mestrap.com/taskssh/internal/resolve"
)

func TestParseCLIVars(t *testing.T) {
	tests := []struct {
		name    string
		in      []string
		want    resolve.Vars
		wantErr bool
	}{
		{
			name: "simple",
			in:   []string{"a=1", "b=2"},
			want: resolve.Vars{"a": "1", "b": "2"},
		},
		{
			name: "trim key and value",
			in:   []string{" a = 1 "},
			want: resolve.Vars{"a": "1"},
		},
		{
			name: "value contains equals",
			in:   []string{"msg=a=b=c"},
			want: resolve.Vars{"msg": "a=b=c"},
		},
		{
			name:    "no equals",
			in:      []string{"abc"},
			wantErr: true,
		},
		{
			name:    "empty key",
			in:      []string{"=1"},
			wantErr: true,
		},
		{
			name:    "reserved date",
			in:      []string{"date=20260101"},
			wantErr: true,
		},
		{
			name:    "reserved exec_id",
			in:      []string{"exec_id=AAAA"},
			wantErr: true,
		},
		{
			name:    "reserved host_name",
			in:      []string{"host_name=foo"},
			wantErr: true,
		},
		{
			name:    "reserved group_name",
			in:      []string{"group_name=prod"},
			wantErr: true,
		},
		{
			name:    "reserved host_addr",
			in:      []string{"host_addr=1.2.3.4"},
			wantErr: true,
		},
		{
			name:    "reserved host_port",
			in:      []string{"host_port=22"},
			wantErr: true,
		},
		{
			name:    "reserved host_user",
			in:      []string{"host_user=root"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCLIVars(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %d vars, want %d", len(got), len(tt.want))
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("got[%q] = %v, want %v", k, got[k], v)
				}
			}
		})
	}
}
