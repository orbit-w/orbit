package actor

import (
	"gitee.com/orbit-w/orbit/lib/logger"
	"github.com/asynkron/protoactor-go/actor"
	"go.uber.org/zap"
)

type Behavior interface {
	HandleRequest(ctx IContext, msg any) (any, error)
	HandleSend(ctx IContext, msg any) error
	HandleForward(ctx IContext, msg any) error
	HandleInit(ctx IContext) error
	HandleStopping(ctx IContext) error
	HandleStopped(ctx IContext) error
}

// ChildActor 表示由SupervisorActor管理的子Actor
// 实现了InitNotifiable接口，允许在初始化完成后发送通知
type ChildActor struct {
	Behavior
	IContext
	initialized  bool
	initCallback func(err error) error
}

// NewChildActor 创建一个新的子Actor
func NewChildActor(behavior Behavior, name, pattern string, meta *Meta, initCB func(err error) error) *ChildActor {
	return &ChildActor{
		IContext:     NewContext(meta, name, pattern),
		Behavior:     behavior,
		initCallback: initCB,
	}
}

func (state *ChildActor) GetContext() IContext {
	return state.IContext
}

// Receive 处理接收到的消息
func (state *ChildActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		state.GetContext().SetActorContext(context)
		// 执行初始化逻辑
		state.HandleInit(state.GetContext(), context)

	case *actor.Stopping:
		_ = state.HandleStopping(state.GetContext())

	case *actor.Stopped:
		state.HandleStopped(state.GetContext())

	case *actor.Restarting:
		logger.GetLogger().Info("Child actor restarting", zap.String("ActorName", state.GetContext().GetActorName()))
		// 重启时重新执行初始化逻辑
		//state.HandleInit(context)

	case *RequestMessage:
		state.handleMessage(state.GetContext(), msg)

	default:
		logger.GetLogger().Info("Child actor received invalid message", zap.String("ActorName", state.GetContext().GetActorName()), zap.Any("Message", msg))
	}
}

// handleMessage 处理常规消息
func (state *ChildActor) handleMessage(ctx IContext, msg *RequestMessage) {
	switch msg.MsgType {
	case MessageTypeRequest:
		result, err := state.HandleRequest(ctx, msg.Message)
		if err != nil {
			ctx.GetActorContext().Respond(err)
		} else {
			ctx.GetActorContext().Respond(result)
		}
	case MessageTypeSend:
		state.HandleSend(ctx, msg.Message)
	case MessageTypeForward:
		state.HandleForward(ctx, msg.Message)
	}
}

// HandleInit 在Actor启动时执行的初始化逻辑
// 返回nil表示成功，否则返回错误
func (state *ChildActor) HandleInit(ctx IContext, actorContext actor.Context) {
	if state.initialized {
		return
	}

	// 执行初始化逻辑
	err := state.Behavior.HandleInit(ctx)
	if err != nil {
		logger.GetLogger().Error("Child actor initialization failed", zap.String("ActorName", ctx.GetActorName()), zap.Error(err))
	}

	// 初始化完成后，通知父进程
	if state.initCallback != nil {
		state.initCallback(err)
	}

	state.initialized = true
	if err == nil {
		logger.GetLogger().Info("Child actor started", zap.String("ActorName", ctx.GetActorName()))
	}
}

func (state *ChildActor) HandleStopping(ctx IContext) error {
	err := state.Behavior.HandleStopping(ctx)
	if err != nil {
		logger.GetLogger().Error("Child actor stopping failed", zap.String("ActorName", ctx.GetActorName()), zap.Error(err))
	} else {
		logger.GetLogger().Info("Child actor stopping", zap.String("ActorName", ctx.GetActorName()))
	}
	return err
}

func (state *ChildActor) HandleStopped(ctx IContext) {
	// 执行初始化逻辑
	err := state.Behavior.HandleStopped(ctx)
	if err != nil {
		logger.GetLogger().Error("Child actor stopped failed", zap.String("ActorName", ctx.GetActorName()), zap.Error(err))
	} else {
		logger.GetLogger().Info("Child actor stopped", zap.String("ActorName", ctx.GetActorName()))
	}

	actorsCache.CompareAndSwap(ctx.GetActorName(), StateNone, StateStopped)
}
