package manager

import (
	"sync"

	"github.com/asynkron/protoactor-go/actor"
)

var (
	actorsCache *ActorsCache
)

type ActorsCache struct {
	cache map[string]*actor.PID
	mu    sync.RWMutex
}

func NewActorsCache() *ActorsCache {
	return &ActorsCache{
		cache: make(map[string]*actor.PID),
		mu:    sync.RWMutex{},
	}
}

func (c *ActorsCache) Get(actorName string) (*actor.PID, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	pid, ok := c.cache[actorName]
	return pid, ok
}

func (c *ActorsCache) Set(actorName string, pid *actor.PID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[actorName] = pid
}

func (c *ActorsCache) Delete(actorSystem *actor.ActorSystem, actorName string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if pid, ok := c.cache[actorName]; ok {
		delete(c.cache, actorName)
		actorSystem.Root.Stop(pid)
	}
}
