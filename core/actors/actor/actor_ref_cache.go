package actor

import (
	cmap "github.com/orcaman/concurrent-map"
)

type ActorRefCache struct {
	cache cmap.ConcurrentMap
}

func NewActorRefCache() *ActorRefCache {
	return &ActorRefCache{
		cache: cmap.New(),
	}
}

func (c *ActorRefCache) Set(actorName string, value *ActorRef) {
	c.cache.Set(actorName, value)
}

func (c *ActorRefCache) Get(actorName string) (*ActorRef, bool) {
	if v, ok := c.cache.Get(actorName); ok {
		return v.(*ActorRef), true
	}
	return nil, false
}

func (c *ActorRefCache) Del(actorName string) {
	c.cache.Remove(actorName)
}
