package routergen

// Mode 定义运行模式
type Mode int

const (
	// ModeNormal 正常模式
	ModeNormal Mode = iota
	// ModeQuiet 静默模式
	ModeQuiet
	// ModeDebug 调试模式
	ModeDebug
)

// String 返回模式的字符串表示
func (m Mode) String() string {
	switch m {
	case ModeNormal:
		return "normal"
	case ModeQuiet:
		return "quiet"
	case ModeDebug:
		return "debug"
	default:
		return "unknown"
	}
}

// IsQuiet 检查是否为静默模式
func (m Mode) IsQuiet() bool {
	return m == ModeQuiet
}

// IsDebug 检查是否为调试模式
func (m Mode) IsDebug() bool {
	return m == ModeDebug
}

// ShouldPrint 检查是否应该打印信息
func (m Mode) ShouldPrint() bool {
	return !m.IsQuiet() || m.IsDebug()
}
