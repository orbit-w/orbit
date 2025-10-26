package mme

import (
	"strings"

	"gitee.com/orbit-w/meteor/bases/dirty/xmap"
	"gitee.com/orbit-w/orbit/app/proto/mme"
	dirtyflag "gitee.com/orbit-w/orbit/lib/base/dirty_flag"
	"gitee.com/orbit-w/orbit/lib/module/db/mgo_builder"
	"github.com/gogo/protobuf/proto"
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
	heroMechanism *mme.HeroMechanism
	dirtyflag.IDirtyFlag

	skillsAccessor xmap.MapAccessor[int32, int32]
}

func NewHeroMechanism(pt *mme.HeroMechanism) *HeroMechanism {
	if pt == nil {
		panic("pt is nil")
	}
	hm := &HeroMechanism{
		heroMechanism: pt,
		IDirtyFlag:    dirtyflag.NewDirtyFlag(),
	}

	hm.skillsAccessor = xmap.NewMapAccessorWithMarker(&hm.heroMechanism.Skills, hm, HeroMechanismDirtySkillsBit)
	return hm
}

// 机制唯一名称
func (m *HeroMechanism) Name() string {
	return "HeroMechanism"
}

func (m *HeroMechanism) GetId() int64 {
	return *m.heroMechanism.Id
}

func (m *HeroMechanism) GetConfId() int32 {
	return *m.heroMechanism.ConfId
}

func (m *HeroMechanism) GetUseTimes() int32 {
	return *m.heroMechanism.UseTimes
}

func (m *HeroMechanism) GetCreateTime() int64 {
	return *m.heroMechanism.CreateTime
}

func (m *HeroMechanism) SetId(v int64) {
	m.heroMechanism.Id = &v
	m.MarkDirty(HeroMechanismDirtyIdBit)
}

func (m *HeroMechanism) SetConfId(v int32) {
	m.heroMechanism.ConfId = &v
	m.MarkDirty(LevelUpMechanismDirtyConfIdBit)
}

func (m *HeroMechanism) SetUseTimes(v int32) {
	m.heroMechanism.UseTimes = &v
	m.MarkDirty(HeroMechanismDirtyUseTimesBit)
}

func (m *HeroMechanism) SetCreateTime(v int64) {
	m.heroMechanism.CreateTime = &v
	m.MarkDirty(HeroMechanismDirtyCreateTimeBit)
}

func (m *HeroMechanism) GetSkills() map[int32]int32 {
	return m.heroMechanism.Skills
}

func (m *HeroMechanism) GetSkillAccessor() xmap.MapAccessor[int32, int32] {
	return m.skillsAccessor
}

// ClearAllDirtyFlags 清除所有脏标记位
func (m *HeroMechanism) ClearAllDirtyFlags() {
	m.ClearAllDirty()

	// 如果Value为引用类型且有脏标记，则清除所有xmap中Value的脏标记

	// 清除所有xmap中的操作记录
	m.skillsAccessor.ResetOperations()
}

// BuildMongoUpdate 构建MongoDB更新操作
func (m *HeroMechanism) BuildMongoUpdate(builder *mgo_builder.MongoUpdateBuilder, prefix string) {
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

// DeepCopy creates a deep copy of HeroMechanism proto data only
// 手写实现，性能优于 proto.Clone（避免反射开销）
func (m *HeroMechanism) DeepCopy(co *mme.HeroMechanism) {
	if m == nil || co == nil {
		return
	}

	// 深拷贝指针字段
	if m.heroMechanism.Id != nil {
		v := *m.heroMechanism.Id
		co.Id = &v
	}

	if m.heroMechanism.ConfId != nil {
		v := *m.heroMechanism.ConfId
		co.ConfId = &v
	}

	if m.heroMechanism.CreateTime != nil {
		v := *m.heroMechanism.CreateTime
		co.CreateTime = &v
	}

	if m.heroMechanism.UseTimes != nil {
		v := *m.heroMechanism.UseTimes
		co.UseTimes = &v
	}

	// 深拷贝 Skills map
	if m.heroMechanism.Skills != nil {
		co.Skills = make(map[int32]int32, len(m.heroMechanism.Skills))
		for k, v := range m.heroMechanism.Skills {
			co.Skills[k] = v
		}
	}
}

// ToIncrementalProto 根据脏标记位构建增量数据的 protoMessage
// 只返回标记为脏的字段数据，用于增量同步
func (m *HeroMechanism) ToIncrementalProto() proto.Message {
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

	if m.IsDirty(LevelUpMechanismDirtyConfIdBit) {
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
