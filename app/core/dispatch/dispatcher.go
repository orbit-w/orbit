package dispatch

import (
	"github.com/orbit-w/orbit/app/controller"
	"github.com/orbit-w/orbit/app/proto/pb"
)

func Dispatch(pid uint32, data []byte) (any, uint32, error) {
	return pb.DispatchCoreRequestByID(controller.GlobalManager(), pid, data)
}

func DispatchByName(msgName string, data []byte) (any, uint32, error) {
	return pb.DispatchCoreRequest(controller.GlobalManager(), msgName, data)
}
