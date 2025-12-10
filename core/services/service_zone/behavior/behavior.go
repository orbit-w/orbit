package servicezone_behavior

import (
	"gitee.com/orbit-w/meteor/bases/misc/utils"
	"gitee.com/orbit-w/meteor/modules/mlog"
	mmeobj "gitee.com/orbit-w/orbit/app/mme"
	"gitee.com/orbit-w/orbit/app/proto/pb"
	"gitee.com/orbit-w/orbit/lib/module/logger"
	"github.com/asynkron/protoactor-go/actor"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

var (
	globalRouter IRouter
)

func SetRouter(router IRouter) {
	globalRouter = router
}

// ZoneActorBehavior ServiceZone Actor 的行为实现
// 实现 actor.Behavior 接口
type ZoneActorBehavior struct {
	ctx    IContext
	logger *mlog.Logger
}

// NewZoneActorBehavior 创建新的 ZoneActorBehavior
// router: 路由分发器，由外部注入，用于解耦 service_zone 和 routers 之间的循环依赖
func NewZoneActorBehavior(ctx IContext) actor.Actor {
	return &ZoneActorBehavior{
		ctx:    ctx,
		logger: logger.GetLogger(),
	}
}

// Start 启动 ServiceZone 的 Actor
// 需要在 Actor 系统启动后调用
// router: 路由分发器，用于处理请求（可选，如果为 nil 则无法处理请求）
func Start(system *actor.ActorSystem, id string, ctx IContext) (*actor.PID, error) {
	// 直接创建持久化Actor，使用supervision策略
	// 这里不使用supervision系统，而是直接创建，但使用supervision策略来保证容错
	decider := func(reason any) actor.Directive {
		// 使用Resume策略，出错后继续运行
		return actor.ResumeDirective
	}
	supervisor := actor.NewOneForOneStrategy(10, 1000, decider)

	props := actor.PropsFromProducer(func() actor.Actor {
		return NewZoneActorBehavior(ctx)
	}, actor.WithSupervisor(supervisor))

	// 直接在Root下创建Actor
	actorCtx := system.Root
	pid, err := actorCtx.SpawnNamed(props, id)
	if err != nil {
		return nil, err
	}

	return pid, nil
}

func Stop(system *actor.ActorSystem, pid *actor.PID) error {
	actorCtx := system.Root
	future := actorCtx.PoisonFuture(pid)
	if err := future.Wait(); err != nil {
		return err
	}
	return nil
}

// HandleRequest 处理请求消息（需要响应）
func (b *ZoneActorBehavior) Receive(ctx actor.Context) {
	defer utils.RecoverPanic()

	switch msg := ctx.Message().(type) {
	case *actor.Started:
		b.logger.Info("ZoneActor started", zap.String("ActorID", ctx.Self().Id))
	case *actor.Stopping:
		b.logger.Info("ZoneActor stopping", zap.String("ActorID", ctx.Self().Id))
	case *actor.Stopped:
		b.logger.Info("ZoneActor stopped", zap.String("ActorID", ctx.Self().Id))
	case *ClientRequest:
		b.HandleRequest(ctx, msg)
	case *AddEntityRequest:
		b.HandleAddEntity(ctx, msg)
	default:
		b.logger.Error("ZoneActor received unknown message", zap.Any("Message", msg))
	}
}

// HandleSend 处理发送消息
func (ab *ZoneActorBehavior) HandleRequest(ctx actor.Context, req *ClientRequest) {
	pid := req.GetPid()
	handler := globalRouter.Dispatch(pid)
	if handler == nil {
		ab.logger.Error("ZoneActor received unknown message", zap.Uint32("Pid", pid), zap.Any("Message", req))
		return
	}

	qw := NewRequestWrapper(pid)
	if err := qw.Unmarshal(req.GetIn()); err != nil {
		ab.logger.Error("ZoneActor unmarshal error", zap.Error(err), zap.Uint32("Pid", pid))
		return
	}

	entities, err := ab.ctx.LoadRefs(qw.GetRefs())
	if err != nil {
		ab.logger.Error("ZoneActor load refs error", zap.Error(err), zap.Uint32("Pid", pid))
		return
	}

	result, respName, err := handler(ab.ctx, qw.GetReq(), entities...)
	if err != nil {
		ab.logger.Error("ZoneActor handler error", zap.Error(err), zap.Uint32("Pid", pid))
	}

	ab.Persist(entities)

	if result == nil {
		rpid, ok := pb.GetProtocolID(respName)
		if !ok {
			ab.logger.Error("ZoneActor received unknown message", zap.Uint32("Pid", pid), zap.Any("Message", req))
			return
		}
		respData, err := proto.Marshal(result)
		if err != nil {
			ab.logger.Error("ZoneActor marshal error", zap.Error(err), zap.Uint32("Pid", pid))
			return
		}

		req.Response(respData, rpid)
	}

	for _, entity := range entities {
		entity.ClearAllDirtyFlags()
	}
}

// TODO：当map Value类型是Wrapper时，当只要有一个字段有变更，就需要无视Wrapper其他字段DirtyFlag，直接持久化Wrapper的子对象
func (ab *ZoneActorBehavior) Persist(entities []mmeobj.IEntity) {
	ab.ctx.Persist(entities...)
}

// HandleInit 处理初始化
func (ab *ZoneActorBehavior) HandleInit(ctx actor.Context) error {
	ab.logger.Info("ServiceZoneActor initialized",
		zap.String("ActorName", ctx.Self().Id))
	return nil
}

// HandleStopping 处理停止中
func (ab *ZoneActorBehavior) HandleStopping(ctx actor.Context) error {
	ab.logger.Info("ServiceZoneActor stopping",
		zap.String("ActorName", ctx.Self().Id))
	return nil
}

// HandleStopped 处理已停止
func (ab *ZoneActorBehavior) HandleStopped(ctx actor.Context) error {
	ab.logger.Info("ServiceZoneActor stopped",
		zap.String("ActorName", ctx.Self().Id))
	return nil
}

// handleAddEntity 处理添加 Entity 请求
func (ab *ZoneActorBehavior) HandleAddEntity(ctx actor.Context, msg *AddEntityRequest) *AddEntityResponse {
	if msg.Entity == nil {
		return &AddEntityResponse{Success: false}
	}

	ab.ctx.AddEntity(msg.Entity)
	//ab.ctx.UpdateSubscriptions(msg.Entity)

	return &AddEntityResponse{Success: true}
}
