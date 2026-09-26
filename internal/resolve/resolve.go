package resolve

import (
	"fmt"
	"regexp"
	"strings"
)

// Vars 是变量池。
type Vars map[string]any

// placeholder 匹配 {{ var }}，允许花括号内前后有空格。
var placeholder = regexp.MustCompile(`\{\{\s*([^{}]+?)\s*\}\}`)

// MaxDepth 是递归展开的最大层数。
const MaxDepth = 4

// Replace 对模板做变量替换。
//
// 递归展开最多 MaxDepth 层。
// 未解析的变量返回错误。
func (v Vars) Replace(template string) (string, error) {
	if template == "" {
		return template, nil
	}
	current := template
	for i := 0; i < MaxDepth; i++ {
		next := v.replaceOnce(current)
		if next == current {
			break
		}
		current = next
	}
	missing := findMissing(current)
	if len(missing) > 0 {
		return "", fmt.Errorf("unresolved: %s", strings.Join(missing, ", "))
	}
	return current, nil
}

// replaceOnce 执行一次替换。
//
// 未定义的变量保持原样，留给后续递归或最终报错。
func (v Vars) replaceOnce(s string) string {
	return placeholder.ReplaceAllStringFunc(s, func(match string) string {
		key := extractKey(match)
		if val, ok := v[key]; ok {
			return fmt.Sprintf("%v", val)
		}
		return match
	})
}

// findMissing 找出未解析的变量名。
func findMissing(s string) []string {
	var keys []string
	for _, m := range placeholder.FindAllStringSubmatch(s, -1) {
		keys = append(keys, strings.TrimSpace(m[1]))
	}
	return keys
}

// extractKey 从 {{ var }} 中提取变量名。
func extractKey(match string) string {
	inner := match[2 : len(match)-2]
	return strings.TrimSpace(inner)
}
