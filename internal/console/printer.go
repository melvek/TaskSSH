// Package console 提供分级、彩色终端输出。
//
// 三个级别：Verbose / Brief / Quiet，由 SetLevel 设置。
// Output 是唯一底层出口，其余预设函数见 presets.go。
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

// CommandWriter 返回远程命令 stdout 的输出目标。
//
// 规则：
//   - Quiet 级别：丢弃
//   - 其他级别：输出到当前 goroutine 的目标
func CommandWriter() io.Writer {
	if currentLevel <= LevelQuiet {
		return io.Discard
	}
	return GetOutput()
}

// StderrWriter 返回远程命令 stderr 的输出目标。
//
// 规则：所有级别都输出。
// stderr 表示错误信息，任何级别下都应可见。
func StderrWriter() io.Writer {
	return GetOutput()
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
