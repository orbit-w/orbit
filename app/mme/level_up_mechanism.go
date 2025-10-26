package mme

import (
	"strings"

	"gitee.com/orbit-w/orbit/app/proto/mme"
	dirtyflag "gitee.com/orbit-w/orbit/lib/base/dirty_flag"
	"gitee.com/orbit-w/orbit/lib/module/db/mgo_builder"
	"google.golang.org/protobuf/proto"
)

// Dirty bits for Mechanism fields
const (
	LevelUpMechanismDirtyCurLevelBit int64 = 1 << 0
	LevelUpMechanismDirtyCurExpBit   int64 = 1 << 1
	LevelUpMechanismDirtyConfIdBit   int64 = 1 << 2
)

type LevelUpMechanism struct {
	levelUpMechanism *mme.LevelUpMechanism
	dirtyflag.IDirtyFlag
}

func NewLevelUpMechanism(levelUpMechanism *mme.LevelUpMechanism) *LevelUpMechanism {
	if levelUpMechanism == nil {
		panic("levelUpMechanism is nil")
	}
	m := &LevelUpMechanism{
		levelUpMechanism: levelUpMechanism,
		IDirtyFlag:       dirtyflag.NewDirtyFlag(),
	}
	return m
}

// 机制唯一名称
func (m *LevelUpMechanism) Name() string {
	return "LevelUpMechanism"
}

func (m *LevelUpMechanism) SetCurLevel(v int32) {
	m.levelUpMechanism.CurLevel = &v
	m.MarkDirty(LevelUpMechanismDirtyCurLevelBit)
}

func (m *LevelUpMechanism) SetCurExp(v int32) {
	m.levelUpMechanism.CurExp = &v
	m.MarkDirty(LevelUpMechanismDirtyCurExpBit)
}

func (m *LevelUpMechanism) SetConfId(v int32) {
	m.levelUpMechanism.ConfId = &v
	m.MarkDirty(LevelUpMechanismDirtyConfIdBit)
}

func (m *LevelUpMechanism) GetCurLevel() int32 {
	return *m.levelUpMechanism.CurLevel
}

func (m *LevelUpMechanism) GetCurExp() int32 {
	return *m.levelUpMechanism.CurExp
}

func (m *LevelUpMechanism) GetConfId() int32 {
	return *m.levelUpMechanism.ConfId
}

// ClearAllDirtyFlags 清除所有脏标记位
func (m *LevelUpMechanism) ClearAllDirtyFlags() {
	m.ClearAllDirty()
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
		{"cur_level", LevelUpMechanismDirtyCurLevelBit, m.GetCurLevel()},
		{"cur_exp", LevelUpMechanismDirtyCurExpBit, m.GetCurExp()},
		{"conf_id", LevelUpMechanismDirtyConfIdBit, m.GetConfId()},
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

func (m *LevelUpMechanism) DeepCopy() *mme.LevelUpMechanism {
	if m == nil {
		return nil
	}
	return proto.Clone(m.levelUpMechanism).(*mme.LevelUpMechanism)

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

	if m.IsDirty(LevelUpMechanismDirtyCurLevelBit) {
		v := m.GetCurLevel()
		incremental.CurLevel = &v
	}

	if m.IsDirty(LevelUpMechanismDirtyCurExpBit) {
		v := m.GetCurExp()
		incremental.CurExp = &v
	}

	if m.IsDirty(LevelUpMechanismDirtyConfIdBit) {
		v := m.GetConfId()
		incremental.ConfId = &v
	}

	return incremental
}
