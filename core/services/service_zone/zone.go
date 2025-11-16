package servicezone

import "gitee.com/orbit-w/orbit/app/proto/core"

type IServiceZone interface {
	AddEntity(entity IEntity)
	SetEntity(entity IEntity)
	RemoveEntity(targetId int64)
	ISubscriber
}

type ISubscriber interface {
	// 根据 ID 列表订阅
	SubscribeByIds(subscriberId string, entityIds []int64) []IEntity
	// 根据类型订阅
	SubscribeByType(subscriberId string, entityTypes []string) []IEntity
	// 订阅所有
	SubscribeAll(subscriberId string) []IEntity
	// 取消订阅
	Unsubscribe(subscriberId string)
	// 获取已订阅的 Entities
	GetSubscribedEntities(subscriberId string) []IEntity
	// 获取订阅策略类型
	GetSubscriberStrategyType(subscriberId string) (core.SubscribeStrategyType, bool)
}
