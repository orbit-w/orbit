package servicezone_behavior

import (
	"gitee.com/orbit-w/orbit/internal/game/mme_agent"
	"google.golang.org/protobuf/proto"
)

// IRouter 路由分发接口，用于解耦 service_zone 和 routers 之间的循环依赖
// routers 包需要实现这个接口
type IRouter interface {
	// Dispatch 根据 pid 分发请求，返回对应的处理器
	// 如果找不到对应的处理器，返回 nil
	Dispatch(pid uint32) func(ctx IContext, req proto.Message, entities ...mme_agent.IEntity) (proto.Message, string, error)
}
