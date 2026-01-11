package servicezone

import (
	"gitee.com/orbit-w/orbit/internal/game/mme_agent"
	"gitee.com/orbit-w/orbit/pkg/proto/mme"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type IServiceZone interface {
	AddEntity(entity mme_agent.IEntity)
	SetEntity(entity mme_agent.IEntity)
	RemoveEntity(targetId int64)
	ISubscriber
}

type ISubscriber interface {
	// 根据 ID 列表订阅
	SubscribeByIds(subscriberId string, entityIds []int64) []mme_agent.IEntity
	// 根据类型订阅
	SubscribeByType(subscriberId string, entityTypes []string) []mme_agent.IEntity
	// 订阅所有
	SubscribeAll(subscriberId string) []mme_agent.IEntity
	// 取消订阅
	Unsubscribe(subscriberId string)
	// 获取已订阅的 Entities
	GetSubscribedEntities(subscriberId string) []mme_agent.IEntity
	// 获取订阅策略类型
	GetSubscriberStrategyType(subscriberId string) (mme.SubscribeStrategyType, bool)
}

type ILoadResponse interface {
	IsSuccess() bool
	IsExists() bool
	GetError() error
	GetData() bson.Raw
}
