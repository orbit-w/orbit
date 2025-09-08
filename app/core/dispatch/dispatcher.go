package dispatch

import (
	"sync"

	"github.com/gogo/protobuf/proto"
)

var (
	// 全局反射路由器实例
	globalRouter *ReflectionRouter

	// 确保全局实例只初始化一次
	routerOnce sync.Once
)

// Init 初始化分发器，注册所有控制器
func Init() {
	routerOnce.Do(func() {
		globalRouter = NewReflectionRouter()
	})
}

func RegisterController(controller any) error {
	return globalRouter.RegisterController(controller)
}

// Dispatch 分发请求到对应的处理方法
func Dispatch(pid uint32, data []byte) (proto.Message, uint32, error) {
	if globalRouter == nil {
		Init()
	}

	// 方式1: 直接使用反射路由器的Dispatch方法（推荐）
	return globalRouter.Dispatch(pid, data)
}

// GetGlobalRouter 获取全局路由器实例（用于调试）
func GetGlobalRouter() *ReflectionRouter {
	return globalRouter
}
