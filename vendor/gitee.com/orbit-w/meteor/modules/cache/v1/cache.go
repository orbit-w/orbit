package v1

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"gitee.com/orbit-w/meteor/modules/mlog"
	"gitee.com/orbit-w/meteor/modules/unique_task_exec"
	"github.com/orca-zhang/ecache"
	"github.com/orca-zhang/ecache/dist"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

const (
	CachePattern = "meteor_cache"
	CacheTTL     = 15 * time.Minute
)

var (
	distInitOnce sync.Once
)

// universalClientAdapter 适配 redis.UniversalClient 到 dist.RedisCli 接口
type universalClientAdapter struct {
	cli redis.UniversalClient
	ctx context.Context
}

func (a *universalClientAdapter) OK() bool {
	return a.cli != nil && a.cli.Ping(a.ctx).Err() == nil
}

func (a *universalClientAdapter) Pub(channel, payload string) error {
	return a.cli.Publish(a.ctx, channel, payload).Err()
}

func (a *universalClientAdapter) Sub(channel string, callback func(payload string)) error {
	pubsub := a.cli.Subscribe(a.ctx, channel)
	ch := pubsub.Channel()
	go func() {
		defer pubsub.Close()
		for msg := range ch {
			if msg != nil {
				callback(msg.Payload)
			}
		}
	}()
	return nil
}

type Cache[V proto.Message] struct {
	cli     redis.UniversalClient
	ttl     time.Duration
	Pattern string
	cache   *ecache.Cache
	exec    *unique_task_exec.UniqueTaskExecutor
	factory func() V
	log     *mlog.Logger
}

func NewCache[V proto.Message](cli redis.UniversalClient, pattern string, factory func() V) *Cache[V] {
	// 初始化 ecache，容量为 128 个分片，最大存储 1024 个元素，过期时间为 15 分钟
	cache := ecache.NewLRUCache(128, 1024, CacheTTL)

	c := &Cache[V]{
		cli:     cli,
		ttl:     CacheTTL,
		Pattern: pattern,
		cache:   cache,
		exec:    unique_task_exec.NewUniqueTaskExecutor(),
		factory: factory,
		log:     mlog.WithPrefix("cache"),
	}

	// 初始化 dist 组件（只初始化一次）
	distInitOnce.Do(func() {
		// 使用适配器包装 UniversalClient
		adapter := &universalClientAdapter{
			cli: cli,
			ctx: context.Background(),
		}
		dist.Init(adapter)
	})

	// 将缓存实例绑定到 pattern 对应的 pool
	// dist 包会自动处理分布式一致性，当调用 dist.OnDel(pattern, key) 时
	// 会通知所有节点删除该 pool 下的所有缓存实例中的 key
	dist.Bind(pattern, cache)

	return c
}

// 更新数据库，并通知所有节点（包括本节点）删除旧缓存，保证分布式一致性
// 其他节点下次访问时会从 Redis 重新加载最新数据
// 当前结点跟其他结点一样，都会有时间窗口去更新Redis，但最终Redis中的数据会是一致的
func (c *Cache[V]) Set(key string, value V) error {
	// 序列化 value
	data, err := proto.Marshal(value)
	if err != nil {
		return err
	}

	// 存储到 Redis
	err = c.cli.Set(context.Background(), GenItemKey(c.Pattern, key), data, c.ttl).Err()
	if err != nil {
		return err
	}

	// 通知其他节点删除旧缓存，保证分布式一致性
	// 其他节点下次访问时会从 Redis 重新加载最新数据
	dist.OnDel(c.Pattern, key)
	return nil
}

func (c *Cache[V]) Get(key string) (V, error) {
	// 先从本地缓存获取
	if v, ok := c.cache.Get(key); ok {
		if value, ok := v.(V); ok {
			return value, nil
		}
	}
	return c.Load(key)
}

func (c *Cache[V]) Remove(key string) error {
	if err := c.cli.Del(context.Background(), GenItemKey(c.Pattern, key)).Err(); err != nil {
		c.log.Error("remove cache del failed", zap.Error(err))
		return err
	}
	// 使用 dist.OnDel 触发分布式删除
	// 这会通知所有节点（包括本节点）删除该 pool 下所有缓存实例中的 key
	dist.OnDel(c.Pattern, key)
	return nil
}

func (c *Cache[V]) Load(key string) (V, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	item := c.exec.ExecuteOnceWithContext(ctx, key, func() any {
		// 再次检查本地缓存（可能在并发情况下已经被其他 goroutine 加载）
		if v, ok := c.cache.Get(key); ok {
			if value, ok := v.(V); ok {
				return value
			}
		}
		// 从 Redis 加载
		v, err := c.cli.Get(context.Background(), GenItemKey(c.Pattern, key)).Result()
		if err != nil {
			return err
		}
		value := c.factory()
		if err := proto.Unmarshal([]byte(v), value); err != nil {
			return err
		}
		// 存入本地缓存
		c.cache.Put(key, value)
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

// Stop 停止缓存（目前 dist 包会自动管理，此方法保留用于未来扩展）
func (c *Cache[V]) Stop() {
	// dist 包会自动管理订阅和清理，无需手动处理
	// 如果需要清理，可以在这里添加逻辑
}

func GenItemKey(pattern string, key string) string {
	return strings.Join([]string{CachePattern, pattern, key}, ":")
}
