package servicezone_behavior

import (
	"gitee.com/orbit-w/meteor/bases/misc/utils"
	"gitee.com/orbit-w/meteor/modules/mlog"
	"gitee.com/orbit-w/orbit/core/network"
	mmeobj "gitee.com/orbit-w/orbit/internal/game/mme"
	"gitee.com/orbit-w/orbit/internal/game/proto/core"
	"gitee.com/orbit-w/orbit/internal/game/proto/mme"
	"gitee.com/orbit-w/orbit/internal/game/proto/pb"
	"gitee.com/orbit-w/orbit/lib/module/logger"
	mmemodel "gitee.com/orbit-w/orbit/lib/module/mme_model"
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
		b.HandleClientRequest(ctx, msg)
	case *AddEntityRequest:
		b.HandleAddEntity(ctx, msg)
	default:
		b.logger.Error("ZoneActor received unknown message", zap.Any("Message", msg))
	}
}

// HandleClientRequest 处理客户端请求
func (ab *ZoneActorBehavior) HandleClientRequest(ctx actor.Context, req *ClientRequest) {
	pid := req.GetPid()
	switch pid {
	case pb.PID_Request_SetEntityRequest: // 设置实体请求
		ab.HandleSystemRequest_SetEntity(ctx, req)
	default:
		ab.HandleClientRequestByRouter(ctx, req) // 其他请求交给Handler处理
	}
}

// HandleSend 处理发送消息
func (ab *ZoneActorBehavior) HandleClientRequestByRouter(ctx actor.Context, req *ClientRequest) {
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

	resp, respName, err := handler(ab.ctx, qw.GetReq(), entities...)
	if err != nil {
		ab.logger.Error("ZoneActor handler error", zap.Error(err), zap.Uint32("Pid", pid))
	}

	ab.Persist(entities)

	ab.SendMessage(req.IClientRequest, entities, resp, respName)

	for _, entity := range entities {
		entity.ClearAllDirtyFlags()
	}
}

func (ab *ZoneActorBehavior) Persist(entities []mmeobj.IEntity) {
	ab.ctx.Persist(entities...)
}

func (ab *ZoneActorBehavior) PackEntityChangeNotify(entities []mmeobj.IEntity, messages []network.Message) []network.Message {
	var changes []*core.EntityChange
	// 发送实体变更消息
	for i := range entities {
		entity := entities[i]
		change, err := packEntityChange(entity)
		if err != nil {
			ab.logger.Error("ZoneActor pack entity change error", zap.Error(err))
			continue
		}
		if change == nil {
			continue
		}
		changes = append(changes, change)
	}

	if len(changes) > 0 {
		rawData, err := proto.Marshal(&core.Notify_EntityChangeNotify{
			EntityChanges: changes,
		})
		if err != nil {
			ab.logger.Error("ZoneActor marshal entity change notify error", zap.Error(err))
			return messages
		}
		messages = append(messages, network.Message{
			Pid:  pb.PID_Notify_EntityChangeNotify,
			Seq:  0,
			Data: rawData,
		})
	}
	return messages
}

func (ab *ZoneActorBehavior) PackResponse(seq uint32, resp proto.Message, respName string, messages []network.Message) []network.Message {
	if resp == nil {
		return messages
	}

	rpid, ok := pb.GetProtocolID(respName)
	if !ok {
		ab.logger.Error("ZoneActor received unknown message", zap.String("RespName", respName))
		return messages
	}
	respData, err := proto.Marshal(resp)
	if err != nil {
		ab.logger.Error("ZoneActor marshal error", zap.Error(err), zap.String("RespName", respName))
		return messages
	}
	messages = append(messages, network.Message{
		Pid:  rpid,
		Seq:  seq,
		Data: respData,
	})

	return messages
}

func (ab *ZoneActorBehavior) SendMessage(request network.IClientRequest, entities []mmeobj.IEntity, resp proto.Message, respName string) {
	messages := make([]network.Message, 0)
	messages = ab.PackEntityChangeNotify(entities, messages)
	messages = ab.PackResponse(request.GetSeq(), resp, respName, messages)

	if len(messages) > 0 {
		request.ResponseBatch(messages)
	}
}

func packEntityChange(entity mmeobj.IEntity) (*core.EntityChange, error) {
	change := entity.ToIncrementalProtoWithContext(mmemodel.SyncContextClient)
	if change == nil {
		return nil, nil
	}

	changeRaw, err := proto.Marshal(change)
	if err != nil {
		return nil, err
	}
	entityId := entity.GetXXXId()
	entityType := entity.GetEntityType()
	return &core.EntityChange{
		EntityRef: &mme.EntityRef{
			EntityId:   &entityId,
			EntityType: &entityType,
		},
		Data: changeRaw,
	}, nil
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
