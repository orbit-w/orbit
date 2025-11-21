package servicezone

import (
	"gitee.com/orbit-w/meteor/bases/misc/utils"
	"gitee.com/orbit-w/meteor/modules/mlog"
	"gitee.com/orbit-w/orbit/lib/module/logger"
	"github.com/asynkron/protoactor-go/actor"
	"go.uber.org/zap"
)

// ZoneActorBehavior ServiceZone Actor 的行为实现
// 实现 actor.Behavior 接口
type ZoneActorBehavior struct {
	zone   *ServiceZone
	logger *mlog.Logger
}

// NewZoneActorBehavior 创建新的 ZoneActorBehavior
func NewZoneActorBehavior(zone *ServiceZone) actor.Actor {
	return &ZoneActorBehavior{
		zone:   zone,
		logger: logger.GetLogger(),
	}
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
	case *Request:
		b.HandleRequest(ctx, msg)
	default:
		b.logger.Error("ZoneActor received unknown message", zap.Any("Message", msg))
	}
}

// HandleSend 处理发送消息（不需要响应）
func (ab *ZoneActorBehavior) HandleRequest(ctx actor.Context, req *Request) {

}

// HandleForward 处理转发消息
func (ab *ZoneActorBehavior) HandleForward(ctx actor.Context, msg any) {

}

// HandleInit 处理初始化
func (ab *ZoneActorBehavior) HandleInit(ctx actor.Context) error {
	ab.logger.Info("ServiceZoneActor initialized",
		zap.String("ActorName", ctx.Self().Id),
		zap.String("ZoneID", ab.zone.ID))
	return nil
}

// HandleStopping 处理停止中
func (ab *ZoneActorBehavior) HandleStopping(ctx actor.Context) error {
	ab.logger.Info("ServiceZoneActor stopping",
		zap.String("ActorName", ctx.Self().Id),
		zap.String("ZoneID", ab.zone.ID))
	return nil
}

// HandleStopped 处理已停止
func (ab *ZoneActorBehavior) HandleStopped(ctx actor.Context) error {
	ab.logger.Info("ServiceZoneActor stopped",
		zap.String("ActorName", ctx.Self().Id),
		zap.String("ZoneID", ab.zone.ID))
	return nil
}

// handleAddEntity 处理添加 Entity 请求
func (ab *ZoneActorBehavior) handleAddEntity(msg *AddEntityRequest) *AddEntityResponse {
	if msg.Entity == nil {
		return &AddEntityResponse{Success: false}
	}

	ab.zone.addEntity(msg.Entity)
	ab.zone.UpdateSubscriptions(msg.Entity)

	return &AddEntityResponse{Success: true}
}

// handleRemoveEntity 处理移除 Entity 请求
func (ab *ZoneActorBehavior) handleRemoveEntity(msg *RemoveEntityRequest) *RemoveEntityResponse {
	entity := ab.zone.getEntityById(msg.EntityID)
	if entity == nil {
		return &RemoveEntityResponse{Success: false}
	}

	ab.zone.removeEntity(msg.EntityID)
	ab.zone.RemoveFromSubscriptions(entity)

	return &RemoveEntityResponse{Success: true}
}

// handleSubscribe 处理订阅请求
func (ab *ZoneActorBehavior) handleSubscribe(_ *SubscribeRequest) *SubscribeResponse {
	return &SubscribeResponse{Entities: nil}
}

// handleSubscribeByIds 处理按 ID 列表订阅请求
func (ab *ZoneActorBehavior) handleSubscribeByIds(_ *SubscribeByIdsRequest) *SubscribeResponse {
	return &SubscribeResponse{Entities: nil}
}

// handleSubscribeByType 处理按类型订阅请求
func (ab *ZoneActorBehavior) handleSubscribeByType(_ *SubscribeByTypeRequest) *SubscribeResponse {
	return &SubscribeResponse{Entities: nil}
}

// handleSubscribeAll 处理订阅所有请求
func (ab *ZoneActorBehavior) handleSubscribeAll(_ *SubscribeAllRequest) *SubscribeResponse {
	return &SubscribeResponse{Entities: nil}
}

// handleUnsubscribe 处理取消订阅请求
func (ab *ZoneActorBehavior) handleUnsubscribe(_ *UnsubscribeRequest) *UnsubscribeResponse {
	return &UnsubscribeResponse{Success: true}
}

// handleGetSubscribedEntities 处理获取已订阅 Entities 请求
func (ab *ZoneActorBehavior) handleGetSubscribedEntities(_ *GetSubscribedEntitiesRequest) *GetSubscribedEntitiesResponse {
	return &GetSubscribedEntitiesResponse{Entities: nil}
}
