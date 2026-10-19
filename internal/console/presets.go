package console

// Debug 输出详细级别的调试信息。
func (l *Logger) Debug(format string, args ...any) {
	l.Output(LevelVerbose, StyleDebug, format, args...)
}

// Info 输出普通信息。
func (l *Logger) Info(format string, args ...any) {
	l.Output(LevelBrief, StyleNone, format, args...)
}

// Hint 输出详细级别的缩进提示。
func (l *Logger) Hint(format string, args ...any) {
	l.Output(LevelVerbose, StyleHint, format, args...)
}

// Warn 输出警告。
func (l *Logger) Warn(format string, args ...any) {
	l.Output(LevelBrief, StyleWarn, format, args...)
}

// Success 输出成功状态。
func (l *Logger) Success(format string, args ...any) {
	l.Output(LevelBrief, StyleSuccess, format, args...)
}

// Error 输出错误信息。
func (l *Logger) Error(format string, args ...any) {
	l.Output(LevelQuiet, StyleError, format, args...)
}

// Section 输出区块标题。
func (l *Logger) Section(title string) {
	l.Output(LevelBrief, StyleSection, "%s", title)
}

// Step 输出任务步骤状态。
func (l *Logger) Step(current, total int, name string) {
	l.Output(LevelBrief, StyleStep, "[STEP %d/%d] %s", current, total, name)
}

// Progress 输出主机执行进度。
func (l *Logger) Progress(current, total int, name, host string) {
	l.Output(LevelBrief, StyleProgress, "[Processing %d/%d] %s [%s]", current, total, name, host)
}

// Skipped 输出跳过步骤的状态。
func (l *Logger) Skipped(format string, args ...any) {
	l.Output(LevelBrief, StyleSkip, format, args...)
}

// Ignored 输出已忽略的错误状态。
func (l *Logger) Ignored(format string, args ...any) {
	l.Output(LevelBrief, StyleIgnore, format, args...)
}

// FailedItem 输出失败的键值项。
func (l *Logger) FailedItem(key, value string) {
	l.Output(LevelQuiet, StyleFailedItem, "%s: %s", key, value)
}

// KeyValue 输出详细级别的键值对。
func (l *Logger) KeyValue(key string, value any) {
	l.Output(LevelVerbose, StyleNone, "%s: %v", key, value)
}

// ListItem 输出详细级别的对齐列表项。
func (l *Logger) ListItem(key, value string) {
	l.Output(LevelVerbose, StyleNone, "%-20s: %s", key, value)
}

// Targets 输出选中的主机列表。
func (l *Logger) Targets(lines []string) {
	l.Section("Target hosts")
	for _, line := range lines {
		l.Output(LevelBrief, StyleNone, "%s", line)
	}
	l.Output(LevelBrief, StyleNone, "")
}

// Summary 输出执行摘要。
func (l *Logger) Summary(total, success, failed int, failedHosts []string) {
	l.Output(LevelBrief, StyleNone, "")
	l.Section("Execution Summary")
	l.Output(LevelBrief, StyleNone, "Total tasks: %d", total)
	l.Output(LevelBrief, StyleNone, "Succeeded: %d", success)
	if failed > 0 {
		l.Output(LevelBrief, StyleFailedItem, "Failed: %d", failed)
		for _, host := range failedHosts {
			l.Output(LevelBrief, StyleFailedItem, "  - %s", host)
		}
		return
	}
	l.Output(LevelBrief, StyleNone, "Failed: 0")
}

// ExecutionID 输出本次执行 ID。
func (l *Logger) ExecutionID(execID string) {
	l.Output(LevelBrief, StyleNone, "")
	l.Output(LevelBrief, StyleNone, "----------------------")
	l.Output(LevelBrief, StyleNone, "Execution ID: %s", execID)
}

// EmptyLine 在详细级别输出空行。
func (l *Logger) EmptyLine() {
	l.Output(LevelVerbose, StyleNone, "")
}
