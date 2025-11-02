package blueprint_gen

import (
	"fmt"
	"os"
	"strings"
)

// GoWrapperGenerator Go Wrapper 生成器
type GoWrapperGenerator struct {
	data *BlueprintData
}

// NewGoWrapperGenerator 创建新的 Go Wrapper 生成器
func NewGoWrapperGenerator(data *BlueprintData) *GoWrapperGenerator {
	return &GoWrapperGenerator{data: data}
}

// Generate 生成所有 Wrapper 文件
func (g *GoWrapperGenerator) Generate(outputDir string) error {
	// 生成 Mechanism Wrapper
	if err := g.generateMechanismWrappers(outputDir); err != nil {
		return fmt.Errorf("failed to generate mechanism wrappers: %w", err)
	}
	
	// TODO: 生成 Module、Manager、Entity Wrapper
	return nil
}

// generateMechanismWrappers 生成 Mechanism Wrapper
func (g *GoWrapperGenerator) generateMechanismWrappers(outputDir string) error {
	packageName := "mme"
	
	for _, mech := range g.data.Mechanisms {
		sb := strings.Builder{}
		
		// Wrapper 代码（不包含 package 和 import，因为这些会从已有文件中获取）
		
		// 生成 Wrapper 结构体
		wrapperName := mech.Name + "Wrapper"
		sb.WriteString(fmt.Sprintf("type %s struct {\n", wrapperName))
		sb.WriteString(fmt.Sprintf("\tdata *%s\n", mech.Name))
		sb.WriteString("\tdirtyflag.IDirtyFlag\n")
		sb.WriteString("\tfieldMetas *fieldmeta.FieldMetas\n\n")
		
		// 生成 map 访问器字段
		for _, field := range mech.DataFields {
			if field.Type.IsXMap || field.Type.IsMap {
				// 判断 Value 类型
				if g.isMMEObjectType(field.Type.ValueType) {
					// MME Object 类型，使用 XMapWrapper
					sb.WriteString(fmt.Sprintf("\t// Value 为 MME Object 类型，使用 XMapWrapper\n"))
					// TODO: 生成 XMapWrapper 字段
				} else {
					// 值类型，使用 MapAccessor
					keyType := toGoBaseTypeSimple(field.Type.KeyType)
					valueType := toGoBaseTypeSimple(field.Type.ValueType)
					accessorFieldName := strings.ToLower(field.Name[0:1]) + field.Name[1:] + "Accessor"
					sb.WriteString(fmt.Sprintf("\t// Value 为值类型，使用 xmap.MapAccessor 进行包装\n"))
					sb.WriteString(fmt.Sprintf("\t%s *xmap.MapAccessor[%s, %s]\n",
						accessorFieldName, keyType, valueType))
				}
			}
		}
		
		sb.WriteString("}\n\n")
		
		// 生成 New 构造函数
		sb.WriteString(fmt.Sprintf("func New%s(data *%s) *%s {\n", wrapperName, mech.Name, wrapperName))
		sb.WriteString("\tif data == nil {\n")
		sb.WriteString("\t\tpanic(\"data is nil\")\n")
		sb.WriteString("\t}\n")
		sb.WriteString(fmt.Sprintf("\tw := &%s{\n", wrapperName))
		sb.WriteString("\t\tdata:       data,\n")
		sb.WriteString("\t\tIDirtyFlag: dirtyflag.NewDirtyFlag(),\n")
		sb.WriteString("\t\tfieldMetas: fieldmeta.NewFieldMetas(),\n")
		sb.WriteString("\t}\n\n")
		
		// 初始化 map 访问器
		for _, field := range mech.DataFields {
			if (field.Type.IsXMap || field.Type.IsMap) && !g.isMMEObjectType(field.Type.ValueType) {
				accessorName := strings.ToLower(field.Name[0:1]) + field.Name[1:] + "Accessor"
				dirtyBitName := fmt.Sprintf("%sDirty%sBit", mech.Name, field.Name)
				sb.WriteString(fmt.Sprintf("\tw.%s = xmap.NewMapAccessorWithMarker(&w.data.%s, w, %s)\n",
					accessorName, field.Name, dirtyBitName))
			}
		}
		
		sb.WriteString("\treturn w\n")
		sb.WriteString("}\n\n")
		
		// 生成 InitFieldContext 方法
		sb.WriteString(fmt.Sprintf("func (w *%s) InitFieldContext() {\n", wrapperName))
		for _, field := range mech.DataFields {
			fieldIndexName := fmt.Sprintf("%sFieldIndex%s", mech.Name, field.Name)
			// 根据 access 选项设置字段类型
			if field.Options.Access == "all" || field.Options.Access == "s" || field.Options.Access == "" {
				sb.WriteString(fmt.Sprintf("\tw.fieldMetas.SetFieldType(%s, fieldmeta.FieldTypeSync)\n",
					fieldIndexName))
			}
		}
		sb.WriteString("}\n\n")
		
		// 生成 Name 方法
		sb.WriteString(fmt.Sprintf("func (w *%s) Name() string {\n", wrapperName))
		sb.WriteString(fmt.Sprintf("\treturn \"%s\"\n", mech.Name))
		sb.WriteString("}\n\n")
		
		// 生成 Getter 和 Setter 方法
		for _, field := range mech.DataFields {
			if field.Type.IsXMap || field.Type.IsMap {
				// Map 字段生成访问器方法
				accessorName := strings.ToLower(field.Name[0:1]) + field.Name[1:]
				methodName := strings.ToUpper(field.Name[0:1]) + field.Name[1:]
				if !g.isMMEObjectType(field.Type.ValueType) {
					keyType := toGoBaseTypeSimple(field.Type.KeyType)
					valueType := toGoBaseTypeSimple(field.Type.ValueType)
					sb.WriteString(fmt.Sprintf("// 包装器-获取%s访问器\n", field.Name))
					sb.WriteString(fmt.Sprintf("func (w *%s) Get%sAccessor() *xmap.MapAccessor[%s, %s] {\n",
						wrapperName, methodName, keyType, valueType))
					sb.WriteString(fmt.Sprintf("\treturn w.%sAccessor\n", accessorName))
					sb.WriteString("}\n\n")
				}
			} else {
				// 普通字段生成 Getter 和 Setter
				goType := ToGoType(field.Type, packageName)
				methodName := strings.ToUpper(field.Name[0:1]) + field.Name[1:]
				dirtyBitName := fmt.Sprintf("%sDirty%sBit", mech.Name, field.Name)
				
				// Getter
				if strings.HasPrefix(goType, "*") {
					// 指针类型
					sb.WriteString(fmt.Sprintf("func (w *%s) Get%s() %s {\n", wrapperName, methodName, goType))
					sb.WriteString(fmt.Sprintf("\treturn w.data.%s\n", field.Name))
					sb.WriteString("}\n\n")
				} else {
					// 值类型
					sb.WriteString(fmt.Sprintf("func (w *%s) Get%s() %s {\n", wrapperName, methodName, strings.TrimPrefix(goType, "*")))
					sb.WriteString(fmt.Sprintf("\treturn w.data.%s\n", field.Name))
					sb.WriteString("}\n\n")
				}
				
				// Setter
				sb.WriteString(fmt.Sprintf("func (w *%s) Set%s(v %s) {\n", wrapperName, methodName, strings.TrimPrefix(goType, "*")))
				sb.WriteString(fmt.Sprintf("\tw.data.%s = v\n", field.Name))
				sb.WriteString(fmt.Sprintf("\tw.MarkDirty(%s)\n", dirtyBitName))
				sb.WriteString("}\n\n")
			}
		}
		
		// 生成 ClearAllDirty 方法
		sb.WriteString(fmt.Sprintf("func (w *%s) ClearAllDirty() {\n", wrapperName))
		sb.WriteString("\tw.IDirtyFlag.ClearAllDirty()\n")
		for _, field := range mech.DataFields {
			if (field.Type.IsXMap || field.Type.IsMap) && !g.isMMEObjectType(field.Type.ValueType) {
				accessorName := strings.ToLower(field.Name[0:1]) + field.Name[1:] + "Accessor"
				sb.WriteString(fmt.Sprintf("\tw.%s.ResetOperations()\n", accessorName))
			}
		}
		sb.WriteString("}\n\n")
		
		// 生成 BuildMongoUpdate 方法
		sb.WriteString(fmt.Sprintf("func (w *%s) BuildMongoUpdate(builder *mgo_builder.MongoUpdateBuilder, path *mgo_builder.NestedPath) {\n", wrapperName))
		sb.WriteString("\tif w == nil {\n")
		sb.WriteString("\t\treturn\n")
		sb.WriteString("\t}\n\n")
		
		for _, field := range mech.DataFields {
			dirtyBitName := fmt.Sprintf("%sDirty%sBit", mech.Name, field.Name)
			fieldNameSnake := CamelToSnake(field.Name)
			
			if field.Type.IsXMap || field.Type.IsMap {
				sb.WriteString(fmt.Sprintf("\tif w.IsDirty(%s) {\n", dirtyBitName))
				if !g.isMMEObjectType(field.Type.ValueType) {
					accessorName := strings.ToLower(field.Name[0:1]) + field.Name[1:] + "Accessor"
					sb.WriteString(fmt.Sprintf("\t\tbuilder.SetNestedPath(path, \"%s\", w.%s.Clone())\n",
						fieldNameSnake, accessorName))
				} else {
					sb.WriteString(fmt.Sprintf("\t\t// TODO: Handle MME Object map field %s\n", field.Name))
				}
				sb.WriteString("\t}\n")
			} else {
				sb.WriteString(fmt.Sprintf("\tif w.IsDirty(%s) {\n", dirtyBitName))
				goType := ToGoType(field.Type, packageName)
				if strings.HasPrefix(goType, "*") {
					sb.WriteString(fmt.Sprintf("\t\tif w.data.%s != nil {\n", field.Name))
					sb.WriteString(fmt.Sprintf("\t\t\tbuilder.SetNestedPath(path, \"%s\", *w.data.%s)\n",
						fieldNameSnake, field.Name))
					sb.WriteString("\t\t}\n")
				} else {
					sb.WriteString(fmt.Sprintf("\t\tbuilder.SetNestedPath(path, \"%s\", w.data.%s)\n",
						fieldNameSnake, field.Name))
				}
				sb.WriteString("\t}\n")
			}
		}
		sb.WriteString("}\n\n")
		
		// 生成 ToIncrementalProtoWithContext 方法
		sb.WriteString("// ToIncrementalProtoWithContext 根据脏标记位构建增量数据的 protoMessage\n")
		sb.WriteString("// 只返回标记为脏的字段数据，用于增量同步\n")
		sb.WriteString(fmt.Sprintf("func (w *%s) ToIncrementalProtoWithContext(ctx mmemodel.SyncContext) proto.Message {\n", wrapperName))
		sb.WriteString("\tif w == nil {\n")
		sb.WriteString("\t\treturn nil\n")
		sb.WriteString("\t}\n\n")
		sb.WriteString("\t// 如果没有脏标记，返回 nil\n")
		sb.WriteString("\tif !w.HasAnyDirty() {\n")
		sb.WriteString("\t\treturn nil\n")
		sb.WriteString("\t}\n\n")
		sb.WriteString(fmt.Sprintf("\tincremental := &mme.%s{}\n\n", mech.Name))
		
		for _, field := range mech.DataFields {
			fieldIndexName := fmt.Sprintf("%sFieldIndex%s", mech.Name, field.Name)
			dirtyBitName := fmt.Sprintf("%sDirty%sBit", mech.Name, field.Name)
			
			if field.Type.IsXMap || field.Type.IsMap {
				// Map 字段生成 ChangeList
				sb.WriteString(fmt.Sprintf("\tif mmemodel.FieldCanBeIncrementalSynced(w, %s, %s, ctx) {\n",
					dirtyBitName, fieldIndexName))
				changeListName := field.Name + "_XXXChangeList"
				recordName := fmt.Sprintf("%s_%s_XXXMapChangeRecord", mech.Name, field.Name)
				
				sb.WriteString(fmt.Sprintf("\t\tincremental.%s = make([]*mme.%s, 0)\n",
					changeListName, recordName))
				
				if !g.isMMEObjectType(field.Type.ValueType) {
					accessorName := strings.ToLower(field.Name[0:1]) + field.Name[1:] + "Accessor"
					sb.WriteString(fmt.Sprintf("\t\tw.%s.RangeOperations(func(key %s, operation xmap.MapOperation[%s]) bool {\n",
						accessorName, toGoBaseTypeSimple(field.Type.KeyType), toGoBaseTypeSimple(field.Type.KeyType)))
					sb.WriteString("\t\t\tswitch operation.Type {\n")
					sb.WriteString("\t\t\tcase xmap.SetOperation:\n")
					sb.WriteString(fmt.Sprintf("\t\t\t\tv, _ := w.%s.Get(key)\n", accessorName))
					sb.WriteString(fmt.Sprintf("\t\t\t\tincremental.%s = append(incremental.%s, &mme.%s{\n",
						changeListName, changeListName, recordName))
					sb.WriteString("\t\t\t\t\tKey:   key,\n")
					sb.WriteString("\t\t\t\t\tValue: v,\n")
					sb.WriteString("\t\t\t\t})\n")
					sb.WriteString("\t\t\tcase xmap.DeleteOperation:\n")
					sb.WriteString(fmt.Sprintf("\t\t\t\tincremental.%s = append(incremental.%s, &mme.%s{\n",
						changeListName, changeListName, recordName))
					sb.WriteString("\t\t\t\t\tKey:      key,\n")
					sb.WriteString("\t\t\t\t\tIsDelete: true,\n")
					sb.WriteString("\t\t\t\t})\n")
					sb.WriteString("\t\t\t}\n")
					sb.WriteString("\t\t\treturn true\n")
					sb.WriteString("\t\t})\n")
				} else {
					sb.WriteString(fmt.Sprintf("\t\t// TODO: Handle MME Object map field %s\n", field.Name))
				}
				sb.WriteString("\t}\n")
			} else {
				// 普通字段
				sb.WriteString(fmt.Sprintf("\tif mmemodel.FieldCanBeIncrementalSynced(w, %s, %s, ctx) {\n",
					dirtyBitName, fieldIndexName))
				goType := ToGoType(field.Type, packageName)
				if strings.HasPrefix(goType, "*") {
					sb.WriteString(fmt.Sprintf("\t\tv := w.Get%s()\n", strings.ToUpper(field.Name[0:1])+field.Name[1:]))
					sb.WriteString(fmt.Sprintf("\t\tif v != nil {\n"))
					sb.WriteString(fmt.Sprintf("\t\t\tincremental.%s = v\n", field.Name))
					sb.WriteString("\t\t}\n")
				} else {
					sb.WriteString(fmt.Sprintf("\t\tv := w.Get%s()\n", strings.ToUpper(field.Name[0:1])+field.Name[1:]))
					sb.WriteString(fmt.Sprintf("\t\tincremental.%s = &v\n", field.Name))
				}
				sb.WriteString("\t}\n")
			}
		}
		
		sb.WriteString("\treturn incremental\n")
		sb.WriteString("}\n\n")
		
		// 写入文件（追加到机制文件）
		fileName := fmt.Sprintf("%s_mechanisms.go", strings.ToLower(mech.Name))
		filePath := fmt.Sprintf("%s/%s", outputDir, fileName)
		
		// 读取现有文件内容（如果存在）
		existingContent := ""
		if FileExists(filePath) {
			if data, err := ReadFileContent(filePath); err == nil {
				existingContent = data
			}
		}
		
		// 追加新内容（如果文件已存在，在末尾添加换行）
		var newContent string
		if existingContent != "" {
			// 确保现有内容以换行结尾
			existingContent = strings.TrimRight(existingContent, " \n\r\t")
			newContent = existingContent + "\n\n" + sb.String()
		} else {
			// 如果文件不存在，需要添加 package 声明
			newContent = fmt.Sprintf("package %s\n\n", packageName) + sb.String()
		}
		if err := WriteFile(filePath, newContent); err != nil {
			return err
		}
	}
	
	return nil
}

// isMMEObjectType 判断是否是 MME Object 类型
func (g *GoWrapperGenerator) isMMEObjectType(typeName string) bool {
	// 检查是否是 Module 或 Mechanism 类型
	for _, module := range g.data.Modules {
		if module.Name == typeName {
			return true
		}
	}
	for _, mech := range g.data.Mechanisms {
		if mech.Name == typeName {
			return true
		}
	}
	return false
}

// toGoBaseTypeSimple 转换基础类型（不带包名）
func toGoBaseTypeSimple(typeStr string) string {
	switch typeStr {
	case "int32":
		return "int32"
	case "int64":
		return "int64"
	case "string":
		return "string"
	case "bool":
		return "bool"
	case "float":
		return "float32"
	case "double":
		return "float64"
	default:
		return typeStr
	}
}

// ReadFileContent 读取文件内容
func ReadFileContent(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}


