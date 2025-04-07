// Code generated by protocol ID generator. DO NOT EDIT.
package pb

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"gitee.com/orbit-w/orbit/lib/utils/proto_utils"
)

// 所有协议ID常量
const (
	// Core 包协议ID
	PID_Core_Fail uint32 = 0xd03670ba // Core.Fail
	PID_Core_Notify_BeAttacked uint32 = 0x8fee7235 // Core.Notify_BeAttacked
	PID_Core_OK uint32 = 0x0ece9291 // Core.OK
	PID_Core_Request_HeartBeat uint32 = 0x95eee555 // Core.Request_HeartBeat
	PID_Core_Request_SearchBook uint32 = 0xd3ecf693 // Core.Request_SearchBook
	PID_Core_Request_SearchBook_Rsp uint32 = 0xf1d19d0a // Core.Request_SearchBook_Rsp

	// Season 包协议ID
	PID_Season_Request_SeasonInfo uint32 = 0xd9714656 // Season.Request_SeasonInfo

)

// AllMessageNameToID 全局消息名称到ID的映射
var AllMessageNameToID = map[string]uint32{
	"Core-Fail": PID_Core_Fail,
	"Core-Notify_BeAttacked": PID_Core_Notify_BeAttacked,
	"Core-OK": PID_Core_OK,
	"Core-Request_HeartBeat": PID_Core_Request_HeartBeat,
	"Core-Request_SearchBook": PID_Core_Request_SearchBook,
	"Core-Request_SearchBook_Rsp": PID_Core_Request_SearchBook_Rsp,
	"Season-Request_SeasonInfo": PID_Season_Request_SeasonInfo,
}

// AllIDToMessageName 全局ID到消息名称的映射
var AllIDToMessageName = map[uint32]string{
	PID_Core_Fail: "Core-Fail",
	PID_Core_Notify_BeAttacked: "Core-Notify_BeAttacked",
	PID_Core_OK: "Core-OK",
	PID_Core_Request_HeartBeat: "Core-Request_HeartBeat",
	PID_Core_Request_SearchBook: "Core-Request_SearchBook",
	PID_Core_Request_SearchBook_Rsp: "Core-Request_SearchBook_Rsp",
	PID_Season_Request_SeasonInfo: "Season-Request_SeasonInfo",
}

// MessagePackageMap 消息名称到包名的映射
var MessagePackageMap = map[string]string{
	"Fail": "Core",
	"Notify_BeAttacked": "Core",
	"OK": "Core",
	"Request_HeartBeat": "Core",
	"Request_SearchBook": "Core",
	"Request_SearchBook_Rsp": "Core",
	"Request_SeasonInfo": "Season",
}

// GetProtocolID 获取指定消息名称的协议ID
func GetProtocolID(messageName string) (uint32, bool) {
	id, ok := AllMessageNameToID[messageName]
	return id, ok
}

// GetMessageName 获取指定协议ID的消息名称
func GetMessageName(pid uint32) (string, bool) {
	name, ok := AllIDToMessageName[pid]
	return name, ok
}

// GetResponsePID 获取响应消息的协议ID
func GetResponsePID(response proto.Message) uint32 {
	if response == nil {
		return 0
	}

	messageName := proto_utils.ParseMessageName(response)
	if messageName == "" {
		return 0
	}

	// 从映射表中查找包名
	packageName, ok := MessagePackageMap[messageName]
	if !ok {
		// 找不到包名直接panic
		panic(fmt.Sprintf("消息 %s 未在映射表中找到对应的包名", messageName))
	}

	fullName := packageName + "-" + messageName
	pid, _ := GetProtocolID(fullName)
	return pid
}
