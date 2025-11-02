package mme

import (
	"gitee.com/orbit-w/orbit/app/proto/mme"
	dirtyflag "gitee.com/orbit-w/orbit/lib/base/dirty_flag"
	fieldmeta "gitee.com/orbit-w/orbit/lib/base/field_meta"
	"gitee.com/orbit-w/orbit/lib/module/db/mgo_builder"
	mmemodel "gitee.com/orbit-w/orbit/lib/module/mme_model"
	"google.golang.org/protobuf/proto"
)

const (
	PlayerEntityFieldIndexId = uint8(0)
	PlayerEntityFieldIndexHeroManager
)

// Dirty bits for LevelUpMechanism fields
const (
	PlayerEntityDirtyIdBit          int64 = 1 << PlayerEntityFieldIndexId
	PlayerEntityDirtyHeroManagerBit int64 = 1 << PlayerEntityFieldIndexHeroManager
)

type PlayerEntity struct {
	XXXId       int64
	HeroManager *HeroManager
}

func NewPlayerEntity() *PlayerEntity {
	return &PlayerEntity{
		HeroManager: NewHeroManager(),
	}
}

func (e *PlayerEntity) DeepCopy(co *PlayerEntity) {
	if e == nil || co == nil {
		return
	}

	*co = *e
	co.HeroManager = &HeroManager{}
	e.HeroManager.DeepCopy(co.HeroManager)
}

// 数据-转换为protobuf
func (e *PlayerEntity) ToProto() *mme.PlayerEntity {
	if e == nil {
		return nil
	}

	pb := &mme.PlayerEntity{}

	id := e.XXXId
	pb.XXXId = id

	pb.HeroManager = e.HeroManager.ToProto()
	return pb
}

type PlayerEntityWrapper struct {
	data *PlayerEntity
	dirtyflag.IDirtyFlag
	fieldMetas *fieldmeta.FieldMetas

	HeroManagerWrapper *HeroManagerWrapper
}

func NewPlayerEntityWrapper(data *PlayerEntity) *PlayerEntityWrapper {
	if data == nil {
		panic("data is nil")
	}

	w := &PlayerEntityWrapper{data: data, IDirtyFlag: dirtyflag.NewDirtyFlag(), fieldMetas: fieldmeta.NewFieldMetas()}
	if data.HeroManager == nil {
		data.HeroManager = NewHeroManager()
	}

	w.HeroManagerWrapper = NewHeroManagerWrapper(data.HeroManager)
	w.HeroManagerWrapper.Link(w.GetDirtyTracker(), PlayerEntityDirtyHeroManagerBit)
	return w
}

func (w *PlayerEntityWrapper) InitFieldContext() {
	w.fieldMetas.SetFieldType(PlayerEntityFieldIndexId, fieldmeta.FieldTypeSync)
	w.fieldMetas.SetFieldType(PlayerEntityFieldIndexHeroManager, fieldmeta.FieldTypeSync)
	w.HeroManagerWrapper.InitFieldContext()
}

func (w *PlayerEntityWrapper) GetId() int64 {
	return w.data.XXXId
}

func (w *PlayerEntityWrapper) SetId(v int64) {
	w.data.XXXId = v
	w.MarkDirty(PlayerEntityDirtyIdBit)
}

func (w *PlayerEntityWrapper) Name() string {
	return "PlayerEntity"
}

func (w *PlayerEntityWrapper) GetHeroManager() *HeroManagerWrapper {
	return w.HeroManagerWrapper
}

func (w *PlayerEntityWrapper) ClearAllDirty() {
	w.IDirtyFlag.ClearAllDirty()
	w.HeroManagerWrapper.ClearAllDirty()
}

func (w *PlayerEntityWrapper) Reset(newData *HeroManager) {
	if w == nil || w.data == nil {
		return
	}

	// 重新构建脏标系统
	w.IDirtyFlag.ClearAllDirty()

	data := NewPlayerEntity()
	w.data = data
	w.HeroManagerWrapper.Reset(data.HeroManager)
}

func (w *PlayerEntityWrapper) BuildMongoUpdate(builder *mgo_builder.MongoUpdateBuilder, path *mgo_builder.NestedPath) {
	if w == nil {
		return
	}

	if w.IsDirty(PlayerEntityDirtyIdBit) {
		builder.SetNestedPath(path, "_id", w.GetId())
	}

	if w.IsDirty(PlayerEntityDirtyHeroManagerBit) {
		w.HeroManagerWrapper.BuildMongoUpdate(builder, path.Field("hero_manager"))
	}
}

func (w *PlayerEntityWrapper) DeepCopy() *PlayerEntity {
	if w == nil {
		return nil
	}
	copy := &PlayerEntity{}
	w.DeepCopyTo(copy)
	return copy
}

// 包装器-深拷贝
func (w *PlayerEntityWrapper) DeepCopyTo(copy *PlayerEntity) {
	if w == nil || w.data == nil || copy == nil {
		return
	}

	// 初始化目标 map
	if copy.HeroManager == nil {
		copy.HeroManager = NewHeroManager()
	}

	// 拷贝所有 HeroModule
	w.HeroManagerWrapper.DeepCopyTo(copy.HeroManager)
}

// MatchesAll 判断字段是否匹配所有类型标记
func (w *PlayerEntityWrapper) MatchesAll(fieldID uint8, fieldTypes ...fieldmeta.FieldType) bool {
	return w.fieldMetas.MatchesAll(fieldID, fieldTypes...)
}

// ToProto 将 PlayerEntity 数据转换为完整的 protobuf 结构体
func (w *PlayerEntityWrapper) ToProto() *mme.PlayerEntity {
	if w == nil || w.data == nil {
		return nil
	}

	return w.data.ToProto()
}

// FromProto 从 protobuf 结构体加载数据到 PlayerEntity
func (w *PlayerEntityWrapper) FromProto(pb *mme.PlayerEntity) {
	if w == nil || w.data == nil || pb == nil {
		return
	}

	w.HeroManagerWrapper.FromProto(pb.HeroManager)
}

// ToIncrementalProto 根据脏标记位构建增量数据的 protoMessage
// 只返回标记为脏的字段数据，用于增量同步
func (w *PlayerEntityWrapper) ToIncrementalProto(ctx mmemodel.SyncContext) proto.Message {
	if w == nil {
		return nil
	}

	// 如果没有脏标记，返回 nil
	if !w.HasAnyDirty() {
		return nil
	}

	incremental := &mme.PlayerEntity{}

	// 根据脏标记位设置对应的字段
	if mmemodel.FieldCanBeIncrementalSynced(w, PlayerEntityDirtyIdBit, PlayerEntityFieldIndexId, ctx) {
		v := w.GetId()
		incremental.XXXId = v
	}

	// 根据脏标记位设置对应的嵌套对象
	if mmemodel.FieldCanBeIncrementalSynced(w, PlayerEntityDirtyHeroManagerBit, PlayerEntityFieldIndexHeroManager, ctx) {
		if w.HeroManagerWrapper != nil {
			pb := w.HeroManagerWrapper.ToIncrementalProto(ctx)
			if pb != nil {
				v, ok := pb.(*mme.HeroManager)
				if ok {
					incremental.HeroManager = v
				}
			}
		}
	}

	return incremental
}
