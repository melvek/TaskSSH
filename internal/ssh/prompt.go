package ssh

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
)

// readPasswordPrompt 从终端读取密码，无回显。
//
// 若终端不支持（IDE / CI），回退到普通读取。
func readPasswordPrompt() string {
	fmt.Print("Password: ")

	if term.IsTerminal(int(syscall.Stdin)) {
		bytePwd, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err == nil {
			return strings.TrimSpace(string(bytePwd))
		}
	}

	// 回退：普通读取
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}
