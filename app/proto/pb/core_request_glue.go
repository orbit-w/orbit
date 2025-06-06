// Code generated by genproto. DO NOT EDIT.
package pb

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"gitee.com/orbit-w/orbit/app/proto/pb/pb_core"
)

// CoreRequestHandler 处理Core包的请求消息
type CoreRequestHandler interface {
	// HandleSearchBook 处理SearchBook请求
	HandleSearchBook(req *pb_core.Request_SearchBook) proto.Message
	// HandleHeartBeat 处理HeartBeat请求
	HandleHeartBeat(req *pb_core.Request_HeartBeat) proto.Message
}

// DispatchCoreRequestByID 根据协议ID分发请求到对应处理函数
func DispatchCoreRequestByID(handler CoreRequestHandler, pid uint32, data []byte) (proto.Message, uint32, error) {
	var response proto.Message
	switch pid {
	case PID_Core_Request_SearchBook: // Request_SearchBook
		req := &pb_core.Request_SearchBook{}
		if err := proto.Unmarshal(data, req); err != nil {
			return nil, 0, fmt.Errorf("unmarshal Request_SearchBook failed: %w", err)
		}

		response = handler.HandleSearchBook(req)
	
	case PID_Core_Request_HeartBeat: // Request_HeartBeat
		req := &pb_core.Request_HeartBeat{}
		if err := proto.Unmarshal(data, req); err != nil {
			return nil, 0, fmt.Errorf("unmarshal Request_HeartBeat failed: %w", err)
		}

		response = handler.HandleHeartBeat(req)
	
	default:
		return nil, 0, fmt.Errorf("unknown request protocol ID: 0x%08x", pid)
	}

	// 使用公共映射文件获取响应ID
	responsePid := GetResponsePID(response)
	return response, responsePid, nil
}
