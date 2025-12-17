package servicezone

import (
	mmeobj "gitee.com/orbit-w/orbit/internal/game/mme"
	"gitee.com/orbit-w/orbit/internal/game/proto/mme"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type IServiceZone interface {
	AddEntity(entity mmeobj.IEntity)
	SetEntity(entity mmeobj.IEntity)
	RemoveEntity(targetId int64)
	ISubscriber
}

type ISubscriber interface {
	// 根据 ID 列表订阅
	SubscribeByIds(subscriberId string, entityIds []int64) []mmeobj.IEntity
	// 根据类型订阅
	SubscribeByType(subscriberId string, entityTypes []string) []mmeobj.IEntity
	// 订阅所有
	SubscribeAll(subscriberId string) []mmeobj.IEntity
	// 取消订阅
	Unsubscribe(subscriberId string)
	// 获取已订阅的 Entities
	GetSubscribedEntities(subscriberId string) []mmeobj.IEntity
	// 获取订阅策略类型
	GetSubscriberStrategyType(subscriberId string) (mme.SubscribeStrategyType, bool)
}

type ILoadResponse interface {
	IsSuccess() bool
	IsExists() bool
	GetError() error
	GetData() bson.Raw
}
