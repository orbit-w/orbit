package actor

import (
	cmap "github.com/orcaman/concurrent-map/v2"
)

var (
	actorsCache *ActorsCache
)

type ActorsCache struct {
	cache cmap.ConcurrentMap[string, *ActorProcess]
}

func NewActorsCache() *ActorsCache {
	return &ActorsCache{
		cache: cmap.New[*ActorProcess](),
	}
}

func (c *ActorsCache) Exist(actorName string) bool {
	return c.cache.Has(actorName)
}

func (c *ActorsCache) Get(actorName string) (*ActorProcess, bool) {
	item, ok := c.cache.Get(actorName)
	if !ok {
		return nil, false
	}
	return item, true
}

func (c *ActorsCache) Set(actorName string, p *ActorProcess) {
	c.cache.Set(actorName, p)
}

func (c *ActorsCache) Delete(actorName string) {
	c.cache.Remove(actorName)
}
