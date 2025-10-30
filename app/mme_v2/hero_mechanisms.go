package mme

import (
	"maps"

	"gitee.com/orbit-w/meteor/bases/dirty/xmap"
	"gitee.com/orbit-w/orbit/app/proto/mme"
	dirtyflag "gitee.com/orbit-w/orbit/lib/base/dirty_flag"
	fieldmeta "gitee.com/orbit-w/orbit/lib/base/field_meta"
	"gitee.com/orbit-w/orbit/lib/module/db/mgo_builder"
	mmemodel "gitee.com/orbit-w/orbit/lib/module/mme_model"
	"google.golang.org/protobuf/proto"
)

const (
	HeroMechanismFieldIndexId = uint8(0)
	HeroMechanismFieldIndexConfId
	HeroMechanismFieldIndexCreateTime
	HeroMechanismFieldIndexUseTimes
	HeroMechanismFieldIndexSkills
)

// Dirty bits for Mechanism fields
const (
	HeroMechanismDirtyIdBit         int64 = 1 << HeroMechanismFieldIndexId
	HeroMechanismDirtyConfIdBit     int64 = 1 << HeroMechanismFieldIndexConfId
	HeroMechanismDirtyCreateTimeBit int64 = 1 << HeroMechanismFieldIndexCreateTime
	HeroMechanismDirtyUseTimesBit   int64 = 1 << HeroMechanismFieldIndexUseTimes
	HeroMechanismDirtySkillsBit     int64 = 1 << HeroMechanismFieldIndexSkills
)

type HeroMechanism struct {
	Id         int64           `bson:"_id"`
	ConfId     int32           `bson:"conf_id"`
	CreateTime int64           `bson:"create_time"`
	UseTimes   int32           `bson:"use_times"`
	Skills     map[int32]int32 `bson:"skills"`
}

// 数据-深拷贝
func (m *HeroMechanism) DeepCopy(co *HeroMechanism) {
	if m == nil || co == nil {
		return
	}

	*co = *m
	maps.Copy(co.Skills, m.Skills)
}

// 数据-转换为protobuf
func (m *HeroMechanism) ToProto() *mme.HeroMechanism {
	if m == nil {
		return nil
	}

	pb := &mme.HeroMechanism{}

	id := m.Id
	pb.Id = &id

	confId := m.ConfId
	pb.ConfId = &confId

	createTime := m.CreateTime
	pb.CreateTime = &createTime

	useTimes := m.UseTimes
	pb.UseTimes = &useTimes

	if m.Skills != nil {
		pb.Skills = make(map[int32]int32, len(m.Skills))
		maps.Copy(pb.Skills, m.Skills)
	}

	return pb
}

type HeroMechanismWrapper struct {
	data *HeroMechanism
	dirtyflag.IDirtyFlag
	fieldMetas *fieldmeta.FieldMetas

	//Value为值类型，使用xmap.MapAccessor进行包装
	skillsAccessor *xmap.MapAccessor[int32, int32]
}

func NewHeroMechanismWrapper(data *HeroMechanism) *HeroMechanismWrapper {
	if data == nil {
		panic("pt is nil")
	}
	hm := &HeroMechanismWrapper{
		data:       data,
		IDirtyFlag: dirtyflag.NewDirtyFlag(),
		fieldMetas: fieldmeta.NewFieldMetas(),
	}

	hm.skillsAccessor = xmap.NewMapAccessorWithMarker(&hm.data.Skills, hm, HeroMechanismDirtySkillsBit)
	return hm
}

func (m *HeroMechanismWrapper) InitFieldContext() {
	m.fieldMetas.SetFieldType(HeroMechanismFieldIndexId, fieldmeta.FieldTypeSync)
	m.fieldMetas.SetFieldType(HeroMechanismFieldIndexConfId, fieldmeta.FieldTypeSync)
	m.fieldMetas.SetFieldType(HeroMechanismFieldIndexUseTimes, fieldmeta.FieldTypeSync)
	m.fieldMetas.SetFieldType(HeroMechanismFieldIndexSkills, fieldmeta.FieldTypeSync)
}

// 机制唯一名称
func (m *HeroMechanismWrapper) Name() string {
	return "HeroMechanism"
}

func (m *HeroMechanismWrapper) GetId() int64 {
	return m.data.Id
}

func (m *HeroMechanismWrapper) GetConfId() int32 {
	return m.data.ConfId
}

func (m *HeroMechanismWrapper) GetUseTimes() int32 {
	return m.data.UseTimes
}

func (m *HeroMechanismWrapper) GetCreateTime() int64 {
	return m.data.CreateTime
}

func (m *HeroMechanismWrapper) SetId(v int64) {
	m.data.Id = v
	m.MarkDirty(HeroMechanismDirtyIdBit)
}

func (m *HeroMechanismWrapper) SetConfId(v int32) {
	m.data.ConfId = v
	m.MarkDirty(HeroMechanismDirtyConfIdBit)
}

func (m *HeroMechanismWrapper) SetUseTimes(v int32) {
	m.data.UseTimes = v
	m.MarkDirty(HeroMechanismDirtyUseTimesBit)
}

func (m *HeroMechanismWrapper) SetCreateTime(v int64) {
	m.data.CreateTime = v
	m.MarkDirty(HeroMechanismDirtyCreateTimeBit)
}

// 包装器-获取技能访问器
func (m *HeroMechanismWrapper) GetSkillAccessor() *xmap.MapAccessor[int32, int32] {
	return m.skillsAccessor
}

// ClearAllDirtyFlags 清除所有脏标记位
func (m *HeroMechanismWrapper) ClearAllDirtyFlags() {
	m.ClearAllDirty()

	// 如果Value为引用类型且有脏标记，则清除所有xmap中Value的脏标记

	// 清除所有xmap中的操作记录
	m.skillsAccessor.ResetOperations()
}

// BuildMongoUpdate 构建MongoDB更新操作
func (m *HeroMechanismWrapper) BuildMongoUpdate(builder *mgo_builder.MongoUpdateBuilder, path *mgo_builder.NestedPath) {
	if m == nil {
		return
	}

	if m.IsDirty(HeroMechanismDirtyIdBit) {
		builder.SetNestedPath(path, "_id", m.GetId())
	}
	if m.IsDirty(HeroMechanismDirtyConfIdBit) {
		builder.SetNestedPath(path, "conf_id", m.GetConfId())
	}
	if m.IsDirty(HeroMechanismDirtyCreateTimeBit) {
		builder.SetNestedPath(path, "create_time", m.GetCreateTime())
	}
	if m.IsDirty(HeroMechanismDirtyUseTimesBit) {
		builder.SetNestedPath(path, "use_times", m.GetUseTimes())
	}
	if m.IsDirty(HeroMechanismDirtySkillsBit) {
		builder.SetNestedPath(path, "skills", m.skillsAccessor.Clone())
	}
}

// 包装器-深拷贝
func (m *HeroMechanismWrapper) DeepCopy() *HeroMechanism {
	if m == nil {
		return nil
	}
	copy := &HeroMechanism{}
	m.DeepCopyTo(copy)
	return copy
}

// 包装器-深拷贝
func (m *HeroMechanismWrapper) DeepCopyTo(copy *HeroMechanism) {
	if m == nil || m.data == nil || copy == nil {
		return
	}

	// 拷贝基础类型字段
	*copy = *m.data

	// Value 为引用类型，需要深拷贝
	// Value 为值类型，浅拷贝即可
	m.skillsAccessor.Copy(copy.Skills)
}

// MatchesAll 判断字段是否匹配所有类型标记
func (m *HeroMechanismWrapper) MatchesAll(fieldID uint8, fieldTypes ...fieldmeta.FieldType) bool {
	return m.fieldMetas.MatchesAll(fieldID, fieldTypes...)
}

// ToProto 将 Mechanism 数据转换为完整的 protobuf 结构体
func (m *HeroMechanismWrapper) ToProto() *mme.HeroMechanism {
	if m == nil || m.data == nil {
		return nil
	}

	return m.data.ToProto()
}

// FromProto 从 protobuf 结构体加载数据到 Mechanism
func (m *HeroMechanismWrapper) FromProto(pb *mme.HeroMechanism) {
	if m == nil || m.data == nil || pb == nil {
		return
	}

	// 加载基础类型字段
	if pb.Id != nil {
		m.SetId(*pb.Id)
	}

	if pb.ConfId != nil {
		m.SetConfId(*pb.ConfId)
	}

	if pb.CreateTime != nil {
		m.SetCreateTime(*pb.CreateTime)
	}

	if pb.UseTimes != nil {
		m.SetUseTimes(*pb.UseTimes)
	}

	// 加载 Skills map
	if pb.Skills != nil {
		// 清空现有的 Skills
		temp := make(map[int32]int32, len(pb.Skills))
		maps.Copy(temp, pb.Skills)
		// 设置新的 Skills数据，并清空所有变化操作记录
		m.skillsAccessor.Reset(&temp)
	}
}

// ToIncrementalProto 根据脏标记位构建增量数据的 protoMessage
// 只返回标记为脏的字段数据，用于增量同步
func (m *HeroMechanismWrapper) ToIncrementalProtoWithContext(ctx mmemodel.SyncContext) proto.Message {
	if m == nil {
		return nil
	}

	// 如果没有脏标记，返回 nil
	if !m.HasAnyDirty() {
		return nil
	}

	incremental := &mme.HeroMechanism{}

	// 根据脏标记位设置对应的字段
	if mmemodel.FieldCanBeIncrementalSynced(m, HeroMechanismDirtyIdBit, HeroMechanismFieldIndexId, ctx) {
		v := m.GetId()
		incremental.Id = &v
	}

	if mmemodel.FieldCanBeIncrementalSynced(m, HeroMechanismDirtyConfIdBit, HeroMechanismFieldIndexConfId, ctx) {
		v := m.GetConfId()
		incremental.ConfId = &v
	}

	if mmemodel.FieldCanBeIncrementalSynced(m, HeroMechanismDirtyCreateTimeBit, HeroMechanismFieldIndexCreateTime, ctx) {
		v := m.GetCreateTime()
		incremental.CreateTime = &v
	}

	if mmemodel.FieldCanBeIncrementalSynced(m, HeroMechanismDirtyUseTimesBit, HeroMechanismFieldIndexUseTimes, ctx) {
		v := m.GetUseTimes()
		incremental.UseTimes = &v
	}

	if mmemodel.FieldCanBeIncrementalSynced(m, HeroMechanismDirtySkillsBit, HeroMechanismFieldIndexSkills, ctx) {
		incremental.Skills_XXXChangeList = make([]*mme.HeroMechanism_Skills_XXXMapChangeRecord, 0)
		m.skillsAccessor.RangeOperations(func(key int32, operation xmap.MapOperation[int32]) bool {
			switch operation.Type {
			case xmap.SetOperation:
				v, _ := m.skillsAccessor.Get(key)
				incremental.Skills_XXXChangeList = append(incremental.Skills_XXXChangeList, &mme.HeroMechanism_Skills_XXXMapChangeRecord{
					ChangeType: mme.ChangeType_CHANGE_TYPE_SET,
					Key:        key,
					Value:      v,
				})
			case xmap.DeleteOperation:
				incremental.Skills_XXXChangeList = append(incremental.Skills_XXXChangeList, &mme.HeroMechanism_Skills_XXXMapChangeRecord{
					ChangeType: mme.ChangeType_CHANGE_TYPE_DELETE,
					Key:        key,
				})
			}
			return true
		})
	}

	return incremental
}
