package actor

import (
	"gitee.com/orbit-w/meteor/bases/misc/utils"
	"gitee.com/orbit-w/orbit/lib/logger"
	"github.com/asynkron/protoactor-go/actor"
	"go.uber.org/zap"
)

type Behavior interface {
	HandleRequest(ctx IContext, msg any) (any, error)
	HandleSend(ctx IContext, msg any)
	HandleForward(ctx IContext, msg any)
	HandleInit(ctx IContext) error
	HandleStopping(ctx IContext) error
	HandleStopped(ctx IContext) error
}

// ChildActor 表示由SupervisorActor管理的子Actor
// 实现了InitNotifiable接口，允许在初始化完成后发送通知
type ChildActor struct {
	Behavior
	*TimerMgr
	metaData     *Meta
	context      actor.Context
	actorName    string
	pattern      string
	initialized  bool
	initCallback func(err error) error
}

// NewChildActor 创建一个新的子Actor
func NewChildActor(behavior Behavior, name, pattern string, meta *Meta, initCB func(err error) error) *ChildActor {
	return &ChildActor{
		metaData:     meta,
		actorName:    name,
		pattern:      pattern,
		Behavior:     behavior,
		initCallback: initCB,
	}
}

func (state *ChildActor) GetContext() IContext {
	return state
}

// Receive 处理接收到的消息
func (state *ChildActor) Receive(context actor.Context) {
	defer utils.RecoverPanic()
	switch msg := context.Message().(type) {
	case *actor.Started:
		state.SetActorContext(context)
		// 执行初始化逻辑
		state.HandleInit(context)

	case *actor.Stopping:
		_ = state.HandleStopping(context)

	case *actor.Stopped:
		state.HandleStopped(context)

	case *actor.Restarting:
		logger.GetLogger().Info("Child actor restarting", zap.String("ActorName", state.GetContext().GetActorName()))
		// 重启时重新执行初始化逻辑
		//state.HandleInit(context)

	case *RequestMessage:
		state.handleMessage(context, msg)

	case *TimerMessage:
		state.Process(func(msg any) {
			state.HandleSend(state, msg)
		})

	default:
		logger.GetLogger().Info("Child actor received invalid message", zap.String("ActorName", state.GetContext().GetActorName()), zap.Any("Message", msg))
	}
}

// handleMessage 处理常规消息
func (state *ChildActor) handleMessage(context actor.Context, msg *RequestMessage) {
	switch msg.MsgType {
	case MessageTypeRequest:
		result, err := state.HandleRequest(state, msg.Message)
		if err != nil {
			context.Respond(err)
		} else {
			context.Respond(result)
		}
	case MessageTypeSend:
		state.HandleSend(state, msg.Message)
	case MessageTypeForward:
		state.HandleForward(state, msg.Message)
	}
}

// HandleInit 在Actor启动时执行的初始化逻辑
// 返回nil表示成功，否则返回错误
func (state *ChildActor) HandleInit(context actor.Context) {
	if state.initialized {
		return
	}

	// 执行初始化逻辑
	err := state.Behavior.HandleInit(state)
	if err != nil {
		logger.GetLogger().Error("Child actor initialization failed", zap.String("ActorName", state.GetActorName()), zap.Error(err))
	}

	// 初始化完成后，通知父进程
	if state.initCallback != nil {
		state.initCallback(err)
	}

	state.initialized = true
	if err == nil {
		logger.GetLogger().Info("Child actor started", zap.String("ActorName", state.GetActorName()))
	}

	// 初始化定时器
	state.TimerMgr = NewTimerMgr(func() {
		context.Send(context.Self(), &TimerMessage{})
	})
}

func (state *ChildActor) HandleStopping(context actor.Context) error {
	err := state.Behavior.HandleStopping(state)
	if err != nil {
		logger.GetLogger().Error("Child actor stopping failed", zap.String("ActorName", state.GetActorName()), zap.Error(err))
	} else {
		logger.GetLogger().Info("Child actor stopping", zap.String("ActorName", state.GetActorName()))
	}
	return err
}

func (state *ChildActor) HandleStopped(context actor.Context) {
	// 执行初始化逻辑
	err := state.Behavior.HandleStopped(state)
	if err != nil {
		logger.GetLogger().Error("Child actor stopped failed", zap.String("ActorName", state.GetActorName()), zap.Error(err))
	} else {
		logger.GetLogger().Info("Child actor stopped", zap.String("ActorName", state.GetActorName()))
	}

	actorsCache.CompareAndSwap(state.GetActorName(), StateNone, StateStopped)
}

func (state *ChildActor) SetMetaData(meta *Meta) {
	state.metaData = meta
}

func (state *ChildActor) GetMetaData() *Meta {
	return state.metaData
}

func (state *ChildActor) GetActorName() string {
	return state.actorName
}

func (state *ChildActor) GetPattern() string {
	return state.pattern
}

func (state *ChildActor) GetActorContext() actor.Context {
	return state.context
}

func (state *ChildActor) SetActorContext(context actor.Context) {
	state.context = context
}

func (state *ChildActor) GetServerId() string {
	if state.metaData == nil {
		return ""
	}
	return state.metaData.ServerId
}
