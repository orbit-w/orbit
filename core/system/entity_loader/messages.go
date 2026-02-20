package entityloader

import (
	"gitee.com/orbit-w/orbit/pkg/proto/mme"
	"github.com/asynkron/protoactor-go/actor"
)

// EntityRef 实体引用，标识一个待加载的 Entity。
type EntityRef struct {
	EntityID   int64
	EntityType mme.EntityType
}

// EntityLoadBatchRequest Worker → EntityLoader 的批量异步加载请求。
//
// Worker 在 TryProcessAnchor 检测到 NeedIO 后，释放所有锁、设置 LoadingAnchors 标记，
// 然后发送此消息给 EntityLoader。
type EntityLoadBatchRequest struct {
	// Refs 待加载的实体引用列表
	Refs []EntityRef
	// AnchorID Worker 调度锚点（最小 EntityID），用于回传给 Worker 关联 StagingArea
	AnchorID int64
	// Requester Worker 的 PID，EntityLoader 完成后向此 PID 发送 EntityLoadBatchComplete
	Requester *actor.PID
}

// EntityLoadBatchComplete EntityLoader → Worker 的批量加载完成通知。
//
// Worker 收到后根据 Success 字段决定：
//   - true  → HandleEventAsyncComplete（正常流程）
//   - false → HandleEventAsyncError（重试/DLQ）
type EntityLoadBatchComplete struct {
	// AnchorID 对应请求的调度锚点
	AnchorID int64
	// Success 批次是否全部加载成功
	Success bool
	// Errors 加载过程中的错误列表（仅 Success=false 时非空）
	Errors []error
}
