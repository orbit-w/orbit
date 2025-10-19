package mme

import (
	"gitee.com/orbit-w/orbit/app/proto/mme"
	"gitee.com/orbit-w/orbit/lib/module/db/mgo_builder"
	"gitee.com/orbit-w/orbit/lib/module/dirty/dirty_tracker"
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
		path := "equipment"
		if prefix != "" {
			path = prefix + "." + path
		}
		if m.IsDirty(LevelUpMechanismDirtyUseTimesBit) {
			builder.Set(path, m.UseTimes)
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
