package protocolids

// MessageInfo 协议ID消息信息
// 用于在生成 protocol_ids.pb.go 文件时存储消息的完整名称和用于 Hash 的名称
type MessageInfo struct {
	// FullName 完整消息名称，用于在 protocol_ids.pb.go 中显示
	// 例如: "Request_SearchBook", "Request_SearchBook_Rsp", "Notify_BeAttacked"
	FullName string

	// PidName 用于生成 Hash 的名称
	// 格式: {PackageName}-{FullName}
	// 例如: "Core-Request_SearchBook", "Core-Request_SearchBook_Rsp", "Core-Notify_BeAttacked"
	PidName string
}

// MessageID 协议ID消息结构
// 包含消息名称和对应的协议ID
type MessageID struct {
	// Name 消息名称，如 "Request_SearchBook"
	Name string

	// ID 协议ID，通过 Hash 生成
	ID uint32
}

