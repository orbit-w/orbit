package protocolids

import (
	"fmt"
	"sort"

	"gitee.com/orbit-w/orbit/lib/base/protoid"
	"gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/pb_gen/net_message"
)

// Generator 协议ID生成器
// 负责从 NetWall 定义中收集 Request、Response 和 Notify 消息，并生成 protocol_ids.pb.go 文件
type Generator struct {
	netWalls []net_message.NetWallFile
}

// NewGenerator 创建新的协议ID生成器
// netWalls: NetWall 文件列表
func NewGenerator(netWalls []net_message.NetWallFile) *Generator {
	return &Generator{
		netWalls: netWalls,
	}
}

// Generate 生成 protocol_ids.pb.go 文件
// 收集所有 NetWall 中的 Request 和 Response，生成一个统一的 protocol_ids.pb.go 文件
// outputDir: 输出目录，通常是 pkg/proto/pb
func (g *Generator) Generate(outputDir string) error {
	// 收集所有 NetWall 包的消息
	allMessages := g.collectAllNetWallMessages()

	// 如果没有任何消息，跳过
	if len(allMessages) == 0 {
		return nil
	}

	// 添加通用的 Response 消息（Rsp_OK 和 Error）
	g.addCommonResponseMessages(&allMessages)

	// 生成协议ID
	messageIDs := g.generateProtocolIDs(allMessages)

	// 生成文件，输出到 outputDir 目录（通常是 pkg/proto/pb）
	return g.generateProtocolIDsFile(messageIDs, outputDir)
}

// collectAllNetWallMessages 收集所有 NetWall 包中的消息
func (g *Generator) collectAllNetWallMessages() []MessageInfo {
	var allMessages []MessageInfo

	for _, netwallFile := range g.netWalls {
		messages := g.collectNetWallMessages(netwallFile)
		allMessages = append(allMessages, messages...)
	}

	return allMessages
}

// collectNetWallMessages 收集指定 NetWall 包中的 Request、Response 和 Notify 消息
func (g *Generator) collectNetWallMessages(netwallFile net_message.NetWallFile) []MessageInfo {
	var messages []MessageInfo
	packageName := netwallFile.GetPackageName()

	// 收集 Request 消息
	for _, req := range netwallFile.GetRequests() {
		messages = append(messages, g.createRequestMessageInfo(packageName, req))

		// 如果有 Response，收集 Response 消息
		if req.GetResponse() != nil {
			messages = append(messages, g.createResponseMessageInfo(packageName, req))
		}
	}

	// 收集 Notify 消息
	for _, notify := range netwallFile.GetNotifies() {
		messages = append(messages, g.createNotifyMessageInfo(packageName, notify))
	}

	return messages
}

// createRequestMessageInfo 创建 Request 消息信息
func (g *Generator) createRequestMessageInfo(packageName string, req *net_message.NetMessage) MessageInfo {
	// Request 消息的 FullName 格式: Request_{RequestName}
	reqFullName := req.GetFullName()
	// Request 消息的 PidName 格式: {PackageName}-Request_{RequestName}
	reqPidName := fmt.Sprintf("%s-%s", packageName, reqFullName)

	return MessageInfo{
		FullName: reqFullName,
		PidName:  reqPidName,
	}
}

// createResponseMessageInfo 创建 Response 消息信息
func (g *Generator) createResponseMessageInfo(packageName string, req *net_message.NetMessage) MessageInfo {
	// Response 消息的 FullName 格式: Request_{RequestName}_Rsp
	rspFullName := req.GetResponse().GetFullName()
	// Response 消息的 PidName 格式: {PackageName}-Request_{RequestName}_Rsp
	rspPidName := fmt.Sprintf("%s-%s", packageName, rspFullName)

	return MessageInfo{
		FullName: rspFullName,
		PidName:  rspPidName,
	}
}

// createNotifyMessageInfo 创建 Notify 消息信息
func (g *Generator) createNotifyMessageInfo(packageName string, notify *net_message.NetMessage) MessageInfo {
	// Notify 消息的 FullName 格式: Notify_{NotifyName}
	notifyFullName := notify.GetFullName()
	// Notify 消息的 PidName 格式: {PackageName}-Notify_{NotifyName}
	notifyPidName := fmt.Sprintf("%s-%s", packageName, notifyFullName)

	return MessageInfo{
		FullName: notifyFullName,
		PidName:  notifyPidName,
	}
}

// addCommonResponseMessages 添加通用的 Response 消息（Rsp_OK 和 Error）
func (g *Generator) addCommonResponseMessages(messages *[]MessageInfo) {
	hasRspOK := false
	hasError := false

	for _, msg := range *messages {
		if msg.FullName == "OK" {
			hasRspOK = true
		}
		if msg.FullName == "Error" {
			hasError = true
		}
	}

	// 如果没有找到，添加通用的 Response 消息
	if !hasRspOK {
		*messages = append(*messages, MessageInfo{
			FullName: "OK",
			PidName:  "Core-OK",
		})
	}
	if !hasError {
		*messages = append(*messages, MessageInfo{
			FullName: "Error",
			PidName:  "Core-Error",
		})
	}
}

// generateProtocolIDs 生成协议ID
func (g *Generator) generateProtocolIDs(messages []MessageInfo) []MessageID {
	messageIDs := make([]MessageID, 0, len(messages))

	for _, msg := range messages {
		// 使用 protoid.HashProtoMessage 生成协议ID，使用 PidName 格式
		pid := protoid.HashProtoMessage(msg.PidName)
		messageIDs = append(messageIDs, MessageID{
			Name: msg.FullName, // 使用 FullName 作为消息名称
			ID:   pid,
		})
	}

	// 按名称排序以得到一致的输出
	sort.Slice(messageIDs, func(i, j int) bool {
		return messageIDs[i].Name < messageIDs[j].Name
	})

	return messageIDs
}
