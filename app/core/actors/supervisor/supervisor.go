package supervisor

import (
	"time"

	"errors"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/orbit-w/orbit/lib/logger"
	"go.uber.org/zap"
)

// SupervisorActor 是一个负责管理子Actor生命周期的管理者
// 它维护了一个从ActorId到PID的映射，用于跟踪所有子Actor
type SupervisorActor struct {
	actor.Actor
}

// NewSupervisorActor 创建一个新的SupervisorActor实例
// 初始化childActors映射表，用于跟踪子Actor
func NewSupervisorActor() *SupervisorActor {
	return &SupervisorActor{}
}

// Receive 处理SupervisorActor接收到的不同类型的消息
func (a *SupervisorActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		// Actor启动时初始化子Actor映射表

	case *SyncStartChildActor:
		// 委托给SyncStartChildActor方法处理
		pid, err := a.SyncStartChildActor(context, msg)
		// 如果启动成功，将PID返回给发送者
		context.Respond(&SyncStartChildActorResponse{
			ActorName: msg.ActorName,
			PID:       pid,
			Error:     err,
		})

	case *actor.Terminated:
		// 处理子Actor终止的消息
		// 当被监视的子Actor终止时，清理相关资源
		RangeActors(func(id string, pid *actor.PID) {
			if pid.String() == msg.Who.String() {
				// 从映射表中删除已终止的子Actor
				RemoveActor(id)
			}
		})
	}
}

// StartChildActor 创建并启动一个子Actor，同步等待其启动完成
// 返回子Actor的PID，或者在出错时返回nil
func (a *SupervisorActor) SyncStartChildActor(context actor.Context, msg *SyncStartChildActor) (*actor.PID, error) {
	// 1. 检查Actor是否已存在
	if pid := GetActor(msg.ActorName); pid != nil {
		return pid, nil
	}

	// 2. 创建Future用于等待Actor启动完成
	future := actor.NewFuture(context.ActorSystem(), 5*time.Second)

	// 3. 创建Actor工厂函数
	actorFactory := func() actor.Actor {
		behaivor := CreateBehaivorWithID(msg.Pattern, msg.ActorName)

		// 设置初始化完成通知
		childActor := NewChildActor(behaivor, msg.ActorName, func() error {
			context.Send(future.PID(), &ChildStartedNotification{ActorName: msg.ActorName})
			return nil
		})
		return childActor
	}

	// 4. 创建Props对象
	props := actor.PropsFromProducer(actorFactory)

	// 5. 创建子Actor
	childPID := context.Spawn(props)
	if childPID == nil {
		return nil, errors.New("failed to spawn actor")
	}

	// 7. 等待Actor启动完成
	result, err := future.Result()
	if err != nil {
		context.Stop(childPID)
		logger.GetLogger().Error("等待子Actor启动超时",
			zap.String("actorName", msg.ActorName),
			zap.Error(err))
		return nil, err
	}

	// 8. 验证启动消息
	startedMsg, ok := result.(*ChildStartedNotification)
	if !ok || startedMsg.ActorName != msg.ActorName {
		context.Stop(childPID)
		logger.GetLogger().Error("收到了错误的启动确认消息",
			zap.String("expected", msg.ActorName),
			zap.Any("received", result))
		return nil, errors.New("收到了错误的启动确认消息")
	}

	// 9. 注册Actor到映射表
	SetActor(msg.ActorName, childPID)

	// 10. 设置Actor终止监视
	context.Watch(childPID)

	// 11. 发送初始消息（如果有）
	if msg.Message != nil {
		context.Send(childPID, msg.Message)
	}

	return childPID, nil
}
