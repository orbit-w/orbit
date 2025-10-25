package mme

import (
	"gitee.com/orbit-w/orbit/app/proto/mme"
	dirtyflag "gitee.com/orbit-w/orbit/lib/base/dirty_flag"
)

// Dirty bits for Module fields
const (
	HeroModuleDirtyXXXIdBit   int64 = 1 << 0
	HeroModuleDirtyBaseBit    int64 = 1 << 1
	HeroModuleDirtyLevelUpBit int64 = 1 << 2
)

type HeroModule[FactsAccessorKey comparable] struct {
	mme *mme.HeroModule
	dirtyflag.IDirtyFlag[FactsAccessorKey]

	Base    *HeroMechanism[any]
	LevelUp *LevelUpMechanism[any]
}

func NewHeroModule[FactsAccessorKey comparable](pt *mme.HeroModule) *HeroModule[FactsAccessorKey] {
	if pt == nil {
		panic("pt is nil")
	}
	m := &HeroModule[FactsAccessorKey]{
		mme:        pt,
		IDirtyFlag: dirtyflag.NewDirtyFlag[FactsAccessorKey](),
	}
	// Mechanism obj 需要Link到DirtyTracker
	// xmap类型，
	// if Value 是Mechanism，则需要通过LinkFactsAccessor到Father和xmap的Key

	m.Base = NewHeroMechanism[any](m.mme.Base)
	m.Base.Link(m.GetDirtyTracker(), HeroModuleDirtyBaseBit)

	m.LevelUp = NewLevelUpMechanism[any](m.mme.LevelUp)
	m.LevelUp.Link(m.GetDirtyTracker(), HeroModuleDirtyLevelUpBit)
	return m
}

func (m *HeroModule[FactsAccessorKey]) Name() string {
	return "HeroModule"
}

func (m *HeroModule[FactsAccessorKey]) DeepCopy(co *mme.HeroModule) {
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
