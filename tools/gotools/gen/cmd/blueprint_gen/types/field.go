package blueprint_types

import (
	"strings"
)

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
// 优先使用 Type.Name，如果为空则从 Type.TypeName 中提取（去掉包名前缀）
func (f *Field) GetTypeName() string {
	if f.Type.Name != "" {
		return f.Type.Name
	}

	// 如果 Name 为空，尝试从 TypeName 提取
	if f.Type.TypeName != "" {
		if idx := strings.LastIndex(f.Type.TypeName, "."); idx >= 0 {
			return f.Type.TypeName[idx+1:]
		}
		return f.Type.TypeName
	}

	return ""
}

func (f *Field) GetValueName() string {
	return f.Type.ValueName()
}

func (f *Field) GetValueType() *FieldType {
	return f.Type.ValueType
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
