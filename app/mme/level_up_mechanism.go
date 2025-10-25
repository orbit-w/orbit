package mme

import (
	"strings"

	"gitee.com/orbit-w/meteor/bases/dirty/dirty_tracker"
	"gitee.com/orbit-w/orbit/app/proto/mme"
	"gitee.com/orbit-w/orbit/lib/module/db/mgo_builder"
	"github.com/gogo/protobuf/proto"
)

// Dirty bits for Mechanism fields
const (
	LevelUpMechanismDirtyXXXIdBit    int64 = 1 << 0
	LevelUpMechanismDirtyCurLevelBit int64 = 1 << 1
	LevelUpMechanismDirtyCurExpBit   int64 = 1 << 2
	LevelUpMechanismDirtyConfIdBit   int64 = 1 << 3
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

func (m *LevelUpMechanism) GetXXXId() int64 {
	if m != nil {
		return m.XXXId
	}
	return 0
}

func (m *LevelUpMechanism) SetXXXId(v int64) {
	m.XXXId = v
	m.MarkDirty(LevelUpMechanismDirtyXXXIdBit)
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

// DeepCopy creates a deep copy of LevelUpMechanism proto data only
// 手写实现，性能优于 proto.Clone（避免反射开销）
func (m *LevelUpMechanism) DeepCopy(co *mme.LevelUpMechanism) {
	if m == nil || co == nil {
		return
	}

	// 深拷贝指针字段
	if m.CurLevel != nil {
		v := *m.CurLevel
		co.CurLevel = &v
	}

	if m.CurExp != nil {
		v := *m.CurExp
		co.CurExp = &v
	}

	if m.ConfId != nil {
		v := *m.ConfId
		co.ConfId = &v
	}

	// 拷贝 XXXId
	co.XXXId = m.XXXId
}

// ToIncrementalProto 根据脏标记位构建增量数据的 protoMessage
// 只返回标记为脏的字段数据，用于增量同步
func (m *LevelUpMechanism) ToIncrementalProto() proto.Message {
	if m == nil {
		return nil
	}

	incremental := &mme.LevelUpMechanism{}

	// 如果没有脏标记，返回 nil
	if !m.HasAnyDirty() {
		return nil
	}

	if m.IsDirty(LevelUpMechanismDirtyXXXIdBit) {
		incremental.XXXId = m.XXXId
	}

	if m.IsDirty(LevelUpMechanismDirtyCurLevelBit) {
		incremental.CurLevel = m.CurLevel
	}

	if m.IsDirty(LevelUpMechanismDirtyCurExpBit) {
		incremental.CurExp = m.CurExp
	}

	if m.IsDirty(LevelUpMechanismDirtyConfIdBit) {
		incremental.ConfId = m.ConfId
	}

	return incremental
}
