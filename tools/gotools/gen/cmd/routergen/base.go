package routergen

import descriptor "github.com/gogo/protobuf/protoc-gen-gogo/descriptor"

// ProtocolIDMapping 用于存储协议ID映射
type ProtocolIDMapping struct {
	PackageName string
	MessageIDs  []MessageID
}

const (
	MessageTypeRequest = iota
	MessageTypeRsp
	MessageTypeNotify
	MessageTypeStruct
)

// MessageName 用于存储消息名称和完整路径
type MessageName struct {
	Type           int
	Name           string
	PackageName    string
	PackageSuffix  string
	MsgWall        string                      //消息墙名称
	ReqFullName    string                      //请求消息完整名称
	RspFullName    string                      //响应消息完整名称
	NotifyFullName string                      //通知消息完整名称
	DP             *descriptor.DescriptorProto //消息DescriptorProto
}

func (m *MessageName) SetPackageName(packageName string) {
	m.PackageName = packageName
	m.PackageSuffix = CapitalizeFirst(packageName)
}

// IsMsgWall 判断是否为消息墙消息
func (m *MessageName) IsMsgWall() bool {
	return m.MsgWall == m.Name
}

// IsMsgWallByWall 判断是否为指定消息墙消息
func (m *MessageName) IsMsgWallByWall(msgWall string) bool {
	return m.MsgWall == msgWall
}

func (m *MessageName) GetPackageName() string {
	return m.PackageName
}

func (m *MessageName) GetReqPidName() string {
	return m.PackageName + "-" + m.ReqFullName
}

func (m *MessageName) GetRspPidName() string {
	return m.PackageName + "-" + m.RspFullName
}

func (m *MessageName) GetNotifyPidName() string {
	return m.PackageName + "-" + m.NotifyFullName
}

// MessageID 用于存储消息ID
type MessageID struct {
	Name string
	ID   uint32
}

// Message 消息结构，用于存储消息定义及其注释
type Message struct {
	Type        int
	PackageName string
	Name        string
	FullName    string // 包含父消息路径的完整名称
	Comment     string
	PidName     string
	Fields      []Field
	Response    string // 响应消息名称，如果有的话
}

// Field 字段结构，用于存储字段定义及其注释
type Field struct {
	Name    string
	Type    string
	Index   int
	Comment string
}
