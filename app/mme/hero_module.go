package mme

import (
	"gitee.com/orbit-w/meteor/bases/dirty/dirty_tracker"
	"gitee.com/orbit-w/orbit/app/proto/mme"
)

// Dirty bits for Module fields
const (
	HeroModuleDirtyXXXIdBit   int64 = 1 << 0
	HeroModuleDirtyBaseBit    int64 = 1 << 1
	HeroModuleDirtyLevelUpBit int64 = 1 << 2
)

type HeroModule struct {
	*mme.HeroModule
	dirty_tracker.DirtyTracker
}

func NewHeroModule() *HeroModule {
	return &HeroModule{
		HeroModule: new(mme.HeroModule),
	}
}

func (m *HeroModule) Name() string {
	return "HeroModule"
}

func (m *HeroModule) Link(parent *dirty_tracker.DirtyTracker, parentBit int64) {
	m.DirtyTracker.Link(parent, parentBit)
}

func (m *HeroModule) GetXXXId() int64 {
	if m != nil {
		return m.XXXId
	}
	return 0
}

func (m *HeroModule) SetXXXId(v int64) {
	m.XXXId = v
	m.MarkDirty(HeroModuleDirtyXXXIdBit)
}
