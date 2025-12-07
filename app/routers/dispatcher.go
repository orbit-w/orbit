package routers

import (
	"fmt"
	"sync"

	mmeobj "gitee.com/orbit-w/orbit/app/mme"
	servicezone_behavior "gitee.com/orbit-w/orbit/core/services/service_zone/behavior"
	"google.golang.org/protobuf/proto"
)

var (
	// 全局路由器实例
	globalRouter *Routers

	// 确保全局实例只初始化一次
	routerOnce sync.Once
)

// Init 初始化分发器，注册所有控制器
func init() {
	routerOnce.Do(func() {
		globalRouter = NewRouter()
	})
}

func GetRouter() *Routers {
	routerOnce.Do(func() {
		globalRouter = NewRouter()
	})
	return globalRouter
}

func RegisterHandler(pid uint32, router func(ctx servicezone_behavior.IContext, req proto.Message, entities ...mmeobj.IEntity) (proto.Message, string, error)) {
	globalRouter.RegisterHandler(pid, router)
}

type Routers struct {
	funcMap map[uint32]func(ctx servicezone_behavior.IContext, req proto.Message, entities ...mmeobj.IEntity) (proto.Message, string, error)
}

func NewRouter() *Routers {
	return &Routers{
		funcMap: make(map[uint32]func(ctx servicezone_behavior.IContext, req proto.Message, entities ...mmeobj.IEntity) (proto.Message, string, error)),
	}
}

func (r *Routers) RegisterHandler(pid uint32, router func(ctx servicezone_behavior.IContext, req proto.Message, entities ...mmeobj.IEntity) (proto.Message, string, error)) {
	if _, ok := r.funcMap[pid]; ok {
		panic(fmt.Sprintf("pid %d already registered", pid))
	}

	r.funcMap[pid] = router
}

func (r *Routers) Dispatch(pid uint32) func(ctx servicezone_behavior.IContext, req proto.Message, entities ...mmeobj.IEntity) (proto.Message, string, error) {
	handler, ok := r.funcMap[pid]
	if !ok {
		return nil
	}

	return handler
}
