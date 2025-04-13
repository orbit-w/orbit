package actor

import (
	"time"

	actor "github.com/asynkron/protoactor-go/actor"
)

func NewActorRef(props *Props, actorName, pattern string, ops ...PropsOption) *ActorRef {
	if props == nil {
		props = NewProps()
	}

	for i := range ops {
		ops[i](props)
	}

	return &ActorRef{
		ActorName: actorName,
		Pattern:   pattern,
		Props:     props,
	}
}

func (actorRef *ActorRef) Send(msg any) {
	pid := actorRef.ref()
	System.ActorSystem().Root.Send(pid, &RequestMessage{
		MsgType: MessageTypeSend,
		Message: msg,
	})
}

func (actorRef *ActorRef) RequestFuture(msg any, timeout ...time.Duration) (any, error) {
	pid := actorRef.ref()
	future := System.ActorSystem().Root.RequestFuture(pid, &RequestMessage{
		MsgType: MessageTypeRequest,
		Message: msg,
	}, parseTimeout(timeout...))
	result, err := future.Result()
	if err != nil {
		return nil, err
	}

	switch v := result.(type) {
	case error:
		return nil, v
	default:
		return v, nil
	}
}

// Stop 停止当前Actor
// 此方法向Actor系统发送停止信号，请求终止目标Actor的执行
// 调用此方法后，目标Actor将完成当前正在处理的消息，然后优雅地关闭
// 注意: 停止操作是异步的，方法调用后立即返回，不等待Actor实际停止
func (actorRef *ActorRef) Stop() {
	StopActor(actorRef.ActorName, actorRef.Pattern)
}

// ref 获取Actor的PID引用
// 此方法返回当前ActorRef对应的Process ID (PID)
// PID是Actor系统中唯一标识一个Actor的标识符
// 返回:
//   - *actor.PID: 返回当前ActorRef的PID引用
//
// 注意: 此方法通常用于内部实现，获取底层Actor系统需要的PID引用
func (actorRef *ActorRef) ref() *actor.PID {
	return actorRef.Props.getOrCreateActorPID(actorRef.ActorName, actorRef.Pattern)
}
