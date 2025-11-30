package zone_meta

import (
	"time"

	cachev1 "gitee.com/orbit-w/meteor/modules/cache/v1"
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
)

func NewZoneMeta(id string, pattern int32, dispatcher *ZoneDispatcher) *ZoneMeta {
	return &ZoneMeta{
		Id:         id,
		Pattern:    pattern,
		Dispatcher: dispatcher,
	}
}

func SetZoneMeta(id string, pattern int32, dispatcher *ZoneDispatcher) error {
	return cache.Set(id, NewZoneMeta(id, pattern, dispatcher))
}

func GetZoneMeta(id string) (*ZoneMeta, error) {
	return cache.Get(id)
}
