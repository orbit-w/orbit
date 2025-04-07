package manager

import (
	"errors"
	"fmt"

	"gitee.com/orbit-w/orbit/lib/logger"
	"github.com/asynkron/protoactor-go/actor"
	"go.uber.org/zap"
)

// ActorManager is responsible for managing actor lifecycle
type ActorManager struct {
	actorSystem *actor.ActorSystem
	starting    *Startings
}

// NewActorManager creates a new instance of ActorManager
func NewActorManager(actorSystem *actor.ActorSystem) *ActorManager {
	return &ActorManager{
		actorSystem: actorSystem,
	}
}

// Receive handles messages sent to the ActorManager
func (m *ActorManager) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		// Start the garbage collection routine
		logger.GetLogger().Info("ActorManager started")

	case *StartActorRequest:
		m.handleStartActor(context, msg)

	case *StopActorMessage:
		m.handleStopActor(context, msg)

	case *actor.Terminated: //system message
		// msg.Who contains the PID of the terminated actor
		// Handle child termination here
		logger.GetLogger().Info("Child actor has terminated", zap.String("ActorName", msg.Who.Id), zap.String("Reason", msg.Why.String()))
		m.handleActorStopped(context, msg.Who.Id)

	case *ChildStartedNotification:
		m.handleNotifyChildStarted(context, msg)

	default:
		logger.GetLogger().Error("ActorManager received unknown message", zap.Any("Message", msg))
	}
}

// handleStartActor handles starting a new actor
// 异步启动Actor
func (m *ActorManager) handleStartActor(context actor.Context, msg *StartActorRequest) {
	// 如果Actor已经存在，则直接返回
	if pid, exists := actorsCache.Get(msg.ActorName); exists {
		context.Respond(pid)
		return
	}

	// 如果正在启动，则将Future添加到队列中
	if m.starting.Exists(msg.ActorName) {
		m.starting.Push(msg.ActorName, msg.Future)
		context.Respond(nil)
		return
	}

	// 创建Actor工厂函数
	actorFactory := func() actor.Actor {
		behaivor := CreateBehaivorWithID(msg.Pattern, msg.ActorName)

		// 设置初始化完成通知
		childActor := NewChildActor(behaivor, msg.ActorName, func(err error) error {
			context.Send(context.Self(), &ChildStartedNotification{ActorName: msg.ActorName, Error: err})
			return nil
		})

		return childActor
	}

	// 创建Props对象
	props := actor.PropsFromProducer(actorFactory)

	// Create new actor
	pid, err := context.SpawnNamed(props, msg.ActorName)
	if err != nil {
		context.Respond(err)
		return
	}

	// 监听Actor的终止
	context.Watch(pid)

	// 将Actor添加到正在启动的列表中
	m.starting.Set(msg.ActorName, pid, msg.Future)

	context.Respond(nil)
}

// handleStopActor handles stopping an actor
func (m *ActorManager) handleStopActor(context actor.Context, msg *StopActorMessage) {
	pid, exists := actorsCache.Get(msg.ActorID)
	if !exists {
		context.Respond(nil)
		return
	}

	// Stop the actor
	context.Poison(pid)

	context.Respond(nil)
}

// handleActorStopped handles notification that an actor has stopped
func (m *ActorManager) handleActorStopped(_ actor.Context, actorName string) {
	// Remove actor from maps
	actorsCache.Delete(m.actorSystem, actorName)
}

func (m *ActorManager) waitChildStartedFuture(context actor.Context, childPid *actor.PID, name string, future *actor.Future) error {
	// 等待Actor启动完成
	result, err := future.Result()
	if err != nil {
		context.Stop(childPid)
		context.Respond(err)
		logger.GetLogger().Error("等待子Actor启动超时",
			zap.String("actorName", name),
			zap.Error(err))
		return err
	}

	// 8. 验证启动消息
	startedMsg, ok := result.(*ChildStartedNotification)
	if !ok || startedMsg.ActorName != name {
		context.Stop(childPid)
		context.Respond(errors.New("收到了错误的启动确认消息"))
		logger.GetLogger().Error("收到了错误的启动确认消息",
			zap.String("expected", name),
			zap.Any("received", result))
		return errors.New("收到了错误的启动确认消息")
	}

	if startedMsg.Error != nil {
		context.Stop(childPid)
		context.Respond(startedMsg.Error)
		logger.GetLogger().Error("子Actor启动失败",
			zap.String("actorName", name),
			zap.Error(startedMsg.Error))
		return fmt.Errorf("子Actor启动失败: %v", startedMsg.Error)
	}
	return nil
}

func (m *ActorManager) handleNotifyChildStarted(context actor.Context, msg *ChildStartedNotification) {
	child, ok := m.starting.PopAndRange(msg.ActorName, func(child *actor.PID, future *actor.Future) {
		if msg.Error != nil {
			context.Send(future.PID(), msg.Error)
		} else {
			context.Send(future.PID(), child)
		}
	})

	if !ok {
		logger.GetLogger().Error("ChildStartedNotification received for unknown actor", zap.String("ActorName", msg.ActorName))
		return
	}

	if msg.Error != nil {
		actorsCache.Set(msg.ActorName, child)
	}
}
