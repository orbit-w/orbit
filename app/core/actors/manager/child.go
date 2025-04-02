package manager

import (
	"fmt"

	"github.com/asynkron/protoactor-go/actor"
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
		fmt.Printf("Child actor %s starting\n", state.ActorName)

		// 执行初始化逻辑
		err := state.HandleInit(context)
		if err != nil {
			fmt.Printf("Child actor %s initialization failed: %v\n", state.ActorName, err)
			context.Stop(context.Self())
			return
		}

	case *actor.Stopping:
		fmt.Printf("Child actor %s stopping\n", state.ActorName)

	case *actor.Stopped:
		fmt.Printf("Child actor %s stopped\n", state.ActorName)
		state.HandleStop(context)

	case *actor.Restarting:
		fmt.Printf("Child actor %s restarting\n", state.ActorName)
		// 重启时重新执行初始化逻辑
		//state.HandleInit(context)

	default:
		// 直接处理消息，不再需要消息缓存逻辑
		state.handleMessage(context, msg)
	}
}

// handleMessage 处理常规消息
func (state *ChildActor) handleMessage(context actor.Context, msg interface{}) {
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
