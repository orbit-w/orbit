package servicezone

import (
	"gitee.com/orbit-w/orbit/internal/game/mme_agent"
	"gitee.com/orbit-w/orbit/pkg/proto/mme"
	"google.golang.org/protobuf/proto"
)

type IContext interface {
	LoadRefs(refs []*mme.EntityRef) ([]mme_agent.IEntity, error)
	Load(id int64, entityType mme.EntityType) (mme_agent.IEntity, error)
}

// IRouter 路由分发接口，用于解耦 service_zone 和 routers 之间的循环依赖
// routers 包需要实现这个接口
type IRouter interface {
	// Dispatch 根据 pid 分发请求，返回对应的处理器
	// 如果找不到对应的处理器，返回 nil
	Dispatch(pid uint32) func(ctx IContext, data []byte) (proto.Message, string, error)
}

type ServiceZoneContext struct {
	*ServiceZone
}

func NewServiceZoneContext(zone *ServiceZone) IContext {
	return &ServiceZoneContext{
		ServiceZone: zone,
	}
}
