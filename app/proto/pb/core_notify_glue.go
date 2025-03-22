// 自动生成的代码 - 请勿手动修改
package pb

import (
	"fmt"

	"google.golang.org/protobuf/proto"
)

// MarshalBeAttacked 序列化BeAttacked通知消息
func MarshalBeAttacked(notify *Notify_BeAttacked) ([]byte, error) {
	return proto.Marshal(notify)
}

// ParseCoreNotify 根据消息名称解析Core包的通知消息
func ParseCoreNotify(msgName string, msgData []byte) (any, error) {
	var err error
	var notification any

	switch msgName {
	case "BeAttacked":
		notify := &Notify_BeAttacked{}
		if err = proto.Unmarshal(msgData, notify); err != nil {
			return nil, fmt.Errorf("unmarshal BeAttacked notification failed: %v", err)
		}
		notification = notify
	default:
		return nil, fmt.Errorf("unknown notification message: %s", msgName)
	}

	return notification, nil
}
