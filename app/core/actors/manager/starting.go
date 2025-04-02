package manager

import (
	"errors"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/orbit-w/meteor/bases/container/linked_list"
)

type Startings struct {
	starting map[string]*Futures
}

type Futures struct {
	Child *actor.PID
	link  *linked_list.LinkedList[string, *actor.Future]
}

func NewSMap() *Startings {
	return &Startings{
		starting: make(map[string]*Futures),
	}
}

func (s *Startings) Set(key string, child *actor.PID, future *actor.Future) error {
	if child == nil {
		return errors.New("child is nil")
	}
	if s.Exists(key) {
		return errors.New("key already exists")
	}

	f := &Futures{
		Child: child,
		link:  linked_list.New[string, *actor.Future](),
	}
	f.link.LPush(key, future)
	s.starting[key] = f
	return nil
}

func (s *Startings) Push(key string, future *actor.Future) error {
	f, ok := s.starting[key]
	if !ok {
		return errors.New("key not found")
	}
	f.link.LPush(key, future)
	return nil
}

func (s *Startings) Delete(key string) {
	delete(s.starting, key)
}

func (s *Startings) PopAndRange(key string, iter func(child *actor.PID, future *actor.Future)) (*actor.PID, bool) {
	f, ok := s.starting[key]
	if !ok {
		return nil, false
	}
	delete(s.starting, key)
	f.link.RRange(f.link.Len(), func(k string, v *actor.Future) {
		iter(f.Child, v)
	})
	return f.Child, true
}

func (s *Startings) Len() int {
	return len(s.starting)
}

func (s *Startings) Clear() {
	s.starting = make(map[string]*Futures)
}

func (s *Startings) Exists(key string) bool {
	_, ok := s.starting[key]
	return ok
}
