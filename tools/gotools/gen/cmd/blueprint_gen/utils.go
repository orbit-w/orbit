package blueprint_gen

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ParseFieldType 解析字段类型字符串
// 支持: int32, int64, string, bool, float, double
// map<KeyType, ValueType>, xmap<KeyType, ValueType>, repeated Type
func ParseFieldType(typeStr string) (FieldType, error) {
	typeStr = strings.TrimSpace(typeStr)

	ft := FieldType{}

	// 检查是否是 repeated
	if strings.HasPrefix(typeStr, "repeated ") {
		ft.IsRepeated = true
		typeStr = strings.TrimPrefix(typeStr, "repeated ")
		typeStr = strings.TrimSpace(typeStr)
		ft.BaseType = typeStr
		return ft, nil
	}

	// 检查是否是 xmap
	if strings.HasPrefix(typeStr, "xmap<") {
		ft.IsXMap = true
		return parseMapType(typeStr[5:], &ft)
	}

	// 检查是否是 map
	if strings.HasPrefix(typeStr, "map<") {
		ft.IsMap = true
		return parseMapType(typeStr[4:], &ft)
	}

	// 基础类型
	ft.BaseType = typeStr
	return ft, nil
}

// parseMapType 解析 map<x, y> 或 xmap<x, y> 类型
func parseMapType(typeStr string, ft *FieldType) (FieldType, error) {
	// 去掉结尾的 >
	if !strings.HasSuffix(typeStr, ">") {
		return *ft, fmt.Errorf("invalid map type format: %s", typeStr)
	}
	typeStr = typeStr[:len(typeStr)-1]

	parts := strings.Split(typeStr, ",")
	if len(parts) != 2 {
		return *ft, fmt.Errorf("invalid map type format, expected key,value: %s", typeStr)
	}

	ft.KeyType = strings.TrimSpace(parts[0])
	ft.ValueType = strings.TrimSpace(parts[1])
	return *ft, nil
}

// ParseFieldOptions 解析字段选项字符串
// 格式: [blueprint:"access=all", orbit:"Access=all"]
func ParseFieldOptions(optionStr string) FieldOption {
	opt := FieldOption{}

	if optionStr == "" {
		return opt
	}

	// 使用正则表达式提取 access=xxx
	re := regexp.MustCompile(`access\s*=\s*([a-z]+)`)
	matches := re.FindStringSubmatch(optionStr)
	if len(matches) >= 2 {
		opt.Access = matches[1]
	} else {
		opt.Access = "all" // 默认值
	}

	return opt
}

// ParseFieldDefinition 解析字段定义行
// 格式: int32 FieldName: 1 [blueprint:"access=all"]
// 或者: HeroManager HeroManager: 1 [blueprint:"access=all"]
func ParseFieldDefinition(line string) (Field, error) {
	field := Field{}

	// 提取注释
	parts := strings.Split(line, "#")
	if len(parts) > 1 {
		field.Comment = strings.TrimSpace(parts[1])
		line = strings.TrimSpace(parts[0])
	}

	// 提取选项 [blueprint:"access=all"]
	var optionStr string
	if idx := strings.Index(line, "["); idx != -1 {
		optionStr = line[idx:]
		line = line[:idx]
		line = strings.TrimSpace(line)
	}
	field.Options = ParseFieldOptions(optionStr)

	// 解析字段编号 : N
	re := regexp.MustCompile(`:\s*(\d+)\s*`)
	matches := re.FindStringSubmatch(line)
	if len(matches) < 2 {
		return field, fmt.Errorf("field number not found in: %s", line)
	}

	var num int32
	if _, err := fmt.Sscanf(matches[1], "%d", &num); err != nil {
		return field, err
	}
	field.Number = num

	// 移除 :N 部分
	line = re.ReplaceAllString(line, "")
	line = strings.TrimSpace(line)

	// 分割类型和字段名
	parts = strings.Fields(line)
	if len(parts) < 2 {
		return field, fmt.Errorf("invalid field format: %s", line)
	}

	// 类型部分是前面的所有部分（可能包含空格，如 "Core.MMELocation"）
	// 字段名是最后一部分
	field.Name = parts[len(parts)-1]
	typeParts := parts[:len(parts)-1]
	typeStr := strings.Join(typeParts, " ")

	fieldType, err := ParseFieldType(typeStr)
	if err != nil {
		return field, err
	}
	field.Type = fieldType

	// 如果类型是消息类型（包含 .），设置 ValueType
	if strings.Contains(typeStr, ".") {
		field.Type.ValueType = typeStr
		field.Type.BaseType = ""
	}

	return field, nil
}

// EnsureDir 确保目录存在
func EnsureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}

// WriteFile 写入文件内容
func WriteFile(filePath, content string) error {
	if err := EnsureDir(filepath.Dir(filePath)); err != nil {
		return err
	}
	return os.WriteFile(filePath, []byte(content), 0644)
}

// FileExists 检查文件是否存在
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

// ToProtoType 将 YAML 类型转换为 Proto 类型
func ToProtoType(ft FieldType) string {
	if ft.IsRepeated {
		return ft.BaseType
	}
	if ft.IsMap || ft.IsXMap {
		return fmt.Sprintf("map<%s, %s>", ft.KeyType, ft.ValueType)
	}
	if ft.ValueType != "" && strings.Contains(ft.ValueType, ".") {
		// 消息类型，如 Core.MMELocation
		return ft.ValueType
	}
	return ft.BaseType
}

// ToGoType 将 YAML 类型转换为 Go 类型
func ToGoType(ft FieldType, packageName string) string {
	if ft.IsRepeated {
		// repeated Type -> []Type
		goType := toGoBaseType(ft.BaseType, packageName)
		return "[]" + goType
	}
	if ft.IsMap || ft.IsXMap {
		// map<Key, Value> -> map[Key]Value
		keyType := toGoBaseType(ft.KeyType, packageName)
		valueType := toGoBaseType(ft.ValueType, packageName)
		return fmt.Sprintf("map[%s]%s", keyType, valueType)
	}
	return toGoBaseType(ft.BaseType, packageName)
}

// toGoBaseType 转换基础类型
func toGoBaseType(typeStr string, packageName string) string {
	switch typeStr {
	case "int32":
		return "int32"
	case "int64":
		return "int64"
	case "string":
		return "string"
	case "bool":
		return "bool"
	case "float":
		return "float32"
	case "double":
		return "float64"
	default:
		// 消息类型
		if strings.Contains(typeStr, ".") {
			// 外部包类型，如 Core.MMELocation
			return typeStr
		}
		// MME 类型，添加包名前缀
		if packageName != "" {
			return fmt.Sprintf("*%s.%s", packageName, typeStr)
		}
		return "*" + typeStr
	}
}

// CamelToSnake 驼峰转蛇形
func CamelToSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteByte('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
