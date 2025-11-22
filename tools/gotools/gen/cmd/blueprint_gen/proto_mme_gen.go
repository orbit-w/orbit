package blueprint_gen

import (
	"fmt"
	"strings"

	blueprint_types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"
)

// generateHeadfileProto 生成 common.proto
func (g *ProtoGenerator) generateHeadfileProto(outputDir string) error {
	sb := strings.Builder{}

	// 收集 DataStruct 字段中引用的枚举类型，确定需要导入的 proto 文件
	imports := g.collectDataStructImports()

	// 文件头部 - common.proto 使用 MME package
	sb.WriteString(g.generateProtoHeader("MME", imports))

	// 生成通用数据结构（DataStruct）
	if g.data.HeadFile != nil {
		for _, ds := range g.data.HeadFile.CommonDataStructs {
			sb.WriteString(fmt.Sprintf("// %s\n", ds.Name))
			sb.WriteString(fmt.Sprintf("message %s {\n", ds.Name))

			for _, field := range ds.Fields {
				// 使用字段自身的编号（已在 YAML 中定义）
				// 传入 "common" 作为 currentSourceProto，因为 DataStruct 生成到 common.proto
				sb.WriteString(g.generateFieldProtoFromOldField(field, field.Number, "common"))
			}

			sb.WriteString("}\n\n")
		}
	}

	// 生成来自 headfile.yaml 的枚举（SourceProto 为 "common"）
	for _, enum := range g.data.Enums {
		if enum.SourceProto == "common" {
			sb.WriteString(generateEnumProto(enum))
		}
	}

	content := sb.String()
	return WriteFile(outputDir+"/common.proto", content)
}

// collectDataStructImports 收集 DataStruct 字段中引用的类型，返回需要导入的 proto 文件列表
func (g *ProtoGenerator) collectDataStructImports() []string {
	if g.data.HeadFile == nil {
		return nil
	}

	// 收集所有 DataStruct 的字段
	fields := make([]*blueprint_types.Field, 0)
	for _, ds := range g.data.HeadFile.CommonDataStructs {
		fields = append(fields, ds.Fields...)
	}

	return g.collectEnumImportsFromFields(fields, "common")
}

// collectEnumImportsFromFields 从字段列表中收集枚举类型的导入
// currentSourceProto: 当前 proto 文件的 SourceProto（如 "entities", "managers", "common" 等）
func (g *ProtoGenerator) collectEnumImportsFromFields(fields []*blueprint_types.Field, currentSourceProto string) []string {
	imports := make(map[string]bool)

	// 构建枚举名称到枚举的映射，便于快速查找
	enumMap := make(map[string]*Enum)
	for _, enum := range g.data.Enums {
		enumMap[enum.Name] = enum
	}

	// 遍历所有字段
	for _, field := range fields {
		// 检查字段类型是否是枚举（通过类型名称匹配）
		typeName := field.Type.Name
		if typeName == "" {
			typeName = field.Type.TypeName
		}

		// 检查类型名称是否对应一个枚举
		if enum, exists := enumMap[typeName]; exists {
			// 如果枚举不在当前 proto 文件中，需要导入
			if enum.SourceProto != currentSourceProto {
				// 根据 SourceProto 确定导入文件
				importFile := g.getProtoImportForSource(enum.SourceProto)
				if importFile != "" {
					imports[importFile] = true
				}
			}
		}

		// 递归检查 map/xmap/repeated 的 value 类型
		if field.Type.ValueType != nil {
			valueTypeName := field.Type.ValueType.Name
			if valueTypeName == "" {
				valueTypeName = field.Type.ValueType.TypeName
			}
			if enum, exists := enumMap[valueTypeName]; exists {
				if enum.SourceProto != currentSourceProto {
					importFile := g.getProtoImportForSource(enum.SourceProto)
					if importFile != "" {
						imports[importFile] = true
					}
				}
			}
		}
	}

	// 转换为列表并排序
	result := make([]string, 0, len(imports))
	for imp := range imports {
		result = append(result, imp)
	}
	return UniqueProtoImports(result)
}

// getProtoImportForSource 根据 SourceProto 返回对应的 proto 导入文件
func (g *ProtoGenerator) getProtoImportForSource(sourceProto string) string {
	switch sourceProto {
	case "entities":
		return "entities.proto"
	case "managers":
		return "managers.proto"
	case "modules":
		return "modules.proto"
	case "mechanisms":
		return "mechanisms.proto"
	case "common":
		return "common.proto"
	default:
		// NetWall 包名，转换为小写并添加 .proto 后缀
		if sourceProto != "" {
			return strings.ToLower(sourceProto) + ".proto"
		}
		return ""
	}
}

// MMEObjectAutoProtoImport 自动生成 MMEObject 的 Proto 导入
// 根据 MMEObject 的类型和名称，生成对应的 Proto 导入（包括枚举类型）
// currentSourceProto: 当前 proto 文件的 SourceProto
// 返回 Proto 导入列表
func (g *ProtoGenerator) MMEObjectAutoProtoImport(mmeObjects []MMEObjectBase, currentSourceProto string) []string {
	imports := []string{}

	// 收集所有字段
	fields := make([]*blueprint_types.Field, 0)
	for _, mmeObject := range mmeObjects {
		fields = append(fields, mmeObject.GetFields()...)
	}

	// 收集枚举类型的导入
	enumImports := g.collectEnumImportsFromFields(fields, currentSourceProto)
	imports = append(imports, enumImports...)

	// 收集其他类型的导入
	for _, mmeObject := range mmeObjects {
		for _, f := range mmeObject.GetFields() {
			switch {
			case f.Type.IsMessage():
				imports = append(imports, g.genMessageProtoImport(f.GetType()))
			case f.Type.IsMMEObjectType():
				objName := f.Type.GetName()
				if objName != "" {
					imports = append(imports, g.genMMEObjectProtoImport(objName))
				}
			case f.Type.IsXMapField() || f.Type.IsMapField():
				switch {
				case f.Type.ValueType.IsMMEObjectType():
					valueName := f.Type.ValueType.GetName()
					imports = append(imports, g.genMMEObjectProtoImport(valueName))
				case f.Type.ValueType.IsFieldBaseType():
					continue
				case f.Type.ValueType.IsMessage():
					imports = append(imports, g.genMessageProtoImport(f.GetValueType()))
				default:
					// 其他类型（如枚举）已在 collectEnumImportsFromFields 中处理
				}
			}
		}
	}

	// 排重&&排序imports
	return UniqueProtoImports(imports)
}

func (g *ProtoGenerator) genMessageProtoImport(field *blueprint_types.FieldType) string {
	nameSpaces := g.data.GetNameSpace()
	namespace, ok := nameSpaces[field.TypeName]
	if !ok {
		return ""
	}
	return GenerateProtoImport(namespace)
}

func (g *ProtoGenerator) genMMEObjectProtoImport(name string) string {
	objType, ok := g.data.GetObjectType(name)
	if !ok {
		return ""
	}
	switch objType {
	case ObjectTypeEntity:
		return GenerateProtoImport("entities")
	case ObjectTypeManager:
		return GenerateProtoImport("managers")
	case ObjectTypeModule:
		return GenerateProtoImport("modules")
	case ObjectTypeMechanism:
		return GenerateProtoImport("mechanisms")
	default:
		panic(fmt.Sprintf("unknown object type: %s", objType))
	}
}

// generateMechanismsProto 生成 mechanisms.proto
func (g *ProtoGenerator) generateMechanismsProto(outputDir string) error {
	sb := strings.Builder{}

	// 收集所有 Mechanism 的字段，用于收集枚举导入
	allFields := make([]*blueprint_types.Field, 0)
	for _, mech := range g.data.Mechanisms {
		allFields = append(allFields, mech.Fields...)
	}

	// 文件头部
	imports := []string{}
	if g.data.HeadFile != nil && len(g.data.HeadFile.CommonDataStructs) > 0 {
		imports = append(imports, "common.proto")
	}
	// 收集枚举类型的导入
	enumImports := g.collectEnumImportsFromFields(allFields, "mechanisms")
	imports = append(imports, enumImports...)
	sb.WriteString(g.generateProtoHeader("MME", imports))

	// 生成所有 Mechanism
	for _, mech := range g.data.Mechanisms {
		// 生成 message（不添加硬编码的注释）
		sb.WriteString(fmt.Sprintf("message %s {\n", mech.Name))

		// 按编号排序字段
		fields := make([]*blueprint_types.Field, len(mech.Fields))
		copy(fields, mech.Fields)
		// 排序（按 Number）
		for i := 0; i < len(fields)-1; i++ {
			for j := i + 1; j < len(fields); j++ {
				if fields[i].Number > fields[j].Number {
					fields[i], fields[j] = fields[j], fields[i]
				}
			}
		}

		// 生成数据字段（按编号排序）
		for _, field := range fields {
			// 使用通用的字段生成方法，自动处理枚举类型引用
			fieldProto := g.generateFieldProto(field, field.Number, "MME")
			// 将缩进从 2 个空格改为 4 个空格（与 managers.proto 格式一致）
			fieldProto = strings.ReplaceAll(fieldProto, "  ", "    ")
			sb.WriteString(fieldProto)
		}

		// 生成 xmap 的增量同步字段（在 message 内部）
		for _, field := range mech.Fields {
			if field.Type.Kind == blueprint_types.FieldKindXMap {
				// 生成 ChangeRecord message（嵌套在父 message 内）
				// 注意：recordName 应该是 "Skills_XXXMapChangeRecord" 而不是 "HeroMechanism_Skills_XXXMapChangeRecord"
				recordName := fmt.Sprintf("%s_XXXMapChangeRecord", field.Name)

				// 获取 key 类型
				keyType := "unknown"
				if field.Type.KeyType != nil {
					keyType = ToProtoTypeFromTypesFieldType(field.Type.KeyType)
				}

				// 获取 value 类型
				valueType := "unknown"
				if field.Type.ValueType != nil {
					valueType = ToProtoTypeFromTypesFieldType(field.Type.ValueType)
				}

				// 生成 ChangeList 字段和 message 定义（在同一行）
				changeListFieldNumber := 1000 + int32(field.Number)
				sb.WriteString(fmt.Sprintf("    repeated %s %s_XXXChangeList = %d;  message %s { %s Key = 1; %s Value = 2; bool IsDelete = 3; }\n",
					recordName, field.Name, changeListFieldNumber, recordName, keyType, valueType))
			}
		}

		sb.WriteString("}\n\n")
	}

	// 生成来自 mechanisms.yaml 的枚举
	for _, enum := range g.data.Enums {
		if enum.SourceProto == "mechanisms" {
			sb.WriteString(generateEnumProto(enum))
		}
	}

	content := sb.String()
	return WriteFile(outputDir+"/mechanisms.proto", content)
}

// generateModulesProto 生成 modules.proto
func (g *ProtoGenerator) generateModulesProto(outputDir string) error {
	sb := strings.Builder{}

	// 文件头部
	objects := []MMEObjectBase{}
	for _, module := range g.data.Modules {
		objects = append(objects, module.MMEObject)
	}
	imports := g.MMEObjectAutoProtoImport(objects, "modules")
	sb.WriteString(g.generateProtoHeader("MME", imports))

	// 生成所有 Module
	for _, module := range g.data.Modules {
		sb.WriteString(fmt.Sprintf("message %s {\n", module.Name))

		// 生成 Mechanism 引用字段（Module 中的 Mechanism 字段不应该是 optional）
		for _, field := range module.Fields {
			// 从 field.Type.TypeName 获取 Mechanism 名称
			mechanismName := field.Type.TypeName
			if mechanismName == "" {
				continue // 跳过无效字段
			}
			fieldType := fmt.Sprintf("MME.%s", mechanismName)
			// 确保编号不为 0
			fieldNum := field.Number
			if fieldNum == 0 {
				fieldNum = 1 // 默认为 1
			}
			// 使用字段名
			fieldName := strings.TrimSpace(field.Name)
			sb.WriteString(fmt.Sprintf("    %s %s = %d;\n",
				fieldType, fieldName, fieldNum))
		}

		sb.WriteString("}\n\n")
	}

	// 生成来自 modules.yaml 的枚举
	for _, enum := range g.data.Enums {
		if enum.SourceProto == "modules" {
			sb.WriteString(generateEnumProto(enum))
		}
	}

	content := sb.String()
	return WriteFile(outputDir+"/modules.proto", content)
}

// generateManagersProto 生成 managers.proto
func (g *ProtoGenerator) generateManagersProto(outputDir string) error {
	sb := strings.Builder{}

	objects := []MMEObjectBase{}
	for _, manager := range g.data.Managers {
		objects = append(objects, manager.MMEObject)
	}
	imports := g.MMEObjectAutoProtoImport(objects, "managers")
	sb.WriteString(g.generateProtoHeader("MME", imports))

	// 生成所有 Manager
	for _, manager := range g.data.Managers {
		sb.WriteString(fmt.Sprintf("message %s {\n", manager.Name))

		// 生成字段
		for _, field := range manager.Fields {
			// 检查是否是 map 或 xmap 类型
			if field.Type.Kind != blueprint_types.FieldKindMap && field.Type.Kind != blueprint_types.FieldKindXMap {
				continue // 跳过非 map/xmap 类型
			}

			// 获取 key 类型
			keyType := "unknown"
			if field.Type.KeyType != nil {
				if field.Type.KeyType.TypeName != "" {
					keyType = field.Type.KeyType.TypeName
				} else {
					keyType = ToProtoTypeFromTypesFieldType(field.Type.KeyType)
				}
			}

			// 获取 value 类型（Module 名称）
			moduleName := "unknown"
			if field.Type.ValueType != nil {
				moduleName = g.typeConverter.GetMMEObjectName(field.Type.ValueType)
				if moduleName == "" {
					moduleName = g.typeConverter.ToProtoType(field.Type.ValueType)
				}
			}

			// 生成字段定义
			fieldType := fmt.Sprintf("map<%s, MME.%s>", keyType, moduleName)
			sb.WriteString(fmt.Sprintf("    %s %s = %d;\n",
				fieldType, field.Name, field.Number))

			// 如果是 xmap，生成增量同步字段（在同一行）
			if field.Type.Kind == blueprint_types.FieldKindXMap {
				// 生成 ChangeRecord message
				// 注意：recordName 应该是 "HeroMap_XXXMapChangeRecord" 而不是 "HeroManager_HeroMap_XXXMapChangeRecord"
				recordName := fmt.Sprintf("%s_XXXMapChangeRecord", field.Name)
				sb.WriteString(fmt.Sprintf("    repeated %s %s_XXXChangeList = %d;  message %s { %s Key = 1; MME.%s Value = 2; bool IsDelete = 3; }\n",
					recordName, field.Name, 1000+int32(field.Number), recordName, keyType, moduleName))
			}
		}

		sb.WriteString("}\n\n")
	}

	// 生成来自 manager.yaml 的枚举
	for _, enum := range g.data.Enums {
		if enum.SourceProto == "managers" {
			sb.WriteString(generateEnumProto(enum))
		}
	}

	content := sb.String()
	return WriteFile(outputDir+"/managers.proto", content)
}

// generateEntitiesProto 生成 entities.proto
func (g *ProtoGenerator) generateEntitiesProto(outputDir string) error {
	sb := strings.Builder{}

	objects := []MMEObjectBase{}
	for _, entity := range g.data.Entities {
		objects = append(objects, entity.MMEObject)
	}
	imports := g.MMEObjectAutoProtoImport(objects, "entities")
	sb.WriteString(g.generateProtoHeader("MME", imports))

	// 生成所有 Entity
	for _, entity := range g.data.Entities {
		sb.WriteString(fmt.Sprintf("message %s {\n", entity.Name))

		// 生成 Manager 字段
		for _, field := range entity.Fields {
			fieldType := field.Type.TypeName
			// 确保编号不为 0
			fieldNum := field.Number
			if fieldNum == 0 {
				fieldNum = 1 // 默认为 1
			}
			// 清理字段名（移除可能的冒号）
			fieldName := strings.TrimSuffix(field.Name, ":")
			fieldName = strings.TrimSpace(fieldName)
			sb.WriteString(fmt.Sprintf("    %s %s = %d;\n",
				fieldType, fieldName, fieldNum))
		}

		// 自动添加 XXXId 字段（编号 10000）
		sb.WriteString("\n    int64 XXXId = 10000;\n")

		sb.WriteString("}\n\n")
	}

	content := sb.String()
	return WriteFile(outputDir+"/entities.proto", content)
}
