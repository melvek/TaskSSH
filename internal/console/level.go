package console

// Level 是日志级别，值越大越详细。
type Level int

const (
	// LevelQuiet 只输出错误。
	LevelQuiet Level = iota
	// LevelBrief 输出步骤状态、区块标题、警告与错误。
	LevelBrief
	// LevelVerbose 输出全部内容，含调试信息与命令输出。
	LevelVerbose
)

var currentLevel = LevelBrief

// SetLevel 设置全局日志级别。
//
// 必须在启动工作 goroutine 之前调用一次。
func SetLevel(l Level) { currentLevel = l }

// GetLevel 返回当前日志级别。
func GetLevel() Level { return currentLevel }

func (l Level) String() string {
	switch l {
	case LevelQuiet:
		return "quiet"
	case LevelBrief:
		return "brief"
	case LevelVerbose:
		return "verbose"
	}
	return "unknown"
}

// Style 是输出样式，决定颜色与固定前缀。
type Style int

const (
	StyleNone Style = iota
	StyleSuccess
	StyleError
	StyleWarn
	StyleHint
	StyleDebug
	StyleSection
	StyleStep
	StyleProgress
	StyleSkip
	StyleIgnore
	StyleFailedItem
)
