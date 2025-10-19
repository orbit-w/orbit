package xmap

// OperationType 操作类型
type OperationType int

const (
	SetOperation    OperationType = iota // 设置操作
	DeleteOperation                      // 删除操作
)

// MapOperation 表示一个 Map 操作
type MapOperation[K comparable] struct {
	Type OperationType
	Key  K
}

// FinegrainedChangeTracker 细粒度变更跟踪器
// 用于跟踪 Map 字段的具体键级别变更，支持高效的 MongoDB 增量更新
type FinegrainedChangeTracker[K comparable, V any] struct {
	operations     map[K]MapOperation[K] // 统一的操作记录，只记录操作类型和键
	isTrackingMode bool                  // 是否启用细粒度跟踪模式
	hasFullReplace bool                  // 是否有完整替换操作
}

// NewFinegrainedChangeTracker 创建新的细粒度变更跟踪器
func NewFinegrainedChangeTracker[K comparable, V any]() *FinegrainedChangeTracker[K, V] {
	return &FinegrainedChangeTracker[K, V]{
		operations:     make(map[K]MapOperation[K]),
		isTrackingMode: true,
	}
}

// EnableTracking 启用细粒度跟踪模式
func (t *FinegrainedChangeTracker[K, V]) EnableTracking() {
	t.isTrackingMode = true
}

// DisableTracking 禁用细粒度跟踪模式
func (t *FinegrainedChangeTracker[K, V]) DisableTracking() {
	t.isTrackingMode = false
}

// TrackSet 跟踪设置操作
func (t *FinegrainedChangeTracker[K, V]) TrackSet(key K, value V) {
	if !t.isTrackingMode {
		return
	}

	t.operations[key] = MapOperation[K]{
		Type: SetOperation,
		Key:  key,
	}
}

// TrackUnset 跟踪删除操作
func (t *FinegrainedChangeTracker[K, V]) TrackUnset(key K) {
	if !t.isTrackingMode {
		return
	}

	t.operations[key] = MapOperation[K]{
		Type: DeleteOperation,
		Key:  key,
	}
}

// TrackFullReplace 跟踪完整替换操作
func (t *FinegrainedChangeTracker[K, V]) TrackFullReplace() {
	t.hasFullReplace = true
	t.clear()
}

// clear 清空所有跟踪的操作
func (t *FinegrainedChangeTracker[K, V]) clear() {
	t.operations = make(map[K]MapOperation[K])
}

// RangeOperations 遍历所有跟踪的map写操作
func (t *FinegrainedChangeTracker[K, V]) RangeOperations(f func(key K, operation MapOperation[K]) bool) {
	for key, operation := range t.operations {
		if !f(key, operation) {
			break
		}
	}
}

// GetSetOperations 获取设置操作的键
func (t *FinegrainedChangeTracker[K, V]) GetSetOperations() []K {
	var setKeys []K
	for _, op := range t.operations {
		if op.Type == SetOperation {
			setKeys = append(setKeys, op.Key)
		}
	}
	return setKeys
}

// GetUnsetKeys 获取删除操作的键
func (t *FinegrainedChangeTracker[K, V]) GetUnsetKeys() []K {
	var unsetKeys []K
	for _, op := range t.operations {
		if op.Type == DeleteOperation {
			unsetKeys = append(unsetKeys, op.Key)
		}
	}
	return unsetKeys
}

// HasChanges 检查是否有变更
func (t *FinegrainedChangeTracker[K, V]) HasChanges() bool {
	return len(t.operations) > 0 || t.hasFullReplace
}

// HasFullReplace 检查是否有完整替换
func (t *FinegrainedChangeTracker[K, V]) HasFullReplace() bool {
	return t.hasFullReplace
}

// Reset 重置跟踪器
func (t *FinegrainedChangeTracker[K, V]) Reset() {
	t.clear()
	t.hasFullReplace = false
}
