package dirtyflag

import (
	dt "gitee.com/orbit-w/meteor/bases/dirty/dirty_tracker"
	"gitee.com/orbit-w/meteor/bases/dirty/xmap"
)

type Linkable interface {
	Link(parent *dt.DirtyTracker, parentBit int64)
	LinkFactsAccessor(parent *dt.DirtyTracker, parentBit int64, factsAccessor xmap.FactsAccessor)
}

type IFactsAccessorLinker interface {
}

type IDirtyFlag interface {
	Linkable
	MarkDirty(dirtyBit int64)
	IsDirty(dirtyBit int64) bool
	HasAnyDirty() bool
	ClearAllDirty()
	GetDirtyTracker() *dt.DirtyTracker
	GetFactsAccessor() xmap.FactsAccessor
	Unlink()
}

type DirtyFlag struct {
	dt.DirtyTracker
	factsAccessor xmap.FactsAccessor
}

func NewDirtyFlag() *DirtyFlag {
	return &DirtyFlag{}
}

func (d *DirtyFlag) Link(parent *dt.DirtyTracker, parentBit int64) {
	d.DirtyTracker.Link(parent, parentBit)
}

func (d *DirtyFlag) LinkFactsAccessor(parent *dt.DirtyTracker, parentBit int64,
	factsAccessor xmap.FactsAccessor) {
	if factsAccessor == nil {
		panic("factsAccessor is nil")
	}
	d.factsAccessor = factsAccessor
	d.DirtyTracker.Link(parent, parentBit)
}

func (d *DirtyFlag) Unlink() {
	d.DirtyTracker.Unlink()

	d.factsAccessor = nil
}

func (d *DirtyFlag) GetFactsAccessor() xmap.FactsAccessor {
	return d.factsAccessor
}

func (d *DirtyFlag) GetDirtyTracker() *dt.DirtyTracker {
	return &d.DirtyTracker
}

func (d *DirtyFlag) MarkDirty(dirtyBit int64) {
	d.DirtyTracker.MarkDirty(dirtyBit)
	if d.factsAccessor != nil {
		d.factsAccessor.TrackSet()
	}
}

func (d *DirtyFlag) IsDirty(dirtyBit int64) bool {
	return d.DirtyTracker.IsDirty(dirtyBit)
}

func (d *DirtyFlag) HasAnyDirty() bool {
	return d.DirtyTracker.HasAnyDirty()
}
