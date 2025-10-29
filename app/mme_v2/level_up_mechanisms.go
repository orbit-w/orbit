package mme

import (
	"gitee.com/orbit-w/orbit/app/proto/mme"
	dirtyflag "gitee.com/orbit-w/orbit/lib/base/dirty_flag"
	fieldmeta "gitee.com/orbit-w/orbit/lib/base/field_meta"
	"gitee.com/orbit-w/orbit/lib/module/db/mgo_builder"
	mmeutils "gitee.com/orbit-w/orbit/lib/module/mme_utils"
	"google.golang.org/protobuf/proto"
)

const (
	LevelUpMechanismFieldIndexCurLevel = uint8(0)
	LevelUpMechanismFieldIndexCurExp
	LevelUpMechanismFieldIndexConfId
)

// Dirty bits for LevelUpMechanism fields
const (
	LevelUpMechanismDirtyCurLevelBit int64 = 1 << LevelUpMechanismFieldIndexCurLevel
	LevelUpMechanismDirtyCurExpBit   int64 = 1 << LevelUpMechanismFieldIndexCurExp
	LevelUpMechanismDirtyConfIdBit   int64 = 1 << LevelUpMechanismFieldIndexConfId
)

type LevelUpMechanism struct {
	CurLevel int32 `bson:"cur_level"`
	CurExp   int32 `bson:"cur_exp"`
	ConfId   int32 `bson:"conf_id"`
}

func (m *LevelUpMechanism) DeepCopy(co *LevelUpMechanism) {
	if m == nil || co == nil {
		return
	}

	*co = *m
}

// ToProto 将 LevelUpMechanism 数据转换为完整的 protobuf 结构体
func (m *LevelUpMechanism) ToProto() *mme.LevelUpMechanism {
	if m == nil {
		return nil
	}

	pb := &mme.LevelUpMechanism{}

	curLevel := m.CurLevel
	pb.CurLevel = &curLevel

	curExp := m.CurExp
	pb.CurExp = &curExp

	confId := m.ConfId
	pb.ConfId = &confId

	return pb
}

type LevelUpMechanismWrapper struct {
	data *LevelUpMechanism
	dirtyflag.IDirtyFlag
	fieldMetas *fieldmeta.FieldMetas
}

func NewLevelUpMechanismWrapper(data *LevelUpMechanism) *LevelUpMechanismWrapper {
	if data == nil {
		panic("data is nil")
	}
	lm := &LevelUpMechanismWrapper{
		data:       data,
		IDirtyFlag: dirtyflag.NewDirtyFlag(),
		fieldMetas: fieldmeta.NewFieldMetas(),
	}

	return lm
}

func (m *LevelUpMechanismWrapper) InitFieldContext() {
	m.fieldMetas.SetFieldType(LevelUpMechanismFieldIndexCurLevel, fieldmeta.FieldTypeSync)
	m.fieldMetas.SetFieldType(LevelUpMechanismFieldIndexCurExp, fieldmeta.FieldTypeSync)
	m.fieldMetas.SetFieldType(LevelUpMechanismFieldIndexConfId, fieldmeta.FieldTypeSync)
}

// 机制唯一名称
func (m *LevelUpMechanismWrapper) Name() string {
	return "LevelUpMechanism"
}

func (m *LevelUpMechanismWrapper) GetCurLevel() int32 {
	return m.data.CurLevel
}

func (m *LevelUpMechanismWrapper) GetCurExp() int32 {
	return m.data.CurExp
}

func (m *LevelUpMechanismWrapper) GetConfId() int32 {
	return m.data.ConfId
}

func (m *LevelUpMechanismWrapper) SetCurLevel(v int32) {
	m.data.CurLevel = v
	m.MarkDirty(LevelUpMechanismDirtyCurLevelBit)
}

func (m *LevelUpMechanismWrapper) SetCurExp(v int32) {
	m.data.CurExp = v
	m.MarkDirty(LevelUpMechanismDirtyCurExpBit)
}

func (m *LevelUpMechanismWrapper) SetConfId(v int32) {
	m.data.ConfId = v
	m.MarkDirty(LevelUpMechanismDirtyConfIdBit)
}

// ClearAllDirtyFlags 清除所有脏标记位
func (m *LevelUpMechanismWrapper) ClearAllDirtyFlags() {
	m.ClearAllDirty()
}

// BuildMongoUpdate 构建MongoDB更新操作
func (m *LevelUpMechanismWrapper) BuildMongoUpdate(builder *mgo_builder.MongoUpdateBuilder, path *mgo_builder.NestedPath) {
	if m == nil {
		return
	}

	if m.IsDirty(LevelUpMechanismDirtyCurLevelBit) {
		builder.SetNestedPath(path, "cur_level", m.GetCurLevel())
	}
	if m.IsDirty(LevelUpMechanismDirtyCurExpBit) {
		builder.SetNestedPath(path, "cur_exp", m.GetCurExp())
	}
	if m.IsDirty(LevelUpMechanismDirtyConfIdBit) {
		builder.SetNestedPath(path, "conf_id", m.GetConfId())
	}
}

func (m *LevelUpMechanismWrapper) Clone() *LevelUpMechanism {
	if m == nil {
		return nil
	}
	copy := &LevelUpMechanism{}
	m.DeepCopy(copy)
	return copy
}

func (m *LevelUpMechanismWrapper) DeepCopy(copy *LevelUpMechanism) {
	if m == nil || m.data == nil || copy == nil {
		return
	}

	// 拷贝基础类型字段
	*copy = *m.data
}

// MatchesAll 判断字段是否匹配所有类型标记
func (m *LevelUpMechanismWrapper) MatchesAll(fieldID uint8, fieldTypes ...fieldmeta.FieldType) bool {
	return m.fieldMetas.MatchesAll(fieldID, fieldTypes...)
}

// ToProto 将 LevelUpMechanism 数据转换为完整的 protobuf 结构体
func (m *LevelUpMechanismWrapper) ToProto() *mme.LevelUpMechanism {
	if m == nil || m.data == nil {
		return nil
	}

	return m.data.ToProto()
}

// FromProto 从 protobuf 结构体加载数据到 LevelUpMechanism
func (m *LevelUpMechanismWrapper) FromProto(pb *mme.LevelUpMechanism) {
	if m == nil || m.data == nil || pb == nil {
		return
	}

	// 加载基础类型字段
	if pb.CurLevel != nil {
		m.SetCurLevel(*pb.CurLevel)
	}

	if pb.CurExp != nil {
		m.SetCurExp(*pb.CurExp)
	}

	if pb.ConfId != nil {
		m.SetConfId(*pb.ConfId)
	}
}

// ToIncrementalProto 根据脏标记位构建增量数据的 protoMessage
// 只返回标记为脏的字段数据，用于增量同步
func (m *LevelUpMechanismWrapper) ToIncrementalProtoWithContext(ctx mmeutils.SyncContext) proto.Message {
	if m == nil {
		return nil
	}

	// 如果没有脏标记，返回 nil
	if !m.HasAnyDirty() {
		return nil
	}

	incremental := &mme.LevelUpMechanism{}

	// 根据脏标记位设置对应的字段
	if mmeutils.FieldCanBeIncrementalSynced(m, LevelUpMechanismDirtyCurLevelBit, LevelUpMechanismFieldIndexCurLevel, ctx) {
		v := m.GetCurLevel()
		incremental.CurLevel = &v
	}

	if mmeutils.FieldCanBeIncrementalSynced(m, LevelUpMechanismDirtyCurExpBit, LevelUpMechanismFieldIndexCurExp, ctx) {
		v := m.GetCurExp()
		incremental.CurExp = &v
	}

	if mmeutils.FieldCanBeIncrementalSynced(m, LevelUpMechanismDirtyConfIdBit, LevelUpMechanismFieldIndexConfId, ctx) {
		v := m.GetConfId()
		incremental.ConfId = &v
	}

	return incremental
}
