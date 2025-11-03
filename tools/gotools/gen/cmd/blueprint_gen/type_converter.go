package blueprint_gen

import (
	"strings"

	types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"
)

// TypeConverter 类型转换器，统一管理所有类型转换逻辑
type TypeConverter struct {
	packageName string
}

// NewTypeConverter 创建新的类型转换器
func NewTypeConverter(packageName string) *TypeConverter {
	return &TypeConverter{
		packageName: packageName,
	}
}

// SetPackageName 设置包名
func (tc *TypeConverter) SetPackageName(packageName string) {
	tc.packageName = packageName
}

// ToProtoType 将 FieldType 转换为 Proto 类型字符串
func (tc *TypeConverter) ToProtoType(ft *types.FieldType) string {
	return ToProtoTypeFromTypesFieldType(ft)
}

// ToGoType 将 FieldType 转换为 Go 类型字符串
func (tc *TypeConverter) ToGoType(ft *types.FieldType) string {
	return ToGoTypeFromTypesFieldType(ft, tc.packageName)
}

// ToGoBaseType 将 FieldType 转换为 Go 基础类型（不带包名）
func (tc *TypeConverter) ToGoBaseType(ft *types.FieldType) string {
	return toGoBaseTypeFromFieldType(ft)
}

// IsOptionalInProto 判断字段在 Proto 中是否需要 optional 标记
func (tc *TypeConverter) IsOptionalInProto(ft *types.FieldType) bool {
	if ft == nil {
		return false
	}

	// map, xmap, repeated 类型不需要 optional
	if ft.Kind == types.FieldKindMap ||
		ft.Kind == types.FieldKindXMap ||
		ft.Kind == types.FieldKindRepeated {
		return false
	}

	// 消息类型（包含 . ）不需要 optional
	protoType := tc.ToProtoType(ft)
	if strings.Contains(protoType, ".") {
		return false
	}

	// 基础类型需要 optional
	return true
}

// FieldKindToProtoString 将 FieldKind 转换为 Proto 类型字符串
func (tc *TypeConverter) FieldKindToProtoString(kind types.FieldKind) string {
	switch kind {
	case types.FieldKindInt32:
		return "int32"
	case types.FieldKindInt64:
		return "int64"
	case types.FieldKindUInt32:
		return "uint32"
	case types.FieldKindUInt64:
		return "uint64"
	case types.FieldKindFloat:
		return "float"
	case types.FieldKindDouble:
		return "double"
	case types.FieldKindBool:
		return "bool"
	case types.FieldKindString:
		return "string"
	case types.FieldKindBytes:
		return "bytes"
	default:
		return "unknown"
	}
}

// FieldKindToGoString 将 FieldKind 转换为 Go 类型字符串
func (tc *TypeConverter) FieldKindToGoString(kind types.FieldKind) string {
	switch kind {
	case types.FieldKindInt32:
		return "int32"
	case types.FieldKindInt64:
		return "int64"
	case types.FieldKindUInt32:
		return "uint32"
	case types.FieldKindUInt64:
		return "uint64"
	case types.FieldKindFloat:
		return "float32"
	case types.FieldKindDouble:
		return "float64"
	case types.FieldKindBool:
		return "bool"
	case types.FieldKindString:
		return "string"
	case types.FieldKindBytes:
		return "[]byte"
	default:
		return "unknown"
	}
}

// IsMMEObjectType 判断类型是否是 MME Object 类型
func (tc *TypeConverter) IsMMEObjectType(ft *types.FieldType) bool {
	if ft == nil {
		return false
	}

	// 如果是 Map/XMap，检查 ValueType
	if ft.Kind == types.FieldKindMap || ft.Kind == types.FieldKindXMap {
		return tc.IsMMEObjectType(ft.ValueType)
	}

	// 检查 TypeName
	if ft.TypeName == "" {
		return false
	}

	// 检查是否是 Module 或 Mechanism 类型
	typeName := ft.TypeName
	// 去掉包名前缀
	if idx := strings.LastIndex(typeName, "."); idx >= 0 {
		typeName = typeName[idx+1:]
	}

	// MME Object 类型的后缀
	suffixes := []string{"Module", "Mechanism", "Manager", "Entity"}
	for _, suffix := range suffixes {
		if strings.HasSuffix(typeName, suffix) {
			return true
		}
	}

	return false
}

// GetMMEObjectName 获取 MME Object 的名称（不带包名）
func (tc *TypeConverter) GetMMEObjectName(ft *types.FieldType) string {
	if ft == nil || ft.TypeName == "" {
		return ""
	}

	typeName := ft.TypeName
	if idx := strings.LastIndex(typeName, "."); idx >= 0 {
		typeName = typeName[idx+1:]
	}

	return typeName
}

// FormatProtoMessage 格式化 Proto message 定义
func (tc *TypeConverter) FormatProtoMessage(name string, fields []*types.Field) string {
	builder := NewCodeBuilder()
	builder.SetIndentStr("  ")

	builder.WriteLine("message %s {", name)
	builder.Indent()

	// 按编号排序字段
	sortedFields := make([]*types.Field, len(fields))
	copy(sortedFields, fields)
	// 简单冒泡排序
	for i := 0; i < len(sortedFields)-1; i++ {
		for j := i + 1; j < len(sortedFields); j++ {
			if sortedFields[i].Number > sortedFields[j].Number {
				sortedFields[i], sortedFields[j] = sortedFields[j], sortedFields[i]
			}
		}
	}

	for _, field := range sortedFields {
		protoType := tc.ToProtoType(&field.Type)
		isOptional := tc.IsOptionalInProto(&field.Type)

		// 添加注释
		if field.Comment != "" {
			builder.WriteLine("// %s", field.Comment)
		}

		// 构建字段定义
		if isOptional {
			builder.WriteLine("optional %s %s = %d;", protoType, field.Name, field.Number)
		} else {
			builder.WriteLine("%s %s = %d;", protoType, field.Name, field.Number)
		}
	}

	builder.Unindent()
	builder.WriteLine("}")

	return builder.String()
}

