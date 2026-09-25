package action

import "strings"

// shellQuote 对字符串做最小限度的 shell 转义。
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
