package blueprint_gen

import (
	"fmt"
	"strings"
)

// GoStructGenerator Go 结构体生成器
type GoStructGenerator struct {
	data *BlueprintData
}

// NewGoStructGenerator 创建新的 Go 结构体生成器
func NewGoStructGenerator(data *BlueprintData) *GoStructGenerator {
	return &GoStructGenerator{data: data}
}

// Generate 生成所有 Go 文件
func (g *GoStructGenerator) Generate(outputDir string) error {
	// 生成 Mechanism Go 文件
	if err := g.generateMechanismFiles(outputDir); err != nil {
		return fmt.Errorf("failed to generate mechanism files: %w", err)
	}
	
	// 生成 Module Go 文件
	if err := g.generateModuleFiles(outputDir); err != nil {
		return fmt.Errorf("failed to generate module files: %w", err)
	}
	
	// 生成 Manager Go 文件
	if err := g.generateManagerFiles(outputDir); err != nil {
		return fmt.Errorf("failed to generate manager files: %w", err)
	}
	
	// 生成 Entity Go 文件
	if err := g.generateEntityFiles(outputDir); err != nil {
		return fmt.Errorf("failed to generate entity files: %w", err)
	}
	
	// 生成 Wrapper 文件
	wrapperGen := NewGoWrapperGenerator(g.data)
	if err := wrapperGen.Generate(outputDir); err != nil {
		return fmt.Errorf("failed to generate wrapper files: %w", err)
	}
	
	return nil
}

// generateMechanismFiles 生成 Mechanism Go 文件
func (g *GoStructGenerator) generateMechanismFiles(outputDir string) error {
	packageName := "mme"
	
	for _, mech := range g.data.Mechanisms {
		sb := strings.Builder{}
		
		// 文件头部
		sb.WriteString(fmt.Sprintf("package %s\n\n", packageName))
		sb.WriteString("import (\n")
		sb.WriteString("\t\"gitee.com/orbit-w/orbit/app/proto/mme\"\n")
		sb.WriteString("\tdirtyflag \"gitee.com/orbit-w/orbit/lib/base/dirty_flag\"\n")
		sb.WriteString("\tfieldmeta \"gitee.com/orbit-w/orbit/lib/base/field_meta\"\n")
		sb.WriteString("\t\"gitee.com/orbit-w/orbit/lib/module/db/mgo_builder\"\n")
		sb.WriteString("\tmmemodel \"gitee.com/orbit-w/orbit/lib/module/mme_model\"\n")
		
		// 检查是否有 map 字段需要 xmap
		hasMap := false
		for _, field := range mech.DataFields {
			if field.Type.IsXMap || field.Type.IsMap {
				hasMap = true
				break
			}
		}
		if hasMap {
			sb.WriteString("\t\"gitee.com/orbit-w/meteor/bases/container/xmap\"\n")
			sb.WriteString("\txmapwrapper \"gitee.com/orbit-w/orbit/lib/module/xmapwrapper\"\n")
		}
		
		sb.WriteString("\t\"google.golang.org/protobuf/proto\"\n")
		sb.WriteString(")\n\n")
		
		// 生成 FieldIndex 常量
		sb.WriteString("const (\n")
		for idx, field := range mech.DataFields {
			fieldIndexName := fmt.Sprintf("%sFieldIndex%s", mech.Name, field.Name)
			sb.WriteString(fmt.Sprintf("\t%s = uint8(%d)\n", fieldIndexName, idx))
		}
		sb.WriteString(")\n\n")
		
		// 生成 DirtyBit 常量
		sb.WriteString("// Dirty bits for Mechanism fields\n")
		sb.WriteString("const (\n")
		for _, field := range mech.DataFields {
			dirtyBitName := fmt.Sprintf("%sDirty%sBit", mech.Name, field.Name)
			sb.WriteString(fmt.Sprintf("\t%s int64 = 1 << %sFieldIndex%s\n",
				dirtyBitName, mech.Name, field.Name))
		}
		sb.WriteString(")\n\n")
		
		// 生成结构体
		sb.WriteString(fmt.Sprintf("type %s struct {\n", mech.Name))
		for _, field := range mech.DataFields {
			goType := ToGoType(field.Type, packageName)
			sb.WriteString(fmt.Sprintf("\t%s %s\n", field.Name, goType))
		}
		sb.WriteString("}\n\n")
		
		// 生成 New 构造函数
		sb.WriteString(fmt.Sprintf("func New%s() *%s {\n", mech.Name, mech.Name))
		sb.WriteString(fmt.Sprintf("\treturn &%s{}\n", mech.Name))
		sb.WriteString("}\n\n")
		
		// 生成 DeepCopy 方法
		sb.WriteString(fmt.Sprintf("func (m *%s) DeepCopy(co *%s) {\n", mech.Name, mech.Name))
		sb.WriteString("\tif m == nil || co == nil {\n")
		sb.WriteString("\t\treturn\n")
		sb.WriteString("\t}\n\n")
		sb.WriteString("\t*co = *m\n")
		sb.WriteString("}\n\n")
		
		// 生成 ToProto 方法
		sb.WriteString(fmt.Sprintf("func (m *%s) ToProto() *mme.%s {\n", mech.Name, mech.Name))
		sb.WriteString("\tif m == nil {\n")
		sb.WriteString("\t\treturn nil\n")
		sb.WriteString("\t}\n\n")
		sb.WriteString(fmt.Sprintf("\tpb := &mme.%s{}\n\n", mech.Name))
		
		for _, field := range mech.DataFields {
			if field.Type.IsXMap || field.Type.IsMap {
				// Map 字段需要特殊处理
				sb.WriteString(fmt.Sprintf("\t// TODO: Handle map field %s\n", field.Name))
			} else if field.Type.BaseType != "" {
				goType := ToGoType(field.Type, packageName)
				if strings.HasPrefix(goType, "*") {
					// 指针类型
					sb.WriteString(fmt.Sprintf("\tif m.%s != nil {\n", field.Name))
					sb.WriteString(fmt.Sprintf("\t\tpb.%s = m.%s\n", field.Name, field.Name))
					sb.WriteString("\t}\n")
				} else {
					sb.WriteString(fmt.Sprintf("\tpb.%s = &m.%s\n", field.Name, field.Name))
				}
			} else if field.Type.ValueType != "" {
				// 消息类型
				sb.WriteString(fmt.Sprintf("\tif m.%s != nil {\n", field.Name))
				sb.WriteString(fmt.Sprintf("\t\tpb.%s = m.%s.ToProto()\n", field.Name, field.Name))
				sb.WriteString("\t}\n")
			}
		}
		
		sb.WriteString("\treturn pb\n")
		sb.WriteString("}\n\n")
		
		// 写入文件
		fileName := fmt.Sprintf("%s_mechanisms.go", strings.ToLower(mech.Name))
		filePath := fmt.Sprintf("%s/%s", outputDir, fileName)
		if err := WriteFile(filePath, sb.String()); err != nil {
			return err
		}
	}
	
	return nil
}

// generateModuleFiles 生成 Module Go 文件
func (g *GoStructGenerator) generateModuleFiles(outputDir string) error {
	// TODO: 实现 Module 文件生成
	return nil
}

// generateManagerFiles 生成 Manager Go 文件
func (g *GoStructGenerator) generateManagerFiles(outputDir string) error {
	// TODO: 实现 Manager 文件生成
	return nil
}

// generateEntityFiles 生成 Entity Go 文件
func (g *GoStructGenerator) generateEntityFiles(outputDir string) error {
	// TODO: 实现 Entity 文件生成
	return nil
}

