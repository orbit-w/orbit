package xmapwrapper

import (
	"gitee.com/orbit-w/meteor/bases/container/xmap"
	dt "gitee.com/orbit-w/meteor/bases/dirty/dirty_tracker"
)

type XMapContainer[K comparable, PbValue any, WrapperValue Linkable[PbValue]] interface {
	Get(key K) (WrapperValue, bool)
	Set(key K, pbValue PbValue) WrapperValue
	Delete(key K) bool
	Reset(pbMap *map[K]PbValue)
	Has(key K) bool
	Len() int
	Range(f func(key K, wrapper WrapperValue) bool)
}

// Linkable 定义可链接对象的接口
// 所有 Mechanism/Module/Manager 都应该实现此接口
type Linkable[PbValue any] interface {
	IncrementalSyncObject
	// Link 链接到父DirtyTracker
	Link(parent *dt.DirtyTracker, parentBit int64)
	// LinkFactsAccessor 链接到父DirtyTracker和xmap的FactsAccessor
	LinkFactsAccessor(parent *dt.DirtyTracker, parentBit int64, factsAccessor func())
	// Unlink 断开链接
	Unlink()
	// GetDirtyTracker 获取DirtyTracker（用于子对象的Link）
	GetDirtyTracker() *dt.DirtyTracker
	// DeepCopy 深拷贝
	DeepCopy() PbValue
}

// IncrementalSyncObject 增量同步对象接口
// 用于增量同步时，获取增量数据
type IncrementalSyncObject interface {
	ClearAllDirty()
	MarkDirty(dirtyBit int64)
	IsDirty(dirtyBit int64) bool
}

// WrapperFactory 包装器工厂函数类型
// 用于将protobuf对象转换为包装对象
type WrapperFactory[PbValue any, WrapperValue any] func(pb PbValue) WrapperValue

// XMapWrapper 通用的xmap链接管理器
// K: map的key类型
// PbValue: protobuf对象类型（通常是指针）
// WrapperValue: 包装对象类型（实现了Linkable接口）
type XMapWrapper[K comparable, PbValue any, WrapperValue Linkable[PbValue]] struct {
	pbMap          *map[K]PbValue                        // protobuf map的引用
	wrapperMap     map[K]WrapperValue                    // 包装对象map
	mapAccessor    *xmap.MapAccessor[K, PbValue]         // xmap访问器
	wrapperFactory WrapperFactory[PbValue, WrapperValue] // 包装器工厂函数
	parentTracker  *dt.DirtyTracker                      // 父DirtyTracker
	parentBit      int64                                 // 父脏标记位
}

// NewXMapWrapper 创建一个新的XMapLink实例
// pbMap: protobuf map的引用
// marker: 脏标记接口
// dirtyBit: 脏标记位
// wrapperFactory: 包装器工厂函数
func NewXMapWrapper[K comparable, PbValue any, WrapperValue Linkable[PbValue]](
	pbMap *map[K]PbValue,
	marker xmap.DirtyMarker,
	parentTracker *dt.DirtyTracker,
	dirtyBit int64,
	wrapperFactory WrapperFactory[PbValue, WrapperValue],
) *XMapWrapper[K, PbValue, WrapperValue] {

	link := &XMapWrapper[K, PbValue, WrapperValue]{
		pbMap:          pbMap,
		wrapperMap:     make(map[K]WrapperValue),
		wrapperFactory: wrapperFactory,
		parentTracker:  parentTracker,
		parentBit:      dirtyBit,
	}

	// 创建MapAccessor
	link.mapAccessor = xmap.NewMapAccessorWithMarkerForRef(pbMap, marker, dirtyBit)

	return link
}

func (x *XMapWrapper[K, PbValue, WrapperValue]) linkAll() {
	// 初始化现有的map元素
	if x.pbMap != nil && *x.pbMap != nil {
		for key, pbValue := range *x.pbMap {
			wrapper := x.wrapperFactory(pbValue)
			tracker := func() {
				x.mapAccessor.TrackSet(key)
			}
			wrapper.LinkFactsAccessor(x.parentTracker, x.parentBit, tracker)
			x.wrapperMap[key] = wrapper
		}
	}
}

// Get 获取包装对象
// 返回包装对象和是否存在的标志
func (x *XMapWrapper[K, PbValue, WrapperValue]) Get(key K) (WrapperValue, bool) {
	wrapper, ok := x.wrapperMap[key]
	return wrapper, ok
}

// Set 设置/添加对象
// 如果key已存在，会先Unlink旧对象
// 返回新的包装对象
func (x *XMapWrapper[K, PbValue, WrapperValue]) Set(key K, pbValue PbValue) WrapperValue {
	// 处理旧对象的Unlink
	x.Delete(key)

	// 创建新的包装对象
	wrapper := x.wrapperFactory(pbValue)

	// Link到父对象和xmap的FactsAccessor
	tracker := func() {
		x.mapAccessor.TrackSet(key)
	}
	wrapper.LinkFactsAccessor(x.parentTracker, x.parentBit, tracker)

	// 更新protobuf map（通过MapAccessor，会自动处理TrackSetWithDelete）
	x.mapAccessor.Set(key, pbValue)

	// 更新包装对象map
	x.wrapperMap[key] = wrapper

	return wrapper
}

func (x *XMapWrapper[K, PbValue, WrapperValue]) SetWithoutTrack(key K, pbValue PbValue) WrapperValue {
	// 处理旧对象的Unlink
	x.Delete(key)

	// 创建新的包装对象
	wrapper := x.wrapperFactory(pbValue)

	// Link到父对象和xmap的FactsAccessor
	tracker := func() {
		x.mapAccessor.TrackSet(key)
	}
	wrapper.LinkFactsAccessor(x.parentTracker, x.parentBit, tracker)

	// 更新protobuf map（通过MapAccessor，会自动处理TrackSetWithDelete）
	x.mapAccessor.SetWithoutTrack(key, pbValue)

	// 更新包装对象map
	x.wrapperMap[key] = wrapper

	return wrapper
}

// Delete 删除对象
// 返回是否成功删除
func (x *XMapWrapper[K, PbValue, WrapperValue]) Delete(key K) bool {
	if wrapper, exists := x.wrapperMap[key]; exists {
		// 从protobuf map中删除（通过MapAccessor）
		x.mapAccessor.Delete(key)

		// Unlink包装对象, 解除与父对象和xmap的关联
		wrapper.Unlink()

		// 从包装对象map中删除
		delete(x.wrapperMap, key)

		return true
	}
	return false
}

func (x *XMapWrapper[K, PbValue, WrapperValue]) Reset(pbMap *map[K]PbValue) {
	x.CleanLinks()

	x.pbMap = pbMap
	x.mapAccessor.Reset(x.pbMap)

	x.linkAll()
}

// Has 检查key是否存在
func (x *XMapWrapper[K, PbValue, WrapperValue]) Has(key K) bool {
	_, ok := x.wrapperMap[key]
	return ok
}

// Len 返回map的长度
func (x *XMapWrapper[K, PbValue, WrapperValue]) Len() int {
	return len(x.wrapperMap)
}

// Range 遍历所有对象
// 回调函数返回false时停止遍历
func (x *XMapWrapper[K, PbValue, WrapperValue]) Range(f func(key K, wrapper WrapperValue) bool) {
	if f == nil {
		return
	}
	for key, wrapper := range x.wrapperMap {
		if !f(key, wrapper) {
			break
		}
	}
}

// Clear 清空所有对象
func (x *XMapWrapper[K, PbValue, WrapperValue]) Clear() {
	x.CleanLinks()

	// 清空protobuf map（通过MapAccessor）
	x.mapAccessor.Clear()
}

func (x *XMapWrapper[K, PbValue, WrapperValue]) CleanLinks() {
	// Unlink所有包装对象
	for _, wrapper := range x.wrapperMap {
		wrapper.Unlink()
	}

	// 清空包装对象map
	x.wrapperMap = make(map[K]WrapperValue)
}

// GetMapAccessor 获取MapAccessor（用于高级操作）
func (x *XMapWrapper[K, PbValue, WrapperValue]) GetMapAccessor() *xmap.MapAccessor[K, PbValue] {
	return x.mapAccessor
}

// Keys 返回所有key
func (x *XMapWrapper[K, PbValue, WrapperValue]) Keys() []K {
	keys := make([]K, 0, len(x.wrapperMap))
	for key := range x.wrapperMap {
		keys = append(keys, key)
	}
	return keys
}

// Values 返回所有包装对象
func (x *XMapWrapper[K, PbValue, WrapperValue]) Values() []WrapperValue {
	values := make([]WrapperValue, 0, len(x.wrapperMap))
	for _, wrapper := range x.wrapperMap {
		values = append(values, wrapper)
	}
	return values
}

func (x *XMapWrapper[K, PbValue, WrapperValue]) RangeOperations(f func(key K, operation xmap.MapOperation[K]) bool) {
	if f == nil {
		return
	}
	x.mapAccessor.RangeOperations(f)
}

// RangeIncrementalSyncObject 遍历所有增量同步对象
func (x *XMapWrapper[K, PbValue, WrapperValue]) RangeIncrementalSyncObject(f func(key K, object IncrementalSyncObject) (stop bool)) {
	if f == nil {
		return
	}
	for k := range x.wrapperMap {
		wrapper := x.wrapperMap[k]
		if f(k, wrapper) {
			break
		}
	}
}

// Clone 对底层Map的浅拷贝
func (x *XMapWrapper[K, PbValue, WrapperValue]) Clone() map[K]PbValue {
	if x == nil {
		return nil
	}
	return x.mapAccessor.Clone()
}

// DeepCopy 对底层Map的深拷贝
func (x *XMapWrapper[K, PbValue, WrapperValue]) DeepCopy(copy *map[K]PbValue) {
	if x == nil || copy == nil {
		return
	}
	x.Range(func(key K, wrapper WrapperValue) bool {
		(*copy)[key] = wrapper.DeepCopy()
		return true
	})
}
