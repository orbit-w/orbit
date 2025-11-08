package blueprint_types

import (
	"fmt"
	"strings"
)

// FieldType 字段类型定义 - 增强以支持深度元数据解析
type FieldType struct {
	Kind         FieldKind  // 字段类型种类（参考 descriptorpb）
	Label        FieldLabel // 字段标签（OPTIONAL, REQUIRED, REPEATED）
	Name         string     // 类型名称
	TypeName     string     // 消息类型完整名称（带包名），如 "mme.HeroManager"
	Extendee     string     // 如果是扩展字段，扩展的消息类型名称
	DefaultValue string     // 默认值
	JsonName     string     // JSON 名称（如果与字段名不同）

	// Map/XMap/Repeated 类型信息（递归类型支持）
	KeyType   *FieldType // map/xmap 的 key 类型，nil 表示无（对于 repeated 类型）
	ValueType *FieldType // map/xmap 的 value 类型，或 repeated 的元素类型
}

func (ft FieldType) GetName() string {
	return ft.Name
}

// GetKind 获取字段类型种类
func (ft FieldType) GetKind() FieldKind {
	return ft.Kind
}

// KeyKind 获取map/xmap的key类型种类
func (ft FieldType) KeyKind() FieldKind {
	return ft.KeyType.Kind
}

// ValueKind 获取map/xmap/repeated的值类型种类
func (ft FieldType) ValueKind() FieldKind {
	return ft.ValueType.Kind
}

func (ft FieldType) GetValueType() *FieldType {
	return ft.ValueType
}

// ValueName 获取map/xmap/repeated的值类型名称，如 HeroManager等MMEObject类型
func (ft FieldType) ValueName() string {
	return ft.ValueType.Name
}

// ValueTypeName 获取map/xmap/repeated的值类型完整名称，如 mme.HeroManager等MMEObject类型
func (ft FieldType) ValueTypeName() string {
	return ft.ValueType.TypeName
}

// IsMapField 判断字段类型是否是map类型
func (ft FieldType) IsMapField() bool {
	return ft.Kind == FieldKindMap
}

// IsXMapField 判断字段类型是否是xmap类型
func (ft FieldType) IsXMapField() bool {
	return ft.Kind == FieldKindXMap
}

// IsMMEObjectType 判断字段类型是否是MMEObject类型
func (ft FieldType) IsMMEObjectType() bool {
	return ft.Kind == FieldKindMMEObject
}

// IsMessage 判断字段类型是否是Message类型
func (ft FieldType) IsMessage() bool {
	return ft.Kind == FieldKindMessage
}

func (ft FieldType) GetTypeName() string {
	if ft.TypeName != "" {
		return ft.TypeName
	}
	return ft.Kind.String()
}

// isXMapValueMMEObject 判断xmap的Value类型是否是MMEObject
// xmap Value不允许是map/xmap/repeated类型
func (ft FieldType) IsXMapValueMMEObject() bool {
	if !ft.IsXMapField() {
		panic("field type is not xmap field")
	}
	if ft.ValueType == nil {
		panic("field value type is nil")
	}

	return ft.ValueType.Kind == FieldKindMMEObject
}

func (ft FieldType) IsMapValueMMEObject() bool {
	if !ft.IsMapField() {
		panic("field type is not map field")
	}
	if ft.ValueType == nil {
		panic("field value type is nil")
	}

	return ft.ValueType.Kind == FieldKindMMEObject
}

// 判断字段类型是基础类型还是MMEObject
// 基础类型包括：int32, int64, uint32, uint64, float, double, bool, string, bytes
func (ft FieldType) IsFieldBaseType() bool {
	switch ft.Kind {
	case FieldKindInt32,
		FieldKindInt64,
		FieldKindUInt32,
		FieldKindUInt64,
		FieldKindFloat,
		FieldKindDouble,
		FieldKindBool,
		FieldKindString,
		FieldKindBytes:
		return true
	case FieldKindMMEObject,
		FieldKindMessage:
		return false
	default:
		panic(fmt.Sprintf("unknown field kind: %s", ft.Kind))
	}
}

// FieldKind 字段类型种类（参考 descriptorpb 的设计）
type FieldKind int32

func (k FieldKind) IsBaseType() bool {
	return k == FieldKindInt32 ||
		k == FieldKindInt64 ||
		k == FieldKindUInt32 ||
		k == FieldKindUInt64 ||
		k == FieldKindFloat ||
		k == FieldKindDouble ||
		k == FieldKindBool ||
		k == FieldKindString ||
		k == FieldKindBytes ||
		k == FieldKindEnum
}

func (k FieldKind) IsMMEObject() bool {
	return k == FieldKindMMEObject
}

func (k FieldKind) IsMessage() bool {
	return k == FieldKindMessage
}

func (k FieldKind) IsMap() bool {
	return k == FieldKindMap
}

// String 返回字段类型的字符串表示
func (k FieldKind) String() string {
	str, ok := FieldKindStringMap[k]
	if !ok {
		return FieldKindUnknownString
	}
	return str

}

// FieldLabel 字段标签类型（参考 descriptorpb 的设计）
type FieldLabel int32

const (
	FieldLabelUnknown  FieldLabel = 0 // 未知
	FieldLabelOptional FieldLabel = 1 // 可选字段
	FieldLabelRepeated FieldLabel = 2 // 重复字段
)

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

// FieldKindFromString 从字符串类型名称获取 FieldKind（深度类型推断）
func FieldKindFromString(typeName string) FieldKind {
	kind, ok := FieldKindStringToKindMap[strings.TrimSpace(typeName)]
	if !ok {
		return FieldKindUnknown
	}
	return kind
}

// ParseTypeString 递归解析类型字符串，支持嵌套的 map/xmap/repeated 类型
// 例如: "int32", "map<int32, string>", "map<int32, map<string, int64>>", "repeated map<string, MessageType>"
func ParseTypeString(typeStr string) (*FieldType, error) {
	typeStr = strings.TrimSpace(typeStr)
	if typeStr == "" {
		return nil, fmt.Errorf("empty type string")
	}

	ft := &FieldType{}

	// 检查是否是 repeated
	if strings.HasPrefix(typeStr, "repeated ") {
		return parseRepeatedType(typeStr, ft)
	}

	// 检查是否是 xmap
	if strings.HasPrefix(typeStr, "xmap<") {
		ft.Kind = FieldKindXMap
		ft.Label = FieldLabelOptional
		ft.Name = "xmap"
		return parseMapTypeRecursive(typeStr[5:], ft)
	}

	// 检查是否是 map
	if strings.HasPrefix(typeStr, "map<") {
		ft.Kind = FieldKindMap
		ft.Label = FieldLabelOptional
		ft.Name = "map"
		return parseMapTypeRecursive(typeStr[4:], ft)
	}

	// 基础类型或消息类型
	return parseBaseOrStructType(typeStr, ft), nil
}

// parseRepeatedType 解析 repeated 类型
func parseRepeatedType(typeStr string, ft *FieldType) (*FieldType, error) {
	ft.Kind = FieldKindRepeated
	ft.Label = FieldLabelRepeated
	ft.Name = "repeated"

	// 提取内部类型字符串
	innerTypeStr := strings.TrimSpace(typeStr[8:])
	if innerTypeStr == "" {
		return nil, fmt.Errorf("repeated type missing element type")
	}

	// 递归解析元素类型
	valueType, err := ParseTypeString(innerTypeStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse repeated element type: %w", err)
	}
	if valueType == nil {
		return nil, fmt.Errorf("repeated element type is nil")
	}

	ft.ValueType = valueType
	return ft, nil
}

// extractTypeName 从完整类型名称中提取不带包名的类型名称
// 例如: "mme.HeroManager" -> "HeroManager", "HeroManager" -> "HeroManager"
func extractTypeName(fullTypeName string) string {
	if lastDot := strings.LastIndex(fullTypeName, "."); lastDot >= 0 {
		return fullTypeName[lastDot+1:]
	}
	return fullTypeName
}

// parseBaseOrStructType 解析基础类型或消息类型
func parseBaseOrStructType(typeStr string, ft *FieldType) *FieldType {
	ft.Label = FieldLabelOptional

	// 首先尝试识别基础类型
	ft.Kind = FieldKindFromString(typeStr)

	// 如果成功识别为基础类型，直接返回
	if ft.Kind != FieldKindUnknown {
		// 对于基础类型，设置 Name 为类型名称
		ft.Name = typeStr
		// 对于已知的基础类型，不需要设置 TypeName
		// 但如果是 enum 或 MMEObject 等需要名称的类型，保留原逻辑
		if ft.Kind == FieldKindEnum || ft.Kind == FieldKindMMEObject {
			ft.TypeName = typeStr
		}
		return ft
	}

	// 检查是否是 MME Object 类型（Entity、Manager、Module、Mechanism）
	if isMMEObjectType(typeStr) {
		ft.Kind = FieldKindMMEObject
		ft.TypeName = typeStr
		ft.Name = extractTypeName(typeStr)
		return ft
	}

	// 无法识别为基础类型或 MME Object，作为普通消息类型处理
	// 包含点号的通常是完整消息类型名称（如 "mme.CommonMessage"）
	// 不包含点号的可能是自定义类型名称
	ft.Kind = FieldKindMessage
	ft.TypeName = typeStr
	ft.Name = extractTypeName(typeStr)
	return ft
}

// isMMEObjectType 判断类型名称是否是 MME Object 类型
// MME Object 包括：Entity（实体）、Manager（管理器）、Module（模块）、Mechanism（机制）
// 命名模式通常以这些后缀结尾，如：PlayerEntity, HeroManager, HeroModule, HeroMechanism
// 也支持带包名的限定名称，如：mme.HeroManager, core.PlayerEntity
func isMMEObjectType(typeName string) bool {
	// MME Object 类型的后缀
	mmeObjectSuffixes := []string{"Entity", "Manager", "Module", "Mechanism"}

	// 提取类型名称（去掉包名前缀）
	typeNameOnly := typeName
	if lastDot := strings.LastIndex(typeName, "."); lastDot >= 0 {
		typeNameOnly = typeName[lastDot+1:]
	}

	// 检查是否以 MME Object 后缀结尾
	for _, suffix := range mmeObjectSuffixes {
		if strings.HasSuffix(typeNameOnly, suffix) {
			return true
		}
	}

	return false
}

// parseMapTypeRecursive 递归解析 map/xmap 的类型参数
// 输入: "KeyType, ValueType>" 或 "KeyType, map<...>>" 等嵌套情况
func parseMapTypeRecursive(typeStr string, ft *FieldType) (*FieldType, error) {
	// 去掉结尾的 >
	if !strings.HasSuffix(typeStr, ">") {
		return nil, fmt.Errorf("invalid map type format: missing closing '>'")
	}
	typeStr = typeStr[:len(typeStr)-1]

	// 查找分隔 Key 和 Value 的逗号，需要处理嵌套的 <>
	keyTypeStr, valueTypeStr, err := splitMapTypes(typeStr)
	if err != nil {
		return nil, fmt.Errorf("invalid map type format: %w", err)
	}

	// 递归解析 Key 类型
	keyTypeStr = strings.TrimSpace(keyTypeStr)
	keyType, err := ParseTypeString(keyTypeStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse key type '%s': %w", keyTypeStr, err)
	}
	ft.KeyType = keyType

	// 递归解析 Value 类型
	valueTypeStr = strings.TrimSpace(valueTypeStr)
	valueType, err := ParseTypeString(valueTypeStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse value type '%s': %w", valueTypeStr, err)
	}
	ft.ValueType = valueType

	return ft, nil
}

// splitMapTypes 分割 map 类型的 Key 和 Value，处理嵌套的 <>
func splitMapTypes(typeStr string) (string, string, error) {
	depth := 0
	commaPos := -1

	for i, char := range typeStr {
		switch char {
		case '<':
			depth++
		case '>':
			depth--
			if depth < 0 {
				return "", "", fmt.Errorf("unmatched '>'")
			}
		case ',':
			if depth == 0 && commaPos == -1 {
				commaPos = i
			}
		}
	}

	if commaPos == -1 {
		return "", "", fmt.Errorf("no comma found to separate key and value types")
	}

	if depth != 0 {
		return "", "", fmt.Errorf("unmatched '<' or '>'")
	}

	keyStr := typeStr[:commaPos]
	valueStr := typeStr[commaPos+1:]
	return keyStr, valueStr, nil
}

// String 返回 FieldType 的字符串表示（用于调试）
func (ft *FieldType) String() string {
	if ft == nil {
		return "nil"
	}

	switch ft.Kind {
	case FieldKindMap, FieldKindXMap:
		if ft.KeyType == nil || ft.ValueType == nil {
			return "invalid map"
		}
		prefix := "map"
		if ft.Kind == FieldKindXMap {
			prefix = "xmap"
		}
		return fmt.Sprintf("%s<%s, %s>", prefix, ft.KeyType.String(), ft.ValueType.String())
	case FieldKindRepeated:
		if ft.ValueType == nil {
			return "invalid repeated"
		}
		return fmt.Sprintf("repeated %s", ft.ValueType.String())
	default:
		if ft.TypeName != "" {
			return ft.TypeName
		}
		return ft.Kind.String()
	}
}
