package mme

import (
	"gitee.com/orbit-w/orbit/app/proto/mme"
	dirtyflag "gitee.com/orbit-w/orbit/lib/base/dirty_flag"
	fieldmeta "gitee.com/orbit-w/orbit/lib/base/field_meta"
	"gitee.com/orbit-w/orbit/lib/module/db/mgo_builder"
	mmemodel "gitee.com/orbit-w/orbit/lib/module/mme_model"
	"google.golang.org/protobuf/proto"
	"maps"
)

const (
	ManualUnlockMechanismFieldIndexUnlockMap = uint8(0)
)

// Dirty bits for Mechanism fields
const (
	ManualUnlockMechanismDirtyUnlockMapBit int64 = 1 << ManualUnlockMechanismFieldIndexUnlockMap
)

type ManualUnlockMechanism struct {
	UnlockMap map[int32]bool
}

func NewManualUnlockMechanism() *ManualUnlockMechanism {
	return &ManualUnlockMechanism{
		UnlockMap: make(map[int32]bool),
	}
}

func (m *ManualUnlockMechanism) DeepCopy(co *ManualUnlockMechanism) {
	if m == nil || co == nil {
		return
	}

	*co = *m
	maps.Copy(co.UnlockMap, m.UnlockMap)
}

func (m *ManualUnlockMechanism) ToProto() *mme.ManualUnlockMechanism {
	if m == nil {
		return nil
	}

	pb := &mme.ManualUnlockMechanism{}

	if m.UnlockMap != nil {
		pb.UnlockMap = make(map[int32]bool, len(m.UnlockMap))
		maps.Copy(pb.UnlockMap, m.UnlockMap)
	}
	return pb
}

func (m *ManualUnlockMechanism) FromProto(pb *mme.ManualUnlockMechanism) {
	if m == nil || pb == nil {
		return
	}

	if pb.UnlockMap != nil {
		m.UnlockMap = make(map[int32]bool, len(pb.UnlockMap))
		maps.Copy(m.UnlockMap, pb.UnlockMap)
	}
}

type ManualUnlockMechanismWrapper struct {
	data *ManualUnlockMechanism
	dirtyflag.IDirtyFlag
	fieldMetas *fieldmeta.FieldMetas
}

func NewManualUnlockMechanismWrapper(data *ManualUnlockMechanism) *ManualUnlockMechanismWrapper {
	if data == nil {
		panic("data is nil")
	}
	w := &ManualUnlockMechanismWrapper{
		data:       data,
		IDirtyFlag: dirtyflag.NewDirtyFlag(),
		fieldMetas: fieldmeta.NewFieldMetas(),
	}

	return w
}

func (w *ManualUnlockMechanismWrapper) InitFieldContext() {
	w.fieldMetas.SetFieldType(ManualUnlockMechanismFieldIndexUnlockMap, fieldmeta.FieldTypeSync)
}

func (w *ManualUnlockMechanismWrapper) Name() string {
	return "ManualUnlockMechanism"
}

// MatchesAll 判断字段是否匹配所有类型标记
func (w *ManualUnlockMechanismWrapper) MatchesAll(fieldID uint8, fieldTypes ...fieldmeta.FieldType) bool {
	return w.fieldMetas.MatchesAll(fieldID, fieldTypes...)
}

// 包装器-获取UnlockMap
func (w *ManualUnlockMechanismWrapper) GetUnlockMap() map[int32]bool {
	return w.data.UnlockMap
}

// 包装器-设置UnlockMap
func (w *ManualUnlockMechanismWrapper) SetUnlockMap(v map[int32]bool) {
	w.data.UnlockMap = v
	w.MarkDirty(ManualUnlockMechanismDirtyUnlockMapBit)
}

// ClearAllDirtyFlags 清除所有脏标记位
func (w *ManualUnlockMechanismWrapper) ClearAllDirtyFlags() {
	w.IDirtyFlag.ClearAllDirty()

	// 如果Value为引用类型且有脏标记，则清除所有xmap中Value的脏标记

	// 清除所有xmap中的操作记录
}

// BuildMongoUpdate 构建MongoDB更新操作
func (w *ManualUnlockMechanismWrapper) BuildMongoUpdate(builder *mgo_builder.MongoUpdateBuilder, path *mgo_builder.NestedPath) {
	if w == nil {
		return
	}

	if w.IsDirty(ManualUnlockMechanismDirtyUnlockMapBit) {
		copy := make(map[int32]bool, len(w.data.UnlockMap))
		maps.Copy(copy, w.data.UnlockMap)
		builder.SetNestedPath(path, "unlock_map", copy)
	}
}

// ToIncrementalProtoWithContext 根据脏标记位构建增量数据的 protoMessage
// 只返回标记为脏的字段数据，用于增量同步
func (w *ManualUnlockMechanismWrapper) ToIncrementalProtoWithContext(ctx mmemodel.SyncContext) proto.Message {
	if w == nil {
		return nil
	}

	// 如果没有脏标记，返回 nil
	if !w.HasAnyDirty() {
		return nil
	}

	incremental := &mme.ManualUnlockMechanism{}

	if mmemodel.FieldCanBeIncrementalSynced(w, ManualUnlockMechanismDirtyUnlockMapBit, ManualUnlockMechanismFieldIndexUnlockMap, ctx) {
		m := w.GetUnlockMap()
		if m != nil {
			incremental.UnlockMap = make(map[int32]bool, len(m))
			maps.Copy(incremental.UnlockMap, m)
		}
	}
	return incremental
}

// ToProto 将 ManualUnlockMechanism 数据转换为完整的 protobuf 结构体
func (w *ManualUnlockMechanismWrapper) ToProto() *mme.ManualUnlockMechanism {
	if w == nil || w.data == nil {
		return nil
	}

	return w.data.ToProto()
}

// FromProto 从 protobuf 结构体加载数据到 ManualUnlockMechanism
func (w *ManualUnlockMechanismWrapper) FromProto(pb *mme.ManualUnlockMechanism) {
	if w == nil || w.data == nil || pb == nil {
		return
	}

	if pb.UnlockMap != nil {
		maps.Copy(w.data.UnlockMap, pb.UnlockMap)
	}
}

// 包装器-深拷贝
func (w *ManualUnlockMechanismWrapper) DeepCopy() *ManualUnlockMechanism {
	if w == nil {
		return nil
	}
	copy := &ManualUnlockMechanism{}
	w.DeepCopyTo(copy)
	return copy
}

// 包装器-深拷贝
func (w *ManualUnlockMechanismWrapper) DeepCopyTo(copy *ManualUnlockMechanism) {
	if w == nil || w.data == nil || copy == nil {
		return
	}

	// 初始化目标 map
	if copy.UnlockMap == nil {
		copy.UnlockMap = make(map[int32]bool, len(w.data.UnlockMap))
	}
	// 拷贝所有 bool
	if w.data.UnlockMap != nil {
		maps.Copy(copy.UnlockMap, w.data.UnlockMap)
	}
}
