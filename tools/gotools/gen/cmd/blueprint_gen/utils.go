package blueprint_gen

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"
)

// ParseFieldOptions 解析字段选项字符串
// 格式: [blueprint:"access=all", orbit:"Access=all"]
func ParseFieldOptions(optionStr string) types.FieldOption {
	opt := types.FieldOption{}

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

// ParseFieldDefinition 解析字段定义行，直接返回 *types.Field
// 格式: int32 FieldName: 1 [blueprint:"access=all"]
// 或者: HeroManager HeroManager: 1 [blueprint:"access=all"]
func ParseFieldDefinition(line string) (*types.Field, error) {
	field := &types.Field{}

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
		return nil, fmt.Errorf("field number not found in: %s", line)
	}

	var num int32
	if _, err := fmt.Sscanf(matches[1], "%d", &num); err != nil {
		return nil, err
	}
	field.Number = num

	// 移除 :N 部分
	line = re.ReplaceAllString(line, "")
	line = strings.TrimSpace(line)

	// 分割类型和字段名
	parts = strings.Fields(line)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid field format: %s", line)
	}

	// 类型部分是前面的所有部分（可能包含空格，如 "Core.MMELocation"）
	// 字段名是最后一部分
	field.Name = parts[len(parts)-1]
	typeParts := parts[:len(parts)-1]
	typeStr := strings.Join(typeParts, " ")

	// 使用 types.ParseTypeString 解析类型
	fieldType, err := types.ParseTypeString(typeStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse type: %w", err)
	}
	field.Type = *fieldType

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

// FormatGoFiles 格式化指定目录下的所有 Go 文件
// 使用 go fmt 命令格式化整个目录及其子目录下的所有 Go 文件
func FormatGoFiles(dir string) error {
	// 检查目录是否存在
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", dir)
	}

	// 获取绝对路径，确保 go fmt 能正确工作
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// 执行 go fmt 格式化整个目录
	// go fmt ./... 会格式化当前目录及所有子目录下的 Go 文件
	cmd := exec.Command("go", "fmt", "./...")
	cmd.Dir = absDir

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to format Go files: %w, output: %s", err, string(output))
	}

	return nil
}

// ToProtoType 将 FieldType 转换为 Proto 类型（统一使用 *types.FieldType）
func ToProtoType(ft *types.FieldType) string {
	return ToProtoTypeFromTypesFieldType(ft)
}

// ToProtoTypeFromTypesFieldType 将新的 types.FieldType 转换为 Proto 类型
func ToProtoTypeFromTypesFieldType(ft *types.FieldType) string {
	if ft == nil {
		return "unknown"
	}

	switch ft.Kind {
	case types.FieldKindRepeated:
		if ft.ValueType != nil {
			return ToProtoTypeFromTypesFieldType(ft.ValueType)
		}
		return "unknown"
	case types.FieldKindMap, types.FieldKindXMap:
		if ft.KeyType != nil && ft.ValueType != nil {
			keyType := ToProtoTypeFromTypesFieldType(ft.KeyType)
			valueType := ToProtoTypeFromTypesFieldType(ft.ValueType)
			return fmt.Sprintf("map<%s, %s>", keyType, valueType)
		}
		return "map<unknown, unknown>"
	default:
		if ft.TypeName != "" {
			return ft.TypeName
		}
		// 基础类型
		return ft.Kind.String()
	}
}

// ToGoType 将 FieldType 转换为 Go 类型（统一使用 *types.FieldType）
func ToGoType(ft *types.FieldType, packageName string) string {
	if ft == nil {
		return "unknown"
	}

	switch ft.Kind {
	case types.FieldKindRepeated:
		if ft.ValueType != nil {
			goType := ToGoType(ft.ValueType, packageName)
			return "[]" + goType
		}
		return "[]unknown"
	case types.FieldKindMap, types.FieldKindXMap:
		if ft.KeyType != nil && ft.ValueType != nil {
			keyType := ToGoType(ft.KeyType, packageName)
			valueType := ToGoType(ft.ValueType, packageName)
			return fmt.Sprintf("map[%s]%s", keyType, valueType)
		}
		return "map[unknown]unknown"
	default:
		if ft.TypeName != "" {
			// 消息类型
			if strings.Contains(ft.TypeName, ".") {
				// 外部包类型，如 Core.MMELocation
				return ft.TypeName
			}
			// MME 类型，添加包名前缀
			if packageName != "" {
				return fmt.Sprintf("*%s.%s", packageName, ft.TypeName)
			}
			return "*" + ft.TypeName
		}
		// 基础类型
		return ft.GetKind().String()
	}
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

// GetMechanismFileName 生成 Mechanism 文件名
// 例如: HeroMechanism -> hero_mechanism.go
func GetMechanismFileName(name string) string {
	snake := CamelToSnake(name)
	// 去掉末尾的 _mechanism，然后加上 _mechanism.go
	if strings.HasSuffix(snake, "_mechanism") {
		return snake + ".go"
	}
	return snake + "_mechanism.go"
}

// GetModuleFileName 生成 Module 文件名
// 例如: HeroModule -> hero_modules.go
func GetModuleFileName(name string) string {
	snake := CamelToSnake(name)
	// 去掉末尾的 _module，然后加上 _modules
	if strings.HasSuffix(snake, "_module") {
		return strings.TrimSuffix(snake, "_module") + "_modules.go"
	}
	return snake + "_modules.go"
}

// GetManagerFileName 生成 Manager 文件名
// 例如: HeroManager -> hero_managers.go
func GetManagerFileName(name string) string {
	snake := CamelToSnake(name)
	// 去掉末尾的 _manager，然后加上 _managers
	if strings.HasSuffix(snake, "_manager") {
		return strings.TrimSuffix(snake, "_manager") + "_managers.go"
	}
	return snake + "_managers.go"
}

// GetEntityFileName 生成 Entity 文件名
// 例如: PlayerEntity -> player_entity.go (注意：Entity 使用单数)
func GetEntityFileName(name string) string {
	snake := CamelToSnake(name)
	// 去掉末尾的 _entity，然后加上 _entity.go (单数)
	if strings.HasSuffix(snake, "_entity") {
		return snake + ".go"
	}
	return snake + "_entity.go"
}

// containsSpecialChars 检查字符串是否包含特殊字符
func containsSpecialChars(s string, specialChars []string) bool {
	for _, char := range specialChars {
		if strings.Contains(s, char) {
			return true
		}
	}
	return false
}

// BuildMongoUpdateCodeGenerator BuildMongoUpdate 代码生成器
type BuildMongoUpdateCodeGenerator struct {
	ObjectName  string     // 对象名称，如 "LevelUpMechanism"
	WrapperName string     // 包装器名称，如 "LevelUpMechanismWrapper"
	Receiver    string     // 接收器名称，如 "w" 或 "m"
	ObjectType  ObjectType // 对象类型：Mechanism/Module/Manager
}

// GenerateBuildMongoUpdateMethod 生成 BuildMongoUpdate 方法的完整代码
func (g *BuildMongoUpdateCodeGenerator) GenerateBuildMongoUpdateMethod(fields []*types.Field) string {
	var sb strings.Builder

	// 方法注释
	sb.WriteString("// BuildMongoUpdate 构建MongoDB更新操作\n")
	if g.ObjectType == ObjectTypeManager {
		sb.WriteString("// 注意：Manager 层级通常不直接构建 MongoDB 更新，而是由 Entity 层处理\n")
	}

	// 方法签名
	sb.WriteString(fmt.Sprintf("func (%s *%s) BuildMongoUpdate(builder *mgo_builder.MongoUpdateBuilder, path *mgo_builder.NestedPath) {\n",
		g.Receiver, g.WrapperName))
	sb.WriteString(fmt.Sprintf("\tif %s == nil {\n", g.Receiver))
	sb.WriteString("\t\treturn\n")
	sb.WriteString("\t}\n\n")

	for i := range fields {
		field := fields[i]
		dirtyBitName := fmt.Sprintf("%sDirty%sBit", g.ObjectName, field.Name)
		fieldNameSnake := CamelToSnake(field.Name)
		sb.WriteString(fmt.Sprintf("\tif %s.IsDirty(%s) {\n", g.Receiver, dirtyBitName))
		switch field.Type.Kind {
		case types.FieldKindMap:
			sb.WriteString(fmt.Sprintf("\t\tbuilder.SetNestedPath(path, \"%s\", %s.Get%s())\n",
				fieldNameSnake, g.Receiver, field.Name))
		case types.FieldKindXMap:
			valueType := field.GetValueType()
			if valueType == nil {
				panic(fmt.Sprintf("field %s value type is nil", field.Name))
			}
			switch {
			case valueType.GetKind().IsBaseType():
				// 值类型 XMap，使用 accessor.Clone()
				accessorName := strings.ToLower(field.Name[0:1]) + field.Name[1:] + "Accessor"
				sb.WriteString(fmt.Sprintf("\t\tbuilder.SetNestedPath(path, \"%s\", %s.%s.Clone())\n",
					fieldNameSnake, g.Receiver, accessorName))
			case valueType.GetKind().IsMessage():
				panic(fmt.Sprintf("field %s value type is message type, not supported", field.Name))
			case valueType.GetKind().IsMMEObject():
				// MME Object 类型 XMap，使用 DeepCopy()
				valueName := field.GetValueName()
				linkFieldName := strings.ToLower(field.Name[0:1]) + field.Name[1:] + "Link"
				sb.WriteString("\t\t//map 结构无法做增量更新，所以需要全量拷贝\n")
				sb.WriteString(fmt.Sprintf("\t\tcopy := make(map[%s]*%s, %s.%s.Len())\n",
					field.Type.KeyKind().String(), valueName, g.Receiver, linkFieldName))
				sb.WriteString(fmt.Sprintf("\t\t%s.%s.DeepCopy(&copy)\n", g.Receiver, linkFieldName))
				sb.WriteString(fmt.Sprintf("\t\tbuilder.SetNestedPath(path, \"%s\", copy)\n", fieldNameSnake))
			default:
				panic(fmt.Sprintf("field %s value type is not supported", field.Name))
			}
		}
		sb.WriteString("\t}\n")
	}

	sb.WriteString("}\n\n")
	return sb.String()
}
