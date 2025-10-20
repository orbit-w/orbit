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
type FinegrainedChangeTracker[K comparable] struct {
	operations     map[K]MapOperation[K] // 统一的操作记录，只记录操作类型和键
	isTrackingMode bool                  // 是否启用细粒度跟踪模式
}

// NewFinegrainedChangeTracker 创建新的细粒度变更跟踪器
func NewFinegrainedChangeTracker[K comparable]() *FinegrainedChangeTracker[K] {
	return &FinegrainedChangeTracker[K]{
		operations:     make(map[K]MapOperation[K]),
		isTrackingMode: true,
	}
}

// EnableTracking 启用细粒度跟踪模式
func (t *FinegrainedChangeTracker[K]) EnableTracking() {
	t.isTrackingMode = true
}

// DisableTracking 禁用细粒度跟踪模式
func (t *FinegrainedChangeTracker[K]) DisableTracking() {
	t.isTrackingMode = false
}

// TrackSet 跟踪设置操作
func (t *FinegrainedChangeTracker[K]) TrackSet(key K) {
	if !t.isTrackingMode {
		return
	}

	t.operations[key] = MapOperation[K]{
		Type: SetOperation,
		Key:  key,
	}
}

// TrackUnset 跟踪删除操作
func (t *FinegrainedChangeTracker[K]) TrackUnset(key K) {
	if !t.isTrackingMode {
		return
	}

	t.operations[key] = MapOperation[K]{
		Type: DeleteOperation,
		Key:  key,
	}
}

// TrackFullReplace 跟踪完整替换操作
func (t *FinegrainedChangeTracker[K]) TrackFullReplace() {

	t.clear()
}

// clear 清空所有跟踪的操作
func (t *FinegrainedChangeTracker[K]) clear() {
	t.operations = make(map[K]MapOperation[K])
}

// RangeOperations 遍历所有跟踪的map写操作
func (t *FinegrainedChangeTracker[K]) RangeOperations(f func(key K, operation MapOperation[K]) bool) {
	for key, operation := range t.operations {
		if !f(key, operation) {
			break
		}
	}
}

// GetSetOperations 获取设置操作的键
func (t *FinegrainedChangeTracker[K]) GetSetOperations() []K {
	var setKeys []K
	for _, op := range t.operations {
		if op.Type == SetOperation {
			setKeys = append(setKeys, op.Key)
		}
	}
	return setKeys
}

// GetUnsetKeys 获取删除操作的键
func (t *FinegrainedChangeTracker[K]) GetUnsetKeys() []K {
	var unsetKeys []K
	for _, op := range t.operations {
		if op.Type == DeleteOperation {
			unsetKeys = append(unsetKeys, op.Key)
		}
	}
	return unsetKeys
}

// HasChanges 检查是否有变更
func (t *FinegrainedChangeTracker[K]) HasChanges() bool {
	return len(t.operations) > 0
}

// Reset 重置跟踪器
func (t *FinegrainedChangeTracker[K]) Reset() {
	t.clear()
}
