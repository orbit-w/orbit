package v1

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"gitee.com/orbit-w/meteor/modules/mlog"
	"gitee.com/orbit-w/meteor/modules/subpub/subpub_redis"
	"gitee.com/orbit-w/meteor/modules/unique_task_exec"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"google.golang.org/protobuf/proto"
)

const (
	CachePattern = "meteor_cache"
)

type Cache[V proto.Message] struct {
	cli     redis.UniversalClient
	ttl     time.Duration
	Pattern string
	cache   cmap.ConcurrentMap[string, *Item[V]]
	exec    *unique_task_exec.UniqueTaskExecutor
	subpub  subpub_redis.IPubSub
	factory func() V
	log     *mlog.Logger
}

func NewCache[V proto.Message](cli redis.UniversalClient, pattern string, factory func() V) *Cache[V] {
	c := &Cache[V]{
		cli:     cli,
		ttl:     15 * time.Minute,
		Pattern: pattern,
		cache:   cmap.New[*Item[V]](),
		exec:    unique_task_exec.NewUniqueTaskExecutor(),
		factory: factory,
		log:     mlog.WithPrefix("cache"),
	}
	c.subpub = subpub_redis.NewPubSub(cli, subpub_redis.CodecString, pattern, c.invoke())
	return c
}

func (c *Cache[V]) Set(key string, value V) error {
	err := c.cli.Set(context.Background(), GenItemKey(c.Pattern, key), value, c.ttl).Err()
	if err != nil {
		return err
	}

	if err := c.subpub.Publish(0, key); err != nil {
		return err
	}
	c.cache.Set(key, &Item[V]{
		Value:    value,
		ExpireAt: time.Now().Add(c.ttl),
	})
	return nil
}

func (c *Cache[V]) Get(key string) (V, error) {
	if item, ok := c.cache.Get(key); ok {
		// 检查是否过期
		if !time.Now().After(item.ExpireAt) {
			return item.Value, nil
		}
		c.cache.Remove(key)
	}
	return c.Load(key)
}

// TODO:
// 边缘情况处理：
//
//	1）极限情况下，A服务删除了数据，并发布，B在临界区成功读去到旧数据，并放入本地缓存中。这种情况下依赖TTL过期机制清除点B中的旧缓存。
func (c *Cache[V]) Remove(key string) error {
	if err := c.cli.Del(context.Background(), GenItemKey(c.Pattern, key)).Err(); err != nil {
		c.log.Error("remove cache del failed", zap.Error(err))
		return err
	}
	c.cache.Remove(key)
	if err := c.subpub.Publish(0, key); err != nil {
		c.log.Error("remove cache publish failed", zap.Error(err))
	}
	return nil
}

func (c *Cache[V]) Load(key string) (V, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	item := c.exec.ExecuteOnceWithContext(ctx, key, func() any {
		if cachedItem, ok := c.cache.Get(key); ok {
			return cachedItem.Value
		}
		v, err := c.cli.Get(context.Background(), GenItemKey(c.Pattern, key)).Result()
		if err != nil {
			return err
		}
		value := c.factory()
		if err := proto.Unmarshal([]byte(v), value); err != nil {
			return err
		}
		// 在缓存中设置值，如果缓存中已存在，则不设置
		cacheItem := &Item[V]{
			Value:    value,
			ExpireAt: time.Now().Add(c.ttl),
		}
		_ = c.cache.SetIfAbsent(key, cacheItem)
		return value
	})

	switch result := item.(type) {
	case error:
		var zeroV V
		return zeroV, result
	case V:
		return result, nil
	default:
		var zeroV V
		return zeroV, fmt.Errorf("unknown type: %T", result)
	}
}

func (c *Cache[V]) invoke() func(pid int64, body []byte, err error) {
	return func(pid int64, body []byte, err error) {
		if err != nil {
			log.Println("invoke subpub failed: ", err)
			return
		}
		key := string(body)
		c.cache.Remove(key)
	}
}

func GenItemKey(pattern string, key string) string {
	return strings.Join([]string{CachePattern, pattern, key}, ":")
}
