package manager

import (
	"gitee.com/orbit-w/orbit/lib/logger"
	"github.com/asynkron/protoactor-go/actor"
	"go.uber.org/zap"
)

type Behavior interface {
	HandleCall(context actor.Context, msg any) error
	HandleCast(context actor.Context, msg any) error
	HandleForward(context actor.Context, msg any) error
	HandleInit(context actor.Context) error
	HandleStopping(context actor.Context) error
	HandleStopped(context actor.Context) error
}

// ChildActor 表示由SupervisorActor管理的子Actor
// 实现了InitNotifiable接口，允许在初始化完成后发送通知
type ChildActor struct {
	Behavior
	ActorName    string
	initialized  bool
	initCallback func(err error) error
}

// NewChildActor 创建一个新的子Actor
func NewChildActor(behavior Behavior, id string, initCB func(err error) error) *ChildActor {
	return &ChildActor{
		ActorName:    id,
		Behavior:     behavior,
		initCallback: initCB,
	}
}

// Receive 处理接收到的消息
func (state *ChildActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		// 执行初始化逻辑
		state.HandleInit(context)

	case *actor.Stopping:
		state.HandleStopping(context)

	case *actor.Stopped:
		state.HandleStopped(context)

	case *actor.Restarting:
		logger.GetLogger().Info("Child actor restarting", zap.String("ActorName", state.ActorName))
		// 重启时重新执行初始化逻辑
		//state.HandleInit(context)

	default:
		// 直接处理消息
		state.handleMessage(context, msg)
	}
}

// handleMessage 处理常规消息
func (state *ChildActor) handleMessage(context actor.Context, msg any) {
	switch msg := msg.(type) {
	case *CallMessage:
		state.HandleCall(context, msg.Message)
	case *CastMessage:
		state.HandleCast(context, msg.Message)
	case *ForwardMessage:
		state.HandleForward(context, msg.Message)
	}
}

// HandleInit 在Actor启动时执行的初始化逻辑
// 返回nil表示成功，否则返回错误
func (state *ChildActor) HandleInit(context actor.Context) {
	if state.initialized {
		return
	}

	// 执行初始化逻辑
	err := state.Behavior.HandleInit(context)
	if err != nil {
		logger.GetLogger().Error("Child actor initialization failed", zap.String("ActorName", state.ActorName), zap.Error(err))
		return
	}

	// 初始化完成后，通知父进程
	if state.initCallback != nil {
		state.initCallback(err)
	}

	state.initialized = true
	if err == nil {
		logger.GetLogger().Info("Child actor started", zap.String("ActorName", state.ActorName))
	}
}

func (state *ChildActor) HandleStopping(context actor.Context) error {
	return nil
}

func (state *ChildActor) HandleStopped(context actor.Context) {
	// 执行初始化逻辑
	err := state.Behavior.HandleStopped(context)
	if err != nil {
		logger.GetLogger().Error("Child actor stopped failed", zap.String("ActorName", state.ActorName), zap.Error(err))
	} else {
		logger.GetLogger().Info("Child actor stopped", zap.String("ActorName", state.ActorName))
	}
	actorsCache.CompareAndSwap(state.ActorName, StateNone, StateStopped)
}
