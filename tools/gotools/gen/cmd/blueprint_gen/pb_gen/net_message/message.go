package net_message

import (
	"fmt"

	types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"
)

// NetWallMessageType NetWall 消息类型
type NetWallMessageType int

const (
	ResponseSuffix = "_Rsp"
	RequestPrefix  = "Request_"
	NotifyPrefix   = "Notify_"
)

const (
	NetWallMessageTypeRequest NetWallMessageType = iota
	NetWallMessageTypeNotify
	NetWallMessageTypeDataStruct
	NetWallMessageTypeResponse
)

func (nt NetWallMessageType) String() string {
	switch nt {
	case NetWallMessageTypeRequest:
		return "Request"
	case NetWallMessageTypeNotify:
		return "Notify"
	case NetWallMessageTypeDataStruct:
		return "DataStruct"
	case NetWallMessageTypeResponse:
		return "Response"
	}
	return ""
}

// NetMessage NetWall 消息定义
type NetMessage struct {
	Type        NetWallMessageType
	Name        string
	FullName    string
	PackageName string
	Fields      []*types.Field
	Comment     string

	// 扩展字段
	RspExt *NetMessage // 仅用于 Request
}

func NewNetMessage(nt NetWallMessageType, name string) *NetMessage {
	return &NetMessage{
		Type:     nt,
		Fields:   make([]*types.Field, 0),
		RspExt:   nil,
		Comment:  "",
		Name:     name,
		FullName: GenFullName(nt, name),
	}
}

func GenFullName(nt NetWallMessageType, name string) string {
	switch nt {
	case NetWallMessageTypeRequest:
		return RequestPrefix + name
	case NetWallMessageTypeNotify:
		return NotifyPrefix + name
	case NetWallMessageTypeResponse:
		panic(fmt.Sprintf("消息类型 %s 不支持生成完整名称", nt.String()))
	default:
		return name
	}
}

func (m *NetMessage) IsRequest() bool {
	return m.Type == NetWallMessageTypeRequest
}

func (m *NetMessage) IsNotify() bool {
	return m.Type == NetWallMessageTypeNotify
}

func (m *NetMessage) IsDataStruct() bool {
	return m.Type == NetWallMessageTypeDataStruct
}

func (m *NetMessage) IsResponse() bool {
	return m.Type == NetWallMessageTypeResponse
}

// GetName 获取消息名称
func (m *NetMessage) GetName() string {
	return m.Name
}

// GetResponse 获取响应消息
func (m *NetMessage) GetResponse() *NetMessage {
	if m.RspExt == nil {
		return nil
	}
	return m.RspExt
}

func (m *NetMessage) HasResponse() bool {
	return m.RspExt != nil
}

// GetType 获取消息类型
func (m *NetMessage) GetType() NetWallMessageType {
	return m.Type
}

// GetPackageName 获取包名
func (m *NetMessage) GetPackageName() string {
	return m.PackageName
}

// GetFields 获取字段列表
func (m *NetMessage) GetFields() []*types.Field {
	return m.Fields
}

func (m *NetMessage) GetFullName() string {
	return m.FullName
}

func (m *NetMessage) GenRspFullName() string {
	if !m.IsRequest() {
		panic(fmt.Sprintf("消息类型 %s 不支持生成响应消息", m.Type.String()))
	}
	return m.FullName + ResponseSuffix
}

func (m *NetMessage) AddResponse() *NetMessage {
	if m.Type != NetWallMessageTypeRequest {
		panic(fmt.Sprintf("消息类型 %s 不支持添加响应", m.Type.String()))
	}
	rsp := &NetMessage{
		Type:        NetWallMessageTypeResponse,
		PackageName: m.PackageName,
		Name:        m.GetName(), // 响应消息名称与请求消息名称相同
		FullName:    m.GenRspFullName(),
		Fields:      make([]*types.Field, 0),
	}

	m.SetResponse(rsp)
	return rsp
}

// 写方法
func (m *NetMessage) SetResponse(rsp *NetMessage) {
	m.RspExt = rsp
}

func (m *NetMessage) SetFields(fields []*types.Field) {
	m.Fields = fields
}
