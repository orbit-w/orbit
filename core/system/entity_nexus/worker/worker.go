package worker

import (
	"time"

	"gitee.com/orbit-w/meteor/bases/misc/utils"
	mlog "gitee.com/orbit-w/meteor/modules/mlog"
	entityloader "gitee.com/orbit-w/orbit/core/system/entity_nexus/entity_loader"
	entitymgr "gitee.com/orbit-w/orbit/core/system/entity_nexus/entity_mgr"
	"gitee.com/orbit-w/orbit/internal/game/mme_agent"
	"gitee.com/orbit-w/orbit/lib/module/logger"
	"github.com/asynkron/protoactor-go/actor"
	"go.uber.org/zap"
)

const (
	DefaultRetryQuotaPerTick = 10
	DefaultTickInterval      = 50 * time.Millisecond
	DefaultMaxRetryCount     = 3
)

// WorkerConfig Worker 配置。
type WorkerConfig struct {
	ID            int
	RetryQuota    int           // RetryProcess 每 Tick 的配额（默认 10）
	TickInterval  time.Duration // Tick 间隔（默认 50ms）
	MaxRetryCount int           // IO 加载失败最大重试次数（默认 3）
}

// Worker 消息调度 Actor，实现文档 Section 12 描述的完整消息调度模型。
//
// 核心职责：
//   - 按 AnchorID 管理消息暂存队列（StagingArea）
//   - TryProcessAnchor 三阶段执行：加锁 → 加载检查 → 业务执行
//   - RetryProcess 配额制重试（防止堆积阻塞新消息）
//   - 处理 EntityLoader 的异步加载完成/失败通知
//   - CooldownSet 防止活锁（Tick 级冷却）
//
// Worker 内部状态全部由 Actor mailbox 串行驱动，无需加锁。
// 跨 Worker 的并发安全由 EntityManager 的 per-entity TryLock 保证。
type Worker struct {
	id     int
	logger *mlog.Logger

	em        *entitymgr.EntityManager
	loaderPID *actor.PID

	// 核心调度数据结构
	stagingArea    *StagingArea
	retryList      *RetryList
	loadingAnchors map[int64]bool // AnchorID → true 表示正在异步加载

	// 配置
	retryQuota    int
	maxRetryCount int
	tickInterval  time.Duration

	// Actor 运行时
	self   *actor.PID
	system *actor.ActorSystem
}

// NewWorker 创建 Worker 实例。由 WorkerPool 内部调用。
func NewWorker(cfg WorkerConfig, em *entitymgr.EntityManager, loaderPID *actor.PID) *Worker {
	retryQuota := cfg.RetryQuota
	if retryQuota <= 0 {
		retryQuota = DefaultRetryQuotaPerTick
	}
	tickInterval := cfg.TickInterval
	if tickInterval <= 0 {
		tickInterval = DefaultTickInterval
	}
	maxRetry := cfg.MaxRetryCount
	if maxRetry <= 0 {
		maxRetry = DefaultMaxRetryCount
	}

	return &Worker{
		id:             cfg.ID,
		logger:         logger.GetLogger(),
		em:             em,
		loaderPID:      loaderPID,
		stagingArea:    newStagingArea(),
		retryList:      newRetryList(),
		loadingAnchors: make(map[int64]bool),
		retryQuota:     retryQuota,
		maxRetryCount:  maxRetry,
		tickInterval:   tickInterval,
	}
}

// ---------------------------------------------------------------------------
// Actor 接口
// ---------------------------------------------------------------------------

// Receive 实现 actor.Actor 接口（文档 Section 12.1）。
func (w *Worker) Receive(ctx actor.Context) {
	defer utils.RecoverPanic()

	switch msg := ctx.Message().(type) {
	case *actor.Started:
		w.self = ctx.Self()
		w.system = ctx.ActorSystem()
		w.scheduleTick()
		w.logger.Info("Worker started", zap.Int("WorkerID", w.id))
	case *actor.Stopping:
		w.logger.Info("Worker stopping", zap.Int("WorkerID", w.id))
	case *actor.Stopped:
		w.logger.Info("Worker stopped", zap.Int("WorkerID", w.id))
	case *WorkerMessage:
		w.processMessage(ctx, msg)
	case *entityloader.EntityLoadBatchComplete:
		w.handleLoadComplete(ctx, msg)
	case *tickEvent:
		w.onTick(ctx)
	}
}

// ---------------------------------------------------------------------------
// 消息入口决策树（文档 Section 2）
// ---------------------------------------------------------------------------

// processMessage 新消息到达时的决策逻辑。
//
// 决策树：
//
//	StagingArea[AnchorID] 存在？
//	  ├── 存在 → PushTail（FIFO 追加，等待统一调度）
//	  └── 不存在 → TryProcessAnchor
//	        ├── Success → 完成
//	        ├── LockBusy → RetryList.Add（msg 已入队首）
//	        └── NeedIO → 等待 AsyncComplete（msg 已入队首，LoadingAnchors 已设置）
func (w *Worker) processMessage(ctx actor.Context, msg *WorkerMessage) {
	anchorID := msg.AnchorID

	if w.stagingArea.Exists(anchorID) {
		w.stagingArea.PushTail(anchorID, msg)
		return
	}

	result := w.tryProcessAnchor(ctx, msg)
	switch result {
	case ResultSuccess:
		// 完成
	case ResultLockBusy:
		w.retryList.Add(anchorID)
	case ResultNeedIO:
		// loadEntities 已处理
	}
}

// ---------------------------------------------------------------------------
// TryProcessAnchor 三阶段执行（文档 Section 3）
// ---------------------------------------------------------------------------

// tryProcessAnchor 完整的三阶段处理流程。
//
//	Phase 1: 按 EntityID 升序逐个 TryLock（INV-5：避免死锁）
//	Phase 2: 在锁保护下检查 Entity 加载状态（消除 TOCTOU 竞态）
//	Phase 3: 执行业务逻辑
//
// 使用 defer unlockAll 保证 panic 安全：Handler panic 时锁仍会被正确释放。
func (w *Worker) tryProcessAnchor(ctx actor.Context, msg *WorkerMessage) ProcessResult {
	refs := msg.EntityRefs
	if len(refs) == 0 {
		return ResultSuccess
	}

	// Phase 1: 加锁序列
	lockedIDs := make([]int64, 0, len(refs))
	unlocked := false

	unlockAll := func() {
		if unlocked {
			return
		}
		unlocked = true
		for i := len(lockedIDs) - 1; i >= 0; i-- {
			w.em.Unlock(lockedIDs[i])
		}
	}
	defer unlockAll()

	for _, ref := range refs {
		if w.em.TryLock(ref.EntityID) {
			lockedIDs = append(lockedIDs, ref.EntityID)
		} else {
			unlockAll()
			w.stagingArea.PushHead(msg.AnchorID, msg)
			return ResultLockBusy
		}
	}

	// Phase 2: Entity 加载检查（在锁保护下）
	var needLoad []entityloader.EntityRef
	for _, ref := range refs {
		if !w.em.Has(ref.EntityID) {
			needLoad = append(needLoad, ref)
		}
	}

	if len(needLoad) > 0 {
		unlockAll()
		w.loadEntities(ctx, msg, needLoad)
		return ResultNeedIO
	}

	// Phase 3: 执行业务逻辑
	entities := make([]mme_agent.IEntity, 0, len(refs))
	for _, ref := range refs {
		if entity, ok := w.em.Get(ref.EntityID); ok {
			entities = append(entities, entity)
		} else {
			w.logger.Error("Entity disappeared under lock",
				zap.Int64("EntityID", ref.EntityID),
				zap.Int64("AnchorID", msg.AnchorID))
		}
	}

	if msg.Handler != nil {
		if err := msg.Handler(entities); err != nil {
			w.logger.Error("Message handler execution failed",
				zap.Int64("AnchorID", msg.AnchorID),
				zap.Error(err))
		} else {
			w.logger.Info("Message handler executed successfully",
				zap.Int("WorkerID", w.id),
				zap.Int64("AnchorID", msg.AnchorID),
				zap.Int("EntityCount", len(entities)))
		}
	}

	// defer unlockAll() 释放所有锁
	return ResultSuccess
}

// ---------------------------------------------------------------------------
// 异步加载（文档 Section 11.2 ~ 11.4）
// ---------------------------------------------------------------------------

// loadEntities 逻辑单元四：触发异步加载。
//
// 步骤：
//  1. 消息入 StagingArea 队首（确保不丢失）
//  2. 设置 LoadingAnchors 标记（防止 RetryProcess 无效遍历）
//  3. 向 EntityLoader 发送 EntityLoadBatchRequest
func (w *Worker) loadEntities(ctx actor.Context, msg *WorkerMessage, needLoad []entityloader.EntityRef) {
	anchorID := msg.AnchorID

	w.stagingArea.PushHead(anchorID, msg)
	w.loadingAnchors[anchorID] = true

	ctx.Send(w.loaderPID, &entityloader.EntityLoadBatchRequest{
		Refs:      needLoad,
		AnchorID:  anchorID,
		Requester: ctx.Self(),
	})
}

// handleLoadComplete 分发 EntityLoader 的加载完成通知。
func (w *Worker) handleLoadComplete(ctx actor.Context, msg *entityloader.EntityLoadBatchComplete) {
	if msg.Success {
		w.handleAsyncComplete(ctx, msg.AnchorID)
	} else {
		w.handleAsyncError(ctx, msg.AnchorID, msg.Errors)
	}
}

// handleAsyncComplete 逻辑单元五：异步加载成功处理（文档 Section 11.3）。
//
// 流程：
//  1. 验证并清除 LoadingAnchors 标记
//  2. 取出队首消息
//  3. TryProcessAnchor 重试执行
//  4. 根据结果维护 RetryList（INV-1 一致性）
func (w *Worker) handleAsyncComplete(ctx actor.Context, anchorID int64) {
	if !w.loadingAnchors[anchorID] {
		w.logger.Warn("Unexpected AsyncComplete for non-loading anchor",
			zap.Int64("AnchorID", anchorID))
		return
	}

	// 必须在处理消息前清除，否则嵌套 NeedIO 会检测到冲突
	delete(w.loadingAnchors, anchorID)

	if !w.stagingArea.Exists(anchorID) {
		w.logger.Warn("Empty staging area after load complete",
			zap.Int64("AnchorID", anchorID))
		return
	}

	msg, ok := w.stagingArea.PopHead(anchorID)
	if !ok {
		return
	}

	result := w.tryProcessAnchor(ctx, msg)

	switch result {
	case ResultSuccess:
		if w.stagingArea.Exists(anchorID) {
			w.retryList.Add(anchorID)
		}
	case ResultLockBusy:
		w.retryList.Add(anchorID)
	case ResultNeedIO:
		// loadEntities 已将消息入队首并设置 LoadingAnchors，无需额外处理
	}
}

// handleAsyncError 异步加载失败处理（文档 Section 11.4）。
//
// 策略：有限重试（RetryCount < MaxRetryCount），超限后丢弃并告警。
func (w *Worker) handleAsyncError(ctx actor.Context, anchorID int64, errors []error) {
	delete(w.loadingAnchors, anchorID)

	for _, err := range errors {
		w.logger.Error("Entity async load failed",
			zap.Int64("AnchorID", anchorID),
			zap.Error(err))
	}

	msg, ok := w.stagingArea.PopHead(anchorID)
	if !ok {
		w.logger.Error("Missing message after async load error",
			zap.Int64("AnchorID", anchorID))
		return
	}

	msg.RetryCount++
	if msg.RetryCount < w.maxRetryCount {
		w.stagingArea.PushTail(anchorID, msg)
	} else {
		w.logger.Error("Message exceeded max retry count, dropped (DLQ candidate)",
			zap.Int64("AnchorID", anchorID),
			zap.Int("RetryCount", msg.RetryCount))
		// TODO: 移入死信队列（DLQ）
	}

	if w.stagingArea.Exists(anchorID) {
		w.retryList.Add(anchorID)
	}
}

// ---------------------------------------------------------------------------
// Tick 调度与 RetryProcess（文档 Section 6 & 12）
// ---------------------------------------------------------------------------

// onTick 定时 Tick 处理入口。
func (w *Worker) onTick(ctx actor.Context) {
	w.retryProcess(ctx, w.retryQuota)
	w.scheduleTick()
}

// retryProcess 配额制重试处理（文档 Section 6 & 10.3）。
//
// 遍历 RetryList 快照，对每个可处理的 Anchor 取出队首消息执行 TryProcessAnchor。
// 跳过 IO 等待中和冷却中的 Anchor，仅在实际执行时消耗配额。
//
// CooldownSet 是 retryProcess 内部的局部 map，每次 Tick 调用时新建，
// 调用结束自动销毁，无需跨 Tick 维护。
func (w *Worker) retryProcess(ctx actor.Context, quota int) {
	if quota <= 0 || w.retryList.Len() == 0 {
		return
	}

	snapshot := w.retryList.Snapshot()
	cooldownThisTick := make(map[int64]struct{})
	remaining := quota

	for _, anchorID := range snapshot {
		if remaining <= 0 {
			break
		}

		if w.loadingAnchors[anchorID] {
			continue
		}

		if _, cooled := cooldownThisTick[anchorID]; cooled {
			continue
		}

		msg, ok := w.stagingArea.PopHead(anchorID)
		if !ok {
			w.retryList.Remove(anchorID)
			continue
		}

		result := w.tryProcessAnchor(ctx, msg)
		remaining--

		switch result {
		case ResultSuccess:
			if !w.stagingArea.Exists(anchorID) {
				w.retryList.Remove(anchorID)
			}
		case ResultLockBusy:
			// msg 已回滚到 StagingArea 队首，设置冷却防止活锁（文档 Section 5）
			cooldownThisTick[anchorID] = struct{}{}
		case ResultNeedIO:
			// 已转入 LoadingAnchors 状态，从 RetryList 移除（INV-1）
			w.retryList.Remove(anchorID)
		}
	}
}

// scheduleTick 通过 time.AfterFunc 向自身投递 tickEvent。
//
// 使用 Actor 消息而非 goroutine 直接访问状态，保证线程安全。
func (w *Worker) scheduleTick() {
	self := w.self
	system := w.system
	interval := w.tickInterval
	time.AfterFunc(interval, func() {
		system.Root.Send(self, &tickEvent{})
	})
}
