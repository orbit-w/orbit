package blueprint_gen

import (
	"fmt"
	"strings"

	types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"
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

		// 检查是否有 map 字段需要 maps 包
		hasMap := false
		for _, field := range mech.Fields {
			if field.Type.Kind == types.FieldKindXMap || field.Type.Kind == types.FieldKindMap {
				hasMap = true
				break
			}
		}
		if hasMap {
			sb.WriteString("\t\"maps\"\n")
			sb.WriteString("\t\"gitee.com/orbit-w/meteor/bases/container/xmap\"\n")
			sb.WriteString("\txmapwrapper \"gitee.com/orbit-w/orbit/lib/module/xmapwrapper\"\n")
		}

		sb.WriteString("\t\"google.golang.org/protobuf/proto\"\n")
		sb.WriteString(")\n\n")

		// 生成 FieldIndex 常量
		sb.WriteString("const (\n")
		for idx, field := range mech.Fields {
			fieldIndexName := fmt.Sprintf("%sFieldIndex%s", mech.Name, field.Name)
			sb.WriteString(fmt.Sprintf("\t%s = uint8(%d)\n", fieldIndexName, idx))
		}
		sb.WriteString(")\n\n")

		// 生成 DirtyBit 常量
		sb.WriteString("// Dirty bits for Mechanism fields\n")
		sb.WriteString("const (\n")
		for _, field := range mech.Fields {
			dirtyBitName := fmt.Sprintf("%sDirty%sBit", mech.Name, field.Name)
			sb.WriteString(fmt.Sprintf("\t%s int64 = 1 << %sFieldIndex%s\n",
				dirtyBitName, mech.Name, field.Name))
		}
		sb.WriteString(")\n\n")

		// 生成结构体
		sb.WriteString(fmt.Sprintf("type %s struct {\n", mech.Name))
		for _, field := range mech.Fields {
			goType := ToGoTypeFromTypesFieldType(&field.Type, packageName)
			sb.WriteString(fmt.Sprintf("\t%s %s\n", field.Name, goType))
		}
		sb.WriteString("}\n\n")

		// 生成 New 构造函数
		sb.WriteString(fmt.Sprintf("func New%s() *%s {\n", mech.Name, mech.Name))
		sb.WriteString(fmt.Sprintf("\treturn &%s{\n", mech.Name))
		// 初始化 map 字段
		hasMapInit := false
		for _, field := range mech.Fields {
			if field.Type.Kind == types.FieldKindXMap || field.Type.Kind == types.FieldKindMap {
				if !hasMapInit {
					hasMapInit = true
				}
				sb.WriteString(fmt.Sprintf("\t\t%s: make(%s),\n", field.Name, ToGoTypeFromTypesFieldType(&field.Type, packageName)))
			}
		}
		if hasMapInit {
			sb.WriteString("\t}\n")
		} else {
			sb.WriteString("\t}\n")
		}
		sb.WriteString("}\n\n")

		// 生成 DeepCopy 方法
		sb.WriteString(fmt.Sprintf("func (m *%s) DeepCopy(co *%s) {\n", mech.Name, mech.Name))
		sb.WriteString("\tif m == nil || co == nil {\n")
		sb.WriteString("\t\treturn\n")
		sb.WriteString("\t}\n\n")
		sb.WriteString("\t*co = *m\n")
		// 深拷贝 map 字段
		for _, field := range mech.Fields {
			if field.Type.Kind == types.FieldKindXMap || field.Type.Kind == types.FieldKindMap {
				sb.WriteString(fmt.Sprintf("\tmaps.Copy(co.%s, m.%s)\n", field.Name, field.Name))
			}
		}
		sb.WriteString("}\n\n")

		// 生成 ToProto 方法
		sb.WriteString(fmt.Sprintf("func (m *%s) ToProto() *mme.%s {\n", mech.Name, mech.Name))
		sb.WriteString("\tif m == nil {\n")
		sb.WriteString("\t\treturn nil\n")
		sb.WriteString("\t}\n\n")
		sb.WriteString(fmt.Sprintf("\tpb := &mme.%s{}\n\n", mech.Name))

		for _, field := range mech.Fields {
			if field.Type.Kind == types.FieldKindXMap || field.Type.Kind == types.FieldKindMap {
				// Map 字段需要特殊处理
				sb.WriteString(fmt.Sprintf("\tif m.%s != nil {\n", field.Name))
				keyType := ToGoBaseTypeFromFieldType(field.Type.KeyType)
				valueType := ToGoBaseTypeFromFieldType(field.Type.ValueType)
				sb.WriteString(fmt.Sprintf("\t\tpb.%s = make(map[%s]%s, len(m.%s))\n", field.Name, keyType, valueType, field.Name))
				sb.WriteString(fmt.Sprintf("\t\tmaps.Copy(pb.%s, m.%s)\n", field.Name, field.Name))
				sb.WriteString("\t}\n")
			} else if field.Type.Kind != types.FieldKindMessage && field.Type.Kind != types.FieldKindMMEObject {
				// 基础类型（值类型）
				sb.WriteString(fmt.Sprintf("\t%s := m.%s\n", strings.ToLower(field.Name), field.Name))
				sb.WriteString(fmt.Sprintf("\tpb.%s = &%s\n", field.Name, strings.ToLower(field.Name)))
			} else {
				// 消息类型或 MME Object 类型
				sb.WriteString(fmt.Sprintf("\tif m.%s != nil {\n", field.Name))
				sb.WriteString(fmt.Sprintf("\t\tpb.%s = m.%s.ToProto()\n", field.Name, field.Name))
				sb.WriteString("\t}\n")
			}
		}

		sb.WriteString("\treturn pb\n")
		sb.WriteString("}\n\n")

		// 写入文件
		fileName := fmt.Sprintf("%s_mechanisms.go", CamelToSnake(mech.Name))
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
