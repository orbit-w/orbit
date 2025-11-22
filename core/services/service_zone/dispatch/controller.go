package dispatch

import (
	"gitee.com/orbit-w/orbit/app/mme"
	"gitee.com/orbit-w/orbit/app/proto/core"
	"google.golang.org/protobuf/proto"
)

type Controller struct {
}

func (c *Controller) HandleLoginRequest(req *core.Request_LoginRequest, playerEntity *mme.PlayerEntityWrapper) (proto.Message, string, error) {
	return nil, "", nil
}
