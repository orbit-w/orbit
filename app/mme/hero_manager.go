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

// GetHero 获取英雄模块（推荐方法）
func (m *HeroManager[FactsAccessorKey]) GetHero(id int64) (*HeroModule[int64], bool) {
	hero, ok := m.heroMapLink.Get(id)
	return hero, ok
}

// SetHero 设置/添加英雄模块
func (m *HeroManager[FactsAccessorKey]) SetHero(id int64, heroModulePb *mme.HeroModule) *HeroModule[int64] {
	// 先处理旧对象的Unlink
	if oldHero, ok := m.heroMapLink.Get(id); ok {
		oldHero.Unlink()
	}

	// 更新protobuf map
	hm := m.heroMapLink.Set(id, heroModulePb)
	return hm
}

// DeleteHero 删除英雄模块
func (m *HeroManager[FactsAccessorKey]) DeleteHero(id int64) bool {
	return m.heroMapLink.Delete(id)
}

// RangeHeroes 遍历所有英雄
func (m *HeroManager[FactsAccessorKey]) RangeHeroes(f func(id int64, hero *HeroModule[int64]) bool) {
	m.heroMapLink.Range(func(id int64, hero *HeroModule[int64]) bool {
		return f(id, hero)
	})
}
