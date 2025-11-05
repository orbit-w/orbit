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

// GetTypeName 获取字段类型名称，如 map, xmap, repeated, HeroManager等基础类型或MMEObject类型
func (f *Field) GetTypeName() string {
	return f.Type.Name
}

func (f *Field) IsMapField() bool {
	return f.Type.IsMapField()
}

func (f *Field) IsXMapField() bool {
	return f.Type.IsXMapField()
}

func (f *Field) IsMMEObjectType() bool {
	return f.Type.IsMMEObjectType()
}

// FieldOption 字段选项
type FieldOption struct {
	Access string // access=all/s/c

	// 扩展选项（自定义选项）
	CustomOptions map[string]any
}
