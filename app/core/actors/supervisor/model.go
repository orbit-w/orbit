package supervisor

type StartChildActor struct {
	ActorName string
	Pattern   string
	Message   any
}

// Message types for actor communication
type SendMessageToChild struct {
	ActorId string
	Message any
}

type ChildNotFound struct {
	ActorId string
}
