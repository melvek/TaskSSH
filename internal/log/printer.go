// Package log 提供彩色终端输出。
//
// 基于 fatih/color，自动检测终端是否支持颜色。
// 设置环境变量 NO_COLOR=1 可强制禁用颜色。
//
// 常用方法支持 printf 风格格式化：
//
//	log.Info("done")
//	log.Info("Total: %d, Success: %d", total, success)
package log

import (
	"fmt"

	"github.com/fatih/color"
)

var (
	green    = color.New(color.FgGreen)
	red      = color.New(color.FgRed)
	yellow   = color.New(color.FgYellow)
	cyan     = color.New(color.FgCyan)
	purple   = color.New(color.FgMagenta)
	bold     = color.New(color.Bold)
	boldBlue = color.New(color.Bold, color.FgBlue)
	boldRed  = color.New(color.Bold, color.FgRed)
)

// ==================== 基础输出 ====================

// Info 普通信息，支持格式化。
func Info(format string, args ...any) {
	fmt.Println(formatMsg(format, args...))
}

// Success 成功信息（绿色），支持格式化。
func Success(format string, args ...any) {
	green.Println(formatMsg(format, args...))
}

// Error 错误信息（红色，输出到 stderr），支持格式化。
func Error(format string, args ...any) {
	red.Fprintln(color.Error, "[ERROR] "+formatMsg(format, args...))
}

// ErrorWithCause 带原因的详细错误信息。
func ErrorWithCause(msg string, cause error) {
	red.Fprintln(color.Error, "[ERROR] "+msg)
	if cause != nil {
		red.Fprintln(color.Error, "  - "+cause.Error())
	}
}

// Warn 警告信息（黄色），支持格式化。
func Warn(format string, args ...any) {
	yellow.Println("[WARN] " + formatMsg(format, args...))
}

// Hint 提示信息（青色），支持格式化。
func Hint(format string, args ...any) {
	cyan.Println("    " + formatMsg(format, args...))
}

// Debug 调试信息（紫色），支持格式化。
func Debug(format string, args ...any) {
	purple.Println("[DEBUG] " + formatMsg(format, args...))
}

// ==================== 结构化输出 ====================

// Section 区块标题。
func Section(title string) {
	boldBlue.Printf("\n=== %s ===\n", title)
}

// KeyValue 键值对输出。
func KeyValue(key string, value any) {
	bold.Printf("  %s: ", key)
	cyan.Printf("%v\n", value)
}

// ListItem 列表项输出，格式：  key    -> value
func ListItem(key, value string) {
	bold.Printf("%-20s", key)
	cyan.Printf(" -> %s\n", value)
}

// FailedItem 失败项输出（红色）。
func FailedItem(key, value string) {
	bold.Printf("  %-20s", key)
	red.Printf(" -> %s\n", value)
}

// Step 步骤输出。
func Step(current, total int, name string) {
	boldBlue.Printf("[STEP %d/%d] ", current, total)
	fmt.Println(name)
}

// Progress 进度输出。
func Progress(current, total int, msg string) {
	cyan.Printf("[%d/%d] %s\n", current, total, msg)
}

// Summary 输出执行摘要。
func Summary(total, success, failed int) {
	Section("Execution Summary")
	KeyValue("Total tasks", total)
	KeyValue("Succeeded", success)
	if failed > 0 {
		boldRed.Printf("  %s: ", "Failed")
		red.Printf("%d\n", failed)
	} else {
		KeyValue("Failed", "0")
	}
}

// EmptyLine 输出空行。
func EmptyLine() {
	fmt.Println()
}

// ==================== 颜色控制 ====================

// Disable 禁用颜色。
func Disable() {
	color.NoColor = true
}

// Enable 强制启用颜色。
func Enable() {
	color.NoColor = false
}

// IsColorEnabled 返回颜色是否启用。
func IsColorEnabled() bool {
	return !color.NoColor
}

// ==================== 内部方法 ====================

// formatMsg 格式化消息。
//
//   - 无参数：返回原字符串
//   - 有参数：按 printf 风格格式化
func formatMsg(format string, args ...any) string {
	if len(args) == 0 {
		return format
	}
	return fmt.Sprintf(format, args...)
}
