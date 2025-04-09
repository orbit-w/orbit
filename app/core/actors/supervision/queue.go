package supervision

import (
	"gitee.com/orbit-w/meteor/bases/container/priority_queue"
	"github.com/asynkron/protoactor-go/actor"
)

type Item struct {
	ActorName string
	Pattern   string
	Future    []*actor.PID
	Child     *actor.PID
}

type Queue struct {
	pq *priority_queue.PriorityQueue[string, *Item, int64]
}

func NewPriorityQueue() *Queue {
	return &Queue{
		pq: priority_queue.New[string, *Item, int64](),
	}
}

func (q *Queue) Insert(actorName, pattern string, child, future *actor.PID, priority int64) {
	ok := q.pq.Exist(actorName)
	if ok {
		return
	}
	q.pq.Push(actorName, &Item{
		ActorName: actorName,
		Pattern:   pattern,
		Future:    []*actor.PID{future},
		Child:     child,
	}, priority)
}

func (q *Queue) BatchInsert(actorName, pattern string, child *actor.PID, futures []*actor.PID, priority int64) {
	ok := q.pq.Exist(actorName)
	if ok {
		return
	}

	q.pq.Push(actorName, &Item{
		ActorName: actorName,
		Pattern:   pattern,
		Future:    futures,
		Child:     child,
	}, priority)
}

func (q *Queue) Push(actorName string, future *actor.PID) {
	ent, ok := q.pq.Get(actorName)
	if !ok {
		return
	}
	v := ent.Value.GetValue()
	v.Future = append(v.Future, future)
	q.pq.UpdateValue(actorName, v)
}

func (q *Queue) Pop(key string) (*Item, bool) {
	return q.pq.PopK(key)
}

func (q *Queue) Exists(key string) bool {
	return q.pq.Exist(key)
}

func (q *Queue) PopAndRangeWithKey(key string, iter func(name string, child, future *actor.PID) bool) (*actor.PID, bool) {
	v, ok := q.pq.PopK(key)
	if !ok {
		return nil, false
	}
	for i := range v.Future {
		f := v.Future[i]
		if !iter(key, v.Child, f) {
			break
		}
	}
	return v.Child, true
}
