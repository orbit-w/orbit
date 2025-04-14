package actor

import actor "github.com/asynkron/protoactor-go/actor"

type IContext interface {
	IBaseContext
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
