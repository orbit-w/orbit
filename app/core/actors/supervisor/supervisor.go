package supervisor

import (
	"time"

	"github.com/asynkron/protoactor-go/actor"
)

// SupervisorActor 是一个负责管理子Actor生命周期的管理者
// 它维护了一个从ActorId到PID的映射，用于跟踪所有子Actor
type SupervisorActor struct {
	actor.Actor
	childActors map[string]*actor.PID // 存储ActorId到PID的映射关系
}

// ChildStartedMessage 子Actor启动完成后发送的消息
type ChildStartedMessage struct {
	ActorId string
}

// NewSupervisorActor 创建一个新的SupervisorActor实例
// 初始化childActors映射表，用于跟踪子Actor
func NewSupervisorActor() *SupervisorActor {
	return &SupervisorActor{
		childActors: make(map[string]*actor.PID),
	}
}

// Receive 处理SupervisorActor接收到的不同类型的消息
func (a *SupervisorActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		// Actor启动时初始化子Actor映射表
		a.childActors = make(map[string]*actor.PID)

	case *SendMessageToChild:
		// 处理发送消息给特定ID子Actor的请求
		if childPID, exists := a.childActors[msg.ActorId]; exists {
			// 如果目标子Actor存在，则转发消息
			context.Send(childPID, msg.Message)
		} else {
			// 如果目标子Actor不存在，则通知发送者
			// 发送者可以随后发送StartChildActor消息来创建该Actor
			context.Send(context.Sender(), &ChildNotFound{ActorId: msg.ActorId})
		}

	case *StartChildActor:
		// 委托给StartChildActor方法处理
		pid := a.StartChildActor(context, msg)
		// 如果启动成功，将PID返回给发送者
		if pid != nil {
			context.Respond(pid)
		}

	case *actor.Terminated:
		// 处理子Actor终止的消息
		// 当被监视的子Actor终止时，清理相关资源
		for id, pid := range a.childActors {
			if pid.String() == msg.Who.String() {
				// 从映射表中删除已终止的子Actor
				delete(a.childActors, id)
				break
			}
		}
	}
}

// StartChildActor 创建并启动一个子Actor，同步等待其启动完成
// 返回子Actor的PID，或者在出错时返回nil
func (a *SupervisorActor) StartChildActor(context actor.Context, msg *StartChildActor) *actor.PID {
	// 检查Actor是否已存在
	if pid, exists := a.childActors[msg.ActorName]; exists {
		return pid
	}

	// 创建一个Future用于等待Actor启动完成
	future := actor.NewFuture(context.ActorSystem(), 5*time.Second)

	// 创建一个工厂函数，用于生成指定ID的Actor实例
	actorFactory := func() actor.Actor {
		// 创建Actor时传入future的PID，让子Actor在启动完成后通知
		childActor := CreateActorWithID(msg.ActorName, msg.ActorName)

		// 如果子Actor实现了InitNotifiable接口，设置通知机制
		if notifiable, ok := childActor.(InitNotifiable); ok {
			notifiable.SetInitCallback(func() {
				// 子Actor初始化完成后，通知future
				context.Send(future.PID(), &ChildStartedMessage{ActorId: msg.ActorName})
			})
		}

		return childActor
	}

	// 创建Props对象，用于配置Actor的行为和特性
	var props *actor.Props

	// 设置监督策略（如果指定）
	if msg.Pattern != "" {
		// 定义故障处理决策函数
		// 当子Actor发生故障时，该函数决定如何处理
		decider := func(reason interface{}) actor.Directive {
			// 默认采用重启策略恢复故障Actor
			return actor.RestartDirective
		}

		// 根据指定的模式选择不同的监督策略
		switch msg.Pattern {
		case "oneforone":
			// 一对一策略：当一个子Actor失败时，只重启该特定Actor
			// 参数含义：最大重启次数10次，时间窗口1000毫秒，使用上述决策函数
			strategy := actor.NewOneForOneStrategy(10, 1000, decider)
			props = actor.PropsFromProducer(actorFactory, actor.WithSupervisor(strategy))
		case "allforone":
			// 全体策略：当一个子Actor失败时，重启所有子Actor
			// 适用于子Actor之间存在强依赖关系的情况
			strategy := actor.NewAllForOneStrategy(10, 1000, decider)
			props = actor.PropsFromProducer(actorFactory, actor.WithSupervisor(strategy))
		default:
			// 无效模式，使用默认配置（无特定监督策略）
			props = actor.PropsFromProducer(actorFactory)
		}
	} else {
		// 未指定监督模式，使用默认配置
		props = actor.PropsFromProducer(actorFactory)
	}

	// 使用配置好的Props创建子Actor
	childPID := context.Spawn(props)
	// 将新创建的子Actor添加到映射表中
	a.childActors[msg.ActorName] = childPID

	// 监视子Actor的终止事件
	// 当子Actor终止时，SupervisorActor会收到Terminated消息
	context.Watch(childPID)

	// 发送初始消息（如果有）
	if msg.Message != nil {
		context.Send(childPID, msg.Message)
	}

	// 等待子Actor启动完成或超时
	result, err := future.Result()
	if err != nil {
		// 超时或发生错误
		context.Logger().Error("等待子Actor启动超时", map[string]interface{}{
			"actorName": msg.ActorName,
			"error":     err.Error(),
		})
		return nil
	}

	// 子Actor成功启动
	startedMsg, ok := result.(*ChildStartedMessage)
	if !ok || startedMsg.ActorId != msg.ActorName {
		context.Logger().Error("收到了错误的启动确认消息", map[string]interface{}{
			"expected": msg.ActorName,
			"received": result,
		})
		return nil
	}

	return childPID
}

// InitNotifiable 定义一个接口，允许Actor在初始化完成后通知
type InitNotifiable interface {
	// SetInitCallback 设置初始化完成后的回调函数
	SetInitCallback(callback func())
}
