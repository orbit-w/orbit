package controller

import (
	"gitee.com/orbit-w/orbit/app/proto/pb"
	"github.com/gogo/protobuf/proto"
)

var (
	GExampleController = &ExampleController{}
)

type ExampleController struct {
}

func (e *ExampleController) HandleSearchBook(req *pb.Request_SearchBook) proto.Message {
	return &pb.Request_SearchBook_Rsp{
		Result: &pb.Book{
			Content: "Hello, World!",
		},
	}
}

func (e *ExampleController) HandleHeartBeat(req *pb.Request_HeartBeat) proto.Message {
	return &pb.Rsp_OK{}
}
