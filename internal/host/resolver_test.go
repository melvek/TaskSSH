package host

import (
	"strings"
	"testing"

	"mestrap.com/taskssh/internal/config"
)

func testInv() *config.Inventory {
	return &config.Inventory{
		Servers: map[string]config.Group{
			"db":     {Hosts: map[string]config.Host{"api": {Host: "10.0.0.1"}, "worker": {Host: "10.0.0.2"}}},
			"prod-a": {Hosts: map[string]config.Host{"api": {Host: "10.1.0.1"}, "web": {Host: "10.1.0.2"}}},
			"prod-b": {Hosts: map[string]config.Host{"api": {Host: "10.2.0.1"}, "cache": {Host: "10.2.0.2"}}},
		},
	}
}

func namesOf(entries []HostEntry) []string {
	out := make([]string, len(entries))
	for i, e := range entries {
		out[i] = e.Name
	}
	return out
}

func resolveNames(pattern string) ([]string, error) {
	entries, err := newResolver(testInv()).resolve(pattern, Overrides{})
	if err != nil {
		return nil, err
	}
	return namesOf(entries), nil
}

func mustNames(t *testing.T, pattern string) []string {
	t.Helper()
	got, err := resolveNames(pattern)
	if err != nil {
		t.Fatalf("resolve(%q) 报错: %v", pattern, err)
	}
	return got
}

func eqNames(t *testing.T, pattern string, got, want []string) {
	t.Helper()
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("pattern %q: got=%v want=%v", pattern, got, want)
	}
}

// 基线：正常写法下的结果（优化版必须与之逐字一致）。
func Test_Baseline(t *testing.T) {
	cases := []struct {
		pattern string
		want    []string
	}{
		{"db", []string{"db/api", "db/worker"}},
		{"api", []string{"db/api", "prod-a/api", "prod-b/api"}},
		{"g:db", []string{"db/api", "db/worker"}},
		{"h:api", []string{"db/api", "prod-a/api", "prod-b/api"}},
		{"g:*", []string{"db/api", "db/worker", "prod-a/api", "prod-a/web", "prod-b/api", "prod-b/cache"}},
		{"db/api", []string{"db/api"}},
		{"web*", []string{"prod-a/web"}},
		{"db/*", []string{"db/api", "db/worker"}},
		{"*/api", []string{"db/api", "prod-a/api", "prod-b/api"}},
		{"*/*", []string{"db/api", "db/worker", "prod-a/api", "prod-a/web", "prod-b/api", "prod-b/cache"}},
		{"prod*/*", []string{"prod-a/api", "prod-a/web", "prod-b/api", "prod-b/cache"}},
		{"g:* !h:api", []string{"db/worker", "prod-a/web", "prod-b/cache"}},
		{"g:db &h:api", []string{"db/api"}},
		{"newhost", []string{"newhost"}},
	}
	for _, c := range cases {
		eqNames(t, c.pattern, mustNames(t, c.pattern), c.want)
	}
}

// 已有防护：非法 glob 会被报出来（不会被当成「匹配不到」）。
func Test_InvalidGlobIsReported(t *testing.T) {
	for _, p := range []string{"g:[", "h:[", "["} {
		_, err := resolveNames(p)
		t.Logf("resolve(%-6q) err = %v", p, err)
		if err == nil {
			t.Fatalf("pattern %q 期望报错", p)
		}
	}
}

func Test_SeparatorOnlyPatternIsRejected(t *testing.T) {
	for _, pattern := range []string{" ", ",", " , , "} {
		_, err := resolveNames(pattern)
		if err == nil {
			t.Errorf("resolve(%q) succeeded; want empty-pattern error", pattern)
		}
	}
}
