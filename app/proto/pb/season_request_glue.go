// Code generated by genproto. DO NOT EDIT.
package pb

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"gitee.com/orbit-w/orbit/app/proto/pb/pb_season"
)

// SeasonRequestHandler 处理Season包的请求消息
type SeasonRequestHandler interface {
	// HandleSeasonInfo 处理SeasonInfo请求
	HandleSeasonInfo(req *pb_season.Request_SeasonInfo) proto.Message
}

// DispatchSeasonRequestByID 根据协议ID分发请求到对应处理函数
func DispatchSeasonRequestByID(handler SeasonRequestHandler, pid uint32, data []byte) (proto.Message, uint32, error) {
	var response proto.Message
	switch pid {
	case PID_Season_Request_SeasonInfo: // Request_SeasonInfo
		req := &pb_season.Request_SeasonInfo{}
		if err := proto.Unmarshal(data, req); err != nil {
			return nil, 0, fmt.Errorf("unmarshal Request_SeasonInfo failed: %w", err)
		}

		response = handler.HandleSeasonInfo(req)
	
	default:
		return nil, 0, fmt.Errorf("unknown request protocol ID: 0x%08x", pid)
	}

	// 使用公共映射文件获取响应ID
	responsePid := GetResponsePID(response)
	return response, responsePid, nil
}
