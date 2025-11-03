package blueprint_types

import (
	"fmt"
	"strings"
)

// FieldType 字段类型定义 - 增强以支持深度元数据解析
type FieldType struct {
	Kind         FieldKind  // 字段类型种类（参考 descriptorpb）
	Label        FieldLabel // 字段标签（OPTIONAL, REQUIRED, REPEATED）
	TypeName     string     // 消息类型完整名称（带包名），如 "mme.HeroManager"
	Extendee     string     // 如果是扩展字段，扩展的消息类型名称
	DefaultValue string     // 默认值
	JsonName     string     // JSON 名称（如果与字段名不同）

	// Map/XMap/Repeated 类型信息（递归类型支持）
	KeyType   *FieldType // map/xmap 的 key 类型，nil 表示无（对于 repeated 类型）
	ValueType *FieldType // map/xmap 的 value 类型，或 repeated 的元素类型
}

// FieldKind 字段类型种类（参考 descriptorpb 的设计）
type FieldKind int32

const (
	FieldKindUnknown   FieldKind = 0
	FieldKindDouble    FieldKind = 1  // double
	FieldKindFloat     FieldKind = 2  // float
	FieldKindInt64     FieldKind = 3  // int64
	FieldKindUInt64    FieldKind = 4  // uint64
	FieldKindInt32     FieldKind = 5  // int32
	FieldKindFixed64   FieldKind = 6  // fixed64
	FieldKindFixed32   FieldKind = 7  // fixed32
	FieldKindBool      FieldKind = 8  // bool
	FieldKindString    FieldKind = 9  // string
	FieldKindMessage   FieldKind = 11 // message
	FieldKindBytes     FieldKind = 12 // bytes
	FieldKindUInt32    FieldKind = 13 // uint32
	FieldKindEnum      FieldKind = 14 // enum
	FieldKindMMEObject FieldKind = 16 // MME对象类型
	FieldKindMap       FieldKind = 18 // map
	FieldKindXMap      FieldKind = 19 // xmap
	FieldKindRepeated  FieldKind = 20 // repeated
)

// String 返回字段类型的字符串表示
func (k FieldKind) String() string {
	switch k {
	case FieldKindDouble:
		return "double"
	case FieldKindFloat:
		return "float"
	case FieldKindInt64:
		return "int64"
	case FieldKindUInt64:
		return "uint64"
	case FieldKindInt32:
		return "int32"
	case FieldKindFixed64:
		return "fixed64"
	case FieldKindFixed32:
		return "fixed32"
	case FieldKindBool:
		return "bool"
	case FieldKindString:
		return "string"
	case FieldKindMessage:
		return "message"
	case FieldKindBytes:
		return "bytes"
	case FieldKindUInt32:
		return "uint32"
	case FieldKindEnum:
		return "enum"
	case FieldKindMMEObject:
		return "MMEObject"
	case FieldKindMap:
		return "map"
	case FieldKindXMap:
		return "xmap"
	case FieldKindRepeated:
		return "repeated"
	default:
		return "unknown"
	}
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
	switch typeName {
	case "double":
		return FieldKindDouble
	case "float":
		return FieldKindFloat
	case "int64":
		return FieldKindInt64
	case "uint64":
		return FieldKindUInt64
	case "int32":
		return FieldKindInt32
	case "fixed64":
		return FieldKindFixed64
	case "fixed32":
		return FieldKindFixed32
	case "bool":
		return FieldKindBool
	case "string":
		return FieldKindString
	case "bytes":
		return FieldKindBytes
	case "uint32":
		return FieldKindUInt32
	case "enum":
		return FieldKindEnum
	case "MMEObject":
		return FieldKindMMEObject
	case "map":
		return FieldKindMap
	case "xmap":
		return FieldKindXMap
	case "repeated":
		return FieldKindRepeated
	case "message":
		return FieldKindMessage
	default:
		return FieldKindUnknown
	}
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
		return parseMapTypeRecursive(typeStr[5:], ft)
	}

	// 检查是否是 map
	if strings.HasPrefix(typeStr, "map<") {
		ft.Kind = FieldKindMap
		ft.Label = FieldLabelOptional
		return parseMapTypeRecursive(typeStr[4:], ft)
	}

	// 基础类型或消息类型
	return parseBaseOrMessageType(typeStr, ft), nil
}

// parseRepeatedType 解析 repeated 类型
func parseRepeatedType(typeStr string, ft *FieldType) (*FieldType, error) {
	ft.Kind = FieldKindRepeated
	ft.Label = FieldLabelRepeated

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

// parseBaseOrMessageType 解析基础类型或消息类型
func parseBaseOrMessageType(typeStr string, ft *FieldType) *FieldType {
	ft.Label = FieldLabelOptional

	// 首先尝试识别基础类型
	ft.Kind = FieldKindFromString(typeStr)

	// 如果成功识别为基础类型，直接返回
	if ft.Kind != FieldKindUnknown {
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
		return ft
	}

	// 无法识别为基础类型或 MME Object，作为普通消息类型处理
	// 包含点号的通常是完整消息类型名称（如 "mme.HeroManager"）
	// 不包含点号的可能是自定义类型名称
	ft.Kind = FieldKindMessage
	ft.TypeName = typeStr
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
