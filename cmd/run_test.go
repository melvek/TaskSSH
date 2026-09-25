package cmd

import "testing"

func TestParseCLIVars(t *testing.T) {
	tests := []struct {
		name    string
		in      []string
		want    map[string]string
		wantErr bool
	}{
		{
			name: "simple",
			in:   []string{"a=1", "b=2"},
			want: map[string]string{"a": "1", "b": "2"},
		},
		{
			name: "trim key and value",
			in:   []string{" a = 1 "},
			want: map[string]string{"a": "1"},
		},
		{
			name: "value contains equals",
			in:   []string{"msg=a=b=c"},
			want: map[string]string{"msg": "a=b=c"},
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
			name:    "reserved execId",
			in:      []string{"execId=AAAA"},
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
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("got[%q] = %q, want %q", k, got[k], v)
				}
			}
		})
	}
}
