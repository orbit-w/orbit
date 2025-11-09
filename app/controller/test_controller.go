package controller

import (
	"gitee.com/orbit-w/orbit/app/proto/pb"
	"google.golang.org/protobuf/proto"
)

var (
	GExampleController = &ExampleController{}
)

type ExampleController struct {
}

//go:generate go run ../../tools/gotools/gen/main.go routergen --controller-dir=. --output-dir=../routers --debug=true
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
