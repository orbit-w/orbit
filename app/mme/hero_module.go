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
	mme *mme.HeroModule
	dirty_tracker.DirtyTracker

	Base    *HeroMechanism[int32]
	LevelUp *LevelUpMechanism[int32]
}

func NewHeroModule(pt *mme.HeroModule) *HeroModule {
	if pt == nil {
		panic("pt is nil")
	}
	m := &HeroModule{
		mme: pt,
	}
	// 非
	m.Base = NewHeroMechanism[int32](m.mme.Base)
	m.Base.Link(&m.DirtyTracker, HeroModuleDirtyBaseBit)

	m.LevelUp = NewLevelUpMechanism[int32](m.mme.LevelUp)
	m.LevelUp.Link(&m.DirtyTracker, HeroModuleDirtyLevelUpBit)
	return m
}

func (m *HeroModule) Name() string {
	return "HeroModule"
}

func (m *HeroModule) Link(parent *dirty_tracker.DirtyTracker, parentBit int64) {
	m.DirtyTracker.Link(parent, parentBit)
}

func (m *HeroModule) DeepCopy(co *mme.HeroModule) {
	if m == nil || co == nil {
		return
	}

	if m.Base != nil {
		m.Base.DeepCopy(co.Base)
	}

	if m.LevelUp != nil {
		m.LevelUp.DeepCopy(co.LevelUp)
	}
}
