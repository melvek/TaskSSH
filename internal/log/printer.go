// Package log 提供彩色终端输出。
//
// 基于 fatih/color，自动检测终端是否支持颜色。
// 设置环境变量 NO_COLOR=1 可强制禁用颜色。
//
// 支持按 goroutine 缓冲输出：
//   - 串行执行时直接输出到 os.Stdout
//   - 并发执行时每个 goroutine 输出到独立缓冲，执行完统一输出
package log

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/fatih/color"
)

var (
	green    = color.New(color.FgGreen)
	red      = color.New(color.FgRed)
	yellow   = color.New(color.FgYellow)
	cyan     = color.New(color.FgCyan)
	purple   = color.New(color.FgMagenta)
	boldBlue = color.New(color.Bold, color.FgBlue)
	boldRed  = color.New(color.Bold, color.FgRed)

	// outputMu 保护输出，避免多个 goroutine 同时写终端
	outputMu sync.Mutex

	// outputs 保存每个 goroutine 的输出目标
	// 默认 os.Stdout，并发时切换为 bytes.Buffer
	outputs sync.Map // int -> io.Writer
)

// ==================== 输出目标控制 ====================

// SetOutput 设置当前 goroutine 的输出目标。
func SetOutput(w io.Writer) {
	outputs.Store(getGID(), w)
}

// ResetOutput 恢复当前 goroutine 的默认输出（os.Stdout）。
func ResetOutput() {
	outputs.Delete(getGID())
}

// GetOutput 返回当前 goroutine 的输出目标。
func GetOutput() io.Writer {
	if w, ok := outputs.Load(getGID()); ok {
		return w.(io.Writer)
	}
	return os.Stdout
}

// getGID 获取当前 goroutine 的 ID。
func getGID() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	fields := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))
	id, _ := strconv.Atoi(fields[0])
	return id
}

// ==================== 基础输出 ====================

// Info 普通信息，支持格式化。
func Info(format string, args ...any) {
	write(formatMsg(format, args...), nil)
}

// Success 成功信息（绿色），支持格式化。
func Success(format string, args ...any) {
	write(formatMsg(format, args...), green)
}

// Error 错误信息（红色），支持格式化。
func Error(format string, args ...any) {
	write("[ERROR] "+formatMsg(format, args...), red)
}

// Warn 警告信息（黄色），支持格式化。
func Warn(format string, args ...any) {
	write("[WARN] "+formatMsg(format, args...), yellow)
}

// Hint 提示信息（青色），支持格式化。
func Hint(format string, args ...any) {
	write("  "+formatMsg(format, args...), cyan)
}

// Debug 调试信息（紫色），支持格式化。
func Debug(format string, args ...any) {
	write("[DEBUG] "+formatMsg(format, args...), purple)
}

// ==================== 结构化输出 ====================

// RawOutput 原样输出字符串（加锁）。
//
// 用于输出预先生成的内容（如按主机缓冲的日志）。
func RawOutput(s string) {
	outputMu.Lock()
	defer outputMu.Unlock()
	fmt.Print(s)
}

// Section 区块标题。
func Section(title string) {
	write("\n>>> "+title, boldBlue)
}

// KeyValue 键值对输出。
func KeyValue(key string, value any) {
	write(fmt.Sprintf("%s: %v", key, value), nil)
}

// ListItem 列表项输出。
func ListItem(key, value string) {
	write(fmt.Sprintf("%-20s: %s", key, value), nil)
}

// FailedItem 失败项输出（红色）。
func FailedItem(key, value string) {
	write(fmt.Sprintf("%-20s: %s", key, value), red)
}

// Step 步骤输出。
func Step(current, total int, name string) {
	write(fmt.Sprintf("[STEP %d/%d] %s", current, total, name), cyan)
}

// Progress 主机进度输出。
func Progress(current, total int, name, host string) {
	write(fmt.Sprintf("[Processing %d/%d] %s [%s]", current, total, name, host), cyan)
}

// Summary 执行摘要。
func Summary(total, success, failed int) {
	Section("Execution Summary")
	KeyValue("Total tasks", total)
	KeyValue("Succeeded", success)
	if failed > 0 {
		write(fmt.Sprintf("Failed: %d", failed), boldRed)
	} else {
		KeyValue("Failed", "0")
	}
}

// EmptyLine 空行。
func EmptyLine() {
	write("", nil)
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

// write 写一行到当前 goroutine 的输出目标。
//
// msg 末尾的换行符会被去掉，由 Fprintln 自动加。
func write(msg string, c *color.Color) {
	outputMu.Lock()
	defer outputMu.Unlock()

	out := GetOutput()
	msg = strings.TrimRight(msg, "\n")

	if c != nil && !color.NoColor {
		c.Fprintln(out, msg)
	} else {
		fmt.Fprintln(out, msg)
	}
}

// formatMsg 格式化消息。
func formatMsg(format string, args ...any) string {
	if len(args) == 0 {
		return format
	}
	return fmt.Sprintf(format, args...)
}
