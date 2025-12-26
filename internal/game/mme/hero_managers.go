package mme

import (
	"gitee.com/orbit-w/meteor/bases/container/xmap"
	dirtyflag "gitee.com/orbit-w/orbit/lib/base/dirty_flag"
	fieldmeta "gitee.com/orbit-w/orbit/lib/base/field_meta"
	"gitee.com/orbit-w/orbit/lib/module/db/mgo_builder"
	mmemodel "gitee.com/orbit-w/orbit/lib/module/mme_model"
	xmapwrapper "gitee.com/orbit-w/orbit/lib/module/xmapwrapper"
	"gitee.com/orbit-w/orbit/pkg/proto/mme"
	"google.golang.org/protobuf/proto"
)

const (
	HeroManagerFieldIndexHeroMap          = uint8(1)
	HeroManagerFieldIndexSingleHeroModule = uint8(2)
)

// Dirty bits for HeroManager fields
const (
	HeroManagerDirtyHeroMapBit          int64 = 1 << (HeroManagerFieldIndexHeroMap - 1)
	HeroManagerDirtySingleHeroModuleBit int64 = 1 << (HeroManagerFieldIndexSingleHeroModule - 1)
)

type HeroManager struct {
	HeroMap          map[int64]*HeroModule `bson:"hero_map"`
	SingleHeroModule *HeroModule           `bson:"single_hero_module"`
}

func NewHeroManager() *HeroManager {
	return &HeroManager{
		HeroMap: make(map[int64]*HeroModule),
	}
}

func (m *HeroManager) DeepCopy(co *HeroManager) {
	if m == nil || co == nil {
		return
	}

	*co = *m
	if m.HeroMap != nil {
		co.HeroMap = make(map[int64]*HeroModule, len(m.HeroMap))
		for k, v := range m.HeroMap {
			if v != nil {
				co.HeroMap[k] = &HeroModule{}
				v.DeepCopy(co.HeroMap[k])
			}
		}
	}
}

func (m *HeroManager) ToProto() *mme.HeroManager {
	if m == nil {
		return nil
	}

	pb := &mme.HeroManager{}
	if m.HeroMap != nil {
		pb.HeroMap = make(map[int64]*mme.HeroModule, len(m.HeroMap))
		for k, v := range m.HeroMap {
			if v != nil {
				pb.HeroMap[k] = v.ToProto()
			}
		}
	}
	return pb
}

func (m *HeroManager) FromProto(pb *mme.HeroManager) {
	if m == nil || pb == nil {
		return
	}

	if pb.HeroMap != nil {
		m.HeroMap = make(map[int64]*HeroModule, len(pb.HeroMap))
		for k, v := range pb.HeroMap {
			if v != nil {
				obj := NewHeroModule()
				obj.FromProto(v)
				m.HeroMap[k] = obj
			}
		}
	}
}

type HeroManagerWrapper struct {
	data *HeroManager
	dirtyflag.IDirtyFlag
	fieldMetas *fieldmeta.FieldMetas

	//Value为MME Object，使用xmap.MapAccessor进行包装
	heroMapLink             *xmapwrapper.XMapWrapper[int64, *HeroModule, *HeroModuleWrapper]
	SingleHeroModuleWrapper *HeroModuleWrapper
}

func NewHeroManagerWrapper(data *HeroManager) *HeroManagerWrapper {
	if data == nil {
		panic("data is nil")
	}
	m := &HeroManagerWrapper{
		data:       data,
		IDirtyFlag: dirtyflag.NewDirtyFlag(),
		fieldMetas: fieldmeta.NewFieldMetas(),
	}

	m.heroMapLink = xmapwrapper.NewXMapWrapperWithParent(
		&data.HeroMap,
		m.GetDirtyTracker(),
		HeroManagerDirtyHeroMapBit,
		NewHeroModuleWrapper,
	)

	if data.SingleHeroModule == nil {
		data.SingleHeroModule = NewHeroModule()
	}
	m.SingleHeroModuleWrapper = NewHeroModuleWrapper(data.SingleHeroModule)
	m.SingleHeroModuleWrapper.Link(m.GetDirtyTracker(), HeroManagerDirtySingleHeroModuleBit)
	return m
}

func (m *HeroManagerWrapper) InitFieldContext() {
	m.fieldMetas.SetFieldType(HeroManagerFieldIndexHeroMap, fieldmeta.FieldTypeSync)
	m.fieldMetas.SetFieldType(HeroManagerFieldIndexSingleHeroModule, fieldmeta.FieldTypeSync)

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

// MatchesAll 判断字段是否匹配所有类型标记
func (m *HeroManagerWrapper) MatchesAll(fieldID uint8, fieldTypes ...fieldmeta.FieldType) bool {
	return m.fieldMetas.MatchesAll(fieldID, fieldTypes...)
}

// SetHeroMap 设置/添加HeroModule模块
func (m *HeroManagerWrapper) HeroMap_Set(id int64, value *HeroModule) *HeroModuleWrapper {
	return m.heroMapLink.Set(id, value)
}

// DeleteHeroMap 删除HeroModule模块
func (m *HeroManagerWrapper) HeroMap_Delete(id int64) bool {
	return m.heroMapLink.Delete(id)
}

// RangeHeroMap 遍历所有HeroModule
func (m *HeroManagerWrapper) HeroMap_Range(f func(id int64, heroModule *HeroModuleWrapper) bool) {
	m.heroMapLink.Range(f)
}

func (m *HeroManagerWrapper) ClearAllDirtyFlags() {
	m.IDirtyFlag.ClearAllDirty()

	// 清空xmaplink中所有object的脏标记
	m.heroMapLink.RangeIncrementalSyncObject(func(key int64, object xmapwrapper.IncrementalSyncObject) (stop bool) {
		object.ClearAllDirty()
		return false
	})
}

func (m *HeroManagerWrapper) Reset(newData *HeroManager) {
	if m == nil || m.data == nil {
		return
	}

	// 重新构建脏标系统
	m.IDirtyFlag.ClearAllDirty()

	// xmap全量覆盖数据，不需要处理Value的逻辑
	data := NewHeroManager()
	m.data = data
	m.heroMapLink.Reset(&m.data.HeroMap)
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
	if m.IsDirty(HeroManagerDirtySingleHeroModuleBit) {
		if m.SingleHeroModuleWrapper != nil {
			m.SingleHeroModuleWrapper.BuildMongoUpdate(builder, path.Field("single_hero_module"))
		}
	}
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

	for key := range pb.HeroMap {
		pbValue := pb.HeroMap[key]
		v := NewHeroModule()
		wrapper := m.heroMapLink.SetWithoutTrack(key, v)
		wrapper.FromProto(pbValue)
	}
	if pb.SingleHeroModule != nil {
		if m.data.SingleHeroModule == nil {
			m.data.SingleHeroModule = NewHeroModule()
		}
		m.SingleHeroModuleWrapper.FromProto(pb.SingleHeroModule)
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
	if m.data.SingleHeroModule != nil {
		if copy.SingleHeroModule == nil {
			copy.SingleHeroModule = &HeroModule{}
		}
		m.SingleHeroModuleWrapper.DeepCopyTo(copy.SingleHeroModule)
	}
}

// ToIncrementalProtoWithContext 根据脏标记位构建增量数据的 protoMessage
// 只返回标记为脏的字段数据，用于增量同步
func (m *HeroManagerWrapper) ToIncrementalProtoWithContext(ctx mmemodel.SyncContext) proto.Message {
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
				heroModule, _ := m.heroMapLink.Get(key)
				pb := heroModule.ToIncrementalProtoWithContext(ctx)
				v, ok := pb.(*mme.HeroModule)
				if ok {
					incremental.HeroMap_XXXChangeList = append(incremental.HeroMap_XXXChangeList, &mme.HeroManager_HeroMap_XXXMapChangeRecord{
						Key:   key,
						Value: v,
					})
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
	if mmemodel.FieldCanBeIncrementalSynced(m, HeroManagerDirtySingleHeroModuleBit, HeroManagerFieldIndexSingleHeroModule, ctx) {
		if m.SingleHeroModuleWrapper != nil {
			pbObj := m.SingleHeroModuleWrapper.ToIncrementalProtoWithContext(ctx)
			if pbObj != nil {
				v, ok := pbObj.(*mme.HeroModule)
				if ok {
					incremental.SingleHeroModule = v
				}
			}
		}
	}
	return incremental
}
