package actor

import (
	sync "sync"
	"time"

	actor "github.com/asynkron/protoactor-go/actor"
)

type Process struct {
	State     int8
	ActorName string
	Pattern   string
	Props     *Props
	PID       *actor.PID
	rw        sync.RWMutex // 读写锁, 用于保护ActorProcess的状态
}

const (
	StateNone = iota
	StateStopped
)

func NewActorProcess(actorName, pattern string, child *actor.PID, props *Props) *Process {
	return &Process{
		ActorName: actorName,
		Pattern:   pattern,
		Props:     props,
		State:     StateNone,
		rw:        sync.RWMutex{},
		PID:       child,
	}
}

func (p *Process) IsStopped() bool {
	p.rw.RLock()
	defer p.rw.RUnlock()
	return p.stopped()
}

func (p *Process) stopped() bool {
	return p.State == StateStopped
}

func (p *Process) Stop() {
	p.rw.Lock()
	defer p.rw.Unlock()
	p.State = StateStopped
}

func (p *Process) GetPID() *actor.PID {
	return p.PID
}

func (p *Process) RequestFuture(msg any, timeout ...time.Duration) (any, error) {
	p.rw.RLock()
	if p.stopped() {
		p.rw.RUnlock()
		return nil, ErrActorStopped
	}

	future := System.ActorSystem().Root.RequestFuture(p.PID, &RequestMessage{
		MsgType: MessageTypeRequest,
		Message: msg,
	}, parseTimeout(timeout...))

	p.rw.RUnlock()

	result, err := future.Result()
	if err != nil {
		return nil, err
	}

	switch v := result.(type) {
	case error:
		return nil, v
	default:
		return v, nil
	}
}

func (p *Process) Send(msg any) error {
	p.rw.RLock()
	defer p.rw.RUnlock()
	if p.stopped() {
		return ErrActorStopped
	}

	System.ActorSystem().Root.Send(p.PID, &RequestMessage{
		MsgType: MessageTypeSend,
		Message: msg,
	})
	return nil
}
