package servicezone

import (
	"gitee.com/orbit-w/orbit/app/proto/mme"
	"gitee.com/orbit-w/orbit/core/network"
	"github.com/asynkron/protoactor-go/actor"
)

// ZoneActorMessages 定义 ServiceZone Actor 的所有消息类型

// AddEntityRequest 添加或更新 Entity 的请求
type AddEntityRequest struct {
	Entity IEntity
}

// AddEntityResponse 添加或更新 Entity 的响应
type AddEntityResponse struct {
	Success bool
}

// RemoveEntityRequest 移除 Entity 的请求
type RemoveEntityRequest struct {
	EntityID int64
}

// RemoveEntityResponse 移除 Entity 的响应
type RemoveEntityResponse struct {
	Success bool
}

// SubscribeRequest 订阅请求
type SubscribeRequest struct {
	SubscriberID     string
	Strategy         ISubscribeStrategy
	ResponseReceiver *actor.PID // 可选，用于接收响应
}

// SubscribeResponse 订阅响应
type SubscribeResponse struct {
	Entities []IEntity
}

// SubscribeByIdsRequest 按 ID 列表订阅请求
type SubscribeByIdsRequest struct {
	SubscriberID     string
	EntityIDs        []int64
	ResponseReceiver *actor.PID
}

// SubscribeByTypeRequest 按类型订阅请求
type SubscribeByTypeRequest struct {
	SubscriberID     string
	EntityTypes      []string
	ResponseReceiver *actor.PID
}

// SubscribeAllRequest 订阅所有请求
type SubscribeAllRequest struct {
	SubscriberID     string
	ResponseReceiver *actor.PID
}

// UnsubscribeRequest 取消订阅请求
type UnsubscribeRequest struct {
	SubscriberID string
}

// UnsubscribeResponse 取消订阅响应
type UnsubscribeResponse struct {
	Success bool
}

// GetSubscribedEntitiesRequest 获取已订阅的 Entities 请求
type GetSubscribedEntitiesRequest struct {
	SubscriberID     string
	ResponseReceiver *actor.PID
}

// GetSubscribedEntitiesResponse 获取已订阅的 Entities 响应
type GetSubscribedEntitiesResponse struct {
	Entities []IEntity
}

// GetSubscriberStrategyTypeResponse 获取订阅策略类型响应
type GetSubscriberStrategyTypeResponse struct {
	StrategyType mme.SubscribeStrategyType
	Exists       bool
}

// SubscribeByStrategyTypeRequest 根据策略类型订阅请求
type SubscribeByStrategyTypeRequest struct {
	SubscriberID     string
	StrategyType     mme.SubscribeStrategyType
	Params           any
	ResponseReceiver *actor.PID
}

type Request struct {
	ZoneId  string
	Session *network.Session
	Bytes   []byte
	Pid     uint32
}

type Response struct {
	Error error
	Msg   any
}
