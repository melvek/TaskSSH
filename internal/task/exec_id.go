package task

import (
	"crypto/rand"
)

// base36Chars 是 Base36 字符集（数字 + 大写字母）。
const base36Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

// generateExecID 生成本次执行的唯一标识，8 位 Base36 构成的随机字符串
func generateExecID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		// 随机失败时返回全零，不阻断流程
		return "00000000"
	}

	var n uint64
	for _, x := range b {
		n = n<<8 | uint64(x)
	}

	result := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		result[i] = base36Chars[n%36]
		n /= 36
	}
	return string(result)
}
