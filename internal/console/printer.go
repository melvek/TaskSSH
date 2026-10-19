// Package console 提供分级、彩色终端输出。
//
// 三个级别：Verbose / Brief / Quiet，由 SetLevel 设置。
// Output 是默认 Logger 的底层出口；并发任务可使用独立 Logger 缓冲输出。
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

	defaultLogger = &Logger{}
)

// Logger 将输出发送到指定 writer。单个 Logger 可以安全地被多个写入者共享。
type Logger struct {
	mu sync.Mutex
	w  io.Writer
}

// NewLogger 创建一个写入指定目标的 Logger。nil writer 使用 os.Stdout。
func NewLogger(w io.Writer) *Logger {
	return &Logger{w: w}
}

// DefaultLogger 返回全局默认 Logger。
func DefaultLogger() *Logger {
	return defaultLogger
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

// Output 是默认 Logger 的底层输出出口。
func Output(level Level, style Style, format string, args ...any) {
	defaultLogger.Output(level, style, format, args...)
}

// Output 按指定级别和样式输出。
func (l *Logger) Output(level Level, style Style, format string, args ...any) {
	if level > currentLevel {
		return
	}
	l.write(style, formatMsg(format, args...))
}

func (l *Logger) write(style Style, msg string) {
	if l == nil {
		l = defaultLogger
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	meta := styleMetas[style]
	body := meta.prefix + strings.TrimRight(msg, "\n")
	fprintStyled(l.writer(), meta, body, true)
}

func (l *Logger) writer() io.Writer {
	if l == nil || l.w == nil {
		return os.Stdout
	}
	return l.w
}

type synchronizedWriter struct {
	logger *Logger
}

func (w synchronizedWriter) Write(p []byte) (int, error) {
	l := w.logger
	if l == nil {
		l = defaultLogger
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.writer().Write(p)
}

// CommandWriter 返回远程命令 stdout 的输出目标。
//
// Quiet 级别下丢弃输出，其他级别写入该 Logger。
func CommandWriter() io.Writer {
	return defaultLogger.CommandWriter()
}

// CommandWriter 返回该 Logger 的远程命令 stdout 输出目标。
func (l *Logger) CommandWriter() io.Writer {
	if currentLevel <= LevelQuiet {
		return io.Discard
	}
	return synchronizedWriter{logger: l}
}

// StderrWriter 返回远程命令 stderr 的默认输出目标。
func StderrWriter() io.Writer {
	return defaultLogger.StderrWriter()
}

// StderrWriter 返回该 Logger 的远程命令 stderr 输出目标。
func (l *Logger) StderrWriter() io.Writer {
	return synchronizedWriter{logger: l}
}

// RawOutput 原样输出字符串（不做级别过滤）。
func RawOutput(s string) {
	defaultLogger.RawOutput(s)
}

// RawOutput 原样输出字符串（不做级别过滤）。
func (l *Logger) RawOutput(s string) {
	if l == nil {
		l = defaultLogger
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	_, _ = io.WriteString(l.writer(), s)
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

// Disable 禁用颜色。
func Disable() { color.NoColor = true }

// Enable 强制启用颜色。
func Enable() { color.NoColor = false }

// IsColorEnabled 返回颜色是否启用。
func IsColorEnabled() bool { return !color.NoColor }
