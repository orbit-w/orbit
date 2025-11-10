package blueprint_gen

import (
	"fmt"
	"regexp"
	"strings"

	blueprint_types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"
)

// generateHeadfileProto 生成 common.proto
func (g *ProtoGenerator) generateHeadfileProto(outputDir string) error {
	sb := strings.Builder{}

	// 文件头部 - headfile.proto 使用 Common package
	sb.WriteString(g.generateProtoHeader("Common", nil))

	// 生成通用数据结构
	if g.data.HeadFile != nil {
		for _, ds := range g.data.HeadFile.CommonDataStructs {
			sb.WriteString(fmt.Sprintf("// %s\n", ds.Name))
			sb.WriteString(fmt.Sprintf("message %s {\n", ds.Name))

			for i, field := range ds.Fields {
				sb.WriteString(g.generateFieldProtoFromOldField(field, int32(i+1)))
			}

			sb.WriteString("}\n\n")
		}
	}

	// 生成 Entity enum
	// 注意：即使没有通用数据结构，也要生成 Entity enum
	if len(g.data.Entities) > 0 {
		sb.WriteString("// EntityType 实体类型枚举\n")
		sb.WriteString("enum EntityType {\n")
		sb.WriteString("    ENTITY_TYPE_UNSPECIFIED = 0;\n")

		for i, entity := range g.data.Entities {
			enumValueName := g.entityNameToEnumValue(entity.Name)
			sb.WriteString(fmt.Sprintf("    %s = %d;\n", enumValueName, i+1))
		}

		sb.WriteString("}\n\n")
	} else {
		// 如果没有 Entity，至少生成空的 enum 定义（可选）
		// 这里不生成，因为用户要求针对每种 Entity 对象生成 enum
	}

	content := sb.String()
	return WriteFile(outputDir+"/common.proto", content)
}

// entityNameToEnumValue 将 Entity 名称转换为 enum 值名称
// 例如: PlayerEntity -> ENTITY_TYPE_PLAYER_ENTITY
func (g *ProtoGenerator) entityNameToEnumValue(entityName string) string {
	// 使用完整的 Entity 名称
	name := entityName

	// 在驼峰命名的大写字母前插入下划线
	re := regexp.MustCompile(`([a-z])([A-Z])`)
	name = re.ReplaceAllString(name, `${1}_${2}`)

	// 全部转换为大写
	name = strings.ToUpper(name)

	// 添加前缀
	return "ENTITY_TYPE_" + name
}

// MMEObjectAutoProtoImport 自动生成 MMEObject 的 Proto 导入
// 根据 MMEObject 的类型和名称，生成对应的 Proto 导入
// 返回 Proto 导入列表
func (g *ProtoGenerator) MMEObjectAutoProtoImport(mmeObjects []MMEObjectBase) []string {
	imports := []string{}
	for _, mmeObject := range mmeObjects {
		for _, f := range mmeObject.GetFields() {
			switch {
			case f.Type.IsMessage():
				imports = append(imports, g.genMessageProtoImport(f.GetType()))
			case f.Type.IsMMEObjectType():
				objName := f.Type.GetName()
				fmt.Println("objName", objName)
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
					panic(fmt.Sprintf("unknown field type: %s", f.Type.ValueType.GetName()))
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

	// 文件头部
	imports := []string{}
	if g.data.HeadFile != nil && len(g.data.HeadFile.CommonDataStructs) > 0 {
		imports = append(imports, "common.proto")
	}
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
			// 获取字段类型
			protoType := g.typeConverter.ToProtoType(&field.Type)
			isOptional := g.typeConverter.IsOptionalInProto(&field.Type)

			// 确保字段名不包含冒号
			fieldName := strings.TrimSuffix(field.Name, ":")
			fieldName = strings.TrimSpace(fieldName)

			// 生成字段定义（使用4个空格缩进，与 managers.proto 格式一致）
			if isOptional {
				sb.WriteString(fmt.Sprintf("    optional %s %s = %d;\n", protoType, fieldName, field.Number))
			} else {
				sb.WriteString(fmt.Sprintf("    %s %s = %d;\n", protoType, fieldName, field.Number))
			}
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
	imports := g.MMEObjectAutoProtoImport(objects)
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
	imports := g.MMEObjectAutoProtoImport(objects)
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
	imports := g.MMEObjectAutoProtoImport(objects)
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
