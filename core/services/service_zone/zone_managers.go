package servicezone

import (
	"fmt"
	"time"

	"gitee.com/orbit-w/orbit/lib/module/unipue_task_exec"
	"github.com/asynkron/protoactor-go/actor"
	cmap "github.com/orcaman/concurrent-map"
)

type ZoneManager struct {
	system *actor.ActorSystem
	cache  cmap.ConcurrentMap
	exec   *unipue_task_exec.UniqueTaskExecutor
	zones  []*ServiceZone
}

func NewZoneManager() *ZoneManager {
	return &ZoneManager{
		system: actor.NewActorSystem(),
		cache:  cmap.New(),
		exec:   unipue_task_exec.NewUniqueTaskExecutor(),
	}
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
func (m *ZoneManager) Load(zoneId string, zoneType ...ZoneType) (*actor.PID, error) {
	// 先从缓存中查找
	if v, exists := m.cache.Get(zoneId); exists {
		if pid, ok := v.(*actor.PID); ok {
			return pid, nil
		}
	}

	// 确定 zoneType，如果没有提供则使用默认值
	var zt ZoneType = ZoneTypeWild
	if len(zoneType) > 0 {
		zt = zoneType[0]
	}

	// 使用 ExecuteOnce 确保并发安全
	re := m.exec.ExecuteOnce(zoneId, func() any {
		// 创建 ServiceZone
		zone := NewServiceZone(zoneId, zt)

		// 启动 Zone Actor，传递 router
		if err := zone.Start(m.system); err != nil {
			return fmt.Errorf("failed to start zone actor for zoneId %s: %w", zoneId, err)
		}

		// 获取 Actor PID
		pid := zone.GetActorPID()
		if pid == nil {
			return fmt.Errorf("zone actor PID is nil for zoneId %s", zoneId)
		}

		// 缓存 PID
		m.cache.Set(zoneId, pid)
		m.zones = append(m.zones, zone)
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
	for _, zone := range m.zones {
		zone.Stop()
	}
	m.cache.Clear()
	m.zones = nil
	m.system.Shutdown()
}
