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
	WearMechanismFieldIndexWearMap = uint8(0)
	WearMechanismFieldIndexConfId  = uint8(1)
)

// Dirty bits for Mechanism fields
const (
	WearMechanismDirtyWearMapBit int64 = 1 << WearMechanismFieldIndexWearMap
	WearMechanismDirtyConfIdBit  int64 = 1 << WearMechanismFieldIndexConfId
)

type WearMechanism struct {
	WearMap map[int32]int32 `bson:"wear_map"`
	ConfId  int32           `bson:"conf_id"`
}

func NewWearMechanism() *WearMechanism {
	return &WearMechanism{
		WearMap: make(map[int32]int32),
	}
}

func (m *WearMechanism) DeepCopy(co *WearMechanism) {
	if m == nil || co == nil {
		return
	}

	*co = *m
	maps.Copy(co.WearMap, m.WearMap)
}

func (m *WearMechanism) ToProto() *mme.WearMechanism {
	if m == nil {
		return nil
	}

	pb := &mme.WearMechanism{}

	if m.WearMap != nil {
		pb.WearMap = make(map[int32]int32, len(m.WearMap))
		maps.Copy(pb.WearMap, m.WearMap)
	}
	confid := m.ConfId
	pb.ConfId = &confid
	return pb
}

func (m *WearMechanism) FromProto(pb *mme.WearMechanism) {
	if m == nil || pb == nil {
		return
	}

	if pb.WearMap != nil {
		m.WearMap = make(map[int32]int32, len(pb.WearMap))
		maps.Copy(m.WearMap, pb.WearMap)
	}
	if pb.ConfId != nil {
		m.ConfId = *pb.ConfId
	}
}

type WearMechanismWrapper struct {
	data *WearMechanism
	dirtyflag.IDirtyFlag
	fieldMetas *fieldmeta.FieldMetas
}

func NewWearMechanismWrapper(data *WearMechanism) *WearMechanismWrapper {
	if data == nil {
		panic("data is nil")
	}
	w := &WearMechanismWrapper{
		data:       data,
		IDirtyFlag: dirtyflag.NewDirtyFlag(),
		fieldMetas: fieldmeta.NewFieldMetas(),
	}

	return w
}

func (w *WearMechanismWrapper) InitFieldContext() {
	w.fieldMetas.SetFieldType(WearMechanismFieldIndexWearMap, fieldmeta.FieldTypeSync)
	w.fieldMetas.SetFieldType(WearMechanismFieldIndexConfId, fieldmeta.FieldTypeSync)
}

func (w *WearMechanismWrapper) Name() string {
	return "WearMechanism"
}

// MatchesAll 判断字段是否匹配所有类型标记
func (w *WearMechanismWrapper) MatchesAll(fieldID uint8, fieldTypes ...fieldmeta.FieldType) bool {
	return w.fieldMetas.MatchesAll(fieldID, fieldTypes...)
}

// 包装器-获取WearMap
func (w *WearMechanismWrapper) GetWearMap() map[int32]int32 {
	return w.data.WearMap
}

// 包装器-设置WearMap
func (w *WearMechanismWrapper) SetWearMap(v map[int32]int32) {
	w.data.WearMap = v
	w.MarkDirty(WearMechanismDirtyWearMapBit)
}

func (w *WearMechanismWrapper) GetConfId() int32 {
	return w.data.ConfId
}

func (w *WearMechanismWrapper) SetConfId(v int32) {
	w.data.ConfId = v
	w.MarkDirty(WearMechanismDirtyConfIdBit)
}

// ClearAllDirtyFlags 清除所有脏标记位
func (w *WearMechanismWrapper) ClearAllDirtyFlags() {
	w.IDirtyFlag.ClearAllDirty()

	// 如果Value为引用类型且有脏标记，则清除所有xmap中Value的脏标记

	// 清除所有xmap中的操作记录
}

// BuildMongoUpdate 构建MongoDB更新操作
func (w *WearMechanismWrapper) BuildMongoUpdate(builder *mgo_builder.MongoUpdateBuilder, path *mgo_builder.NestedPath) {
	if w == nil {
		return
	}

	if w.IsDirty(WearMechanismDirtyWearMapBit) {
		copy := make(map[int32]int32, len(w.data.WearMap))
		maps.Copy(copy, w.data.WearMap)
		builder.SetNestedPath(path, "wear_map", copy)
	}
	if w.IsDirty(WearMechanismDirtyConfIdBit) {
		builder.SetNestedPath(path, "conf_id", w.GetConfId())
	}
}

// ToIncrementalProtoWithContext 根据脏标记位构建增量数据的 protoMessage
// 只返回标记为脏的字段数据，用于增量同步
func (w *WearMechanismWrapper) ToIncrementalProtoWithContext(ctx mmemodel.SyncContext) proto.Message {
	if w == nil {
		return nil
	}

	// 如果没有脏标记，返回 nil
	if !w.HasAnyDirty() {
		return nil
	}

	incremental := &mme.WearMechanism{}

	if mmemodel.FieldCanBeIncrementalSynced(w, WearMechanismDirtyWearMapBit, WearMechanismFieldIndexWearMap, ctx) {
		m := w.GetWearMap()
		if m != nil {
			incremental.WearMap = make(map[int32]int32, len(m))
			maps.Copy(incremental.WearMap, m)
		}
	}
	if mmemodel.FieldCanBeIncrementalSynced(w, WearMechanismDirtyConfIdBit, WearMechanismFieldIndexConfId, ctx) {
		v := w.GetConfId()
		incremental.ConfId = &v
	}
	return incremental
}

// ToProto 将 WearMechanism 数据转换为完整的 protobuf 结构体
func (w *WearMechanismWrapper) ToProto() *mme.WearMechanism {
	if w == nil || w.data == nil {
		return nil
	}

	return w.data.ToProto()
}

// FromProto 从 protobuf 结构体加载数据到 WearMechanism
func (w *WearMechanismWrapper) FromProto(pb *mme.WearMechanism) {
	if w == nil || w.data == nil || pb == nil {
		return
	}

	if pb.WearMap != nil {
		maps.Copy(w.data.WearMap, pb.WearMap)
	}
	if pb.ConfId != nil {
		w.SetConfId(*pb.ConfId)
	}
}

// 包装器-深拷贝
func (w *WearMechanismWrapper) DeepCopy() *WearMechanism {
	if w == nil {
		return nil
	}
	copy := &WearMechanism{}
	w.DeepCopyTo(copy)
	return copy
}

// 包装器-深拷贝
func (w *WearMechanismWrapper) DeepCopyTo(copy *WearMechanism) {
	if w == nil || w.data == nil || copy == nil {
		return
	}

	// 拷贝基础类型字段
	*copy = *w.data

	// 初始化目标 map
	if copy.WearMap == nil {
		copy.WearMap = make(map[int32]int32, len(w.data.WearMap))
	}
	// 拷贝所有 int32
	if w.data.WearMap != nil {
		maps.Copy(copy.WearMap, w.data.WearMap)
	}
}
