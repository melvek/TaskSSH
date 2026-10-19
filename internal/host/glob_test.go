package host

import (
	"path"
	"testing"
)

func TestCompiledGlobMatchesPathMatch(t *testing.T) {
	patterns := []string{
		"",
		"web1",
		"web*",
		"web?",
		"web[1-3]",
		"web[^1-3]",
		"*/web*",
		"prod*/web[ab]?",
		`literal\*`,
		`literal\?`,
		`literal\[`,
		"星*",
		"?/web",
		"[z-a]",
		"[",
		"[a-]",
		"trailing\\",
	}
	names := []string{
		"",
		"web1",
		"web2",
		"web4",
		"web",
		"webserver",
		"prod-a/webab",
		"prod-b/webac",
		"prod-b/web1",
		"literal*",
		"literal?",
		"literal[",
		"星-node",
		"x/web",
		"x/y/web",
		"z",
		"z-a",
	}

	for _, pattern := range patterns {
		t.Run(pattern, func(t *testing.T) {
			compiled, compileErr := compileGlobPattern(pattern)
			for _, name := range names {
				want, pathErr := path.Match(pattern, name)
				if (compileErr != nil) != (pathErr != nil) {
					t.Fatalf("pattern %q: compile error = %v, path.Match error = %v", pattern, compileErr, pathErr)
				}
				if pathErr == nil {
					if got := compiled.Match(name); got != want {
						t.Errorf("pattern %q against %q: got %v, want %v", pattern, name, got, want)
					}
				}
			}
		})
	}
}

func BenchmarkGlobMatching(b *testing.B) {
	pattern := "app-*-node-[0-9]?.example.com"
	names := make([]string, 256)
	for i := range names {
		names[i] = "app-prod-node-12.example.com"
	}
	compiled, err := compileGlobPattern(pattern)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("path.Match", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			for _, name := range names {
				if _, err := path.Match(pattern, name); err != nil {
					b.Fatal(err)
				}
			}
		}
	})
	b.Run("compiled", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			for _, name := range names {
				compiled.Match(name)
			}
		}
	})
}
