package cmd

import (
	"fmt"
	"strings"

	"mestrap.com/taskssh/internal/config"
	"mestrap.com/taskssh/internal/resolve"
)

// parseCLIVars 解析 -D 传入的 key=value 变量。
//
// 规则：
//   - 格式必须是 key=value，value 可含 '='
//   - key 和 value 均做 trim
//   - key 不能为空
//   - 不能覆盖保留变量
//   - 重复的 key 以后出现的为准
func parseCLIVars(vars []string) (resolve.Vars, error) {
	out := make(resolve.Vars, len(vars))

	for _, kv := range vars {
		before, after, ok := strings.Cut(kv, "=")
		if !ok {
			return nil, fmt.Errorf("invalid -D %q, expected key=value", kv)
		}

		key := strings.TrimSpace(before)
		val := strings.TrimSpace(after)

		if key == "" {
			return nil, fmt.Errorf("invalid -D %q, key is empty", kv)
		}

		if config.IsReservedVar(key) {
			return nil, fmt.Errorf(
				"%s is a reserved variable and cannot be overridden by -D", key)
		}

		out[key] = val
	}

	return out, nil
}
