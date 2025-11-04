package blueprint_gen

import (
	"fmt"
	"strings"

	types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"
)

// GoStructGenerator Go 结构体生成器
type GoStructGenerator struct {
	ctx *BlueprintContext
}

// NewGoStructGenerator 创建新的 Go 结构体生成器
func NewGoStructGenerator(_ctx *BlueprintContext) *GoStructGenerator {
	return &GoStructGenerator{ctx: _ctx}
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
	wrapperGen := NewGoWrapperGenerator(g.ctx)
	if err := wrapperGen.Generate(outputDir); err != nil {
		return fmt.Errorf("failed to generate wrapper files: %w", err)
	}

	return nil
}

// generateMechanismFiles 生成 Mechanism Go 文件
func (g *GoStructGenerator) generateMechanismFiles(outputDir string) error {
	packageName := "mme"

	for _, mech := range g.ctx.Mechanisms {
		sb := strings.Builder{}

		// 文件头部
		sb.WriteString(fmt.Sprintf("package %s\n\n", packageName))
		sb.WriteString("import (\n")
		sb.WriteString("\t\"gitee.com/orbit-w/orbit/app/proto/mme\"\n")
		sb.WriteString("\tdirtyflag \"gitee.com/orbit-w/orbit/lib/base/dirty_flag\"\n")
		sb.WriteString("\tfieldmeta \"gitee.com/orbit-w/orbit/lib/base/field_meta\"\n")
		sb.WriteString("\t\"gitee.com/orbit-w/orbit/lib/module/db/mgo_builder\"\n")
		sb.WriteString("\tmmemodel \"gitee.com/orbit-w/orbit/lib/module/mme_model\"\n")

		// 检查是否有 map/xmap 字段需要 maps 包
		hasMap := false
		hasXMapWithMMEObject := false
		for _, field := range mech.Fields {
			if field.Type.Kind == types.FieldKindXMap || field.Type.Kind == types.FieldKindMap {
				hasMap = true
			}
			// 检查 xmap 字段的 Value 类型是否是 MME Object（Linkable）
			if field.Type.Kind == types.FieldKindXMap && g.isMMEObjectType(field.Type.ValueType) {
				hasXMapWithMMEObject = true
			}
		}
		if hasMap {
			sb.WriteString("\t\"maps\"\n")
			sb.WriteString("\t\"gitee.com/orbit-w/meteor/bases/container/xmap\"\n")
		}
		// 只有当 xmap Value 是 MME Object（Linkable）类型时才需要 xmapwrapper
		if hasXMapWithMMEObject {
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

		// 生成 FromProto 方法
		sb.WriteString(fmt.Sprintf("func (m *%s) FromProto(pb *mme.%s) {\n", mech.Name, mech.Name))
		sb.WriteString("\tif m == nil || pb == nil {\n")
		sb.WriteString("\t\treturn\n")
		sb.WriteString("\t}\n\n")

		for _, field := range mech.Fields {
			if field.Type.Kind == types.FieldKindXMap || field.Type.Kind == types.FieldKindMap {
				// Map 字段需要特殊处理
				sb.WriteString(fmt.Sprintf("\tif pb.%s != nil {\n", field.Name))
				keyType := ToGoBaseTypeFromFieldType(field.Type.KeyType)
				valueType := ToGoBaseTypeFromFieldType(field.Type.ValueType)
				sb.WriteString(fmt.Sprintf("\t\tm.%s = make(map[%s]%s, len(pb.%s))\n", field.Name, keyType, valueType, field.Name))

				// 检查 Value 类型是否是 MME Object 类型
				if g.isMMEObjectType(field.Type.ValueType) {
					// MME Object 类型，需要调用 FromProto
					sb.WriteString(fmt.Sprintf("\t\tfor k, v := range pb.%s {\n", field.Name))
					valueTypeName := ToGoBaseTypeFromFieldType(field.Type.ValueType)
					sb.WriteString("\t\t\tif v != nil {\n")
					sb.WriteString(fmt.Sprintf("\t\t\t\tobj := New%s()\n", valueTypeName))
					sb.WriteString("\t\t\t\tobj.FromProto(v)\n")
					sb.WriteString(fmt.Sprintf("\t\t\t\tm.%s[k] = obj\n", field.Name))
					sb.WriteString("\t\t\t}\n")
					sb.WriteString("\t\t}\n")
				} else {
					// 值类型，直接复制
					sb.WriteString(fmt.Sprintf("\t\tmaps.Copy(m.%s, pb.%s)\n", field.Name, field.Name))
				}
				sb.WriteString("\t}\n")
			} else if field.Type.Kind != types.FieldKindMessage && field.Type.Kind != types.FieldKindMMEObject {
				// 基础类型（值类型），需要解引用指针
				sb.WriteString(fmt.Sprintf("\tif pb.%s != nil {\n", field.Name))
				sb.WriteString(fmt.Sprintf("\t\tm.%s = *pb.%s\n", field.Name, field.Name))
				sb.WriteString("\t}\n")
			} else {
				// 消息类型或 MME Object 类型
				sb.WriteString(fmt.Sprintf("\tif pb.%s != nil {\n", field.Name))
				sb.WriteString(fmt.Sprintf("\t\tif m.%s == nil {\n", field.Name))
				// 需要创建新对象
				valueTypeName := ""
				if field.Type.TypeName != "" {
					valueTypeName = field.Type.TypeName
					// 去掉包名前缀
					if idx := strings.LastIndex(valueTypeName, "."); idx >= 0 {
						valueTypeName = valueTypeName[idx+1:]
					}
				} else {
					valueTypeName = strings.TrimPrefix(ToGoBaseTypeFromFieldType(&field.Type), "*")
				}
				sb.WriteString(fmt.Sprintf("\t\t\tm.%s = New%s()\n", field.Name, valueTypeName))
				sb.WriteString("\t\t}\n")
				sb.WriteString(fmt.Sprintf("\t\tm.%s.FromProto(pb.%s)\n", field.Name, field.Name))
				sb.WriteString("\t}\n")
			}
		}

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
	packageName := "mme"

	for _, module := range g.ctx.Modules {
		sb := strings.Builder{}

		// 文件头部
		sb.WriteString(fmt.Sprintf("package %s\n\n", packageName))
		sb.WriteString("import (\n")
		sb.WriteString("\t\"gitee.com/orbit-w/orbit/app/proto/mme\"\n")
		sb.WriteString("\t\"maps\"\n")
		sb.WriteString(")\n\n")

		// 生成 FieldIndex 常量
		sb.WriteString("const (\n")
		for idx, field := range module.Mechanisms {
			fieldIndexName := fmt.Sprintf("%sFieldIndex%s", module.Name, field.Name)
			sb.WriteString(fmt.Sprintf("\t%s = uint8(%d)\n", fieldIndexName, idx))
		}
		sb.WriteString(")\n\n")

		// 生成 DirtyBit 常量
		sb.WriteString(fmt.Sprintf("// Dirty bits for %s fields\n", module.Name))
		sb.WriteString("const (\n")
		for _, field := range module.Mechanisms {
			dirtyBitName := fmt.Sprintf("%sDirty%sBit", module.Name, field.Name)
			sb.WriteString(fmt.Sprintf("\t%s int64 = 1 << %sFieldIndex%s\n",
				dirtyBitName, module.Name, field.Name))
		}
		sb.WriteString(")\n\n")

		// 生成结构体
		sb.WriteString(fmt.Sprintf("type %s struct {\n", module.Name))
		for _, field := range module.Mechanisms {
			goType := ToGoTypeFromTypesFieldType(&field.Type, packageName)
			sb.WriteString(fmt.Sprintf("\t%s %s\n", field.Name, goType))
		}
		sb.WriteString("}\n\n")

		// 生成 New 构造函数
		sb.WriteString(fmt.Sprintf("func New%s() *%s {\n", module.Name, module.Name))
		sb.WriteString(fmt.Sprintf("\treturn &%s{\n", module.Name))
		for _, field := range module.Mechanisms {
			// Mechanism 类型需要创建新对象
			if g.isMMEObjectType(&field.Type) {
				typeName := strings.TrimPrefix(ToGoBaseTypeFromFieldType(&field.Type), "*")
				sb.WriteString(fmt.Sprintf("\t\t%s: New%s(),\n", field.Name, typeName))
			}
		}
		sb.WriteString("\t}\n")
		sb.WriteString("}\n\n")

		// 生成 DeepCopy 方法
		sb.WriteString(fmt.Sprintf("func (m *%s) DeepCopy(co *%s) {\n", module.Name, module.Name))
		sb.WriteString("\tif m == nil || co == nil {\n")
		sb.WriteString("\t\treturn\n")
		sb.WriteString("\t}\n\n")
		sb.WriteString("\t*co = *m\n")
		for _, field := range module.Mechanisms {
			sb.WriteString(fmt.Sprintf("\tif m.%s != nil {\n", field.Name))
			typeName := strings.TrimPrefix(ToGoBaseTypeFromFieldType(&field.Type), "*")
			sb.WriteString(fmt.Sprintf("\t\tco.%s = &%s{}\n", field.Name, typeName))
			sb.WriteString(fmt.Sprintf("\t\tm.%s.DeepCopy(co.%s)\n", field.Name, field.Name))
			sb.WriteString("\t}\n")
		}
		sb.WriteString("}\n\n")

		// 生成 ToProto 方法
		sb.WriteString(fmt.Sprintf("func (m *%s) ToProto() *mme.%s {\n", module.Name, module.Name))
		sb.WriteString("\tif m == nil {\n")
		sb.WriteString("\t\treturn nil\n")
		sb.WriteString("\t}\n\n")
		sb.WriteString(fmt.Sprintf("\tpb := &mme.%s{}\n\n", module.Name))
		for _, field := range module.Mechanisms {
			sb.WriteString(fmt.Sprintf("\tif m.%s != nil {\n", field.Name))
			sb.WriteString(fmt.Sprintf("\t\tpb.%s = m.%s.ToProto()\n", field.Name, field.Name))
			sb.WriteString("\t}\n")
		}
		sb.WriteString("\treturn pb\n")
		sb.WriteString("}\n\n")

		// 生成 FromProto 方法
		sb.WriteString(fmt.Sprintf("func (m *%s) FromProto(pb *mme.%s) {\n", module.Name, module.Name))
		sb.WriteString("\tif m == nil || pb == nil {\n")
		sb.WriteString("\t\treturn\n")
		sb.WriteString("\t}\n\n")
		for _, field := range module.Mechanisms {
			sb.WriteString(fmt.Sprintf("\tif pb.%s != nil {\n", field.Name))
			sb.WriteString(fmt.Sprintf("\t\tif m.%s == nil {\n", field.Name))
			typeName := strings.TrimPrefix(ToGoBaseTypeFromFieldType(&field.Type), "*")
			sb.WriteString(fmt.Sprintf("\t\t\tm.%s = New%s()\n", field.Name, typeName))
			sb.WriteString("\t\t}\n")
			sb.WriteString(fmt.Sprintf("\t\tm.%s.FromProto(pb.%s)\n", field.Name, field.Name))
			sb.WriteString("\t}\n")
		}
		sb.WriteString("}\n\n")

		// 写入文件
		fileName := fmt.Sprintf("%s_modules.go", CamelToSnake(module.Name))
		filePath := fmt.Sprintf("%s/%s", outputDir, fileName)
		if err := WriteFile(filePath, sb.String()); err != nil {
			return err
		}
	}

	return nil
}

// generateManagerFiles 生成 Manager Go 文件
func (g *GoStructGenerator) generateManagerFiles(outputDir string) error {
	packageName := "mme"

	for _, manager := range g.ctx.Managers {
		sb := strings.Builder{}

		// 文件头部
		sb.WriteString(fmt.Sprintf("package %s\n\n", packageName))
		sb.WriteString("import (\n")
		sb.WriteString("\t\"gitee.com/orbit-w/orbit/app/proto/mme\"\n")
		sb.WriteString("\t\"maps\"\n")
		sb.WriteString(")\n\n")

		// 生成 FieldIndex 常量
		sb.WriteString("const (\n")
		for idx, field := range manager.Fields {
			fieldIndexName := fmt.Sprintf("%sFieldIndex%s", manager.Name, field.Name)
			sb.WriteString(fmt.Sprintf("\t%s = uint8(%d)\n", fieldIndexName, idx))
		}
		sb.WriteString(")\n\n")

		// 生成 DirtyBit 常量
		sb.WriteString(fmt.Sprintf("// Dirty bits for %s fields\n", manager.Name))
		sb.WriteString("const (\n")
		for _, field := range manager.Fields {
			dirtyBitName := fmt.Sprintf("%sDirty%sBit", manager.Name, field.Name)
			sb.WriteString(fmt.Sprintf("\t%s int64 = 1 << %sFieldIndex%s\n",
				dirtyBitName, manager.Name, field.Name))
		}
		sb.WriteString(")\n\n")

		// 生成结构体
		sb.WriteString(fmt.Sprintf("type %s struct {\n", manager.Name))
		for _, field := range manager.Fields {
			goType := ToGoTypeFromTypesFieldType(&field.Type, packageName)
			sb.WriteString(fmt.Sprintf("\t%s %s\n", field.Name, goType))
		}
		sb.WriteString("}\n\n")

		// 生成 New 构造函数
		sb.WriteString(fmt.Sprintf("func New%s() *%s {\n", manager.Name, manager.Name))
		sb.WriteString(fmt.Sprintf("\treturn &%s{\n", manager.Name))
		for _, field := range manager.Fields {
			if field.Type.Kind == types.FieldKindXMap || field.Type.Kind == types.FieldKindMap {
				sb.WriteString(fmt.Sprintf("\t\t%s: make(%s),\n", field.Name, ToGoTypeFromTypesFieldType(&field.Type, packageName)))
			}
		}
		sb.WriteString("\t}\n")
		sb.WriteString("}\n\n")

		// 生成 DeepCopy 方法
		sb.WriteString(fmt.Sprintf("func (m *%s) DeepCopy(co *%s) {\n", manager.Name, manager.Name))
		sb.WriteString("\tif m == nil || co == nil {\n")
		sb.WriteString("\t\treturn\n")
		sb.WriteString("\t}\n\n")
		sb.WriteString("\t*co = *m\n")
		for _, field := range manager.Fields {
			if field.Type.Kind == types.FieldKindXMap || field.Type.Kind == types.FieldKindMap {
				sb.WriteString(fmt.Sprintf("\tif m.%s != nil {\n", field.Name))
				keyType := ToGoBaseTypeFromFieldType(field.Type.KeyType)
				valueType := strings.TrimPrefix(ToGoBaseTypeFromFieldType(field.Type.ValueType), "*")
				sb.WriteString(fmt.Sprintf("\t\tco.%s = make(map[%s]*%s, len(m.%s))\n", field.Name, keyType, valueType, field.Name))
				sb.WriteString(fmt.Sprintf("\t\tfor k, v := range m.%s {\n", field.Name))
				sb.WriteString("\t\t\tif v != nil {\n")
				sb.WriteString(fmt.Sprintf("\t\t\t\tco.%s[k] = &%s{}\n", field.Name, valueType))
				sb.WriteString(fmt.Sprintf("\t\t\t\tv.DeepCopy(co.%s[k])\n", field.Name))
				sb.WriteString("\t\t\t}\n")
				sb.WriteString("\t\t}\n")
				sb.WriteString("\t}\n")
			}
		}
		sb.WriteString("}\n\n")

		// 生成 ToProto 方法
		sb.WriteString(fmt.Sprintf("func (m *%s) ToProto() *mme.%s {\n", manager.Name, manager.Name))
		sb.WriteString("\tif m == nil {\n")
		sb.WriteString("\t\treturn nil\n")
		sb.WriteString("\t}\n\n")
		sb.WriteString(fmt.Sprintf("\tpb := &mme.%s{}\n", manager.Name))
		for _, field := range manager.Fields {
			if field.Type.Kind == types.FieldKindXMap || field.Type.Kind == types.FieldKindMap {
				sb.WriteString(fmt.Sprintf("\tif m.%s != nil {\n", field.Name))
				keyType := ToGoBaseTypeFromFieldType(field.Type.KeyType)
				valueType := ToGoBaseTypeFromFieldType(field.Type.ValueType)
				// 去掉指针符号，获取proto类型
				protoValueType := valueType
				if strings.HasPrefix(valueType, "*") {
					protoValueType = valueType[1:]
				}
				sb.WriteString(fmt.Sprintf("\t\tpb.%s = make(map[%s]*mme.%s, len(m.%s))\n", field.Name, keyType, protoValueType, field.Name))
				sb.WriteString(fmt.Sprintf("\t\tfor k, v := range m.%s {\n", field.Name))
				sb.WriteString("\t\t\tif v != nil {\n")
				sb.WriteString(fmt.Sprintf("\t\t\t\tpb.%s[k] = v.ToProto()\n", field.Name))
				sb.WriteString("\t\t\t}\n")
				sb.WriteString("\t\t}\n")
				sb.WriteString("\t}\n")
			}
		}
		sb.WriteString("\treturn pb\n")
		sb.WriteString("}\n\n")

		// 生成 FromProto 方法
		sb.WriteString(fmt.Sprintf("func (m *%s) FromProto(pb *mme.%s) {\n", manager.Name, manager.Name))
		sb.WriteString("\tif m == nil || pb == nil {\n")
		sb.WriteString("\t\treturn\n")
		sb.WriteString("\t}\n\n")
		for _, field := range manager.Fields {
			if field.Type.Kind == types.FieldKindXMap || field.Type.Kind == types.FieldKindMap {
				sb.WriteString(fmt.Sprintf("\tif pb.%s != nil {\n", field.Name))
				keyType := ToGoBaseTypeFromFieldType(field.Type.KeyType)
				valueType := strings.TrimPrefix(ToGoBaseTypeFromFieldType(field.Type.ValueType), "*")
				sb.WriteString(fmt.Sprintf("\t\tm.%s = make(map[%s]*%s, len(pb.%s))\n", field.Name, keyType, valueType, field.Name))
				sb.WriteString(fmt.Sprintf("\t\tfor k, v := range pb.%s {\n", field.Name))
				sb.WriteString("\t\t\tif v != nil {\n")
				sb.WriteString(fmt.Sprintf("\t\t\t\tobj := New%s()\n", valueType))
				sb.WriteString("\t\t\t\tobj.FromProto(v)\n")
				sb.WriteString(fmt.Sprintf("\t\t\t\tm.%s[k] = obj\n", field.Name))
				sb.WriteString("\t\t\t}\n")
				sb.WriteString("\t\t}\n")
				sb.WriteString("\t}\n")
			}
		}
		sb.WriteString("}\n\n")

		// 写入文件
		fileName := fmt.Sprintf("%s_managers.go", CamelToSnake(manager.Name))
		filePath := fmt.Sprintf("%s/%s", outputDir, fileName)
		if err := WriteFile(filePath, sb.String()); err != nil {
			return err
		}
	}

	return nil
}

// generateEntityFiles 生成 Entity Go 文件
func (g *GoStructGenerator) generateEntityFiles(outputDir string) error {
	packageName := "mme"

	for _, entity := range g.ctx.Entities {
		sb := strings.Builder{}

		// 文件头部
		sb.WriteString(fmt.Sprintf("package %s\n\n", packageName))
		sb.WriteString("import (\n")
		sb.WriteString("\t\"gitee.com/orbit-w/orbit/app/proto/mme\"\n")
		sb.WriteString("\t\"maps\"\n")
		sb.WriteString(")\n\n")

		// 生成 FieldIndex 常量（Entity的Id字段默认是0）
		sb.WriteString("const (\n")
		sb.WriteString(fmt.Sprintf("\t%sFieldIndexXXXId = uint8(0)\n", entity.Name))
		for idx, field := range entity.Fields {
			fieldIndexName := fmt.Sprintf("%sFieldIndex%s", entity.Name, field.Name)
			sb.WriteString(fmt.Sprintf("\t%s = uint8(%d)\n", fieldIndexName, idx+1))
		}
		sb.WriteString(")\n\n")

		// 生成 DirtyBit 常量
		sb.WriteString(fmt.Sprintf("// Dirty bits for %s fields\n", entity.Name))
		sb.WriteString("const (\n")
		sb.WriteString(fmt.Sprintf("\t%sDirtyXXXIdBit int64 = 1 << %sFieldIndexXXXId\n", entity.Name, entity.Name))
		for _, field := range entity.Fields {
			dirtyBitName := fmt.Sprintf("%sDirty%sBit", entity.Name, field.Name)
			sb.WriteString(fmt.Sprintf("\t%s int64 = 1 << %sFieldIndex%s\n",
				dirtyBitName, entity.Name, field.Name))
		}
		sb.WriteString(")\n\n")

		// 生成结构体
		sb.WriteString(fmt.Sprintf("type %s struct {\n", entity.Name))
		sb.WriteString("\tXXXId int64\n")
		for _, field := range entity.Fields {
			goType := ToGoTypeFromTypesFieldType(&field.Type, packageName)
			sb.WriteString(fmt.Sprintf("\t%s %s\n", field.Name, goType))
		}
		sb.WriteString("}\n\n")

		// 生成 New 构造函数
		sb.WriteString(fmt.Sprintf("func New%s() *%s {\n", entity.Name, entity.Name))
		sb.WriteString(fmt.Sprintf("\treturn &%s{\n", entity.Name))
		for _, field := range entity.Fields {
			typeName := field.Type.TypeName
			sb.WriteString(fmt.Sprintf("\t\t%s: New%s(),\n", field.Name, typeName))
		}
		sb.WriteString("\t}\n")
		sb.WriteString("}\n\n")

		// 生成 DeepCopy 方法
		sb.WriteString(fmt.Sprintf("func (e *%s) DeepCopy(co *%s) {\n", entity.Name, entity.Name))
		sb.WriteString("\tif e == nil || co == nil {\n")
		sb.WriteString("\t\treturn\n")
		sb.WriteString("\t}\n\n")
		sb.WriteString("\t*co = *e\n")
		for _, field := range entity.Fields {
			typeName := field.Type.TypeName
			sb.WriteString(fmt.Sprintf("\tco.%s = &%s{}\n", field.Name, typeName))
			sb.WriteString(fmt.Sprintf("\te.%s.DeepCopy(co.%s)\n", field.Name, field.Name))
		}
		sb.WriteString("}\n\n")

		// 生成 ToProto 方法
		sb.WriteString(fmt.Sprintf("func (e *%s) ToProto() *mme.%s {\n", entity.Name, entity.Name))
		sb.WriteString("\tif e == nil {\n")
		sb.WriteString("\t\treturn nil\n")
		sb.WriteString("\t}\n\n")
		sb.WriteString(fmt.Sprintf("\tpb := &mme.%s{}\n\n", entity.Name))
		sb.WriteString("\tid := e.XXXId\n")
		sb.WriteString("\tpb.XXXId = id\n\n")
		for _, field := range entity.Fields {
			sb.WriteString(fmt.Sprintf("\tif e.%s != nil {\n", field.Name))
			sb.WriteString(fmt.Sprintf("\t\tpb.%s = e.%s.ToProto()\n", field.Name, field.Name))
			sb.WriteString("\t}\n")
		}
		sb.WriteString("\treturn pb\n")
		sb.WriteString("}\n\n")

		// 生成 FromProto 方法
		sb.WriteString(fmt.Sprintf("func (e *%s) FromProto(pb *mme.%s) {\n", entity.Name, entity.Name))
		sb.WriteString("\tif e == nil || pb == nil {\n")
		sb.WriteString("\t\treturn\n")
		sb.WriteString("\t}\n\n")
		sb.WriteString("\te.XXXId = pb.XXXId\n\n")
		for _, field := range entity.Fields {
			typeName := field.Type.TypeName
			sb.WriteString(fmt.Sprintf("\tif pb.%s != nil {\n", field.Name))
			sb.WriteString(fmt.Sprintf("\t\tif e.%s == nil {\n", field.Name))
			sb.WriteString(fmt.Sprintf("\t\t\te.%s = New%s()\n", field.Name, typeName))
			sb.WriteString("\t\t}\n")
			sb.WriteString(fmt.Sprintf("\t\te.%s.FromProto(pb.%s)\n", field.Name, field.Name))
			sb.WriteString("\t}\n")
		}
		sb.WriteString("}\n\n")

		// 写入文件
		fileName := fmt.Sprintf("%s_entities.go", CamelToSnake(entity.Name))
		filePath := fmt.Sprintf("%s/%s", outputDir, fileName)
		if err := WriteFile(filePath, sb.String()); err != nil {
			return err
		}
	}

	return nil
}

// isMMEObjectType 判断是否是 MME Object 类型（Linkable）
func (g *GoStructGenerator) isMMEObjectType(fieldType *types.FieldType) bool {
	if fieldType == nil {
		return false
	}

	// 如果是 Map/XMap，检查 ValueType
	if fieldType.Kind == types.FieldKindMap || fieldType.Kind == types.FieldKindXMap {
		if fieldType.ValueType != nil {
			return g.isMMEObjectType(fieldType.ValueType)
		}
		return false
	}

	// 获取类型名称
	typeName := ""
	if fieldType.TypeName != "" {
		typeName = fieldType.TypeName
	} else {
		typeName = fieldType.Kind.String()
	}

	// 检查是否是 Module、Mechanism、Manager 或 Entity 类型
	objectType, ok := g.ctx.ObjectTypeMap[typeName]
	if !ok {
		return false
	}

	return isMMEObjectType(objectType)
}

// isMMEObjectType 判断是否是 MME Object 类型
func isMMEObjectType(objectType ObjectType) bool {
	return objectType == ObjectTypeModule || objectType == ObjectTypeMechanism ||
		objectType == ObjectTypeManager || objectType == ObjectTypeEntity
}
