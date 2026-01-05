package servicezone

import (
	"gitee.com/orbit-w/orbit/internal/game/mme_agent/entities"
	"gitee.com/orbit-w/orbit/pkg/proto/mme"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type IServiceZone interface {
	AddEntity(entity entities.IEntity)
	SetEntity(entity entities.IEntity)
	RemoveEntity(targetId int64)
	ISubscriber
}

type ISubscriber interface {
	// 根据 ID 列表订阅
	SubscribeByIds(subscriberId string, entityIds []int64) []entities.IEntity
	// 根据类型订阅
	SubscribeByType(subscriberId string, entityTypes []string) []entities.IEntity
	// 订阅所有
	SubscribeAll(subscriberId string) []entities.IEntity
	// 取消订阅
	Unsubscribe(subscriberId string)
	// 获取已订阅的 Entities
	GetSubscribedEntities(subscriberId string) []entities.IEntity
	// 获取订阅策略类型
	GetSubscriberStrategyType(subscriberId string) (mme.SubscribeStrategyType, bool)
}

type ILoadResponse interface {
	IsSuccess() bool
	IsExists() bool
	GetError() error
	GetData() bson.Raw
}
