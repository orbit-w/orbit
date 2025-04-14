package actor

import (
	actor "github.com/asynkron/protoactor-go/actor"
)

type Props struct {
	InitHandler func() error
	Meta        *Meta
	kvs         map[string]any
}

func NewProps() *Props {
	return &Props{
		kvs: make(map[string]any),
	}
}

func (pp *Props) GetInitHandler() func() error {
	if pp == nil {
		return nil
	}
	if pp.InitHandler == nil {
		return nil
	}
	return pp.InitHandler
}

func (pp *Props) GetMeta() *Meta {
	if pp == nil {
		return nil
	}
	if pp.Meta == nil {
		return nil
	}
	return pp.Meta
}

func (pp *Props) GetKvs(iter func(k string, v any)) {
	if pp == nil {
		return
	}
	if pp.kvs == nil {
		return
	}
	for k, v := range pp.kvs {
		iter(k, v)
	}
}

func (pp *Props) getOrCreateActorPID(name, pattern string) *actor.PID {
	pid, err := GetOrStartActor(name, pattern, pp)
	if err != nil {
		panic(err)
	}
	return pid
}

type PropsOption func(pp *Props)

func WithInitHandler(handler func() error) PropsOption {
	return func(pp *Props) {
		pp.InitHandler = handler
	}
}

func WithMeta(meta *Meta) PropsOption {
	return func(pp *Props) {
		pp.Meta = meta
	}
}
