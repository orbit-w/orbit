package entityloader

import (
	"fmt"
	"strconv"
	"time"

	"gitee.com/orbit-w/meteor/bases/misc/utils"
	mlog "gitee.com/orbit-w/meteor/modules/mlog"
	mmeobj "gitee.com/orbit-w/orbit/internal/game/mme"
	"gitee.com/orbit-w/orbit/internal/game/mme_agent"
	entitymgr "gitee.com/orbit-w/orbit/core/system/entity_mgr"
	"gitee.com/orbit-w/orbit/lib/module/logger"
	"gitee.com/orbit-w/orbit/lib/module/persistence"
	"gitee.com/orbit-w/orbit/pkg/proto/mme"
	"github.com/asynkron/protoactor-go/actor"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"
)

const (
	// DefaultLoadTimeout 是单个批次从发出 AsyncLoad 到收到 LoadResponse 的最大等待时间。
	// 超时后 EntityLoader 主动发送 Success=false 通知 Worker，防止 StagingArea 永久阻塞。
	DefaultLoadTimeout = 10 * time.Second

	// timeoutCheckInterval 是超时检查的轮询间隔，设为 loadTimeout 的一半，
	// 保证最坏情况下超时延迟不超过 1.5× loadTimeout。
	timeoutCheckInterval = DefaultLoadTimeout / 2
)

// checkTimeoutsMsg 触发一次超时扫描，由 scheduleTimeoutCheck 通过 AfterFunc 投递。
type checkTimeoutsMsg struct{}

// entityLoaderActor 异步实体加载 Actor。
//
// 接收 Worker 的 EntityLoadBatchRequest，逐个调用 persistence.AsyncLoad，
// 以自身 PID 为 ResponseReceiver 收集 *persistence.LoadResponse，
// 创建 Entity 实例并注册到 EntityManager，批次完成后通知 Worker。
//
// 单线程模型（Actor mailbox 保证），内部状态无需加锁。
type entityLoaderActor struct {
	em          *entitymgr.EntityManager
	dbResolver  DatabaseResolver
	logger      *mlog.Logger
	loadTimeout time.Duration

	// self / system 在 actor.Started 时由 ctx 保存，供 time.AfterFunc 回调使用。
	self   *actor.PID
	system *actor.ActorSystem

	// entityID → 加载追踪（去重：多个 batch 请求同一 entity 只发起一次 DB 查询）
	pendingEntities map[int64]*pendingEntity
	// batchID → 批次追踪
	pendingBatches map[string]*pendingBatch
	// 全局递增批次计数器（用于生成 batchID）
	batchCounter uint64
}

// pendingEntity 单个 entity 的加载追踪。
type pendingEntity struct {
	entityID   int64
	entityType mme.EntityType
	batchIDs   []string // 等待此 entity 加载完成的所有批次
}

// pendingBatch 一次批量加载请求的追踪。
type pendingBatch struct {
	batchID   string
	anchorID  int64
	requester *actor.PID
	remaining int // 还有多少 entity 未收到响应
	errors    []error
	startTime time.Time // 批次创建时间，用于超时检测
}

func newEntityLoaderActor(em *entitymgr.EntityManager, dbResolver DatabaseResolver, loadTimeout time.Duration) *entityLoaderActor {
	if loadTimeout <= 0 {
		loadTimeout = DefaultLoadTimeout
	}
	return &entityLoaderActor{
		em:              em,
		dbResolver:      dbResolver,
		logger:          logger.GetLogger(),
		loadTimeout:     loadTimeout,
		pendingEntities: make(map[int64]*pendingEntity),
		pendingBatches:  make(map[string]*pendingBatch),
	}
}

func (a *entityLoaderActor) Receive(ctx actor.Context) {
	defer utils.RecoverPanic()

	switch msg := ctx.Message().(type) {
	case *actor.Started:
		a.self = ctx.Self()
		a.system = ctx.ActorSystem()
		a.logger.Info("EntityLoaderActor started", zap.String("ActorID", ctx.Self().Id))
		a.scheduleTimeoutCheck()
	case *actor.Stopping:
		a.logger.Info("EntityLoaderActor stopping", zap.String("ActorID", ctx.Self().Id))
	case *actor.Stopped:
		a.logger.Info("EntityLoaderActor stopped", zap.String("ActorID", ctx.Self().Id))
	case *EntityLoadBatchRequest:
		a.handleLoadBatch(ctx, msg)
	case *persistence.LoadResponse:
		a.handleLoadResponse(ctx, msg)
	case *checkTimeoutsMsg:
		a.handleCheckTimeouts(ctx)
	default:
		a.logger.Error("EntityLoaderActor received unknown message", zap.Any("Message", msg))
	}
}

// handleLoadBatch 处理来自 Worker 的批量加载请求。
//
// 对每个 EntityRef：
//  1. 跳过 EntityManager 中已存在的 entity
//  2. 去重：如果 entity 已在加载中，仅追加 batchID 到 pendingEntity.batchIDs
//  3. 否则解析 database + collection，发起 persistence.AsyncLoad
//
// 如果所有 entity 均已加载，立即回复 EntityLoadBatchComplete。
func (a *entityLoaderActor) handleLoadBatch(ctx actor.Context, req *EntityLoadBatchRequest) {
	batchID := a.nextBatchID()
	batch := &pendingBatch{
		batchID:   batchID,
		anchorID:  req.AnchorID,
		requester: req.Requester,
		startTime: time.Now(),
	}

	for _, ref := range req.Refs {
		if a.em.Has(ref.EntityID) {
			continue
		}

		if pending, ok := a.pendingEntities[ref.EntityID]; ok {
			pending.batchIDs = append(pending.batchIDs, batchID)
			batch.remaining++
			continue
		}

		wrapperFactory := mmeobj.GetEntityWrapperFactory(ref.EntityType)
		if wrapperFactory == nil {
			batch.errors = append(batch.errors, fmt.Errorf(
				"entity wrapper factory not found for type %d (entityID=%d)", ref.EntityType, ref.EntityID))
			continue
		}

		wrapper := wrapperFactory()
		database := a.dbResolver(ref.EntityType)
		collection := wrapper.Collection()

		a.pendingEntities[ref.EntityID] = &pendingEntity{
			entityID:   ref.EntityID,
			entityType: ref.EntityType,
			batchIDs:   []string{batchID},
		}
		batch.remaining++

		if err := persistence.AsyncLoad(database, collection, ref.EntityID, ctx.Self()); err != nil {
			a.logger.Error("Failed to initiate async load",
				zap.Int64("EntityID", ref.EntityID),
				zap.Error(err))
			batch.errors = append(batch.errors, err)
			batch.remaining--
			delete(a.pendingEntities, ref.EntityID)
		}
	}

	a.pendingBatches[batchID] = batch

	if batch.remaining == 0 {
		a.completeBatch(ctx, batch)
	}
}

// handleLoadResponse 处理来自 PersistenceActor 的加载响应。
//
// 根据 DocumentID 关联到 pendingEntity，创建 Entity 实例并注册到 EntityManager，
// 然后递减所有等待此 entity 的 batch 的 remaining 计数。
// 当某个 batch 的 remaining 归零时，向 Worker 发送完成通知。
func (a *entityLoaderActor) handleLoadResponse(ctx actor.Context, resp *persistence.LoadResponse) {
	entityID, ok := toInt64(resp.DocumentID)
	if !ok {
		a.logger.Error("LoadResponse has unexpected DocumentID type",
			zap.Any("DocumentID", resp.DocumentID))
		return
	}

	pending, exists := a.pendingEntities[entityID]
	if !exists {
		a.logger.Warn("Received LoadResponse for unknown entity",
			zap.Int64("EntityID", entityID))
		return
	}

	var loadErr error
	if resp.GetError() != nil {
		loadErr = resp.GetError()
		a.logger.Error("Entity load failed",
			zap.Int64("EntityID", entityID),
			zap.Error(loadErr))
	} else {
		loadErr = a.createAndRegisterEntity(entityID, pending.entityType, resp)
	}

	for _, batchID := range pending.batchIDs {
		batch, ok := a.pendingBatches[batchID]
		if !ok {
			// 批次已超时完成，跳过（timed-out batch 已从 pendingBatches 删除）
			continue
		}
		if loadErr != nil {
			batch.errors = append(batch.errors, loadErr)
		}
		batch.remaining--
		if batch.remaining == 0 {
			a.completeBatch(ctx, batch)
		}
	}

	delete(a.pendingEntities, entityID)
}

// handleCheckTimeouts 扫描所有待处理批次，对超过 loadTimeout 的批次
// 主动发送 Success=false 通知，防止 StagingArea 永久阻塞（对应文档 Section 15：加载超时处理）。
//
// 超时的批次仍留在 pendingEntities 中等待 LoadResponse（避免重复 DB 查询）；
// 当 LoadResponse 最终到达时，handleLoadResponse 的 `!ok` 检查会安全跳过已超时的批次。
func (a *entityLoaderActor) handleCheckTimeouts(ctx actor.Context) {
	now := time.Now()

	var timedOut []*pendingBatch
	for _, batch := range a.pendingBatches {
		if now.Sub(batch.startTime) > a.loadTimeout {
			timedOut = append(timedOut, batch)
		}
	}

	for _, batch := range timedOut {
		elapsed := now.Sub(batch.startTime)
		a.logger.Warn("Entity load batch timed out",
			zap.String("BatchID", batch.batchID),
			zap.Int64("AnchorID", batch.anchorID),
			zap.Int("Remaining", batch.remaining),
			zap.Duration("Elapsed", elapsed),
			zap.Duration("Timeout", a.loadTimeout))
		batch.errors = append(batch.errors, fmt.Errorf(
			"load timed out after %s (anchorID=%d, batchID=%s)", elapsed.Round(time.Millisecond), batch.anchorID, batch.batchID))
		a.completeBatch(ctx, batch)
	}

	a.scheduleTimeoutCheck()
}

// scheduleTimeoutCheck 使用 time.AfterFunc 在 timeoutCheckInterval 后
// 向自身投递一条 checkTimeoutsMsg，驱动周期性超时扫描。
//
// 使用 time.AfterFunc + Actor 消息（而非 goroutine 直接访问状态）保证线程安全：
// 消息进入 mailbox 后在 Actor 单线程上下文中执行，无需额外同步。
func (a *entityLoaderActor) scheduleTimeoutCheck() {
	self := a.self
	system := a.system
	time.AfterFunc(timeoutCheckInterval, func() {
		system.Root.Send(self, &checkTimeoutsMsg{})
	})
}

// createAndRegisterEntity 根据 LoadResponse 创建 Entity 实例并注册到 EntityManager。
func (a *entityLoaderActor) createAndRegisterEntity(
	entityID int64, entityType mme.EntityType, resp *persistence.LoadResponse,
) error {
	entityFactory := mme_agent.GetEntityFactory(entityType)
	if entityFactory == nil {
		return fmt.Errorf("entity factory not found for type %d", entityType)
	}

	entity := entityFactory()

	var (
		raw   bson.Raw
		isNew bool
	)

	if !resp.IsExists() {
		isNew = true
		data := bson.M{"_id": entityID}
		var err error
		raw, err = bson.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal default entity data: %w", err)
		}
	} else {
		raw = resp.GetData()
	}

	if err := entity.OnLoad(raw, isNew); err != nil {
		return fmt.Errorf("entity OnLoad failed for %d: %w", entityID, err)
	}

	a.em.Add(entity)

	a.logger.Debug("Entity loaded and registered",
		zap.Int64("EntityID", entityID),
		zap.Bool("IsNew", isNew))

	return nil
}

// completeBatch 完成一个批次，向 Worker 发送通知并清理追踪状态。
func (a *entityLoaderActor) completeBatch(ctx actor.Context, batch *pendingBatch) {
	ctx.Send(batch.requester, &EntityLoadBatchComplete{
		AnchorID: batch.anchorID,
		Success:  len(batch.errors) == 0,
		Errors:   batch.errors,
	})
	delete(a.pendingBatches, batch.batchID)
}

// nextBatchID 生成递增的批次 ID。
func (a *entityLoaderActor) nextBatchID() string {
	a.batchCounter++
	return strconv.FormatUint(a.batchCounter, 10)
}

// toInt64 将 any 类型的 DocumentID 转换为 int64。
func toInt64(v any) (int64, bool) {
	switch id := v.(type) {
	case int64:
		return id, true
	case int:
		return int64(id), true
	case int32:
		return int64(id), true
	case float64:
		return int64(id), true
	default:
		return 0, false
	}
}
