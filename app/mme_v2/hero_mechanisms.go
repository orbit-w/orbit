package mme

import (
	"strings"

	"gitee.com/orbit-w/meteor/bases/dirty/xmap"
	"gitee.com/orbit-w/orbit/app/proto/mme"
	dirtyflag "gitee.com/orbit-w/orbit/lib/base/dirty_flag"
	fieldmeta "gitee.com/orbit-w/orbit/lib/base/field_meta"
	"gitee.com/orbit-w/orbit/lib/module/db/mgo_builder"
	mmeutils "gitee.com/orbit-w/orbit/lib/module/mme_utils"
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

type HeroMechanismWrapper struct {
	data *HeroMechanism
	dirtyflag.IDirtyFlag
	fieldMetas *fieldmeta.FieldMetas

	skillsAccessor xmap.MapAccessor[int32, int32]
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

func (m *HeroMechanismWrapper) GetSkillAccessor() xmap.MapAccessor[int32, int32] {
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
func (m *HeroMechanismWrapper) BuildMongoUpdate(builder *mgo_builder.MongoUpdateBuilder, prefix string) {
	if m == nil {
		return
	}

	// 定义字段配置，避免重复代码
	fieldConfigs := []struct {
		fieldName string
		dirtyBit  int64
		value     any
	}{
		{"_id", HeroMechanismDirtyIdBit, m.GetId()},
		{"conf_id", HeroMechanismDirtyConfIdBit, m.GetConfId()},
		{"create_time", HeroMechanismDirtyCreateTimeBit, m.GetCreateTime()},
		{"use_times", HeroMechanismDirtyUseTimesBit, m.GetUseTimes()},
		{"skills", HeroMechanismDirtySkillsBit, m.skillsAccessor.Clone()},
	}

	// 使用 strings.Builder 优化路径构建性能
	var pathBuilder strings.Builder
	hasPrefix := prefix != ""

	// 预分配容量，减少内存重新分配
	if hasPrefix {
		pathBuilder.Grow(len(prefix) + 20) // 预估最大路径长度
	}

	// 批量处理字段更新
	for _, config := range fieldConfigs {
		if m.IsDirty(config.dirtyBit) {
			pathBuilder.Reset()

			if hasPrefix {
				pathBuilder.WriteString(prefix)
				pathBuilder.WriteByte('.')
			}
			pathBuilder.WriteString(config.fieldName)

			builder.Set(pathBuilder.String(), config.value)
		}
	}
}

func (m *HeroMechanismWrapper) DeepCopy(copy *HeroMechanism) {
	if m == nil || m.data == nil || copy == nil {
		return
	}

	// 拷贝基础类型字段
	*copy = *m.data

	m.skillsAccessor.DeepCopy(copy.Skills)
}

func (m *HeroMechanismWrapper) MatchesAll(fieldID uint8, fieldTypes ...fieldmeta.FieldType) bool {
	return m.fieldMetas.MatchesAll(fieldID, fieldTypes...)
}

// ToProto 将 HeroMechanism 数据转换为完整的 protobuf 结构体
func (m *HeroMechanismWrapper) ToProto() *mme.HeroMechanism {
	if m == nil || m.data == nil {
		return nil
	}

	pb := &mme.HeroMechanism{}

	// 设置所有字段
	id := m.data.Id
	pb.Id = &id

	confId := m.data.ConfId
	pb.ConfId = &confId

	createTime := m.data.CreateTime
	pb.CreateTime = &createTime

	useTimes := m.data.UseTimes
	pb.UseTimes = &useTimes

	// 转换 Skills map
	if m.data.Skills != nil {
		pb.Skills = make(map[int32]int32, len(m.data.Skills))
		for k, v := range m.data.Skills {
			pb.Skills[k] = v
		}
	}

	return pb
}

// ToIncrementalProto 根据脏标记位构建增量数据的 protoMessage
// 只返回标记为脏的字段数据，用于增量同步
func (m *HeroMechanismWrapper) ToIncrementalProto(ctx mmeutils.SyncContext) proto.Message {
	if m == nil {
		return nil
	}

	// 如果没有脏标记，返回 nil
	if !m.HasAnyDirty() {
		return nil
	}

	incremental := &mme.HeroMechanism{}

	// 根据脏标记位设置对应的字段
	if mmeutils.FieldCanBeIncrementalSynced(m, HeroMechanismDirtyIdBit, HeroMechanismFieldIndexId, ctx) {
		v := m.GetId()
		incremental.Id = &v
	}

	if mmeutils.FieldCanBeIncrementalSynced(m, HeroMechanismDirtyConfIdBit, HeroMechanismFieldIndexConfId, ctx) {
		v := m.GetConfId()
		incremental.ConfId = &v
	}

	if mmeutils.FieldCanBeIncrementalSynced(m, HeroMechanismDirtyCreateTimeBit, HeroMechanismFieldIndexCreateTime, ctx) {
		v := m.GetCreateTime()
		incremental.CreateTime = &v
	}

	if mmeutils.FieldCanBeIncrementalSynced(m, HeroMechanismDirtyUseTimesBit, HeroMechanismFieldIndexUseTimes, ctx) {
		v := m.GetUseTimes()
		incremental.UseTimes = &v
	}

	if mmeutils.FieldCanBeIncrementalSynced(m, HeroMechanismDirtySkillsBit, HeroMechanismFieldIndexSkills, ctx) {
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
