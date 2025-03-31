package supervisor

import (
	"fmt"

	"github.com/asynkron/protoactor-go/actor"
	cmap "github.com/orbit-w/meteor/bases/container/map/concurrent_map"
)

var (
	cache = cmap.ConcurrentMap[string, *actor.PID]{}
)

func GetActor(id string) *actor.PID {
	pid, ok := cache.Get(id)
	if !ok {
		return nil
	}
	return pid
}

func SetActor(id string, pid *actor.PID) {
	cache.Set(id, pid)
}

func RemoveActor(id string) {
	cache.Remove(id)
}

// ChildActor 表示由SupervisorActor管理的子Actor
// 实现了InitNotifiable接口，允许在初始化完成后发送通知
type ChildActor struct {
	ActorId      string
	initCallback func()        // 初始化完成后的回调函数
	initialized  bool          // 标记是否已初始化完成
	pendingMsgs  []interface{} // 存储初始化前收到的消息
}

// NewChildActor 创建一个新的子Actor
func NewChildActor(id string) *ChildActor {
	return &ChildActor{
		ActorId:     id,
		initialized: false,
		pendingMsgs: make([]interface{}, 0),
	}
}

// SetInitCallback 实现InitNotifiable接口
// 设置初始化完成后的回调函数
func (state *ChildActor) SetInitCallback(callback func()) {
	state.initCallback = callback
}

// Receive 处理接收到的消息
func (state *ChildActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		fmt.Printf("Child actor %s starting\n", state.ActorId)

		// 在Actor系统中注册该Actor
		SetActor(state.ActorId, context.Self())

		// 执行初始化逻辑
		state.initialize(context)

		// 标记初始化完成
		state.initialized = true

		// 如果设置了回调函数，则调用它通知初始化完成
		if state.initCallback != nil {
			state.initCallback()
		}

		// 处理之前存储的消息
		state.processPendingMessages(context)

	case *actor.Stopping:
		fmt.Printf("Child actor %s stopping\n", state.ActorId)
		// 清理资源
		RemoveActor(state.ActorId)

	case *actor.Stopped:
		fmt.Printf("Child actor %s stopped\n", state.ActorId)

	case *actor.Restarting:
		fmt.Printf("Child actor %s restarting\n", state.ActorId)
		// 重启时需要清理资源
		RemoveActor(state.ActorId)

	default:
		if !state.initialized {
			// 如果Actor尚未初始化完成，将消息存储起来，稍后处理
			state.pendingMsgs = append(state.pendingMsgs, msg)
		} else {
			// 处理消息
			state.handleMessage(context, msg)
		}
	}
}

// initialize 执行Actor的初始化逻辑
// 在实际应用中，这里可能包含数据库连接、状态恢复等耗时操作
func (state *ChildActor) initialize(context actor.Context) {
	// 模拟初始化过程，实际应用中可能包含更复杂的逻辑
	fmt.Printf("Initializing child actor %s...\n", state.ActorId)
}

// processPendingMessages 处理初始化前收到的消息
func (state *ChildActor) processPendingMessages(context actor.Context) {
	// 处理所有挂起的消息
	for _, msg := range state.pendingMsgs {
		state.handleMessage(context, msg)
	}
	// 清空挂起消息列表
	state.pendingMsgs = make([]interface{}, 0)
}

// handleMessage 处理常规消息
func (state *ChildActor) handleMessage(context actor.Context, msg interface{}) {
	fmt.Printf("Child actor %s received message: %v\n", state.ActorId, msg)
	// 根据消息类型执行相应操作
}
