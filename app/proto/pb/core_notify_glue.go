// 自动生成的代码 - 请勿手动修改
package pb

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"github.com/orbit-w/orbit/lib/utils/proto_utils"
)

// Base notification message type that doesn't need protocol ID
const (
	PID_Core_Notify uint32 = 0
)

// getCoreNotificationPID 通过反射获取通知消息的协议ID
func getCoreNotificationPID(notification any) uint32 {
	// 对基础消息类型特殊处理
	if notification == nil {
		return 0
	}

	// 获取消息名称
	typeName := proto_utils.ParseMessageName(notification)
	if typeName == "" {
		return 0
	}

	// 基础消息类型特殊处理
	if typeName == "Notify" {
		return PID_Core_Notify
	}

	// 查找类型对应的协议ID
	pid, ok := GetCoreProtocolID(typeName)
	if ok {
		return pid
	}

	// 未找到对应的协议ID
	return 0
}

// MarshalBeAttacked 序列化BeAttacked通知消息
func MarshalBeAttacked(notify *Notify_BeAttacked) ([]byte, uint32, error) {
	data, err := proto.Marshal(notify)
	return data, PID_Core_Notify_BeAttacked, err
}

// ParseCoreNotify 根据消息名称解析Core包的通知消息
func ParseCoreNotify(msgName string, data []byte) (any, uint32, error) {
	var err error
	var notification any
	var notificationPid uint32

	switch msgName {
	case "Notify_BeAttacked":
		notify := &Notify_BeAttacked{}
		if err = proto.Unmarshal(data, notify); err != nil {
			return nil, 0, fmt.Errorf("unmarshal BeAttacked notification failed: %v", err)
		}
		notification = notify
		notificationPid = getCoreNotificationPID(notification)
	default:
		return nil, 0, fmt.Errorf("unknown notification message: %s", msgName)
	}

	return notification, notificationPid, nil
}

// ParseCoreNotifyByID 根据协议ID解析Core包的通知消息
func ParseCoreNotifyByID(pid uint32, data []byte) (any, uint32, error) {
	var err error
	var notification any
	var notificationPid uint32

	switch pid {
	case PID_Core_Notify:
		// 基础通知类型特殊处理
		return nil, 0, fmt.Errorf("cannot unmarshal base Notify type directly")

	case PID_Core_Notify_BeAttacked:
		notify := &Notify_BeAttacked{}
		if err = proto.Unmarshal(data, notify); err != nil {
			return nil, 0, fmt.Errorf("unmarshal notification with ID 0x%08x failed: %v", pid, err)
		}
		notification = notify
		notificationPid = getCoreNotificationPID(notification)
	default:
		return nil, 0, fmt.Errorf("unknown notification protocol ID: 0x%08x", pid)
	}

	return notification, notificationPid, nil
}
