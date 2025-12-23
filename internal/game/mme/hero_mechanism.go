package mme

import (
	"gitee.com/orbit-w/meteor/bases/container/xmap"
	dirtyflag "gitee.com/orbit-w/orbit/lib/base/dirty_flag"
	fieldmeta "gitee.com/orbit-w/orbit/lib/base/field_meta"
	"gitee.com/orbit-w/orbit/lib/module/db/mgo_builder"
	mmemodel "gitee.com/orbit-w/orbit/lib/module/mme_model"
	"gitee.com/orbit-w/orbit/pkg/proto/mme"
	"google.golang.org/protobuf/proto"
	"maps"
)

const (
	HeroMechanismFieldIndexId         = uint8(1)
	HeroMechanismFieldIndexConfId     = uint8(2)
	HeroMechanismFieldIndexCreateTime = uint8(3)
	HeroMechanismFieldIndexUseTimes   = uint8(4)
	HeroMechanismFieldIndexSkills     = uint8(5)
)

// Dirty bits for Mechanism fields
const (
	HeroMechanismDirtyIdBit         int64 = 1 << (HeroMechanismFieldIndexId - 1)
	HeroMechanismDirtyConfIdBit     int64 = 1 << (HeroMechanismFieldIndexConfId - 1)
	HeroMechanismDirtyCreateTimeBit int64 = 1 << (HeroMechanismFieldIndexCreateTime - 1)
	HeroMechanismDirtyUseTimesBit   int64 = 1 << (HeroMechanismFieldIndexUseTimes - 1)
	HeroMechanismDirtySkillsBit     int64 = 1 << (HeroMechanismFieldIndexSkills - 1)
)

type HeroMechanism struct {
	Id         int64           `bson:"id"`
	ConfId     int32           `bson:"conf_id"`
	CreateTime int64           `bson:"create_time"`
	UseTimes   int32           `bson:"use_times"`
	Skills     map[int32]int32 `bson:"skills"`
}

func NewHeroMechanism() *HeroMechanism {
	return &HeroMechanism{
		Skills: make(map[int32]int32),
	}
}

func (m *HeroMechanism) DeepCopy(co *HeroMechanism) {
	if m == nil || co == nil {
		return
	}

	*co = *m
	maps.Copy(co.Skills, m.Skills)
}

func (m *HeroMechanism) ToProto() *mme.HeroMechanism {
	if m == nil {
		return nil
	}

	pb := &mme.HeroMechanism{}

	id := m.Id
	pb.Id = &id
	confid := m.ConfId
	pb.ConfId = &confid
	createtime := m.CreateTime
	pb.CreateTime = &createtime
	usetimes := m.UseTimes
	pb.UseTimes = &usetimes
	if m.Skills != nil {
		pb.Skills = make(map[int32]int32, len(m.Skills))
		maps.Copy(pb.Skills, m.Skills)
	}
	return pb
}

func (m *HeroMechanism) FromProto(pb *mme.HeroMechanism) {
	if m == nil || pb == nil {
		return
	}

	if pb.Id != nil {
		m.Id = *pb.Id
	}
	if pb.ConfId != nil {
		m.ConfId = *pb.ConfId
	}
	if pb.CreateTime != nil {
		m.CreateTime = *pb.CreateTime
	}
	if pb.UseTimes != nil {
		m.UseTimes = *pb.UseTimes
	}
	if pb.Skills != nil {
		m.Skills = make(map[int32]int32, len(pb.Skills))
		maps.Copy(m.Skills, pb.Skills)
	}
}

type HeroMechanismWrapper struct {
	data *HeroMechanism
	dirtyflag.IDirtyFlag
	fieldMetas *fieldmeta.FieldMetas

	// Value 为值类型，使用 xmap.MapAccessor 进行包装
	skillsAccessor *xmap.MapAccessor[int32, int32]
}

func NewHeroMechanismWrapper(data *HeroMechanism) *HeroMechanismWrapper {
	if data == nil {
		panic("data is nil")
	}
	w := &HeroMechanismWrapper{
		data:       data,
		IDirtyFlag: dirtyflag.NewDirtyFlag(),
		fieldMetas: fieldmeta.NewFieldMetas(),
	}

	w.skillsAccessor = xmap.NewMapAccessorWithMarker(&w.data.Skills, w, HeroMechanismDirtySkillsBit)
	return w
}

func (w *HeroMechanismWrapper) InitFieldContext() {
	w.fieldMetas.SetFieldType(HeroMechanismFieldIndexId, fieldmeta.FieldTypeSync)
	w.fieldMetas.SetFieldType(HeroMechanismFieldIndexConfId, fieldmeta.FieldTypeSync)
	w.fieldMetas.SetFieldType(HeroMechanismFieldIndexCreateTime, fieldmeta.FieldTypeSync)
	w.fieldMetas.SetFieldType(HeroMechanismFieldIndexUseTimes, fieldmeta.FieldTypeSync)
	w.fieldMetas.SetFieldType(HeroMechanismFieldIndexSkills, fieldmeta.FieldTypeSync)
}

func (w *HeroMechanismWrapper) Name() string {
	return "HeroMechanism"
}

// MatchesAll 判断字段是否匹配所有类型标记
func (w *HeroMechanismWrapper) MatchesAll(fieldID uint8, fieldTypes ...fieldmeta.FieldType) bool {
	return w.fieldMetas.MatchesAll(fieldID, fieldTypes...)
}

func (w *HeroMechanismWrapper) GetId() int64 {
	return w.data.Id
}

func (w *HeroMechanismWrapper) SetId(v int64) {
	w.data.Id = v
	w.MarkDirty(HeroMechanismDirtyIdBit)
}

func (w *HeroMechanismWrapper) GetConfId() int32 {
	return w.data.ConfId
}

func (w *HeroMechanismWrapper) SetConfId(v int32) {
	w.data.ConfId = v
	w.MarkDirty(HeroMechanismDirtyConfIdBit)
}

func (w *HeroMechanismWrapper) GetCreateTime() int64 {
	return w.data.CreateTime
}

func (w *HeroMechanismWrapper) SetCreateTime(v int64) {
	w.data.CreateTime = v
	w.MarkDirty(HeroMechanismDirtyCreateTimeBit)
}

func (w *HeroMechanismWrapper) GetUseTimes() int32 {
	return w.data.UseTimes
}

func (w *HeroMechanismWrapper) SetUseTimes(v int32) {
	w.data.UseTimes = v
	w.MarkDirty(HeroMechanismDirtyUseTimesBit)
}

// 包装器-获取Skills访问器
func (w *HeroMechanismWrapper) GetSkillsAccessor() *xmap.MapAccessor[int32, int32] {
	return w.skillsAccessor
}

// ClearAllDirtyFlags 清除所有脏标记位
func (w *HeroMechanismWrapper) ClearAllDirtyFlags() {
	w.IDirtyFlag.ClearAllDirty()

	// 如果Value为引用类型且有脏标记，则清除所有xmap中Value的脏标记

	// 清除所有xmap中的操作记录
	w.skillsAccessor.ResetOperations()
}

// BuildMongoUpdate 构建MongoDB更新操作
func (w *HeroMechanismWrapper) BuildMongoUpdate(builder *mgo_builder.MongoUpdateBuilder, path *mgo_builder.NestedPath) {
	if w == nil {
		return
	}

	if w.IsDirty(HeroMechanismDirtyIdBit) {
		builder.SetNestedPath(path, "id", w.GetId())
	}
	if w.IsDirty(HeroMechanismDirtyConfIdBit) {
		builder.SetNestedPath(path, "conf_id", w.GetConfId())
	}
	if w.IsDirty(HeroMechanismDirtyCreateTimeBit) {
		builder.SetNestedPath(path, "create_time", w.GetCreateTime())
	}
	if w.IsDirty(HeroMechanismDirtyUseTimesBit) {
		builder.SetNestedPath(path, "use_times", w.GetUseTimes())
	}
	if w.IsDirty(HeroMechanismDirtySkillsBit) {
		builder.SetNestedPath(path, "skills", w.skillsAccessor.Clone())
	}
}

// ToIncrementalProtoWithContext 根据脏标记位构建增量数据的 protoMessage
// 只返回标记为脏的字段数据，用于增量同步
func (w *HeroMechanismWrapper) ToIncrementalProtoWithContext(ctx mmemodel.SyncContext) proto.Message {
	if w == nil {
		return nil
	}

	// 如果没有脏标记，返回 nil
	if !w.HasAnyDirty() {
		return nil
	}

	incremental := &mme.HeroMechanism{}

	if mmemodel.FieldCanBeIncrementalSynced(w, HeroMechanismDirtyIdBit, HeroMechanismFieldIndexId, ctx) {
		v := w.GetId()
		incremental.Id = &v
	}
	if mmemodel.FieldCanBeIncrementalSynced(w, HeroMechanismDirtyConfIdBit, HeroMechanismFieldIndexConfId, ctx) {
		v := w.GetConfId()
		incremental.ConfId = &v
	}
	if mmemodel.FieldCanBeIncrementalSynced(w, HeroMechanismDirtyCreateTimeBit, HeroMechanismFieldIndexCreateTime, ctx) {
		v := w.GetCreateTime()
		incremental.CreateTime = &v
	}
	if mmemodel.FieldCanBeIncrementalSynced(w, HeroMechanismDirtyUseTimesBit, HeroMechanismFieldIndexUseTimes, ctx) {
		v := w.GetUseTimes()
		incremental.UseTimes = &v
	}
	if mmemodel.FieldCanBeIncrementalSynced(w, HeroMechanismDirtySkillsBit, HeroMechanismFieldIndexSkills, ctx) {
		incremental.Skills = w.skillsAccessor.Clone()
	}
	return incremental
}

// ToProto 将 HeroMechanism 数据转换为完整的 protobuf 结构体
func (w *HeroMechanismWrapper) ToProto() *mme.HeroMechanism {
	if w == nil || w.data == nil {
		return nil
	}

	return w.data.ToProto()
}

// FromProto 从 protobuf 结构体加载数据到 HeroMechanism
func (w *HeroMechanismWrapper) FromProto(pb *mme.HeroMechanism) {
	if w == nil || w.data == nil || pb == nil {
		return
	}

	if pb.Id != nil {
		w.SetId(*pb.Id)
	}
	if pb.ConfId != nil {
		w.SetConfId(*pb.ConfId)
	}
	if pb.CreateTime != nil {
		w.SetCreateTime(*pb.CreateTime)
	}
	if pb.UseTimes != nil {
		w.SetUseTimes(*pb.UseTimes)
	}
	if pb.Skills != nil {
		// 清空现有的数据
		temp := make(map[int32]int32, len(pb.Skills))
		maps.Copy(temp, pb.Skills)
		// 设置新的数据，并清空所有变化操作记录
		w.skillsAccessor.Reset(&temp)
	}
}

// 包装器-深拷贝
func (w *HeroMechanismWrapper) DeepCopy() *HeroMechanism {
	if w == nil {
		return nil
	}
	copy := &HeroMechanism{}
	w.DeepCopyTo(copy)
	return copy
}

// 包装器-深拷贝
func (w *HeroMechanismWrapper) DeepCopyTo(copy *HeroMechanism) {
	if w == nil || w.data == nil || copy == nil {
		return
	}

	// 拷贝基础类型字段
	*copy = *w.data

	// 拷贝 SkillsAccessor
	if w.skillsAccessor != nil {
		if copy.Skills == nil {
			copy.Skills = make(map[int32]int32, w.skillsAccessor.Len())
		}
		w.skillsAccessor.Copy(copy.Skills)
	}
}
