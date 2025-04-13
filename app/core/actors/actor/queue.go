package actor

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

func NewItem(actorName, pattern string, child *actor.PID) *Item {
	return &Item{
		ActorName: actorName,
		Pattern:   pattern,
		Future:    make([]*actor.PID, 0),
		Child:     child,
	}
}

func (i *Item) AddFuture(future ...*actor.PID) {
	if i.Future == nil {
		i.Future = make([]*actor.PID, 0)
	}
	if future != nil {
		i.Future = append(i.Future, future...)
	}
}

func (i *Item) FuturesNum() int {
	num := 0
	for _, future := range i.Future {
		if future != nil {
			num++
		}
	}
	return num
}

func (i *Item) Futures() []*actor.PID {
	for j := range i.Future {
		if i.Future[j] == nil {
			i.Future = append(i.Future[:j], i.Future[j+1:]...)
		}
	}
	return i.Future
}

type Queue struct {
	pq *priority_queue.PriorityQueue[string, *Item, int64]
}

func NewPriorityQueue() *Queue {
	return &Queue{
		pq: priority_queue.New[string, *Item, int64](),
	}
}

func (q *Queue) Insert(actorName string, item *Item, priority int64) {
	ok := q.pq.Exist(actorName)
	if ok {
		return
	}
	q.pq.Push(actorName, item, priority)
}

func (q *Queue) PushFuture(actorName string, future *actor.PID) {
	ent, ok := q.pq.Get(actorName)
	if !ok {
		return
	}
	v := ent.Value.GetValue()
	v.AddFuture(future)
	q.pq.UpdateValue(actorName, v)
}

func (q *Queue) Pop(key string) (*Item, bool) {
	return q.pq.PopK(key)
}

func (q *Queue) Exists(key string) bool {
	return q.pq.Exist(key)
}

func (q *Queue) PopAndRangeWithKey(key string, iter func(name, pattern string, child, future *actor.PID) bool) (*Item, bool) {
	v, ok := q.pq.PopK(key)
	if !ok {
		return nil, false
	}
	for i := range v.Future {
		f := v.Future[i]
		if !iter(key, v.Pattern, v.Child, f) {
			break
		}
	}
	return v, true
}
