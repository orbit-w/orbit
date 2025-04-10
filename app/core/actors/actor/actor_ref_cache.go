package actor

import (
	"context"
	"fmt"

	"gitee.com/orbit-w/orbit/lib/unipue_task_exec"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"
)

type ActorRefCache struct {
	cli   *redis.Client
	cache cmap.ConcurrentMap
	exec  *unipue_task_exec.UniqueTaskExecutor
}

func NewActorRefCache(cli *redis.Client) *ActorRefCache {
	return &ActorRefCache{
		cli:   cli,
		cache: cmap.New(),
		exec:  unipue_task_exec.NewUniqueTaskExecutor(),
	}
}

func (c *ActorRefCache) Load(actorName string) (*ActorRef, error) {
	if ar, exists := c.cache.Get(actorName); exists {
		return ar.(*ActorRef), nil
	}

	re := c.exec.ExecuteOnce(actorName, func() any {
		key := genRedisKey(actorName)
		content, err := c.cli.Get(context.Background(), key).Result()
		if err != nil {
			return err
		}

		ar := &ActorRef{}
		err = proto.Unmarshal([]byte(content), ar)
		if err != nil {
			return err
		}

		c.cache.Set(actorName, ar)
		return ar
	})

	switch v := re.(type) {
	case error:
		return nil, v
	case *ActorRef:
		return v, nil
	default:
		return nil, fmt.Errorf("unknown error: %v", re)
	}
}

func (c *ActorRefCache) Store(actorName string, value *ActorRef) (*ActorRef, error) {
	content, err := proto.Marshal(value)
	if err != nil {
		return nil, err
	}

	c.cli.Set(context.Background(), genRedisKey(actorName), content, 0)
	c.cache.Set(actorName, value)
	return value, nil
}

func (c *ActorRefCache) Set(key string, value *ActorRef) {
	c.cache.Set(key, value)
}

func (c *ActorRefCache) Get(key string) (*ActorRef, bool) {
	if v, ok := c.cache.Get(key); ok {
		return v.(*ActorRef), true
	}
	return nil, false
}

func (c *ActorRefCache) Del(key string) {
	c.cli.Del(context.Background(), key)
}

func genRedisKey(actorName string) string {
	return fmt.Sprintf("actor_ref:%s", actorName)
}
