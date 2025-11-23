package servicezone

import (
	"sync"
	"time"

	"github.com/asynkron/protoactor-go/actor"
)

var (
	globalManager *ZoneManager
	once          sync.Once
	globalRouter  IRouter
)

func GetZoneManager() *ZoneManager {
	once.Do(func() {
		globalManager = NewZoneManager()
	})
	return globalManager
}

// InitWithRouter 初始化 ZoneManager 并设置路由分发器
// 应该在应用启动时调用，在注册路由之后
func InitWithRouter(router IRouter) {
	globalRouter = router
}

func Cast(zoneId string, msg any) error {
	return GetZoneManager().Cast(zoneId, msg)
}

func Call(zoneId string, req any, timeout ...time.Duration) (*actor.Future, error) {
	return GetZoneManager().Call(zoneId, req, timeout...)
}

func Stop() {
	GetZoneManager().Stop()
}
