
package routers

import (
	"gitee.com/orbit-w/orbit/app/controller"
	"gitee.com/orbit-w/orbit/app/core/dispatch"
	"gitee.com/orbit-w/orbit/app/proto/pb"
	"github.com/gogo/protobuf/proto"
)

func init() {

	dispatch.Register(pb.PID_Request_SearchBook, func(data []byte) (proto.Message, uint32, error) {
		req := &pb.Request_SearchBook{}
		if err := proto.Unmarshal(data, req); err != nil {
			return nil, 0, err
		}
		return controller.GExampleController.HandleSearchBook(req), pb.PID_Request_SearchBook_Rsp, nil
	})

	dispatch.Register(pb.PID_Request_HeartBeat, func(data []byte) (proto.Message, uint32, error) {
		req := &pb.Request_HeartBeat{}
		if err := proto.Unmarshal(data, req); err != nil {
			return nil, 0, err
		}
		return controller.GExampleController.HandleHeartBeat(req), pb.PID_Request_HeartBeat_Rsp, nil
	})

}
