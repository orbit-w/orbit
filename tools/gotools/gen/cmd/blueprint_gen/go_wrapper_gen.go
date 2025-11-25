package blueprint_gen

import (
	"fmt"
	"os"
	"strings"

	types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"
)

// GoWrapperGenerator Go Wrapper 生成器
type GoWrapperGenerator struct {
	data *BlueprintContext
}

// NewGoWrapperGenerator 创建新的 Go Wrapper 生成器
func NewGoWrapperGenerator(data *BlueprintContext) *GoWrapperGenerator {
	return &GoWrapperGenerator{data: data}
}

// Generate 生成所有 Wrapper 文件
func (g *GoWrapperGenerator) Generate(outputDir string) error {
	// 生成 Mechanism Wrapper
	if err := g.generateMechanismWrappers(outputDir); err != nil {
		return fmt.Errorf("failed to generate mechanism wrappers: %w", err)
	}

	// 生成 Module Wrapper
	if err := g.generateModuleWrappers(outputDir); err != nil {
		return fmt.Errorf("failed to generate module wrappers: %w", err)
	}

	// 生成 Manager Wrapper
	if err := g.generateManagerWrappers(outputDir); err != nil {
		return fmt.Errorf("failed to generate manager wrappers: %w", err)
	}

	// 生成 Entity Wrapper
	if err := g.generateEntityWrappers(outputDir); err != nil {
		return fmt.Errorf("failed to generate entity wrappers: %w", err)
	}

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

		// 生成 map 访问器字段（只有xmap字段才需要accessor）
		for _, field := range mech.Fields {
			if field.Type.IsXMapField() {
				// 判断 Value 类型
				if field.Type.IsXMapValueMMEObject() {
					// 目前机制中，不支持用xmap 编排MME Object 类型
					panic(fmt.Sprintf("mechanism %s 的xmap字段 %s 的Value类型是 MMEObject 类型，不支持用xmap编排MME Object 类型", mech.Name, field.Name))
					// TODO: 生成 XMapWrapper 字段
				} else {
					// 值类型，使用 MapAccessor
					keyType := field.Type.KeyKind().String()
					valueType := field.Type.ValueKind().String()
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

		// 初始化 map 访问器（只有xmap字段才需要accessor，普通map不需要）
		for _, field := range mech.Fields {
			if field.Type.IsXMapField() && !field.Type.IsXMapValueMMEObject() {
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
			if field.Type.IsXMapField() {
				// XMap 字段生成访问器方法
				accessorName := strings.ToLower(field.Name[0:1]) + field.Name[1:]
				methodName := strings.ToUpper(field.Name[0:1]) + field.Name[1:]
				if field.Type.IsXMapValueMMEObject() {
					panic(fmt.Sprintf("mechanism %s 的xmap字段 %s 的Value类型是 MMEObject 类型，不支持用xmap编排MME Object 类型", mech.Name, field.Name))
				} else {
					keyType := field.Type.KeyKind().String()
					valueType := field.Type.ValueKind().String()
					sb.WriteString(fmt.Sprintf("// 包装器-获取%s访问器\n", field.Name))
					sb.WriteString(fmt.Sprintf("func (w *%s) Get%sAccessor() *xmap.MapAccessor[%s, %s] {\n",
						wrapperName, methodName, keyType, valueType))
					sb.WriteString(fmt.Sprintf("\treturn w.%sAccessor\n", accessorName))
					sb.WriteString("}\n\n")
				}
			} else if field.Type.IsMapField() {
				// 普通 Map 字段生成 Getter 方法，直接返回 map
				methodName := strings.ToUpper(field.Name[0:1]) + field.Name[1:]
				goType := ToGoTypeFromTypesFieldType(&field.Type, packageName)
				sb.WriteString(fmt.Sprintf("// 包装器-获取%s\n", field.Name))
				sb.WriteString(fmt.Sprintf("func (w *%s) Get%s() %s {\n", wrapperName, methodName, goType))
				sb.WriteString(fmt.Sprintf("\treturn w.data.%s\n", field.Name))
				sb.WriteString("}\n\n")
				// 生成 Setter 方法，标记脏标记
				dirtyBitName := fmt.Sprintf("%sDirty%sBit", mech.Name, field.Name)
				sb.WriteString(fmt.Sprintf("// 包装器-设置%s\n", field.Name))
				sb.WriteString(fmt.Sprintf("func (w *%s) Set%s(v %s) {\n", wrapperName, methodName, goType))
				sb.WriteString(fmt.Sprintf("\tw.data.%s = v\n", field.Name))
				sb.WriteString(fmt.Sprintf("\tw.MarkDirty(%s)\n", dirtyBitName))
				sb.WriteString("}\n\n")
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

		// 生成 ClearAllDirtyFlags 方法
		sb.WriteString("// ClearAllDirtyFlags 清除所有脏标记位\n")
		sb.WriteString(fmt.Sprintf("func (w *%s) ClearAllDirtyFlags() {\n", wrapperName))
		sb.WriteString("\tw.IDirtyFlag.ClearAllDirty()\n\n")
		sb.WriteString("\t// 如果Value为引用类型且有脏标记，则清除所有xmap中Value的脏标记\n\n")
		sb.WriteString("\t// 清除所有xmap中的操作记录\n")
		// 只处理值类型的 xmap 字段（普通map不需要增量策略）
		for _, field := range mech.Fields {
			if field.IsXMapField() && !field.Type.IsXMapValueMMEObject() {
				accessorName := strings.ToLower(field.Name[0:1]) + field.Name[1:] + "Accessor"
				sb.WriteString(fmt.Sprintf("\tw.%s.ResetOperations()\n", accessorName))
			}
		}
		sb.WriteString("}\n\n")

		// 生成 BuildMongoUpdate 方法
		generator := &BuildMongoUpdateCodeGenerator{
			ObjectName:  mech.Name,
			WrapperName: wrapperName,
			Receiver:    "w",
			ObjectType:  ObjectTypeMechanism,
		}
		sb.WriteString(generator.GenerateBuildMongoUpdateMethod(mech.Fields))

		generatorToIncrementalProto := &ToIncrementalProtoCodeGenerator{
			ObjectName:  mech.Name,
			WrapperName: wrapperName,
			Receiver:    "w",
			ObjectType:  ObjectTypeMechanism,
		}
		sb.WriteString(generatorToIncrementalProto.GenerateToIncrementalProtoMethod(mech.Fields))

		wrapperMethodCodeGenerator := &WrapperMethodCodeGenerator{
			ObjectName:  mech.Name,
			WrapperName: wrapperName,
			Receiver:    "w",
			ProtoPkg:    "mme",
			ObjectType:  ObjectTypeMechanism,
		}

		sb.WriteString(wrapperMethodCodeGenerator.GenerateToProtoMethod())
		sb.WriteString(wrapperMethodCodeGenerator.GenerateFromProtoMethod(mech.Fields))
		sb.WriteString(wrapperMethodCodeGenerator.GenerateDeepCopyMethod())
		sb.WriteString(wrapperMethodCodeGenerator.GenerateDeepCopyToMethod(mech.Fields))

		// 写入文件（追加到机制文件）
		fileName := GetMechanismFileName(mech.Name)
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

const (
	unknownType = "unknown"
)

// ToGoTypeFromTypesFieldType 将新的 types.FieldType 转换为 Go 类型
func ToGoTypeFromTypesFieldType(ft *types.FieldType, packageName string) string {
	if ft == nil {
		return unknownType
	}

	switch ft.Kind {
	case types.FieldKindRepeated:
		if ft.ValueType == nil {
			return "[]" + unknownType
		}
		return "[]" + ToGoTypeFromTypesFieldType(ft.ValueType, packageName)

	case types.FieldKindMap, types.FieldKindXMap:
		if ft.KeyType == nil || ft.ValueType == nil {
			return fmt.Sprintf("map[%s]%s", unknownType, unknownType)
		}
		keyType := ToGoTypeFromTypesFieldType(ft.KeyType, packageName)
		valueType := ToGoTypeFromTypesFieldType(ft.ValueType, packageName)
		return fmt.Sprintf("map[%s]%s", keyType, valueType)

	case types.FieldKindInt32, types.FieldKindInt64, types.FieldKindUInt32, types.FieldKindUInt64,
		types.FieldKindFloat, types.FieldKindDouble, types.FieldKindBool, types.FieldKindString, types.FieldKindBytes:
		// 基础类型直接返回值类型，不带指针和包名
		return ft.GetKind().String()

	case types.FieldKindMMEObject:
		// MME Object 类型，添加指针前缀
		typeName := ft.Name
		return "*" + typeName

	case types.FieldKindMessage:
		// 消息类型，外部包类型直接返回，本地类型保持原样
		return ft.GetTypeName()

	default:
		// 其他类型（如 Enum），尝试获取类型名
		typeName := ft.GetTypeName()
		if typeName != "" && typeName != unknownType {
			return typeName
		}
		return unknownType
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

// generateModuleWrappers 生成 Module Wrapper
func (g *GoWrapperGenerator) generateModuleWrappers(outputDir string) error {
	packageName := "mme"

	for _, module := range g.data.Modules {
		sb := strings.Builder{}

		// Wrapper 代码（不包含 package 和 import，因为这些会从已有文件中获取）

		// 生成 Wrapper 结构体
		wrapperName := module.Name + "Wrapper"
		sb.WriteString(fmt.Sprintf("type %s struct {\n", wrapperName))
		sb.WriteString(fmt.Sprintf("\tdata *%s\n", module.Name))
		sb.WriteString("\tdirtyflag.IDirtyFlag\n")
		sb.WriteString("\tfieldMetas *fieldmeta.FieldMetas\n\n")

		// 生成嵌套的 Mechanism Wrapper 字段
		for _, field := range module.Fields {
			if field.Type.IsMMEObjectType() {
				typeName := field.GetTypeName()
				wrapperTypeName := typeName + "Wrapper"
				sb.WriteString(fmt.Sprintf("\t%s *%s\n", field.Name+"Wrapper", wrapperTypeName))
			}
		}

		sb.WriteString("}\n\n")

		// 生成 New 构造函数
		sb.WriteString(fmt.Sprintf("func New%s(data *%s) *%s {\n", wrapperName, module.Name, wrapperName))
		sb.WriteString("\tif data == nil {\n")
		sb.WriteString("\t\tpanic(\"data is nil\")\n")
		sb.WriteString("\t}\n")
		sb.WriteString(fmt.Sprintf("\tw := &%s{\n", wrapperName))
		sb.WriteString("\t\tdata:       data,\n")
		sb.WriteString("\t\tIDirtyFlag: dirtyflag.NewDirtyFlag(),\n")
		sb.WriteString("\t\tfieldMetas: fieldmeta.NewFieldMetas(),\n")
		sb.WriteString("\t}\n\n")

		// 初始化嵌套的 Mechanism Wrapper
		for _, field := range module.Fields {
			if field.Type.IsMMEObjectType() {
				typeName := field.GetTypeName()
				wrapperFieldName := field.Name + "Wrapper"
				dirtyBitName := fmt.Sprintf("%sDirty%sBit", module.Name, field.Name)
				sb.WriteString(fmt.Sprintf("\t// 初始化嵌套的 %s Wrapper\n", typeName))
				sb.WriteString(fmt.Sprintf("\tif data.%s == nil {\n", field.Name))
				sb.WriteString(fmt.Sprintf("\t\tdata.%s = New%s()\n", field.Name, typeName))
				sb.WriteString("\t}\n")
				sb.WriteString(fmt.Sprintf("\tw.%s = New%sWrapper(data.%s)\n",
					wrapperFieldName, typeName, field.Name))
				sb.WriteString(fmt.Sprintf("\tw.%s.Link(w.GetDirtyTracker(), %s)\n",
					wrapperFieldName, dirtyBitName))
				sb.WriteString("\n")
			}
		}

		sb.WriteString("\treturn w\n")
		sb.WriteString("}\n\n")

		// 生成 InitFieldContext 方法
		sb.WriteString(fmt.Sprintf("func (w *%s) InitFieldContext() {\n", wrapperName))
		for _, field := range module.Fields {
			fieldIndexName := fmt.Sprintf("%sFieldIndex%s", module.Name, field.Name)
			// 根据 access 选项设置字段类型
			if field.Options.Access == "all" || field.Options.Access == "s" || field.Options.Access == "" {
				sb.WriteString(fmt.Sprintf("\tw.fieldMetas.SetFieldType(%s, fieldmeta.FieldTypeSync)\n",
					fieldIndexName))
			}
		}
		sb.WriteString("\n")
		// 初始化嵌套 Mechanism 的字段上下文
		for _, field := range module.Fields {
			if field.Type.IsMMEObjectType() {
				wrapperFieldName := field.Name + "Wrapper"
				sb.WriteString(fmt.Sprintf("\tif w.%s != nil {\n", wrapperFieldName))
				sb.WriteString(fmt.Sprintf("\t\tw.%s.InitFieldContext()\n", wrapperFieldName))
				sb.WriteString("\t}\n")
			}
		}
		sb.WriteString("}\n\n")

		// 生成 Name 方法
		sb.WriteString(fmt.Sprintf("func (w *%s) Name() string {\n", wrapperName))
		sb.WriteString(fmt.Sprintf("\treturn \"%s\"\n", module.Name))
		sb.WriteString("}\n\n")

		// 生成 MatchesAll 方法（实现 IFieldMetaContext 接口）
		sb.WriteString("// MatchesAll 判断字段是否匹配所有类型标记\n")
		sb.WriteString(fmt.Sprintf("func (w *%s) MatchesAll(fieldID uint8, fieldTypes ...fieldmeta.FieldType) bool {\n", wrapperName))
		sb.WriteString("\treturn w.fieldMetas.MatchesAll(fieldID, fieldTypes...)\n")
		sb.WriteString("}\n\n")

		// 生成 Getter 方法（获取嵌套的 Mechanism Wrapper）
		for _, field := range module.Fields {
			if field.Type.IsMMEObjectType() {
				typeName := field.GetTypeName()
				wrapperFieldName := field.Name + "Wrapper"
				methodName := strings.ToUpper(field.Name[0:1]) + field.Name[1:]
				sb.WriteString(fmt.Sprintf("func (w *%s) Get%s() *%sWrapper {\n",
					wrapperName, methodName, typeName))
				sb.WriteString(fmt.Sprintf("\treturn w.%s\n", wrapperFieldName))
				sb.WriteString("}\n\n")
			}
		}

		// 生成 ClearAllDirtyFlags 方法
		sb.WriteString("// ClearAllDirtyFlags 清除所有脏标记位\n")
		sb.WriteString(fmt.Sprintf("func (w *%s) ClearAllDirtyFlags() {\n", wrapperName))
		sb.WriteString("\tw.ClearAllDirty()\n\n")
		sb.WriteString("\t// 清除嵌套 Mechanism 的脏标记\n")
		for _, field := range module.Fields {
			if field.Type.IsMMEObjectType() {
				wrapperFieldName := field.Name + "Wrapper"
				sb.WriteString(fmt.Sprintf("\tif w.%s != nil {\n", wrapperFieldName))
				sb.WriteString(fmt.Sprintf("\t\tw.%s.ClearAllDirtyFlags()\n", wrapperFieldName))
				sb.WriteString("\t}\n")
			}
		}
		sb.WriteString("}\n\n")

		// 生成 BuildMongoUpdate 方法
		generator := &BuildMongoUpdateCodeGenerator{
			ObjectName:  module.Name,
			WrapperName: wrapperName,
			Receiver:    "w",
			ObjectType:  ObjectTypeModule,
		}
		sb.WriteString(generator.GenerateBuildMongoUpdateMethod(module.Fields))

		wrapperMethodCodeGenerator := &WrapperMethodCodeGenerator{
			ObjectName:  module.Name,
			WrapperName: wrapperName,
			Receiver:    "w",
			ProtoPkg:    "mme",
			ObjectType:  ObjectTypeModule,
		}
		sb.WriteString(wrapperMethodCodeGenerator.GenerateToProtoMethod())
		sb.WriteString(wrapperMethodCodeGenerator.GenerateFromProtoMethod(module.Fields))
		sb.WriteString(wrapperMethodCodeGenerator.GenerateDeepCopyMethod())
		sb.WriteString(wrapperMethodCodeGenerator.GenerateDeepCopyToMethod(module.Fields))

		// 生成 ToIncrementalProtoWithContext 方法
		generatorToIncrementalProto := &ToIncrementalProtoCodeGenerator{
			ObjectName:  module.Name,
			WrapperName: wrapperName,
			Receiver:    "w",
			ObjectType:  ObjectTypeModule,
		}
		sb.WriteString(generatorToIncrementalProto.GenerateToIncrementalProtoMethod(module.Fields))

		// 写入文件（追加到模块文件）
		fileName := GetModuleFileName(module.Name)
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
			// 如果文件不存在，需要添加 package 声明和导入
			imports := "import (\n"
			imports += "\t\"gitee.com/orbit-w/orbit/app/proto/mme\"\n"
			imports += "\tdirtyflag \"gitee.com/orbit-w/orbit/lib/base/dirty_flag\"\n"
			imports += "\tfieldmeta \"gitee.com/orbit-w/orbit/lib/base/field_meta\"\n"
			imports += "\t\"gitee.com/orbit-w/orbit/lib/module/db/mgo_builder\"\n"
			imports += "\tmmemodel \"gitee.com/orbit-w/orbit/lib/module/mme_model\"\n"
			imports += "\t\"google.golang.org/protobuf/proto\"\n"
			imports += ")\n\n"
			newContent = fmt.Sprintf("package %s\n\n%s", packageName, imports) + sb.String()
		}
		if err := WriteFile(filePath, newContent); err != nil {
			return err
		}
	}

	return nil
}

// generateManagerWrappers 生成 Manager Wrapper
func (g *GoWrapperGenerator) generateManagerWrappers(outputDir string) error {
	for _, manager := range g.data.Managers {
		sb := strings.Builder{}

		// 生成 Wrapper 结构体
		wrapperName := manager.Name + "Wrapper"
		sb.WriteString(fmt.Sprintf("type %s struct {\n", wrapperName))
		sb.WriteString(fmt.Sprintf("\tdata *%s\n", manager.Name))
		sb.WriteString("\tdirtyflag.IDirtyFlag\n")
		sb.WriteString("\tfieldMetas *fieldmeta.FieldMetas\n\n")

		// 生成 map 字段的 XMapWrapper（Manager 的 map 字段的 value 通常是 Module 类型）
		for _, field := range manager.Fields {
			if field.Type.IsXMapField() || field.Type.IsMapField() {
				// Manager 的 map 字段的 value 必须是 MME Object（Module 类型）
				if field.Type.IsXMapField() {
					if !field.Type.IsXMapValueMMEObject() {
						panic(fmt.Sprintf("manager %s 的xmap字段 %s 的Value类型必须是 MMEObject 类型（Module）", manager.Name, field.Name))
					}
				} else if field.Type.IsMapField() {
					if !field.Type.IsMapValueMMEObject() {
						panic(fmt.Sprintf("manager %s 的map字段 %s 的Value类型必须是 MMEObject 类型（Module）", manager.Name, field.Name))
					}
				}

				keyType := field.Type.KeyKind().String()
				valueType := field.GetValueName()
				wrapperTypeName := valueType + "Wrapper"
				linkFieldName := strings.ToLower(field.Name[0:1]) + field.Name[1:] + "Link"
				sb.WriteString("\t//Value为MME Object，使用xmap.MapAccessor进行包装\n")
				sb.WriteString(fmt.Sprintf("\t%s *xmapwrapper.XMapWrapper[%s, *%s, *%s]\n",
					linkFieldName, keyType, valueType, wrapperTypeName))
			}
		}

		sb.WriteString("}\n\n")

		// 生成 New 构造函数
		sb.WriteString(fmt.Sprintf("func New%s(data *%s) *%s {\n", wrapperName, manager.Name, wrapperName))
		sb.WriteString("\tif data == nil {\n")
		sb.WriteString("\t\tpanic(\"data is nil\")\n")
		sb.WriteString("\t}\n")
		sb.WriteString(fmt.Sprintf("\tm := &%s{\n", wrapperName))
		sb.WriteString("\t\tdata:       data,\n")
		sb.WriteString("\t\tIDirtyFlag: dirtyflag.NewDirtyFlag(),\n")
		sb.WriteString("\t\tfieldMetas: fieldmeta.NewFieldMetas(),\n")
		sb.WriteString("\t}\n\n")

		// 初始化 XMapWrapper
		for _, field := range manager.Fields {
			if field.Type.IsXMapField() || field.Type.IsMapField() {
				valueName := field.GetValueName()
				linkFieldName := strings.ToLower(field.Name[0:1]) + field.Name[1:] + "Link"
				dirtyBitName := fmt.Sprintf("%sDirty%sBit", manager.Name, field.Name)
				newWrapperFuncName := fmt.Sprintf("New%sWrapper", valueName)
				sb.WriteString(fmt.Sprintf("\tm.%s = xmapwrapper.NewXMapWrapperWithParent(\n",
					linkFieldName))
				sb.WriteString(fmt.Sprintf("\t\t&data.%s,\n", field.Name))
				sb.WriteString("\t\tm.GetDirtyTracker(),\n")
				sb.WriteString(fmt.Sprintf("\t\t%s,\n", dirtyBitName))
				sb.WriteString(fmt.Sprintf("\t\t%s,\n", newWrapperFuncName))
				sb.WriteString("\t)\n\n")
			}
		}

		sb.WriteString("\treturn m\n")
		sb.WriteString("}\n\n")

		// 生成 InitFieldContext 方法
		sb.WriteString(fmt.Sprintf("func (m *%s) InitFieldContext() {\n", wrapperName))
		for _, field := range manager.Fields {
			fieldIndexName := fmt.Sprintf("%sFieldIndex%s", manager.Name, field.Name)
			// 根据 access 选项设置字段类型
			if field.Options.Access == "all" || field.Options.Access == "s" || field.Options.Access == "" {
				sb.WriteString(fmt.Sprintf("\tm.fieldMetas.SetFieldType(%s, fieldmeta.FieldTypeSync)\n",
					fieldIndexName))
			}
		}
		sb.WriteString("\n")
		// 初始化所有 Module 的字段上下文
		for _, field := range manager.Fields {
			if field.Type.IsXMapField() || field.Type.IsMapField() {
				linkFieldName := strings.ToLower(field.Name[0:1]) + field.Name[1:] + "Link"
				keyType := field.Type.KeyKind().String()
				valueName := field.GetValueName()
				wrapperTypeName := valueName + "Wrapper"
				sb.WriteString(fmt.Sprintf("\t// 初始化所有 %s 的字段上下文\n", valueName))
				sb.WriteString(fmt.Sprintf("\tm.%s.Range(func(key %s, wrapper *%s) bool {\n",
					linkFieldName, keyType, wrapperTypeName))
				sb.WriteString("\t\twrapper.InitFieldContext()\n")
				sb.WriteString("\t\treturn true\n")
				sb.WriteString("\t})\n")
			}
		}
		sb.WriteString("}\n\n")

		// 生成 Name 方法
		sb.WriteString("// Manager 唯一名称\n")
		sb.WriteString(fmt.Sprintf("func (m *%s) Name() string {\n", wrapperName))
		sb.WriteString(fmt.Sprintf("\treturn \"%s\"\n", manager.Name))
		sb.WriteString("}\n\n")

		// 生成 MatchesAll 方法
		sb.WriteString("// MatchesAll 判断字段是否匹配所有类型标记\n")
		sb.WriteString(fmt.Sprintf("func (m *%s) MatchesAll(fieldID uint8, fieldTypes ...fieldmeta.FieldType) bool {\n", wrapperName))
		sb.WriteString("\treturn m.fieldMetas.MatchesAll(fieldID, fieldTypes...)\n")
		sb.WriteString("}\n\n")

		// 生成 map 字段的工具方法（Set, Delete, Range）
		for _, field := range manager.Fields {
			if field.Type.IsXMapField() || field.Type.IsMapField() {
				keyType := field.Type.KeyKind().String()
				valueName := field.GetValueName()
				wrapperTypeName := valueName + "Wrapper"
				linkFieldName := strings.ToLower(field.Name[0:1]) + field.Name[1:] + "Link"
				methodPrefix := strings.ToUpper(field.Name[0:1]) + field.Name[1:]

				// Set 方法
				sb.WriteString(fmt.Sprintf("// Set%s 设置/添加%s模块\n", methodPrefix, valueName))
				sb.WriteString(fmt.Sprintf("func (m *%s) %s_Set(id %s, value *%s) *%s {\n",
					wrapperName, field.Name, keyType, valueName, wrapperTypeName))
				sb.WriteString(fmt.Sprintf("\treturn m.%s.Set(id, value)\n", linkFieldName))
				sb.WriteString("}\n\n")

				// Delete 方法
				sb.WriteString(fmt.Sprintf("// Delete%s 删除%s模块\n", methodPrefix, valueName))
				sb.WriteString(fmt.Sprintf("func (m *%s) %s_Delete(id %s) bool {\n",
					wrapperName, field.Name, keyType))
				sb.WriteString(fmt.Sprintf("\treturn m.%s.Delete(id)\n", linkFieldName))
				sb.WriteString("}\n\n")

				// Range 方法
				sb.WriteString(fmt.Sprintf("// Range%s 遍历所有%s\n", methodPrefix, valueName))
				sb.WriteString(fmt.Sprintf("func (m *%s) %s_Range(f func(id %s, %s *%s) bool) {\n",
					wrapperName, field.Name, keyType, strings.ToLower(valueName[0:1])+valueName[1:], wrapperTypeName))
				sb.WriteString(fmt.Sprintf("\tm.%s.Range(f)\n", linkFieldName))
				sb.WriteString("}\n\n")
			}
		}

		// 生成 ClearAllDirty 方法
		sb.WriteString(fmt.Sprintf("func (m *%s) ClearAllDirtyFlags() {\n", wrapperName))
		sb.WriteString("\tm.IDirtyFlag.ClearAllDirty()\n\n")
		sb.WriteString("\t// 清空xmaplink中所有object的脏标记\n")
		for _, field := range manager.Fields {
			if field.Type.IsXMapField() || field.Type.IsMapField() {
				linkFieldName := strings.ToLower(field.Name[0:1]) + field.Name[1:] + "Link"
				keyKind := field.Type.KeyKind()
				sb.WriteString(fmt.Sprintf("\tm.%s.RangeIncrementalSyncObject(func(key %s, object xmapwrapper.IncrementalSyncObject) (stop bool) {\n",
					linkFieldName, keyKind.String()))
				sb.WriteString("\t\tobject.ClearAllDirty()\n")
				sb.WriteString("\t\treturn false\n")
				sb.WriteString("\t})\n")
			}
		}
		sb.WriteString("}\n\n")

		// 生成 Reset 方法
		sb.WriteString(fmt.Sprintf("func (m *%s) Reset(newData *%s) {\n", wrapperName, manager.Name))
		sb.WriteString("\tif m == nil || m.data == nil {\n")
		sb.WriteString("\t\treturn\n")
		sb.WriteString("\t}\n\n")
		sb.WriteString("\t// 重新构建脏标系统\n")
		sb.WriteString("\tm.IDirtyFlag.ClearAllDirty()\n\n")
		sb.WriteString("\t// xmap全量覆盖数据，不需要处理Value的逻辑\n")
		sb.WriteString(fmt.Sprintf("\tdata := New%s()\n", manager.Name))
		sb.WriteString("\tm.data = data\n")
		for _, field := range manager.Fields {
			if field.Type.IsXMapField() || field.Type.IsMapField() {
				linkFieldName := strings.ToLower(field.Name[0:1]) + field.Name[1:] + "Link"
				sb.WriteString(fmt.Sprintf("\tm.%s.Reset(&m.data.%s)\n", linkFieldName, field.Name))
			}
		}
		sb.WriteString("}\n\n")

		// 生成 BuildMongoUpdate 方法
		generator := &BuildMongoUpdateCodeGenerator{
			ObjectName:  manager.Name,
			WrapperName: wrapperName,
			Receiver:    "m",
			ObjectType:  ObjectTypeManager,
		}
		sb.WriteString(generator.GenerateBuildMongoUpdateMethod(manager.Fields))

		wrapperMethodCodeGenerator := &WrapperMethodCodeGenerator{
			ObjectName:  manager.Name,
			WrapperName: wrapperName,
			Receiver:    "m",
			ProtoPkg:    "mme",
			ObjectType:  ObjectTypeManager,
		}
		sb.WriteString(wrapperMethodCodeGenerator.GenerateToProtoMethod())
		sb.WriteString(wrapperMethodCodeGenerator.GenerateFromProtoMethod(manager.Fields))
		sb.WriteString(wrapperMethodCodeGenerator.GenerateDeepCopyMethod())
		sb.WriteString(wrapperMethodCodeGenerator.GenerateDeepCopyToMethod(manager.Fields))

		// 生成 ToIncrementalProto 方法
		generatorToIncrementalProto := &ToIncrementalProtoCodeGenerator{
			ObjectName:  manager.Name,
			WrapperName: wrapperName,
			Receiver:    "m",
			ObjectType:  ObjectTypeManager,
		}
		sb.WriteString(generatorToIncrementalProto.GenerateToIncrementalProtoMethod(manager.Fields))

		// 写入文件（追加到 Manager 文件）
		fileName := GetManagerFileName(manager.Name)
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
		}
		if err := WriteFile(filePath, newContent); err != nil {
			return err
		}
	}

	return nil
}

// generateEntityWrappers 生成 Entity Wrapper
func (g *GoWrapperGenerator) generateEntityWrappers(outputDir string) error {
	if len(g.data.Entities) == 0 {
		return nil
	}

	packageName := "mme"
	firstEntity := g.data.Entities[0]
	firstEntityFileName := GetEntityFileName(firstEntity.Name)
	firstEntityFilePath := fmt.Sprintf("%s/%s", outputDir, firstEntityFileName)

	// 检查并生成样板代码
	existingContent, err := g.ensureEntityBoilerplate(firstEntityFilePath, packageName)
	if err != nil {
		return fmt.Errorf("failed to ensure entity boilerplate: %w", err)
	}

	// 生成所有实体的 Wrapper 代码
	for idx, entity := range g.data.Entities {
		sb := strings.Builder{}

		// Wrapper 代码（不包含 package 和 import，因为这些会从已有文件中获取）

		// 生成 Wrapper 结构体
		wrapperName := entity.Name + "Wrapper"
		sb.WriteString(fmt.Sprintf("type %s struct {\n", wrapperName))
		sb.WriteString(fmt.Sprintf("\tdata *%s\n", entity.Name))
		sb.WriteString("\tdirtyflag.IDirtyFlag\n")
		sb.WriteString("\tfieldMetas *fieldmeta.FieldMetas\n\n")

		// 生成嵌套的 Manager Wrapper 字段
		for _, field := range entity.Fields {
			if field.Type.IsMMEObjectType() {
				typeName := field.GetTypeName()
				wrapperTypeName := typeName + "Wrapper"
				sb.WriteString(fmt.Sprintf("\t%s *%s\n", field.Name+"Wrapper", wrapperTypeName))
			}
		}

		sb.WriteString("}\n\n")

		// 生成 New 构造函数（无参数版本）
		sb.WriteString(fmt.Sprintf("func New%s() *%s {\n", wrapperName, wrapperName))
		sb.WriteString(fmt.Sprintf("\treturn &%s{\n", wrapperName))
		sb.WriteString("\t\tIDirtyFlag: dirtyflag.NewDirtyFlag(),\n")
		sb.WriteString("\t\tfieldMetas: fieldmeta.NewFieldMetas(),\n")
		sb.WriteString("\t}\n")
		sb.WriteString("}\n\n")

		// 生成 Collection 方法
		collectionName := GetEntityCollectionName(entity.Name)
		sb.WriteString(fmt.Sprintf("func (e *%s) Collection() string {\n", wrapperName))
		sb.WriteString(fmt.Sprintf("\treturn \"%s\"\n", collectionName))
		sb.WriteString("}\n\n")

		// 生成 HasAnyDirty 方法
		sb.WriteString(fmt.Sprintf("func (e *%s) HasAnyDirty() bool {\n", wrapperName))
		sb.WriteString("\treturn e.IDirtyFlag.HasAnyDirty()\n")
		sb.WriteString("}\n\n")

		// 生成 Load 方法
		sb.WriteString(fmt.Sprintf("func (e *%s) Load(raw bson.Raw) error {\n", wrapperName))
		sb.WriteString(fmt.Sprintf("\tdata := New%s()\n", entity.Name))
		sb.WriteString("\tif err := bson.Unmarshal(raw, data); err != nil {\n")
		sb.WriteString("\t\treturn err\n")
		sb.WriteString("\t}\n")
		sb.WriteString("\te.data = data\n\n")

		// 初始化嵌套的 Manager Wrapper
		for _, field := range entity.Fields {
			if field.Type.IsMMEObjectType() {
				typeName := field.GetTypeName()
				wrapperFieldName := field.Name + "Wrapper"
				dirtyBitName := fmt.Sprintf("%sDirty%sBit", entity.Name, field.Name)
				sb.WriteString(fmt.Sprintf("\t// 初始化嵌套的 %s Wrapper\n", typeName))
				sb.WriteString(fmt.Sprintf("\tif data.%s == nil {\n", field.Name))
				sb.WriteString(fmt.Sprintf("\t\tdata.%s = New%s()\n", field.Name, typeName))
				sb.WriteString("\t}\n")
				sb.WriteString(fmt.Sprintf("\te.%s = New%sWrapper(data.%s)\n",
					wrapperFieldName, typeName, field.Name))
				sb.WriteString(fmt.Sprintf("\te.%s.Link(e.GetDirtyTracker(), %s)\n",
					wrapperFieldName, dirtyBitName))
				sb.WriteString("\n")
			}
		}

		sb.WriteString("\treturn nil\n")
		sb.WriteString("}\n\n")

		// 生成 InitFieldContext 方法
		sb.WriteString(fmt.Sprintf("func (e *%s) InitFieldContext() {\n", wrapperName))
		// XXXId 字段的字段上下文
		fieldIndexName := fmt.Sprintf("%sFieldIndexXXXId", entity.Name)
		sb.WriteString(fmt.Sprintf("\te.fieldMetas.SetFieldType(%s, fieldmeta.FieldTypeSync)\n",
			fieldIndexName))
		for _, field := range entity.Fields {
			fieldIndexName := fmt.Sprintf("%sFieldIndex%s", entity.Name, field.Name)
			// 根据 access 选项设置字段类型
			if field.Options.Access == "all" || field.Options.Access == "s" || field.Options.Access == "" {
				sb.WriteString(fmt.Sprintf("\te.fieldMetas.SetFieldType(%s, fieldmeta.FieldTypeSync)\n",
					fieldIndexName))
			}
		}
		sb.WriteString("\n")
		// 初始化嵌套 Manager 的字段上下文
		for _, field := range entity.Fields {
			if field.Type.IsMMEObjectType() {
				wrapperFieldName := field.Name + "Wrapper"
				sb.WriteString(fmt.Sprintf("\tif e.%s != nil {\n", wrapperFieldName))
				sb.WriteString(fmt.Sprintf("\t\te.%s.InitFieldContext()\n", wrapperFieldName))
				sb.WriteString("\t}\n")
			}
		}
		sb.WriteString("}\n\n")

		// 生成 Name 方法
		sb.WriteString(fmt.Sprintf("func (e *%s) Name() string {\n", wrapperName))
		sb.WriteString(fmt.Sprintf("\treturn \"%s\"\n", entity.Name))
		sb.WriteString("}\n\n")

		// 生成 GetEntityType 方法
		sb.WriteString("// GetEntityType 返回实体类型枚举\n")
		sb.WriteString(fmt.Sprintf("func (e *%s) GetEntityType() mme.EntityType {\n", wrapperName))
		sb.WriteString(fmt.Sprintf("\treturn mme.EntityType_%s\n", GenEntityTypeEnumName(entity.Name)))
		sb.WriteString("}\n\n")

		// 生成 MatchesAll 方法（实现 IFieldMetaContext 接口）
		sb.WriteString("// MatchesAll 判断字段是否匹配所有类型标记\n")
		sb.WriteString(fmt.Sprintf("func (e *%s) MatchesAll(fieldID uint8, fieldTypes ...fieldmeta.FieldType) bool {\n", wrapperName))
		sb.WriteString("\treturn e.fieldMetas.MatchesAll(fieldID, fieldTypes...)\n")
		sb.WriteString("}\n\n")

		// 生成 XXXId 的 Getter 和 Setter 方法
		sb.WriteString("// 包装器-获取XXXId\n")
		sb.WriteString(fmt.Sprintf("func (e *%s) GetXXXId() int64 {\n", wrapperName))
		sb.WriteString("\treturn e.data.XXXId\n")
		sb.WriteString("}\n\n")
		dirtyBitName := fmt.Sprintf("%sDirtyXXXIdBit", entity.Name)
		sb.WriteString("// 包装器-设置XXXId\n")
		sb.WriteString(fmt.Sprintf("func (e *%s) SetXXXId(v int64) {\n", wrapperName))
		sb.WriteString("\te.data.XXXId = v\n")
		sb.WriteString(fmt.Sprintf("\te.MarkDirty(%s)\n", dirtyBitName))
		sb.WriteString("}\n\n")

		// 生成 Getter 方法（获取嵌套的 Manager Wrapper）
		for _, field := range entity.Fields {
			if field.Type.IsMMEObjectType() {
				typeName := field.GetTypeName()
				wrapperFieldName := field.Name + "Wrapper"
				methodName := strings.ToUpper(field.Name[0:1]) + field.Name[1:]
				sb.WriteString(fmt.Sprintf("func (e *%s) Get%s() *%sWrapper {\n",
					wrapperName, methodName, typeName))
				sb.WriteString(fmt.Sprintf("\treturn e.%s\n", wrapperFieldName))
				sb.WriteString("}\n\n")
			}
		}

		// 生成 ClearAllDirtyFlags 方法
		sb.WriteString("// ClearAllDirtyFlags 清除所有脏标记位\n")
		sb.WriteString(fmt.Sprintf("func (e *%s) ClearAllDirtyFlags() {\n", wrapperName))
		sb.WriteString("\te.ClearAllDirty()\n\n")
		sb.WriteString("\t// 清除嵌套 Manager 的脏标记\n")
		for _, field := range entity.Fields {
			if field.Type.IsMMEObjectType() {
				wrapperFieldName := field.Name + "Wrapper"
				sb.WriteString(fmt.Sprintf("\tif e.%s != nil {\n", wrapperFieldName))
				sb.WriteString(fmt.Sprintf("\t\te.%s.ClearAllDirtyFlags()\n", wrapperFieldName))
				sb.WriteString("\t}\n")
			}
		}
		sb.WriteString("}\n\n")

		// 生成 BuildMongoUpdate 方法
		generator := &BuildMongoUpdateCodeGenerator{
			ObjectName:  entity.Name,
			WrapperName: wrapperName,
			Receiver:    "e",
			ObjectType:  ObjectTypeEntity,
		}
		sb.WriteString(generator.GenerateBuildMongoUpdateMethod(entity.Fields))

		wrapperMethodCodeGenerator := &WrapperMethodCodeGenerator{
			ObjectName:  entity.Name,
			WrapperName: wrapperName,
			Receiver:    "e",
			ProtoPkg:    "mme",
			ObjectType:  ObjectTypeEntity,
		}
		sb.WriteString(wrapperMethodCodeGenerator.GenerateToProtoMethod())
		sb.WriteString(wrapperMethodCodeGenerator.GenerateFromProtoMethod(entity.Fields))
		sb.WriteString(wrapperMethodCodeGenerator.GenerateDeepCopyMethod())
		sb.WriteString(wrapperMethodCodeGenerator.GenerateDeepCopyToMethod(entity.Fields))

		// 生成 ToIncrementalProtoWithContext 方法
		generatorToIncrementalProto := &ToIncrementalProtoCodeGenerator{
			ObjectName:  entity.Name,
			WrapperName: wrapperName,
			Receiver:    "e",
			ObjectType:  ObjectTypeEntity,
		}
		sb.WriteString(generatorToIncrementalProto.GenerateToIncrementalProtoMethod(entity.Fields))

		// 写入文件（追加到 Entity 文件）
		fileName := GetEntityFileName(entity.Name)
		filePath := fmt.Sprintf("%s/%s", outputDir, fileName)

		// 对于第一个实体，使用已更新的内容（包含样板代码）
		if idx == 0 {
			// 确保现有内容以换行结尾
			existingContent = strings.TrimRight(existingContent, " \n\r\t")
			newContent := existingContent + "\n\n" + sb.String()
			if err := WriteFile(filePath, newContent); err != nil {
				return err
			}
		} else {
			// 对于其他实体，读取现有文件内容（如果存在）
			entityExistingContent := ""
			if FileExists(filePath) {
				if data, err := ReadFileContent(filePath); err == nil {
					entityExistingContent = data
				}
			}

			// 追加新内容（如果文件已存在，在末尾添加换行）
			var newContent string
			if entityExistingContent != "" {
				// 确保现有内容以换行结尾
				entityExistingContent = strings.TrimRight(entityExistingContent, " \n\r\t")
				newContent = entityExistingContent + "\n\n" + sb.String()
			} else {
				// 如果文件不存在，需要添加 package 声明和导入
				imports := "import (\n"
				imports += "\t\"gitee.com/orbit-w/orbit/app/proto/mme\"\n"
				imports += "\tdirtyflag \"gitee.com/orbit-w/orbit/lib/base/dirty_flag\"\n"
				imports += "\tfieldmeta \"gitee.com/orbit-w/orbit/lib/base/field_meta\"\n"
				imports += "\t\"gitee.com/orbit-w/orbit/lib/module/db/mgo_builder\"\n"
				imports += "\tmmemodel \"gitee.com/orbit-w/orbit/lib/module/mme_model\"\n"
				imports += "\t\"go.mongodb.org/mongo-driver/v2/bson\"\n"
				imports += "\t\"google.golang.org/protobuf/proto\"\n"
				imports += ")\n\n"
				newContent = fmt.Sprintf("package %s\n\n%s", packageName, imports) + sb.String()
			}

			if err := WriteFile(filePath, newContent); err != nil {
				return err
			}
		}
	}

	return nil
}

// generateEntityBoilerplate 生成实体样板代码（IEntity接口、EntityFactory类型等）
func (g *GoWrapperGenerator) generateEntityBoilerplate() string {
	var sb strings.Builder

	// 生成 IEntity 接口
	sb.WriteString("type IEntity interface {\n")
	sb.WriteString("\tCollection() string\n")
	sb.WriteString("\tHasAnyDirty() bool\n")
	sb.WriteString("\tLoad(raw bson.Raw) error\n")
	sb.WriteString("\tName() string\n")
	sb.WriteString("\tGetXXXId() int64\n")
	sb.WriteString("\tSetXXXId(id int64)\n")
	sb.WriteString("\tGetEntityType() mme.EntityType\n")
	sb.WriteString("\tInitFieldContext()\n")
	sb.WriteString("\tClearAllDirtyFlags()\n")
	sb.WriteString("\tBuildMongoUpdate(builder *mgo_builder.MongoUpdateBuilder, path *mgo_builder.NestedPath)\n")
	sb.WriteString("\tToProto() proto.Message\n")
	sb.WriteString("\tFromProto(msg proto.Message)\n")
	sb.WriteString("\tToIncrementalProtoWithContext(ctx mmemodel.SyncContext) proto.Message\n")
	sb.WriteString("}\n\n")

	// 生成 EntityFactory 类型
	sb.WriteString("type EntityFactory func() IEntity\n\n")

	// 生成 mapEntityFactories 变量
	sb.WriteString("var (\n")
	sb.WriteString("\tmapEntityFactories = make(map[mme.EntityType]EntityFactory)\n")
	sb.WriteString(")\n\n")

	// 生成 init() 函数（包含所有实体的注册）
	sb.WriteString("func init() {\n")
	for _, entity := range g.data.Entities {
		entityTypeEnumName := GenEntityTypeEnumName(entity.Name)
		wrapperName := entity.Name + "Wrapper"
		sb.WriteString(fmt.Sprintf("\tRegisterEntityFactory(mme.EntityType_%s, func() IEntity {\n", entityTypeEnumName))
		sb.WriteString(fmt.Sprintf("\t\treturn New%s()\n", wrapperName))
		sb.WriteString("\t})\n")
	}
	sb.WriteString("}\n\n")

	// 生成 RegisterEntityFactory 函数
	sb.WriteString("func RegisterEntityFactory(entityType mme.EntityType, factory EntityFactory) {\n")
	sb.WriteString("\tmapEntityFactories[entityType] = factory\n")
	sb.WriteString("}\n\n")

	// 生成 GetEntityFactory 函数
	sb.WriteString("func GetEntityFactory(entityType mme.EntityType) EntityFactory {\n")
	sb.WriteString("\tfactory, ok := mapEntityFactories[entityType]\n")
	sb.WriteString("\tif !ok {\n")
	sb.WriteString("\t\treturn nil\n")
	sb.WriteString("\t}\n")
	sb.WriteString("\treturn factory\n")
	sb.WriteString("}\n")

	return sb.String()
}

// updateInitFunction 更新或创建 init() 函数，添加所有实体的注册
func (g *GoWrapperGenerator) updateInitFunction(content string) string {
	// 检查是否已存在 init() 函数
	if strings.Contains(content, "func init()") {
		// 查找现有的 init() 函数
		lines := strings.Split(content, "\n")
		inInit := false
		initStartIdx := -1
		initEndIdx := -1
		braceCount := 0

		for i, line := range lines {
			trimmed := strings.TrimSpace(line)

			if strings.HasPrefix(trimmed, "func init()") {
				inInit = true
				initStartIdx = i
				braceCount = 0
			}

			if inInit {
				// 计算大括号
				braceCount += strings.Count(line, "{")
				braceCount -= strings.Count(line, "}")

				if braceCount == 0 && initStartIdx >= 0 {
					initEndIdx = i
					break
				}
			}
		}

		if initStartIdx >= 0 && initEndIdx >= 0 {
			// 提取现有的 init() 函数内容
			existingInitLines := lines[initStartIdx : initEndIdx+1]
			existingInitContent := strings.Join(existingInitLines, "\n")

			// 检查是否已包含所有实体的注册
			allRegistered := true
			for _, entity := range g.data.Entities {
				entityTypeEnumName := GenEntityTypeEnumName(entity.Name)
				registerPattern := fmt.Sprintf("RegisterEntityFactory(mme.EntityType_%s", entityTypeEnumName)
				if !strings.Contains(existingInitContent, registerPattern) {
					allRegistered = false
					break
				}
			}

			if !allRegistered {
				// 在 init() 函数结束前添加缺失的注册
				var updatedInitLines []string
				updatedInitLines = append(updatedInitLines, existingInitLines[0]) // func init() {

				// 添加现有的注册（如果有）
				for i := 1; i < len(existingInitLines)-1; i++ {
					updatedInitLines = append(updatedInitLines, existingInitLines[i])
				}

				// 添加缺失的注册
				for _, entity := range g.data.Entities {
					entityTypeEnumName := GenEntityTypeEnumName(entity.Name)
					registerPattern := fmt.Sprintf("RegisterEntityFactory(mme.EntityType_%s", entityTypeEnumName)
					if !strings.Contains(existingInitContent, registerPattern) {
						wrapperName := entity.Name + "Wrapper"
						updatedInitLines = append(updatedInitLines,
							fmt.Sprintf("\tRegisterEntityFactory(mme.EntityType_%s, func() IEntity {", entityTypeEnumName))
						updatedInitLines = append(updatedInitLines,
							fmt.Sprintf("\t\treturn New%s()", wrapperName))
						updatedInitLines = append(updatedInitLines, "\t})")
					}
				}

				updatedInitLines = append(updatedInitLines, existingInitLines[len(existingInitLines)-1]) // }

				// 重建内容
				var result []string
				result = append(result, lines[:initStartIdx]...)
				result = append(result, updatedInitLines...)
				result = append(result, lines[initEndIdx+1:]...)
				return strings.Join(result, "\n")
			}
		}
	} else {
		// 如果不存在 init() 函数，在样板代码后添加
		initCode := "\nfunc init() {\n"
		for _, entity := range g.data.Entities {
			entityTypeEnumName := GenEntityTypeEnumName(entity.Name)
			wrapperName := entity.Name + "Wrapper"
			initCode += fmt.Sprintf("\tRegisterEntityFactory(mme.EntityType_%s, func() IEntity {\n", entityTypeEnumName)
			initCode += fmt.Sprintf("\t\treturn New%s()\n", wrapperName)
			initCode += "\t})\n"
		}
		initCode += "}\n"

		// 在 GetEntityFactory 函数后插入
		if idx := strings.Index(content, "func GetEntityFactory"); idx >= 0 {
			// 找到 GetEntityFactory 函数的结束位置
			lines := strings.Split(content[idx:], "\n")
			braceCount := 0
			endIdx := 0
			for i, line := range lines {
				braceCount += strings.Count(line, "{")
				braceCount -= strings.Count(line, "}")
				if braceCount == 0 && i > 0 {
					endIdx = i
					break
				}
			}
			if endIdx > 0 {
				insertPos := idx + len(strings.Join(lines[:endIdx+1], "\n"))
				return content[:insertPos] + "\n" + initCode + content[insertPos:]
			}
		}
	}

	return content
}

// parsedFileContent 表示解析后的文件内容结构
type parsedFileContent struct {
	packageLine    string
	importBlock    []string
	otherContent   []string
	importStartIdx int
	importEndIdx   int
}

// ensureEntityBoilerplate 确保实体文件包含样板代码，如果不存在则生成并插入
func (g *GoWrapperGenerator) ensureEntityBoilerplate(filePath, packageName string) (string, error) {
	existingContent, hasBoilerplate := g.checkBoilerplate(filePath)

	if hasBoilerplate {
		return existingContent, nil
	}

	boilerplate := g.generateEntityBoilerplate()
	requiredImports := g.getRequiredImports()

	var newContent string
	if existingContent != "" {
		// 文件已存在，需要重新组织文件结构
		parsed := g.parseFileContent(existingContent)
		newContent = g.reorganizeFileContent(parsed, packageName, requiredImports, boilerplate)
	} else {
		// 文件不存在，创建新文件并包含样板代码
		newContent = fmt.Sprintf("package %s\n\n%s\n\n%s", packageName, requiredImports, boilerplate)
	}

	// 更新 init() 函数，添加所有实体的注册
	newContent = g.updateInitFunction(newContent)

	// 写入更新的内容回文件
	if err := WriteFile(filePath, newContent); err != nil {
		return "", fmt.Errorf("failed to write boilerplate to %s: %w", filePath, err)
	}

	return newContent, nil
}

// checkBoilerplate 检查文件是否包含样板代码
func (g *GoWrapperGenerator) checkBoilerplate(filePath string) (content string, hasBoilerplate bool) {
	if !FileExists(filePath) {
		return "", false
	}

	data, err := ReadFileContent(filePath)
	if err != nil {
		return "", false
	}

	hasBoilerplate = strings.Contains(data, "type IEntity interface") &&
		strings.Contains(data, "type EntityFactory") &&
		strings.Contains(data, "mapEntityFactories")

	return data, hasBoilerplate
}

// parseFileContent 解析文件内容，提取 package、import 和其他内容
func (g *GoWrapperGenerator) parseFileContent(content string) *parsedFileContent {
	parsed := &parsedFileContent{
		importStartIdx: -1,
		importEndIdx:   -1,
	}

	lines := strings.Split(content, "\n")
	pkgFound := false
	inImport := false
	braceCount := 0

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// 查找 package 声明
		if !pkgFound && strings.HasPrefix(trimmed, "package ") {
			parsed.packageLine = line
			pkgFound = true
			continue
		}

		// 查找 import 块
		if pkgFound && !inImport && strings.HasPrefix(trimmed, "import") {
			inImport = true
			parsed.importStartIdx = i
			parsed.importBlock = append(parsed.importBlock, line)
			if strings.Contains(line, "(") {
				braceCount = 1
			} else {
				// 单行 import
				parsed.importEndIdx = i
				inImport = false
			}
			continue
		}

		if inImport {
			parsed.importBlock = append(parsed.importBlock, line)
			braceCount += strings.Count(line, "(")
			braceCount -= strings.Count(line, ")")
			if braceCount == 0 {
				parsed.importEndIdx = i
				inImport = false
			}
			continue
		}

		// 其他内容（在 package 之后且不在 import 块中）
		if pkgFound && (parsed.importEndIdx < 0 || i > parsed.importEndIdx) {
			// 跳过错误位置的 import 块（在类型声明之后）
			if strings.HasPrefix(trimmed, "import") {
				continue
			}
			parsed.otherContent = append(parsed.otherContent, line)
		}
	}

	return parsed
}

// reorganizeFileContent 重新组织文件内容：package -> import -> 样板代码 -> 其他内容
func (g *GoWrapperGenerator) reorganizeFileContent(parsed *parsedFileContent, packageName, requiredImports, boilerplate string) string {
	var newContent strings.Builder

	// 写入 package 声明
	if parsed.packageLine != "" {
		newContent.WriteString(parsed.packageLine)
		newContent.WriteString("\n\n")
	} else {
		newContent.WriteString("package " + packageName)
		newContent.WriteString("\n\n")
	}

	// 写入 import 块（使用现有的或新的）
	imports := g.selectImportBlock(parsed, requiredImports)
	newContent.WriteString(imports)
	newContent.WriteString("\n\n")

	// 写入样板代码
	newContent.WriteString(boilerplate)
	newContent.WriteString("\n\n")

	// 写入其他内容
	if len(parsed.otherContent) > 0 {
		otherText := strings.Join(parsed.otherContent, "\n")
		otherText = strings.TrimLeft(otherText, " \t\r\n")
		if otherText != "" {
			newContent.WriteString(otherText)
		}
	}

	return newContent.String()
}

// selectImportBlock 选择使用现有的 import 块还是新的 import 块
func (g *GoWrapperGenerator) selectImportBlock(parsed *parsedFileContent, requiredImports string) string {
	hasValidImport := len(parsed.importBlock) > 0 &&
		parsed.importStartIdx >= 0 &&
		parsed.importEndIdx >= 0

	if !hasValidImport {
		return requiredImports
	}

	// 检查 import 块是否在正确位置（紧跟在 package 之后，即索引为 1）
	// 或者 package 和 import 之间只有空行
	isInCorrectPosition := parsed.importStartIdx == 1

	if isInCorrectPosition {
		return strings.Join(parsed.importBlock, "\n")
	}

	// import 块在错误位置，使用新的
	return requiredImports
}

// getRequiredImports 返回必需的 import 语句
func (g *GoWrapperGenerator) getRequiredImports() string {
	return `import (
	"gitee.com/orbit-w/orbit/app/proto/mme"
	dirtyflag "gitee.com/orbit-w/orbit/lib/base/dirty_flag"
	fieldmeta "gitee.com/orbit-w/orbit/lib/base/field_meta"
	"gitee.com/orbit-w/orbit/lib/module/db/mgo_builder"
	mmemodel "gitee.com/orbit-w/orbit/lib/module/mme_model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"google.golang.org/protobuf/proto"
)`
}
