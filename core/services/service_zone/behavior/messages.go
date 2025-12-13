package servicezone_behavior

import (
	mmeobj "gitee.com/orbit-w/orbit/app/mme"
	"gitee.com/orbit-w/orbit/app/proto/core"
	"gitee.com/orbit-w/orbit/app/proto/mme"
	"gitee.com/orbit-w/orbit/app/proto/pb"
	"gitee.com/orbit-w/orbit/core/network"
	"github.com/asynkron/protoactor-go/actor"
	"google.golang.org/protobuf/proto"
)

// ZoneActorMessages 定义 ServiceZone Actor 的所有消息类型

// AddEntityRequest 添加或更新 Entity 的请求
type AddEntityRequest struct {
	Entity mmeobj.IEntity
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
	Strategy         mme.SubscribeStrategyType
	ResponseReceiver *actor.PID // 可选，用于接收响应
}

// SubscribeResponse 订阅响应
type SubscribeResponse struct {
	Entities []mmeobj.IEntity
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
	Entities []mmeobj.IEntity
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

type ClientRequest struct {
	network.IClientRequest
	ZoneId string
}

type Response struct {
	Error error
	Msg   any
}

type IResponse interface {
	Response(data []byte, pid uint32) error
}

func ResponseOK(response IResponse) {
	rawData, err := proto.Marshal(&core.OK{})
	if err != nil {
		return
	}
	response.Response(rawData, pb.PID_Rsp_OK)
}

func ResponseError(response IResponse, reason string) {
	rawData, err := proto.Marshal(&core.Error{
		Reason: proto.String(reason),
	})
	if err != nil {
		return
	}
	response.Response(rawData, pb.PID_Rsp_Error)
}
