package console

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
