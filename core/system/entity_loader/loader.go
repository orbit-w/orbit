package entityloader

import (
	"time"

	entitymgr "gitee.com/orbit-w/orbit/core/system/entity_mgr"
	"gitee.com/orbit-w/orbit/pkg/proto/mme"
	"github.com/asynkron/protoactor-go/actor"
)

// DatabaseResolver 数据库解析器，根据 EntityType 返回目标数据库名称。
// 由上层（全局配置 / 分区映射）注入，EntityLoader 不关心具体映射逻辑。
type DatabaseResolver func(entityType mme.EntityType) string

// EntityType 类型别名，方便内部使用。
type EntityType = mme.EntityType

// EntityLoader 异步实体加载组件。
//
// 基于 Actor 模型，桥接 Worker 和 PersistenceActor：
//   - 接收 Worker 的 EntityLoadBatchRequest（通过 Actor 消息）
//   - 调用 persistence.AsyncLoad 发起异步 DB 加载（以自身 PID 为 ResponseReceiver）
//   - 收到 LoadResponse 后创建 Entity、调用 OnLoad、注册到 EntityManager
//   - 批次完成后发送 EntityLoadBatchComplete 通知 Worker
//
// EntityManager 保持纯数据结构层不变，加载逻辑由 EntityLoader 承担。
//
// ActorSystem 必须与调用方 Worker 共享，以保证本地消息传递走直接内存路径，
// 避免跨系统路由引入的序列化开销。
type EntityLoader struct {
	em          *entitymgr.EntityManager
	dbResolver  DatabaseResolver
	loadTimeout time.Duration
	pid         *actor.PID
	system      *actor.ActorSystem
}

// NewEntityLoader 创建 EntityLoader 实例。
//
// system: 与 Worker 共享的 ActorSystem（保证本地消息传递走直接内存路径）
// em: EntityManager 实例（由 EntityLoader 向其注册加载完成的 Entity）
// resolver: 数据库解析器（entityType → database name）
func NewEntityLoader(system *actor.ActorSystem, em *entitymgr.EntityManager, resolver DatabaseResolver) *EntityLoader {
	return &EntityLoader{
		em:          em,
		dbResolver:  resolver,
		loadTimeout: DefaultLoadTimeout,
		system:      system,
	}
}

// WithLoadTimeout 设置批次加载超时时间（默认 DefaultLoadTimeout）。
// 超时后 EntityLoader 主动通知 Worker Success=false，防止 StagingArea 永久阻塞。
func (l *EntityLoader) WithLoadTimeout(d time.Duration) *EntityLoader {
	if d > 0 {
		l.loadTimeout = d
	}
	return l
}

// Start 在共享 ActorSystem 中 Spawn EntityLoaderActor。
func (l *EntityLoader) Start() error {
	decider := func(reason any) actor.Directive {
		return actor.ResumeDirective
	}
	supervisor := actor.NewOneForOneStrategy(10, 1000, decider)

	em := l.em
	resolver := l.dbResolver
	loadTimeout := l.loadTimeout
	props := actor.PropsFromProducer(func() actor.Actor {
		return newEntityLoaderActor(em, resolver, loadTimeout)
	}, actor.WithSupervisor(supervisor))

	pid, err := l.system.Root.SpawnNamed(props, "entity-loader-actor")
	if err != nil {
		return err
	}

	l.pid = pid
	return nil
}

// Stop 优雅停止 EntityLoader，等待所有进行中的加载请求处理完毕。
func (l *EntityLoader) Stop() error {
	if l.pid != nil {
		future := l.system.Root.PoisonFuture(l.pid)
		l.pid = nil
		if err := future.Wait(); err != nil {
			return err
		}
	}
	return nil
}

// GetPID 返回 EntityLoaderActor 的 PID。
// Worker 通过此 PID 发送 EntityLoadBatchRequest。
func (l *EntityLoader) GetPID() *actor.PID {
	return l.pid
}
