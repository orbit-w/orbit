package actor

import (
	"context"
	"fmt"

	"gitee.com/orbit-w/orbit/lib/unipue_task_exec"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"
)

/*
	ActorMetaCache 用于缓存Actor的元数据
	缓存结构:
	actor_meta:{actorName} -> Meta

	type Meta struct {
		...
		ActorName     string                 `protobuf:"bytes,1,opt,name=ActorName,proto3" json:"ActorName,omitempty"`
		Pattern       string                 `protobuf:"bytes,2,opt,name=Pattern,proto3" json:"Pattern,omitempty"`
		ServerId      string                 `protobuf:"bytes,3,opt,name=ServerId,proto3" json:"ServerId,omitempty"` //初始逻辑服ID
		Dispatcher    *Dispatcher            `protobuf:"bytes,4,opt,name=Dispatcher,proto3" json:"Dispatcher,omitempty"`
		...
	}
*/

type ActorMetaCache struct {
	cli   *redis.Client
	cache cmap.ConcurrentMap
	exec  *unipue_task_exec.UniqueTaskExecutor
}

func NewMetaCache(cli *redis.Client) *ActorMetaCache {
	return &ActorMetaCache{
		cli:   cli,
		cache: cmap.New(),
		exec:  unipue_task_exec.NewUniqueTaskExecutor(),
	}
}

func (c *ActorMetaCache) Load(actorName string) (*Meta, error) {
	if v, exists := c.cache.Get(actorName); exists {
		return v.(*Meta), nil
	}

	re := c.exec.ExecuteOnce(actorName, func() any {
		key := genRedisKey(actorName)
		content, err := c.cli.Get(context.Background(), key).Result()
		if err != nil {
			return err
		}

		ar := &Meta{}
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
	case *Meta:
		return v, nil
	default:
		return nil, fmt.Errorf("unknown error: %v", re)
	}
}

func (c *ActorMetaCache) Store(actorName string, value *Meta) (*Meta, error) {
	content, err := proto.Marshal(value)
	if err != nil {
		return nil, err
	}

	c.cli.Set(context.Background(), genRedisKey(actorName), content, 0)
	c.cache.Set(actorName, value)
	return value, nil
}

func (c *ActorMetaCache) Set(key string, value *Meta) {
	c.cache.Set(key, value)
}

func (c *ActorMetaCache) Get(key string) (*Meta, bool) {
	if v, ok := c.cache.Get(key); ok {
		return v.(*Meta), true
	}
	return nil, false
}

func (c *ActorMetaCache) Del(key string) {
	c.cli.Del(context.Background(), key)
	c.cache.Remove(key)
}

func genRedisKey(actorName string) string {
	return fmt.Sprintf("actor_meta:%s", actorName)
}
