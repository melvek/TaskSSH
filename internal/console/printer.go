// Package console 提供分级、彩色终端输出。
//
// 三个级别：Verbose / Brief / Quiet，由 SetLevel 设置。
// 并发任务可创建独立 Logger，避免不同主机的输出混杂。
//
// 颜色基于 fatih/color 自动检测，NO_COLOR=1 可强制禁用。
package console

import (
	"fmt"
	"io"
	"os"
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

	// Default 用于不关联特定主机的输出。
	Default = &Logger{w: os.Stdout}
)

// Logger 将格式化输出写入自己的目标。
type Logger struct {
	mu sync.Mutex
	w  io.Writer
}

// NewLogger 创建写入 w 的 Logger；w 为 nil 时使用 os.Stdout。
func NewLogger(w io.Writer) *Logger {
	if w == nil {
		w = os.Stdout
	}
	return &Logger{w: w}
}

type styleMeta struct {
	color  *color.Color
	prefix string
}

var styleMetas = map[Style]styleMeta{
	StyleNone:       {},
	StyleSuccess:    {green, ""},
	StyleError:      {red, "Error: "},
	StyleWarn:       {yellow, "Warning: "},
	StyleHint:       {cyan, "  "},
	StyleDebug:      {gray, ""},
	StyleSection:    {blue, ">>> "},
	StyleStep:       {cyan, ""},
	StyleProgress:   {cyan, ""},
	StyleSkip:       {cyan, ""},
	StyleIgnore:     {yellow, ""},
	StyleFailedItem: {red, ""},
}

// Output 在应用级别过滤和样式后输出消息。
func (l *Logger) Output(level Level, style Style, format string, args ...any) {
	if level > currentLevel {
		return
	}

	message := format
	if len(args) > 0 {
		message = fmt.Sprintf(format, args...)
	}

	meta := styleMetas[style]
	message = meta.prefix + strings.TrimRight(message, "\n")

	l.mu.Lock()
	defer l.mu.Unlock()
	if meta.color != nil && !color.NoColor {
		meta.color.Fprintln(l.w, message)
		return
	}
	fmt.Fprintln(l.w, message)
}

// CommandWriter 返回远程命令 stdout 的输出目标。
// Quiet 级别丢弃 stdout，其他级别写入此 Logger。
func (l *Logger) CommandWriter() io.Writer {
	if currentLevel <= LevelQuiet {
		return io.Discard
	}
	return synchronizedWriter{logger: l}
}

// StderrWriter 返回远程命令 stderr 的输出目标。
func (l *Logger) StderrWriter() io.Writer {
	return synchronizedWriter{logger: l}
}

// RawOutput 原样输出 s，不进行级别过滤。
func (l *Logger) RawOutput(s string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	_, _ = io.WriteString(l.w, s)
}

type synchronizedWriter struct {
	logger *Logger
}

func (w synchronizedWriter) Write(p []byte) (int, error) {
	w.logger.mu.Lock()
	defer w.logger.mu.Unlock()
	return w.logger.w.Write(p)
}

// Disable 禁用颜色。
func Disable() { color.NoColor = true }

// Enable 强制启用颜色。
func Enable() { color.NoColor = false }

// IsColorEnabled 返回颜色是否启用。
func IsColorEnabled() bool { return !color.NoColor }
