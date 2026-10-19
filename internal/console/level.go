package console

// Level 是日志级别，值越大越详细。
type Level int

const (
	// LevelError 只输出错误。
	LevelError Level = iota
	// LevelWarn 增加警告。
	LevelWarn
	// LevelInfo 增加常规执行状态。
	LevelInfo
	// LevelDebug 增加调试信息和命令输出。
	LevelDebug
	// LevelTrace 增加高频细节。
	LevelTrace
)

var currentLevel = LevelInfo

// SetLevel 设置全局日志级别。
//
// 必须在启动工作 goroutine 之前调用一次。
func SetLevel(l Level) { currentLevel = l }

// GetLevel 返回当前日志级别。
func GetLevel() Level { return currentLevel }

func (l Level) String() string {
	switch l {
	case LevelError:
		return "error"
	case LevelWarn:
		return "warn"
	case LevelInfo:
		return "info"
	case LevelDebug:
		return "debug"
	case LevelTrace:
		return "trace"
	}
	return "unknown"
}
