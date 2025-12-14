package v1

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"gitee.com/orbit-w/meteor/bases/misc/utils"
	"gitee.com/orbit-w/meteor/modules/mlog"
	"gitee.com/orbit-w/meteor/modules/unique_task_exec"
	"github.com/orca-zhang/ecache"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

const (
	CachePattern = "meteor_cache"
	CacheTTL     = 15 * time.Minute

	topic = "meteor_cache"

	stateReady   = 1
	stateStopped = 2

	defaultPublishTimeout = time.Second * 5
	defaultPingTimeout    = time.Second * 5
)

type Cache[V proto.Message] struct {
	state   atomic.Uint32
	topic   string
	cli     redis.UniversalClient
	ttl     time.Duration
	Pattern string
	cache   *ecache.Cache
	exec    *unique_task_exec.UniqueTaskExecutor
	factory func() V
	log     *mlog.Logger
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewCache[V proto.Message](cli redis.UniversalClient, pattern string, factory func() V) *Cache[V] {
	// 初始化 ecache，容量为 128 个分片，最大存储 1024 个元素，过期时间为 15 分钟
	cache := ecache.NewLRUCache(128, 1024, CacheTTL)

	ctx, cancel := context.WithCancel(context.Background())
	c := &Cache[V]{
		cli:     cli,
		ttl:     CacheTTL,
		Pattern: pattern,
		cache:   cache,
		exec:    unique_task_exec.NewUniqueTaskExecutor(),
		factory: factory,
		log:     mlog.WithPrefix("cache"),
		topic:   topic + "_" + pattern,
		ctx:     ctx,
		cancel:  cancel,
	}

	c.state.Store(stateReady)
	go c.runSubscribe()
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
	c.onDel(key)
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
	c.onDel(key)
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

func (c *Cache[V]) Close() {
	c.state.CompareAndSwap(stateReady, stateStopped)
	if c.cancel != nil {
		c.cancel() // 取消 context，让 goroutine 立即退出
	}
}

func (c *Cache[V]) runSubscribe() {
	defer utils.RecoverPanic()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		for c.cli == nil || !c.ping() {
			if c.state.Load() == stateStopped {
				return
			}

			// 简单的 sleep，但无法中断
			time.Sleep(10 * time.Millisecond)

			// 睡眠后再次检查 context
			select {
			case <-c.ctx.Done():
				return
			default:
			}
		}

		if c.state.Load() == stateStopped {
			return
		}

		c.sub()
	}
}

func (c *Cache[V]) sub() {
	defer utils.RecoverPanic()

	pubsub := c.cli.Subscribe(c.ctx, c.topic)
	ch := pubsub.Channel()
	defer pubsub.Close()

	for {
		select {
		case <-c.ctx.Done():
			// context 被取消，立即退出
			return
		case msg, ok := <-ch:
			if !ok {
				// channel 已关闭
				c.log.Error("subscribe channel closed", zap.String("topic", c.topic))
				return
			}
			if msg != nil {
				key := msg.Payload
				c.cache.Del(key)
			}
		}
	}
}

func (c *Cache[V]) pub(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultPublishTimeout)
	defer cancel()
	return c.cli.Publish(ctx, c.topic, key).Err()
}

func (c *Cache[V]) ping() bool {
	ctx, cancel := context.WithTimeout(context.Background(), defaultPingTimeout)
	defer cancel()
	return c.cli != nil && c.cli.Ping(ctx).Err() == nil
}

func (c *Cache[V]) onDel(key string) error {
	// pub to remote nodes
	if c.pub(key) == nil {
		return nil
	}
	c.cache.Del(key)
	return nil
}

func GenItemKey(pattern string, key string) string {
	return strings.Join([]string{CachePattern, pattern, key}, ":")
}
