package manager

import (
	"fmt"

	"gitee.com/orbit-w/orbit/lib/logger"
	"github.com/asynkron/protoactor-go/actor"
	"go.uber.org/zap"
)

type Behavior interface {
	HandleCall(context actor.Context, msg any) error
	HandleCast(context actor.Context, msg any) error
	HandleForward(context actor.Context, msg any) error
	HandleInit(context actor.Context) error
	HandleStop(context actor.Context) error
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
func NewChildActor(behavior Behavior, id string, cb func(err error) error) *ChildActor {
	return &ChildActor{
		ActorName:    id,
		Behavior:     behavior,
		initCallback: cb,
	}
}

// Receive 处理接收到的消息
func (state *ChildActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		// 执行初始化逻辑
		err := state.HandleInit(context)
		if err != nil {
			fmt.Printf("Child actor %s initialization failed: %v\n", state.ActorName, err)
			context.Stop(context.Self())
			return
		}

		logger.GetLogger().Info("Child actor started", zap.String("ActorName", state.ActorName))

	case *actor.Stopping:
		logger.GetLogger().Info("Child actor stopping", zap.String("ActorName", state.ActorName))

	case *actor.Stopped:
		state.HandleStop(context)
		logger.GetLogger().Info("Child actor stopped", zap.String("ActorName", state.ActorName))

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
func (state *ChildActor) HandleInit(context actor.Context) error {
	if state.initialized {
		return nil
	}

	// 执行初始化逻辑
	err := state.Behavior.HandleInit(context)

	// 初始化完成后，通知父进程
	if state.initCallback != nil {
		// 只有在初始化成功时才通知父进程
		if err := state.initCallback(err); err != nil {
			err = fmt.Errorf("child actor %s initialization failed: %v", state.ActorName, err)
			context.Stop(context.Self())
			return err
		}
	}
	return nil
}
