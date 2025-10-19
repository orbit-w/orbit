package mme

import (
	"gitee.com/orbit-w/orbit/app/proto/mme"
	"gitee.com/orbit-w/orbit/lib/module/db/mgo_builder"
	"gitee.com/orbit-w/orbit/lib/module/dirty/dirty_tracker"
	"github.com/gogo/protobuf/proto"
)

// Dirty bits for component fields
const (
	LevelUpMechanismDirtyIdBit         int64 = 1 << 0
	LevelUpMechanismDirtyConfIdBit     int64 = 1 << 1
	LevelUpMechanismDirtyUseTimesBit   int64 = 1 << 2
	LevelUpMechanismDirtyCreateTimeBit int64 = 1 << 3
	LevelUpMechanismDirtySkillsBit     int64 = 1 << 4
)

type HeroMechanism struct {
	mme.HeroMechanism
	dirty_tracker.DirtyTracker
}

func NewHeroMechanism() *HeroMechanism {
	return &HeroMechanism{}
}

// 机制唯一名称
func (m *HeroMechanism) Name() string {
	return "HeroMechanism"
}

func (m *HeroMechanism) Link(parent *dirty_tracker.DirtyTracker, parentBit int64) {
	m.DirtyTracker.Link(parent, parentBit)
}

func (m *HeroMechanism) GetConfId() int32 {
	return m.ConfId
}

func (m *HeroMechanism) SetConfId(v int32) {
	m.ConfId = v
	m.MarkDirty(LevelUpMechanismDirtyConfIdBit)
}

func (m *HeroMechanism) GetUseTimes() int32 {
	return m.UseTimes
}

func (m *HeroMechanism) SetUseTimes(v int32) {
	m.UseTimes = v
	m.MarkDirty(LevelUpMechanismDirtyUseTimesBit)
}

func (m *HeroMechanism) GetCreateTime() int64 {
	return m.CreateTime
}

func (m *HeroMechanism) SetCreateTime(v int64) {
	m.CreateTime = v
	m.MarkDirty(LevelUpMechanismDirtyCreateTimeBit)
}

func (m *HeroMechanism) GetSkills() map[int32]int32 {
	return m.Skills
}

func (m *HeroMechanism) SetSkills(v map[int32]int32) {
	m.Skills = v
	m.MarkDirty(LevelUpMechanismDirtySkillsBit)
}

// ClearAllDirtyFlags 清除所有脏标记位
func (m *HeroMechanism) ClearAllDirtyFlags() {
	m.DirtyTracker.ClearAllDirty()
}

func (m *HeroMechanism) BuildMongoUpdate(builder *mgo_builder.MongoUpdateBuilder, prefix string) {
	if m == nil {
		return
	}
	{
		path := "_id"
		if prefix != "" {
			path = prefix + "." + path
		}
		if m.IsDirty(LevelUpMechanismDirtyIdBit) {
			builder.Set(path, m.Id)
		}
	}
	{
		path := "conf_id"
		if prefix != "" {
			path = prefix + "." + path
		}
		if m.IsDirty(LevelUpMechanismDirtyConfIdBit) {
			builder.Set(path, m.ConfId)
		}
	}
	{
		path := "use_times"
		if prefix != "" {
			path = prefix + "." + path
		}
		if m.IsDirty(LevelUpMechanismDirtyUseTimesBit) {
			builder.Set(path, m.UseTimes)
		}
	}
	{
		path := "skills"
		if prefix != "" {
			path = prefix + "." + path
		}
		if m.IsDirty(LevelUpMechanismDirtySkillsBit) {
			builder.Set(path, m.Skills)
		}
	}
	{
		path := "create_time"
		if prefix != "" {
			path = prefix + "." + path
		}
		if m.IsDirty(LevelUpMechanismDirtyCreateTimeBit) {
			builder.Set(path, m.CreateTime)
		}
	}
}

// DeepCopy creates a deep copy of LevelUpMechanism
func (s *HeroMechanism) DeepCopy(co *HeroMechanism) {
	if s == nil {
		return
	}

	*co = *s

	if s.Skills != nil {
		co.Skills = make(map[int32]int32, len(s.Skills))
		for k, v := range s.Skills {
			co.Skills[k] = v
		}
	}
}

// PB 根据脏标记位构建增量数据的 protoMessage
// 只返回标记为脏的字段数据，用于增量同步
func (m *HeroMechanism) PB() proto.Message {
	if m == nil {
		return nil
	}

	// 如果没有脏标记，返回 nil
	if !m.HasAnyDirty() {
		return nil
	}

	incremental := &mme.HeroMechanism{}

	// 根据脏标记位设置对应的字段
	if m.IsDirty(LevelUpMechanismDirtyIdBit) {
		incremental.Id = m.Id
	}

	if m.IsDirty(LevelUpMechanismDirtyConfIdBit) {
		incremental.ConfId = m.ConfId
	}

	if m.IsDirty(LevelUpMechanismDirtyCreateTimeBit) {
		incremental.CreateTime = m.CreateTime
	}

	if m.IsDirty(LevelUpMechanismDirtyUseTimesBit) {
		incremental.UseTimes = m.UseTimes
	}

	if m.IsDirty(LevelUpMechanismDirtySkillsBit) {
		// 深拷贝 Skills map
		if m.Skills != nil {
			incremental.Skills = make(map[int32]int32, len(m.Skills))
			for k, v := range m.Skills {
				incremental.Skills[k] = v
			}
		}
	}

	return incremental
}
