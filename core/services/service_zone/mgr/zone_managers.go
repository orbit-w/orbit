package servicezone_mgr

import (
	"fmt"
	"time"

	"gitee.com/orbit-w/orbit/core/network"
	servicezone_behavior "gitee.com/orbit-w/orbit/core/services/service_zone/behavior"
	servicezone "gitee.com/orbit-w/orbit/core/services/service_zone/zone"
	"gitee.com/orbit-w/orbit/lib/module/unipue_task_exec"
	"github.com/asynkron/protoactor-go/actor"
	cmap "github.com/orcaman/concurrent-map"
)

type ZoneManager struct {
	system *actor.ActorSystem
	cache  cmap.ConcurrentMap
	exec   *unipue_task_exec.UniqueTaskExecutor
}

func NewZoneManager() *ZoneManager {
	return &ZoneManager{
		system: actor.NewActorSystem(),
		cache:  cmap.New(),
		exec:   unipue_task_exec.NewUniqueTaskExecutor(),
	}
}

func (z *ZoneManager) ClientRequest(zoneId string, originRequest network.IClientRequest) error {
	return z.Cast(zoneId, &servicezone_behavior.ClientRequest{
		IClientRequest: originRequest,
		ZoneId:         zoneId,
	})
}

func (z *ZoneManager) Cast(zoneId string, msg any) error {
	pid, err := z.Load(zoneId)
	if err != nil {
		return fmt.Errorf("zone %s not found: %w", zoneId, err)
	}

	z.system.Root.Send(pid, msg)
	return nil
}

func (z *ZoneManager) Call(zoneId string, req any, timeout ...time.Duration) (*actor.Future, error) {
	pid, err := z.Load(zoneId)
	if err != nil {
		return nil, fmt.Errorf("zone %s not found: %w", zoneId, err)
	}

	future := z.system.Root.RequestFuture(pid, req, parseTimeout(timeout...))
	return future, nil
}

// Load 加载或创建 Zone Actor
// zoneId: 服务区 ID
// zoneType: 服务区类型（可选，如果为 nil 则使用默认类型 ZoneTypeWild）
// 返回: Zone Actor 的 PID
func (m *ZoneManager) Load(zoneId string, zoneType ...servicezone.ZoneType) (*actor.PID, error) {
	// 先从缓存中查找
	if v, exists := m.cache.Get(zoneId); exists {
		if pid, ok := v.(*actor.PID); ok {
			return pid, nil
		}
	}

	// 确定 zoneType，如果没有提供则使用默认值
	var zt servicezone.ZoneType = servicezone.ZoneTypeWild
	if len(zoneType) > 0 {
		zt = zoneType[0]
	}

	// 使用 ExecuteOnce 确保并发安全
	re := m.exec.ExecuteOnce(zoneId, func() any {
		// 创建 ServiceZone
		zone := servicezone.NewServiceZone(zoneId, zt)

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

func (m *ZoneManager) Stop() {
	for _, pid := range m.cache.Items() {
		if pid, ok := pid.(*actor.PID); ok {
			servicezone_behavior.Stop(m.system, pid)
		}
	}
	m.cache.Clear()
	m.system.Shutdown()
}

func parseTimeout(timeout ...time.Duration) time.Duration {
	if len(timeout) > 0 {
		return timeout[0]
	}
	return 5 * time.Second
}
