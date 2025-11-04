package blueprint_types

// Field 字段定义 - 增强以支持深度元数据解析
type Field struct {
	Name    string
	Type    FieldType
	Number  int32 // 字段编号
	Options FieldOption
	Comment string

	// 深度元数据（可选，用于存储完整的解析信息）
	Metadata *FieldMetadata
}

func (f *Field) IsMapField() bool {
	return f.Type.Kind == FieldKindMap
}

func (f *Field) IsXMapField() bool {
	return f.Type.Kind == FieldKindXMap
}

// FieldOption 字段选项
type FieldOption struct {
	Access string // access=all/s/c

	// 扩展选项（自定义选项）
	CustomOptions map[string]any
}
