// Package console 提供分级、彩色终端输出。
//
// 三个级别：Verbose / Brief / Quiet，由 SetLevel 设置。
// Output 是唯一底层出口，其余预设函数内部转调 Output。
//
// 颜色基于 fatih/color 自动检测，NO_COLOR=1 可强制禁用。
//
// 支持按 goroutine 缓冲输出：串行时直接写 os.Stdout，并发时每个 goroutine
// 输出到独立缓冲，执行完统一刷出。所有输出均独立成行。
package console

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
	green  = color.New(color.FgGreen)
	red    = color.New(color.FgRed)
	yellow = color.New(color.FgYellow)
	cyan   = color.New(color.FgCyan)
	blue   = color.New(color.FgBlue)
	gray   = color.New(color.FgHiBlack)

	// outputMu 保护所有输出。
	outputMu sync.Mutex

	// states 保存每个 goroutine 的输出目标。
	states sync.Map // int -> *outputState
)

type outputState struct {
	w io.Writer
}

type styleMeta struct {
	color  *color.Color
	prefix string
}

var styleMetas = map[Style]styleMeta{
	StyleNone:       {},
	StyleSuccess:    {green, ""},
	StyleError:      {red, "[ERROR] "},
	StyleWarn:       {yellow, "[WARN] "},
	StyleHint:       {cyan, "  "},
	StyleDebug:      {gray, "[DEBUG] "},
	StyleSection:    {blue, ">>> "},
	StyleStep:       {cyan, ""},
	StyleProgress:   {cyan, ""},
	StyleSkip:       {cyan, ""},
	StyleIgnore:     {yellow, ""},
	StyleFailedItem: {red, ""},
}

// ==================== 输出目标控制 ====================

// SetOutput 设置当前 goroutine 的输出目标。
func SetOutput(w io.Writer) {
	outputMu.Lock()
	defer outputMu.Unlock()
	states.Store(getGID(), &outputState{w: w})
}

// ResetOutput 恢复当前 goroutine 的默认输出。
func ResetOutput() {
	outputMu.Lock()
	defer outputMu.Unlock()
	states.Delete(getGID())
}

// GetOutput 返回当前 goroutine 的输出目标。
func GetOutput() io.Writer {
	return getState().w
}

// VerboseWriter 返回一个仅在 Verbose 级别下生效的输出目标。
//
// 非 Verbose 级别返回 io.Discard，写入被静默丢弃。
// 用于命令 stdout/stderr 这类「只在详细模式下显示」的输出。
func VerboseWriter() io.Writer {
	if currentLevel >= LevelVerbose {
		return GetOutput()
	}
	return io.Discard
}

// getState 返回当前 goroutine 的状态，不存在则创建。
func getState() *outputState {
	gid := getGID()
	if v, ok := states.Load(gid); ok {
		return v.(*outputState)
	}
	st := &outputState{w: os.Stdout}
	actual, _ := states.LoadOrStore(gid, st)
	return actual.(*outputState)
}

// getGID 获取当前 goroutine 的 ID。
func getGID() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	fields := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))
	id, _ := strconv.Atoi(fields[0])
	return id
}

// ==================== 底层出口 ====================

// Output 是唯一底层输出出口。
//
// 无 args 时按纯字符串输出，有 args 时按 fmt.Sprintf 语义格式化。
// 调用方优先使用预设函数（Success / Error / Step 等）。
func Output(level Level, style Style, format string, args ...any) {
	if level > currentLevel {
		return
	}
	write(style, formatMsg(format, args...))
}

func write(style Style, msg string) {
	outputMu.Lock()
	defer outputMu.Unlock()

	st := getState()
	out := st.w
	meta := styleMetas[style]
	body := meta.prefix + strings.TrimRight(msg, "\n")

	fprintStyled(out, meta, body, true)
}

func fprintStyled(out io.Writer, meta styleMeta, s string, newline bool) {
	if meta.color != nil && !color.NoColor {
		if newline {
			meta.color.Fprintln(out, s)
		} else {
			meta.color.Fprint(out, s)
		}
		return
	}
	if newline {
		fmt.Fprintln(out, s)
	} else {
		fmt.Fprint(out, s)
	}
}

func formatMsg(format string, args ...any) string {
	if len(args) == 0 {
		return format
	}
	return fmt.Sprintf(format, args...)
}

// ==================== 预设输出 ====================

// Debug 调试信息（Verbose）。
func Debug(format string, args ...any) { Output(LevelVerbose, StyleDebug, format, args...) }

// Info 普通信息（Verbose）。
func Info(format string, args ...any) { Output(LevelVerbose, StyleNone, format, args...) }

// Hint 提示信息（Verbose，青色缩进）。
func Hint(format string, args ...any) { Output(LevelVerbose, StyleHint, format, args...) }

// Warn 警告（Brief）。
func Warn(format string, args ...any) { Output(LevelBrief, StyleWarn, format, args...) }

// Success 成功状态（Brief）。
func Success(format string, args ...any) { Output(LevelBrief, StyleSuccess, format, args...) }

// Error 错误（Quiet）。
func Error(format string, args ...any) { Output(LevelQuiet, StyleError, format, args...) }

// Section 区块标题（Brief）。
func Section(title string) { Output(LevelBrief, StyleSection, "%s", title) }

// Step 步骤（Brief）。
func Step(current, total int, name string) {
	Output(LevelBrief, StyleStep, "[STEP %d/%d] %s", current, total, name)
}

// Progress 主机进度（Brief）。
func Progress(current, total int, name, host string) {
	Output(LevelBrief, StyleProgress, "[Processing %d/%d] %s [%s]", current, total, name, host)
}

// Skipped 步骤跳过（Brief）。
func Skipped(format string, args ...any) { Output(LevelBrief, StyleSkip, format, args...) }

// Ignored 步骤失败但忽略（Brief）。
func Ignored(format string, args ...any) { Output(LevelBrief, StyleIgnore, format, args...) }

// FailedItem 失败项（Quiet）。
func FailedItem(key, value string) {
	Output(LevelQuiet, StyleFailedItem, "%s: %s", key, value)
}

// KeyValue 键值对（Verbose）。
func KeyValue(key string, value any) {
	Output(LevelVerbose, StyleNone, "%s: %v", key, value)
}

// ListItem 列表项（Verbose）。
func ListItem(key, value string) {
	Output(LevelVerbose, StyleNone, "%-20s: %s", key, value)
}

// Targets 输出目标主机清单（Brief）。
//
// 每行格式为 "name: host"，由调用方拼好。
func Targets(lines []string) {
	Section("Target hosts")
	for _, l := range lines {
		Output(LevelBrief, StyleNone, "%s", l)
	}
	Output(LevelBrief, StyleNone, "")
}

// Summary 执行摘要（Brief）。
func Summary(total, success, failed int, failedHosts []string) {
	Section("Execution Summary")
	Output(LevelBrief, StyleNone, "Total tasks: %d", total)
	Output(LevelBrief, StyleNone, "Succeeded: %d", success)
	if failed > 0 {
		Output(LevelBrief, StyleFailedItem, "Failed: %d", failed)
		for _, h := range failedHosts {
			Output(LevelBrief, StyleFailedItem, "  - %s", h)
		}
	} else {
		Output(LevelBrief, StyleNone, "Failed: 0")
	}
	Output(LevelBrief, StyleNone, "")
}

func ExecutionID(execId string) {
	Output(LevelBrief, StyleNone, "----------------------")
	Output(LevelBrief, StyleNone, "Execution ID: %s", execId)
}

// EmptyLine 空行（Verbose）。
func EmptyLine() { Output(LevelVerbose, StyleNone, "") }

// RawOutput 原样输出字符串（不做级别过滤，用于刷出缓冲区）。
func RawOutput(s string) {
	outputMu.Lock()
	defer outputMu.Unlock()

	st := getState()
	fmt.Fprint(st.w, s)
}

// ==================== 颜色控制 ====================

// Disable 禁用颜色。
func Disable() { color.NoColor = true }

// Enable 强制启用颜色。
func Enable() { color.NoColor = false }

// IsColorEnabled 返回颜色是否启用。
func IsColorEnabled() bool { return !color.NoColor }
