package persistence

import (
	"context"
	"time"

	mongodbdriver "gitee.com/orbit-w/meteor/modules/database/no_sql/mongodb_driver"
	"github.com/asynkron/protoactor-go/actor"
)

// Persistence 持久化系统，提供消息驱动的持久化服务
type Persistence struct {
	db          *mongodbdriver.VirtualMongoClient
	actorPID    *actor.PID
	actorSystem *actor.ActorSystem
	pattern     string
}

// NewPersistence 创建新的持久化系统实例
func NewPersistence(db *mongodbdriver.VirtualMongoClient) *Persistence {
	if db == nil {
		panic(ErrDBNotInitialized)
	}
	return &Persistence{
		db:      db,
		pattern: PersistencePattern,
	}
}

// Start 启动持久化系统
// 需要在Actor系统启动后调用
func (p *Persistence) Start() error {
	p.actorSystem = actor.NewActorSystem()
	// 直接创建持久化Actor，使用supervision策略
	// 这里不使用supervision系统，而是直接创建，但使用supervision策略来保证容错
	decider := func(reason any) actor.Directive {
		// 使用Resume策略，出错后继续运行
		return actor.ResumeDirective
	}
	supervisor := actor.NewOneForOneStrategy(10, 1000, decider)

	props := actor.PropsFromProducer(func() actor.Actor {
		return NewPersistenceActor(p.db)
	}, actor.WithSupervisor(supervisor))

	// 直接在Root下创建Actor
	ctx := p.actorSystem.Root
	pid, err := ctx.SpawnNamed(props, "persistence-actor")
	if err != nil {
		return err
	}

	p.actorPID = pid
	return nil
}

// Poison 停止持久化系统, 同步等待
func (p *Persistence) Poison(ctx context.Context) error {
	if p.actorPID != nil {
		// 直接停止Actor
		ctx := p.actorSystem.Root
		future := ctx.PoisonFuture(p.actorPID)
		p.actorPID = nil
		if err := future.Wait(); err != nil {
			return err
		}
	}
	return nil
}

// Persist 异步持久化数据
// collection: MongoDB集合名称
// documentID: 文档ID
// wrapper: 实现了Wrapper接口的数据包装器
// 返回Future，可以通过Future.Result()获取结果
func (p *Persistence) Call(req any, timeout ...time.Duration) (*actor.Future, error) {
	if p.actorPID == nil {
		return nil, ErrDBNotInitialized
	}

	root := p.actorSystem.Root
	future := root.RequestFuture(p.actorPID, req, parseTimeout(timeout...))
	return future, nil
}

// GetActorPID 获取持久化Actor的PID
func (p *Persistence) GetActorPID() *actor.PID {
	return p.actorPID
}

func parseTimeout(timeout ...time.Duration) time.Duration {
	if len(timeout) > 0 {
		return timeout[0]
	}
	return 5 * time.Second
}
