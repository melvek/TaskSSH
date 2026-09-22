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
	bold     = color.New(color.Bold)
	boldBlue = color.New(color.Bold, color.FgBlue)
	boldRed  = color.New(color.Bold, color.FgRed)
)

// Info 普通信息。
func Info(msg string) {
	fmt.Println(msg)
}

// Success 成功信息（绿色）。
func Success(msg string) {
	green.Println(msg)
}

// Error 错误信息（红色，输出到 stderr）。
func Error(msg string) {
	red.Fprintln(color.Error, "[ERROR] "+msg)
}

// Warn 警告信息（黄色）。
func Warn(msg string) {
	yellow.Println("[WARN] " + msg)
}

// Hint 提示信息（青色）。
func Hint(msg string) {
	cyan.Println("    " + msg)
}

// Section 区块标题。
func Section(title string) {
	boldBlue.Printf("\n=== %s ===\n", title)
}

// KeyValue 键值对。
func KeyValue(key string, value any) {
	bold.Printf("  %s: ", key)
	cyan.Printf("%v\n", value)
}

// ListItem 列表项。
func ListItem(key, value string) {
	bold.Printf("  %-20s", key)
	cyan.Printf(" -> %s\n", value)
}

// EmptyLine 空行。
func EmptyLine() {
	fmt.Println()
}

// Disable 禁用颜色。
func Disable() {
	color.NoColor = true
}
