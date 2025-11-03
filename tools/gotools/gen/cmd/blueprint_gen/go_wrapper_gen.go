package blueprint_gen

import (
	"fmt"
	"os"
	"strings"

	types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"
)

// GoWrapperGenerator Go Wrapper 生成器
type GoWrapperGenerator struct {
	data *BlueprintData
}

// NewGoWrapperGenerator 创建新的 Go Wrapper 生成器
func NewGoWrapperGenerator(data *BlueprintData) *GoWrapperGenerator {
	return &GoWrapperGenerator{data: data}
}

// Generate 生成所有 Wrapper 文件
func (g *GoWrapperGenerator) Generate(outputDir string) error {
	// 生成 Mechanism Wrapper
	if err := g.generateMechanismWrappers(outputDir); err != nil {
		return fmt.Errorf("failed to generate mechanism wrappers: %w", err)
	}

	// TODO: 生成 Module、Manager、Entity Wrapper
	return nil
}

// generateMechanismWrappers 生成 Mechanism Wrapper
func (g *GoWrapperGenerator) generateMechanismWrappers(outputDir string) error {
	packageName := "mme"

	for _, mech := range g.data.Mechanisms {
		sb := strings.Builder{}

		// Wrapper 代码（不包含 package 和 import，因为这些会从已有文件中获取）

		// 生成 Wrapper 结构体
		wrapperName := mech.Name + "Wrapper"
		sb.WriteString(fmt.Sprintf("type %s struct {\n", wrapperName))
		sb.WriteString(fmt.Sprintf("\tdata *%s\n", mech.Name))
		sb.WriteString("\tdirtyflag.IDirtyFlag\n")
		sb.WriteString("\tfieldMetas *fieldmeta.FieldMetas\n\n")

		// 生成 map 访问器字段
		for _, field := range mech.Fields {
			if field.Type.Kind == types.FieldKindXMap || field.Type.Kind == types.FieldKindMap {
				// 判断 Value 类型
				if g.isMMEObjectType(field.Type.ValueType) {
					// MME Object 类型，使用 XMapWrapper
					sb.WriteString("\t// Value 为 MME Object 类型，使用 XMapWrapper\n")
					// TODO: 生成 XMapWrapper 字段
				} else {
					// 值类型，使用 MapAccessor
					keyType := ToGoBaseTypeFromFieldType(field.Type.KeyType)
					valueType := ToGoBaseTypeFromFieldType(field.Type.ValueType)
					accessorFieldName := strings.ToLower(field.Name[0:1]) + field.Name[1:] + "Accessor"
					sb.WriteString("\t// Value 为值类型，使用 xmap.MapAccessor 进行包装\n")
					sb.WriteString(fmt.Sprintf("\t%s *xmap.MapAccessor[%s, %s]\n",
						accessorFieldName, keyType, valueType))
				}
			}
		}

		sb.WriteString("}\n\n")

		// 生成 New 构造函数
		sb.WriteString(fmt.Sprintf("func New%s(data *%s) *%s {\n", wrapperName, mech.Name, wrapperName))
		sb.WriteString("\tif data == nil {\n")
		sb.WriteString("\t\tpanic(\"data is nil\")\n")
		sb.WriteString("\t}\n")
		sb.WriteString(fmt.Sprintf("\tw := &%s{\n", wrapperName))
		sb.WriteString("\t\tdata:       data,\n")
		sb.WriteString("\t\tIDirtyFlag: dirtyflag.NewDirtyFlag(),\n")
		sb.WriteString("\t\tfieldMetas: fieldmeta.NewFieldMetas(),\n")
		sb.WriteString("\t}\n\n")

		// 初始化 map 访问器
		for _, field := range mech.Fields {
			if (field.Type.Kind == types.FieldKindXMap || field.Type.Kind == types.FieldKindMap) && !g.isMMEObjectType(field.Type.ValueType) {
				accessorName := strings.ToLower(field.Name[0:1]) + field.Name[1:] + "Accessor"
				dirtyBitName := fmt.Sprintf("%sDirty%sBit", mech.Name, field.Name)
				sb.WriteString(fmt.Sprintf("\tw.%s = xmap.NewMapAccessorWithMarker(&w.data.%s, w, %s)\n",
					accessorName, field.Name, dirtyBitName))
			}
		}

		sb.WriteString("\treturn w\n")
		sb.WriteString("}\n\n")

		// 生成 InitFieldContext 方法
		sb.WriteString(fmt.Sprintf("func (w *%s) InitFieldContext() {\n", wrapperName))
		for _, field := range mech.Fields {
			fieldIndexName := fmt.Sprintf("%sFieldIndex%s", mech.Name, field.Name)
			// 根据 access 选项设置字段类型
			if field.Options.Access == "all" || field.Options.Access == "s" || field.Options.Access == "" {
				sb.WriteString(fmt.Sprintf("\tw.fieldMetas.SetFieldType(%s, fieldmeta.FieldTypeSync)\n",
					fieldIndexName))
			}
		}
		sb.WriteString("}\n\n")

		// 生成 Name 方法
		sb.WriteString(fmt.Sprintf("func (w *%s) Name() string {\n", wrapperName))
		sb.WriteString(fmt.Sprintf("\treturn \"%s\"\n", mech.Name))
		sb.WriteString("}\n\n")

		// 生成 MatchesAll 方法（实现 IFieldMetaContext 接口）
		sb.WriteString("// MatchesAll 判断字段是否匹配所有类型标记\n")
		sb.WriteString(fmt.Sprintf("func (w *%s) MatchesAll(fieldID uint8, fieldTypes ...fieldmeta.FieldType) bool {\n", wrapperName))
		sb.WriteString("\treturn w.fieldMetas.MatchesAll(fieldID, fieldTypes...)\n")
		sb.WriteString("}\n\n")

		// 生成 Getter 和 Setter 方法
		for _, field := range mech.Fields {
			if field.Type.Kind == types.FieldKindXMap || field.Type.Kind == types.FieldKindMap {
				// Map 字段生成访问器方法
				accessorName := strings.ToLower(field.Name[0:1]) + field.Name[1:]
				methodName := strings.ToUpper(field.Name[0:1]) + field.Name[1:]
				if !g.isMMEObjectType(field.Type.ValueType) {
					keyType := ToGoBaseTypeFromFieldType(field.Type.KeyType)
					valueType := ToGoBaseTypeFromFieldType(field.Type.ValueType)
					sb.WriteString(fmt.Sprintf("// 包装器-获取%s访问器\n", field.Name))
					sb.WriteString(fmt.Sprintf("func (w *%s) Get%sAccessor() *xmap.MapAccessor[%s, %s] {\n",
						wrapperName, methodName, keyType, valueType))
					sb.WriteString(fmt.Sprintf("\treturn w.%sAccessor\n", accessorName))
					sb.WriteString("}\n\n")
				}
			} else {
				// 普通字段生成 Getter 和 Setter
				goType := ToGoTypeFromTypesFieldType(&field.Type, packageName)
				methodName := strings.ToUpper(field.Name[0:1]) + field.Name[1:]
				dirtyBitName := fmt.Sprintf("%sDirty%sBit", mech.Name, field.Name)

				// Getter - 基础类型返回值类型，不是指针
				sb.WriteString(fmt.Sprintf("func (w *%s) Get%s() %s {\n", wrapperName, methodName, goType))
				sb.WriteString(fmt.Sprintf("\treturn w.data.%s\n", field.Name))
				sb.WriteString("}\n\n")

				// Setter - 基础类型接受值类型
				sb.WriteString(fmt.Sprintf("func (w *%s) Set%s(v %s) {\n", wrapperName, methodName, goType))
				sb.WriteString(fmt.Sprintf("\tw.data.%s = v\n", field.Name))
				sb.WriteString(fmt.Sprintf("\tw.MarkDirty(%s)\n", dirtyBitName))
				sb.WriteString("}\n\n")
			}
		}

		// 生成 ClearAllDirty 方法
		sb.WriteString(fmt.Sprintf("func (w *%s) ClearAllDirty() {\n", wrapperName))
		sb.WriteString("\tw.IDirtyFlag.ClearAllDirty()\n")
		for _, field := range mech.Fields {
			if (field.Type.Kind == types.FieldKindXMap || field.Type.Kind == types.FieldKindMap) && !g.isMMEObjectType(field.Type.ValueType) {
				accessorName := strings.ToLower(field.Name[0:1]) + field.Name[1:] + "Accessor"
				sb.WriteString(fmt.Sprintf("\tw.%s.ResetOperations()\n", accessorName))
			}
		}
		sb.WriteString("}\n\n")

		// 生成 BuildMongoUpdate 方法
		sb.WriteString(fmt.Sprintf("func (w *%s) BuildMongoUpdate(builder *mgo_builder.MongoUpdateBuilder, path *mgo_builder.NestedPath) {\n", wrapperName))
		sb.WriteString("\tif w == nil {\n")
		sb.WriteString("\t\treturn\n")
		sb.WriteString("\t}\n\n")

		for _, field := range mech.Fields {
			dirtyBitName := fmt.Sprintf("%sDirty%sBit", mech.Name, field.Name)
			fieldNameSnake := CamelToSnake(field.Name)

			if field.Type.Kind == types.FieldKindXMap || field.Type.Kind == types.FieldKindMap {
				sb.WriteString(fmt.Sprintf("\tif w.IsDirty(%s) {\n", dirtyBitName))
				if !g.isMMEObjectType(field.Type.ValueType) {
					accessorName := strings.ToLower(field.Name[0:1]) + field.Name[1:] + "Accessor"
					sb.WriteString(fmt.Sprintf("\t\tbuilder.SetNestedPath(path, \"%s\", w.%s.Clone())\n",
						fieldNameSnake, accessorName))
				} else {
					sb.WriteString(fmt.Sprintf("\t\t// TODO: Handle MME Object map field %s\n", field.Name))
				}
				sb.WriteString("\t}\n")
			} else {
				sb.WriteString(fmt.Sprintf("\tif w.IsDirty(%s) {\n", dirtyBitName))
				methodName := strings.ToUpper(field.Name[0:1]) + field.Name[1:]
				sb.WriteString(fmt.Sprintf("\t\tbuilder.SetNestedPath(path, \"%s\", w.Get%s())\n",
					fieldNameSnake, methodName))
				sb.WriteString("\t}\n")
			}
		}
		sb.WriteString("}\n\n")

		// 生成 ToIncrementalProtoWithContext 方法
		sb.WriteString("// ToIncrementalProtoWithContext 根据脏标记位构建增量数据的 protoMessage\n")
		sb.WriteString("// 只返回标记为脏的字段数据，用于增量同步\n")
		sb.WriteString(fmt.Sprintf("func (w *%s) ToIncrementalProtoWithContext(ctx mmemodel.SyncContext) proto.Message {\n", wrapperName))
		sb.WriteString("\tif w == nil {\n")
		sb.WriteString("\t\treturn nil\n")
		sb.WriteString("\t}\n\n")
		sb.WriteString("\t// 如果没有脏标记，返回 nil\n")
		sb.WriteString("\tif !w.HasAnyDirty() {\n")
		sb.WriteString("\t\treturn nil\n")
		sb.WriteString("\t}\n\n")
		sb.WriteString(fmt.Sprintf("\tincremental := &mme.%s{}\n\n", mech.Name))

		for _, field := range mech.Fields {
			fieldIndexName := fmt.Sprintf("%sFieldIndex%s", mech.Name, field.Name)
			dirtyBitName := fmt.Sprintf("%sDirty%sBit", mech.Name, field.Name)

			if field.Type.Kind == types.FieldKindXMap || field.Type.Kind == types.FieldKindMap {
				// Map 类型直接全量同步，不使用增量同步逻辑
				// 所有 map/xmap 字段都使用 mmemodel.FieldCanBeIncrementalSynced 快捷方法进行脏标记判断
				sb.WriteString(fmt.Sprintf("\tif mmemodel.FieldCanBeIncrementalSynced(w, %s, %s, ctx) {\n",
					dirtyBitName, fieldIndexName))

				accessorName := strings.ToLower(field.Name[0:1]) + field.Name[1:] + "Accessor"

				if !g.isReferenceType(field.Type.ValueType) {
					// 值类型，直接调用 Clone
					sb.WriteString(fmt.Sprintf("\t\tincremental.%s = w.%s.Clone()\n",
						field.Name, accessorName))
				} else {
					// 引用类型，留 TODO
					sb.WriteString(fmt.Sprintf("\t\t// TODO: Handle reference type map field %s (full sync)\n", field.Name))
				}
				sb.WriteString("\t}\n")
			} else {
				// 所有字段都使用 mmemodel.FieldCanBeIncrementalSynced 快捷方法进行脏标记判断
				sb.WriteString(fmt.Sprintf("\tif mmemodel.FieldCanBeIncrementalSynced(w, %s, %s, ctx) {\n",
					dirtyBitName, fieldIndexName))
				methodName := strings.ToUpper(field.Name[0:1]) + field.Name[1:]
				sb.WriteString(fmt.Sprintf("\t\tv := w.Get%s()\n", methodName))
				sb.WriteString(fmt.Sprintf("\t\tincremental.%s = &v\n", field.Name))
				sb.WriteString("\t}\n")
			}
		}

		sb.WriteString("\treturn incremental\n")
		sb.WriteString("}\n\n")

		// 写入文件（追加到机制文件）
		fileName := fmt.Sprintf("%s_mechanisms.go", CamelToSnake(mech.Name))
		filePath := fmt.Sprintf("%s/%s", outputDir, fileName)

		// 读取现有文件内容（如果存在）
		existingContent := ""
		if FileExists(filePath) {
			if data, err := ReadFileContent(filePath); err == nil {
				existingContent = data
			}
		}

		// 追加新内容（如果文件已存在，在末尾添加换行）
		var newContent string
		if existingContent != "" {
			// 确保现有内容以换行结尾
			existingContent = strings.TrimRight(existingContent, " \n\r\t")
			newContent = existingContent + "\n\n" + sb.String()
		} else {
			// 如果文件不存在，需要添加 package 声明
			newContent = fmt.Sprintf("package %s\n\n", packageName) + sb.String()
		}
		if err := WriteFile(filePath, newContent); err != nil {
			return err
		}
	}

	return nil
}

// isMMEObjectType 判断是否是 MME Object 类型
func (g *GoWrapperGenerator) isMMEObjectType(fieldType *types.FieldType) bool {
	if fieldType == nil {
		return false
	}

	// 如果是 Map/XMap，检查 ValueType
	if fieldType.Kind == types.FieldKindMap || fieldType.Kind == types.FieldKindXMap {
		if fieldType.ValueType != nil {
			return g.isMMEObjectType(fieldType.ValueType)
		}
		return false
	}

	// 获取类型名称
	typeName := ""
	if fieldType.TypeName != "" {
		typeName = fieldType.TypeName
	} else {
		typeName = fieldType.Kind.String()
	}

	// 检查是否是 Module 或 Mechanism 类型
	for _, module := range g.data.Modules {
		if module.Name == typeName {
			return true
		}
	}
	for _, mech := range g.data.Mechanisms {
		if mech.Name == typeName {
			return true
		}
	}
	return false
}

// isReferenceType 判断是否是引用类型（MME Object 或 Message 类型）
func (g *GoWrapperGenerator) isReferenceType(fieldType *types.FieldType) bool {
	if fieldType == nil {
		return false
	}

	// 如果是 Map/XMap，检查 ValueType
	if fieldType.Kind == types.FieldKindMap || fieldType.Kind == types.FieldKindXMap {
		if fieldType.ValueType != nil {
			return g.isReferenceType(fieldType.ValueType)
		}
		return false
	}

	// Message 类型和 MME Object 类型都是引用类型
	return fieldType.Kind == types.FieldKindMessage || fieldType.Kind == types.FieldKindMMEObject || g.isMMEObjectType(fieldType)
}

// isContainerType 判断是否是容器类型（xmap、xslice等，需要增量同步的容器）
// 注意可扩展性，后续可能还会有 xslice 容器
func (g *GoWrapperGenerator) isContainerType(fieldType *types.FieldType) bool {
	if fieldType == nil {
		return false
	}
	// xmap 是容器类型，后续可能还会有 xslice 等
	return fieldType.Kind == types.FieldKindXMap
}

// getTypeNameFromFieldType 从 FieldType 获取类型名称字符串
func getTypeNameFromFieldType(ft *types.FieldType) string {
	if ft == nil {
		return "unknown"
	}
	if ft.TypeName != "" {
		return ft.TypeName
	}
	return ft.Kind.String()
}

// toGoBaseTypeSimple 转换基础类型（不带包名）
func toGoBaseTypeSimple(typeStr string) string {
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
		return typeStr
	}
}

// ToGoTypeFromTypesFieldType 将新的 types.FieldType 转换为 Go 类型
func ToGoTypeFromTypesFieldType(ft *types.FieldType, packageName string) string {
	if ft == nil {
		return "unknown"
	}

	switch ft.Kind {
	case types.FieldKindRepeated:
		if ft.ValueType != nil {
			return "[]" + ToGoTypeFromTypesFieldType(ft.ValueType, packageName)
		}
		return "[]unknown"
	case types.FieldKindMap, types.FieldKindXMap:
		if ft.KeyType != nil && ft.ValueType != nil {
			keyType := ToGoTypeFromTypesFieldType(ft.KeyType, packageName)
			valueType := ToGoTypeFromTypesFieldType(ft.ValueType, packageName)
			return fmt.Sprintf("map[%s]%s", keyType, valueType)
		}
		return "map[unknown]unknown"
	case types.FieldKindInt32, types.FieldKindInt64, types.FieldKindUInt32, types.FieldKindUInt64,
		types.FieldKindFloat, types.FieldKindDouble, types.FieldKindBool, types.FieldKindString, types.FieldKindBytes:
		// 基础类型直接返回值类型，不带指针和包名
		return ToGoBaseTypeFromFieldType(ft)
	case types.FieldKindMMEObject:
		// MME Object 类型，添加包名前缀和指针
		typeName := getTypeNameFromFieldType(ft)
		// 去掉包名前缀，只取类型名
		if idx := strings.LastIndex(typeName, "."); idx >= 0 {
			typeName = typeName[idx+1:]
		}
		return "*" + typeName
	case types.FieldKindMessage:
		// 消息类型
		typeName := getTypeNameFromFieldType(ft)
		if strings.Contains(typeName, ".") {
			// 外部包类型，直接返回
			return typeName
		}
		// 本地消息类型，可能是其他包的类型，保持原样
		return typeName
	default:
		// 其他类型（如 Enum），尝试获取类型名
		typeName := getTypeNameFromFieldType(ft)
		if typeName != "" && typeName != "unknown" {
			return typeName
		}
		return "unknown"
	}
}

// ReadFileContent 读取文件内容
func ReadFileContent(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
