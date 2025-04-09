package dispatch

import (
	"gitee.com/orbit-w/orbit/app/controller"
	"gitee.com/orbit-w/orbit/app/proto/pb"
	"google.golang.org/protobuf/proto"
)

func Dispatch(pid uint32, data []byte) (proto.Message, uint32, error) {
	return pb.DispatchCoreRequestByID(controller.GlobalManager(), pid, data)
}
