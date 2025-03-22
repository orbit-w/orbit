// 自动生成的代码 - 请勿手动修改
package pb

import (
	"fmt"
	"google.golang.org/protobuf/proto"
)

// MarshalBeAttacked 序列化BeAttacked通知消息
func MarshalBeAttacked(notify *Notify_BeAttacked) ([]byte, uint32, error) {
	data, err := proto.Marshal(notify)
	return data, PID_Core_BeAttacked, err
}

// ParseCoreNotify 根据消息名称解析Core包的通知消息
func ParseCoreNotify(msgName string, data []byte) (any, error) {
	var err error
	var notification any

	switch msgName {
	case "BeAttacked":
		notify := &Notify_BeAttacked{}
		if err = proto.Unmarshal(data, notify); err != nil {
			return nil, fmt.Errorf("unmarshal BeAttacked notification failed: %v", err)
		}
		notification = notify
	default:
		return nil, fmt.Errorf("unknown notification message: %s", msgName)
	}

	return notification, nil
}

// ParseCoreNotifyByID 根据协议ID解析Core包的通知消息
func ParseCoreNotifyByID(pid uint32, data []byte) (any, error) {
	var err error
	var notification any

	switch pid {
	case PID_Core_BeAttacked:
		notify := &Notify_BeAttacked{}
		if err = proto.Unmarshal(data, notify); err != nil {
			return nil, fmt.Errorf("unmarshal notification with ID 0x%08x failed: %v", pid, err)
		}
		notification = notify
	default:
		return nil, fmt.Errorf("unknown notification protocol ID: 0x%08x", pid)
	}

	return notification, nil
}
