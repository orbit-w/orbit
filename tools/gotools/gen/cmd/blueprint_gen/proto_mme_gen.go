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
	// 格式化 proto 内容
	content = FormatProtoContent(content)
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

	// 使用统一的解析器收集所有导入（包括枚举和消息类型）
	// 传入 "common" 作为 currentSourceProto，因为 DataStruct 生成到 common.proto
	return g.resolver.CollectImportsFromFieldsWithSourceProto(fields, "common")
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
	// 使用统一的解析器收集所有导入（包括枚举和消息类型）
	imports := g.resolver.CollectImportsFromFieldsWithSourceProto(allFields, "mechanisms")
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

				// 获取 value 类型，使用解析器处理类型引用
				valueType := "unknown"
				if field.Type.ValueType != nil {
					// 使用解析器解析类型引用，自动处理包名前缀
					valueType = g.resolver.ResolveTypeReference(field.Type.ValueType, "MME")
				}

				// 生成 ChangeList 字段和 message 定义（格式化后分多行）
				changeListFieldNumber := 1000 + int32(field.Number)
				sb.WriteString(fmt.Sprintf("    repeated %s %s_XXXChangeList = %d;\n", recordName, field.Name, changeListFieldNumber))
				sb.WriteString(fmt.Sprintf("    message %s {\n", recordName))
				sb.WriteString(fmt.Sprintf("        %s Key = 1;\n", keyType))
				sb.WriteString(fmt.Sprintf("        %s Value = 2;\n", valueType))
				sb.WriteString("        bool IsDelete = 3;\n")
				sb.WriteString("    }\n")
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
	// 格式化 proto 内容
	content = FormatProtoContent(content)
	return WriteFile(outputDir+"/mechanisms.proto", content)
}

// generateModulesProto 生成 modules.proto
func (g *ProtoGenerator) generateModulesProto(outputDir string) error {
	sb := strings.Builder{}

	// 文件头部
	// 收集所有 Module 的字段
	allFields := make([]*blueprint_types.Field, 0)
	for _, module := range g.data.Modules {
		allFields = append(allFields, module.Fields...)
	}
	// 使用统一的解析器收集所有导入
	imports := g.resolver.CollectImportsFromFieldsWithSourceProto(allFields, "modules")
	sb.WriteString(g.generateProtoHeader("MME", imports))

	// 生成所有 Module
	for _, module := range g.data.Modules {
		sb.WriteString(fmt.Sprintf("message %s {\n", module.Name))

		// 生成 Mechanism 引用字段（Module 中的 Mechanism 字段不应该是 optional）
		for _, field := range module.Fields {
			// 使用通用的字段生成方法，自动处理类型引用和导入
			fieldProto := g.generateFieldProto(field, field.Number, "MME")
			sb.WriteString(fieldProto)
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
	// 格式化 proto 内容
	content = FormatProtoContent(content)
	return WriteFile(outputDir+"/modules.proto", content)
}

// generateManagersProto 生成 managers.proto
func (g *ProtoGenerator) generateManagersProto(outputDir string) error {
	sb := strings.Builder{}

	// 收集所有 Manager 的字段
	allFields := make([]*blueprint_types.Field, 0)
	for _, manager := range g.data.Managers {
		allFields = append(allFields, manager.Fields...)
	}
	// 使用统一的解析器收集所有导入
	imports := g.resolver.CollectImportsFromFieldsWithSourceProto(allFields, "managers")
	sb.WriteString(g.generateProtoHeader("MME", imports))

	// 生成所有 Manager
	for _, manager := range g.data.Managers {
		sb.WriteString(fmt.Sprintf("message %s {\n", manager.Name))

		// 生成字段
		for _, field := range manager.Fields {
			switch field.Type.Kind {
			case blueprint_types.FieldKindMap:
				g.generateMessageByMapOrXMapField(field, &sb)
			case blueprint_types.FieldKindXMap:
				g.generateMessageByMapOrXMapField(field, &sb)
			case blueprint_types.FieldKindRepeated:
				panic(fmt.Sprintf("manager %s 的字段 %s 类型不能是 repeated", manager.Name, field.Name))
			default:
				g.generateMessageByField(field, &sb)
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
	// 格式化 proto 内容
	content = FormatProtoContent(content)
	return WriteFile(outputDir+"/managers.proto", content)
}

func (g *ProtoGenerator) generateMessageByField(field *blueprint_types.Field, sb *strings.Builder) {
	fmt.Fprintf(sb, "    %s %s = %d;\n",
		field.Type.GetName(), field.Name, field.Number)
}

func (g *ProtoGenerator) generateMessageByMapOrXMapField(field *blueprint_types.Field, sb *strings.Builder) {
	// 获取 key 类型
	keyType := "unknown"
	if field.Type.KeyType != nil {
		if field.Type.KeyType.TypeName != "" {
			keyType = field.Type.KeyType.TypeName
		} else {
			keyType = ToProtoTypeFromTypesFieldType(field.Type.KeyType)
		}
	}

	// 获取 value 类型（Module 名称），使用解析器处理类型引用
	moduleName := "unknown"
	if field.Type.ValueType != nil {
		// 使用解析器解析类型引用，自动处理包名前缀
		moduleName = g.resolver.ResolveTypeReference(field.Type.ValueType, "MME")
	}

	// 生成字段定义
	fieldType := fmt.Sprintf("map<%s, %s>", keyType, moduleName)
	sb.WriteString(fmt.Sprintf("    %s %s = %d;\n",
		fieldType, field.Name, field.Number))

	// 如果是 xmap，生成增量同步字段（格式化后分多行）
	if field.Type.Kind == blueprint_types.FieldKindXMap {
		// 生成 ChangeRecord message
		// 注意：recordName 应该是 "HeroMap_XXXMapChangeRecord" 而不是 "HeroManager_HeroMap_XXXMapChangeRecord"
		recordName := fmt.Sprintf("%s_XXXMapChangeRecord", field.Name)
		changeListFieldNumber := 1000 + int32(field.Number)
		sb.WriteString(fmt.Sprintf("    repeated %s %s_XXXChangeList = %d;\n", recordName, field.Name, changeListFieldNumber))
		sb.WriteString(fmt.Sprintf("    message %s {\n", recordName))
		sb.WriteString(fmt.Sprintf("        %s Key = 1;\n", keyType))
		sb.WriteString(fmt.Sprintf("        %s Value = 2;\n", moduleName))
		sb.WriteString("        bool IsDelete = 3;\n")
		sb.WriteString("    }\n")
	}
}

// generateEntitiesProto 生成 entities.proto
func (g *ProtoGenerator) generateEntitiesProto(outputDir string) error {
	sb := strings.Builder{}

	// 收集所有 Entity 的字段
	allFields := make([]*blueprint_types.Field, 0)
	for _, entity := range g.data.Entities {
		allFields = append(allFields, entity.Fields...)
	}
	// 使用统一的解析器收集所有导入
	imports := g.resolver.CollectImportsFromFieldsWithSourceProto(allFields, "entities")
	sb.WriteString(g.generateProtoHeader("MME", imports))

	// 生成所有 Entity
	for _, entity := range g.data.Entities {
		sb.WriteString(fmt.Sprintf("message %s {\n", entity.Name))

		// 生成 Manager 字段
		for _, field := range entity.Fields {
			// 使用通用的字段生成方法，自动处理类型引用和导入
			fieldProto := g.generateFieldProto(field, field.Number, "MME")
			sb.WriteString(fieldProto)
		}

		// 自动添加 XXXId 字段（编号 10000）
		sb.WriteString("\n    int64 XXXId = 10000;\n")

		sb.WriteString("}\n\n")
	}

	content := sb.String()
	// 格式化 proto 内容
	content = FormatProtoContent(content)
	return WriteFile(outputDir+"/entities.proto", content)
}
