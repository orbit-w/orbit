package persistence

import (
	"context"
	"time"

	"gitee.com/orbit-w/meteor/bases/misc/utils"
	mongodbdriver "gitee.com/orbit-w/meteor/modules/database/no_sql/mongodb_driver"
	mlog "gitee.com/orbit-w/meteor/modules/mlog"
	"gitee.com/orbit-w/orbit/lib/module/db/mgo_builder"
	"gitee.com/orbit-w/orbit/lib/module/logger"
	"github.com/asynkron/protoactor-go/actor"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

// UpdateResult 定义MongoDB更新结果的接口
// 用于从VirtualMongoClient.UpdateOne的返回值中提取计数信息
type UpdateResult interface {
	GetMatchedCount() int64
	GetModifiedCount() int64
}

// PersistenceActor 持久化Actor，负责处理持久化请求
type PersistenceActor struct {
	db     *mongodbdriver.VirtualMongoClient
	logger *mlog.Logger
}

// NewPersistenceActor 创建新的持久化Actor
func NewPersistenceActor(db *mongodbdriver.VirtualMongoClient) *PersistenceActor {
	return &PersistenceActor{
		db:     db,
		logger: logger.GetLogger(),
	}
}

// Receive 处理接收到的消息
func (p *PersistenceActor) Receive(ctx actor.Context) {
	defer utils.RecoverPanic()

	switch msg := ctx.Message().(type) {
	case *actor.Started:
		p.logger.Info("PersistenceActor started", zap.String("ActorID", ctx.Self().Id))

	case *PersistenceRequest[any]:
		p.handlePersistenceRequest(ctx, msg)

	case *BatchPersistenceRequest[any]:
		p.handleBatchPersistenceRequest(ctx, msg)

	default:
		p.logger.Error("PersistenceActor received unknown message", zap.Any("Message", msg))
	}
}

// handlePersistenceRequest 处理单个持久化请求
func (p *PersistenceActor) handlePersistenceRequest(ctx actor.Context, req *PersistenceRequest[any]) {
	response := p.persist(req)

	// 如果有响应接收者，发送响应
	if req.ResponseReceiver != nil {
		ctx.Send(req.ResponseReceiver, response)
	} else {
		// 否则直接响应给发送者
		ctx.Respond(response)
	}
}

// handleBatchPersistenceRequest 处理批量持久化请求
func (p *PersistenceActor) handleBatchPersistenceRequest(ctx actor.Context, req *BatchPersistenceRequest[any]) {
	results := make([]*PersistenceResponse, 0, len(req.Requests))
	successCount := 0
	failCount := 0

	for _, persistenceReq := range req.Requests {
		response := p.persist(persistenceReq)
		results = append(results, response)

		if response.Success {
			successCount++
		} else {
			failCount++
		}
	}

	batchResponse := &BatchPersistenceResponse{
		Results:      results,
		SuccessCount: successCount,
		FailCount:    failCount,
	}

	// 如果有响应接收者，发送响应
	if req.ResponseReceiver != nil {
		ctx.Send(req.ResponseReceiver, batchResponse)
	} else {
		// 否则直接响应给发送者
		ctx.Respond(batchResponse)
	}
}

// persist 执行实际的持久化操作
func (p *PersistenceActor) persist(req *PersistenceRequest[any]) *PersistenceResponse {
	if req == nil {
		return &PersistenceResponse{
			Success:    false,
			Error:      ErrInvalidRequest,
			Collection: "",
			DocumentID: nil,
		}
	}

	if req.Wrapper == nil {
		return &PersistenceResponse{
			Success:    false,
			Error:      ErrWrapperIsNil,
			Collection: req.Collection,
			DocumentID: req.DocID,
		}
	}

	if req.Collection == "" {
		return &PersistenceResponse{
			Success:    false,
			Error:      ErrEmptyCollection,
			Collection: "",
			DocumentID: req.DocID,
		}
	}

	if req.DocID == nil {
		return &PersistenceResponse{
			Success:    false,
			Error:      ErrEmptyDocumentID,
			Collection: req.Collection,
			DocumentID: nil,
		}
	}

	// 构建更新操作
	updateDoc := mgo_builder.WithBuilderResult(func(builder *mgo_builder.MongoUpdateBuilder) map[string]any {
		// 创建根路径
		path := &mgo_builder.NestedPath{}
		// 调用Wrapper的BuildMongoUpdate方法
		req.Wrapper.BuildMongoUpdate(builder, path)
		return builder.Build()
	})

	// 如果没有更新操作，返回成功（无需更新）
	if updateDoc == nil || len(updateDoc) == 0 {
		p.logger.Debug("No update operations, skipping persistence",
			zap.String("Collection", req.Collection),
			zap.Any("DocumentID", req.DocID))
		return &PersistenceResponse{
			Success:       true,
			Collection:    req.Collection,
			DocumentID:    req.DocID,
			MatchedCount:  0,
			ModifiedCount: 0,
		}
	}

	// 构建查询条件（使用_id）
	filter := bson.M{"_id": req.DocID}

	// 使用请求的上下文或默认上下文
	persistCtx := req.Context
	if persistCtx == nil {
		var cancel context.CancelFunc
		timeout := req.Timeout
		if timeout == 0 {
			timeout = 30 * time.Second
		}
		persistCtx, cancel = context.WithTimeout(context.Background(), timeout)
		defer cancel()
	}

	// 执行MongoDB更新操作
	// 注意：这里假设VirtualMongoClient有UpdateOne方法
	// 如果接口不同，需要根据实际接口调整
	result, err := p.db.Client().Database(req.Collection).Collection(req.Collection).UpdateOne(persistCtx, filter, updateDoc, options.UpdateOne())
	if err != nil {
		p.logger.Error("Failed to persist data",
			zap.String("Collection", req.Collection),
			zap.Any("DocumentID", req.DocID),
			zap.Error(err))
		return &PersistenceResponse{
			Success:    false,
			Error:      err,
			Collection: req.Collection,
			DocumentID: req.DocID,
		}
	}

	matchedCount := int64(0)
	modifiedCount := int64(0)
	// 根据实际的UpdateResult类型提取计数
	matchedCount = result.MatchedCount
	modifiedCount = result.ModifiedCount

	p.logger.Debug("Data persisted successfully",
		zap.String("Collection", req.Collection),
		zap.Any("DocumentID", req.DocID),
		zap.Int64("MatchedCount", matchedCount),
		zap.Int64("ModifiedCount", modifiedCount))

	return &PersistenceResponse{
		Success:       true,
		Collection:    req.Collection,
		DocumentID:    req.DocID,
		MatchedCount:  matchedCount,
		ModifiedCount: modifiedCount,
	}
}

// NewNestedPath 创建新的嵌套路径
func NewNestedPath() *mgo_builder.NestedPath {
	return &mgo_builder.NestedPath{}
}

// NewNestedPathWithField 创建带字段名的嵌套路径
func NewNestedPathWithField(fieldName string) *mgo_builder.NestedPath {
	path := NewNestedPath()
	return path.Field(fieldName)
}
