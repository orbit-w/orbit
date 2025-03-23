// 自动生成的代码 - 请勿手动修改
package Core

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"github.com/orbit-w/orbit/lib/utils/proto_utils"
)

// CoreRequestHandler 定义处理Core包Request消息的接口
type CoreRequestHandler interface {
	// HandleSearchBook 处理SearchBook请求
	HandleSearchBook(req *Request_SearchBook) any
	// HandleHeartBeat 处理HeartBeat请求
	HandleHeartBeat(req *Request_HeartBeat) any
}

// DispatchCoreRequest 分发Core包的请求消息到对应处理函数
func DispatchCoreRequest(handler CoreRequestHandler, msgName string, data []byte) (any, uint32, error) {
	var err error
	var response any
	var responsePid uint32

	switch msgName {
	case "SearchBook":
		req := &Request_SearchBook{}
		if err = proto.Unmarshal(data, req); err != nil {
			return nil, 0, fmt.Errorf("unmarshal SearchBook failed: %v", err)
		}
		response = handler.HandleSearchBook(req)
		responsePid = getCoreResponsePID(response)
	case "HeartBeat":
		req := &Request_HeartBeat{}
		if err = proto.Unmarshal(data, req); err != nil {
			return nil, 0, fmt.Errorf("unmarshal HeartBeat failed: %v", err)
		}
		response = handler.HandleHeartBeat(req)
		responsePid = getCoreResponsePID(response)
	default:
		return nil, 0, fmt.Errorf("unknown request message: %s", msgName)
	}

	return response, responsePid, nil
}

// DispatchCoreRequestByID 根据协议ID分发Core包的请求消息到对应处理函数
func DispatchCoreRequestByID(handler CoreRequestHandler, pid uint32, data []byte) (any, uint32, error) {
	var err error
	var response any
	var responsePid uint32

	switch pid {
	case PID_Core_SearchBook:
		req := &Request_SearchBook{}
		if err = proto.Unmarshal(data, req); err != nil {
			return nil, 0, fmt.Errorf("unmarshal SearchBook failed: %v", err)
		}
		response = handler.HandleSearchBook(req)
		responsePid = getCoreResponsePID(response)
	case PID_Core_HeartBeat:
		req := &Request_HeartBeat{}
		if err = proto.Unmarshal(data, req); err != nil {
			return nil, 0, fmt.Errorf("unmarshal HeartBeat failed: %v", err)
		}
		response = handler.HandleHeartBeat(req)
		responsePid = getCoreResponsePID(response)
	default:
		return nil, 0, fmt.Errorf("unknown request protocol ID: 0x%x", pid)
	}

	return response, responsePid, nil
}

// getCoreResponsePID 通过反射获取响应消息的协议ID
func getCoreResponsePID(response any) uint32 {
	// 获取消息名称
	typeName := proto_utils.ParseMessageName(response)
	if typeName == "" {
		return 0
	}

	// 查找类型对应的协议ID
	pid, ok := GetCoreProtocolID(typeName)
	if ok {
		return pid
	}

	// 未找到对应的协议ID
	return 0
}
