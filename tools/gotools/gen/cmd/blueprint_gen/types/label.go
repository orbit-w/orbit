package blueprint_types

// FieldLabel 字段标签类型（参考 descriptorpb 的设计）
type FieldLabel int32

const (
	FieldLabelUnknown  FieldLabel = 0 // 未知
	FieldLabelOptional FieldLabel = 1 // 可选字段
	FieldLabelRepeated FieldLabel = 2 // 重复字段
)

func (l FieldLabel) IsRepeated() bool {
	return l == FieldLabelRepeated
}

func (l FieldLabel) IsOptional() bool {
	return l == FieldLabelOptional
}

// String 返回字段标签的字符串表示
func (l FieldLabel) String() string {
	switch l {
	case FieldLabelOptional:
		return "OPTIONAL"
	case FieldLabelRepeated:
		return "REPEATED"
	default:
		return "UNKNOWN"
	}
}
