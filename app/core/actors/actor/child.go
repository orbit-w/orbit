package actor

import (
	"time"

	"gitee.com/orbit-w/meteor/bases/misc/utils"
	"gitee.com/orbit-w/orbit/lib/logger"
	"github.com/asynkron/protoactor-go/actor"
	"go.uber.org/zap"
)

type Behavior interface {
	HandleRequest(ctx IContext, msg any) (any, error)
	HandleSend(ctx IContext, msg any)
	HandleForward(ctx IContext, msg any)
	HandleInit(ctx IContext) error
	HandleStopping(ctx IContext) error
	HandleStopped(ctx IContext) error
}

// ChildActor 表示由SupervisorActor管理的子Actor
// 实现了InitNotifiable接口，允许在初始化完成后发送通知
type ChildActor struct {
	Behavior
	*TimerMgr
	metaData           *Meta
	context            actor.Context
	actorName          string
	pattern            string
	initCallback       func(err error) error
	aliveTimeout       time.Duration
	lastActivityTime   time.Time
	aliveCheckInterval time.Duration
}

// NewChildActor 创建一个新的子Actor
func NewChildActor(behavior Behavior, name, pattern string, meta *Meta, aliveTimeout time.Duration, initCB func(err error) error) *ChildActor {
	if aliveTimeout <= 0 {
		//默认保活时间是30Min，如果期间内没有消息
		aliveTimeout = DefaultAliveTimeout
	}
	return &ChildActor{
		metaData:           meta,
		actorName:          name,
		pattern:            pattern,
		Behavior:           behavior,
		initCallback:       initCB,
		aliveCheckInterval: AliveCheckInterval,
	}
}

func (state *ChildActor) GetContext() IContext {
	return state
}

// Receive 处理接收到的消息
func (state *ChildActor) Receive(context actor.Context) {
	defer utils.RecoverPanic()
	switch msg := context.Message().(type) {
	case *actor.Started:
		state.SetActorContext(context)
		// 执行初始化逻辑
		state.HandleInit(context)

	case *actor.Stopping:
		_ = state.HandleStopping(context)

	case *actor.Stopped:
		state.HandleStopped(context)

	case *actor.Restarting:
		logger.GetLogger().Info("Child actor restarting", zap.String("ActorName", state.GetContext().GetActorName()))
		// 重启时重新执行初始化逻辑
		//state.HandleInit(context)

	case *RequestMessage:
		state.handleMessage(context, msg)

	case *TimerMessage:
		state.Process(func(msg any) {
			switch msg.(type) {
			case *CheckAliveMessage:
				state.handleAliveCheck(context)
			default:
				state.HandleSend(state, msg)
			}
		})

	default:
		logger.GetLogger().Info("Child actor received invalid message", zap.String("ActorName", state.GetContext().GetActorName()), zap.Any("Message", msg))
	}
}

// handleMessage 处理常规消息
func (state *ChildActor) handleMessage(context actor.Context, msg *RequestMessage) {
	state.updateActivityTime()

	switch msg.MsgType {
	case MessageTypeRequest:
		result, err := state.HandleRequest(state, msg.Message)
		if err != nil {
			context.Respond(err)
		} else {
			context.Respond(result)
		}
	case MessageTypeSend:
		state.HandleSend(state, msg.Message)
	case MessageTypeForward:
		state.HandleForward(state, msg.Message)
	}
}

// HandleInit 在Actor启动时执行的初始化逻辑
// 返回nil表示成功，否则返回错误
func (state *ChildActor) HandleInit(context actor.Context) {
	// 执行初始化逻辑
	err := state.Behavior.HandleInit(state)
	if err != nil {
		logger.GetLogger().Error("Child actor initialization failed", zap.String("ActorName", state.GetActorName()), zap.Error(err))
	}

	// 初始化完成后，通知父进程
	if state.initCallback != nil {
		state.initCallback(err)
	}

	if err == nil {
		logger.GetLogger().Info("Child actor started", zap.String("ActorName", state.GetActorName()))
	}

	// 初始化定时器
	state.TimerMgr = NewTimerMgr(func() {
		context.Send(context.Self(), &TimerMessage{})
	})

	// 更新最后活动时间
	state.updateActivityTime()

	// 启动定时器
	state.schedule()

}

func (state *ChildActor) HandleStopping(context actor.Context) error {
	err := state.Behavior.HandleStopping(state)
	if err != nil {
		logger.GetLogger().Error("Child actor stopping failed", zap.String("ActorName", state.GetActorName()), zap.Error(err))
	} else {
		logger.GetLogger().Info("Child actor stopping", zap.String("ActorName", state.GetActorName()))
	}
	return err
}

func (state *ChildActor) HandleStopped(context actor.Context) {
	// 执行初始化逻辑
	err := state.Behavior.HandleStopped(state)
	if err != nil {
		logger.GetLogger().Error("Child actor stopped failed", zap.String("ActorName", state.GetActorName()), zap.Error(err))
	} else {
		logger.GetLogger().Info("Child actor stopped", zap.String("ActorName", state.GetActorName()))
	}
}

// 处理活跃检测
func (state *ChildActor) handleAliveCheck(context actor.Context) {
	if time.Since(state.lastActivityTime) > state.aliveTimeout {
		context.Send(context.Parent(), &PoisonActorMessage{
			ActorName: state.GetActorName(),
			Pattern:   state.GetPattern(),
		})
		logger.GetLogger().Info("Actor is not active, stopping",
			zap.String("ActorName", state.GetActorName()),
			zap.Duration("AliveTimeout", state.aliveTimeout),
			zap.Duration("LastActivityTime", time.Since(state.lastActivityTime)))
	} else {
		logger.GetLogger().Debug("Alive check",
			zap.String("ActorName", state.GetActorName()))
	}
}

// 启动活跃检测定时器
func (state *ChildActor) schedule() {
	state.TimerMgr.AddSystemTimer(aliveCheckTimerKey, state.aliveCheckInterval, checkAliveMessage)
	logger.GetLogger().Debug("Started alive check timer",
		zap.String("ActorName", state.GetActorName()),
		zap.Duration("Interval", state.aliveCheckInterval))
}

// IsActive 检查Actor是否活跃
func (state *ChildActor) IsActive() bool {
	return time.Since(state.lastActivityTime) < state.aliveTimeout
}

func (state *ChildActor) SetMetaData(meta *Meta) {
	state.metaData = meta
}

func (state *ChildActor) GetMetaData() *Meta {
	return state.metaData
}

func (state *ChildActor) GetActorName() string {
	return state.actorName
}

func (state *ChildActor) GetPattern() string {
	return state.pattern
}

func (state *ChildActor) GetActorContext() actor.Context {
	return state.context
}

func (state *ChildActor) SetActorContext(context actor.Context) {
	state.context = context
}

func (state *ChildActor) GetServerId() string {
	if state.metaData == nil {
		return ""
	}
	return state.metaData.ServerId
}

// 更新Actor最后活动时间
func (state *ChildActor) updateActivityTime() {
	state.lastActivityTime = time.Now()
}
