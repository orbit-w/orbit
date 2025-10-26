package mme

import (
	"gitee.com/orbit-w/orbit/app/proto/mme"
	dirtyflag "gitee.com/orbit-w/orbit/lib/base/dirty_flag"
	xmaplink "gitee.com/orbit-w/orbit/lib/base/xmap_link"
)

const (
	HeroManagerDirtyHeroMapBit = 1 << 0
)

type HeroManager[FactsAccessorKey comparable] struct {
	heroManager *mme.HeroManager
	dirtyflag.IDirtyFlag[FactsAccessorKey]

	heroMapLink *xmaplink.XMapLink[int64, *mme.HeroModule, *HeroModule[int64]]
}

func NewHeroManager[FactsAccessorKey comparable](pt *mme.HeroManager) *HeroManager[FactsAccessorKey] {
	if pt == nil {
		panic("pt is nil")
	}

	m := &HeroManager[FactsAccessorKey]{
		heroManager: pt,
		IDirtyFlag:  dirtyflag.NewDirtyFlag[FactsAccessorKey](),
	}

	m.heroMapLink = xmaplink.NewXMapLinkWithParent(
		&m.heroManager.HeroMap, m.GetDirtyTracker(), HeroManagerDirtyHeroMapBit, NewHeroModule[int64], true)

	return m
}

// SetHero 设置/添加英雄模块
func (m *HeroManager[FactsAccessorKey]) HeroMap_Set(id int64, heroModulePb *mme.HeroModule) *HeroModule[int64] {
	return m.heroMapLink.Set(id, heroModulePb)
}

// DeleteHero 删除英雄模块
func (m *HeroManager[FactsAccessorKey]) HeroMap_Delete(id int64) bool {
	return m.heroMapLink.Delete(id)
}

// RangeHeroes 遍历所有英雄
func (m *HeroManager[FactsAccessorKey]) HeroMap_Range(f func(id int64, hero *HeroModule[int64]) bool) {
	m.heroMapLink.Range(func(id int64, hero *HeroModule[int64]) bool {
		return f(id, hero)
	})
}
