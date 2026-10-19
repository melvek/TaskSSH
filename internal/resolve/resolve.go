package resolve

import (
	"fmt"
	"regexp"
	"strings"
)

// 变量池，仅支持扁平化键值对
type Vars map[string]any

// 预编译正则：匹配 {{ var }}，允许内部有空格，禁止嵌套花括号
// 注意：[^{}]+? 确保不会跨越多层花括号，适合扁平替换
var placeholder = regexp.MustCompile(`\{\{\s*([^{}]+?)\s*\}\}`)

// 定义递归展开的最大层数，防止恶意构造的循环引用导致栈溢出或死循环
const MaxDepth = 5

// 执行递归变量替换，并返回适用于 Shell/命令行的安全字符串
func (v Vars) Replace(template string) (string, error) {
	if template == "" {
		return "", nil
	}

	current := template
	for range MaxDepth {
		// 1. 执行单次替换
		next, changed := v.replaceOnce(current)

		// 2. 如果没有发生变化，说明所有可替换变量已处理完毕，退出循环
		if !changed {
			break
		}
		current = next
	}

	// 3. 检查是否还有未解析的变量
	missing := findMissing(current)
	if len(missing) > 0 {
		return "", fmt.Errorf("unresolved variables: %s", strings.Join(missing, ", "))
	}

	return current, nil
}

// replaceOnce 执行单轮替换，返回新字符串和是否发生变化的标志
func (v Vars) replaceOnce(s string) (string, bool) {
	changed := false
	result := placeholder.ReplaceAllStringFunc(s, func(match string) string {
		key := extractKey(match)
		val, ok := v[key]
		if !ok {
			// 变量不存在，保持原样，留给后续检查报错
			return match
		}

		// 标记发生了变化
		changed = true

		// 转换为字符串。注意：这里暂不转义，因为变量值可能还需要作为模板的一部分再次被解析
		// 例如 service_path = package/{{date}}，如果此时转义了斜杠或花括号，会导致下一轮解析失败
		return fmt.Sprintf("%v", val)
	})
	return result, changed
}

// findMissing 找出所有未被替换的变量名
func findMissing(s string) []string {
	var keys []string
	matches := placeholder.FindAllStringSubmatch(s, -1)
	for _, m := range matches {
		if len(m) > 1 {
			keys = append(keys, strings.TrimSpace(m[1]))
		}
	}
	return keys
}

// extractKey 从匹配字符串中提取变量名
func extractKey(match string) string {
	if len(match) < 4 {
		return ""
	}
	// 去掉前后的 {{ 和 }}
	inner := match[2 : len(match)-2]
	return strings.TrimSpace(inner)
}
