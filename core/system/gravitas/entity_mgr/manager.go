package entitymgr

import (
	"sync"
	"sync/atomic"

	"gitee.com/orbit-w/meteor/bases/sync/reentrantlock"
	"gitee.com/orbit-w/orbit/internal/game/mme_agent"
	"gitee.com/orbit-w/orbit/lib/utils"
	"gitee.com/orbit-w/orbit/pkg/proto/mme"
	cmap "github.com/orcaman/concurrent-map/v2"
)

// ---------------------------------------------------------------------------
// EntityManager — 已加载 Entity 的线程安全管理器
// ---------------------------------------------------------------------------
//
// 设计原则（方案 B：单一存储 + 路由索引）：
//   - Entity 只有一份存储在 EntityGroup.entities 中（唯一数据源）
//   - typeIndex 仅作为 entityId → entityType 的轻量路由表
//   - locks 独立于 Entity 生命周期，支持先锁后加载
//   - Add/Remove 在 group.mu.Lock 下同步写入 typeIndex 和 group.entities
//   - typeIndex 只是路由提示，最终一致性由 group.mu 保证
//
// 详见 docs/EntityManager设计方案.md
//
// 注意：EntityId唯一不能复用！！！
type EntityManager struct {
	// typeIndex: entityId → entityType 路由表。
	// 用于 Get(id) 将请求路由到正确的 EntityGroup。
	// 不存储 Entity 本身，仅作路由提示。
	typeIndex cmap.ConcurrentMap[int64, mme.EntityType]

	// groups: entityType → *EntityGroup（唯一 Entity 存储）。
	// EntityType 首次注册时创建，永不删除。
	groups sync.Map

	// locks: entityId → *reentrantlock.ReentrantLock（独立锁表）。
	// 与 Entity 存储解耦，支持先锁后加载（Worker 先 TryLock 再 AsyncLoad）。
	// 使用可重入锁，同一 goroutine 内可重复加锁。
	locks cmap.ConcurrentMap[int64, *reentrantlock.ReentrantLock]

	// count: 全局 Entity 计数（原子操作，无锁竞争）。
	count atomic.Int64
}

// int64ShardingFunc 将 int64 key 映射到 cmap shard 槽位。
// 复用 MurmurHash3 fmix64 finalizer，兼顾高低位信息，降低碰撞率。
func int64ShardingFunc(key int64) uint32 {
	return uint32(utils.Fmix64(uint64(key)))
}

// NewEntityManager 创建 EntityManager 实例。
func NewEntityManager() *EntityManager {
	return &EntityManager{
		typeIndex: cmap.NewWithCustomShardingFunction[int64, mme.EntityType](int64ShardingFunc),
		locks:     cmap.NewWithCustomShardingFunction[int64, *reentrantlock.ReentrantLock](int64ShardingFunc),
	}
}

func (m *EntityManager) Get(entityId int64) (mme_agent.IEntity, bool) {
	entityType, ok := m.typeIndex.Get(entityId)
	if !ok {
		return nil, false
	}
	group := m.getGroup(entityType)
	if group == nil {
		return nil, false
	}
	return group.Get(entityId)
}

// Has 检查 Entity 是否已加载。
// 走完整 Get 路径，与 Get 一致性保证相同，避免 typeIndex 单独查询的假阳性。
func (m *EntityManager) Has(entityId int64) bool {
	_, ok := m.Get(entityId)
	return ok
}

// HasFast 变种，不需要与 Get 保证一致性保证，直接查 typeIndex 即可。
// 用于 Has 的快速路径，避免重复 Get 操作。
func (m *EntityManager) HasFast(entityId int64) bool {
	_, ok := m.typeIndex.Get(entityId)
	return ok
}

// ---------------------------------------------------------------------------
// 需求 2：按类型加锁遍历
// ---------------------------------------------------------------------------

// RangeByType 遍历指定类型的所有 Entity。
//
// 遍历期间持有组级读锁（group.mu.RLock），阻止同类型的 Add/Remove 结构变更。
// 多个 RangeByType / Get 可并发执行（均为 RLock）。
// 不同 EntityType 的操作完全隔离。
//
// fn 返回 false 提前终止遍历。
func (m *EntityManager) RangeByType(entityType mme.EntityType, fn func(mme_agent.IEntity) bool) {
	group := m.getGroup(entityType)
	if group == nil {
		return
	}
	group.Range(fn)
}

// ---------------------------------------------------------------------------
// 需求 3：Per-Entity TryLock（可重入锁）
// ---------------------------------------------------------------------------

// TryLock 尝试对指定 entityId 加锁，立即返回，非阻塞。
//
// 锁独立于 Entity 存储：Entity 未加载时即可加锁（先锁后加载场景）。
// 使用 ReentrantLock，同一 goroutine 内可重复加锁（重入）。
// 返回 true 表示加锁成功（包括重入成功），false 表示锁已被其他 goroutine 持有。
func (m *EntityManager) TryLock(entityId int64) bool {
	mu := m.getOrCreateLock(entityId)
	return mu.TryLock()
}

// TryLockWithSpin 带有限自旋的 TryLock，支持重入。
//
// 内部由 ReentrantLock 自行管理自旋策略（前 20 次 CPU 自旋，后续 Sleep(1ms)），
// 最多尝试 100 次。在短暂锁竞争场景下比 TryLock 有更高的成功率。
// 返回 true 表示加锁成功，false 表示超过最大自旋次数仍未获取到锁。
func (m *EntityManager) TryLockWithSpin(entityId int64) bool {
	mu := m.getOrCreateLock(entityId)
	return mu.TryLockWithSpin()
}

// Unlock 释放指定 entityId 的锁。
//
// 使用 ReentrantLock，支持重入解锁：
//   - 若重入计数 > 1，仅递减计数，锁仍被当前 goroutine 持有。
//   - 若重入计数 == 1，完全释放锁。
//
// 调用方必须确保当前 goroutine 持有该锁，否则会 panic。
func (m *EntityManager) Unlock(entityId int64) {
	mu, ok := m.locks.Get(entityId)
	if !ok {
		panic("entitymgr: Unlock of unknown entity lock")
	}
	mu.Unlock()
}

// CleanupLock 清理指定 entityId 的锁。
//
// 如果锁正在被其他 goroutine 持有，不会清理，返回 false。
// 如果锁未被持有且成功清理，返回 true。
//
// 注意：此方法仅应在确认安全时调用（如 Entity 已卸载且不会再被访问）。
// 在并发环境下，调用方需确保没有其他 goroutine 正在或即将使用该锁。
func (m *EntityManager) CleanupLock(entityId int64) bool {
	mu, ok := m.locks.Get(entityId)
	if !ok {
		return false
	}
	// 检查锁是否被其他 goroutine 持有
	if mu.IsLockedByOther() {
		return false
	}
	m.locks.Remove(entityId)
	return true
}

// ---------------------------------------------------------------------------
// 管理操作
// ---------------------------------------------------------------------------

// Add 添加 Entity。
//
// 在 group.mu.Lock 下同步写入 group.entities 和 typeIndex，
// 对并发的 Get/RangeByType 而言是原子的。
//
// 并发安全性证明：
//   - 并发 Get 若在 typeIndex 写入前到达 → typeIndex 查不到 → 返回 false（安全）
//   - 并发 Get 若在 typeIndex 写入后到达 → group.mu.RLock 阻塞 →
//     等 Add 释放 Lock 后获得 RLock → 看到完整 entity（安全）
func (m *EntityManager) Add(entity mme_agent.IEntity) {
	group := m.getOrCreateGroup(entity.GetEntityType())
	group.mu.Lock()
	group.add(entity)
	m.typeIndex.Set(entity.GetId(), entity.GetEntityType())
	group.mu.Unlock()
	m.count.Add(1)
}

// Remove 按 entityId 移除 Entity。
//
// 在 group.mu.Lock 下同步删除 group.entities 和 typeIndex，
// 对并发的 Get/RangeByType 而言是原子的。
// 不销毁 locks 表中的锁（锁可能仍被持有或后续复用）。
//
// 并发安全性证明：
//   - 并发 Get 若在 typeIndex 删除前到达 → group.mu.RLock 阻塞 →
//     等 Remove 释放 Lock 后获得 RLock → entity 已删 → 返回 false（安全）
//   - 并发 Get 若在 typeIndex 删除后到达 → typeIndex 查不到 → 返回 false（安全）
func (m *EntityManager) Remove(entityId int64) (mme_agent.IEntity, bool) {
	entityType, ok := m.typeIndex.Get(entityId)
	if !ok {
		return nil, false
	}
	group := m.getGroup(entityType)
	if group == nil {
		return nil, false
	}

	group.mu.Lock()
	entity, removed := group.remove(entityId)
	if removed {
		m.typeIndex.Remove(entityId)
	}
	group.mu.Unlock()

	if removed {
		m.count.Add(-1)
	}
	return entity, removed
}

// Len 返回全局已加载的 Entity 总数。
func (m *EntityManager) Len() int {
	return int(m.count.Load())
}

// LenByType 返回指定类型的 Entity 数量。
func (m *EntityManager) LenByType(entityType mme.EntityType) int {
	group := m.getGroup(entityType)
	if group == nil {
		return 0
	}
	return group.Len()
}

// ---------------------------------------------------------------------------
// 内部辅助
// ---------------------------------------------------------------------------

// getGroup 获取指定类型的 EntityGroup，不存在返回 nil。
func (m *EntityManager) getGroup(entityType mme.EntityType) *EntityGroup {
	val, ok := m.groups.Load(entityType)
	if !ok {
		return nil
	}
	return val.(*EntityGroup)
}

// getOrCreateGroup 获取或创建指定类型的 EntityGroup。
// 使用 sync.Map.LoadOrStore 保证并发安全：仅第一个到达的 goroutine 创建 Group。
func (m *EntityManager) getOrCreateGroup(entityType mme.EntityType) *EntityGroup {
	val, ok := m.groups.Load(entityType)
	if ok {
		return val.(*EntityGroup)
	}
	group := newEntityGroup()
	actual, _ := m.groups.LoadOrStore(entityType, group)
	return actual.(*EntityGroup)
}

// getOrCreateLock 获取或创建指定 entityId 的可重入锁。
// 快速路径：锁已存在，直接返回（cmap shard RLock）。
// 慢速路径：Upsert 在 shard 锁内原子创建（保证同一 entityId 全局唯一锁实例）。
func (m *EntityManager) getOrCreateLock(entityId int64) *reentrantlock.ReentrantLock {
	if mu, ok := m.locks.Get(entityId); ok {
		return mu
	}
	newMu := &reentrantlock.ReentrantLock{}
	return m.locks.Upsert(entityId, newMu,
		func(exist bool, valueInMap, newValue *reentrantlock.ReentrantLock) *reentrantlock.ReentrantLock {
			if exist {
				return valueInMap
			}
			return newValue
		})
}
