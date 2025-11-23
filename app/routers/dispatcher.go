package routers

import (
	"fmt"
	"sync"

	servicezone "gitee.com/orbit-w/orbit/core/services/service_zone"
	"google.golang.org/protobuf/proto"
)

var (
	// 全局路由器实例
	globalRouter *Router

	// 确保全局实例只初始化一次
	routerOnce sync.Once
)

// Init 初始化分发器，注册所有控制器
func Init() {
	routerOnce.Do(func() {
		globalRouter = NewRouter()
	})
}

func GetRouter() *Router {
	routerOnce.Do(func() {
		globalRouter = NewRouter()
	})
	return globalRouter
}

func Register(pid uint32, router Handler) {
	globalRouter.Register(pid, router)
}

// Dispatch 分发请求到对应的处理方法
func Dispatch(pid uint32, data []byte) Handler {
	if globalRouter == nil {
		Init()
	}

	handler := globalRouter.Dispatch(pid)
	if handler == nil {
		return nil
	}

	return handler
}

type Handler func(ctx servicezone.IContext, data []byte) (proto.Message, string, error)

type Router struct {
	funcMap map[uint32]Handler
}

func NewRouter() *Router {
	return &Router{
		funcMap: make(map[uint32]Handler),
	}
}

func (r *Router) Register(pid uint32, router Handler) {
	if _, ok := r.funcMap[pid]; ok {
		panic(fmt.Sprintf("pid %d already registered", pid))
	}

	r.funcMap[pid] = router
}

func (r *Router) Dispatch(pid uint32) Handler {
	router, ok := r.funcMap[pid]
	if !ok {
		return nil
	}

	return router
}
