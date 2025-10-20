package mme

import (
	"gitee.com/orbit-w/orbit/app/proto/mme"
	"gitee.com/orbit-w/orbit/lib/module/db/mgo_builder"
	"gitee.com/orbit-w/orbit/lib/module/dirty/dirty_tracker"
	"gitee.com/orbit-w/orbit/lib/module/dirty/xmap"
	"github.com/gogo/protobuf/proto"
)

// Dirty bits for component fields
const (
	LevelUpMechanismDirtyIdBit         int64 = 1 << 0
	LevelUpMechanismDirtyConfIdBit     int64 = 1 << 1
	LevelUpMechanismDirtyCreateTimeBit int64 = 1 << 2
	LevelUpMechanismDirtyUseTimesBit   int64 = 1 << 3
	LevelUpMechanismDirtySkillsBit     int64 = 1 << 4
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

	hm.SkillsAccessor = xmap.NewMapAccessorWithMarker(&hm.Skills, hm, LevelUpMechanismDirtySkillsBit)
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
	m.MarkDirty(LevelUpMechanismDirtyIdBit)
}

func (m *HeroMechanism) SetConfId(v int32) {
	m.ConfId = &v
	m.MarkDirty(LevelUpMechanismDirtyConfIdBit)
}

func (m *HeroMechanism) SetUseTimes(v int32) {
	m.UseTimes = &v
	m.MarkDirty(LevelUpMechanismDirtyUseTimesBit)
}

func (m *HeroMechanism) SetCreateTime(v int64) {
	m.CreateTime = &v
	m.MarkDirty(LevelUpMechanismDirtyCreateTimeBit)
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
		path := "create_time"
		if prefix != "" {
			path = prefix + "." + path
		}
		if m.IsDirty(LevelUpMechanismDirtyCreateTimeBit) {
			builder.Set(path, m.CreateTime)
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
		incremental.SkillChanges = make([]*mme.SkillXXXChange, 0)
		m.SkillsAccessor.RangeOperations(func(key int32, operation xmap.MapOperation[int32]) bool {
			switch operation.Type {
			case xmap.SetOperation:
				v, _ := m.SkillsAccessor.Get(key)
				incremental.SkillChanges = append(incremental.SkillChanges, &mme.SkillXXXChange{
					Key:   key,
					Value: v,
				})
			case xmap.DeleteOperation:
				incremental.SkillChanges = append(incremental.SkillChanges, &mme.SkillXXXChange{
					Key: key,
				})
			}
			return true
		})
	}

	return incremental
}
