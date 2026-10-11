package ssh

import (
	"mestrap.com/taskssh/internal/prompt"
)

// readPasswordPrompt 从终端读取密码，无回显。
//
// 读取失败时返回空字符串，由上层处理。
func readPasswordPrompt() string {
	pwd, err := prompt.Hidden("Password: ")
	if err != nil {
		return ""
	}
	return pwd
}
