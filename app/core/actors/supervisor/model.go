package supervisor

import "github.com/asynkron/protoactor-go/actor"

type SyncStartChildActor struct {
	ActorName string
	Pattern   string
	Message   any
}

type SyncStartChildActorResponse struct {
	ActorName string
	PID       *actor.PID
	Error     error
}

// Message types for actor communication
type SendMessageToChild struct {
	ActorName string
	Message   any
}

type ChildNotFound struct {
	ActorName string
}

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
}
