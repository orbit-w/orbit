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
	HeroModuleFieldIndexBase         = uint8(0)
	HeroModuleFieldIndexLevelUp      = uint8(1)
	HeroModuleFieldIndexTalentUnlock = uint8(2)
	HeroModuleFieldIndexSkinWear     = uint8(3)
)

// Dirty bits for HeroModule fields
const (
	HeroModuleDirtyBaseBit         int64 = 1 << HeroModuleFieldIndexBase
	HeroModuleDirtyLevelUpBit      int64 = 1 << HeroModuleFieldIndexLevelUp
	HeroModuleDirtyTalentUnlockBit int64 = 1 << HeroModuleFieldIndexTalentUnlock
	HeroModuleDirtySkinWearBit     int64 = 1 << HeroModuleFieldIndexSkinWear
)

type HeroModule struct {
	Base         *HeroMechanism         `bson:"base"`
	LevelUp      *LevelUpMechanism      `bson:"level_up"`
	TalentUnlock *ManualUnlockMechanism `bson:"talent_unlock"`
	SkinWear     *WearMechanism         `bson:"skin_wear"`
}

func NewHeroModule() *HeroModule {
	return &HeroModule{
		Base:         NewHeroMechanism(),
		LevelUp:      NewLevelUpMechanism(),
		TalentUnlock: NewManualUnlockMechanism(),
		SkinWear:     NewWearMechanism(),
	}
}

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
	if m.TalentUnlock != nil {
		co.TalentUnlock = &ManualUnlockMechanism{}
		m.TalentUnlock.DeepCopy(co.TalentUnlock)
	}
	if m.SkinWear != nil {
		co.SkinWear = &WearMechanism{}
		m.SkinWear.DeepCopy(co.SkinWear)
	}
}

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
	if m.TalentUnlock != nil {
		pb.TalentUnlock = m.TalentUnlock.ToProto()
	}
	if m.SkinWear != nil {
		pb.SkinWear = m.SkinWear.ToProto()
	}
	return pb
}

func (m *HeroModule) FromProto(pb *mme.HeroModule) {
	if m == nil || pb == nil {
		return
	}

	if pb.Base != nil {
		if m.Base == nil {
			m.Base = NewHeroMechanism()
		}
		m.Base.FromProto(pb.Base)
	}
	if pb.LevelUp != nil {
		if m.LevelUp == nil {
			m.LevelUp = NewLevelUpMechanism()
		}
		m.LevelUp.FromProto(pb.LevelUp)
	}
	if pb.TalentUnlock != nil {
		if m.TalentUnlock == nil {
			m.TalentUnlock = NewManualUnlockMechanism()
		}
		m.TalentUnlock.FromProto(pb.TalentUnlock)
	}
	if pb.SkinWear != nil {
		if m.SkinWear == nil {
			m.SkinWear = NewWearMechanism()
		}
		m.SkinWear.FromProto(pb.SkinWear)
	}
}

type HeroModuleWrapper struct {
	data *HeroModule
	dirtyflag.IDirtyFlag
	fieldMetas *fieldmeta.FieldMetas

	BaseWrapper         *HeroMechanismWrapper
	LevelUpWrapper      *LevelUpMechanismWrapper
	TalentUnlockWrapper *ManualUnlockMechanismWrapper
	SkinWearWrapper     *WearMechanismWrapper
}

func NewHeroModuleWrapper(data *HeroModule) *HeroModuleWrapper {
	if data == nil {
		panic("data is nil")
	}
	w := &HeroModuleWrapper{
		data:       data,
		IDirtyFlag: dirtyflag.NewDirtyFlag(),
		fieldMetas: fieldmeta.NewFieldMetas(),
	}

	// 初始化嵌套的 HeroMechanism Wrapper
	if data.Base == nil {
		data.Base = NewHeroMechanism()
	}
	w.BaseWrapper = NewHeroMechanismWrapper(data.Base)
	w.BaseWrapper.Link(w.GetDirtyTracker(), HeroModuleDirtyBaseBit)

	// 初始化嵌套的 LevelUpMechanism Wrapper
	if data.LevelUp == nil {
		data.LevelUp = NewLevelUpMechanism()
	}
	w.LevelUpWrapper = NewLevelUpMechanismWrapper(data.LevelUp)
	w.LevelUpWrapper.Link(w.GetDirtyTracker(), HeroModuleDirtyLevelUpBit)

	// 初始化嵌套的 ManualUnlockMechanism Wrapper
	if data.TalentUnlock == nil {
		data.TalentUnlock = NewManualUnlockMechanism()
	}
	w.TalentUnlockWrapper = NewManualUnlockMechanismWrapper(data.TalentUnlock)
	w.TalentUnlockWrapper.Link(w.GetDirtyTracker(), HeroModuleDirtyTalentUnlockBit)

	// 初始化嵌套的 WearMechanism Wrapper
	if data.SkinWear == nil {
		data.SkinWear = NewWearMechanism()
	}
	w.SkinWearWrapper = NewWearMechanismWrapper(data.SkinWear)
	w.SkinWearWrapper.Link(w.GetDirtyTracker(), HeroModuleDirtySkinWearBit)

	return w
}

func (w *HeroModuleWrapper) InitFieldContext() {
	w.fieldMetas.SetFieldType(HeroModuleFieldIndexBase, fieldmeta.FieldTypeSync)
	w.fieldMetas.SetFieldType(HeroModuleFieldIndexLevelUp, fieldmeta.FieldTypeSync)
	w.fieldMetas.SetFieldType(HeroModuleFieldIndexTalentUnlock, fieldmeta.FieldTypeSync)
	w.fieldMetas.SetFieldType(HeroModuleFieldIndexSkinWear, fieldmeta.FieldTypeSync)

	if w.BaseWrapper != nil {
		w.BaseWrapper.InitFieldContext()
	}
	if w.LevelUpWrapper != nil {
		w.LevelUpWrapper.InitFieldContext()
	}
	if w.TalentUnlockWrapper != nil {
		w.TalentUnlockWrapper.InitFieldContext()
	}
	if w.SkinWearWrapper != nil {
		w.SkinWearWrapper.InitFieldContext()
	}
}

func (w *HeroModuleWrapper) Name() string {
	return "HeroModule"
}

// MatchesAll 判断字段是否匹配所有类型标记
func (w *HeroModuleWrapper) MatchesAll(fieldID uint8, fieldTypes ...fieldmeta.FieldType) bool {
	return w.fieldMetas.MatchesAll(fieldID, fieldTypes...)
}

func (w *HeroModuleWrapper) GetBase() *HeroMechanismWrapper {
	return w.BaseWrapper
}

func (w *HeroModuleWrapper) GetLevelUp() *LevelUpMechanismWrapper {
	return w.LevelUpWrapper
}

func (w *HeroModuleWrapper) GetTalentUnlock() *ManualUnlockMechanismWrapper {
	return w.TalentUnlockWrapper
}

func (w *HeroModuleWrapper) GetSkinWear() *WearMechanismWrapper {
	return w.SkinWearWrapper
}

// ClearAllDirtyFlags 清除所有脏标记位
func (w *HeroModuleWrapper) ClearAllDirtyFlags() {
	w.ClearAllDirty()

	// 清除嵌套 Mechanism 的脏标记
	if w.BaseWrapper != nil {
		w.BaseWrapper.ClearAllDirtyFlags()
	}
	if w.LevelUpWrapper != nil {
		w.LevelUpWrapper.ClearAllDirtyFlags()
	}
	if w.TalentUnlockWrapper != nil {
		w.TalentUnlockWrapper.ClearAllDirtyFlags()
	}
	if w.SkinWearWrapper != nil {
		w.SkinWearWrapper.ClearAllDirtyFlags()
	}
}

// BuildMongoUpdate 构建MongoDB更新操作
func (w *HeroModuleWrapper) BuildMongoUpdate(builder *mgo_builder.MongoUpdateBuilder, path *mgo_builder.NestedPath) {
	if w == nil {
		return
	}

	if w.IsDirty(HeroModuleDirtyBaseBit) {
		if w.BaseWrapper != nil {
			w.BaseWrapper.BuildMongoUpdate(builder, path.Field("base"))
		}
	}
	if w.IsDirty(HeroModuleDirtyLevelUpBit) {
		if w.LevelUpWrapper != nil {
			w.LevelUpWrapper.BuildMongoUpdate(builder, path.Field("level_up"))
		}
	}
	if w.IsDirty(HeroModuleDirtyTalentUnlockBit) {
		if w.TalentUnlockWrapper != nil {
			w.TalentUnlockWrapper.BuildMongoUpdate(builder, path.Field("talent_unlock"))
		}
	}
	if w.IsDirty(HeroModuleDirtySkinWearBit) {
		if w.SkinWearWrapper != nil {
			w.SkinWearWrapper.BuildMongoUpdate(builder, path.Field("skin_wear"))
		}
	}
}

// ToProto 将 HeroModule 数据转换为完整的 protobuf 结构体
func (w *HeroModuleWrapper) ToProto() *mme.HeroModule {
	if w == nil || w.data == nil {
		return nil
	}

	return w.data.ToProto()
}

// FromProto 从 protobuf 结构体加载数据到 HeroModule
func (w *HeroModuleWrapper) FromProto(pb *mme.HeroModule) {
	if w == nil || w.data == nil || pb == nil {
		return
	}

	if pb.Base != nil {
		if w.data.Base == nil {
			w.data.Base = NewHeroMechanism()
		}
		w.BaseWrapper.FromProto(pb.Base)
	}
	if pb.LevelUp != nil {
		if w.data.LevelUp == nil {
			w.data.LevelUp = NewLevelUpMechanism()
		}
		w.LevelUpWrapper.FromProto(pb.LevelUp)
	}
	if pb.TalentUnlock != nil {
		if w.data.TalentUnlock == nil {
			w.data.TalentUnlock = NewManualUnlockMechanism()
		}
		w.TalentUnlockWrapper.FromProto(pb.TalentUnlock)
	}
	if pb.SkinWear != nil {
		if w.data.SkinWear == nil {
			w.data.SkinWear = NewWearMechanism()
		}
		w.SkinWearWrapper.FromProto(pb.SkinWear)
	}
}

// 包装器-深拷贝
func (w *HeroModuleWrapper) DeepCopy() *HeroModule {
	if w == nil {
		return nil
	}
	copy := &HeroModule{}
	w.DeepCopyTo(copy)
	return copy
}

// 包装器-深拷贝
func (w *HeroModuleWrapper) DeepCopyTo(copy *HeroModule) {
	if w == nil || w.data == nil || copy == nil {
		return
	}

	if w.data.Base != nil {
		if copy.Base == nil {
			copy.Base = &HeroMechanism{}
		}
		w.BaseWrapper.DeepCopyTo(copy.Base)
	}
	if w.data.LevelUp != nil {
		if copy.LevelUp == nil {
			copy.LevelUp = &LevelUpMechanism{}
		}
		w.LevelUpWrapper.DeepCopyTo(copy.LevelUp)
	}
	if w.data.TalentUnlock != nil {
		if copy.TalentUnlock == nil {
			copy.TalentUnlock = &ManualUnlockMechanism{}
		}
		w.TalentUnlockWrapper.DeepCopyTo(copy.TalentUnlock)
	}
	if w.data.SkinWear != nil {
		if copy.SkinWear == nil {
			copy.SkinWear = &WearMechanism{}
		}
		w.SkinWearWrapper.DeepCopyTo(copy.SkinWear)
	}
}

// ToIncrementalProtoWithContext 根据脏标记位构建增量数据的 protoMessage
// 只返回标记为脏的字段数据，用于增量同步
func (w *HeroModuleWrapper) ToIncrementalProtoWithContext(ctx mmemodel.SyncContext) proto.Message {
	if w == nil {
		return nil
	}

	// 如果没有脏标记，返回 nil
	if !w.HasAnyDirty() {
		return nil
	}

	incremental := &mme.HeroModule{}

	if mmemodel.FieldCanBeIncrementalSynced(w, HeroModuleDirtyBaseBit, HeroModuleFieldIndexBase, ctx) {
		if w.BaseWrapper != nil {
			pb := w.BaseWrapper.ToIncrementalProtoWithContext(ctx)
			if pb != nil {
				v, ok := pb.(*mme.HeroMechanism)
				if ok {
					incremental.Base = v
				}
			}
		}
	}
	if mmemodel.FieldCanBeIncrementalSynced(w, HeroModuleDirtyLevelUpBit, HeroModuleFieldIndexLevelUp, ctx) {
		if w.LevelUpWrapper != nil {
			pb := w.LevelUpWrapper.ToIncrementalProtoWithContext(ctx)
			if pb != nil {
				v, ok := pb.(*mme.LevelUpMechanism)
				if ok {
					incremental.LevelUp = v
				}
			}
		}
	}
	if mmemodel.FieldCanBeIncrementalSynced(w, HeroModuleDirtyTalentUnlockBit, HeroModuleFieldIndexTalentUnlock, ctx) {
		if w.TalentUnlockWrapper != nil {
			pb := w.TalentUnlockWrapper.ToIncrementalProtoWithContext(ctx)
			if pb != nil {
				v, ok := pb.(*mme.ManualUnlockMechanism)
				if ok {
					incremental.TalentUnlock = v
				}
			}
		}
	}
	if mmemodel.FieldCanBeIncrementalSynced(w, HeroModuleDirtySkinWearBit, HeroModuleFieldIndexSkinWear, ctx) {
		if w.SkinWearWrapper != nil {
			pb := w.SkinWearWrapper.ToIncrementalProtoWithContext(ctx)
			if pb != nil {
				v, ok := pb.(*mme.WearMechanism)
				if ok {
					incremental.SkinWear = v
				}
			}
		}
	}
	return incremental
}
