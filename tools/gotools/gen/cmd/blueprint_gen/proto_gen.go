package blueprint_gen

import (
	"fmt"
	"strings"
)

// ProtoGenerator Proto 文件生成器
type ProtoGenerator struct {
	data *BlueprintData
}

// NewProtoGenerator 创建新的 Proto 生成器
func NewProtoGenerator(data *BlueprintData) *ProtoGenerator {
	return &ProtoGenerator{data: data}
}

// Generate 生成所有 Proto 文件
func (g *ProtoGenerator) Generate(outputDir string) error {
	// 生成 common.proto
	if err := g.generateCommonProto(outputDir); err != nil {
		return fmt.Errorf("failed to generate common.proto: %w", err)
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
func generateProtoHeader(packageName string, imports []string) string {
	sb := strings.Builder{}
	sb.WriteString("syntax = \"proto3\";\n\n")
	sb.WriteString(fmt.Sprintf("package %s;\n", packageName))
	sb.WriteString("option go_package = \"./mme\";\n\n")
	
	if len(imports) > 0 {
		sb.WriteString("\n")
		for _, imp := range imports {
			sb.WriteString(fmt.Sprintf("import \"%s\";\n", imp))
		}
		sb.WriteString("\n")
	}
	
	return sb.String()
}

// generateFieldProto 生成字段的 Proto 定义
func generateFieldProto(field Field, fieldNumber int32) string {
	sb := strings.Builder{}
	
	// 添加注释
	if field.Comment != "" {
		sb.WriteString(fmt.Sprintf("  // %s\n", field.Comment))
	}
	
	// 添加 optional 标记（仅对非 map/repeated 的标量字段）
	protoType := ToProtoType(field.Type)
	isOptional := !field.Type.IsMap && !field.Type.IsXMap && !field.Type.IsRepeated && 
		field.Type.BaseType != "" && !strings.Contains(protoType, ".")
	
	if isOptional {
		sb.WriteString("  optional ")
	}
	
	// 确保字段名不包含冒号
	fieldName := strings.TrimSuffix(field.Name, ":")
	fieldName = strings.TrimSpace(fieldName)
	
	// map 字段前面需要加空格
	if field.Type.IsMap || field.Type.IsXMap {
		sb.WriteString(fmt.Sprintf("  %s %s = %d;", protoType, fieldName, fieldNumber))
	} else {
		sb.WriteString(fmt.Sprintf("%s %s = %d;", protoType, fieldName, fieldNumber))
	}
	
	if field.Comment != "" {
		sb.WriteString(fmt.Sprintf("  // %s", field.Comment))
	}
	sb.WriteString("\n")
	
	return sb.String()
}

