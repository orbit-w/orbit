package pb

import (
	"fmt"

	"gitee.com/orbit-w/orbit/app/proto/core"
	"gitee.com/orbit-w/orbit/app/proto/mme"
	"gitee.com/orbit-w/orbit/app/proto/sample"
	"google.golang.org/protobuf/proto"
)

type pbFactory func() proto.Message

var (
	pbFactories = make(map[uint32]pbFactory)
)

func init() {
	RegisterPBFactory(PID_Request_LoginRequest, func() proto.Message {
		return &core.Request_LoginRequest{}
	})
	RegisterPBFactory(PID_Request_AskLevelUp, func() proto.Message {
		return &mme.Request_AskLevelUp{}
	})
	RegisterPBFactory(PID_Request_SearchBook, func() proto.Message {
		return &core.Request_SearchBook{}
	})
	RegisterPBFactory(PID_Request_SearchNewsPaper, func() proto.Message {
		return &sample.Request_SearchNewsPaper{}
	})
	RegisterPBFactory(PID_Request_HeartBeat, func() proto.Message {
		return &core.Request_HeartBeat{}
	})
}

func RegisterPBFactory(pid uint32, factory pbFactory) {
	if _, ok := pbFactories[pid]; ok {
		panic(fmt.Sprintf("create pb from request already registered for pid: %d", pid))
	}
	pbFactories[pid] = factory
}

func GetPBFactory(pid uint32) pbFactory {
	return pbFactories[pid]
}
