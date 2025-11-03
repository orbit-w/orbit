package blueprint_types

// FieldMetadata 字段的深度元数据（参考 descriptorpb.FieldDescriptorProto 的结构）
type FieldMetadata struct {
	// 基础字段信息
	Name         string
	Number       int32
	Label        FieldLabel
	Kind         FieldKind
	TypeName     string // 消息类型名称
	Extendee     string // 扩展字段的目标类型
	DefaultValue string // 默认值
	JsonName     string // JSON 名称

	Retention string

	// 扩展信息
	CustomOptions map[string]any

	// 类型相关元数据
	MapKeyKind   FieldKind // map key 类型（如果是 map）
	MapValueKind FieldKind // map value 类型（如果是 map）
	MapValueType string    // map value 消息类型名称
}

// BuildFieldMetadata 构建完整的字段元数据（深度解析）
func BuildFieldMetadata(field Field) *FieldMetadata {
	meta := &FieldMetadata{
		Name:          field.Name,
		Number:        field.Number,
		Label:         field.Type.Label,
		Kind:          field.Type.Kind,
		TypeName:      field.Type.TypeName,
		Extendee:      field.Type.Extendee,
		DefaultValue:  field.Type.DefaultValue,
		JsonName:      field.Type.JsonName,
		CustomOptions: field.Options.CustomOptions,
	}

	// 如果是 Map/XMap 类型，设置 Key 和 Value 类型信息
	tp := field.Type
	if tp.Kind == FieldKindMap || tp.Kind == FieldKindXMap {
		if tp.KeyType != nil {
			meta.MapKeyKind = tp.KeyType.Kind
		}
		if tp.ValueType != nil {
			meta.MapValueKind = tp.ValueType.Kind
			if tp.ValueType.TypeName != "" {
				meta.MapValueType = tp.ValueType.TypeName
			} else if tp.ValueType.Kind != FieldKindUnknown {
				// 对于基础类型，使用字符串表示
				meta.MapValueType = tp.ValueType.String()
			}
		}
	}

	// 如果是 Repeated 类型，ValueType 存储元素类型
	// 注意：对于 repeated 类型，MapValueKind 和 MapValueType 也可以用来存储元素类型信息
	if field.Type.Kind == FieldKindRepeated && field.Type.ValueType != nil {
		meta.MapValueKind = field.Type.ValueType.Kind
		if field.Type.ValueType.TypeName != "" {
			meta.MapValueType = field.Type.ValueType.TypeName
		} else if field.Type.ValueType.Kind != FieldKindUnknown {
			meta.MapValueType = field.Type.ValueType.String()
		}
	}

	return meta
}

// InferFieldLabel 从字段类型信息推断字段标签
func InferFieldLabel(fieldType FieldType) FieldLabel {
	if fieldType.Kind == FieldKindRepeated {
		return FieldLabelRepeated
	}
	// 根据业务逻辑，默认是 OPTIONAL
	return FieldLabelOptional
}
