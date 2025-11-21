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

	// 生成 enum.proto（包含所有 NetWall 中的 Enum）
	if err := g.generateEnumProtoFile(outputDir); err != nil {
		return fmt.Errorf("failed to generate enum.proto: %w", err)
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
func (g *ProtoGenerator) generateFieldProto(field *blueprint_types.Field, fieldNumber int32) string {
	builder := NewCodeBuilder()
	builder.SetIndentStr("  ")

	// 添加注释
	if field.Comment != "" {
		builder.WriteLine("// %s", field.Comment)
	}

	// 添加 optional 标记
	protoType := g.typeConverter.ToProtoType(&field.Type)
	isOptional := g.typeConverter.IsOptionalInProto(&field.Type)

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

// generateFieldProtoFromOldField 从 Field 类型生成 Proto 定义（已统一使用 *blueprint_types.Field）
func (g *ProtoGenerator) generateFieldProtoFromOldField(field *blueprint_types.Field, fieldNumber int32) string {
	if field == nil {
		return ""
	}
	return g.generateFieldProto(field, fieldNumber)
}
