package zone_meta

import (
	sync "sync"
	"time"

	cachev1 "gitee.com/orbit-w/meteor/modules/cache/v1"
	"gitee.com/orbit-w/meteor/modules/database/rdb"
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
	cache             *cachev1.Cache[*ZoneMeta]
	once              sync.Once
	dispatcherService *ZoneDispatcherService
)

type ZoneMetaService struct {
}

func NewZoneMetaService() *ZoneMetaService {
	return &ZoneMetaService{}
}

func (s *ZoneMetaService) Start() error {
	once.Do(func() {
		cache = cachev1.NewCache(rdb.UniversalClient(), CachePattern, func() *ZoneMeta {
			return &ZoneMeta{}
		})
		dispatcherService = NewZoneDispatcherService()
	})
	return nil
}

func (s *ZoneMetaService) Stop() error {
	cache.Close()
	return nil
}

func NewZoneMeta(id string, pattern int32, dispatcher *ZoneDispatcher) (*ZoneMeta, error) {
	meta := &ZoneMeta{
		Id:         id,
		Pattern:    pattern,
		Dispatcher: dispatcher,
	}
	if dispatcher != nil {
		node, err := dispatcherService.SelectNode(dispatcher)
		if err != nil {
			return nil, err
		}
		meta.node = node
	}
	return meta, nil
}

func SetZoneMeta(id string, pattern int32, dispatcher *ZoneDispatcher) (*ZoneMeta, error) {
	meta, err := NewZoneMeta(id, pattern, dispatcher)
	if err != nil {
		return nil, err
	}
	err = cache.Set(id, meta)
	if err != nil {
		return nil, err
	}
	return meta, nil
}

func GetZoneMeta(id string) (*ZoneMeta, error) {
	meta, err := cache.Get(id)
	if err != nil {
		return nil, err
	}
	if meta.IsNodeExpired() || meta.NodeInvalid() {
		if meta.Dispatcher != nil {
			node, err := dispatcherService.SelectNode(meta.Dispatcher)
			if err != nil {
				return nil, err
			}
			meta.node = node
		}

	}
	return meta, nil
}
