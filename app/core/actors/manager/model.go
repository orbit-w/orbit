package manager

import (
	"time"

	"github.com/asynkron/protoactor-go/actor"
)

type CastMessage struct {
	ActorName string
	Message   any
}

type CallMessage struct {
	ActorName string
	Message   any
}

type CallMessageResponse struct {
	ActorName string
	Message   any
	Error     error
}

type ForwardMessage struct {
	ActorName string
	Message   any
}

type ForwardMessageResponse struct {
	ActorName string
	Message   any
	Error     error
}

// ChildStartedNotification 子Actor启动完成并执行Behavior HandleInit后发送的通知
type ChildStartedNotification struct {
	ActorName string
	Error     error
	PID       *actor.PID
}

// Message types for ActorManager
type StartActorMessage struct {
	Pattern   string
	ActorName string
}

type StartActorWithFutureMessage struct {
	Pattern   string
	ActorName string
	Timeout   time.Duration
}

type StopActorMessage struct {
	ActorID string
}

type ActorStoppedMessage struct {
	ActorID string
}
