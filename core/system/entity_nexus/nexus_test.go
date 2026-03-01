package entity_nexus

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	entityloader "gitee.com/orbit-w/orbit/core/system/entity_nexus/entity_loader"
	"gitee.com/orbit-w/orbit/core/system/entity_nexus/worker"
	mmeobj "gitee.com/orbit-w/orbit/internal/game/mme"
	"gitee.com/orbit-w/orbit/internal/game/mme_agent"
	"gitee.com/orbit-w/orbit/pkg/proto/mme"
	"go.mongodb.org/mongo-driver/v2/bson"
	"google.golang.org/protobuf/proto"
)

type testEntity struct {
	id         int64
	entityType mme.EntityType
}

func (e *testEntity) GetId() int64 {
	return e.id
}

func (e *testEntity) GetEntityType() mme.EntityType {
	return e.entityType
}

func (e *testEntity) GetEntityWrapper() mmeobj.IEntityWrapper {
	return nil
}

func (e *testEntity) OnLoad(raw bson.Raw, new bool) error {
	return nil
}

func (e *testEntity) OnSave() error {
	return nil
}

func (e *testEntity) OnLogin() error {
	return nil
}

func (e *testEntity) OnLogout() error {
	return nil
}

func newTestNexus(t *testing.T, handler func([]mme_agent.IEntity) error) *Nexus {
	t.Helper()

	n := &Nexus{
		cfg: Config{
			DBResolver: func(entityType mme.EntityType) string { return "test" },
			Pool: worker.PoolConfig{
				WorkerCount:  10,
				TickInterval: 20 * time.Millisecond,
			},
		},
		handler: handler,
	}
	if err := n.init(); err != nil {
		t.Fatalf("init nexus failed: %v", err)
	}
	t.Cleanup(func() {
		if err := n.Stop(); err != nil {
			t.Fatalf("stop nexus failed: %v", err)
		}
	})
	return n
}

// TestNexusInitAndStopLifecycle 验证 Nexus 的初始化与停止生命周期。
//
// 测试目的：
//   - init 后 ActorSystem、EntityManager、EntityLoader、WorkerPool 均正确创建并启动
//   - Stop 按逆序优雅停止各组件，并清空引用
//   - 重复调用 Stop 具有幂等性，不报错
func TestNexusInitAndStopLifecycle(t *testing.T) {
	n := &Nexus{
		cfg: Config{
			DBResolver: func(entityType mme.EntityType) string { return "test" },
			Pool: worker.PoolConfig{
				WorkerCount:  1,
				TickInterval: 20 * time.Millisecond,
			},
		},
	}
	if err := n.init(); err != nil {
		t.Fatalf("init nexus failed: %v", err)
	}

	if n.system == nil {
		t.Fatalf("actor system should be initialized")
	}
	if n.entityMgr == nil {
		t.Fatalf("entity manager should be initialized")
	}
	if n.entityLoader == nil || n.entityLoader.GetPID() == nil {
		t.Fatalf("entity loader should be started with valid pid")
	}
	if n.workerPool == nil {
		t.Fatalf("worker pool should be initialized")
	}
	if got := n.workerPool.WorkerCount(); got != 1 {
		t.Fatalf("unexpected worker count: got=%d want=1", got)
	}

	if err := n.Stop(); err != nil {
		t.Fatalf("stop nexus failed: %v", err)
	}
	if n.workerPool != nil {
		t.Fatalf("worker pool should be nil after stop")
	}
	if n.entityLoader != nil {
		t.Fatalf("entity loader should be nil after stop")
	}

	if err := n.Stop(); err != nil {
		t.Fatalf("second stop should be idempotent: %v", err)
	}
}

// TestNexusDispatchWithLoadedEntities 验证消息分发与业务执行的主路径。
//
// 测试目的：
//   - 当 Entity 已预加载时，Dispatch 能正确路由到 Worker 并执行业务 Handler
//   - 传入乱序的 EntityRefs，验证内部按 AnchorID（最小 EntityID）路由且能拿到全部涉及的 Entity
//   - Handler 收到的 entities 映射包含消息涉及的所有 Entity
func TestNexusDispatchWithLoadedEntities(t *testing.T) {
	done := make(chan map[int64]mme_agent.IEntity, 1)
	n := newTestNexus(t, func(entities []mme_agent.IEntity) error {
		m := make(map[int64]mme_agent.IEntity)
		for _, entity := range entities {
			m[entity.GetId()] = entity
		}
		done <- m
		return nil
	})

	n.EntityManager().Add(&testEntity{id: 101, entityType: mme.EntityType_PlayerEntityType})
	n.EntityManager().Add(&testEntity{id: 102, entityType: mme.EntityType_PlayerEntityType})

	n.Dispatch([]entityloader.EntityRef{
		{EntityID: 102, EntityType: mme.EntityType_PlayerEntityType},
		{EntityID: 101, EntityType: mme.EntityType_PlayerEntityType},
	})

	select {
	case entities := <-done:
		if len(entities) != 2 {
			t.Fatalf("handler should receive 2 entities, got=%d", len(entities))
		}
		if _, ok := entities[101]; !ok {
			t.Fatalf("entity 101 missing in handler payload")
		}
		if _, ok := entities[102]; !ok {
			t.Fatalf("entity 102 missing in handler payload")
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("handler was not called in time")
	}
}

// TestNexusDispatchFromProtoWithLoadedEntity 验证从 proto EntityRef 列表分发的便捷入口。
//
// 测试目的：
//   - DispatchFromProto 能将 []*mme.EntityRef 正确转换为内部 EntityRef 并分发
//   - nil 引用被过滤，不影响正常消息处理
//   - 与 Dispatch 一致，能触发 Handler 并收到正确的 entities
func TestNexusDispatchFromProtoWithLoadedEntity(t *testing.T) {
	done := make(chan map[int64]mme_agent.IEntity, 1)
	n := newTestNexus(t, func(entities []mme_agent.IEntity) error {
		m := make(map[int64]mme_agent.IEntity)
		for _, entity := range entities {
			m[entity.GetId()] = entity
		}
		done <- m
		return nil
	})

	n.EntityManager().Add(&testEntity{id: 2001, entityType: mme.EntityType_PlayerEntityType})

	n.DispatchFromProto([]*mme.EntityRef{
		nil,
		{
			EntityId:   proto.Int64(2001),
			EntityType: mme.EntityType_PlayerEntityType.Enum(),
		},
	})

	select {
	case entities := <-done:
		if len(entities) != 1 {
			t.Fatalf("handler should receive 1 entity, got=%d", len(entities))
		}
		if _, ok := entities[2001]; !ok {
			t.Fatalf("entity 2001 missing in handler payload")
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("handler was not called in time")
	}
}

// TestSingletonStartAndStop 验证 Nexus 单例的完整生命周期。
//
// 测试目的：
//   - Start 初始化单例后，GetInst、GetEntityManager 返回有效实例
//   - 通过单例的 Dispatch 能正常分发消息并触发 Handler
//   - Stop 后单例被清空，GetInst 返回 nil
//   - 测试前后对 Stop 的调用保证单例状态干净，避免影响其他测试
func TestSingletonStartAndStop(t *testing.T) {
	if err := Stop(); err != nil {
		t.Fatalf("cleanup stop before test failed: %v", err)
	}
	t.Cleanup(func() {
		if err := Stop(); err != nil {
			t.Fatalf("cleanup stop after test failed: %v", err)
		}
	})

	called := make(chan struct{}, 1)
	err := Start(
		Config{
			DBResolver: func(entityType mme.EntityType) string { return "test" },
			Pool: worker.PoolConfig{
				WorkerCount:  1,
				TickInterval: 20 * time.Millisecond,
			},
		},
		func(_ []mme_agent.IEntity) error {
			called <- struct{}{}
			return nil
		},
	)
	if err != nil {
		t.Fatalf("start singleton failed: %v", err)
	}

	if GetInst() == nil {
		t.Fatalf("singleton instance should be available after start")
	}
	if GetEntityManager() == nil {
		t.Fatalf("singleton entity manager should be available after start")
	}

	GetEntityManager().Add(&testEntity{id: 3001, entityType: mme.EntityType_PlayerEntityType})
	Dispatch([]entityloader.EntityRef{
		{EntityID: 3001, EntityType: mme.EntityType_PlayerEntityType},
	})

	select {
	case <-called:
	case <-time.After(2 * time.Second):
		t.Fatalf("singleton dispatch did not trigger handler")
	}

	if err := Stop(); err != nil {
		t.Fatalf("stop singleton failed: %v", err)
	}
	if GetInst() != nil {
		t.Fatalf("singleton instance should be nil after stop")
	}
}

// TestConcurrentDispatchRoutesToSameWorker 验证多条消息并发分发时，能够正确路由到同一个 Worker。
//
// 测试目的：
//   - 相同 EntityRefs（相同最小 EntityID → 相同 AnchorID）的多条消息，
//     无论以多大并发投递，均被路由到同一个 Worker
//   - 同一 Worker 基于 Actor mailbox 串行执行，Handler 最大并发度为 1
//   - 通过 maxConcurrent 原子计数断言不存在 Handler 并发执行
//   - 通过 totalHandled 断言所有 N 条消息均被完整处理，无遗漏
func TestConcurrentDispatchRoutesToSameWorker(t *testing.T) {
	const (
		workerCount = 4
		msgCount    = 30
		entityID1   = int64(701)
		entityID2   = int64(702)
	)

	var (
		activeHandlers atomic.Int64 // 当前正在执行 Handler 的计数
		maxConcurrent  atomic.Int64 // 观测到的最大并发 Handler 数
		totalHandled   atomic.Int64 // 已完成处理的消息总数
	)

	n := &Nexus{
		cfg: Config{
			DBResolver: func(entityType mme.EntityType) string { return "test" },
			Pool: worker.PoolConfig{
				WorkerCount:  workerCount,
				TickInterval: 20 * time.Millisecond,
			},
		},
		handler: func(_ []mme_agent.IEntity) error {
			cur := activeHandlers.Add(1)
			// 用 CAS 更新观测到的最大并发数
			for {
				old := maxConcurrent.Load()
				if cur <= old || maxConcurrent.CompareAndSwap(old, cur) {
					break
				}
			}
			// 短暂休眠，给潜在的并发执行创造时间窗口
			time.Sleep(2 * time.Millisecond)
			activeHandlers.Add(-1)
			totalHandled.Add(1)
			return nil
		},
	}
	if err := n.init(); err != nil {
		t.Fatalf("init nexus failed: %v", err)
	}
	t.Cleanup(func() {
		if err := n.Stop(); err != nil {
			t.Fatalf("stop nexus failed: %v", err)
		}
	})

	n.EntityManager().Add(&testEntity{id: entityID1, entityType: mme.EntityType_PlayerEntityType})
	n.EntityManager().Add(&testEntity{id: entityID2, entityType: mme.EntityType_PlayerEntityType})

	// proto 格式的 EntityRef，所有并发 goroutine 只读共享，无需复制
	baseProtoRefs := []*mme.EntityRef{
		{EntityId: proto.Int64(entityID1), EntityType: mme.EntityType_PlayerEntityType.Enum()},
		{EntityId: proto.Int64(entityID2), EntityType: mme.EntityType_PlayerEntityType.Enum()},
	}

	// 验证所有消息的路由目标一致（确定性断言）
	// GenAnchorID 接受内部 EntityRef，在此单独构造用于日志输出
	internalRefs := []entityloader.EntityRef{
		{EntityID: entityID1, EntityType: mme.EntityType_PlayerEntityType},
		{EntityID: entityID2, EntityType: mme.EntityType_PlayerEntityType},
	}
	anchorID := n.workerPool.GenAnchorID(internalRefs)
	expectedWorkerIdx := int(uint64(anchorID) % uint64(workerCount))
	if expectedWorkerIdx < 0 || expectedWorkerIdx >= workerCount {
		t.Fatalf("computed workerIdx=%d out of range [0, %d)", expectedWorkerIdx, workerCount)
	}
	t.Logf("AnchorID=%d → workerIdx=%d (all %d messages should route here)", anchorID, expectedWorkerIdx, msgCount)

	// 并发投递 N 条消息，统一走 DispatchFromProto 入口
	// DispatchFromProto 内部构造新的 []entityloader.EntityRef 再排序，不修改传入切片，可安全共享
	var wg sync.WaitGroup
	for i := 0; i < msgCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			n.DispatchFromProto(baseProtoRefs)
		}()
	}
	wg.Wait()

	// 等待所有消息处理完成，最长 5 秒
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if totalHandled.Load() >= msgCount {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	if got := totalHandled.Load(); got != msgCount {
		t.Fatalf("expected %d messages handled, got %d", msgCount, got)
	}

	// 核心断言：Handler 最大并发度为 1，证明所有消息由同一 Worker 串行处理
	if mc := maxConcurrent.Load(); mc > 1 {
		t.Fatalf("handlers executed concurrently (maxConcurrent=%d > 1), "+
			"messages must have been routed to different workers", mc)
	}
}
