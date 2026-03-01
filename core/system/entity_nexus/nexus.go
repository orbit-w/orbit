package entity_nexus

import (
	entityloader "gitee.com/orbit-w/orbit/core/system/entity_nexus/entity_loader"
	entitymgr "gitee.com/orbit-w/orbit/core/system/entity_nexus/entity_mgr"
	"gitee.com/orbit-w/orbit/core/system/entity_nexus/worker"
	"gitee.com/orbit-w/orbit/pkg/proto/mme"
	"github.com/asynkron/protoactor-go/actor"
)

// Config Nexus 整体配置。
type Config struct {
	// DBResolver 根据 EntityType 返回目标数据库名称，由上层注入。
	DBResolver entityloader.DatabaseResolver
	// Pool WorkerPool 配置，零值时各字段使用内部默认值。
	Pool worker.PoolConfig
}

// Nexus 消息调度系统的统一入口，整合 EntityManager / EntityLoader / WorkerPool。
//
// 启动顺序：EntityManager → EntityLoader（Spawn Actor）→ WorkerPool（Spawn Workers）
// 停止顺序：WorkerPool → EntityLoader（保证 Worker 不再产生新加载请求后再关闭 Loader）
type Nexus struct {
	system       *actor.ActorSystem
	cfg          Config
	entityMgr    *entitymgr.EntityManager
	entityLoader *entityloader.EntityLoader
	workerPool   *worker.WorkerPool
	handler      worker.MessageHandler
}

// init 按顺序初始化并启动所有子组件。
//
// 由包级 Start 函数调用；若任一组件启动失败，已启动的组件会被回滚停止。
func (n *Nexus) init() error {
	n.system = actor.NewActorSystem()
	n.entityMgr = entitymgr.NewEntityManager()

	n.entityLoader = entityloader.NewEntityLoader(n.system, n.entityMgr, n.cfg.DBResolver)
	if err := n.entityLoader.Start(); err != nil {
		return err
	}

	pool, err := worker.NewWorkerPool(n.system, n.entityMgr, n.entityLoader.GetPID(), n.cfg.Pool)
	if err != nil {
		_ = n.entityLoader.Stop()
		return err
	}
	n.workerPool = pool
	return nil
}

// Stop 按逆序优雅停止所有子组件。
func (n *Nexus) Stop() error {
	if n.workerPool != nil {
		n.workerPool.Stop()
		n.workerPool = nil
	}
	if n.entityLoader != nil {
		if err := n.entityLoader.Stop(); err != nil {
			return err
		}
		n.entityLoader = nil
	}
	return nil
}

// Dispatch 将消息分发到对应的 Worker。
//
// 内部按 EntityID 升序排序 refs，取最小 EntityID 作为 AnchorID 路由到固定 Worker。
// 同一组 Entity 的消息总是路由到同一个 Worker，保证 FIFO 顺序。
func (n *Nexus) Dispatch(refs []entityloader.EntityRef) {
	n.workerPool.Dispatch(refs, n.handler)
}

// DispatchFromProto 从 proto EntityRef 列表分发消息（便捷方法）。
func (n *Nexus) DispatchFromProto(refs []*mme.EntityRef) {
	n.workerPool.DispatchFromProto(refs, n.handler)
}

// EntityManager 返回底层 EntityManager，供外部直接查询已加载的 Entity。
func (n *Nexus) EntityManager() *entitymgr.EntityManager {
	return n.entityMgr
}
