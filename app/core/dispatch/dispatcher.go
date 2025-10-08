package dispatch

import (
	"sync"

	"gitee.com/orbit-w/orbit/app/proto/pb"
	"github.com/gogo/protobuf/proto"
)

var (
	// 全局反射路由器实例
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

func Register(pid uint32, callback func(data []byte) (proto.Message, error)) {
	globalRouter.Register(pid, callback)
}

// Dispatch 分发请求到对应的处理方法
func Dispatch(pid uint32, data []byte) (proto.Message, uint32, error) {
	if globalRouter == nil {
		Init()
	}

	rsp, err := globalRouter.Dispatch(pid, data)
	if err != nil {
		return nil, 0, err
	}

	return rsp, pb.GetResponsePID(rsp), nil
}
