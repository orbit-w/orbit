package actor

import (
	"errors"
	"time"
)

// NewActorRef 创建一个新的ActorRef实例
// ActorRef为Actor的代理。它主要的作用是支持向它所代表的Actor发送消息，
// 从而实现Actor之间的通信。通过ActorRef，可以避免直接访问或操作Actor的内部信息和状态。
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

// Send 发送消息到Actor
// 如果Actor正在停止，则尝试排队订阅新Actor事件，
// 当新的Actor启动就绪后，重新发送消息
func (actorRef *ActorRef) Send(msg any) error {
	if err := actorRef.ref().Send(msg); err != nil {
		if errors.Is(err, ErrActorStopped) {
			return actorRef.ref().Send(msg)
		}
		return err
	}
	return nil
}

// RequestFuture 发送消息到Actor并等待消息回复
// 如果Actor正在停止，则尝试排队订阅新Actor事件，
// 当新的Actor启动就绪后，重新发送消息
func (actorRef *ActorRef) RequestFuture(msg any, timeout ...time.Duration) (any, error) {
	re, err := actorRef.ref().RequestFuture(msg, timeout...)
	if err == nil {
		return re, nil
	}

	if errors.Is(err, ErrActorStopped) {
		re, err = actorRef.ref().RequestFuture(msg, timeout...)
		if err != nil {
			return nil, err
		}
	}
	return re, err
}

// Stop 停止当前Actor
// 此方法向Actor系统发送停止信号，请求终止目标Actor的执行.
// 当有新消息发送到目标Actor，会将Actor重新激活。
// 调用此方法后，目标Actor将完成当前正在处理的消息，然后优雅地关闭
// 注意: 停止操作是异步的，方法调用后立即返回，不等待Actor实际停止
func (actorRef *ActorRef) Stop() {
	_ = StopActor(actorRef.ActorName, actorRef.Pattern)
}

// ref 获取Actor的 Process 引用
// 此方法返回当前ActorRef对应的Process
// 返回:
//   - *Process: 返回当前Actor 引用
func (actorRef *ActorRef) ref() *Process {
	return actorRef.Props.getOrCreateActorPID(actorRef.ActorName, actorRef.Pattern)
}
