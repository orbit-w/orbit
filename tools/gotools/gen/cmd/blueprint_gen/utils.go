package blueprint_gen

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"
	mmeobject "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types/mme_obj"
)

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

// FormatProtoContent 格式化 proto 文件内容
// 主要处理：
// 1. 确保所有字段使用4个空格缩进
// 2. 清理多余的空行
// 3. 确保格式一致性
// 注意：长行拆分已经在生成代码时处理，这里主要做最后的清理
func FormatProtoContent(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	var lastEmpty bool

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// 处理空行：保留单个空行，但去除连续的空行
		if trimmed == "" {
			if !lastEmpty {
				result = append(result, "")
				lastEmpty = true
			}
			continue
		}
		lastEmpty = false

		// 保持原有行的格式，因为生成代码时已经正确格式化了
		// 这里主要是确保缩进一致性
		result = append(result, line)
	}

	// 移除末尾的空行
	for len(result) > 0 && result[len(result)-1] == "" {
		result = result[:len(result)-1]
	}

	return strings.Join(result, "\n")
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

// GetEntityCollectionName 生成 Entity 集合名称
// 例如: PlayerEntity -> player_entities (复数形式)
func GetEntityCollectionName(name string) string {
	snake := CamelToSnake(name)
	// 转为复数形式：简单地在末尾加 s
	if strings.HasSuffix(snake, "_entity") {
		return snake + "s"
	}
	return snake + "s"
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
	ObjectName  string               // 对象名称，如 "LevelUpMechanism"
	WrapperName string               // 包装器名称，如 "LevelUpMechanismWrapper"
	Receiver    string               // 接收器名称，如 "w" 或 "m"
	ObjectType  mmeobject.ObjectType // 对象类型：Mechanism/Module/Manager/Entity
}

// GenerateBuildMongoUpdateMethod 生成 BuildMongoUpdate 方法的完整代码
func (g *BuildMongoUpdateCodeGenerator) GenerateBuildMongoUpdateMethod(fields []*types.Field) string {
	var sb strings.Builder

	// 生成方法签名
	g.GenerateBuildMongoUpdateMethodFuncSign(&sb)

	// 确定缩进级别
	indent := "\t"
	if g.ObjectType == mmeobject.ObjectTypeEntity {
		indent = "\t\t" // Entity 类型多一层缩进（因为有 HasAnyDirty 检查）
	}

	for i := range fields {
		field := fields[i]

		switch {
		case g.ObjectType.IsEntity():
			// Entity 类型只处理 Manager 类型的字段
			if field.IsMMEObjectType() {
				fieldType := field.GetMMEObjectType()
				if !fieldType.IsManager() {
					panic(fmt.Sprintf("field %s is not MMEObject type, it is %s", field.Name, fieldType.String()))
				}
			}
		}

		dirtyBitName := fmt.Sprintf("%sDirty%sBit", g.ObjectName, field.Name)
		fieldNameSnake := CamelToSnake(field.Name)
		sb.WriteString(fmt.Sprintf("%sif %s.IsDirty(%s) {\n", indent, g.Receiver, dirtyBitName))
		switch field.Type.Kind {
		case types.FieldKindMap:
			switch {
			case field.Type.ValueKind().IsBaseType():
				keyType := field.Type.KeyKind().String()
				valueType := field.Type.ValueKind().String()
				sb.WriteString(fmt.Sprintf("%s\tcopy := make(map[%s]%s, len(%s.data.%s))\n",
					indent, keyType, valueType, g.Receiver, field.Name))
				sb.WriteString(fmt.Sprintf("%s\tmaps.Copy(copy, %s.data.%s)\n",
					indent, g.Receiver, field.Name))
				sb.WriteString(fmt.Sprintf("%s\tbuilder.SetNestedPath(path, \"%s\", copy)\n",
					indent, fieldNameSnake))
			case field.Type.ValueKind().IsMMEObject():
				panic(fmt.Sprintf("field %s value type is MMEObject, not supported for Map", field.Name))
			default:
				panic(fmt.Sprintf("field %s value type is not supported for Map", field.Name))
			}
		case types.FieldKindXMap:
			valueType := field.GetValueType()
			if valueType == nil {
				panic(fmt.Sprintf("field %s value type is nil", field.Name))
			}
			switch {
			case valueType.GetKind().IsBaseType():
				// 值类型 XMap，使用 accessor.Clone()
				accessorName := strings.ToLower(field.Name[0:1]) + field.Name[1:] + "Accessor"
				sb.WriteString(fmt.Sprintf("%s\tbuilder.SetNestedPath(path, \"%s\", %s.%s.Clone())\n",
					indent, fieldNameSnake, g.Receiver, accessorName))
			case valueType.GetKind().IsMessage():
				panic(fmt.Sprintf("field %s value type is message type, not supported", field.Name))
			case valueType.GetKind().IsMMEObject():
				// MME Object 类型 XMap，使用 DeepCopy()
				valueName := field.GetValueName()
				linkFieldName := strings.ToLower(field.Name[0:1]) + field.Name[1:] + "Link"
				sb.WriteString(fmt.Sprintf("%s\t//map 结构无法做增量更新，所以需要全量拷贝\n", indent))
				sb.WriteString(fmt.Sprintf("%s\tcopy := make(map[%s]*%s, %s.%s.Len())\n",
					indent, field.Type.KeyKind().String(), valueName, g.Receiver, linkFieldName))
				sb.WriteString(fmt.Sprintf("%s\t%s.%s.DeepCopy(&copy)\n", indent, g.Receiver, linkFieldName))
				sb.WriteString(fmt.Sprintf("%s\tbuilder.SetNestedPath(path, \"%s\", copy)\n", indent, fieldNameSnake))
			default:
				panic(fmt.Sprintf("field %s value type is not supported", field.Name))
			}
		default:
			switch {
			case field.Type.GetKind().IsBaseType():
				// 基础类型字段处理
				methodName := strings.ToUpper(field.Name[0:1]) + field.Name[1:]
				// 所有字段统一使用 snake_case 命名，与 bson 标签保持一致
				sb.WriteString(fmt.Sprintf("%s\tbuilder.SetNestedPath(path, \"%s\", %s.Get%s())\n",
					indent, fieldNameSnake, g.Receiver, methodName))
			case field.Type.GetKind().IsMMEObject():
				// MMEObject 类型字段处理：调用嵌套 wrapper 的 BuildMongoUpdate 方法
				wrapperFieldName := field.Name + "Wrapper"
				sb.WriteString(fmt.Sprintf("%s\tif %s.%s != nil {\n",
					indent, g.Receiver, wrapperFieldName))
				sb.WriteString(fmt.Sprintf("%s\t\t%s.%s.BuildMongoUpdate(builder, path.Field(\"%s\"))\n",
					indent, g.Receiver, wrapperFieldName, fieldNameSnake))
				sb.WriteString(fmt.Sprintf("%s\t}\n", indent))
			default:
				panic(fmt.Sprintf("field %s value type is not supported", field.Name))
			}
		}
		sb.WriteString(fmt.Sprintf("%s}\n", indent))
	}

	// Entity 类型需要关闭 HasAnyDirty 的大括号
	if g.ObjectType == mmeobject.ObjectTypeEntity {
		sb.WriteString("\t}\n")
	}

	sb.WriteString("}\n\n")
	return sb.String()
}

// GenerateBuildMongoUpdateMethodFuncSign 生成 BuildMongoUpdate 方法的函数签名
func (g *BuildMongoUpdateCodeGenerator) GenerateBuildMongoUpdateMethodFuncSign(sb *strings.Builder) {
	// 方法注释
	sb.WriteString("// BuildMongoUpdate 构建MongoDB更新操作\n")
	if g.ObjectType == mmeobject.ObjectTypeManager {
		sb.WriteString("// 注意：Manager 层级通常不直接构建 MongoDB 更新，而是由 Entity 层处理\n")
	}

	// Entity 类型的方法签名不同：不接受 path 参数
	if g.ObjectType == mmeobject.ObjectTypeEntity {
		sb.WriteString(fmt.Sprintf("func (%s *%s) BuildMongoUpdate(builder *mgo_builder.MongoUpdateBuilder) {\n",
			g.Receiver, g.WrapperName))
		sb.WriteString(fmt.Sprintf("\tif %s == nil {\n", g.Receiver))
		sb.WriteString("\t\treturn\n")
		sb.WriteString("\t}\n\n")

		// Entity 类型需要先检查是否有脏数据
		sb.WriteString(fmt.Sprintf("\tif %s.HasAnyDirty() {\n", g.Receiver))
		// Entity 类型在内部创建 path
		sb.WriteString(fmt.Sprintf("\t\tpath := mgo_builder.NewNestedPathWithField(%s.Collection())\n", g.Receiver))
	} else {
		// 其他类型（Mechanism, Module, Manager）的方法签名：接受 path 参数
		sb.WriteString(fmt.Sprintf("func (%s *%s) BuildMongoUpdate(builder *mgo_builder.MongoUpdateBuilder, path *mgo_builder.NestedPath) {\n",
			g.Receiver, g.WrapperName))
		sb.WriteString(fmt.Sprintf("\tif %s == nil {\n", g.Receiver))
		sb.WriteString("\t\treturn\n")
		sb.WriteString("\t}\n\n")
	}

}

// ToIncrementalProtoCodeGenerator ToIncrementalProto 代码生成器
type ToIncrementalProtoCodeGenerator struct {
	ObjectName  string               // 对象名称，如 "LevelUpMechanism"
	WrapperName string               // 包装器名称，如 "LevelUpMechanismWrapper"
	Receiver    string               // 接收器名称，如 "w" 或 "m"
	ObjectType  mmeobject.ObjectType // 对象类型：Mechanism/Module/Manager
}

// GenerateToIncrementalProtoMethod 生成 ToIncrementalProto 方法的完整代码
func (g *ToIncrementalProtoCodeGenerator) GenerateToIncrementalProtoMethod(fields []*types.Field) string {
	var sb strings.Builder

	// 根据对象类型确定方法名
	methodName := "ToIncrementalProtoWithContext"
	// 方法注释
	sb.WriteString(fmt.Sprintf("// %s 根据脏标记位构建增量数据的 protoMessage\n", methodName))
	sb.WriteString("// 只返回标记为脏的字段数据，用于增量同步\n")

	// 方法签名
	sb.WriteString(fmt.Sprintf("func (%s *%s) %s(ctx mmemodel.SyncContext) proto.Message {\n",
		g.Receiver, g.WrapperName, methodName))
	sb.WriteString(fmt.Sprintf("\tif %s == nil {\n", g.Receiver))
	sb.WriteString("\t\treturn nil\n")
	sb.WriteString("\t}\n\n")
	sb.WriteString("\t// 如果没有脏标记，返回 nil\n")
	sb.WriteString(fmt.Sprintf("\tif !%s.HasAnyDirty() {\n", g.Receiver))
	sb.WriteString("\t\treturn nil\n")
	sb.WriteString("\t}\n\n")
	sb.WriteString(fmt.Sprintf("\tincremental := &mme.%s{}\n\n", g.ObjectName))

	for i := range fields {
		field := fields[i]
		fieldIndexName := fmt.Sprintf("%sFieldIndex%s", g.ObjectName, field.Name)
		dirtyBitName := fmt.Sprintf("%sDirty%sBit", g.ObjectName, field.Name)

		switch field.Type.Kind {
		case types.FieldKindXMap:
			// XMap 类型字段处理
			valueType := field.GetValueType()
			if valueType == nil {
				panic(fmt.Sprintf("field %s value type is nil", field.Name))
			}

			switch {
			case valueType.GetKind().IsMMEObject():
				// MME Object 类型 XMap（仅 Manager 支持）
				if g.ObjectType != mmeobject.ObjectTypeManager {
					panic(fmt.Sprintf("field %s is XMap with MME Object value type, only supported in Manager", field.Name))
				}

				linkFieldName := strings.ToLower(field.Name[0:1]) + field.Name[1:] + "Link"
				keyType := field.Type.KeyKind().String()
				valueName := field.GetValueName()
				protoValueType := fmt.Sprintf("*mme.%s", valueName)

				sb.WriteString(fmt.Sprintf("\tif mmemodel.FieldCanBeIncrementalSynced(%s, %s, %s, ctx) {\n",
					g.Receiver, dirtyBitName, fieldIndexName))
				sb.WriteString(fmt.Sprintf("\t\tincremental.%s_XXXChangeList = make([]*mme.%s_%s_XXXMapChangeRecord, 0)\n",
					field.Name, g.ObjectName, field.Name))
				sb.WriteString(fmt.Sprintf("\t\t%s.%s.RangeOperations(func(key %s, operation xmap.MapOperation[%s]) bool {\n",
					g.Receiver, linkFieldName, keyType, keyType))
				sb.WriteString("\t\t\tswitch operation.Type {\n")
				sb.WriteString("\t\t\tcase xmap.SetOperation:\n")
				sb.WriteString(fmt.Sprintf("\t\t\t\t%s, _ := %s.%s.Get(key)\n",
					strings.ToLower(valueName[0:1])+valueName[1:], g.Receiver, linkFieldName))
				sb.WriteString(fmt.Sprintf("\t\t\t\tpb := %s.ToIncrementalProtoWithContext(ctx)\n",
					strings.ToLower(valueName[0:1])+valueName[1:]))
				sb.WriteString(fmt.Sprintf("\t\t\t\tv, ok := pb.(%s)\n", protoValueType))
				sb.WriteString("\t\t\t\tif ok {\n")
				sb.WriteString(fmt.Sprintf("\t\t\t\t\tincremental.%s_XXXChangeList = append(incremental.%s_XXXChangeList, &mme.%s_%s_XXXMapChangeRecord{\n",
					field.Name, field.Name, g.ObjectName, field.Name))
				sb.WriteString("\t\t\t\t\t\tKey:   key,\n")
				sb.WriteString("\t\t\t\t\t\tValue: v,\n")
				sb.WriteString("\t\t\t\t\t})\n")
				sb.WriteString("\t\t\t\t}\n")
				sb.WriteString("\t\t\t\treturn true\n")
				sb.WriteString("\t\t\tcase xmap.DeleteOperation:\n")
				sb.WriteString(fmt.Sprintf("\t\t\t\tincremental.%s_XXXChangeList = append(incremental.%s_XXXChangeList, &mme.%s_%s_XXXMapChangeRecord{\n",
					field.Name, field.Name, g.ObjectName, field.Name))
				sb.WriteString("\t\t\t\t\tKey:      key,\n")
				sb.WriteString("\t\t\t\t\tIsDelete: true,\n")
				sb.WriteString("\t\t\t\t})\n")
				sb.WriteString("\t\t\t\treturn true\n")
				sb.WriteString("\t\t\t}\n")
				sb.WriteString("\t\t\treturn true\n")
				sb.WriteString("\t\t})\n\n")
				sb.WriteString("\t}\n")
			case valueType.GetKind().IsBaseType():
				// 基础类型值 XMap，使用 Clone()
				accessorName := strings.ToLower(field.Name[0:1]) + field.Name[1:] + "Accessor"
				sb.WriteString(fmt.Sprintf("\tif mmemodel.FieldCanBeIncrementalSynced(%s, %s, %s, ctx) {\n",
					g.Receiver, dirtyBitName, fieldIndexName))
				sb.WriteString(fmt.Sprintf("\t\tincremental.%s = %s.%s.Clone()\n",
					field.Name, g.Receiver, accessorName))
				sb.WriteString("\t}\n")
			default:
				panic(fmt.Sprintf("field %s value type is not supported for XMap", field.Name))
			}
		case types.FieldKindMap:
			// Map 类型字段处理
			valueType := field.GetValueType()
			if valueType == nil {
				panic(fmt.Sprintf("field %s value type is nil", field.Name))
			}
			switch {
			case valueType.GetKind().IsMMEObject():
				panic(fmt.Sprintf("field %s value type is MMEObject, not supported for Map", field.Name))
			case valueType.GetKind().IsBaseType():
				// 普通 Map 类型字段处理
				methodName := strings.ToUpper(field.Name[0:1]) + field.Name[1:]
				keyType := field.Type.KeyKind().String()
				valueType := field.Type.ValueKind().String()

				sb.WriteString(fmt.Sprintf("\tif mmemodel.FieldCanBeIncrementalSynced(%s, %s, %s, ctx) {\n",
					g.Receiver, dirtyBitName, fieldIndexName))
				sb.WriteString(fmt.Sprintf("\t\tm := %s.Get%s()\n", g.Receiver, methodName))
				sb.WriteString("\t\tif m != nil {\n")
				sb.WriteString(fmt.Sprintf("\t\t\tincremental.%s = make(map[%s]%s, len(m))\n",
					field.Name, keyType, valueType))
				sb.WriteString(fmt.Sprintf("\t\t\tmaps.Copy(incremental.%s, m)\n", field.Name))
				sb.WriteString("\t\t}\n")
				sb.WriteString("\t}\n")
			default:
				panic(fmt.Sprintf("field %s value type is not supported for Map", field.Name))
			}
		case types.FieldKindMMEObject:
			// MME Object 类型字段处理
			wrapperFieldName := field.Name + "Wrapper"
			typeName := field.GetTypeName()
			protoTypeName := fmt.Sprintf("*mme.%s", typeName)

			sb.WriteString(fmt.Sprintf("\tif mmemodel.FieldCanBeIncrementalSynced(%s, %s, %s, ctx) {\n",
				g.Receiver, dirtyBitName, fieldIndexName))
			sb.WriteString(fmt.Sprintf("\t\tif %s.%s != nil {\n", g.Receiver, wrapperFieldName))
			sb.WriteString(fmt.Sprintf("\t\t\tpb := %s.%s.ToIncrementalProtoWithContext(ctx)\n",
				g.Receiver, wrapperFieldName))
			sb.WriteString("\t\t\tif pb != nil {\n")
			sb.WriteString(fmt.Sprintf("\t\t\t\tv, ok := pb.(%s)\n", protoTypeName))
			sb.WriteString("\t\t\t\tif ok {\n")
			sb.WriteString(fmt.Sprintf("\t\t\t\t\tincremental.%s = v\n", field.Name))
			sb.WriteString("\t\t\t\t}\n")
			sb.WriteString("\t\t\t}\n")
			sb.WriteString("\t\t}\n")
			sb.WriteString("\t}\n")
		case types.FieldKindMessage:
			panic(fmt.Sprintf("field %s value type is message type, not supported", field.Name))
		default:
			// 基础类型字段处理
			if field.Type.GetKind().IsBaseType() {
				methodName := strings.ToUpper(field.Name[0:1]) + field.Name[1:]
				sb.WriteString(fmt.Sprintf("\tif mmemodel.FieldCanBeIncrementalSynced(%s, %s, %s, ctx) {\n",
					g.Receiver, dirtyBitName, fieldIndexName))
				sb.WriteString(fmt.Sprintf("\t\tv := %s.Get%s()\n", g.Receiver, methodName))
				sb.WriteString(fmt.Sprintf("\t\tincremental.%s = &v\n", field.Name))
				sb.WriteString("\t}\n")
			} else {
				panic(fmt.Sprintf("field %s value type is not supported", field.Name))
			}
		}
	}

	sb.WriteString("\treturn incremental\n")
	sb.WriteString("}\n\n")
	return sb.String()
}

// WrapperMethodCodeGenerator Wrapper 方法代码生成器
type WrapperMethodCodeGenerator struct {
	ObjectName  string               // 对象名称，如 "HeroModule"
	WrapperName string               // 包装器名称，如 "HeroModuleWrapper"
	Receiver    string               // 接收器名称，如 "w" 或 "m"
	ProtoPkg    string               // Proto 包名，如 "mme"
	ObjectType  mmeobject.ObjectType // 对象类型：Mechanism/Module/Manager/Entity
}

// GenerateDeepCopyMethod 生成 DeepCopy 方法
func (g *WrapperMethodCodeGenerator) GenerateDeepCopyMethod() string {
	var sb strings.Builder
	sb.WriteString("// 包装器-深拷贝\n")
	sb.WriteString(fmt.Sprintf("func (%s *%s) DeepCopy() *%s {\n", g.Receiver, g.WrapperName, g.ObjectName))
	sb.WriteString(fmt.Sprintf("\tif %s == nil {\n", g.Receiver))
	sb.WriteString("\t\treturn nil\n")
	sb.WriteString("\t}\n")
	sb.WriteString(fmt.Sprintf("\tcopy := &%s{}\n", g.ObjectName))
	sb.WriteString(fmt.Sprintf("\t%s.DeepCopyTo(copy)\n", g.Receiver))
	sb.WriteString("\treturn copy\n")
	sb.WriteString("}\n\n")
	return sb.String()
}

// GenerateDeepCopyToMethod 生成 DeepCopyTo 方法
func (g *WrapperMethodCodeGenerator) GenerateDeepCopyToMethod(fields []*types.Field) string {
	var sb strings.Builder
	sb.WriteString("// 包装器-深拷贝\n")
	sb.WriteString(fmt.Sprintf("func (%s *%s) DeepCopyTo(copy *%s) {\n", g.Receiver, g.WrapperName, g.ObjectName))
	sb.WriteString(fmt.Sprintf("\tif %s == nil || %s.data == nil || copy == nil {\n", g.Receiver, g.Receiver))
	sb.WriteString("\t\treturn\n")
	sb.WriteString("\t}\n\n")

	// 检查是否有基础类型字段，如果有则先拷贝整个 data（一次性拷贝所有基础类型字段）
	hasBaseTypeFields := false
	for _, field := range fields {
		if field.Type.GetKind().IsBaseType() {
			hasBaseTypeFields = true
			break
		}
	}
	if hasBaseTypeFields {
		sb.WriteString("\t// 拷贝基础类型字段\n")
		sb.WriteString(fmt.Sprintf("\t*copy = *%s.data\n\n", g.Receiver))
	}

	// 根据字段类型处理每个字段
	for _, field := range fields {
		switch {
		case field.Type.GetKind().IsBaseType():
			// 基础类型字段已经在上面通过 *copy = *w.data 拷贝了，跳过
			continue
		case field.Type.IsMMEObjectType():
			// MMEObject 类型字段：递归调用 DeepCopyTo
			typeName := field.GetTypeName()
			wrapperFieldName := field.Name + "Wrapper"
			sb.WriteString(fmt.Sprintf("\tif %s.data.%s != nil {\n", g.Receiver, field.Name))
			sb.WriteString(fmt.Sprintf("\t\tif copy.%s == nil {\n", field.Name))
			sb.WriteString(fmt.Sprintf("\t\t\tcopy.%s = &%s{}\n", field.Name, typeName))
			sb.WriteString("\t\t}\n")
			sb.WriteString(fmt.Sprintf("\t\t%s.%s.DeepCopyTo(copy.%s)\n", g.Receiver, wrapperFieldName, field.Name))
			sb.WriteString("\t}\n")
		case field.Type.IsMapField():
			valueType := field.GetValueType()
			switch {
			case valueType.IsMMEObjectType():
				panic(fmt.Sprintf("field %s value type is MMEObject, not supported for Map", field.Name))
			case valueType.IsFieldBaseType():
				// Map 类型字段：使用 maps.Copy()
				keyType := field.Type.KeyKind().String()
				valueName := field.GetValueName()
				sb.WriteString("\t// 初始化目标 map\n")
				sb.WriteString(fmt.Sprintf("\tif copy.%s == nil {\n", field.Name))
				sb.WriteString(fmt.Sprintf("\t\tcopy.%s = make(map[%s]%s, len(%s.data.%s))\n",
					field.Name, keyType, valueName, g.Receiver, field.Name))
				sb.WriteString("\t}\n")
				sb.WriteString(fmt.Sprintf("\t// 拷贝所有 %s\n", valueName))
				sb.WriteString(fmt.Sprintf("\tif %s.data.%s != nil {\n", g.Receiver, field.Name))
				sb.WriteString(fmt.Sprintf("\t\tmaps.Copy(copy.%s, %s.data.%s)\n", field.Name, g.Receiver, field.Name))
				sb.WriteString("\t}\n")
			default:
				panic(fmt.Sprintf("field %s value type is not supported for Map", field.Name))
			}
		case field.Type.IsXMapField():
			// XMap 类型字段：根据 ValueType 决定使用 Accessor.Copy() 还是 Link.DeepCopy()
			valueType := field.GetValueType()
			switch {
			case valueType.IsMMEObjectType():
				// ValueType 是 MMEObject，使用 Link.DeepCopy()
				linkFieldName := strings.ToLower(field.Name[0:1]) + field.Name[1:] + "Link"
				keyType := field.Type.KeyKind().String()
				valueName := field.GetValueName()
				sb.WriteString("\t// 初始化目标 map\n")
				sb.WriteString(fmt.Sprintf("\tif copy.%s == nil {\n", field.Name))
				sb.WriteString(fmt.Sprintf("\t\tcopy.%s = make(map[%s]*%s, len(%s.data.%s))\n",
					field.Name, keyType, valueName, g.Receiver, field.Name))
				sb.WriteString("\t}\n")
				sb.WriteString(fmt.Sprintf("\t// 拷贝所有 %s\n", valueName))
				sb.WriteString(fmt.Sprintf("\t%s.%s.DeepCopy(&copy.%s)\n", g.Receiver, linkFieldName, field.Name))

			case valueType.IsFieldBaseType():
				// ValueType 是基础类型，使用 Accessor.Copy()
				accessorName := strings.ToLower(field.Name[0:1]) + field.Name[1:] + "Accessor"
				sb.WriteString(fmt.Sprintf("\t// 拷贝 %sAccessor\n", field.Name))
				sb.WriteString(fmt.Sprintf("\tif %s.%s != nil {\n", g.Receiver, accessorName))
				sb.WriteString(fmt.Sprintf("\t\tif copy.%s == nil {\n", field.Name))
				sb.WriteString(fmt.Sprintf("\t\t\tcopy.%s = make(map[%s]%s, %s.%s.Len())\n",
					field.Name, field.Type.KeyKind().String(), field.Type.ValueKind().String(), g.Receiver, accessorName))
				sb.WriteString("\t\t}\n")
				sb.WriteString(fmt.Sprintf("\t%s.%s.Copy(copy.%s)\n", g.Receiver, accessorName, field.Name))
				sb.WriteString("\t}\n")

			default:
				panic(fmt.Sprintf("field %s value type is not supported for XMap", field.Name))
			}

		}
	}

	sb.WriteString("}\n\n")
	return sb.String()
}

// GenerateToProtoMethod 生成 ToProto 方法
func (g *WrapperMethodCodeGenerator) GenerateToProtoMethod() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("// ToProto 将 %s 数据转换为完整的 protobuf 结构体\n", g.ObjectName))
	if g.ObjectType == mmeobject.ObjectTypeEntity {
		sb.WriteString(fmt.Sprintf("func (%s *%s) ToProto() proto.Message {\n", g.Receiver, g.WrapperName))
	} else {
		sb.WriteString(fmt.Sprintf("func (%s *%s) ToProto() *%s.%s {\n", g.Receiver, g.WrapperName, g.ProtoPkg, g.ObjectName))
	}
	sb.WriteString(fmt.Sprintf("\tif %s == nil || %s.data == nil {\n", g.Receiver, g.Receiver))
	sb.WriteString("\t\treturn nil\n")
	sb.WriteString("\t}\n\n")
	sb.WriteString(fmt.Sprintf("\treturn %s.data.ToProto()\n", g.Receiver))
	sb.WriteString("}\n\n")
	return sb.String()
}

// GenerateFromProtoMethod 生成 FromProto 方法
func (g *WrapperMethodCodeGenerator) GenerateFromProtoMethod(fields []*types.Field) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("// FromProto 从 protobuf 结构体加载数据到 %s\n", g.ObjectName))
	if g.ObjectType == mmeobject.ObjectTypeEntity {
		sb.WriteString(fmt.Sprintf("func (%s *%s) FromProto(msg proto.Message) {\n", g.Receiver, g.WrapperName))
		sb.WriteString(fmt.Sprintf("\tif %s == nil || %s.data == nil || msg == nil {\n", g.Receiver, g.Receiver))
		sb.WriteString("\t\treturn\n")
		sb.WriteString("\t}\n\n")
		sb.WriteString(fmt.Sprintf("\tpb, ok := msg.(*%s.%s)\n", g.ProtoPkg, g.ObjectName))
		sb.WriteString("\tif !ok {\n")
		sb.WriteString("\t\treturn\n")
		sb.WriteString("\t}\n\n")
	} else {
		sb.WriteString(fmt.Sprintf("func (%s *%s) FromProto(pb *%s.%s) {\n", g.Receiver, g.WrapperName, g.ProtoPkg, g.ObjectName))
		sb.WriteString(fmt.Sprintf("\tif %s == nil || %s.data == nil || pb == nil {\n", g.Receiver, g.Receiver))
		sb.WriteString("\t\treturn\n")
		sb.WriteString("\t}\n\n")
	}
	// 根据字段类型处理每个字段
	for _, field := range fields {
		if field.Type.GetKind().IsBaseType() {
			// 基础类型字段
			methodName := strings.ToUpper(field.Name[0:1]) + field.Name[1:]
			sb.WriteString(fmt.Sprintf("\tif pb.%s != nil {\n", field.Name))
			sb.WriteString(fmt.Sprintf("\t\t%s.Set%s(*pb.%s)\n", g.Receiver, methodName, field.Name))
			sb.WriteString("\t}\n")
		} else if field.Type.IsMMEObjectType() {
			// MMEObject 类型字段：递归调用 FromProto
			typeName := field.GetTypeName()
			wrapperFieldName := field.Name + "Wrapper"
			sb.WriteString(fmt.Sprintf("\tif pb.%s != nil {\n", field.Name))
			sb.WriteString(fmt.Sprintf("\t\tif %s.data.%s == nil {\n", g.Receiver, field.Name))
			sb.WriteString(fmt.Sprintf("\t\t\t%s.data.%s = New%s()\n", g.Receiver, field.Name, typeName))
			sb.WriteString("\t\t}\n")
			sb.WriteString(fmt.Sprintf("\t\t%s.%s.FromProto(pb.%s)\n", g.Receiver, wrapperFieldName, field.Name))
			sb.WriteString("\t}\n")
		} else if field.Type.IsXMapField() {
			// XMap 类型字段：根据 ValueType 决定处理方式
			valueType := field.GetValueType()
			if valueType != nil && valueType.IsMMEObjectType() {
				// ValueType 是 MMEObject，使用 Link.SetWithoutTrack() 和 FromProto()
				linkFieldName := strings.ToLower(field.Name[0:1]) + field.Name[1:] + "Link"
				valueName := field.GetValueName()
				sb.WriteString(fmt.Sprintf("\tfor key := range pb.%s {\n", field.Name))
				sb.WriteString(fmt.Sprintf("\t\tpbValue := pb.%s[key]\n", field.Name))
				sb.WriteString(fmt.Sprintf("\t\tv := New%s()\n", valueName))
				sb.WriteString(fmt.Sprintf("\t\twrapper := %s.%s.SetWithoutTrack(key, v)\n", g.Receiver, linkFieldName))
				sb.WriteString("\t\twrapper.FromProto(pbValue)\n")
				sb.WriteString("\t}\n")
			} else {
				// ValueType 是基础类型，使用 Accessor.Reset()
				accessorName := strings.ToLower(field.Name[0:1]) + field.Name[1:] + "Accessor"
				sb.WriteString(fmt.Sprintf("\tif pb.%s != nil {\n", field.Name))
				sb.WriteString("\t\t// 清空现有的数据\n")
				sb.WriteString(fmt.Sprintf("\t\ttemp := make(map[%s]%s, len(pb.%s))\n",
					field.Type.KeyKind().String(), field.Type.ValueKind().String(), field.Name))
				sb.WriteString(fmt.Sprintf("\t\tmaps.Copy(temp, pb.%s)\n", field.Name))
				sb.WriteString("\t\t// 设置新的数据，并清空所有变化操作记录\n")
				sb.WriteString(fmt.Sprintf("\t\t%s.%s.Reset(&temp)\n", g.Receiver, accessorName))
				sb.WriteString("\t}\n")
			}
		} else if field.Type.IsMapField() {
			// Map 类型字段：根据 ValueType 决定处理方式
			valueType := field.GetValueType()
			switch {
			case valueType.IsMMEObjectType():
				panic(fmt.Sprintf("field %s value type is MMEObject, not supported for Map", field.Name))
			case valueType.IsFieldBaseType():
				// ValueType 是基础类型，使用 maps.Copy()
				sb.WriteString(fmt.Sprintf("\tif pb.%s != nil {\n", field.Name))
				sb.WriteString(fmt.Sprintf("\t\tmaps.Copy(%s.data.%s, pb.%s)\n", g.Receiver, field.Name, field.Name))
				sb.WriteString("\t}\n")
			default:
				panic(fmt.Sprintf("field %s value type is not supported for Map", field.Name))
			}
		}
	}

	sb.WriteString("}\n\n")
	return sb.String()
}

// generateEnumProto 生成 Enum 的 Proto 定义
func generateEnumProto(enum *Enum) string {
	sb := strings.Builder{}

	// 添加注释
	if enum.Comment != "" {
		sb.WriteString(fmt.Sprintf("// %s\n", enum.Comment))
	}

	// 生成 enum 定义
	sb.WriteString(fmt.Sprintf("enum %s {\n", enum.Name))

	// 生成枚举值定义
	for _, value := range enum.Values {
		// 添加注释
		if value.Comment != "" {
			sb.WriteString(fmt.Sprintf("    // %s\n", value.Comment))
		}
		// 生成枚举值定义
		sb.WriteString(fmt.Sprintf("    %s = %d;\n", value.Name, value.Number))
	}

	sb.WriteString("}\n\n")

	return sb.String()
}

func GenEntityTypeEnumName(entityName string) string {
	return fmt.Sprintf("%sType", entityName)
}

// GenerateBsonTag 生成 bson tag
// 将字段名转换为小写加下划线的格式
// 例如: ConfId -> bson:"conf_id"
func GenerateBsonTag(fieldName string) string {
	snakeName := CamelToSnake(fieldName)
	return fmt.Sprintf("`bson:\"%s\"`", snakeName)
}
