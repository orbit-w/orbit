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
