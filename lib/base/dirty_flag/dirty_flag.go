package dirtyflag

import (
	dt "gitee.com/orbit-w/meteor/bases/dirty/dirty_tracker"
	"gitee.com/orbit-w/meteor/bases/dirty/xmap"
)

type Linkable[FactsAccessorKey comparable] interface {
	Link(parent *dt.DirtyTracker, parentBit int64)
	LinkFactsAccessor(parent *dt.DirtyTracker, parentBit int64, factsAccessor xmap.FactsAccessor[FactsAccessorKey], key FactsAccessorKey)
}

type IDirtyFlag[FactsAccessorKey comparable] interface {
	Linkable[FactsAccessorKey]
	MarkDirty(dirtyBit int64)
	IsDirty(dirtyBit int64) bool
	HasAnyDirty() bool
	ClearAllDirty()
	GetDirtyTracker() *dt.DirtyTracker
	GetFactsAccessor() xmap.FactsAccessor[FactsAccessorKey]
	GetFactsAccessorKey() FactsAccessorKey
	Unlink()
}

type DirtyFlag[FactsAccessorKey comparable] struct {
	dt.DirtyTracker
	factsAccessorKey FactsAccessorKey
	factsAccessor    xmap.FactsAccessor[FactsAccessorKey]
}

func NewDirtyFlag[FactsAccessorKey comparable]() *DirtyFlag[FactsAccessorKey] {
	return &DirtyFlag[FactsAccessorKey]{}
}

func (d *DirtyFlag[FactsAccessorKey]) Link(parent *dt.DirtyTracker, parentBit int64) {
	d.DirtyTracker.Link(parent, parentBit)
}

func (d *DirtyFlag[FactsAccessorKey]) LinkFactsAccessor(parent *dt.DirtyTracker, parentBit int64,
	factsAccessor xmap.FactsAccessor[FactsAccessorKey], key FactsAccessorKey) {
	if factsAccessor == nil {
		panic("factsAccessor is nil")
	}
	d.factsAccessor = factsAccessor
	d.factsAccessorKey = key
	d.DirtyTracker.Link(parent, parentBit)
}

func (d *DirtyFlag[FactsAccessorKey]) Unlink() {
	d.DirtyTracker.Unlink()

	d.factsAccessor = nil
	var zeroValue FactsAccessorKey
	d.factsAccessorKey = zeroValue
}

func (d *DirtyFlag[FactsAccessorKey]) GetFactsAccessor() xmap.FactsAccessor[FactsAccessorKey] {
	return d.factsAccessor
}

func (d *DirtyFlag[FactsAccessorKey]) GetFactsAccessorKey() FactsAccessorKey {
	return d.factsAccessorKey
}

func (d *DirtyFlag[FactsAccessorKey]) GetDirtyTracker() *dt.DirtyTracker {
	return &d.DirtyTracker
}

func (d *DirtyFlag[FactsAccessorKey]) MarkDirty(dirtyBit int64) {
	d.DirtyTracker.MarkDirty(dirtyBit)
	if d.factsAccessor != nil {
		d.factsAccessor.TrackSet(d.factsAccessorKey)
	}
}

func (d *DirtyFlag[FactsAccessorKey]) IsDirty(dirtyBit int64) bool {
	return d.DirtyTracker.IsDirty(dirtyBit)
}

func (d *DirtyFlag[FactsAccessorKey]) HasAnyDirty() bool {
	return d.DirtyTracker.HasAnyDirty()
}
