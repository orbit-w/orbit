package mme

import (
	"strings"

	"gitee.com/orbit-w/meteor/bases/dirty/dirty_tracker"
	"gitee.com/orbit-w/meteor/bases/dirty/xmap"
	"gitee.com/orbit-w/orbit/app/proto/mme"
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
	*mme.HeroMechanism
	dirty_tracker.DirtyTracker

	SkillsAccessor xmap.MapAccessor[int32, int32]
}

func NewHeroMechanism() *HeroMechanism {
	hm := &HeroMechanism{
		HeroMechanism: new(mme.HeroMechanism),
	}

	hm.SkillsAccessor = xmap.NewMapAccessorWithMarker(&hm.Skills, hm, HeroMechanismDirtySkillsBit)
	return hm
}

// 机制唯一名称
func (m *HeroMechanism) Name() string {
	return "HeroMechanism"
}

func (m *HeroMechanism) Link(parent *dirty_tracker.DirtyTracker, parentBit int64) {
	m.DirtyTracker.Link(parent, parentBit)
}

func (m *HeroMechanism) SetId(v int64) {
	m.Id = &v
	m.MarkDirty(HeroMechanismDirtyIdBit)
}

func (m *HeroMechanism) SetConfId(v int32) {
	m.ConfId = &v
	m.MarkDirty(LevelUpMechanismDirtyConfIdBit)
}

func (m *HeroMechanism) SetUseTimes(v int32) {
	m.UseTimes = &v
	m.MarkDirty(HeroMechanismDirtyUseTimesBit)
}

func (m *HeroMechanism) SetCreateTime(v int64) {
	m.CreateTime = &v
	m.MarkDirty(HeroMechanismDirtyCreateTimeBit)
}

func (m *HeroMechanism) GetSkillAccessor() xmap.MapAccessor[int32, int32] {
	return m.SkillsAccessor
}

// ClearAllDirtyFlags 清除所有脏标记位
func (m *HeroMechanism) ClearAllDirtyFlags() {
	m.DirtyTracker.ClearAllDirty()
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
		{"_id", HeroMechanismDirtyIdBit, m.Id},
		{"conf_id", LevelUpMechanismDirtyConfIdBit, m.ConfId},
		{"create_time", HeroMechanismDirtyCreateTimeBit, m.CreateTime},
		{"use_times", HeroMechanismDirtyUseTimesBit, m.UseTimes},
		{"skills", HeroMechanismDirtySkillsBit, m.Skills},
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

// DeepCopy creates a deep copy of HeroMechanism
func (m *HeroMechanism) DeepCopy(co *HeroMechanism) {
	if m == nil {
		return
	}

	// 创建新的 protobuf 对象
	co.HeroMechanism = &mme.HeroMechanism{}

	// 深拷贝指针字段
	if m.Id != nil {
		v := *m.Id
		co.Id = &v
	}

	if m.ConfId != nil {
		v := *m.ConfId
		co.ConfId = &v
	}

	if m.CreateTime != nil {
		v := *m.CreateTime
		co.CreateTime = &v
	}

	if m.UseTimes != nil {
		v := *m.UseTimes
		co.UseTimes = &v
	}

	// 深拷贝 Skills map
	if m.Skills != nil {
		co.Skills = make(map[int32]int32, len(m.Skills))
		for k, v := range m.Skills {
			co.Skills[k] = v
		}
	}

	// 拷贝 XXXId（非指针）
	co.XXXId = m.XXXId

	// 重新创建 SkillsAccessor
	co.SkillsAccessor = xmap.NewMapAccessorWithMarker(&co.Skills, co, HeroMechanismDirtySkillsBit)

	// 拷贝 DirtyTracker 状态
	co.DirtyTracker = m.DirtyTracker
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

	if m.IsDirty(HeroMechanismDirtyXXXIdBit) {
		incremental.XXXId = m.XXXId
	}

	// 根据脏标记位设置对应的字段
	if m.IsDirty(HeroMechanismDirtyIdBit) {
		incremental.Id = m.Id
	}

	if m.IsDirty(LevelUpMechanismDirtyConfIdBit) {
		incremental.ConfId = m.ConfId
	}

	if m.IsDirty(HeroMechanismDirtyCreateTimeBit) {
		incremental.CreateTime = m.CreateTime
	}

	if m.IsDirty(HeroMechanismDirtyUseTimesBit) {
		incremental.UseTimes = m.UseTimes
	}

	if m.IsDirty(HeroMechanismDirtySkillsBit) {
		incremental.XXXChange_Skills = make([]*mme.XXXChange_Skills, 0)
		m.SkillsAccessor.RangeOperations(func(key int32, operation xmap.MapOperation[int32]) bool {
			switch operation.Type {
			case xmap.SetOperation:
				v, _ := m.SkillsAccessor.Get(key)
				incremental.XXXChange_Skills = append(incremental.XXXChange_Skills, &mme.XXXChange_Skills{
					Key:   key,
					Value: v,
				})
			case xmap.DeleteOperation:
				incremental.XXXChange_Skills = append(incremental.XXXChange_Skills, &mme.XXXChange_Skills{
					Key: key,
				})
			}
			return true
		})
	}

	return incremental
}
