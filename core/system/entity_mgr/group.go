package entitymgr

import (
	"sync"

	"gitee.com/orbit-w/orbit/internal/game/mme_agent"
)

// ---------------------------------------------------------------------------
// EntityGroup — 同类型 Entity 的线程安全并发集合
// ---------------------------------------------------------------------------
//
// 使用 sync.RWMutex + map 实现，提供组级读锁遍历语义：
//   - Range/Get 持有 RLock，遍历和查找期间阻止 Add/Remove 结构变更
//   - 多个 Range/Get 可并发执行（多读者共享 RLock）
//   - Add/Remove 持有 Lock，与所有读操作互斥
//
// 注意：add/remove 是内部方法，不自行加锁。
// 由 EntityManager 在持有 mu.Lock 的前提下调用，以保证 typeIndex 和
// group.entities 在同一锁区间内原子更新。
type EntityGroup struct {
	mu       sync.RWMutex
	entities map[int64]mme_agent.IEntity
}

func newEntityGroup() *EntityGroup {
	return &EntityGroup{
		entities: make(map[int64]mme_agent.IEntity),
	}
}

// add 在已持有 mu.Lock 的前提下写入 Entity（内部方法，由 EntityManager 调用）。
func (g *EntityGroup) add(entity mme_agent.IEntity) {
	g.entities[entity.GetId()] = entity
}

// remove 在已持有 mu.Lock 的前提下删除 Entity（内部方法，由 EntityManager 调用）。
// 返回被移除的 Entity 和是否存在。
func (g *EntityGroup) remove(entityID int64) (mme_agent.IEntity, bool) {
	entity, ok := g.entities[entityID]
	if ok {
		delete(g.entities, entityID)
	}
	return entity, ok
}

// Get 按 ID 查找，O(1)（线程安全，组级读锁）。
func (g *EntityGroup) Get(entityID int64) (mme_agent.IEntity, bool) {
	g.mu.RLock()
	entity, ok := g.entities[entityID]
	g.mu.RUnlock()
	return entity, ok
}

// Range 遍历所有 Entity，fn 返回 false 提前终止。
// 遍历期间持有组级读锁，阻止 Add/Remove 结构变更。
// 多个 Range/Get 可并发执行（均为 RLock）。
func (g *EntityGroup) Range(fn func(entity mme_agent.IEntity) bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	for _, e := range g.entities {
		if !fn(e) {
			return
		}
	}
}

// Len 返回集合大小（线程安全，组级读锁）。
func (g *EntityGroup) Len() int {
	g.mu.RLock()
	n := len(g.entities)
	g.mu.RUnlock()
	return n
}
