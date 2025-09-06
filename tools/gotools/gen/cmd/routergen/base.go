package routergen

// ProtocolIDMapping 用于存储协议ID映射
type ProtocolIDMapping struct {
	PackageName string
	MessageIDs  []MessageID
}

// MessageName 用于存储消息名称和完整路径
type MessageName struct {
	Name     string
	FullName string
	MsgWall  string //消息墙名称
}

// MessageID 用于存储消息ID
type MessageID struct {
	Name string
	ID   uint32
}

// Message 消息结构，用于存储消息定义及其注释
type Message struct {
	Name     string
	FullName string // 包含父消息路径的完整名称
	Comment  string
	Fields   []Field
	Response string // 响应消息名称，如果有的话
}

// Field 字段结构，用于存储字段定义及其注释
type Field struct {
	Name    string
	Type    string
	Index   int
	Comment string
}
