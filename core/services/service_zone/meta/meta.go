package zone_meta

import (
	sync "sync"
	"time"

	cachev1 "gitee.com/orbit-w/meteor/modules/cache/v1"
	"github.com/redis/go-redis/v9"
)

const (
	// CacheTTL 缓存过期时间：15分钟
	CacheTTL = 15 * time.Minute
	// CachePattern Redis key 前缀
	CachePattern = "zone_meta"
	// PubSubTopic 发布订阅主题，用于分布式缓存失效通知
	PubSubTopic = "ecache:zone_meta:invalidate"
)

var (
	cache *cachev1.Cache[*ZoneMeta]
	once  sync.Once
)

type ZoneMetaService struct {
	cli redis.UniversalClient
}

func NewZoneMetaService(_cli redis.UniversalClient) *ZoneMetaService {
	return &ZoneMetaService{
		cli: _cli,
	}
}

func (s *ZoneMetaService) Start() error {
	once.Do(func() {
		cache = cachev1.NewCache(s.cli, CachePattern, func() *ZoneMeta {
			return &ZoneMeta{}
		})
	})
	return nil
}

func (s *ZoneMetaService) Stop() error {
	return nil
}

func NewZoneMeta(id string, pattern int32, dispatcher *ZoneDispatcher) *ZoneMeta {
	return &ZoneMeta{
		Id:         id,
		Pattern:    pattern,
		Dispatcher: dispatcher,
	}
}

func SetZoneMeta(id string, pattern int32, dispatcher *ZoneDispatcher) (*ZoneMeta, error) {
	meta := NewZoneMeta(id, pattern, dispatcher)
	err := cache.Set(id, meta)
	if err != nil {
		return nil, err
	}
	return meta, nil
}

func GetZoneMeta(id string) (*ZoneMeta, error) {
	return cache.Get(id)
}
