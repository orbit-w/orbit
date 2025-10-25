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
	operations     map[K][]MapOperation[K] // 操作序列记录，支持同一个 key 的多个操作
	isTrackingMode bool                    // 是否启用细粒度跟踪模式
	isRefType      bool                    // 是否为引用类型（引用类型需要特殊处理）
}

// NewFinegrainedChangeTracker 创建新的细粒度变更跟踪器（值类型）
func NewFinegrainedChangeTracker[K comparable]() *FinegrainedChangeTracker[K] {
	return &FinegrainedChangeTracker[K]{
		operations:     make(map[K][]MapOperation[K]),
		isTrackingMode: true,
		isRefType:      false,
	}
}

// NewFinegrainedChangeTrackerForRef 创建新的细粒度变更跟踪器（引用类型）
func NewFinegrainedChangeTrackerForRef[K comparable]() *FinegrainedChangeTracker[K] {
	return &FinegrainedChangeTracker[K]{
		operations:     make(map[K][]MapOperation[K]),
		isTrackingMode: true,
		isRefType:      true,
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

	newOp := MapOperation[K]{
		Type: SetOperation,
		Key:  key,
	}

	if t.isRefType {
		// 引用类型：SetOperation 和 DeleteOperation 不能彼此覆盖
		ops := t.operations[key]
		if len(ops) == 0 {
			// 没有操作，直接添加
			t.operations[key] = []MapOperation[K]{newOp}
		} else {
			lastOp := ops[len(ops)-1]
			if lastOp.Type == SetOperation {
				// 新操作与末尾操作相同，不发生变化（覆盖最后一个）
				ops[len(ops)-1] = newOp
				t.operations[key] = ops
			} else {
				// 新操作与末尾操作不同，Append 后移除前面所有相同类型的操作
				// 先追加新操作
				ops = append(ops, newOp)
				// 移除前面所有 SetOperation（除了最后一个）
				t.operations[key] = t.removeOperationsExceptLast(ops, SetOperation)
			}
		}
	} else {
		// 值类型：直接覆盖
		t.operations[key] = []MapOperation[K]{newOp}
	}
}

// TrackSetWithDelete 跟踪引用类型的设置操作（先删除旧值再设置新值）
// 当 keyExists 为 true 时，会先记录 DeleteOperation，然后记录 SetOperation
func (t *FinegrainedChangeTracker[K]) TrackSetWithDelete(key K, keyExists bool) {
	if !t.isTrackingMode {
		return
	}

	if !t.isRefType {
		// 值类型直接调用 TrackSet
		t.TrackSet(key)
		return
	}

	// 引用类型：如果 key 已存在，需要先记录删除操作
	if keyExists {
		ops := t.operations[key]

		if len(ops) == 0 {
			// 没有操作，直接添加 Delete + Set
			t.operations[key] = t.makeOperationPair(key, DeleteOperation, SetOperation)
		} else {
			lastOp := ops[len(ops)-1]

			if lastOp.Type == SetOperation {
				// 最后是 Set，追加 Delete + Set
				t.operations[key] = t.appendAndOptimize(ops, key, DeleteOperation, SetOperation)
			} else {
				// 最后是 Delete，覆盖 Delete 后追加 Set
				t.operations[key] = t.replaceLastAndAppend(ops, key, DeleteOperation, SetOperation)
			}
		}
	} else {
		// key 不存在，直接设置
		t.TrackSet(key)
	}
}

// TrackUnset 跟踪删除操作
func (t *FinegrainedChangeTracker[K]) TrackUnset(key K) {
	if !t.isTrackingMode {
		return
	}

	newOp := MapOperation[K]{
		Type: DeleteOperation,
		Key:  key,
	}

	if t.isRefType {
		// 引用类型：SetOperation 和 DeleteOperation 不能彼此覆盖
		ops := t.operations[key]
		if len(ops) == 0 {
			// 没有操作，直接添加
			t.operations[key] = []MapOperation[K]{newOp}
		} else {
			lastOp := ops[len(ops)-1]
			if lastOp.Type == DeleteOperation {
				// 新操作与末尾操作相同，不发生变化（覆盖最后一个）
				ops[len(ops)-1] = newOp
				t.operations[key] = ops
			} else {
				// 新操作与末尾操作不同，Append 后移除前面所有相同类型的操作
				// 先追加新操作
				ops = append(ops, newOp)
				// 移除前面所有 DeleteOperation（除了最后一个）
				t.operations[key] = t.removeOperationsExceptLast(ops, DeleteOperation)
			}
		}
	} else {
		// 值类型：直接覆盖
		t.operations[key] = []MapOperation[K]{newOp}
	}
}

// removeOperationsExceptLast 移除操作序列中除了最后一个指定类型的操作
// 例如：[Set1, Delete, Set2, Delete, Set3] 移除 SetOperation 后变成 [Delete, Delete, Set3]
func (t *FinegrainedChangeTracker[K]) removeOperationsExceptLast(ops []MapOperation[K], opType OperationType) []MapOperation[K] {
	if len(ops) <= 1 {
		return ops
	}

	// 从后往前找到最后一个指定类型的操作的索引
	lastIndex := -1
	for i := len(ops) - 1; i >= 0; i-- {
		if ops[i].Type == opType {
			lastIndex = i
			break
		}
	}

	if lastIndex == -1 {
		// 没有找到指定类型的操作
		return ops
	}

	// 构建新的操作序列，跳过前面所有与指定类型相同的操作
	result := make([]MapOperation[K], 0, len(ops))
	for i := 0; i < len(ops); i++ {
		if i == lastIndex {
			// 保留最后一个指定类型的操作
			result = append(result, ops[i])
		} else if ops[i].Type != opType {
			// 保留不同类型的操作
			result = append(result, ops[i])
		}
		// 跳过前面所有与指定类型相同的操作
	}

	return result
}

// makeOperationPair 创建一对操作
func (t *FinegrainedChangeTracker[K]) makeOperationPair(key K, firstType, secondType OperationType) []MapOperation[K] {
	return []MapOperation[K]{
		{Type: firstType, Key: key},
		{Type: secondType, Key: key},
	}
}

// appendAndOptimize 追加两个操作并优化
// 先追加 firstType，优化后再追加 secondType，再次优化
func (t *FinegrainedChangeTracker[K]) appendAndOptimize(ops []MapOperation[K], key K, firstType, secondType OperationType) []MapOperation[K] {
	// 追加第一个操作
	ops = append(ops, MapOperation[K]{Type: firstType, Key: key})
	// 优化第一个操作类型
	ops = t.removeOperationsExceptLast(ops, firstType)
	// 追加第二个操作
	ops = append(ops, MapOperation[K]{Type: secondType, Key: key})
	// 优化第二个操作类型
	return t.removeOperationsExceptLast(ops, secondType)
}

// replaceLastAndAppend 替换最后一个操作并追加新操作
// 用 firstType 覆盖最后一个操作，然后追加 secondType 并优化
func (t *FinegrainedChangeTracker[K]) replaceLastAndAppend(ops []MapOperation[K], key K, firstType, secondType OperationType) []MapOperation[K] {
	if len(ops) > 0 {
		// 覆盖最后一个操作
		ops[len(ops)-1] = MapOperation[K]{Type: firstType, Key: key}
	}
	// 追加第二个操作
	ops = append(ops, MapOperation[K]{Type: secondType, Key: key})
	// 优化第二个操作类型
	return t.removeOperationsExceptLast(ops, secondType)
}

// TrackFullReplace 跟踪完整替换操作
func (t *FinegrainedChangeTracker[K]) TrackFullReplace() {
	t.clear()
}

// clear 清空所有跟踪的操作
func (t *FinegrainedChangeTracker[K]) clear() {
	t.operations = make(map[K][]MapOperation[K])
}

// RangeOperations 遍历所有跟踪的map写操作（遍历每个 key 的所有操作）
func (t *FinegrainedChangeTracker[K]) RangeOperations(f func(key K, operation MapOperation[K]) bool) {
	for key, ops := range t.operations {
		for _, operation := range ops {
			if !f(key, operation) {
				return
			}
		}
	}
}

// GetSetOperations 获取设置操作的键（只返回最终状态为 Set 的键）
func (t *FinegrainedChangeTracker[K]) GetSetOperations() []K {
	var setKeys []K
	for key, ops := range t.operations {
		if len(ops) > 0 && ops[len(ops)-1].Type == SetOperation {
			setKeys = append(setKeys, key)
		}
	}
	return setKeys
}

// GetUnsetKeys 获取删除操作的键（只返回最终状态为 Delete 的键）
func (t *FinegrainedChangeTracker[K]) GetUnsetKeys() []K {
	var unsetKeys []K
	for key, ops := range t.operations {
		if len(ops) > 0 && ops[len(ops)-1].Type == DeleteOperation {
			unsetKeys = append(unsetKeys, key)
		}
	}
	return unsetKeys
}

// GetOperations 获取指定 key 的所有操作序列
func (t *FinegrainedChangeTracker[K]) GetOperations(key K) []MapOperation[K] {
	return t.operations[key]
}

// HasChanges 检查是否有变更
func (t *FinegrainedChangeTracker[K]) HasChanges() bool {
	return len(t.operations) > 0
}

// Reset 重置跟踪器
func (t *FinegrainedChangeTracker[K]) Reset() {
	t.clear()
}
