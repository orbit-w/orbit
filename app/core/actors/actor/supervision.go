package actor

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"gitee.com/orbit-w/meteor/bases/misc/utils"
	"gitee.com/orbit-w/orbit/lib/logger"
	"github.com/asynkron/protoactor-go/actor"
	"go.uber.org/zap"
)

const (
	StateActorSupervisionNormal int32 = iota
	StateActorSupervisionStopping
	StateActorSupervisionStopped
)

// ActorSupervision is responsible for managing actor lifecycle
type ActorSupervision struct {
	state       atomic.Int32
	level       Level
	actorSystem *actor.ActorSystem
	starting    *Queue
	stopping    *Queue
	restarting  *Queue
}

// NewActorSupervision creates a new instance of ActorManager
func NewActorSupervision(actorSystem *actor.ActorSystem, level Level) *ActorSupervision {
	return &ActorSupervision{
		level:       level,
		actorSystem: actorSystem,
		starting:    NewPriorityQueue(),
		stopping:    NewPriorityQueue(),
		restarting:  NewPriorityQueue(),
	}
}

func GenManagerName(level Level) string {
	return strings.Join([]string{ManagerName, fmt.Sprintf("level-%d", level)}, "-")
}

// Receive handles messages sent to the ActorManager
func (m *ActorSupervision) Receive(context actor.Context) {
	defer utils.RecoverPanic()

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
		m.handleActorStopped(context, ExtractActorName(msg.Who))
		logger.GetLogger().Info("Child actor has terminated", zap.String("ActorName", msg.Who.Id), zap.String("Reason", msg.Why.String()))

	case *ChildStartedNotification:
		m.handleNotifyChildStarted(context, msg)

	case *StopAllRequest:
		m.handleStoppingAll(context)

	case *actor.Stopped:
		m.handleStopped(context)

	default:
		logger.GetLogger().Error("ActorManager received unknown message", zap.Any("Message", msg))
	}
}

// handleStartActor handles starting a new actor
// 异步启动Actor
func (m *ActorSupervision) handleStartActor(context actor.Context, msg *StartActorRequest) {
	// 如果Actor已经存在，则直接返回
	if p, exists := actorsCache.Get(msg.ActorName); exists {
		context.Respond(p)
		return
	}

	// 如果正在启动，则将Future添加到队列中
	if m.starting.Exists(msg.ActorName) {
		m.starting.PushFuture(msg.ActorName, msg.Future)
		context.Respond(startActorWaitMessage)
		return
	}

	// 如果正在停止，则将Future添加到队列中
	if m.stopping.Exists(msg.ActorName) {
		if m.restarting.Exists(msg.ActorName) {
			m.restarting.PushFuture(msg.ActorName, msg.Future)
		} else {
			m.restarting.Insert(msg.ActorName, NewItem(msg.ActorName, msg.Pattern, nil, msg.Props, msg.Future), time.Now().UnixNano())
		}
		context.Respond(startActorWaitMessage)
		return
	}

	if m.isAllActorStopped() {
		context.Respond(ErrSupervisionStopped)
		return
	}

	pid, err := m.startActor(context, msg.Pattern, msg.ActorName, msg.Props)
	if err != nil {
		context.Respond(err)
		return
	}

	// 将Actor添加到正在启动的列表中
	item := NewItem(msg.ActorName, msg.Pattern, pid, msg.Props, msg.Future)
	err = m.starting.Insert(msg.ActorName, item, time.Now().UnixNano())
	if err != nil {
		logger.GetLogger().Error("[StartActor] Failed to insert starting queue", zap.String("ActorName", msg.ActorName), zap.Error(err))
		context.Respond(err)
		return
	}

	context.Respond(startActorWaitMessage)
}

func (m *ActorSupervision) startActor(context actor.Context, pattern, actorName string, props *Props) (*actor.PID, error) {
	// 创建Actor工厂函数
	actorFactory := func() actor.Actor {
		behavior := CreateBehaviorWithID(pattern, actorName)

		// 设置初始化完成通知
		childActor := NewChildActor(behavior, actorName, pattern, props.GetMeta(), props.GetAliveTimeout(), func(err error) error {
			context.Send(context.Self(), &ChildStartedNotification{ActorName: actorName, Error: err})
			return nil
		})

		return childActor
	}

	// Create new actor
	pid, err := context.SpawnNamed(actor.PropsFromProducer(actorFactory), actorName)
	if err != nil {
		return nil, err
	}

	return pid, nil
}

// handleStopActor handles stopping an actor
func (m *ActorSupervision) handleStopActor(context actor.Context, msg *StopActorMessage) {
	p, exists := actorsCache.Get(msg.ActorName)
	if !exists {
		context.Respond(nil)
		return
	}

	m.stopActor(context, msg.ActorName, p)
	context.Respond(nil)
}

// Pid 属于需要停止的目标Actor
// 异步停止，当发送Poison信号后立即返回调用者，
// 而不是等到Supervisor收到目标Actor Terminated的信号后才返回调用者
func (m *ActorSupervision) stopActor(context actor.Context, name string, p *Process) {
	// 立刻将Process从缓存中删除
	actorsCache.Delete(name)
	// 立刻将Process状态设置为停止状态
	p.Stop()
	// 异步停止Actor
	m.stopActorWithPID(context, name, p.Pattern, p.GetPID(), false)
}

func (m *ActorSupervision) stopActorWithPID(context actor.Context, actorName, pattern string, pid *actor.PID, clear bool) {
	// Stop the actor
	context.Poison(pid)

	// 将Actor从正在停止的列表中删除
	item := NewItem(actorName, pattern, nil, nil)
	m.stopping.Insert(actorName, item, time.Now().UnixNano())

	// 从缓存中删除Actor
	if clear {
		actorsCache.Delete(actorName)
	}
}

// handleActorStopped handles notification that an actor has stopped
func (m *ActorSupervision) handleActorStopped(context actor.Context, actorName string) {
	defer func() {
		actorsCache.Delete(actorName)
	}()

	_, ok := m.stopping.Pop(actorName)
	if !ok {
		if m.state.Load() < StateActorSupervisionStopping {
			logger.GetLogger().Error("Actor stopped notification received for unknown actor", zap.String("ActorName", actorName))
		}
		return
	}

	// 如果所有Actor都已经停止，则不启动新的Actor
	if m.isAllActorStopped() {
		return
	}

	watching, ok := m.restarting.Pop(actorName)
	if ok && watching.FuturesNum() > 0 {
		pid, err := m.startActor(context, watching.Pattern, actorName, watching.Props)
		if err != nil {
			logger.GetLogger().Error("Failed to start actor", zap.String("ActorName", actorName), zap.Error(err))
			for i := range watching.Future {
				target := watching.Future[i]
				context.Send(target, err)
			}
			return
		}

		newItem := NewItem(actorName, watching.Pattern, pid, watching.Props, watching.Future...)
		if err := m.starting.Insert(actorName, newItem, time.Now().UnixNano()); err != nil {
			logger.GetLogger().Error("[HandleActorRestart] Failed to insert starting queue", zap.String("ActorName", actorName), zap.Error(err))
		}
	}
}

func (m *ActorSupervision) handleNotifyChildStarted(context actor.Context, msg *ChildStartedNotification) {
	//TODO: 需要修复逻辑顺序问题
	watchers := make([]*actor.PID, 0)
	item, ok := m.starting.PopAndRangeWithKey(msg.ActorName, func(name, pattern string, child, future *actor.PID) bool {
		watchers = append(watchers, future)
		return true
	})

	if !ok {
		logger.GetLogger().Error("ChildStartedNotification received for unknown actor", zap.String("ActorName", msg.ActorName))
		return
	}

	if msg.Error == nil {
		logger.GetLogger().Info("Child actor started", zap.String("ActorName", msg.ActorName))
		p := NewActorProcess(msg.ActorName, item.Pattern, item.Child, item.Props)
		actorsCache.Set(msg.ActorName, p)
		for i := range watchers {
			w := watchers[i]
			context.Send(w, p)
		}
	} else {
		logger.GetLogger().Error("Child actor started with error", zap.String("ActorName", msg.ActorName), zap.Error(msg.Error))
		m.stopActorWithPID(context, msg.ActorName, item.Pattern, item.Child, false)
		for i := range watchers {
			w := watchers[i]
			context.Send(w, msg.Error)
		}
	}
}

func (m *ActorSupervision) handleStoppingAll(context actor.Context) {
	m.state.Store(StateActorSupervisionStopping)
	children := context.Children()
	for i := range children {
		child := children[i]
		name := ExtractActorName(child)
		if m.stopping.Exists(name) {
			continue
		}
		p, exists := actorsCache.Get(name)
		if exists {
			m.stopActor(context, name, p)
		} else {
			logger.GetLogger().Error("StoppingAll child not found", zap.String("ActorName", name))
		}
	}

	if !m.starting.Empty() || !m.stopping.Empty() || !m.restarting.Empty() {
		context.Respond(&StopAllResponse{Complete: false})
		return
	}

	complete := len(context.Children()) == 0
	if complete {
		m.state.Store(StateActorSupervisionStopped)
	}
	context.Respond(&StopAllResponse{Complete: complete})
}

func (m *ActorSupervision) handleStopped(_ actor.Context) {
	// 释放队列
	m.stopping.Free()
	m.restarting.Free()
	m.starting.Free()
}

func (m *ActorSupervision) isAllActorStopped() bool {
	return m.state.Load() == StateActorSupervisionStopped
}
