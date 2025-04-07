package manager

import (
	"github.com/asynkron/protoactor-go/actor"
	cmap "github.com/orcaman/concurrent-map/v2"
)

var (
	actorsCache *ActorsCache
)

type ActorsCache struct {
	cache cmap.ConcurrentMap[string, *actor.PID]
}

func NewActorsCache() *ActorsCache {
	return &ActorsCache{
		cache: cmap.New[*actor.PID](),
	}
}

func (c *ActorsCache) Get(actorName string) (*actor.PID, bool) {
	pid, ok := c.cache.Get(actorName)
	if !ok {
		return nil, false
	}
	return pid, true
}

func (c *ActorsCache) Set(actorName string, pid *actor.PID) {
	c.cache.Set(actorName, pid)
}

func (c *ActorsCache) Delete(actorName string) {
	c.cache.Remove(actorName)
}
