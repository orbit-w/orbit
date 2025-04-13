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

type Context struct {
	metaData  *Meta
	context   actor.Context
	actorName string
	pattern   string
}

func NewContext(meta *Meta, actorName, pattern string) IContext {
	return &Context{
		metaData:  meta,
		actorName: actorName,
		pattern:   pattern,
	}
}

func (c *Context) SetMetaData(meta *Meta) {
	c.metaData = meta
}

func (c *Context) GetMetaData() *Meta {
	return c.metaData
}

func (c *Context) GetActorName() string {
	return c.actorName
}

func (c *Context) GetPattern() string {
	return c.pattern
}

func (c *Context) GetActorContext() actor.Context {
	return c.context
}

func (c *Context) SetActorContext(context actor.Context) {
	c.context = context
}

func (c *Context) GetServerId() string {
	if c.metaData == nil {
		return ""
	}
	return c.metaData.ServerId
}
