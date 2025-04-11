package actor

import (
	"sync/atomic"

	"github.com/asynkron/protoactor-go/actor"
	cmap "github.com/orcaman/concurrent-map/v2"
)

var (
	actorsCache *ActorsCache
)

const (
	StateNone = iota
	StateStopped
)

type ActorItem struct {
	Pattern string
	PID     *actor.PID
	State   atomic.Int32
}

func (item *ActorItem) IsStopped() bool {
	return item.State.Load() == StateStopped
}

type ActorsCache struct {
	cache cmap.ConcurrentMap[string, *ActorItem]
}

func NewActorsCache() *ActorsCache {
	return &ActorsCache{
		cache: cmap.New[*ActorItem](),
	}
}

func (c *ActorsCache) Get(actorName string) (*actor.PID, bool) {
	item, ok := c.cache.Get(actorName)
	if !ok {
		return nil, false
	}
	if item.IsStopped() {
		c.cache.Remove(actorName)
		if item.PID != nil {
			system := actor.NewActorSystem()
			system.Root.Stop(item.PID)
		}
		return nil, false
	}
	return item.PID, true
}

func (c *ActorsCache) GetItem(actorName string) (*ActorItem, bool) {
	item, ok := c.cache.Get(actorName)
	if !ok {
		return nil, false
	}
	if item.IsStopped() {
		c.cache.Remove(actorName)
		if item.PID != nil {
			system := actor.NewActorSystem()
			system.Root.Stop(item.PID)
		}
		return nil, false
	}
	return item, true
}

func (c *ActorsCache) Set(actorName, pattern string, pid *actor.PID) {
	c.cache.Set(actorName, &ActorItem{
		Pattern: pattern,
		PID:     pid,
		State:   atomic.Int32{},
	})
}

func (c *ActorsCache) CompareAndSwap(actorName string, old, new int32) {
	item, ok := c.cache.Get(actorName)
	if !ok {
		return
	}

	if item.State.CompareAndSwap(old, new) {
		c.cache.Set(actorName, item)
	}
}

func (c *ActorsCache) Delete(actorName string) {
	c.cache.Remove(actorName)
}
