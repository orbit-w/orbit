package controller

import (
	"github.com/orbit-w/orbit/app/proto/pb"
)

type ExampleController struct {
}

func (e *ExampleController) HandleSearchBook(req *pb.Request_SearchBook) any {
	return nil
}

func (e *ExampleController) HandleHeartBeat(req *pb.Request_HeartBeat) any {
	return nil
}
