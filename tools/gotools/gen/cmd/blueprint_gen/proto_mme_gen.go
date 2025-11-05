package blueprint_gen

import (
	"fmt"
	"strings"

	blueprint_types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"
)

// generateCommonProto 生成 common.proto
func (g *ProtoGenerator) generateCommonProto(outputDir string) error {
	sb := strings.Builder{}

	// 文件头部
	sb.WriteString(g.generateProtoHeader("MME", nil))

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

	content := sb.String()
	return WriteFile(outputDir+"/common.proto", content)
}

// generateMechanismsProto 生成 mechanisms.proto
func (g *ProtoGenerator) generateMechanismsProto(outputDir string) error {
	sb := strings.Builder{}

	// 文件头部
	imports := []string{}
	if g.data.HeadFile != nil && len(g.data.HeadFile.CommonDataStructs) > 0 {
		imports = append(imports, "protocol/common.proto")
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
			sb.WriteString(g.generateFieldProto(field, field.Number))
		}

		// 生成 xmap 的增量同步字段（在 message 内部）
		for _, field := range mech.Fields {
			if field.Type.Kind == blueprint_types.FieldKindXMap {
				// 生成 ChangeRecord message（嵌套在父 message 内）
				// 注意：recordName 应该是 "Skills_XXXMapChangeRecord" 而不是 "HeroMechanism_Skills_XXXMapChangeRecord"
				recordName := fmt.Sprintf("%s_XXXMapChangeRecord", field.Name)
				sb.WriteString(fmt.Sprintf("\n  message %s {\n", recordName))
				if field.Type.KeyType != nil {
					keyType := ToProtoTypeFromTypesFieldType(field.Type.KeyType)
					sb.WriteString(fmt.Sprintf("    %s Key = 1;\n", keyType))
				}
				if field.Type.ValueType != nil {
					valueType := ToProtoTypeFromTypesFieldType(field.Type.ValueType)
					sb.WriteString(fmt.Sprintf("    %s Value = 2;\n", valueType))
				}
				sb.WriteString("    bool IsDelete = 3;\n")
				sb.WriteString("  }\n")

				// 生成 ChangeList 字段
				changeListFieldNumber := 1000 + int32(field.Number)
				sb.WriteString(fmt.Sprintf("  repeated %s %s_XXXChangeList = %d; // %s变化\n",
					recordName, field.Name, changeListFieldNumber, field.Name))
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
	imports := []string{"protocol/mechanisms.proto"}
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

	// 文件头部（注意 import 在 package 之前）
	sb.WriteString("syntax = \"proto3\";\n\n")
	sb.WriteString("import \"protocol/modules.proto\";\n\n")
	sb.WriteString("option go_package = \"./mme\";\n\n")
	sb.WriteString("package MME;\n\n")

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

	// 文件头部（注意 import 在 package 之后）
	sb.WriteString("syntax = \"proto3\";\n\n")
	sb.WriteString("package MME;\n")
	sb.WriteString("option go_package = \"./mme\";\n\n")
	sb.WriteString("import \"protocol/managers.proto\";\n\n")

	// 生成所有 Entity
	for _, entity := range g.data.Entities {
		sb.WriteString(fmt.Sprintf("message %s {\n", entity.Name))

		// 生成 Manager 字段
		for _, field := range entity.Fields {
			fieldType := fmt.Sprintf("MME.%s", field.Type.TypeName)
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
