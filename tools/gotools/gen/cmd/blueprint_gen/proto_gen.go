package blueprint_gen

import (
	"fmt"
	"strings"

	blueprint_types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"
)

// ProtoGenerator Proto 文件生成器
type ProtoGenerator struct {
	data          *BlueprintContext
	typeConverter *TypeConverter
}

// NewProtoGenerator 创建新的 Proto 生成器
func NewProtoGenerator(data *BlueprintContext) *ProtoGenerator {
	return &ProtoGenerator{
		data:          data,
		typeConverter: NewTypeConverter("mme"),
	}
}

// Generate 生成所有 Proto 文件
func (g *ProtoGenerator) Generate(outputDir string) error {
	// 生成 headfile.proto
	if err := g.generateHeadfileProto(outputDir); err != nil {
		return fmt.Errorf("failed to generate headfile.proto: %w", err)
	}

	// 生成 mechanisms.proto
	if err := g.generateMechanismsProto(outputDir); err != nil {
		return fmt.Errorf("failed to generate mechanisms.proto: %w", err)
	}

	// 生成 modules.proto
	if err := g.generateModulesProto(outputDir); err != nil {
		return fmt.Errorf("failed to generate modules.proto: %w", err)
	}

	// 生成 managers.proto
	if err := g.generateManagersProto(outputDir); err != nil {
		return fmt.Errorf("failed to generate managers.proto: %w", err)
	}

	// 生成 entities.proto
	if err := g.generateEntitiesProto(outputDir); err != nil {
		return fmt.Errorf("failed to generate entities.proto: %w", err)
	}

	// 生成 NetWall proto 文件
	if err := g.generateNetWallProto(outputDir); err != nil {
		return fmt.Errorf("failed to generate NetWall proto: %w", err)
	}

	return nil
}

// generateProtoHeader 生成 Proto 文件头部
func (g *ProtoGenerator) generateProtoHeader(packageName string, imports []string) string {
	builder := NewCodeBuilder()
	builder.SetIndentStr("")

	builder.WriteLine("syntax = \"proto3\";")
	builder.WriteEmptyLine()
	goPackageName := strings.ToLower(packageName)
	builder.WriteLine("package %s;", packageName)
	builder.WriteLine("option go_package = \"gitee.com/orbit-w/orbit/app/proto/%s\";", goPackageName)
	builder.WriteEmptyLine()

	if len(imports) > 0 {
		imports = UniqueProtoImports(imports)
		for _, imp := range imports {
			builder.WriteLine("import \"%s\";", imp)
		}
		builder.WriteEmptyLine()
	}

	return builder.String()
}

// generateFieldProto 生成字段的 Proto 定义（新版本，使用 *blueprint_types.Field）
// currentPackageName: 当前 proto 文件的包名，用于判断是否需要添加包名前缀
// baseIndent: 基础缩进级别（默认为 1，表示 message 内的字段）
func (g *ProtoGenerator) generateFieldProto(field *blueprint_types.Field, fieldNumber int32, currentPackageName string) string {
	return g.generateFieldProtoWithIndent(field, fieldNumber, currentPackageName, 1)
}

// generateFieldProtoWithIndent 生成字段的 Proto 定义，支持自定义缩进级别
// currentPackageName: 当前 proto 文件的包名，用于判断是否需要添加包名前缀
// indentLevel: 缩进级别（1 = 4个空格，2 = 8个空格，以此类推）
func (g *ProtoGenerator) generateFieldProtoWithIndent(field *blueprint_types.Field, fieldNumber int32, currentPackageName string, indentLevel int) string {
	builder := NewCodeBuilder()
	builder.SetIndentStr("    ") // 使用 4 个空格作为缩进
	for i := 0; i < indentLevel; i++ {
		builder.Indent() // 设置缩进级别
	}

	// 添加注释
	if field.Comment != "" {
		builder.WriteLine("// %s", field.Comment)
	}

	// 获取字段类型
	protoType := g.typeConverter.ToProtoType(&field.Type)
	isOptional := g.typeConverter.IsOptionalInProto(&field.Type)

	// 处理枚举类型：检查类型名称是否对应枚举，并添加包名前缀
	protoType = g.resolveEnumTypeReference(protoType, &field.Type, currentPackageName)

	// 处理消息类型：添加包名前缀（排除枚举类型，因为枚举已经在上面处理了）
	if field.Type.IsMessage() && !g.isEnumType(&field.Type) {
		nameSpaces := g.data.GetNameSpace()
		if protoPackageName, ok := nameSpaces[field.Type.Name]; ok && protoPackageName != currentPackageName {
			protoType = fmt.Sprintf("%s.%s", protoPackageName, field.Type.Name)
		}
	}

	// 确保字段名不包含冒号
	fieldName := strings.TrimSuffix(field.Name, ":")
	fieldName = strings.TrimSpace(fieldName)

	// 构建字段定义
	if isOptional {
		builder.WriteLine("optional %s %s = %d;", protoType, fieldName, fieldNumber)
	} else {
		builder.WriteLine("%s %s = %d;", protoType, fieldName, fieldNumber)
	}

	return builder.String()
}

// resolveEnumTypeReference 解析枚举类型引用，如果类型名称对应枚举，则添加包名前缀
func (g *ProtoGenerator) resolveEnumTypeReference(protoType string, fieldType *blueprint_types.FieldType, currentPackageName string) string {
	// 获取类型名称
	typeName := fieldType.Name
	if typeName == "" {
		typeName = fieldType.TypeName
	}

	// 检查类型名称是否对应枚举
	for _, enum := range g.data.Enums {
		if enum.Name == typeName {
			// 如果枚举不在当前包中，需要添加包名前缀
			if enum.SourceProto != currentPackageName {
				// 根据 SourceProto 确定包名
				packageName := g.getPackageNameForSource(enum.SourceProto)
				if packageName != "" && packageName != currentPackageName {
					// 如果包名是 MME，则不需要前缀（因为都在同一个包中）
					// 否则添加包名前缀
					if packageName != "MME" {
						return fmt.Sprintf("%s.%s", packageName, typeName)
					}
				}
			}
			break
		}
	}

	return protoType
}

// isEnumType 检查字段类型是否是枚举类型
func (g *ProtoGenerator) isEnumType(fieldType *blueprint_types.FieldType) bool {
	// 获取类型名称
	typeName := fieldType.Name
	if typeName == "" {
		typeName = fieldType.TypeName
	}

	// 检查类型名称是否对应枚举
	for _, enum := range g.data.Enums {
		if enum.Name == typeName {
			return true
		}
	}

	return false
}

// getPackageNameForSource 根据 SourceProto 返回对应的包名
func (g *ProtoGenerator) getPackageNameForSource(sourceProto string) string {
	switch sourceProto {
	case "entities", "managers", "modules", "mechanisms", "common":
		return "MME" // 这些都在 MME 包中
	default:
		// NetWall 包名，返回原值（首字母大写）
		if sourceProto != "" {
			return strings.ToUpper(sourceProto[:1]) + sourceProto[1:]
		}
		return sourceProto
	}
}

// generateFieldProtoFromOldField 从 Field 类型生成 Proto 定义（已统一使用 *blueprint_types.Field）
// currentPackageName: 当前 proto 文件的包名，默认为 "MME"
func (g *ProtoGenerator) generateFieldProtoFromOldField(field *blueprint_types.Field, fieldNumber int32, currentPackageName string) string {
	if field == nil {
		return ""
	}
	if currentPackageName == "" {
		currentPackageName = "MME"
	}
	return g.generateFieldProto(field, fieldNumber, currentPackageName)
}
