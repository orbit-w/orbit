package xmaplink

import (
	dt "gitee.com/orbit-w/meteor/bases/dirty/dirty_tracker"
	"gitee.com/orbit-w/meteor/bases/dirty/xmap"
)

// Linkable 定义可链接对象的接口
// 所有 Mechanism/Module/Manager 都应该实现此接口
type Linkable[K comparable] interface {
	// Link 链接到父DirtyTracker
	Link(parent *dt.DirtyTracker, parentBit int64)
	// LinkFactsAccessor 链接到父DirtyTracker和xmap的FactsAccessor
	LinkFactsAccessor(parent *dt.DirtyTracker, parentBit int64, factsAccessor xmap.FactsAccessor[K], key K)
	// Unlink 断开链接
	Unlink()
	// GetDirtyTracker 获取DirtyTracker（用于子对象的Link）
	GetDirtyTracker() *dt.DirtyTracker
}

// WrapperFactory 包装器工厂函数类型
// 用于将protobuf对象转换为包装对象
type WrapperFactory[PbValue any, WrapperValue any] func(pb PbValue) WrapperValue

// XMapLink 通用的xmap链接管理器
// K: map的key类型
// PbValue: protobuf对象类型（通常是指针）
// WrapperValue: 包装对象类型（实现了Linkable接口）
type XMapLink[K comparable, PbValue any, WrapperValue Linkable[K]] struct {
	pbMap          *map[K]PbValue                        // protobuf map的引用
	wrapperMap     map[K]WrapperValue                    // 包装对象map
	mapAccessor    xmap.MapAccessor[K, PbValue]          // xmap访问器
	wrapperFactory WrapperFactory[PbValue, WrapperValue] // 包装器工厂函数
	parentTracker  *dt.DirtyTracker                      // 父DirtyTracker
	parentBit      int64                                 // 父脏标记位
}

// NewXMapLink 创建一个新的XMapLink实例
// pbMap: protobuf map的引用
// marker: 脏标记接口
// dirtyBit: 脏标记位
// wrapperFactory: 包装器工厂函数
// isRefType: 是否为引用类型（如果是引用类型，Set时会使用TrackSetWithDelete）
func NewXMapLink[K comparable, PbValue any, WrapperValue Linkable[K]](
	pbMap *map[K]PbValue,
	marker xmap.DirtyMarker,
	dirtyBit int64,
	wrapperFactory WrapperFactory[PbValue, WrapperValue],
	isRefType bool,
) *XMapLink[K, PbValue, WrapperValue] {

	link := &XMapLink[K, PbValue, WrapperValue]{
		pbMap:          pbMap,
		wrapperMap:     make(map[K]WrapperValue),
		wrapperFactory: wrapperFactory,
	}

	// 创建MapAccessor
	if isRefType {
		link.mapAccessor = xmap.NewMapAccessorWithMarkerForRef(pbMap, marker, dirtyBit)
	} else {
		link.mapAccessor = xmap.NewMapAccessorWithMarker(pbMap, marker, dirtyBit)
	}

	// 初始化现有的map元素
	if pbMap != nil && *pbMap != nil {
		for key, pbValue := range *pbMap {
			wrapper := wrapperFactory(pbValue)
			// 使用LinkFactsAccessor链接到父对象和Key
			wrapper.LinkFactsAccessor(link.parentTracker, link.parentBit, &link.mapAccessor, key)
			link.wrapperMap[key] = wrapper
		}
	}

	return link
}

// SetParent 设置父DirtyTracker（必须在初始化完成后调用）
func (x *XMapLink[K, PbValue, WrapperValue]) SetParent(parentTracker *dt.DirtyTracker, parentBit int64) {
	x.parentTracker = parentTracker
	x.parentBit = parentBit

	// 重新链接所有已存在的包装对象
	for key, wrapper := range x.wrapperMap {
		wrapper.Unlink()
		wrapper.LinkFactsAccessor(x.parentTracker, x.parentBit, &x.mapAccessor, key)
	}
}

// Get 获取包装对象
// 返回包装对象和是否存在的标志
func (x *XMapLink[K, PbValue, WrapperValue]) Get(key K) (WrapperValue, bool) {
	wrapper, ok := x.wrapperMap[key]
	return wrapper, ok
}

// Set 设置/添加对象
// 如果key已存在，会先Unlink旧对象
// 返回新的包装对象
func (x *XMapLink[K, PbValue, WrapperValue]) Set(key K, pbValue PbValue) WrapperValue {
	// 处理旧对象的Unlink
	x.Delete(key)

	// 创建新的包装对象
	wrapper := x.wrapperFactory(pbValue)

	// Link到父对象和xmap的FactsAccessor
	wrapper.LinkFactsAccessor(x.parentTracker, x.parentBit, &x.mapAccessor, key)

	// 更新protobuf map（通过MapAccessor，会自动处理TrackSetWithDelete）
	x.mapAccessor.Set(key, pbValue)

	// 更新包装对象map
	x.wrapperMap[key] = wrapper

	return wrapper
}

// Delete 删除对象
// 返回是否成功删除
func (x *XMapLink[K, PbValue, WrapperValue]) Delete(key K) bool {
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

// Has 检查key是否存在
func (x *XMapLink[K, PbValue, WrapperValue]) Has(key K) bool {
	_, ok := x.wrapperMap[key]
	return ok
}

// Len 返回map的长度
func (x *XMapLink[K, PbValue, WrapperValue]) Len() int {
	return len(x.wrapperMap)
}

// Range 遍历所有对象
// 回调函数返回false时停止遍历
func (x *XMapLink[K, PbValue, WrapperValue]) Range(f func(key K, wrapper WrapperValue) bool) {
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
func (x *XMapLink[K, PbValue, WrapperValue]) Clear() {
	// Unlink所有包装对象
	for _, wrapper := range x.wrapperMap {
		wrapper.Unlink()
	}

	// 清空包装对象map
	x.wrapperMap = make(map[K]WrapperValue)

	// 清空protobuf map（通过MapAccessor）
	x.mapAccessor.Clear()
}

// GetMapAccessor 获取MapAccessor（用于高级操作）
func (x *XMapLink[K, PbValue, WrapperValue]) GetMapAccessor() *xmap.MapAccessor[K, PbValue] {
	return &x.mapAccessor
}

// Keys 返回所有key
func (x *XMapLink[K, PbValue, WrapperValue]) Keys() []K {
	keys := make([]K, 0, len(x.wrapperMap))
	for key := range x.wrapperMap {
		keys = append(keys, key)
	}
	return keys
}

// Values 返回所有包装对象
func (x *XMapLink[K, PbValue, WrapperValue]) Values() []WrapperValue {
	values := make([]WrapperValue, 0, len(x.wrapperMap))
	for _, wrapper := range x.wrapperMap {
		values = append(values, wrapper)
	}
	return values
}
