package blueprint_gen

import (
	"fmt"
	"strings"
)

// generateCommonProto 生成 common.proto
func (g *ProtoGenerator) generateCommonProto(outputDir string) error {
	sb := strings.Builder{}

	// 文件头部
	sb.WriteString(generateProtoHeader("MME", nil))

	// 生成通用数据结构
	if g.data.HeadFile != nil {
		for _, ds := range g.data.HeadFile.CommonDataStructs {
			sb.WriteString(fmt.Sprintf("// %s\n", ds.Name))
			sb.WriteString(fmt.Sprintf("message %s {\n", ds.Name))

			for i, field := range ds.Fields {
				sb.WriteString(generateFieldProto(field, int32(i+1)))
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
	sb.WriteString(generateProtoHeader("MME", imports))

	// 生成所有 Mechanism
	for _, mech := range g.data.Mechanisms {
		// 生成 message（不添加硬编码的注释）
		sb.WriteString(fmt.Sprintf("message %s {\n", mech.Name))

		// 按编号排序字段
		dataFields := make([]MechanismDataField, len(mech.DataFields))
		copy(dataFields, mech.DataFields)
		// 排序（按 Number）
		for i := 0; i < len(dataFields)-1; i++ {
			for j := i + 1; j < len(dataFields); j++ {
				if dataFields[i].Number > dataFields[j].Number {
					dataFields[i], dataFields[j] = dataFields[j], dataFields[i]
				}
			}
		}

		// 生成数据字段（按编号排序）
		for _, field := range dataFields {
			sb.WriteString(generateFieldProto(field.Field, field.Number))
		}

		// 生成 xmap 的增量同步字段（在 message 内部）
		for _, field := range mech.DataFields {
			if field.Type.IsXMap {
				// 生成 ChangeRecord message（嵌套在父 message 内）
				// 注意：recordName 应该是 "Skills_XXXMapChangeRecord" 而不是 "HeroMechanism_Skills_XXXMapChangeRecord"
				recordName := fmt.Sprintf("%s_XXXMapChangeRecord", field.Name)
				sb.WriteString(fmt.Sprintf("\n  message %s {\n", recordName))
				sb.WriteString(fmt.Sprintf("    %s Key = 1;\n", field.Type.KeyType))
				sb.WriteString(fmt.Sprintf("    %s Value = 2;\n", field.Type.ValueType))
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
	sb.WriteString(generateProtoHeader("MME", imports))

	// 生成所有 Module
	for _, module := range g.data.Modules {
		sb.WriteString(fmt.Sprintf("message %s {\n", module.Name))

		// 生成 Mechanism 引用字段（Module 中的 Mechanism 字段不应该是 optional）
		for _, mechRef := range module.Mechanisms {
			fieldType := fmt.Sprintf("MME.%s", mechRef.MechanismName)
			// 确保编号不为 0
			fieldNum := mechRef.Number
			if fieldNum == 0 {
				fieldNum = 1 // 默认为 1
			}
			// 清理字段名（移除可能的冒号）
			fieldName := strings.TrimSuffix(mechRef.FieldName, ":")
			fieldName = strings.TrimSpace(fieldName)
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
			fieldType := fmt.Sprintf("map<%s, MME.%s>", field.KeyType, field.ModuleName)
			sb.WriteString(fmt.Sprintf("    %s %s = %d;\n",
				fieldType, field.Name, field.Number))

			// 如果是 xmap，生成增量同步字段（在同一行）
			if field.IsXMap {
				// 生成 ChangeRecord message
				// 注意：recordName 应该是 "HeroMap_XXXMapChangeRecord" 而不是 "HeroManager_HeroMap_XXXMapChangeRecord"
				recordName := fmt.Sprintf("%s_XXXMapChangeRecord", field.Name)
				sb.WriteString(fmt.Sprintf("    repeated %s %s_XXXChangeList = %d;  message %s { %s Key = 1; MME.%s Value = 2; bool IsDelete = 3; }\n",
					recordName, field.Name, 1000+int32(field.Number), recordName, field.KeyType, field.ModuleName))
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
			fieldType := fmt.Sprintf("MME.%s", field.ManagerName)
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
