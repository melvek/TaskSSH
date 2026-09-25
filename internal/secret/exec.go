package secret

import (
	"fmt"
	"os/exec"
	"strings"
)

// runKeyScript 执行可执行文件，把输出作为密钥。
func runKeyScript(path string) (string, error) {
	cmd := exec.Command(path)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("run key script %s: %w", path, err)
	}

	key := strings.TrimSpace(string(out))
	if key == "" {
		return "", fmt.Errorf("key script %s produced empty output", path)
	}

	return key, nil
}
