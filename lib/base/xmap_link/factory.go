package xmaplink

import (
	dt "gitee.com/orbit-w/meteor/bases/dirty/dirty_tracker"
)

// NewXMapLinkWithParent 创建XMapLink并立即设置父节点
// 这是最常用的构造方式
func NewXMapLinkWithParent[K comparable, PbValue any, WrapperValue Linkable](
	pbMap *map[K]PbValue,
	parentTracker *dt.DirtyTracker,
	parentBit int64,
	wrapperFactory WrapperFactory[PbValue, WrapperValue],
) *XMapLink[K, PbValue, WrapperValue] {
	// 创建一个临时的DirtyMarker
	marker := &tempDirtyMarker{
		tracker: parentTracker,
	}

	link := NewXMapLink(pbMap, marker, parentBit, wrapperFactory)
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
