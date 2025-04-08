package manager

import (
	"fmt"
	"strings"
	"time"

	"gitee.com/orbit-w/orbit/lib/logger"
	"github.com/asynkron/protoactor-go/actor"
	"go.uber.org/zap"
)

// ActorManager is responsible for managing actor lifecycle
type ActorManager struct {
	level       Level
	actorSystem *actor.ActorSystem
	starting    *Queue
	stopping    *Queue
	watchActors map[string]bool
}

// NewActorManager creates a new instance of ActorManager
func NewActorManager(actorSystem *actor.ActorSystem, level Level) *ActorManager {
	return &ActorManager{
		level:       level,
		actorSystem: actorSystem,
		starting:    NewPriorityQueue(),
		stopping:    NewPriorityQueue(),
		watchActors: make(map[string]bool, 0),
	}
}

func GenManagerName(level Level) string {
	return strings.Join([]string{ManagerName, fmt.Sprintf("level-%d", level)}, "-")
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
		m.handleActorStopped(context, msg.Who.Id)
		logger.GetLogger().Info("Child actor has terminated", zap.String("ActorName", msg.Who.Id), zap.String("Reason", msg.Why.String()))

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

	// 如果正在停止，则将Future添加到队列中
	if m.stopping.Exists(msg.ActorName) {
		m.stopping.Push(msg.ActorName, msg.Future)
		context.Respond(nil)
		return
	}

	pid, err := m.startActor(context, msg.Pattern, msg.ActorName)
	if err != nil {
		context.Respond(err)
		return
	}

	// 将Actor添加到正在启动的列表中
	m.starting.Insert(msg.ActorName, msg.Pattern, pid, msg.Future, time.Now().UnixNano())

	context.Respond(nil)
}

func (m *ActorManager) startActor(context actor.Context, pattern, actorName string) (*actor.PID, error) {
	// 创建Actor工厂函数
	actorFactory := func() actor.Actor {
		behaivor := CreateBehaivorWithID(pattern, actorName)

		// 设置初始化完成通知
		childActor := NewChildActor(behaivor, actorName, func(err error) error {
			context.Send(context.Self(), &ChildStartedNotification{ActorName: actorName, Error: err})
			return nil
		})

		return childActor
	}

	// 创建Props对象
	props := actor.PropsFromProducer(actorFactory)

	// Create new actor
	pid, err := context.SpawnNamed(props, actorName)
	if err != nil {
		return nil, err
	}

	// 监听Actor的终止
	context.Watch(pid)
	m.watchActors[actorName] = true

	return pid, nil
}

// handleStopActor handles stopping an actor
func (m *ActorManager) handleStopActor(context actor.Context, msg *StopActorMessage) {
	pid, exists := actorsCache.Get(msg.ActorName)
	if !exists {
		context.Respond(nil)
		return
	}

	// Stop the actor
	context.Poison(pid)

	// 将Actor从正在停止的列表中删除
	m.stopping.Insert(msg.ActorName, "", nil, nil, time.Now().UnixNano())

	// 从缓存中删除Actor
	actorsCache.Delete(msg.ActorName)

	context.Respond(nil)
}

// handleActorStopped handles notification that an actor has stopped
func (m *ActorManager) handleActorStopped(context actor.Context, actorName string) {
	defer func() {
		actorsCache.Delete(actorName)
		delete(m.watchActors, actorName)
	}()

	item, ok := m.stopping.Pop(actorName)
	if !ok {
		logger.GetLogger().Error("Actor stopped notification received for unknown actor", zap.String("ActorName", actorName))
		return
	}

	if item != nil && len(item.Future) > 0 {
		pid, err := m.startActor(context, item.Pattern, actorName)
		if err != nil {
			logger.GetLogger().Error("Failed to start actor", zap.String("ActorName", actorName), zap.Error(err))
			for i := range item.Future {
				target := item.Future[i]
				context.Send(target, err)
			}
			return
		}
		item.Child = pid
		m.starting.BatchInsert(actorName, item.Pattern, pid, item.Future, time.Now().UnixNano())
	}
}

func (m *ActorManager) handleNotifyChildStarted(context actor.Context, msg *ChildStartedNotification) {
	target, ok := m.starting.PopAndRangeWithKey(msg.ActorName, func(name string, child, future *actor.PID) bool {
		if msg.Error != nil {
			context.Send(future, msg.Error)
		} else {
			context.Send(future, child)
		}
		return false
	})

	if !ok {
		logger.GetLogger().Error("ChildStartedNotification received for unknown actor", zap.String("ActorName", msg.ActorName))
		return
	}

	if msg.Error != nil {
		actorsCache.Set(msg.ActorName, target)
	}
}

// 新增方法：检查Actor是否处于活跃状态（非stopping）
func (m *ActorManager) isActorAlive(pid *actor.PID) bool {
	// 使用Touch消息测试Actor是否响应
	// Touch是一种特殊消息，如果Actor活跃会回复Touched消息
	future := m.actorSystem.Root.RequestFuture(pid, &actor.Touch{}, 100*time.Millisecond)

	// 如果能够获得响应，说明Actor是活跃的
	result, err := future.Result()
	if err != nil {
		// 超时或其他错误，说明Actor可能已停止或正在停止
		return false
	}

	// 检查返回值是否为Touched消息
	_, ok := result.(*actor.Touched)
	return ok
}
