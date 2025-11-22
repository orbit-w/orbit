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
	resolver      *TypeReferenceResolver
}

// NewProtoGenerator 创建新的 Proto 生成器
func NewProtoGenerator(data *BlueprintContext) *ProtoGenerator {
	typeConverter := NewTypeConverter("mme")
	return &ProtoGenerator{
		data:          data,
		typeConverter: typeConverter,
		resolver:      NewTypeReferenceResolver(data, typeConverter),
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
	isOptional := g.typeConverter.IsOptionalInProto(&field.Type)

	// 使用统一的类型引用解析器处理跨命名空间引用
	// 对于 MME 包内的类型，不添加包名前缀（因为它们都在同一个包中）
	protoType := g.resolver.ResolveTypeReference(&field.Type, currentPackageName)

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

// resolveEnumTypeReference 解析枚举类型引用（已废弃，使用 TypeReferenceResolver）
// 保留此方法以保持向后兼容，但实际使用 resolver.ResolveTypeReference
func (g *ProtoGenerator) resolveEnumTypeReference(protoType string, fieldType *blueprint_types.FieldType, currentPackageName string) string {
	return g.resolver.ResolveTypeReference(fieldType, currentPackageName)
}

// isEnumType 检查字段类型是否是枚举类型（已废弃，使用 TypeReferenceResolver）
func (g *ProtoGenerator) isEnumType(fieldType *blueprint_types.FieldType) bool {
	return g.resolver.isEnumType(fieldType.Name)
}

// getPackageNameForSource 根据 SourceProto 返回对应的包名（已废弃，使用 TypeReferenceResolver）
func (g *ProtoGenerator) getPackageNameForSource(sourceProto string) string {
	return g.resolver.getPackageNameForSource(sourceProto)
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
