package servicezone_mgr

import (
	"fmt"
	"time"

	"gitee.com/orbit-w/orbit/core/network"
	zone_meta "gitee.com/orbit-w/orbit/core/services/service_zone/meta"
	servicezone "gitee.com/orbit-w/orbit/core/services/service_zone/zone"
	"github.com/asynkron/protoactor-go/actor"
)

var (
	globalManager *ZoneManager
)

func GetZoneManager() *ZoneManager {
	return globalManager
}

func StartZoneWithMeta(zoneId string, meta *zone_meta.ZoneMeta) (*actor.PID, error) {
	return GetZoneManager().StartZoneWithMeta(zoneId, meta)
}

func GetZone(zoneId string) (*actor.PID, error) {
	return GetZoneManager().GetZone(zoneId)
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

func GenLocalZoneId(pattern servicezone.ZonePattern, serverId string) string {
	return fmt.Sprintf("%d_%s", int32(pattern), serverId)
}
