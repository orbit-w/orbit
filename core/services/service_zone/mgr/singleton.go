package servicezone_mgr

import (
	"sync"
	"time"

	"gitee.com/orbit-w/orbit/core/network"
	"github.com/asynkron/protoactor-go/actor"
)

var (
	globalManager *ZoneManager
	once          sync.Once
)

func GetZoneManager() *ZoneManager {
	once.Do(func() {
		globalManager = NewZoneManager()
	})
	return globalManager
}

func ClientRequest(zoneId string, originRequest network.IClientRequest) error {
	return GetZoneManager().ClientRequest(zoneId, originRequest)
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
