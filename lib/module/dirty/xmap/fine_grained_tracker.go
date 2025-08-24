package xmap

// FinegrainedChangeTracker 细粒度变更跟踪器
// 用于跟踪 Map 字段的具体键级别变更，支持高效的 MongoDB 增量更新
type FinegrainedChangeTracker[K comparable, V any] struct {
	setOperations  map[K]V    // 设置的键值对
	unsetKeys      map[K]bool // 删除的键
	incOperations  map[K]V    // 递增操作（仅适用于数值类型）
	isTrackingMode bool       // 是否启用细粒度跟踪模式
	hasFullReplace bool       // 是否有完整替换操作
}

// NewFinegrainedChangeTracker 创建新的细粒度变更跟踪器
func NewFinegrainedChangeTracker[K comparable, V any]() *FinegrainedChangeTracker[K, V] {
	return &FinegrainedChangeTracker[K, V]{
		setOperations:  make(map[K]V),
		unsetKeys:      make(map[K]bool),
		incOperations:  make(map[K]V),
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

	t.setOperations[key] = value
	// 如果之前标记为删除，现在取消
	delete(t.unsetKeys, key)
}

// TrackUnset 跟踪删除操作
func (t *FinegrainedChangeTracker[K, V]) TrackUnset(key K) {
	if !t.isTrackingMode {
		return
	}

	t.unsetKeys[key] = true
	// 如果之前标记为设置，现在取消
	delete(t.setOperations, key)
}

// TrackInc 跟踪递增操作（仅适用于数值类型）
func (t *FinegrainedChangeTracker[K, V]) TrackInc(key K, value V) {
	if !t.isTrackingMode {
		return
	}

	t.incOperations[key] = value
}

// TrackFullReplace 跟踪完整替换操作
func (t *FinegrainedChangeTracker[K, V]) TrackFullReplace() {
	t.hasFullReplace = true
	t.clear()
}

// clear 清空所有跟踪的操作
func (t *FinegrainedChangeTracker[K, V]) clear() {
	t.setOperations = make(map[K]V)
	t.unsetKeys = make(map[K]bool)
	t.incOperations = make(map[K]V)
}

// GetSetOperations 获取设置操作
func (t *FinegrainedChangeTracker[K, V]) GetSetOperations() map[K]V {
	return t.setOperations
}

// GetUnsetKeys 获取删除的键
func (t *FinegrainedChangeTracker[K, V]) GetUnsetKeys() map[K]bool {
	return t.unsetKeys
}

// GetIncOperations 获取递增操作
func (t *FinegrainedChangeTracker[K, V]) GetIncOperations() map[K]V {
	return t.incOperations
}

// HasChanges 检查是否有变更
func (t *FinegrainedChangeTracker[K, V]) HasChanges() bool {
	return len(t.setOperations) > 0 || len(t.unsetKeys) > 0 || len(t.incOperations) > 0 || t.hasFullReplace
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

// Enhanced MapAccessor with fine-grained change tracking
type EnhancedMapAccessor[K comparable, V any] struct {
	MapAccessor[K, V]
	changeTracker *FinegrainedChangeTracker[K, V]
}

// NewEnhancedMapAccessor 创建增强的 Map 访问器
func NewEnhancedMapAccessor[K comparable, V any](m *map[K]V, marker DirtyMarker, dirtyBit int64) EnhancedMapAccessor[K, V] {
	return EnhancedMapAccessor[K, V]{
		MapAccessor:   NewMapAccessorWithMarker(m, marker, dirtyBit),
		changeTracker: NewFinegrainedChangeTracker[K, V](),
	}
}

// Set 重写设置方法，增加细粒度跟踪
func (op EnhancedMapAccessor[K, V]) Set(key K, value V) {
	op.MapAccessor.Set(key, value)
	op.changeTracker.TrackSet(key, value)
}

// Delete 重写删除方法，增加细粒度跟踪
func (op EnhancedMapAccessor[K, V]) Delete(key K) {
	op.MapAccessor.Delete(key)
	op.changeTracker.TrackUnset(key)
}

// ReplaceAll 重写替换所有方法，标记为完整替换
func (op EnhancedMapAccessor[K, V]) ReplaceAll(entries map[K]V) {
	op.MapAccessor.ReplaceAll(entries)
	op.changeTracker.TrackFullReplace()
}

// Inc 数值递增操作（仅适用于数值类型）
func (op EnhancedMapAccessor[K, V]) Inc(key K, delta V) V {
	// 注意：这里需要类型约束来确保 V 是数值类型
	// 实际使用时可能需要使用泛型约束或接口
	current, exists := op.Get(key)
	if !exists {
		op.Set(key, delta)
		op.changeTracker.TrackInc(key, delta)
		return delta
	}

	// 这里需要根据具体类型进行加法运算
	// 为了简化，假设有一个加法接口
	if adder, ok := any(current).(interface{ Add(V) V }); ok {
		newValue := adder.Add(delta)
		op.Set(key, newValue)
		op.changeTracker.TrackInc(key, delta)
		return newValue
	}

	// 回退到设置操作
	op.Set(key, delta)
	return delta
}

// GetChangeTracker 获取变更跟踪器
func (op EnhancedMapAccessor[K, V]) GetChangeTracker() *FinegrainedChangeTracker[K, V] {
	return op.changeTracker
}

// EnableFinegrainedTracking 启用细粒度跟踪
func (op EnhancedMapAccessor[K, V]) EnableFinegrainedTracking() {
	op.changeTracker.EnableTracking()
}

// DisableFinegrainedTracking 禁用细粒度跟踪
func (op EnhancedMapAccessor[K, V]) DisableFinegrainedTracking() {
	op.changeTracker.DisableTracking()
}

// ResetChangeTracking 重置变更跟踪
func (op EnhancedMapAccessor[K, V]) ResetChangeTracking() {
	op.changeTracker.Reset()
}
