package mme

import (
	"gitee.com/orbit-w/orbit/app/proto/enum"
	"gitee.com/orbit-w/orbit/app/proto/mme"
	dirtyflag "gitee.com/orbit-w/orbit/lib/base/dirty_flag"
	fieldmeta "gitee.com/orbit-w/orbit/lib/base/field_meta"
	"gitee.com/orbit-w/orbit/lib/module/db/mgo_builder"
	mmemodel "gitee.com/orbit-w/orbit/lib/module/mme_model"
	"google.golang.org/protobuf/proto"
)

const (
	PlayerEntityFieldIndexXXXId       = uint8(0)
	PlayerEntityFieldIndexHeroManager = uint8(1)
)

// Dirty bits for PlayerEntity fields
const (
	PlayerEntityDirtyXXXIdBit       int64 = 1 << PlayerEntityFieldIndexXXXId
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

func (e *PlayerEntity) ToProto() proto.Message {
	if e == nil {
		return nil
	}

	pb := &mme.PlayerEntity{}

	id := e.XXXId
	pb.XXXId = id

	if e.HeroManager != nil {
		pb.HeroManager = e.HeroManager.ToProto()
	}
	return pb
}

func (e *PlayerEntity) FromProto(msg proto.Message) {
	if e == nil || msg == nil {
		return
	}

	pb, ok := msg.(*mme.PlayerEntity)
	if !ok {
		return
	}

	e.XXXId = pb.XXXId

	if pb.HeroManager != nil {
		if e.HeroManager == nil {
			e.HeroManager = NewHeroManager()
		}
		e.HeroManager.FromProto(pb.HeroManager)
	}
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
	e := &PlayerEntityWrapper{
		data:       data,
		IDirtyFlag: dirtyflag.NewDirtyFlag(),
		fieldMetas: fieldmeta.NewFieldMetas(),
	}

	// 初始化嵌套的 HeroManager Wrapper
	if data.HeroManager == nil {
		data.HeroManager = NewHeroManager()
	}
	e.HeroManagerWrapper = NewHeroManagerWrapper(data.HeroManager)
	e.HeroManagerWrapper.Link(e.GetDirtyTracker(), PlayerEntityDirtyHeroManagerBit)

	return e
}

func (e *PlayerEntityWrapper) InitFieldContext() {
	e.fieldMetas.SetFieldType(PlayerEntityFieldIndexXXXId, fieldmeta.FieldTypeSync)
	e.fieldMetas.SetFieldType(PlayerEntityFieldIndexHeroManager, fieldmeta.FieldTypeSync)

	if e.HeroManagerWrapper != nil {
		e.HeroManagerWrapper.InitFieldContext()
	}
}

func (e *PlayerEntityWrapper) Name() string {
	return "PlayerEntity"
}

// GetEntityType 返回实体类型枚举
func (e *PlayerEntityWrapper) GetEntityType() enum.EntityType {
	return enum.EntityType_PlayerEntity
}

// MatchesAll 判断字段是否匹配所有类型标记
func (e *PlayerEntityWrapper) MatchesAll(fieldID uint8, fieldTypes ...fieldmeta.FieldType) bool {
	return e.fieldMetas.MatchesAll(fieldID, fieldTypes...)
}

// 包装器-获取XXXId
func (e *PlayerEntityWrapper) GetXXXId() int64 {
	return e.data.XXXId
}

// 包装器-设置XXXId
func (e *PlayerEntityWrapper) SetXXXId(v int64) {
	e.data.XXXId = v
	e.MarkDirty(PlayerEntityDirtyXXXIdBit)
}

func (e *PlayerEntityWrapper) GetHeroManager() *HeroManagerWrapper {
	return e.HeroManagerWrapper
}

// ClearAllDirtyFlags 清除所有脏标记位
func (e *PlayerEntityWrapper) ClearAllDirtyFlags() {
	e.ClearAllDirty()

	// 清除嵌套 Manager 的脏标记
	if e.HeroManagerWrapper != nil {
		e.HeroManagerWrapper.ClearAllDirtyFlags()
	}
}

// BuildMongoUpdate 构建MongoDB更新操作
func (e *PlayerEntityWrapper) BuildMongoUpdate(builder *mgo_builder.MongoUpdateBuilder, path *mgo_builder.NestedPath) {
	if e == nil {
		return
	}

	if e.IsDirty(PlayerEntityDirtyHeroManagerBit) {
		if e.HeroManagerWrapper != nil {
			e.HeroManagerWrapper.BuildMongoUpdate(builder, path.Field("hero_manager"))
		}
	}
}

// ToProto 将 PlayerEntity 数据转换为完整的 protobuf 结构体
func (e *PlayerEntityWrapper) ToProto() proto.Message {
	if e == nil || e.data == nil {
		return nil
	}

	return e.data.ToProto()
}

// FromProto 从 protobuf 结构体加载数据到 PlayerEntity
func (e *PlayerEntityWrapper) FromProto(msg proto.Message) {
	if e == nil || e.data == nil || msg == nil {
		return
	}

	pb, ok := msg.(*mme.PlayerEntity)
	if !ok {
		return
	}

	if pb.HeroManager != nil {
		if e.data.HeroManager == nil {
			e.data.HeroManager = NewHeroManager()
		}
		e.HeroManagerWrapper.FromProto(pb.HeroManager)
	}
}

// 包装器-深拷贝
func (e *PlayerEntityWrapper) DeepCopy() *PlayerEntity {
	if e == nil {
		return nil
	}
	copy := &PlayerEntity{}
	e.DeepCopyTo(copy)
	return copy
}

// 包装器-深拷贝
func (e *PlayerEntityWrapper) DeepCopyTo(copy *PlayerEntity) {
	if e == nil || e.data == nil || copy == nil {
		return
	}

	if e.data.HeroManager != nil {
		if copy.HeroManager == nil {
			copy.HeroManager = &HeroManager{}
		}
		e.HeroManagerWrapper.DeepCopyTo(copy.HeroManager)
	}
}

// ToIncrementalProtoWithContext 根据脏标记位构建增量数据的 protoMessage
// 只返回标记为脏的字段数据，用于增量同步
func (e *PlayerEntityWrapper) ToIncrementalProtoWithContext(ctx mmemodel.SyncContext) proto.Message {
	if e == nil {
		return nil
	}

	// 如果没有脏标记，返回 nil
	if !e.HasAnyDirty() {
		return nil
	}

	incremental := &mme.PlayerEntity{}

	if mmemodel.FieldCanBeIncrementalSynced(e, PlayerEntityDirtyHeroManagerBit, PlayerEntityFieldIndexHeroManager, ctx) {
		if e.HeroManagerWrapper != nil {
			pb := e.HeroManagerWrapper.ToIncrementalProtoWithContext(ctx)
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
