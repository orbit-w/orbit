package controller

import (
	"gitee.com/orbit-w/orbit/app/proto/pb/pb_core"
	"google.golang.org/protobuf/proto"
)

type ExampleController struct {
}

func (e *ExampleController) HandleSearchBook(req *pb_core.Request_SearchBook) proto.Message {
	return &pb_core.Request_SearchBook_Rsp{
		Result: &pb_core.Book{
			Content: "Hello, World!",
		},
	}
}

func (e *ExampleController) HandleHeartBeat(req *pb_core.Request_HeartBeat) proto.Message {
	return &pb_core.OK{}
}
