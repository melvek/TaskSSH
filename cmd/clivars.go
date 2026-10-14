package cmd

import (
	"fmt"
	"strings"

	"mestrap.com/taskssh/internal/config"
	"mestrap.com/taskssh/internal/resolve"
)

// 解析 -Dkey=value 自定义参数，自定义参数不能使用保留变量
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
			return nil, fmt.Errorf("%s is a reserved variable and cannot be overridden by -D", key)
		}

		out[key] = val
	}

	return out, nil
}
