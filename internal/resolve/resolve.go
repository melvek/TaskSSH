package resolve

import (
	"fmt"
	"regexp"
	"strings"
)

type Vars map[string]any

var placeholder = regexp.MustCompile(`\$\{([^}]+)\}`)

const MaxDepth = 4

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

func (v Vars) replaceOnce(s string) string {
	return placeholder.ReplaceAllStringFunc(s, func(match string) string {
		key := strings.TrimSpace(match[2 : len(match)-1])
		if val, ok := v[key]; ok {
			return fmt.Sprintf("%v", val)
		}
		return match
	})
}

func findMissing(s string) []string {
	var keys []string
	for _, m := range placeholder.FindAllStringSubmatch(s, -1) {
		keys = append(keys, strings.TrimSpace(m[1]))
	}
	return keys
}
