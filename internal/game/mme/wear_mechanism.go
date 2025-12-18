package mme

import (
	"gitee.com/orbit-w/meteor/bases/container/xmap"
	dirtyflag "gitee.com/orbit-w/orbit/lib/base/dirty_flag"
	fieldmeta "gitee.com/orbit-w/orbit/lib/base/field_meta"
	"gitee.com/orbit-w/orbit/lib/module/db/mgo_builder"
	mmemodel "gitee.com/orbit-w/orbit/lib/module/mme_model"
	"gitee.com/orbit-w/orbit/pkg/proto/mme"
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

	// Value 为值类型，使用 xmap.MapAccessor 进行包装
	wearMapAccessor *xmap.MapAccessor[int32, int32]
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

	w.wearMapAccessor = xmap.NewMapAccessorWithMarker(&w.data.WearMap, w, WearMechanismDirtyWearMapBit)
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

// 包装器-获取WearMap访问器
func (w *WearMechanismWrapper) GetWearMapAccessor() *xmap.MapAccessor[int32, int32] {
	return w.wearMapAccessor
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
	w.wearMapAccessor.ResetOperations()
}

// BuildMongoUpdate 构建MongoDB更新操作
func (w *WearMechanismWrapper) BuildMongoUpdate(builder *mgo_builder.MongoUpdateBuilder, path *mgo_builder.NestedPath) {
	if w == nil {
		return
	}

	if w.IsDirty(WearMechanismDirtyWearMapBit) {
		builder.SetNestedPath(path, "wear_map", w.wearMapAccessor.Clone())
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
		incremental.WearMap = w.wearMapAccessor.Clone()
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
		// 清空现有的数据
		temp := make(map[int32]int32, len(pb.WearMap))
		maps.Copy(temp, pb.WearMap)
		// 设置新的数据，并清空所有变化操作记录
		w.wearMapAccessor.Reset(&temp)
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

	// 拷贝 WearMapAccessor
	if w.wearMapAccessor != nil {
		if copy.WearMap == nil {
			copy.WearMap = make(map[int32]int32, w.wearMapAccessor.Len())
		}
		w.wearMapAccessor.Copy(copy.WearMap)
	}
}
