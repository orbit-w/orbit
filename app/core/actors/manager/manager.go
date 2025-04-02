package manager

import (
	"errors"
	"fmt"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/orbit-w/orbit/lib/logger"
	"go.uber.org/zap"
)

// ActorManager is responsible for managing actor lifecycle
type ActorManager struct {
	actorSystem *actor.ActorSystem
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

	case *StartActorMessage:
		m.handleStartActor(context, msg)

	case *StartActorWithFutureMessage:
		m.handleStartActorWithFuture(context, msg)

	case *StopActorMessage:
		m.handleStopActor(context, msg)

	case *actor.Terminated: //system message
		// msg.Who contains the PID of the terminated actor
		// Handle child termination here
		fmt.Printf("Child actor %s has terminated\n", msg.Who.Id)
		m.handleActorStopped(context, msg.Who.Id)

	case *ChildStartedNotification:
		if msg.Error != nil {
			actorsCache.Delete(m.actorSystem, msg.ActorName)
		}

	default:
		logger.GetLogger().Error("ActorManager received unknown message", zap.Any("message", msg))
	}
}

// handleStartActor handles starting a new actor
// 异步启动Actor
func (m *ActorManager) handleStartActor(context actor.Context, msg *StartActorMessage) {
	// Check if actor already exists
	if pid, exists := actorsCache.Get(msg.ActorName); exists {
		context.Respond(pid)
		return
	}

	// 创建Actor工厂函数
	actorFactory := func() actor.Actor {
		behaivor := CreateBehaivorWithID(msg.Pattern, msg.ActorName)

		// 设置初始化完成通知
		childActor := NewChildActor(behaivor, msg.ActorName, func(err error) error {
			context.Send(context.Self(), &ChildStartedNotification{ActorName: msg.ActorName})
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

	// Watch the child actor for termination
	context.Watch(pid)

	// Store actor PID and update activity time
	actorsCache.Set(msg.ActorName, pid)

	context.Respond(pid)
}

// 同步启动Actor
func (m *ActorManager) handleStartActorWithFuture(context actor.Context, msg *StartActorWithFutureMessage) {
	// Check if actor already exists
	if pid, exists := actorsCache.Get(msg.ActorName); exists {
		context.Respond(pid)
		return
	}

	// 创建Future用于等待Actor启动完成
	future := actor.NewFuture(context.ActorSystem(), msg.Timeout)

	// 创建Actor工厂函数
	actorFactory := func() actor.Actor {
		behaivor := CreateBehaivorWithID(msg.Pattern, msg.ActorName)

		// 设置初始化完成通知
		childActor := NewChildActor(behaivor, msg.ActorName, func(err error) error {
			fmt.Printf("Child actor %s started\n", msg.ActorName)
			context.Send(future.PID(), &ChildStartedNotification{ActorName: msg.ActorName})
			return nil
		})
		return childActor
	}

	// 创建Props对象
	props := actor.PropsFromProducer(actorFactory)

	// Create new actor
	pid, err := context.SpawnNamed(props, msg.ActorName)
	if err != nil {
		if err == actor.ErrNameExists {
			context.Respond(pid)
			return
		}
		context.Respond(err)
		return
	}

	// Watch the child actor for termination
	context.Watch(pid)

	err = m.waitChildStartedFuture(context, pid, msg.ActorName, future)
	if err != nil {
		context.Stop(pid)
		context.Respond(err)
		return
	}

	// Store actor PID and update activity time
	actorsCache.Set(msg.ActorName, pid)

	context.Respond(pid)
}

// handleStopActor handles stopping an actor
func (m *ActorManager) handleStopActor(context actor.Context, msg *StopActorMessage) {
	pid, exists := actorsCache.Get(msg.ActorID)
	if !exists {
		context.Respond(nil)
		return
	}

	// Stop the actor
	context.Stop(pid)

	// Remove actor from maps
	actorsCache.Delete(m.actorSystem, msg.ActorID)

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
