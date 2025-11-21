package persistence

import (
	"context"
	"time"

	"github.com/asynkron/protoactor-go/actor"
	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	// PersistencePattern 持久化Actor的模式标识
	PersistencePattern = "persistence"
)

// LoadRequest 加载请求消息
type LoadRequest struct {
	// Database 数据库名称
	Database string
	// Collection 集合名称
	Collection string
	// DocID 文档ID
	DocID any
	// Timeout 超时时间（可选）
	Timeout time.Duration
	// Context 请求上下文（可选）
	Context context.Context
}

// PersistenceRequest 持久化请求消息
type PersistenceRequest struct {
	// Database 数据库名称
	Database string
	// Collection 集合名称
	Collection string
	// DocID 文档ID（MongoDB的_id字段）
	DocID any
	// Doc 文档
	Doc bson.M
	// ResponseReceiver 响应接收者（可选），用于接收持久化结果
	ResponseReceiver *actor.PID
	// Timeout 超时时间（可选）
	Timeout time.Duration
	// Context 请求上下文（可选）
	Context context.Context
}

// LoadResponse 加载响应消息
type LoadResponse struct {
	// Success 是否成功
	Success bool
	// Error 错误信息（如果失败）
	Error error
	// Collection 集合名称
	Collection string
	// DocumentID 文档ID
	DocumentID any
	// Data 加载的文档数据（如果成功）
	Data bson.M
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
	Requests []*PersistenceRequest
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
