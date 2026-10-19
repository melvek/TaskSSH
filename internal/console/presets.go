package console

// Debug 输出详细级别的调试信息。
func (l *Logger) Debug(format string, args ...any) {
	l.Output(LevelDebug, StyleDebug, format, args...)
}

// Trace 输出高频详细信息。
func (l *Logger) Trace(format string, args ...any) {
	l.Output(LevelTrace, StyleDebug, format, args...)
}

// Info 输出普通信息。
func (l *Logger) Info(format string, args ...any) {
	l.Output(LevelInfo, StyleNone, format, args...)
}

// Hint 输出详细级别的缩进提示。
func (l *Logger) Hint(format string, args ...any) {
	l.Output(LevelDebug, StyleHint, format, args...)
}

// Warn 输出警告。
func (l *Logger) Warn(format string, args ...any) {
	l.Output(LevelWarn, StyleWarn, format, args...)
}

// Success 输出成功状态。
func (l *Logger) Success(format string, args ...any) {
	l.Output(LevelDebug, StyleSuccess, format, args...)
}

// Error 输出错误信息。
func (l *Logger) Error(format string, args ...any) {
	l.Output(LevelError, StyleError, format, args...)
}

// Section 输出区块标题。
func (l *Logger) Section(title string) {
	l.Output(LevelInfo, StyleSection, "%s", title)
}

// Step 输出任务步骤状态。
func (l *Logger) Step(current, total int, name string) {
	l.Output(LevelInfo, StyleStep, "[STEP %d/%d] %s", current, total, name)
}

// Progress 输出主机执行进度。
func (l *Logger) Progress(current, total int, name, host string) {
	l.Output(LevelInfo, StyleProgress, "[Processing %d/%d] %s [%s]", current, total, name, host)
}

// Skipped 输出跳过步骤的状态。
func (l *Logger) Skipped(format string, args ...any) {
	l.Output(LevelInfo, StyleSkip, format, args...)
}

// Ignored 输出已忽略的错误状态。
func (l *Logger) Ignored(format string, args ...any) {
	l.Output(LevelWarn, StyleIgnore, format, args...)
}

// FailedItem 输出失败的键值项。
func (l *Logger) FailedItem(key, value string) {
	l.Output(LevelError, StyleFailedItem, "%s: %s", key, value)
}

// KeyValue 输出详细级别的键值对。
func (l *Logger) KeyValue(key string, value any) {
	l.Output(LevelDebug, StyleNone, "%s: %v", key, value)
}

// ListItem 输出详细级别的对齐列表项。
func (l *Logger) ListItem(key, value string) {
	l.Output(LevelDebug, StyleNone, "%-20s: %s", key, value)
}

// Targets 输出选中的主机列表。
func (l *Logger) Targets(lines []string) {
	l.Section("Target hosts")
	for _, line := range lines {
		l.Output(LevelInfo, StyleNone, "%s", line)
	}
	l.Output(LevelInfo, StyleNone, "")
}

// Summary 输出执行摘要。
func (l *Logger) Summary(total, success, failed int, failedHosts []string) {
	l.Output(LevelInfo, StyleNone, "")
	l.Section("Execution Summary")
	l.Output(LevelInfo, StyleNone, "Total tasks: %d", total)
	l.Output(LevelInfo, StyleNone, "Succeeded: %d", success)
	if failed > 0 {
		l.Output(LevelInfo, StyleFailedItem, "Failed: %d", failed)
		for _, host := range failedHosts {
			l.Output(LevelInfo, StyleFailedItem, "  - %s", host)
		}
		return
	}
	l.Output(LevelInfo, StyleNone, "Failed: 0")
}

// ExecutionID 输出本次执行 ID。
func (l *Logger) ExecutionID(execID string) {
	l.Output(LevelInfo, StyleNone, "")
	l.Output(LevelInfo, StyleNone, "----------------------")
	l.Output(LevelInfo, StyleNone, "Execution ID: %s", execID)
}

// EmptyLine 在调试级别输出空行。
func (l *Logger) EmptyLine() {
	l.Output(LevelDebug, StyleNone, "")
}
