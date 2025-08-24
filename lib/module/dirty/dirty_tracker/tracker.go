package dirty

// ChangeCallback 脏标记变化回调函数类型
// oldFlag: 变更前的标记值
// newFlag: 变更后的标记值
// changedBits: 发生变化的位
type ChangeCallback func(oldFlag, newFlag, changedBits int64)

type DirtyTracker struct {
	dirtyFlag int64 // 普通的脏标记位，支持64个不同的脏标记

	// 嵌套支持字段
	parent    *DirtyTracker // 父级追踪器
	parentBit int64         // 在父级中对应的脏标记位
}

// =============================================================================
// 非线程安全版本 - 嵌套支持操作（极致性能）
// =============================================================================

// Link 将当前 DirtyTracker 链接到父级 DirtyTracker
// 当当前对象变脏时，会自动标记父对象的指定位为脏
// parent: 父级 DirtyTracker
// parentBit: 在父级中对应的脏标记位
// 注意：非线程安全，需要外部同步
func (d *DirtyTracker) Link(parent *DirtyTracker, parentBit int64) {
	if parent == nil {
		return
	}
	d.parent = parent
	d.parentBit = parentBit
}

// Unlink 断开与父级 FastDirtyTracker 的链接
// 注意：非线程安全，需要外部同步
func (d *DirtyTracker) Unlink() {
	d.parent = nil
	d.parentBit = 0
}

// GetParent 获取父级 FastDirtyTracker（非线程安全）
func (d *DirtyTracker) GetParent() *DirtyTracker {
	return d.parent
}

// propagateToParent 向父级传播脏标记（内部方法，非线程安全）
func (d *DirtyTracker) propagateToParent() {
	if d.parent != nil && d.parentBit != 0 {
		d.parent.MarkDirty(d.parentBit)
	}
}

// =============================================================================
// 非线程安全版本 - 核心脏标记操作（极致性能）
// =============================================================================

// MarkDirty 标记指定位为脏（非线程安全，极致性能）
// 支持嵌套传播：如果有父级对象，会自动标记父级对应位为脏
func (d *DirtyTracker) MarkDirty(bit int64) {
	wasClean := d.dirtyFlag == 0
	d.dirtyFlag |= bit

	// 如果从干净状态变为脏状态，向父级传播
	if wasClean && d.dirtyFlag != 0 {
		d.propagateToParent()
	}
}

// IsDirty 检查指定位是否为脏（非线程安全）
func (d *DirtyTracker) IsDirty(bit int64) bool {
	return d.dirtyFlag&bit != 0
}

// ClearDirty 清除指定的脏标记位（非线程安全）
func (d *DirtyTracker) ClearDirty(bit int64) {
	d.dirtyFlag &^= bit
}

// ClearAllDirty 清除所有脏标记（非线程安全）
func (d *DirtyTracker) ClearAllDirty() {
	d.dirtyFlag = 0
}

// GetDirtyFlag 获取完整的脏标记值（非线程安全）
func (d *DirtyTracker) GetDirtyFlag() int64 {
	return d.dirtyFlag
}

// SetDirtyFlag 直接设置脏标记值（非线程安全）
// 通常用于序列化/反序列化场景
func (d *DirtyTracker) SetDirtyFlag(flag int64) {
	d.dirtyFlag = flag
}

// =============================================================================
// 非线程安全版本 - 便捷操作方法
// =============================================================================

// HasAnyDirty 检查是否有任何脏标记（非线程安全）
func (d *DirtyTracker) HasAnyDirty() bool {
	return d.dirtyFlag != 0
}

// MarkMultipleDirty 同时标记多个位为脏（非线程安全）
// 支持嵌套传播：如果有父级对象，会自动标记父级对应位为脏
func (d *DirtyTracker) MarkMultipleDirty(bits ...int64) {
	wasClean := d.dirtyFlag == 0
	for _, bit := range bits {
		d.dirtyFlag |= bit
	}

	// 如果从干净状态变为脏状态，向父级传播
	if wasClean && d.dirtyFlag != 0 {
		d.propagateToParent()
	}
}

// ClearMultipleDirty 同时清除多个脏标记位（非线程安全）
func (d *DirtyTracker) ClearMultipleDirty(bits ...int64) {
	for _, bit := range bits {
		d.dirtyFlag &^= bit
	}
}

// IsAnyDirty 检查指定的任意一个位是否为脏（非线程安全）
func (d *DirtyTracker) IsAnyDirty(bits ...int64) bool {
	var combined int64
	for _, bit := range bits {
		combined |= bit
	}
	return d.dirtyFlag&combined != 0
}

// IsAllDirty 检查指定的所有位是否都为脏（非线程安全）
func (d *DirtyTracker) IsAllDirty(bits ...int64) bool {
	var combined int64
	for _, bit := range bits {
		combined |= bit
	}
	return d.dirtyFlag&combined == combined
}

// =============================================================================
// 非线程安全版本 - 扩展功能（支持回调观察者）
// =============================================================================

// FastDirtyMarker 接口，用于支持快速版本的不同类型的 MarkDirty 调用
type FastDirtyMarker interface {
	MarkDirty(bit int64)
}

// DirtyTrackerWithCallbacks 支持回调的快速脏标记追踪器（非线程安全版本）
// 当脏标记发生变化时，会自动调用注册的回调函数
type DirtyTrackerWithCallbacks struct {
	DirtyTracker
	callbacks    []ChangeCallback
	parentMarker FastDirtyMarker // 父级标记器，支持回调传播
	parentBit    int64           // 在父级中对应的脏标记位
}

// NewFastDirtyTrackerWithCallbacks 创建支持回调的快速脏标记追踪器（非线程安全）
func NewFastDirtyTrackerWithCallbacks() *DirtyTrackerWithCallbacks {
	return &DirtyTrackerWithCallbacks{
		callbacks: make([]ChangeCallback, 0),
	}
}

// AddCallback 添加脏标记变化回调（非线程安全版本）
// 回调函数会在脏标记发生变化时被调用
func (d *DirtyTrackerWithCallbacks) AddCallback(callback ChangeCallback) {
	d.callbacks = append(d.callbacks, callback)
}

// Link 将当前 FastDirtyTrackerWithCallbacks 链接到父级 FastDirtyTracker
// 支持回调版本链接到普通版本或其他回调版本
func (d *DirtyTrackerWithCallbacks) Link(parent interface{}, parentBit int64) {
	switch p := parent.(type) {
	case *DirtyTracker:
		d.DirtyTracker.Link(p, parentBit)
		d.parentMarker = p
		d.parentBit = parentBit
	case *DirtyTrackerWithCallbacks:
		d.DirtyTracker.Link(&p.DirtyTracker, parentBit)
		d.parentMarker = p
		d.parentBit = parentBit
	}
}

// propagateToParent 重写父级传播方法，确保使用正确的 MarkDirty
func (d *DirtyTrackerWithCallbacks) propagateToParent() {
	if d.parentMarker != nil && d.parentBit != 0 {
		d.parentMarker.MarkDirty(d.parentBit)
	}
}

// MarkDirty 重写标记方法，支持回调通知（非线程安全版本）
// 支持嵌套传播：如果有父级对象，会自动标记父级对应位为脏
func (d *DirtyTrackerWithCallbacks) MarkDirty(bit int64) {
	old := d.dirtyFlag
	wasClean := d.dirtyFlag == 0
	d.dirtyFlag |= bit
	// 只有在实际发生变化时才通知回调
	if old != d.dirtyFlag {
		changedBits := d.dirtyFlag &^ old
		for _, callback := range d.callbacks {
			callback(old, d.dirtyFlag, changedBits)
		}
	}

	// 如果从干净状态变为脏状态，向父级传播
	if wasClean && d.dirtyFlag != 0 {
		d.propagateToParent()
	}
}
