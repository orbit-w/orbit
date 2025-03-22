// 自动生成的代码 - 请勿手动修改
package pb

import (
	"fmt"
	"google.golang.org/protobuf/proto"
)

// CoreRequestHandler 定义处理Core包Request消息的接口
type CoreRequestHandler interface {
	// HandleSearchBook 处理SearchBook请求
	HandleSearchBook(req *Request_SearchBook) any
	// HandleHeartBeat 处理HeartBeat请求
	HandleHeartBeat(req *Request_HeartBeat) any
}

// DispatchCoreRequest 分发Core包的请求消息到对应处理函数
func DispatchCoreRequest(handler CoreRequestHandler, msgName string, msgData []byte) (any, error) {
	var err error
	var response any

	switch msgName {
	case "SearchBook":
		req := &Request_SearchBook{}
		if err = proto.Unmarshal(msgData, req); err != nil {
			return nil, fmt.Errorf("unmarshal SearchBook failed: %v", err)
		}
		response = handler.HandleSearchBook(req)
	case "HeartBeat":
		req := &Request_HeartBeat{}
		if err = proto.Unmarshal(msgData, req); err != nil {
			return nil, fmt.Errorf("unmarshal HeartBeat failed: %v", err)
		}
		response = handler.HandleHeartBeat(req)
	default:
		return nil, fmt.Errorf("unknown request message: %s", msgName)
	}

	return response, nil
}
