package blueprint_types

import mmeobject "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types/mme_obj"

// FieldMetadata 字段的深度元数据（参考 descriptorpb.FieldDescriptorProto 的结构）
type FieldMetadata struct {
	*Field
	IsMMEObject   bool                 // 是否是MMEObject类型
	MMEObjectType mmeobject.ObjectType // 如果是MMEObject类型，则记录其类型

	// 扩展信息
	CustomOptions map[string]any
}

func (meta *FieldMetadata) SetMMEObjectKind(mmeObjectKind mmeobject.ObjectType) {
	meta.IsMMEObject = true
	meta.MMEObjectType = mmeObjectKind
}

func BuildFieldMetadata(field *Field) *FieldMetadata {
	meta := &FieldMetadata{
		Field:         field,
		IsMMEObject:   false,
		CustomOptions: field.Options.CustomOptions,
	}
	buildFieldMetadata(meta, field)
	return meta
}

// buildFieldMetadata 构建完整的字段元数据（深度解析）
func buildFieldMetadata(meta *FieldMetadata, field *Field) {

}

// InferFieldLabel 从字段类型信息推断字段标签
func InferFieldLabel(fieldType FieldType) FieldLabel {
	if fieldType.Kind == FieldKindRepeated {
		return FieldLabelRepeated
	}
	// 根据业务逻辑，默认是 OPTIONAL
	return FieldLabelOptional
}
