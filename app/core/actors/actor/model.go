package actor

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
}

// Message types for ActorManager
type StartActorRequest struct {
	Pattern   string
	ActorName string
	Timeout   time.Duration
	Future    *actor.PID
	Props     *Props
}

type StartActorWait struct {
}

type StopActorMessage struct {
	ActorName string
	Pattern   string
}

type StopAllRequest struct {
}

type StopAllResponse struct {
	Complete bool
}

type TimerMessage struct{}

var (
	startActorWaitMessage = &StartActorWait{}
	checkAliveMessage     = &CheckAliveMessage{}
)

type ActorInfo struct {
	ActorName string
	Pattern   string
}

const (
	MessageTypeSend int8 = iota
	MessageTypeRequest
	MessageTypeForward
)

type RequestMessage struct {
	MsgType int8
	Message any
}

type CheckAliveMessage struct{}
