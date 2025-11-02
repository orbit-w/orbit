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
	HeroModuleFieldIndexBase = uint8(0)
	HeroModuleFieldIndexLevelUp
)

// Dirty bits for HeroModule fields
const (
	HeroModuleDirtyBaseBit    int64 = 1 << HeroModuleFieldIndexBase
	HeroModuleDirtyLevelUpBit int64 = 1 << HeroModuleFieldIndexLevelUp
)

type HeroModule struct {
	Base    *HeroMechanism    `bson:"base"`
	LevelUp *LevelUpMechanism `bson:"level_up"`
}

func NewHeroModule() *HeroModule {
	return &HeroModule{
		Base:    NewHeroMechanism(),
		LevelUp: NewLevelUpMechanism(),
	}
}

// 数据-深拷贝
func (m *HeroModule) DeepCopy(co *HeroModule) {
	if m == nil || co == nil {
		return
	}

	*co = *m

	if m.Base != nil {
		co.Base = &HeroMechanism{}
		m.Base.DeepCopy(co.Base)
	}

	if m.LevelUp != nil {
		co.LevelUp = &LevelUpMechanism{}
		m.LevelUp.DeepCopy(co.LevelUp)
	}
}

// 数据-转换为protobuf
func (m *HeroModule) ToProto() *mme.HeroModule {
	if m == nil {
		return nil
	}

	pb := &mme.HeroModule{}

	if m.Base != nil {
		pb.Base = m.Base.ToProto()
	}

	if m.LevelUp != nil {
		pb.LevelUp = m.LevelUp.ToProto()
	}

	return pb
}

type HeroModuleWrapper struct {
	data *HeroModule
	dirtyflag.IDirtyFlag
	fieldMetas *fieldmeta.FieldMetas

	BaseWrapper    *HeroMechanismWrapper
	LevelUpWrapper *LevelUpMechanismWrapper
}

func NewHeroModuleWrapper(data *HeroModule) *HeroModuleWrapper {
	if data == nil {
		panic("data is nil")
	}
	hm := &HeroModuleWrapper{
		data:       data,
		IDirtyFlag: dirtyflag.NewDirtyFlag(),
		fieldMetas: fieldmeta.NewFieldMetas(),
	}

	// 初始化嵌套的 Mechanism Wrapper
	if data.Base == nil {
		data.Base = NewHeroMechanism()
	}
	hm.BaseWrapper = NewHeroMechanismWrapper(data.Base)
	hm.BaseWrapper.Link(hm.GetDirtyTracker(), HeroModuleDirtyBaseBit)

	if data.LevelUp == nil {
		data.LevelUp = NewLevelUpMechanism()
	}
	hm.LevelUpWrapper = NewLevelUpMechanismWrapper(data.LevelUp)
	hm.LevelUpWrapper.Link(hm.GetDirtyTracker(), HeroModuleDirtyLevelUpBit)

	return hm
}

func (m *HeroModuleWrapper) InitFieldContext() {
	m.fieldMetas.SetFieldType(HeroModuleFieldIndexBase, fieldmeta.FieldTypeSync)
	m.fieldMetas.SetFieldType(HeroModuleFieldIndexLevelUp, fieldmeta.FieldTypeSync)

	// 初始化嵌套 Mechanism 的字段上下文
	if m.BaseWrapper != nil {
		m.BaseWrapper.InitFieldContext()
	}
	if m.LevelUpWrapper != nil {
		m.LevelUpWrapper.InitFieldContext()
	}
}

// 模块唯一名称
func (m *HeroModuleWrapper) Name() string {
	return "HeroModule"
}

func (m *HeroModuleWrapper) GetBase() *HeroMechanismWrapper {
	return m.BaseWrapper
}

func (m *HeroModuleWrapper) GetLevelUp() *LevelUpMechanismWrapper {
	return m.LevelUpWrapper
}

// ClearAllDirtyFlags 清除所有脏标记位
func (m *HeroModuleWrapper) ClearAllDirtyFlags() {
	m.ClearAllDirty()

	// 清除嵌套 Mechanism 的脏标记
	if m.BaseWrapper != nil {
		m.BaseWrapper.ClearAllDirtyFlags()
	}
	if m.LevelUpWrapper != nil {
		m.LevelUpWrapper.ClearAllDirtyFlags()
	}
}

// BuildMongoUpdate 构建MongoDB更新操作
func (m *HeroModuleWrapper) BuildMongoUpdate(builder *mgo_builder.MongoUpdateBuilder, path *mgo_builder.NestedPath) {
	if m == nil {
		return
	}

	// 构建嵌套字段路径
	if m.BaseWrapper != nil && m.IsDirty(HeroModuleDirtyBaseBit) {
		m.BaseWrapper.BuildMongoUpdate(builder, path.Field("base"))
	}

	if m.LevelUpWrapper != nil && m.IsDirty(HeroModuleDirtyLevelUpBit) {
		m.LevelUpWrapper.BuildMongoUpdate(builder, path.Field("level_up"))
	}
}

// 包装器-深拷贝
func (m *HeroModuleWrapper) DeepCopy() *HeroModule {
	if m == nil {
		return nil
	}
	copy := &HeroModule{}
	m.DeepCopyTo(copy)
	return copy
}

// 包装器-深拷贝
func (m *HeroModuleWrapper) DeepCopyTo(copy *HeroModule) {
	if m == nil || m.data == nil || copy == nil {
		return
	}

	// 拷贝嵌套的 Mechanism 对象
	if m.data.Base != nil {
		if copy.Base == nil {
			copy.Base = &HeroMechanism{}
		}
		m.BaseWrapper.DeepCopyTo(copy.Base)
	}

	if m.data.LevelUp != nil {
		if copy.LevelUp == nil {
			copy.LevelUp = &LevelUpMechanism{}
		}
		m.LevelUpWrapper.DeepCopyTo(copy.LevelUp)
	}
}

// MatchesAll 判断字段是否匹配所有类型标记
func (m *HeroModuleWrapper) MatchesAll(fieldID uint8, fieldTypes ...fieldmeta.FieldType) bool {
	return m.fieldMetas.MatchesAll(fieldID, fieldTypes...)
}

// ToProto 将 HeroModule 数据转换为完整的 protobuf 结构体
func (m *HeroModuleWrapper) ToProto() *mme.HeroModule {
	if m == nil || m.data == nil {
		return nil
	}

	return m.data.ToProto()
}

// FromProto 从 protobuf 结构体加载数据到 HeroModule
func (m *HeroModuleWrapper) FromProto(pb *mme.HeroModule) {
	if m == nil || m.data == nil || pb == nil {
		return
	}

	// 加载 Base 数据
	if pb.Base != nil {
		m.data.Base = &HeroMechanism{}
		m.BaseWrapper = NewHeroMechanismWrapper(m.data.Base)
		m.BaseWrapper.Link(m.GetDirtyTracker(), HeroModuleDirtyBaseBit)
		m.BaseWrapper.FromProto(pb.Base)
	}

	// 加载 LevelUp 数据
	if pb.LevelUp != nil {
		m.data.LevelUp = &LevelUpMechanism{}
		m.LevelUpWrapper = NewLevelUpMechanismWrapper(m.data.LevelUp)
		m.LevelUpWrapper.Link(m.GetDirtyTracker(), HeroModuleDirtyLevelUpBit)
		m.LevelUpWrapper.FromProto(pb.LevelUp)
	}
}

// ToIncrementalProto 根据脏标记位构建增量数据的 protoMessage
// 只返回标记为脏的字段数据，用于增量同步
func (m *HeroModuleWrapper) ToIncrementalProtoWithContext(ctx mmemodel.SyncContext) proto.Message {
	if m == nil {
		return nil
	}

	// 如果没有脏标记，返回 nil
	if !m.HasAnyDirty() {
		return nil
	}

	incremental := &mme.HeroModule{}

	// 根据脏标记位设置对应的嵌套对象
	if mmemodel.FieldCanBeIncrementalSynced(m, HeroModuleDirtyBaseBit, HeroModuleFieldIndexBase, ctx) {
		if m.BaseWrapper != nil {
			pb := m.BaseWrapper.ToIncrementalProtoWithContext(ctx)
			if pb != nil {
				v, ok := pb.(*mme.HeroMechanism)
				if ok {
					incremental.Base = v
				}
			}
		}
	}

	if mmemodel.FieldCanBeIncrementalSynced(m, HeroModuleDirtyLevelUpBit, HeroModuleFieldIndexLevelUp, ctx) {
		if m.LevelUpWrapper != nil {
			pb := m.LevelUpWrapper.ToIncrementalProtoWithContext(ctx)
			if pb != nil {
				v, ok := pb.(*mme.LevelUpMechanism)
				if ok {
					incremental.LevelUp = v
				}
			}
		}
	}

	return incremental
}
