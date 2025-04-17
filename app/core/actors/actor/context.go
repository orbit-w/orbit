package actor

import (
	"time"

	actor "github.com/asynkron/protoactor-go/actor"
)

type IContext interface {
	IBaseContext
	ITimerContext
}

type IBaseContext interface {
	SetMetaData(meta *Meta)
	GetMetaData() *Meta
	GetActorName() string
	GetPattern() string
	GetActorContext() actor.Context
	SetActorContext(context actor.Context)
	GetServerId() string
}

type ITimerContext interface {
	AddTimerRepeat(key string, duration time.Duration, msg any) *Timer
	AddTimerOnce(key string, duration time.Duration, msg any) *Timer
	RemoveTimer(key string)
}
