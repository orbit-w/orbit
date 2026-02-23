package worker

import (
	"time"

	entityloader "gitee.com/orbit-w/orbit/core/system/gravitas/entity_loader"
	"gitee.com/orbit-w/orbit/internal/game/mme_agent"
)

// ProcessResult TryProcessAnchor 的三种结果（文档 Section 3.1）
type ProcessResult int

const (
	ResultSuccess  ProcessResult = iota // 全锁成功 + 全加载 + 执行完毕，锁已全部释放
	ResultLockBusy                      // 某个锁获取失败，已全部回滚，消息进入 StagingArea 队首
	ResultNeedIO                        // 全锁成功但有 Entity 未加载，已全部释放，消息进入 StagingArea 队首
)

// MessageHandler 业务逻辑执行函数。
//
// Worker 在获取所有锁并确保所有 Entity 已加载后调用。
// entities: EntityID → IEntity 映射，包含消息涉及的所有 Entity。
type MessageHandler func(entities map[int64]mme_agent.IEntity) error

// WorkerMessage Worker 处理的消息单元。
//
// 由 WorkerPool.Dispatch 创建，EntityRefs 已按 EntityID 升序排列（INV-5），
// AnchorID = min(EntityRefs.EntityID)。
type WorkerMessage struct {
	EntityRefs []entityloader.EntityRef
	AnchorID   int64
	Handler    MessageHandler
	RetryCount int
	CreateTime time.Time
}

// tickEvent 内部定时 Tick，由 time.AfterFunc 投递到 Worker mailbox。
type tickEvent struct{}
