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

func TestSetTargetHostPattern(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantPattern string
		wantErr     bool
	}{
		{
			name:        "positional host",
			args:        []string{"prod"},
			wantPattern: "prod",
		},
		{
			name:        "multiple hosts in one pattern",
			args:        []string{"prod_server,web_server"},
			wantPattern: "prod_server,web_server",
		},
		{
			name:        "space-separated hosts in one quoted pattern",
			args:        []string{"prod_server web_server"},
			wantPattern: "prod_server web_server",
		},
		{
			name:    "multiple positional arguments",
			args:    []string{"prod_server", "web_server"},
			wantErr: true,
		},
		{
			name:    "empty host",
			args:    []string{" "},
			wantErr: true,
		},
		{
			name:    "missing pattern",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := setTargetHostPattern(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if targetHostPattern != tt.wantPattern {
				t.Errorf("targetHostPattern = %q, want %q", targetHostPattern, tt.wantPattern)
			}
		})
	}
}

func TestBuiltinActionArgumentsRequireFlags(t *testing.T) {
	previousPattern := targetHostPattern
	previousExecute := executeFlag
	previousPushFile, previousPushDest := pushFile, pushDest
	previousFetchFile, previousScriptFile := fetchFile, scriptFile
	t.Cleanup(func() {
		targetHostPattern = previousPattern
		executeFlag = previousExecute
		pushFile, pushDest = previousPushFile, previousPushDest
		fetchFile, scriptFile = previousFetchFile, previousScriptFile
	})

	tests := []struct {
		name string
		run  func() error
		want string
	}{
		{
			name: "command requires execute flag",
			run: func() error {
				return commandCmd.RunE(commandCmd, []string{"prod"})
			},
			want: "command requires -e/--execute",
		},
		{
			name: "push requires file flag",
			run: func() error {
				return pushCmd.RunE(pushCmd, []string{"prod"})
			},
			want: "push requires -f/--file",
		},
		{
			name: "push requires destination flag",
			run: func() error {
				pushFile = "app.jar"
				return pushCmd.RunE(pushCmd, []string{"prod"})
			},
			want: "push requires -d/--dest",
		},
		{
			name: "fetch requires file flag",
			run: func() error {
				return fetchCmd.RunE(fetchCmd, []string{"prod"})
			},
			want: "fetch requires -f/--file",
		},
		{
			name: "script requires file flag",
			run: func() error {
				return scriptCmd.RunE(scriptCmd, []string{"prod"})
			},
			want: "script requires -f/--file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targetHostPattern = ""
			executeFlag = ""
			pushFile, pushDest = "", ""
			fetchFile = ""
			scriptFile = ""

			err := tt.run()
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if err.Error() != tt.want {
				t.Fatalf("error = %q, want %q", err, tt.want)
			}
		})
	}

	if err := pushCmd.Args(pushCmd, []string{"prod", "web_server"}); err == nil {
		t.Fatal("expected multiple host arguments to be rejected")
	}
	if err := commandCmd.Args(commandCmd, []string{"prod", "web_server"}); err == nil {
		t.Fatal("expected multiple host arguments to be rejected")
	}
	if err := pushCmd.Args(pushCmd, nil); err == nil {
		t.Fatal("expected missing host pattern to be rejected")
	}
	if err := commandCmd.Args(commandCmd, nil); err == nil {
		t.Fatal("expected missing host pattern to be rejected")
	}
}

func TestHostFlagRemoved(t *testing.T) {
	if rootCmd.PersistentFlags().Lookup("host") != nil {
		t.Fatal("unexpected --host flag")
	}
	if rootCmd.PersistentFlags().ShorthandLookup("H") != nil {
		t.Fatal("unexpected -H shorthand")
	}
}
