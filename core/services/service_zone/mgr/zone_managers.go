package servicezone_mgr

import (
	"fmt"
	"time"

	"gitee.com/orbit-w/orbit/core/network"
	servicezone_behavior "gitee.com/orbit-w/orbit/core/services/service_zone/behavior"
	zone_meta "gitee.com/orbit-w/orbit/core/services/service_zone/meta"
	servicezone "gitee.com/orbit-w/orbit/core/services/service_zone/zone"
	"gitee.com/orbit-w/orbit/lib/module/unipue_task_exec"
	"github.com/asynkron/protoactor-go/actor"
	cmap "github.com/orcaman/concurrent-map/v2"
)

type ZoneManager struct {
	system *actor.ActorSystem
	cache  cmap.ConcurrentMap[string, *actor.PID]
	exec   *unipue_task_exec.UniqueTaskExecutor
}

func NewZoneManager() *ZoneManager {
	globalManager = &ZoneManager{
		system: actor.NewActorSystem(),
		cache:  cmap.New[*actor.PID](),
		exec:   unipue_task_exec.NewUniqueTaskExecutor(),
	}
	return globalManager
}

func (z *ZoneManager) Start() error {
	return nil
}

func (m *ZoneManager) Stop() error {
	for _, pid := range m.cache.Items() {
		servicezone_behavior.Stop(m.system, pid)
	}
	m.cache.Clear()
	m.system.Shutdown()
	return nil
}

func (z *ZoneManager) ClientRequest(zoneId string, originRequest network.IClientRequest) error {
	return z.Cast(zoneId, &servicezone_behavior.ClientRequest{
		IClientRequest: originRequest,
		ZoneId:         zoneId,
	})
}

func (z *ZoneManager) Cast(zoneId string, msg any) error {
	pid, err := z.GetZone(zoneId)
	if err != nil {
		return fmt.Errorf("zone %s not found: %w", zoneId, err)
	}

	z.system.Root.Send(pid, msg)
	return nil
}

func (z *ZoneManager) Call(zoneId string, req any, timeout ...time.Duration) (*actor.Future, error) {
	pid, err := z.GetZone(zoneId)
	if err != nil {
		return nil, fmt.Errorf("zone %s not found: %w", zoneId, err)
	}

	future := z.system.Root.RequestFuture(pid, req, parseTimeout(timeout...))
	return future, nil
}

func (z *ZoneManager) StartZoneWithMeta(zoneId string, meta *zone_meta.ZoneMeta) (*actor.PID, error) {
	return z.Load(zoneId, meta)
}

func (m *ZoneManager) GetZone(zoneId string) (*actor.PID, error) {
	if v, exists := m.cache.Get(zoneId); exists {
		return v, nil
	}
	meta, err := zone_meta.GetZoneMeta(zoneId)
	if err != nil {
		return nil, fmt.Errorf("failed to get zone meta for zoneId %s: %w", zoneId, err)
	}

	return m.Load(zoneId, meta)
}

// Load 加载或创建 Zone Actor
// zoneId: 服务区 ID
// zoneType: 服务区类型（可选，如果为 nil 则使用默认类型 ZoneTypeWild）
// 返回: Zone Actor 的 PID
func (m *ZoneManager) Load(zoneId string, zoneMeta *zone_meta.ZoneMeta) (*actor.PID, error) {
	// 使用 ExecuteOnce 确保并发安全
	re := m.exec.ExecuteOnce(zoneId, func() any {
		// 创建 ServiceZone
		zone := servicezone.NewServiceZone(zoneId, zoneMeta)

		// 启动 Zone Actor，传递 router
		pid, err := servicezone_behavior.Start(m.system, servicezone.GenActorId(zoneId), zone)
		if err != nil {
			return fmt.Errorf("failed to start zone actor for zoneId %s: %w", zoneId, err)
		}

		// 缓存 PID
		m.cache.Set(zoneId, pid)
		return pid
	})

	switch v := re.(type) {
	case error:
		return nil, v
	case *actor.PID:
		return v, nil
	default:
		return nil, fmt.Errorf("unknown error: %v", re)
	}
}

func (m *ZoneManager) ZoneExists(zoneId string) bool {
	return m.cache.Has(zoneId)
}

func parseTimeout(timeout ...time.Duration) time.Duration {
	if len(timeout) > 0 {
		return timeout[0]
	}
	return 5 * time.Second
}
