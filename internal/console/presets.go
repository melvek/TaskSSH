package console

// ==================== 预设输出 ====================
//
// 所有预设函数内部转调 Output，级别与样式固定。

// Debug 调试信息（Verbose）。
func Debug(format string, args ...any) { defaultLogger.Debug(format, args...) }
func (l *Logger) Debug(format string, args ...any) {
	l.Output(LevelVerbose, StyleDebug, format, args...)
}

// Info 普通信息（Verbose）。
func Info(format string, args ...any) { defaultLogger.Info(format, args...) }
func (l *Logger) Info(format string, args ...any) {
	l.Output(LevelBrief, StyleNone, format, args...)
}

// Hint 提示信息（Verbose，青色缩进）。
func Hint(format string, args ...any) { defaultLogger.Hint(format, args...) }
func (l *Logger) Hint(format string, args ...any) {
	l.Output(LevelVerbose, StyleHint, format, args...)
}

// Warn 警告（Brief）。
func Warn(format string, args ...any) { defaultLogger.Warn(format, args...) }
func (l *Logger) Warn(format string, args ...any) {
	l.Output(LevelBrief, StyleWarn, format, args...)
}

// Success 成功状态（Brief）。
func Success(format string, args ...any) { defaultLogger.Success(format, args...) }
func (l *Logger) Success(format string, args ...any) {
	l.Output(LevelBrief, StyleSuccess, format, args...)
}

// Error 错误（Quiet）。
func Error(format string, args ...any) { defaultLogger.Error(format, args...) }
func (l *Logger) Error(format string, args ...any) {
	l.Output(LevelQuiet, StyleError, format, args...)
}

// Section 区块标题（Brief）。
func Section(title string)             { defaultLogger.Section(title) }
func (l *Logger) Section(title string) { l.Output(LevelBrief, StyleSection, "%s", title) }

// Step 步骤（Brief）。
func Step(current, total int, name string) {
	defaultLogger.Step(current, total, name)
}
func (l *Logger) Step(current, total int, name string) {
	l.Output(LevelBrief, StyleStep, "[STEP %d/%d] %s", current, total, name)
}

// Progress 主机进度（Brief）。
func Progress(current, total int, name, host string) {
	defaultLogger.Progress(current, total, name, host)
}
func (l *Logger) Progress(current, total int, name, host string) {
	l.Output(LevelBrief, StyleProgress, "[Processing %d/%d] %s [%s]", current, total, name, host)
}

// Skipped 步骤跳过（Brief）。
func Skipped(format string, args ...any) { defaultLogger.Skipped(format, args...) }
func (l *Logger) Skipped(format string, args ...any) {
	l.Output(LevelBrief, StyleSkip, format, args...)
}

// Ignored 步骤失败但忽略（Brief）。
func Ignored(format string, args ...any) { defaultLogger.Ignored(format, args...) }
func (l *Logger) Ignored(format string, args ...any) {
	l.Output(LevelBrief, StyleIgnore, format, args...)
}

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
	Output(LevelBrief, StyleNone, "")
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
}

// ExecutionID 输出执行 ID（Brief）。
func ExecutionID(execID string) {
	Output(LevelBrief, StyleNone, "")
	Output(LevelBrief, StyleNone, "----------------------")
	Output(LevelBrief, StyleNone, "Execution ID: %s", execID)
}

// EmptyLine 空行（Verbose）。
func EmptyLine() { Output(LevelVerbose, StyleNone, "") }
