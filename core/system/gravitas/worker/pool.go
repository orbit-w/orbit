package worker

import (
	"fmt"
	"sort"
	"time"

	entityloader "gitee.com/orbit-w/orbit/core/system/gravitas/entity_loader"
	entitymgr "gitee.com/orbit-w/orbit/core/system/gravitas/entity_mgr"
	orbitutils "gitee.com/orbit-w/orbit/lib/utils"
	"gitee.com/orbit-w/orbit/pkg/proto/mme"
	"github.com/asynkron/protoactor-go/actor"
)

// PoolConfig WorkerPool 配置。
type PoolConfig struct {
	WorkerCount   int           // Worker 数量（默认 4）
	RetryQuota    int           // 每 Tick 重试配额（默认 10）
	TickInterval  time.Duration // Tick 间隔（默认 50ms）
	MaxRetryCount int           // IO 失败最大重试次数（默认 3）
}

// WorkerPool 管理多个 Worker Actor，提供基于 AnchorID 的消息路由。
//
// 路由策略：AnchorID = min(EntityRefs.EntityID)，
// 通过 hash(AnchorID) % WorkerCount 路由到固定 Worker。
// 同一组 Entity 的消息总是路由到同一个 Worker，保证 FIFO 顺序（INV-4）。
type WorkerPool struct {
	workers []*actor.PID
	count   int
	system  *actor.ActorSystem
	em      *entitymgr.EntityManager
}

// NewWorkerPool 创建并启动 WorkerPool。
//
// system:    与 Worker/EntityLoader 共享的 ActorSystem
// em:        EntityManager 实例
// loaderPID: EntityLoader Actor PID
// cfg:       配置参数
func NewWorkerPool(
	system *actor.ActorSystem,
	em *entitymgr.EntityManager,
	loaderPID *actor.PID,
	cfg PoolConfig,
) (*WorkerPool, error) {
	count := cfg.WorkerCount
	if count <= 0 {
		count = 4
	}

	pool := &WorkerPool{
		workers: make([]*actor.PID, count),
		count:   count,
		system:  system,
		em:      em,
	}

	decider := func(reason any) actor.Directive {
		return actor.ResumeDirective
	}
	supervisor := actor.NewOneForOneStrategy(10, 1000, decider)

	for i := 0; i < count; i++ {
		workerID := i
		props := actor.PropsFromProducer(func() actor.Actor {
			return NewWorker(WorkerConfig{
				ID:            workerID,
				RetryQuota:    cfg.RetryQuota,
				TickInterval:  cfg.TickInterval,
				MaxRetryCount: cfg.MaxRetryCount,
			}, em, loaderPID)
		}, actor.WithSupervisor(supervisor))

		pid, err := system.Root.SpawnNamed(props, fmt.Sprintf("worker-%d", i))
		if err != nil {
			pool.Stop()
			return nil, fmt.Errorf("failed to spawn worker %d: %w", i, err)
		}
		pool.workers[i] = pid
	}

	return pool, nil
}

// Dispatch 将消息发送到对应的 Worker。
//
// 内部按 EntityID 升序排序 refs（保证 INV-5：加锁顺序一致），
// 取 min(EntityID) 作为 AnchorID 路由到固定 Worker。
func (p *WorkerPool) Dispatch(refs []entityloader.EntityRef, handler MessageHandler) {
	if len(refs) == 0 {
		return
	}

	sort.Slice(refs, func(i, j int) bool {
		return refs[i].EntityID < refs[j].EntityID
	})

	anchorID := refs[0].EntityID
	msg := &WorkerMessage{
		EntityRefs: refs,
		AnchorID:   anchorID,
		Handler:    handler,
		CreateTime: time.Now(),
	}

	workerIdx := int(uint64(orbitutils.Fmix64(uint64(anchorID))) % uint64(p.count))
	p.system.Root.Send(p.workers[workerIdx], msg)
}

// DispatchFromProto 从 proto EntityRef 列表分发消息（便捷方法）。
//
// 将 []*mme.EntityRef 转换为 []entityloader.EntityRef 后调用 Dispatch。
func (p *WorkerPool) DispatchFromProto(refs []*mme.EntityRef, handler MessageHandler) {
	entityRefs := make([]entityloader.EntityRef, 0, len(refs))
	for _, ref := range refs {
		if ref == nil {
			continue
		}
		entityRefs = append(entityRefs, entityloader.EntityRef{
			EntityID:   ref.GetEntityId(),
			EntityType: ref.GetEntityType(),
		})
	}
	p.Dispatch(entityRefs, handler)
}

// Stop 优雅停止所有 Worker Actor。
func (p *WorkerPool) Stop() {
	for i, pid := range p.workers {
		if pid != nil {
			_ = p.system.Root.PoisonFuture(pid).Wait()
			p.workers[i] = nil
		}
	}
}

// WorkerCount 返回 Worker 数量。
func (p *WorkerPool) WorkerCount() int {
	return p.count
}

// GetWorkerPID 返回指定索引的 Worker PID（高级用法 / 测试用）。
func (p *WorkerPool) GetWorkerPID(index int) *actor.PID {
	if index < 0 || index >= p.count {
		return nil
	}
	return p.workers[index]
}
