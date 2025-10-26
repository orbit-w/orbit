package xmaplink

import (
	dt "gitee.com/orbit-w/meteor/bases/dirty/dirty_tracker"
	"gitee.com/orbit-w/meteor/bases/dirty/xmap"
)

// NewXMapLinkForRef 创建引用类型的XMapLink（自动使用TrackSetWithDelete）
// 适用于Value是指针类型的场景（如 *mme.HeroModule）
func NewXMapLinkForRef[K comparable, PbValue any, WrapperValue Linkable](
	pbMap *map[K]PbValue,
	marker xmap.DirtyMarker,
	dirtyBit int64,
	wrapperFactory WrapperFactory[PbValue, WrapperValue],
) *XMapLink[K, PbValue, WrapperValue] {
	return NewXMapLink(pbMap, marker, dirtyBit, wrapperFactory, true)
}

// NewXMapLinkForValue 创建值类型的XMapLink
// 适用于Value是值类型的场景（如 int32, string）
func NewXMapLinkForValue[K comparable, PbValue any, WrapperValue Linkable](
	pbMap *map[K]PbValue,
	marker xmap.DirtyMarker,
	dirtyBit int64,
	wrapperFactory WrapperFactory[PbValue, WrapperValue],
) *XMapLink[K, PbValue, WrapperValue] {
	return NewXMapLink(pbMap, marker, dirtyBit, wrapperFactory, false)
}

// NewXMapLinkWithParent 创建XMapLink并立即设置父节点
// 这是最常用的构造方式
func NewXMapLinkWithParent[K comparable, PbValue any, WrapperValue Linkable](
	pbMap *map[K]PbValue,
	parentTracker *dt.DirtyTracker,
	parentBit int64,
	wrapperFactory WrapperFactory[PbValue, WrapperValue],
	isRefType bool,
) *XMapLink[K, PbValue, WrapperValue] {
	// 创建一个临时的DirtyMarker
	marker := &tempDirtyMarker{
		tracker: parentTracker,
	}

	link := NewXMapLink(pbMap, marker, parentBit, wrapperFactory, isRefType)
	link.SetParent(parentTracker, parentBit)

	return link
}

// tempDirtyMarker 临时的DirtyMarker实现，用于初始化
type tempDirtyMarker struct {
	tracker *dt.DirtyTracker
}

func (m *tempDirtyMarker) MarkDirty(bit int64) {
	if m.tracker != nil {
		m.tracker.MarkDirty(bit)
	}
}
