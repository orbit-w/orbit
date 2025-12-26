package blueprint_gen

import (
	"fmt"
	"os"
	"strings"

	"gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/mmeobj"
	types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"
	mmeobject "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types/mme_obj"
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

		// 如何是Entity，生成对应的Init方法

		// 生成 BuildMongoUpdate 方法
		generator := &BuildMongoUpdateCodeGenerator{
			ObjectName:  mech.Name,
			WrapperName: wrapperName,
			Receiver:    "w",
			ObjectType:  mmeobject.ObjectTypeMechanism,
		}
		sb.WriteString(generator.GenerateBuildMongoUpdateMethod(mech.Fields))

		generatorToIncrementalProto := &ToIncrementalProtoCodeGenerator{
			ObjectName:  mech.Name,
			WrapperName: wrapperName,
			Receiver:    "w",
			ObjectType:  mmeobject.ObjectTypeMechanism,
		}
		sb.WriteString(generatorToIncrementalProto.GenerateToIncrementalProtoMethod(mech.Fields))

		wrapperMethodCodeGenerator := &WrapperMethodCodeGenerator{
			ObjectName:  mech.Name,
			WrapperName: wrapperName,
			Receiver:    "w",
			ProtoPkg:    "mme",
			ObjectType:  mmeobject.ObjectTypeMechanism,
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
			ObjectType:  mmeobject.ObjectTypeModule,
		}
		sb.WriteString(generator.GenerateBuildMongoUpdateMethod(module.Fields))

		wrapperMethodCodeGenerator := &WrapperMethodCodeGenerator{
			ObjectName:  module.Name,
			WrapperName: wrapperName,
			Receiver:    "w",
			ProtoPkg:    "mme",
			ObjectType:  mmeobject.ObjectTypeModule,
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
			ObjectType:  mmeobject.ObjectTypeModule,
		}
		sb.WriteString(generatorToIncrementalProto.GenerateToIncrementalProtoMethod(module.Fields))

		// 生成 Location 方法
		sb.WriteString(g.generateModulesLocationMethod(module, wrapperName))

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
			imports += "\t\"errors\"\n\n"
			imports += "\tdirtyflag \"gitee.com/orbit-w/orbit/lib/base/dirty_flag\"\n"
			imports += "\tfieldmeta \"gitee.com/orbit-w/orbit/lib/base/field_meta\"\n"
			imports += "\t\"gitee.com/orbit-w/orbit/lib/module/db/mgo_builder\"\n"
			imports += "\tmmemodel \"gitee.com/orbit-w/orbit/lib/module/mme_model\"\n"
			imports += "\t\"gitee.com/orbit-w/orbit/pkg/proto/mme\"\n"
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

// generateLocationMethod 生成 Location 方法
// Location 方法根据 MMELocation 查找对应的 Mechanism Wrapper 和 MechanismType
func (g *GoWrapperGenerator) generateModulesLocationMethod(module *mmeobj.Module, wrapperName string) string {
	sb := strings.Builder{}

	// 只有包含 MMEObject 字段的 Module 才生成 Location 方法
	hasMMEObjectField := false
	for _, field := range module.Fields {
		if field.Type.IsMMEObjectType() {
			hasMMEObjectField = true
			break
		}
	}

	if !hasMMEObjectField {
		return ""
	}

	// 生成方法签名和开头
	sb.WriteString("// Location 根据位置信息查找对应的 Mechanism Wrapper 和 MechanismType\n")
	sb.WriteString(fmt.Sprintf("func (w *%s) Location(loc *mme.MMELocation) (any, mme.MechanismType) {\n", wrapperName))
	sb.WriteString("\tif loc == nil {\n")
	sb.WriteString("\t\treturn nil, mme.MechanismType_Unknown\n")
	sb.WriteString("\t}\n\n")
	sb.WriteString("\tindex := loc.GetModuleIndex()\n")
	sb.WriteString("\tswitch index {\n")

	// 生成 case 分支
	for _, field := range module.Fields {
		if field.Type.IsMMEObjectType() {
			fieldIndexName := fmt.Sprintf("%sFieldIndex%s", module.Name, field.Name)
			wrapperFieldName := field.Name + "Wrapper"
			typeName := field.GetTypeName()

			// 生成 MechanismType 枚举值名称
			mechanismTypeEnumName := GenMechanismTypeEnumName(typeName)

			sb.WriteString(fmt.Sprintf("\tcase int32(%s):\n", fieldIndexName))
			sb.WriteString(fmt.Sprintf("\t\treturn w.%s, mme.MechanismType_%s\n",
				wrapperFieldName, mechanismTypeEnumName))
		}
	}

	// 生成默认分支
	sb.WriteString("\t}\n")
	sb.WriteString("\treturn nil, mme.MechanismType_Unknown\n")
	sb.WriteString("}\n\n")

	return sb.String()
}

func (g *GoWrapperGenerator) generateMMEGOStructWrapper(objName string, _ mmeobject.ObjectType, fields []*types.Field, sb *strings.Builder) {
	// 生成 Wrapper 结构体
	wrapperName := objName + "Wrapper"
	sb.WriteString(fmt.Sprintf("type %s struct {\n", wrapperName))
	sb.WriteString(fmt.Sprintf("\tdata *%s\n", objName))
	sb.WriteString("\tdirtyflag.IDirtyFlag\n")
	sb.WriteString("\tfieldMetas *fieldmeta.FieldMetas\n\n")

	for _, field := range fields {
		switch {
		case field.Type.IsXMapField() || field.Type.IsMapField():
			if field.Type.IsXMapField() {
				if !field.Type.IsXMapValueMMEObject() {
					panic(fmt.Sprintf("%s 的xmap字段 %s 的Value类型必须是 MMEObject 类型", objName, field.Name))
				}
			} else if field.Type.IsMapField() {
				if !field.Type.IsMapValueMMEObject() {
					panic(fmt.Sprintf("%s 的map字段 %s 的Value类型必须是 MMEObject 类型", objName, field.Name))
				}
			}

			keyType := field.Type.KeyKind().String()
			valueType := field.GetValueName()
			wrapperTypeName := valueType + "Wrapper"
			linkFieldName := strings.ToLower(field.Name[0:1]) + field.Name[1:] + "Link"
			sb.WriteString("\t//Value为MME Object，使用xmap.MapAccessor进行包装\n")
			sb.WriteString(fmt.Sprintf("\t%s *xmapwrapper.XMapWrapper[%s, *%s, *%s]\n",
				linkFieldName, keyType, valueType, wrapperTypeName))
		case field.Type.IsMMEObjectType():
			wrapperTypeName := field.GetTypeName() + "Wrapper"
			sb.WriteString(fmt.Sprintf("\t%sWrapper *%s\n", field.Name, wrapperTypeName))
		default:
			panic(fmt.Sprintf("field %s type is not supported", field.Name))
		}

	}

	sb.WriteString("}\n\n")

}

func (g *GoWrapperGenerator) generateMMEGOStructWrapperNewConstructor(objName string, _ mmeobject.ObjectType, fields []*types.Field, sb *strings.Builder) {
	// 生成 New 构造函数
	wrapperName := objName + "Wrapper"
	sb.WriteString(fmt.Sprintf("func New%s(data *%s) *%s {\n", wrapperName, objName, wrapperName))
	sb.WriteString("\tif data == nil {\n")
	sb.WriteString("\t\tpanic(\"data is nil\")\n")
	sb.WriteString("\t}\n")
	sb.WriteString(fmt.Sprintf("\tm := &%s{\n", wrapperName))
	sb.WriteString("\t\tdata:       data,\n")
	sb.WriteString("\t\tIDirtyFlag: dirtyflag.NewDirtyFlag(),\n")
	sb.WriteString("\t\tfieldMetas: fieldmeta.NewFieldMetas(),\n")
	sb.WriteString("\t}\n\n")

	// 初始化 XMapWrapper
	for _, field := range fields {
		switch {
		case field.Type.IsXMapField() || field.Type.IsMapField():
			valueName := field.GetValueName()
			linkFieldName := strings.ToLower(field.Name[0:1]) + field.Name[1:] + "Link"
			dirtyBitName := fmt.Sprintf("%sDirty%sBit", objName, field.Name)
			newWrapperFuncName := fmt.Sprintf("New%sWrapper", valueName)
			sb.WriteString(fmt.Sprintf("\tm.%s = xmapwrapper.NewXMapWrapperWithParent(\n",
				linkFieldName))
			sb.WriteString(fmt.Sprintf("\t\t&data.%s,\n", field.Name))
			sb.WriteString("\t\tm.GetDirtyTracker(),\n")
			sb.WriteString(fmt.Sprintf("\t\t%s,\n", dirtyBitName))
			sb.WriteString(fmt.Sprintf("\t\t%s,\n", newWrapperFuncName))
			sb.WriteString("\t)\n\n")
		case field.Type.IsMMEObjectType():
			// 当 data = nil 时，需要初始化
			sb.WriteString(fmt.Sprintf("\tif data.%s == nil {\n", field.Name))
			sb.WriteString(fmt.Sprintf("\t\tdata.%s = New%s()\n", field.Name, field.GetTypeName()))
			sb.WriteString("\t}\n")
			dirtyBitName := fmt.Sprintf("%sDirty%sBit", objName, field.Name)
			wrapperTypeName := field.GetTypeName() + "Wrapper"
			sb.WriteString(fmt.Sprintf("\tm.%sWrapper = New%s(data.%s)\n", field.Name, wrapperTypeName, field.Name))
			sb.WriteString(fmt.Sprintf("\tm.%sWrapper.Link(m.GetDirtyTracker(), %s)\n", field.Name, dirtyBitName))
		default:
			panic(fmt.Sprintf("field %s type is not supported", field.Name))
		}
	}

	sb.WriteString("\treturn m\n")
	sb.WriteString("}\n\n")
}

// generateManagerWrappers 生成 Manager Wrapper
func (g *GoWrapperGenerator) generateManagerWrappers(outputDir string) error {
	for _, manager := range g.data.Managers {
		sb := strings.Builder{}

		// 生成 Wrapper 结构体
		wrapperName := manager.Name + "Wrapper"
		g.generateMMEGOStructWrapper(manager.Name, mmeobject.ObjectTypeManager, manager.Fields, &sb)

		// 生成 New 构造函数
		g.generateMMEGOStructWrapperNewConstructor(manager.Name, mmeobject.ObjectTypeManager, manager.Fields, &sb)

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
			ObjectType:  mmeobject.ObjectTypeManager,
		}
		sb.WriteString(generator.GenerateBuildMongoUpdateMethod(manager.Fields))

		wrapperMethodCodeGenerator := &WrapperMethodCodeGenerator{
			ObjectName:  manager.Name,
			WrapperName: wrapperName,
			Receiver:    "m",
			ProtoPkg:    "mme",
			ObjectType:  mmeobject.ObjectTypeManager,
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
			ObjectType:  mmeobject.ObjectTypeManager,
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

// generateEntityInitFunction 生成 Entity 的 init 函数
func generateEntityInitFunction(entityName string, wrapperName string) string {
	entityTypeEnumName := GenEntityTypeEnumName(entityName)
	return fmt.Sprintf("func init() {\n\tRegisterEntityFactory(mme.EntityType_%s, func() IEntity {\n\t\treturn New%s()\n\t})\n}\n\n", entityTypeEnumName, wrapperName)
}

// generateEntityWrappers 生成 Entity Wrapper
func (g *GoWrapperGenerator) generateEntityWrappers(outputDir string) error {
	if len(g.data.Entities) == 0 {
		return nil
	}

	packageName := "mme"

	// 生成所有实体的 Wrapper 代码
	for _, entity := range g.data.Entities {
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
			ObjectType:  mmeobject.ObjectTypeEntity,
		}
		sb.WriteString(generator.GenerateBuildMongoUpdateMethod(entity.Fields))

		wrapperMethodCodeGenerator := &WrapperMethodCodeGenerator{
			ObjectName:  entity.Name,
			WrapperName: wrapperName,
			Receiver:    "e",
			ProtoPkg:    "mme",
			ObjectType:  mmeobject.ObjectTypeEntity,
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
			ObjectType:  mmeobject.ObjectTypeEntity,
		}
		sb.WriteString(generatorToIncrementalProto.GenerateToIncrementalProtoMethod(entity.Fields))

		// 写入文件（追加到 Entity 文件）
		fileName := GetEntityFileName(entity.Name)
		filePath := fmt.Sprintf("%s/%s", outputDir, fileName)

		// 生成 init 函数（wrapperName 已在上面声明）
		initFunction := generateEntityInitFunction(entity.Name, wrapperName)

		// 读取现有文件内容（如果存在）
		entityExistingContent := ""
		if FileExists(filePath) {
			if data, err := ReadFileContent(filePath); err == nil {
				entityExistingContent = data
			}
		}

		// 追加新内容（如果文件已存在，在末尾添加换行）
		var newContent string
		if entityExistingContent != "" {
			// 检查是否已经有 init 函数
			hasInitFunction := strings.Contains(entityExistingContent, "func init()")

			// 确保现有内容以换行结尾
			entityExistingContent = strings.TrimRight(entityExistingContent, " \n\r\t")

			if !hasInitFunction {
				// 如果没有 init 函数，需要在 import 之后插入
				// 查找 import 块的结束位置
				importEndIndex := strings.LastIndex(entityExistingContent, ")\n\n")
				if importEndIndex == -1 {
					// 如果没有找到 import 块，在 package 之后插入
					packageEndIndex := strings.Index(entityExistingContent, "\n\n")
					if packageEndIndex != -1 {
						beforeInit := entityExistingContent[:packageEndIndex+2]
						afterInit := entityExistingContent[packageEndIndex+2:]
						newContent = beforeInit + initFunction + afterInit + "\n\n" + sb.String()
					} else {
						newContent = entityExistingContent + "\n\n" + initFunction + sb.String()
					}
				} else {
					// 在 import 块之后插入 init 函数
					beforeInit := entityExistingContent[:importEndIndex+3]
					afterInit := entityExistingContent[importEndIndex+3:]
					newContent = beforeInit + initFunction + afterInit + "\n\n" + sb.String()
				}
			} else {
				// 如果已经有 init 函数，直接追加 Wrapper 代码
				newContent = entityExistingContent + "\n\n" + sb.String()
			}
		} else {
			// 如果文件不存在，需要添加 package 声明和导入
			imports := "import (\n"
			imports += "\t\"gitee.com/orbit-w/orbit/pkg/proto/mme\"\n"
			imports += "\tdirtyflag \"gitee.com/orbit-w/orbit/lib/base/dirty_flag\"\n"
			imports += "\tfieldmeta \"gitee.com/orbit-w/orbit/lib/base/field_meta\"\n"
			imports += "\t\"gitee.com/orbit-w/orbit/lib/module/db/mgo_builder\"\n"
			imports += "\tmmemodel \"gitee.com/orbit-w/orbit/lib/module/mme_model\"\n"
			imports += "\t\"go.mongodb.org/mongo-driver/v2/bson\"\n"
			imports += "\t\"google.golang.org/protobuf/proto\"\n"
			imports += ")\n\n"
			newContent = fmt.Sprintf("package %s\n\n%s%s", packageName, imports, initFunction) + sb.String()
		}

		if err := WriteFile(filePath, newContent); err != nil {
			return err
		}
	}

	return nil
}
