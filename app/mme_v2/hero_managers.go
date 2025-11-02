package mme

import (
	"gitee.com/orbit-w/meteor/bases/container/xmap"
	"gitee.com/orbit-w/meteor/modules/mlog"
	"gitee.com/orbit-w/orbit/app/proto/mme"
	dirtyflag "gitee.com/orbit-w/orbit/lib/base/dirty_flag"
	fieldmeta "gitee.com/orbit-w/orbit/lib/base/field_meta"
	"gitee.com/orbit-w/orbit/lib/module/db/mgo_builder"
	mmemodel "gitee.com/orbit-w/orbit/lib/module/mme_model"
	xmapwrapper "gitee.com/orbit-w/orbit/lib/module/xmapwrapper"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

const (
	HeroManagerFieldIndexHeroMap = uint8(0)
)

// Dirty bits for HeroManager fields
const (
	HeroManagerDirtyHeroMapBit int64 = 1 << HeroManagerFieldIndexHeroMap
)

type HeroManager struct {
	HeroMap map[int64]*HeroModule `bson:"hero_map"`
}

func NewHeroManager() *HeroManager {
	return &HeroManager{
		HeroMap: make(map[int64]*HeroModule),
	}
}

// 数据-深拷贝
func (m *HeroManager) DeepCopy(co *HeroManager) {
	if m == nil || co == nil {
		return
	}

	*co = *m
	if m.HeroMap != nil {
		co.HeroMap = make(map[int64]*HeroModule, len(m.HeroMap))
		for k, v := range m.HeroMap {
			co.HeroMap[k] = &HeroModule{}
			v.DeepCopy(co.HeroMap[k])
		}
	}
}

// 数据-转换为protobuf
func (m *HeroManager) ToProto() *mme.HeroManager {
	if m == nil {
		return nil
	}

	pb := &mme.HeroManager{}
	if m.HeroMap != nil {
		pb.HeroMap = make(map[int64]*mme.HeroModule, len(m.HeroMap))
		for k, v := range m.HeroMap {
			pb.HeroMap[k] = v.ToProto()
		}
	}
	return pb
}

type HeroManagerWrapper struct {
	data *HeroManager
	dirtyflag.IDirtyFlag
	fieldMetas *fieldmeta.FieldMetas

	//Value为MME Object，使用xmap.MapAccessor进行包装
	heroMapLink *xmapwrapper.XMapWrapper[int64, *HeroModule, *HeroModuleWrapper]
}

func NewHeroManagerWrapper(data *HeroManager) *HeroManagerWrapper {
	if data == nil {
		panic("data is nil")
	}

	hm := &HeroManagerWrapper{
		data:       data,
		IDirtyFlag: dirtyflag.NewDirtyFlag(),
		fieldMetas: fieldmeta.NewFieldMetas(),
	}
	hm.heroMapLink = xmapwrapper.NewXMapWrapperWithParent(
		&data.HeroMap,
		hm.GetDirtyTracker(),
		HeroManagerDirtyHeroMapBit,
		NewHeroModuleWrapper,
	)

	return hm
}

func (m *HeroManagerWrapper) InitFieldContext() {
	m.fieldMetas.SetFieldType(HeroManagerFieldIndexHeroMap, fieldmeta.FieldTypeSync)

	// 初始化所有 HeroModule 的字段上下文
	m.heroMapLink.Range(func(key int64, wrapper *HeroModuleWrapper) bool {
		wrapper.InitFieldContext()
		return true
	})
}

// Manager 唯一名称
func (m *HeroManagerWrapper) Name() string {
	return "HeroManager"
}

// SetHero 设置/添加英雄模块
func (m *HeroManagerWrapper) HeroMap_Set(id int64, heroModulePb *HeroModule) *HeroModuleWrapper {
	return m.heroMapLink.Set(id, heroModulePb)
}

// DeleteHero 删除英雄模块
func (m *HeroManagerWrapper) HeroMap_Delete(id int64) bool {
	return m.heroMapLink.Delete(id)
}

// RangeHeroes 遍历所有英雄
func (m *HeroManagerWrapper) HeroMap_Range(f func(id int64, hero *HeroModuleWrapper) bool) {
	m.heroMapLink.Range(f)
}

func (m *HeroManagerWrapper) ClearAllDirty() {
	m.IDirtyFlag.ClearAllDirty()

	// 清空xmaplink中所有object的脏标记
	m.heroMapLink.RangeIncrementalSyncObject(func(key int64, object xmapwrapper.IncrementalSyncObject) (stop bool) {
		object.ClearAllDirty()
		return false
	})
}

// BuildMongoUpdate 构建MongoDB更新操作
// 注意：Manager 层级通常不直接构建 MongoDB 更新，而是由 Entity 层处理
func (m *HeroManagerWrapper) BuildMongoUpdate(builder *mgo_builder.MongoUpdateBuilder, path *mgo_builder.NestedPath) {
	if m == nil {
		return
	}

	if m.IsDirty(HeroManagerDirtyHeroMapBit) {
		//map 结构无法做增量更新，所以需要全量拷贝
		copy := make(map[int64]*HeroModule, m.heroMapLink.Len())
		m.heroMapLink.DeepCopy(&copy)
		builder.SetNestedPath(path, "hero_map", copy)
	}
}

// 包装器-深拷贝
func (m *HeroManagerWrapper) DeepCopy() *HeroManager {
	if m == nil {
		return nil
	}
	copy := &HeroManager{}
	m.DeepCopyTo(copy)
	return copy
}

// 包装器-深拷贝
func (m *HeroManagerWrapper) DeepCopyTo(copy *HeroManager) {
	if m == nil || m.data == nil || copy == nil {
		return
	}

	// 初始化目标 map
	if copy.HeroMap == nil {
		copy.HeroMap = make(map[int64]*HeroModule, len(m.data.HeroMap))
	}

	// 拷贝所有 HeroModule
	m.heroMapLink.DeepCopy(&copy.HeroMap)
}

// MatchesAll 判断字段是否匹配所有类型标记
func (m *HeroManagerWrapper) MatchesAll(fieldID uint8, fieldTypes ...fieldmeta.FieldType) bool {
	return m.fieldMetas.MatchesAll(fieldID, fieldTypes...)
}

// ToProto 将 HeroManager 数据转换为完整的 protobuf 结构体
func (m *HeroManagerWrapper) ToProto() *mme.HeroManager {
	if m == nil || m.data == nil {
		return nil
	}

	return m.data.ToProto()
}

// FromProto 从 protobuf 结构体加载数据到 HeroManager
func (m *HeroManagerWrapper) FromProto(pb *mme.HeroManager) {
	if m == nil || m.data == nil || pb == nil {
		return
	}

	// 重新构建脏标系统
	m.IDirtyFlag = dirtyflag.NewDirtyFlag()

	// xmap全量覆盖数据
	data := NewHeroManager()
	m.data = data
	m.heroMapLink.Reset(&m.data.HeroMap)

	for key := range pb.HeroMap {
		pbValue := pb.HeroMap[key]
		v := NewHeroModule()
		wrapper := m.heroMapLink.SetWithoutTrack(key, v)
		wrapper.FromProto(pbValue)
	}

}

// ToIncrementalProto 根据脏标记位构建增量数据的 protoMessage
// 只返回标记为脏的字段数据，用于增量同步
func (m *HeroManagerWrapper) ToIncrementalProto(ctx mmemodel.SyncContext) proto.Message {
	if m == nil {
		return nil
	}

	// 如果没有脏标记，返回 nil
	if !m.HasAnyDirty() {
		return nil
	}

	incremental := &mme.HeroManager{}
	if mmemodel.FieldCanBeIncrementalSynced(m, HeroManagerDirtyHeroMapBit, HeroManagerFieldIndexHeroMap, ctx) {
		incremental.HeroMap_XXXChangeList = make([]*mme.HeroManager_HeroMap_XXXMapChangeRecord, 0)
		m.heroMapLink.RangeOperations(func(key int64, operation xmap.MapOperation[int64]) bool {
			switch operation.Type {
			case xmap.SetOperation:
				hero, _ := m.heroMapLink.Get(key)
				pb := hero.ToIncrementalProtoWithContext(ctx)
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
