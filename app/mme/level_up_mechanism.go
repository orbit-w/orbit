package mme

import (
	"gitee.com/orbit-w/orbit/app/proto/mme"
	dirtyflag "gitee.com/orbit-w/orbit/lib/base/dirty_flag"
	fieldmeta "gitee.com/orbit-w/orbit/lib/base/field_meta"
	"gitee.com/orbit-w/orbit/lib/module/db/mgo_builder"
	mmemodel "gitee.com/orbit-w/orbit/lib/module/mme_model"
	"google.golang.org/protobuf/proto"
)

const (
	LevelUpMechanismFieldIndexCurLevel = uint8(0)
	LevelUpMechanismFieldIndexCurExp   = uint8(1)
	LevelUpMechanismFieldIndexConfId   = uint8(2)
)

// Dirty bits for Mechanism fields
const (
	LevelUpMechanismDirtyCurLevelBit int64 = 1 << LevelUpMechanismFieldIndexCurLevel
	LevelUpMechanismDirtyCurExpBit   int64 = 1 << LevelUpMechanismFieldIndexCurExp
	LevelUpMechanismDirtyConfIdBit   int64 = 1 << LevelUpMechanismFieldIndexConfId
)

type LevelUpMechanism struct {
	CurLevel int32
	CurExp   int32
	ConfId   int32
}

func NewLevelUpMechanism() *LevelUpMechanism {
	return &LevelUpMechanism{}
}

func (m *LevelUpMechanism) DeepCopy(co *LevelUpMechanism) {
	if m == nil || co == nil {
		return
	}

	*co = *m
}

func (m *LevelUpMechanism) ToProto() *mme.LevelUpMechanism {
	if m == nil {
		return nil
	}

	pb := &mme.LevelUpMechanism{}

	curlevel := m.CurLevel
	pb.CurLevel = &curlevel
	curexp := m.CurExp
	pb.CurExp = &curexp
	confid := m.ConfId
	pb.ConfId = &confid
	return pb
}

func (m *LevelUpMechanism) FromProto(pb *mme.LevelUpMechanism) {
	if m == nil || pb == nil {
		return
	}

	if pb.CurLevel != nil {
		m.CurLevel = *pb.CurLevel
	}
	if pb.CurExp != nil {
		m.CurExp = *pb.CurExp
	}
	if pb.ConfId != nil {
		m.ConfId = *pb.ConfId
	}
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
	w := &LevelUpMechanismWrapper{
		data:       data,
		IDirtyFlag: dirtyflag.NewDirtyFlag(),
		fieldMetas: fieldmeta.NewFieldMetas(),
	}

	return w
}

func (w *LevelUpMechanismWrapper) InitFieldContext() {
	w.fieldMetas.SetFieldType(LevelUpMechanismFieldIndexCurLevel, fieldmeta.FieldTypeSync)
	w.fieldMetas.SetFieldType(LevelUpMechanismFieldIndexCurExp, fieldmeta.FieldTypeSync)
	w.fieldMetas.SetFieldType(LevelUpMechanismFieldIndexConfId, fieldmeta.FieldTypeSync)
}

func (w *LevelUpMechanismWrapper) Name() string {
	return "LevelUpMechanism"
}

// MatchesAll 判断字段是否匹配所有类型标记
func (w *LevelUpMechanismWrapper) MatchesAll(fieldID uint8, fieldTypes ...fieldmeta.FieldType) bool {
	return w.fieldMetas.MatchesAll(fieldID, fieldTypes...)
}

func (w *LevelUpMechanismWrapper) GetCurLevel() int32 {
	return w.data.CurLevel
}

func (w *LevelUpMechanismWrapper) SetCurLevel(v int32) {
	w.data.CurLevel = v
	w.MarkDirty(LevelUpMechanismDirtyCurLevelBit)
}

func (w *LevelUpMechanismWrapper) GetCurExp() int32 {
	return w.data.CurExp
}

func (w *LevelUpMechanismWrapper) SetCurExp(v int32) {
	w.data.CurExp = v
	w.MarkDirty(LevelUpMechanismDirtyCurExpBit)
}

func (w *LevelUpMechanismWrapper) GetConfId() int32 {
	return w.data.ConfId
}

func (w *LevelUpMechanismWrapper) SetConfId(v int32) {
	w.data.ConfId = v
	w.MarkDirty(LevelUpMechanismDirtyConfIdBit)
}

// ClearAllDirtyFlags 清除所有脏标记位
func (w *LevelUpMechanismWrapper) ClearAllDirtyFlags() {
	w.IDirtyFlag.ClearAllDirty()

	// 如果Value为引用类型且有脏标记，则清除所有xmap中Value的脏标记

	// 清除所有xmap中的操作记录
}

// BuildMongoUpdate 构建MongoDB更新操作
func (w *LevelUpMechanismWrapper) BuildMongoUpdate(builder *mgo_builder.MongoUpdateBuilder, path *mgo_builder.NestedPath) {
	if w == nil {
		return
	}

	if w.IsDirty(LevelUpMechanismDirtyCurLevelBit) {
		builder.SetNestedPath(path, "cur_level", w.GetCurLevel())
	}
	if w.IsDirty(LevelUpMechanismDirtyCurExpBit) {
		builder.SetNestedPath(path, "cur_exp", w.GetCurExp())
	}
	if w.IsDirty(LevelUpMechanismDirtyConfIdBit) {
		builder.SetNestedPath(path, "conf_id", w.GetConfId())
	}
}

// ToIncrementalProtoWithContext 根据脏标记位构建增量数据的 protoMessage
// 只返回标记为脏的字段数据，用于增量同步
func (w *LevelUpMechanismWrapper) ToIncrementalProtoWithContext(ctx mmemodel.SyncContext) proto.Message {
	if w == nil {
		return nil
	}

	// 如果没有脏标记，返回 nil
	if !w.HasAnyDirty() {
		return nil
	}

	incremental := &mme.LevelUpMechanism{}

	if mmemodel.FieldCanBeIncrementalSynced(w, LevelUpMechanismDirtyCurLevelBit, LevelUpMechanismFieldIndexCurLevel, ctx) {
		v := w.GetCurLevel()
		incremental.CurLevel = &v
	}
	if mmemodel.FieldCanBeIncrementalSynced(w, LevelUpMechanismDirtyCurExpBit, LevelUpMechanismFieldIndexCurExp, ctx) {
		v := w.GetCurExp()
		incremental.CurExp = &v
	}
	if mmemodel.FieldCanBeIncrementalSynced(w, LevelUpMechanismDirtyConfIdBit, LevelUpMechanismFieldIndexConfId, ctx) {
		v := w.GetConfId()
		incremental.ConfId = &v
	}
	return incremental
}

// ToProto 将 LevelUpMechanism 数据转换为完整的 protobuf 结构体
func (w *LevelUpMechanismWrapper) ToProto() *mme.LevelUpMechanism {
	if w == nil || w.data == nil {
		return nil
	}

	return w.data.ToProto()
}

// FromProto 从 protobuf 结构体加载数据到 LevelUpMechanism
func (w *LevelUpMechanismWrapper) FromProto(pb *mme.LevelUpMechanism) {
	if w == nil || w.data == nil || pb == nil {
		return
	}

	if pb.CurLevel != nil {
		w.SetCurLevel(*pb.CurLevel)
	}
	if pb.CurExp != nil {
		w.SetCurExp(*pb.CurExp)
	}
	if pb.ConfId != nil {
		w.SetConfId(*pb.ConfId)
	}
}

// 包装器-深拷贝
func (w *LevelUpMechanismWrapper) DeepCopy() *LevelUpMechanism {
	if w == nil {
		return nil
	}
	copy := &LevelUpMechanism{}
	w.DeepCopyTo(copy)
	return copy
}

// 包装器-深拷贝
func (w *LevelUpMechanismWrapper) DeepCopyTo(copy *LevelUpMechanism) {
	if w == nil || w.data == nil || copy == nil {
		return
	}

	// 拷贝基础类型字段
	*copy = *w.data

}
