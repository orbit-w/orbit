package mme

import (
	"gitee.com/orbit-w/meteor/bases/dirty/xmap"
	"gitee.com/orbit-w/meteor/modules/mlog"
	"gitee.com/orbit-w/orbit/app/proto/mme"
	dirtyflag "gitee.com/orbit-w/orbit/lib/base/dirty_flag"
	"gitee.com/orbit-w/orbit/lib/base/xmapwrapper"
	"github.com/gogo/protobuf/proto"
	"go.uber.org/zap"
)

const (
	HeroManagerDirtyHeroMapBit = 1 << 0
)

type HeroManager struct {
	heroManager *mme.HeroManager
	dirtyflag.IDirtyFlag

	heroMapLink *xmapwrapper.XMapWrapper[int64, *mme.HeroModule, *HeroModule]
}

func NewHeroManager(pt *mme.HeroManager) *HeroManager {
	if pt == nil {
		panic("pt is nil")
	}

	m := &HeroManager{
		heroManager: pt,
		IDirtyFlag:  dirtyflag.NewDirtyFlag(),
	}

	m.heroMapLink = xmapwrapper.NewXMapWrapperWithParent(
		&m.heroManager.HeroMap, m.GetDirtyTracker(), HeroManagerDirtyHeroMapBit, NewHeroModule)

	return m
}

// SetHero 设置/添加英雄模块
func (m *HeroManager) HeroMap_Set(id int64, heroModulePb *mme.HeroModule) *HeroModule {
	return m.heroMapLink.Set(id, heroModulePb)
}

// DeleteHero 删除英雄模块
func (m *HeroManager) HeroMap_Delete(id int64) bool {
	return m.heroMapLink.Delete(id)
}

// RangeHeroes 遍历所有英雄
func (m *HeroManager) HeroMap_Range(f func(id int64, hero *HeroModule) bool) {
	m.heroMapLink.Range(func(id int64, hero *HeroModule) bool {
		return f(id, hero)
	})
}

func (m *HeroManager) ClearAllDirty() {
	m.IDirtyFlag.ClearAllDirty()

	// 清空xmaplink中所有object的脏标记
	m.heroMapLink.RangeIncrementalSyncObject(func(key int64, object xmapwrapper.IncrementalSyncObject) (stop bool) {
		object.ClearAllDirty()
		return false
	})
}

// ToIncrementalProto 根据脏标记位构建增量数据的 protoMessage
// 只返回标记为脏的字段数据，用于增量同步
func (m *HeroManager) ToIncrementalProto() proto.Message {
	if m == nil {
		return nil
	}

	incremental := &mme.HeroManager{}

	if m.IsDirty(HeroManagerDirtyHeroMapBit) {
		incremental.HeroMap_XXXChangeList = make([]*mme.HeroManager_HeroMap_XXXMapChangeRecord, 0)
		m.heroMapLink.RangeOperations(func(key int64, operation xmap.MapOperation[int64]) bool {
			switch operation.Type {
			case xmap.SetOperation:
				hero, _ := m.heroMapLink.Get(key)
				pb := hero.ToIncrementalProto()
				v, ok := pb.(*mme.HeroModule)
				if ok {
					incremental.HeroMap_XXXChangeList = append(incremental.HeroMap_XXXChangeList, &mme.HeroManager_HeroMap_XXXMapChangeRecord{
						Key:   key,
						Value: v,
					})
				} else {
					mlog.Error("HeroManager.ToIncrementalProto", zap.Any("key", key), zap.Any("operation", operation))
				}
				return true
			case xmap.DeleteOperation:
				incremental.HeroMap_XXXChangeList = append(incremental.HeroMap_XXXChangeList, &mme.HeroManager_HeroMap_XXXMapChangeRecord{
					Key:      key,
					IsDelete: true,
				})
				return true
			}
			return true
		})

	}
	return incremental
}
