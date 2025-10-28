package mme

import (
	"maps"
	"strings"

	"gitee.com/orbit-w/meteor/bases/dirty/xmap"
	"gitee.com/orbit-w/orbit/app/proto/mme"
	dirtyflag "gitee.com/orbit-w/orbit/lib/base/dirty_flag"
	"gitee.com/orbit-w/orbit/lib/module/db/mgo_builder"
	"google.golang.org/protobuf/proto"
)

// Dirty bits for Mechanism fields
const (
	HeroMechanismDirtyXXXIdBit      int64 = 1 << 0
	HeroMechanismDirtyIdBit         int64 = 1 << 1
	HeroMechanismDirtyConfIdBit     int64 = 1 << 2
	HeroMechanismDirtyCreateTimeBit int64 = 1 << 3
	HeroMechanismDirtyUseTimesBit   int64 = 1 << 4
	HeroMechanismDirtySkillsBit     int64 = 1 << 5
)

type HeroMechanism struct {
	Id         int64           `bson:"_id"`
	ConfId     int32           `bson:"conf_id"`
	CreateTime int64           `bson:"create_time"`
	UseTimes   int32           `bson:"use_times"`
	Skills     map[int32]int32 `bson:"skills"`
}

type HeroMechanismAccessor struct {
	impl *HeroMechanism
	dirtyflag.IDirtyFlag

	skillsAccessor xmap.MapAccessor[int32, int32]
}

func NewHeroMechanismAccessor(impl *HeroMechanism) *HeroMechanismAccessor {
	if impl == nil {
		panic("pt is nil")
	}
	hm := &HeroMechanismAccessor{
		impl:       impl,
		IDirtyFlag: dirtyflag.NewDirtyFlag(),
	}

	hm.skillsAccessor = xmap.NewMapAccessorWithMarker(&hm.impl.Skills, hm, HeroMechanismDirtySkillsBit)
	return hm
}

// 机制唯一名称
func (m *HeroMechanismAccessor) Name() string {
	return "HeroMechanism"
}

func (m *HeroMechanismAccessor) GetId() int64 {
	return m.impl.Id
}

func (m *HeroMechanismAccessor) GetConfId() int32 {
	return m.impl.ConfId
}

func (m *HeroMechanismAccessor) GetUseTimes() int32 {
	return m.impl.UseTimes
}

func (m *HeroMechanismAccessor) GetCreateTime() int64 {
	return m.impl.CreateTime
}

func (m *HeroMechanismAccessor) SetId(v int64) {
	m.impl.Id = v
	m.MarkDirty(HeroMechanismDirtyIdBit)
}

func (m *HeroMechanismAccessor) SetConfId(v int32) {
	m.impl.ConfId = v
	m.MarkDirty(HeroMechanismDirtyConfIdBit)
}

func (m *HeroMechanismAccessor) SetUseTimes(v int32) {
	m.impl.UseTimes = v
	m.MarkDirty(HeroMechanismDirtyUseTimesBit)
}

func (m *HeroMechanismAccessor) SetCreateTime(v int64) {
	m.impl.CreateTime = v
	m.MarkDirty(HeroMechanismDirtyCreateTimeBit)
}

func (m *HeroMechanismAccessor) GetSkills() map[int32]int32 {
	return m.impl.Skills
}

func (m *HeroMechanismAccessor) GetSkillAccessor() xmap.MapAccessor[int32, int32] {
	return m.skillsAccessor
}

// ClearAllDirtyFlags 清除所有脏标记位
func (m *HeroMechanismAccessor) ClearAllDirtyFlags() {
	m.ClearAllDirty()

	// 如果Value为引用类型且有脏标记，则清除所有xmap中Value的脏标记

	// 清除所有xmap中的操作记录
	m.skillsAccessor.ResetOperations()
}

// BuildMongoUpdate 构建MongoDB更新操作
func (m *HeroMechanismAccessor) BuildMongoUpdate(builder *mgo_builder.MongoUpdateBuilder, prefix string) {
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
		{"skills", HeroMechanismDirtySkillsBit, m.GetSkills()},
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

func (m *HeroMechanismAccessor) DeepCopy(co *HeroMechanism) {
	if m == nil {
		return
	}

	*co = *m.impl

	co.Skills = make(map[int32]int32, len(m.impl.Skills))
	maps.Copy(co.Skills, m.impl.Skills)
}

// ToIncrementalProto 根据脏标记位构建增量数据的 protoMessage
// 只返回标记为脏的字段数据，用于增量同步
func (m *HeroMechanismAccessor) ToIncrementalProto() proto.Message {
	if m == nil {
		return nil
	}

	// 如果没有脏标记，返回 nil
	if !m.HasAnyDirty() {
		return nil
	}

	incremental := &mme.HeroMechanism{}

	// 根据脏标记位设置对应的字段
	if m.IsDirty(HeroMechanismDirtyIdBit) {
		v := m.GetId()
		incremental.Id = &v
	}

	if m.IsDirty(HeroMechanismDirtyConfIdBit) {
		v := m.GetConfId()
		incremental.ConfId = &v
	}

	if m.IsDirty(HeroMechanismDirtyCreateTimeBit) {
		v := m.GetCreateTime()
		incremental.CreateTime = &v
	}

	if m.IsDirty(HeroMechanismDirtyUseTimesBit) {
		v := m.GetUseTimes()
		incremental.UseTimes = &v
	}

	if m.IsDirty(HeroMechanismDirtySkillsBit) {
		incremental.Skills_XXXChangeList = make([]*mme.HeroMechanism_Skills_XXXMapChangeRecord, 0)
		m.skillsAccessor.RangeOperations(func(key int32, operation xmap.MapOperation[int32]) bool {
			switch operation.Type {
			case xmap.SetOperation:
				v, _ := m.skillsAccessor.Get(key)
				incremental.Skills_XXXChangeList = append(incremental.Skills_XXXChangeList, &mme.HeroMechanism_Skills_XXXMapChangeRecord{
					Key:   key,
					Value: v,
				})
			case xmap.DeleteOperation:
				incremental.Skills_XXXChangeList = append(incremental.Skills_XXXChangeList, &mme.HeroMechanism_Skills_XXXMapChangeRecord{
					Key: key,
				})
			}
			return true
		})
	}

	return incremental
}

// ToFullProto 生成全量的 mme.HeroMechanism 数据
// 返回包含所有字段数据的 protoMessage，用于全量同步
func (m *HeroMechanismAccessor) ToFullProto() proto.Message {
	if m == nil {
		return nil
	}

	full := &mme.HeroMechanism{}

	// 设置所有字段
	id := m.GetId()
	full.Id = &id

	confId := m.GetConfId()
	full.ConfId = &confId

	createTime := m.GetCreateTime()
	full.CreateTime = &createTime

	useTimes := m.GetUseTimes()
	full.UseTimes = &useTimes

	// 复制 Skills map
	skills := m.GetSkills()
	if len(skills) > 0 {
		full.Skills = make(map[int32]int32, len(skills))
		for k, v := range skills {
			full.Skills[k] = v
		}
	}

	return full
}
