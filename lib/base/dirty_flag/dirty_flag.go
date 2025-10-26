package dirtyflag

import (
	dt "gitee.com/orbit-w/meteor/bases/dirty/dirty_tracker"
)

type Linkable interface {
	Link(parent *dt.DirtyTracker, parentBit int64)
	LinkFactsAccessor(parent *dt.DirtyTracker, parentBit int64, trackSet func())
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
	Unlink()
}

type DirtyFlag struct {
	dt.DirtyTracker
	trackSet func()
}

func NewDirtyFlag() *DirtyFlag {
	return &DirtyFlag{}
}

func (d *DirtyFlag) Link(parent *dt.DirtyTracker, parentBit int64) {
	d.DirtyTracker.Link(parent, parentBit)
}

func (d *DirtyFlag) LinkFactsAccessor(parent *dt.DirtyTracker, parentBit int64,
	trackSet func()) {
	if trackSet == nil {
		panic("factsAccessor is nil")
	}
	d.trackSet = trackSet
	d.DirtyTracker.Link(parent, parentBit)
}

func (d *DirtyFlag) Unlink() {
	d.DirtyTracker.Unlink()

	d.trackSet = nil
}

func (d *DirtyFlag) GetDirtyTracker() *dt.DirtyTracker {
	return &d.DirtyTracker
}

func (d *DirtyFlag) MarkDirty(dirtyBit int64) {
	d.DirtyTracker.MarkDirty(dirtyBit)
	if d.trackSet != nil {
		d.trackSet()
	}
}

func (d *DirtyFlag) IsDirty(dirtyBit int64) bool {
	return d.DirtyTracker.IsDirty(dirtyBit)
}

func (d *DirtyFlag) HasAnyDirty() bool {
	return d.DirtyTracker.HasAnyDirty()
}
