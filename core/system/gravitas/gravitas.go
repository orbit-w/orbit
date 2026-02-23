package gravitas

import (
	entityloader "gitee.com/orbit-w/orbit/core/system/gravitas/entity_loader"
	entitymgr "gitee.com/orbit-w/orbit/core/system/gravitas/entity_mgr"
	"gitee.com/orbit-w/orbit/core/system/gravitas/worker"
	"gitee.com/orbit-w/orbit/pkg/proto/mme"
	"github.com/asynkron/protoactor-go/actor"
)

// Config Gravitas 整体配置。
type Config struct {
	// DBResolver 根据 EntityType 返回目标数据库名称，由上层注入。
	DBResolver entityloader.DatabaseResolver
	// Pool WorkerPool 配置，零值时各字段使用内部默认值。
	Pool worker.PoolConfig
}

// Gravitas 消息调度系统的统一入口，整合 EntityManager / EntityLoader / WorkerPool。
//
// 启动顺序：EntityManager → EntityLoader（Spawn Actor）→ WorkerPool（Spawn Workers）
// 停止顺序：WorkerPool → EntityLoader（保证 Worker 不再产生新加载请求后再关闭 Loader）
type GravitasImpl struct {
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
func (g *GravitasImpl) init() error {
	g.system = actor.NewActorSystem()
	g.entityMgr = entitymgr.NewEntityManager()

	g.entityLoader = entityloader.NewEntityLoader(g.system, g.entityMgr, g.cfg.DBResolver)
	if err := g.entityLoader.Start(); err != nil {
		return err
	}

	pool, err := worker.NewWorkerPool(g.system, g.entityMgr, g.entityLoader.GetPID(), g.cfg.Pool)
	if err != nil {
		_ = g.entityLoader.Stop()
		return err
	}
	g.workerPool = pool
	return nil
}

// Stop 按逆序优雅停止所有子组件。
func (g *GravitasImpl) Stop() error {
	if g.workerPool != nil {
		g.workerPool.Stop()
		g.workerPool = nil
	}
	if g.entityLoader != nil {
		if err := g.entityLoader.Stop(); err != nil {
			return err
		}
		g.entityLoader = nil
	}
	return nil
}

// Dispatch 将消息分发到对应的 Worker。
//
// 内部按 EntityID 升序排序 refs，取最小 EntityID 作为 AnchorID 路由到固定 Worker。
// 同一组 Entity 的消息总是路由到同一个 Worker，保证 FIFO 顺序。
func (g *GravitasImpl) Dispatch(refs []entityloader.EntityRef) {
	g.workerPool.Dispatch(refs, g.handler)
}

// DispatchFromProto 从 proto EntityRef 列表分发消息（便捷方法）。
func (g *GravitasImpl) DispatchFromProto(refs []*mme.EntityRef) {
	g.workerPool.DispatchFromProto(refs, g.handler)
}

// EntityManager 返回底层 EntityManager，供外部直接查询已加载的 Entity。
func (g *GravitasImpl) EntityManager() *entitymgr.EntityManager {
	return g.entityMgr
}
