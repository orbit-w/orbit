package mme

import (
	"strings"

	"gitee.com/orbit-w/meteor/bases/dirty/dirty_tracker"
	"gitee.com/orbit-w/orbit/app/proto/mme"
	"gitee.com/orbit-w/orbit/lib/module/db/mgo_builder"
)

// Dirty bits for Mechanism fields
const (
	LevelUpMechanismDirtyCurLevelBit int64 = 1 << 0
	LevelUpMechanismDirtyCurExpBit   int64 = 1 << 1
	LevelUpMechanismDirtyConfIdBit   int64 = 1 << 2
)

type LevelUpMechanism struct {
	*mme.LevelUpMechanism
	dirty_tracker.DirtyTracker
}

func NewLevelUpMechanism() *LevelUpMechanism {
	return &LevelUpMechanism{
		LevelUpMechanism: new(mme.LevelUpMechanism),
	}
}

// 机制唯一名称
func (m *LevelUpMechanism) Name() string {
	return "LevelUpMechanism"
}

func (m *LevelUpMechanism) Link(parent *dirty_tracker.DirtyTracker, parentBit int64) {
	m.DirtyTracker.Link(parent, parentBit)
}

func (m *LevelUpMechanism) SetCurLevel(v int32) {
	m.CurLevel = &v
	m.MarkDirty(LevelUpMechanismDirtyCurLevelBit)
}

func (m *LevelUpMechanism) SetCurExp(v int32) {
	m.CurExp = &v
	m.MarkDirty(LevelUpMechanismDirtyCurExpBit)
}

func (m *LevelUpMechanism) SetConfId(v int32) {
	m.ConfId = &v
	m.MarkDirty(LevelUpMechanismDirtyConfIdBit)
}

// ClearAllDirtyFlags 清除所有脏标记位
func (m *LevelUpMechanism) ClearAllDirtyFlags() {
	m.DirtyTracker.ClearAllDirty()
}

// BuildMongoUpdate 构建MongoDB更新操作
func (m *LevelUpMechanism) BuildMongoUpdate(builder *mgo_builder.MongoUpdateBuilder, prefix string) {
	if m == nil {
		return
	}

	// 定义字段配置，避免重复代码
	fieldConfigs := []struct {
		fieldName string
		dirtyBit  int64
		value     any
	}{
		{"cur_level", LevelUpMechanismDirtyCurLevelBit, m.CurLevel},
		{"cur_exp", LevelUpMechanismDirtyCurExpBit, m.CurExp},
		{"conf_id", LevelUpMechanismDirtyConfIdBit, m.ConfId},
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
