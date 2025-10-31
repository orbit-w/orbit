package persistence

import (
	"context"
	"time"

	"github.com/asynkron/protoactor-go/actor"
)

const (
	// PersistencePattern 持久化Actor的模式标识
	PersistencePattern = "persistence"
)

// PersistenceRequest 持久化请求消息
type PersistenceRequest[IDType any] struct {
	// Collection 集合名称
	Collection string
	// DocID 文档ID（MongoDB的_id字段）
	DocID IDType
	// Wrapper 实现了Wrapper接口的数据包装器
	Wrapper Wrapper
	// ResponseReceiver 响应接收者（可选），用于接收持久化结果
	ResponseReceiver *actor.PID
	// Timeout 超时时间（可选）
	Timeout time.Duration
	// Context 请求上下文（可选）
	Context context.Context
}

// PersistenceResponse 持久化响应消息
type PersistenceResponse struct {
	// Success 是否成功
	Success bool
	// Error 错误信息（如果失败）
	Error error
	// Collection 集合名称
	Collection string
	// DocumentID 文档ID
	DocumentID any
	// MatchedCount 匹配的文档数量（更新操作）
	MatchedCount int64
	// ModifiedCount 修改的文档数量（更新操作）
	ModifiedCount int64
}

// BatchPersistenceRequest 批量持久化请求消息
type BatchPersistenceRequest[IDType any] struct {
	// Requests 持久化请求列表
	Requests []*PersistenceRequest[IDType]
	// ResponseReceiver 响应接收者（可选）
	ResponseReceiver *actor.PID
}

// BatchPersistenceResponse 批量持久化响应消息
type BatchPersistenceResponse struct {
	// Results 每个请求的结果
	Results []*PersistenceResponse
	// SuccessCount 成功的数量
	SuccessCount int
	// FailCount 失败的数量
	FailCount int
}
