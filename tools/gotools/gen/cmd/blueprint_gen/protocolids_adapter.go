package blueprint_gen

import (
	"gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/protocolids"
)

// netWallFileAdapter 适配器，将 blueprint_gen.NetWallFile 转换为 protocolids.NetWallFile
type netWallFileAdapter struct {
	file *NetWallFile
}

// GetPackageName 获取包名
func (a *netWallFileAdapter) GetPackageName() string {
	return a.file.PackageName
}

// GetRequests 获取请求列表
func (a *netWallFileAdapter) GetRequests() []protocolids.NetMessage {
	requests := make([]protocolids.NetMessage, len(a.file.Requests))
	for i, req := range a.file.Requests {
		requests[i] = &netMessageAdapter{msg: req}
	}
	return requests
}

// GetNotifies 获取通知列表
func (a *netWallFileAdapter) GetNotifies() []protocolids.NetMessage {
	notifies := make([]protocolids.NetMessage, len(a.file.Notifies))
	for i, notify := range a.file.Notifies {
		notifies[i] = &netMessageAdapter{msg: notify}
	}
	return notifies
}

// netMessageAdapter 适配器，将 blueprint_gen.NetMessage 转换为 protocolids.NetMessage
type netMessageAdapter struct {
	msg *NetMessage
}

// GetName 获取消息名称
func (a *netMessageAdapter) GetName() string {
	return a.msg.Name
}

// GetResponse 获取响应消息
func (a *netMessageAdapter) GetResponse() protocolids.NetMessage {
	if a.msg.Rsp == nil {
		return nil
	}
	return &netMessageAdapter{msg: a.msg.Rsp}
}

// newProtocolIDGenerator 从 BlueprintContext 创建协议ID生成器
func newProtocolIDGenerator(ctx *BlueprintContext) *protocolids.Generator {
	netWalls := make([]protocolids.NetWallFile, len(ctx.NetWalls))
	for i, netwall := range ctx.NetWalls {
		netWalls[i] = &netWallFileAdapter{file: netwall}
	}
	return protocolids.NewGenerator(netWalls)
}

