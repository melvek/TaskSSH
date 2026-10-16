package host

import (
	"path"
	"strings"
	"testing"
	"time"
)

func matchTimed(pattern, name string, d time.Duration) (bool, error, time.Duration, bool) {
	type res struct {
		ok  bool
		err error
	}
	ch := make(chan res, 1)
	start := time.Now()
	go func() {
		ok, err := path.Match(pattern, name)
		ch <- res{ok, err}
	}()
	select {
	case r := <-ch:
		return r.ok, r.err, time.Since(start), true
	case <-time.After(d):
		return false, nil, time.Since(start), false
	}
}

// 探针一：path.Match 对非法模式是否「与候选串无关」地返回 ErrBadPattern。
// 这决定能不能把语法校验从热循环里挪出去。
func TestPathMatchBadPatternDetection(t *testing.T) {
	pats := []string{"[", "[a", "a[b", "[]", "[^", "[a-", "*[", "*", "?", "a/b", "*/*", "[a-z]", "prod-?/*"}
	for _, p := range pats {
		_, eEmpty, _, _ := matchTimed(p, "", 2*time.Second)
		_, eX, _, _ := matchTimed(p, "x", 2*time.Second)
		_, eLong, _, _ := matchTimed(p, strings.Repeat("xyz", 40), 2*time.Second)
		t.Logf("pattern=%-12q err(empty)=%-28v err(x)=%-28v err(long)=%v", p, eEmpty, eX, eLong)
	}
}

// 探针二：多星号 glob 的匹配耗时是否随规模爆炸（ReDoS 风险）。
func TestGlobComplexity(t *testing.T) {
	for n := 4; n <= 24; n += 4 {
		pat := strings.Repeat("*a", n) + "b"
		name := strings.Repeat("a", 2*n)
		ok, err, el, done := matchTimed(pat, name, 5*time.Second)
		t.Logf("n=%-3d lenPat=%-4d lenName=%-4d match=%-5v done=%-5v err=%-20v elapsed=%v",
			n, len(pat), len(name), ok, done, err, el)
		if !done || el > 2*time.Second {
			t.Logf("==> 出现超线性耗时，后续跳过")
			break
		}
	}
}

// 探针三：放大规模，确认多星号 / 字符类 glob 不会出现超线性回溯。
func TestGlobComplexityLarge(t *testing.T) {
	shapes := []struct {
		name  string
		build func(n int) (string, string)
	}{
		{"星号+字母", func(n int) (string, string) { return strings.Repeat("*a", n) + "b", strings.Repeat("a", 2*n) }},
		{"字母+星号", func(n int) (string, string) { return strings.Repeat("a*", n) + "b", strings.Repeat("a", 2*n) }},
		{"纯星号", func(n int) (string, string) { return strings.Repeat("*", n), strings.Repeat("a", n) }},
		{"字符类", func(n int) (string, string) { return strings.Repeat("[a-z]*", n) + "b", strings.Repeat("a", 2*n) }},
	}
	for _, s := range shapes {
		for _, n := range []int{20, 60, 120, 240} {
			pat, name := s.build(n)
			ok, err, el, done := matchTimed(pat, name, 5*time.Second)
			t.Logf("%-9s n=%-4d lenPat=%-5d lenName=%-5d match=%-5v done=%-5v err=%-18v elapsed=%v",
				s.name, n, len(pat), len(name), ok, done, err, el)
			if !done {
				t.Errorf("%s n=%d 超过 5s 未返回，存在复杂度风险", s.name, n)
				break
			}
		}
	}
}

// Test_Standalone_SingleCharLabel_NoPanic 单字符 label 曾导致切片越界 panic。
func Test_Standalone_SingleCharLabel_NoPanic(t *testing.T) {
	for _, name := range []string{"a", "1", "x", "a.b", "a1", "9.z"} {
		got, err := resolveNames(name)
		if err != nil {
			t.Errorf("resolve(%q) 意外报错: %v", name, err)
			continue
		}
		if len(got) != 1 || got[0] != name {
			t.Errorf("resolve(%q) = %v, 期望独立主机 [%s]", name, got, name)
		}
	}
}

// Test_Standalone_Validate 独立主机名/域名/IP 校验。
func Test_Standalone_Validate(t *testing.T) {
	valid := []string{
		"web1", "db-01", "api.service.consul", "k8s_pod-1",
		"a", "1", "1234", "web1.prod.example.com",
		"192.168.1.1", "10.0.0.1", "127.0.0.1",
		"::1", "fe80::1", "fe80::1%eth0", "[::1]", "[fe80::1%eth0]", "2001:db8::1",
	}
	for _, name := range valid {
		if _, err := resolveNames(name); err != nil {
			t.Errorf("合法独立主机 %q 被拒: %v", name, err)
		}
	}

	invalid := []string{
		"", "db-", "-db", "a..b", ".a", "a.",
		"192.168.1.999", "1::2::3", "[::1", "fe80::1%",
		// 注意："a b" 不在此列——空格会被 splitPattern 拆成两个 token
		// （a 与 b），二者都是合法短名，属于预期的多模式写法。
		"foo;id", "foo*", "foo?bar", "a`b", "a$b", "a&b", "a|b",
		"café", "a@b", "a:b", "a/b", "a=b",
		strings.Repeat("a", 64),
	}
	for _, name := range invalid {
		if name == "" {
			continue // 空 pattern 语义为全选，另行约定
		}
		got, err := resolveNames(name)
		if err == nil {
			t.Errorf("非法独立主机 %q 未被拒，返回 %v", name, got)
		}
	}
}

// Test_Standalone_LengthBoundary 总长度边界。
func Test_Standalone_LengthBoundary(t *testing.T) {
	// 每段 63：63+1+63+1+63+1+59 = 251，合法
	okName := strings.Repeat("a", 63) + "." + strings.Repeat("b", 63) + "." +
		strings.Repeat("c", 63) + "." + strings.Repeat("d", 59)
	if len(okName) != 251 {
		t.Fatalf("用例构造错误: len=%d", len(okName))
	}
	if _, err := resolveNames(okName); err != nil {
		t.Errorf("长度 251 的合法域名被拒: %v", err)
	}

	// 254 长度超限
	badName := strings.Repeat("a", 63) + "." + strings.Repeat("b", 63) + "." +
		strings.Repeat("c", 63) + "." + strings.Repeat("d", 62)
	if len(badName) != 254 {
		t.Fatalf("用例构造错误: len=%d", len(badName))
	}
	if _, err := resolveNames(badName); err == nil {
		t.Errorf("长度 254 的域名未被拒")
	}
}
