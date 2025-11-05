package blueprint_gen

import (
	"fmt"
	"os"
	"strings"

	types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"
)

// GoWrapperGenerator Go Wrapper 生成器
type GoWrapperGenerator struct {
	data *BlueprintContext
}

// NewGoWrapperGenerator 创建新的 Go Wrapper 生成器
func NewGoWrapperGenerator(data *BlueprintContext) *GoWrapperGenerator {
	return &GoWrapperGenerator{data: data}
}

// Generate 生成所有 Wrapper 文件
func (g *GoWrapperGenerator) Generate(outputDir string) error {
	// 生成 Mechanism Wrapper
	if err := g.generateMechanismWrappers(outputDir); err != nil {
		return fmt.Errorf("failed to generate mechanism wrappers: %w", err)
	}

	// 生成 Module Wrapper
	if err := g.generateModuleWrappers(outputDir); err != nil {
		return fmt.Errorf("failed to generate module wrappers: %w", err)
	}

	// TODO: 生成 Manager、Entity Wrapper
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

		// 生成 map 访问器字段（只有xmap字段才需要accessor）
		for _, field := range mech.Fields {
			if field.Type.IsXMapField() {
				// 判断 Value 类型
				if field.Type.IsXMapValueMMEObject() {
					// 目前机制中，不支持用xmap 编排MME Object 类型
					panic(fmt.Sprintf("mechanism %s 的xmap字段 %s 的Value类型是 MMEObject 类型，不支持用xmap编排MME Object 类型", mech.Name, field.Name))
					// TODO: 生成 XMapWrapper 字段
				} else {
					// 值类型，使用 MapAccessor
					keyType := ToGoBaseTypeFromFieldType(field.Type.KeyType)
					valueType := ToGoBaseTypeFromFieldType(field.Type.ValueType)
					accessorFieldName := strings.ToLower(field.Name[0:1]) + field.Name[1:] + "Accessor"
					sb.WriteString("\t// Value 为值类型，使用 xmap.MapAccessor 进行包装\n")
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

		// 初始化 map 访问器（只有xmap字段才需要accessor，普通map不需要）
		for _, field := range mech.Fields {
			if field.Type.IsXMapField() && !field.Type.IsXMapValueMMEObject() {
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
		for _, field := range mech.Fields {
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

		// 生成 MatchesAll 方法（实现 IFieldMetaContext 接口）
		sb.WriteString("// MatchesAll 判断字段是否匹配所有类型标记\n")
		sb.WriteString(fmt.Sprintf("func (w *%s) MatchesAll(fieldID uint8, fieldTypes ...fieldmeta.FieldType) bool {\n", wrapperName))
		sb.WriteString("\treturn w.fieldMetas.MatchesAll(fieldID, fieldTypes...)\n")
		sb.WriteString("}\n\n")

		// 生成 Getter 和 Setter 方法
		for _, field := range mech.Fields {
			if field.Type.IsXMapField() {
				// XMap 字段生成访问器方法
				accessorName := strings.ToLower(field.Name[0:1]) + field.Name[1:]
				methodName := strings.ToUpper(field.Name[0:1]) + field.Name[1:]
				if field.Type.IsXMapValueMMEObject() {
					panic(fmt.Sprintf("mechanism %s 的xmap字段 %s 的Value类型是 MMEObject 类型，不支持用xmap编排MME Object 类型", mech.Name, field.Name))
				} else {
					keyType := ToGoBaseTypeFromFieldType(field.Type.KeyType)
					valueType := ToGoBaseTypeFromFieldType(field.Type.ValueType)
					sb.WriteString(fmt.Sprintf("// 包装器-获取%s访问器\n", field.Name))
					sb.WriteString(fmt.Sprintf("func (w *%s) Get%sAccessor() *xmap.MapAccessor[%s, %s] {\n",
						wrapperName, methodName, keyType, valueType))
					sb.WriteString(fmt.Sprintf("\treturn w.%sAccessor\n", accessorName))
					sb.WriteString("}\n\n")
				}
			} else if field.Type.IsMapField() {
				// 普通 Map 字段生成 Getter 方法，直接返回 map
				methodName := strings.ToUpper(field.Name[0:1]) + field.Name[1:]
				goType := ToGoTypeFromTypesFieldType(&field.Type, packageName)
				sb.WriteString(fmt.Sprintf("// 包装器-获取%s\n", field.Name))
				sb.WriteString(fmt.Sprintf("func (w *%s) Get%s() %s {\n", wrapperName, methodName, goType))
				sb.WriteString(fmt.Sprintf("\treturn w.data.%s\n", field.Name))
				sb.WriteString("}\n\n")
				// 生成 Setter 方法，标记脏标记
				dirtyBitName := fmt.Sprintf("%sDirty%sBit", mech.Name, field.Name)
				sb.WriteString(fmt.Sprintf("// 包装器-设置%s\n", field.Name))
				sb.WriteString(fmt.Sprintf("func (w *%s) Set%s(v %s) {\n", wrapperName, methodName, goType))
				sb.WriteString(fmt.Sprintf("\tw.data.%s = v\n", field.Name))
				sb.WriteString(fmt.Sprintf("\tw.MarkDirty(%s)\n", dirtyBitName))
				sb.WriteString("}\n\n")
			} else {
				// 普通字段生成 Getter 和 Setter
				goType := ToGoTypeFromTypesFieldType(&field.Type, packageName)
				methodName := strings.ToUpper(field.Name[0:1]) + field.Name[1:]
				dirtyBitName := fmt.Sprintf("%sDirty%sBit", mech.Name, field.Name)

				// Getter - 基础类型返回值类型，不是指针
				sb.WriteString(fmt.Sprintf("func (w *%s) Get%s() %s {\n", wrapperName, methodName, goType))
				sb.WriteString(fmt.Sprintf("\treturn w.data.%s\n", field.Name))
				sb.WriteString("}\n\n")

				// Setter - 基础类型接受值类型
				sb.WriteString(fmt.Sprintf("func (w *%s) Set%s(v %s) {\n", wrapperName, methodName, goType))
				sb.WriteString(fmt.Sprintf("\tw.data.%s = v\n", field.Name))
				sb.WriteString(fmt.Sprintf("\tw.MarkDirty(%s)\n", dirtyBitName))
				sb.WriteString("}\n\n")
			}
		}

		// 生成 ClearAllDirtyFlags 方法
		sb.WriteString("// ClearAllDirtyFlags 清除所有脏标记位\n")
		sb.WriteString(fmt.Sprintf("func (w *%s) ClearAllDirtyFlags() {\n", wrapperName))
		sb.WriteString("\tw.IDirtyFlag.ClearAllDirty()\n\n")
		sb.WriteString("\t// 如果Value为引用类型且有脏标记，则清除所有xmap中Value的脏标记\n\n")
		sb.WriteString("\t// 清除所有xmap中的操作记录\n")
		// 只处理值类型的 xmap 字段（普通map不需要增量策略）
		for _, field := range mech.Fields {
			if field.IsXMapField() && !field.Type.IsXMapValueMMEObject() {
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

		for _, field := range mech.Fields {
			dirtyBitName := fmt.Sprintf("%sDirty%sBit", mech.Name, field.Name)
			fieldNameSnake := CamelToSnake(field.Name)

			if field.Type.IsXMapField() {
				sb.WriteString(fmt.Sprintf("\tif w.IsDirty(%s) {\n", dirtyBitName))
				if !field.Type.IsXMapValueMMEObject() {
					accessorName := strings.ToLower(field.Name[0:1]) + field.Name[1:] + "Accessor"
					sb.WriteString(fmt.Sprintf("\t\tbuilder.SetNestedPath(path, \"%s\", w.%s.Clone())\n",
						fieldNameSnake, accessorName))
				} else {
					sb.WriteString(fmt.Sprintf("\t\t// TODO: Handle MME Object map field %s\n", field.Name))
				}
				sb.WriteString("\t}\n")
			} else if field.Type.Kind == types.FieldKindMap {
				// 普通map字段，直接使用Get方法
				sb.WriteString(fmt.Sprintf("\tif w.IsDirty(%s) {\n", dirtyBitName))
				methodName := strings.ToUpper(field.Name[0:1]) + field.Name[1:]
				sb.WriteString(fmt.Sprintf("\t\tbuilder.SetNestedPath(path, \"%s\", w.Get%s())\n",
					fieldNameSnake, methodName))
				sb.WriteString("\t}\n")
			} else {
				sb.WriteString(fmt.Sprintf("\tif w.IsDirty(%s) {\n", dirtyBitName))
				methodName := strings.ToUpper(field.Name[0:1]) + field.Name[1:]
				sb.WriteString(fmt.Sprintf("\t\tbuilder.SetNestedPath(path, \"%s\", w.Get%s())\n",
					fieldNameSnake, methodName))
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

		for _, field := range mech.Fields {
			fieldIndexName := fmt.Sprintf("%sFieldIndex%s", mech.Name, field.Name)
			dirtyBitName := fmt.Sprintf("%sDirty%sBit", mech.Name, field.Name)

			if field.Type.IsXMapField() {
				// XMap 类型使用增量同步逻辑
				// 所有 xmap 字段都使用 mmemodel.FieldCanBeIncrementalSynced 快捷方法进行脏标记判断
				sb.WriteString(fmt.Sprintf("\tif mmemodel.FieldCanBeIncrementalSynced(w, %s, %s, ctx) {\n",
					dirtyBitName, fieldIndexName))

				accessorName := strings.ToLower(field.Name[0:1]) + field.Name[1:] + "Accessor"

				if !g.isReferenceType(field.Type.ValueType) {
					// 值类型，直接调用 Clone
					sb.WriteString(fmt.Sprintf("\t\tincremental.%s = w.%s.Clone()\n",
						field.Name, accessorName))
				} else {
					// 引用类型，留 TODO
					sb.WriteString(fmt.Sprintf("\t\t// TODO: Handle reference type map field %s (full sync)\n", field.Name))
				}
				sb.WriteString("\t}\n")
			} else if field.Type.IsMapField() {
				// 普通 Map 类型直接全量同步，不使用增量同步逻辑
				// 所有 map 字段都使用 mmemodel.FieldCanBeIncrementalSynced 快捷方法进行脏标记判断
				sb.WriteString(fmt.Sprintf("\tif mmemodel.FieldCanBeIncrementalSynced(w, %s, %s, ctx) {\n",
					dirtyBitName, fieldIndexName))
				methodName := strings.ToUpper(field.Name[0:1]) + field.Name[1:]
				// 普通map直接返回map的副本
				keyType := ToGoBaseTypeFromFieldType(field.Type.KeyType)
				valueType := ToGoBaseTypeFromFieldType(field.Type.ValueType)
				sb.WriteString(fmt.Sprintf("\t\tm := w.Get%s()\n", methodName))
				sb.WriteString("\t\tif m != nil {\n")
				sb.WriteString(fmt.Sprintf("\t\t\tincremental.%s = make(map[%s]%s, len(m))\n",
					field.Name, keyType, valueType))
				sb.WriteString(fmt.Sprintf("\t\t\tmaps.Copy(incremental.%s, m)\n", field.Name))
				sb.WriteString("\t\t}\n")
				sb.WriteString("\t}\n")
			} else {
				// 所有字段都使用 mmemodel.FieldCanBeIncrementalSynced 快捷方法进行脏标记判断
				sb.WriteString(fmt.Sprintf("\tif mmemodel.FieldCanBeIncrementalSynced(w, %s, %s, ctx) {\n",
					dirtyBitName, fieldIndexName))
				methodName := strings.ToUpper(field.Name[0:1]) + field.Name[1:]
				sb.WriteString(fmt.Sprintf("\t\tv := w.Get%s()\n", methodName))
				sb.WriteString(fmt.Sprintf("\t\tincremental.%s = &v\n", field.Name))
				sb.WriteString("\t}\n")
			}
		}

		sb.WriteString("\treturn incremental\n")
		sb.WriteString("}\n\n")

		// 生成 ToProto 方法
		sb.WriteString("// ToProto 将 Mechanism 数据转换为完整的 protobuf 结构体\n")
		sb.WriteString(fmt.Sprintf("func (w *%s) ToProto() *mme.%s {\n", wrapperName, mech.Name))
		sb.WriteString("\tif w == nil || w.data == nil {\n")
		sb.WriteString("\t\treturn nil\n")
		sb.WriteString("\t}\n\n")
		sb.WriteString("\treturn w.data.ToProto()\n")
		sb.WriteString("}\n\n")

		// 生成 FromProto 方法
		sb.WriteString("// FromProto 从 protobuf 结构体加载数据到 Mechanism\n")
		sb.WriteString(fmt.Sprintf("func (w *%s) FromProto(pb *mme.%s) {\n", wrapperName, mech.Name))
		sb.WriteString("\tif w == nil || w.data == nil || pb == nil {\n")
		sb.WriteString("\t\treturn\n")
		sb.WriteString("\t}\n\n")

		// 按字段编号排序，保证与 YAML 中定义的顺序一致
		sortedFields := make([]*types.Field, len(mech.Fields))
		copy(sortedFields, mech.Fields)
		// 冒泡排序按 Number 字段排序
		for i := 0; i < len(sortedFields)-1; i++ {
			for j := i + 1; j < len(sortedFields); j++ {
				if sortedFields[i].Number > sortedFields[j].Number {
					sortedFields[i], sortedFields[j] = sortedFields[j], sortedFields[i]
				}
			}
		}

		for _, field := range sortedFields {
			if field.Type.IsXMapField() {
				// XMap 字段需要特殊处理（使用accessor）
				sb.WriteString(fmt.Sprintf("\t// 加载 %s xmap\n", field.Name))
				sb.WriteString(fmt.Sprintf("\tif pb.%s != nil {\n", field.Name))

				if !field.Type.IsXMapValueMMEObject() {
					// 值类型 xmap，使用 accessor 的 Reset 方法
					accessorName := strings.ToLower(field.Name[0:1]) + field.Name[1:] + "Accessor"
					sb.WriteString("\t\t// 清空现有的 map\n")
					keyType := ToGoBaseTypeFromFieldType(field.Type.KeyType)
					valueType := ToGoBaseTypeFromFieldType(field.Type.ValueType)
					sb.WriteString(fmt.Sprintf("\t\ttemp := make(map[%s]%s, len(pb.%s))\n",
						keyType, valueType, field.Name))
					sb.WriteString(fmt.Sprintf("\t\tmaps.Copy(temp, pb.%s)\n", field.Name))
					sb.WriteString("\t\t// 设置新的 map 数据，并清空所有变化操作记录\n")
					sb.WriteString(fmt.Sprintf("\t\tw.%s.Reset(&temp)\n", accessorName))
				} else {
					// MME Object 类型 xmap，需要递归处理
					sb.WriteString(fmt.Sprintf("\t\t// TODO: Handle MME Object map field %s\n", field.Name))
				}
				sb.WriteString("\t}\n")
			} else if field.Type.Kind == types.FieldKindMap {
				// 普通 Map 字段，直接使用 Setter 方法
				methodName := strings.ToUpper(field.Name[0:1]) + field.Name[1:]
				sb.WriteString(fmt.Sprintf("\t// 加载 %s map\n", field.Name))
				sb.WriteString(fmt.Sprintf("\tif pb.%s != nil {\n", field.Name))
				keyType := ToGoBaseTypeFromFieldType(field.Type.KeyType)
				valueType := ToGoBaseTypeFromFieldType(field.Type.ValueType)
				sb.WriteString(fmt.Sprintf("\t\ttemp := make(map[%s]%s, len(pb.%s))\n",
					keyType, valueType, field.Name))
				sb.WriteString(fmt.Sprintf("\t\tmaps.Copy(temp, pb.%s)\n", field.Name))
				sb.WriteString(fmt.Sprintf("\t\tw.Set%s(temp)\n", methodName))
				sb.WriteString("\t}\n")
			} else {
				// 普通字段，使用 Setter 方法
				methodName := strings.ToUpper(field.Name[0:1]) + field.Name[1:]
				sb.WriteString(fmt.Sprintf("\t// 加载基础类型字段 %s\n", field.Name))
				sb.WriteString(fmt.Sprintf("\tif pb.%s != nil {\n", field.Name))
				sb.WriteString(fmt.Sprintf("\t\tw.Set%s(*pb.%s)\n", methodName, field.Name))
				sb.WriteString("\t}\n")
			}
		}

		sb.WriteString("}\n\n")

		// 生成 DeepCopy 方法
		sb.WriteString("// 包装器-深拷贝\n")
		sb.WriteString(fmt.Sprintf("func (w *%s) DeepCopy() *%s {\n", wrapperName, mech.Name))
		sb.WriteString("\tif w == nil {\n")
		sb.WriteString("\t\treturn nil\n")
		sb.WriteString("\t}\n")
		sb.WriteString(fmt.Sprintf("\tcopy := &%s{}\n", mech.Name))
		sb.WriteString("\tw.DeepCopyTo(copy)\n")
		sb.WriteString("\treturn copy\n")
		sb.WriteString("}\n\n")

		// 生成 DeepCopyTo 方法
		sb.WriteString("// 包装器-深拷贝\n")
		sb.WriteString(fmt.Sprintf("func (w *%s) DeepCopyTo(copy *%s) {\n", wrapperName, mech.Name))
		sb.WriteString("\tif w == nil || w.data == nil || copy == nil {\n")
		sb.WriteString("\t\treturn\n")
		sb.WriteString("\t}\n\n")
		sb.WriteString("\t// 拷贝基础类型字段\n")
		sb.WriteString("\t*copy = *w.data\n\n")

		// 处理 map/xmap 字段
		for _, field := range mech.Fields {
			if field.Type.IsXMapField() {
				accessorName := strings.ToLower(field.Name[0:1]) + field.Name[1:] + "Accessor"
				if !field.Type.IsXMapValueMMEObject() {
					// 值类型 xmap，使用 accessor.Copy() 方法
					sb.WriteString("\t// Value 为值类型，浅拷贝即可\n")
					sb.WriteString(fmt.Sprintf("\tw.%s.Copy(copy.%s)\n", accessorName, field.Name))
				} else {
					// MME Object 类型 xmap，需要深拷贝每个元素
					keyType := ToGoBaseTypeFromFieldType(field.Type.KeyType)
					valueType := strings.TrimPrefix(ToGoBaseTypeFromFieldType(field.Type.ValueType), "*")
					sb.WriteString("\t// Value 为引用类型，需要深拷贝\n")
					sb.WriteString(fmt.Sprintf("\tif copy.%s == nil {\n", field.Name))
					sb.WriteString(fmt.Sprintf("\t\tcopy.%s = make(map[%s]*%s)\n", field.Name, keyType, valueType))
					sb.WriteString("\t}\n")
					sb.WriteString(fmt.Sprintf("\tfor k, v := range w.data.%s {\n", field.Name))
					sb.WriteString("\t\tif v != nil {\n")
					sb.WriteString(fmt.Sprintf("\t\t\tcopy.%s[k] = &%s{}\n", field.Name, valueType))
					sb.WriteString(fmt.Sprintf("\t\t\tv.DeepCopy(copy.%s[k])\n", field.Name))
					sb.WriteString("\t\t}\n")
					sb.WriteString("\t}\n")
				}
			} else if field.Type.IsMapField() {
				// 普通 map 字段，直接使用 maps.Copy
				if !field.Type.IsMapValueMMEObject() {
					// 值类型 map，使用 maps.Copy
					sb.WriteString("\t// Value 为值类型，浅拷贝即可\n")
					sb.WriteString(fmt.Sprintf("\tif w.data.%s != nil {\n", field.Name))
					sb.WriteString(fmt.Sprintf("\t\tif copy.%s == nil {\n", field.Name))
					keyType := ToGoBaseTypeFromFieldType(field.Type.KeyType)
					valueType := ToGoBaseTypeFromFieldType(field.Type.ValueType)
					sb.WriteString(fmt.Sprintf("\t\t\tcopy.%s = make(map[%s]%s, len(w.data.%s))\n",
						field.Name, keyType, valueType, field.Name))
					sb.WriteString("\t\t}\n")
					sb.WriteString(fmt.Sprintf("\t\tmaps.Copy(copy.%s, w.data.%s)\n", field.Name, field.Name))
					sb.WriteString("\t}\n")
				} else {
					// MME Object 类型 map，需要深拷贝每个元素
					keyType := ToGoBaseTypeFromFieldType(field.Type.KeyType)
					valueType := strings.TrimPrefix(ToGoBaseTypeFromFieldType(field.Type.ValueType), "*")
					sb.WriteString("\t// Value 为引用类型，需要深拷贝\n")
					sb.WriteString(fmt.Sprintf("\tif copy.%s == nil {\n", field.Name))
					sb.WriteString(fmt.Sprintf("\t\tcopy.%s = make(map[%s]*%s)\n", field.Name, keyType, valueType))
					sb.WriteString("\t}\n")
					sb.WriteString(fmt.Sprintf("\tfor k, v := range w.data.%s {\n", field.Name))
					sb.WriteString("\t\tif v != nil {\n")
					sb.WriteString(fmt.Sprintf("\t\t\tcopy.%s[k] = &%s{}\n", field.Name, valueType))
					sb.WriteString(fmt.Sprintf("\t\t\tv.DeepCopy(copy.%s[k])\n", field.Name))
					sb.WriteString("\t\t}\n")
					sb.WriteString("\t}\n")
				}
			}
		}
		sb.WriteString("}\n\n")

		// 写入文件（追加到机制文件）
		fileName := GetMechanismFileName(mech.Name)
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

// isReferenceType 判断是否是引用类型（MME Object 或 Message 类型）
func (g *GoWrapperGenerator) isReferenceType(fieldType *types.FieldType) bool {
	if fieldType == nil {
		return false
	}

	// 如果是 Map/XMap，检查 ValueType
	if fieldType.Kind == types.FieldKindMap || fieldType.Kind == types.FieldKindXMap {
		if fieldType.ValueType != nil {
			return g.isReferenceType(fieldType.ValueType)
		}
		return false
	}

	// Message 类型和 MME Object 类型都是引用类型
	return fieldType.Kind == types.FieldKindMessage || fieldType.Kind == types.FieldKindMMEObject || fieldType.IsMMEObjectType()
}

// getTypeNameFromFieldType 从 FieldType 获取类型名称字符串
func getTypeNameFromFieldType(ft *types.FieldType) string {
	if ft == nil {
		return "unknown"
	}
	if ft.TypeName != "" {
		return ft.TypeName
	}
	return ft.Kind.String()
}

// ToGoTypeFromTypesFieldType 将新的 types.FieldType 转换为 Go 类型
func ToGoTypeFromTypesFieldType(ft *types.FieldType, packageName string) string {
	if ft == nil {
		return "unknown"
	}

	switch ft.Kind {
	case types.FieldKindRepeated:
		if ft.ValueType != nil {
			return "[]" + ToGoTypeFromTypesFieldType(ft.ValueType, packageName)
		}
		return "[]unknown"
	case types.FieldKindMap, types.FieldKindXMap:
		if ft.KeyType != nil && ft.ValueType != nil {
			keyType := ToGoTypeFromTypesFieldType(ft.KeyType, packageName)
			valueType := ToGoTypeFromTypesFieldType(ft.ValueType, packageName)
			return fmt.Sprintf("map[%s]%s", keyType, valueType)
		}
		return "map[unknown]unknown"
	case types.FieldKindInt32, types.FieldKindInt64, types.FieldKindUInt32, types.FieldKindUInt64,
		types.FieldKindFloat, types.FieldKindDouble, types.FieldKindBool, types.FieldKindString, types.FieldKindBytes:
		// 基础类型直接返回值类型，不带指针和包名
		return ToGoBaseTypeFromFieldType(ft)
	case types.FieldKindMMEObject:
		// MME Object 类型，添加包名前缀和指针
		typeName := getTypeNameFromFieldType(ft)
		// 去掉包名前缀，只取类型名
		if idx := strings.LastIndex(typeName, "."); idx >= 0 {
			typeName = typeName[idx+1:]
		}
		return "*" + typeName
	case types.FieldKindMessage:
		// 消息类型
		typeName := getTypeNameFromFieldType(ft)
		if strings.Contains(typeName, ".") {
			// 外部包类型，直接返回
			return typeName
		}
		// 本地消息类型，可能是其他包的类型，保持原样
		return typeName
	default:
		// 其他类型（如 Enum），尝试获取类型名
		typeName := getTypeNameFromFieldType(ft)
		if typeName != "" && typeName != "unknown" {
			return typeName
		}
		return "unknown"
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

// ensureModuleImports 确保Module文件包含必要的导入
// 如果导入不存在，则添加到import块中
func ensureModuleImports(content string) string {
	// 检查是否已经包含必要的导入
	hasDirtyFlag := strings.Contains(content, "dirtyflag")
	hasFieldMeta := strings.Contains(content, "fieldmeta")
	hasMgoBuilder := strings.Contains(content, "mgo_builder")
	hasMmeModel := strings.Contains(content, "mmemodel")
	hasProto := strings.Contains(content, "google.golang.org/protobuf/proto")

	// 如果所有导入都已存在，直接返回
	if hasDirtyFlag && hasFieldMeta && hasMgoBuilder && hasMmeModel && hasProto {
		return content
	}

	// 查找 import 块的位置
	importStart := strings.Index(content, "import (")
	if importStart == -1 {
		// 如果没有 import 块，在 package 声明后添加
		packageEnd := strings.Index(content, "\n\n")
		if packageEnd == -1 {
			packageEnd = len(content)
		}
		imports := "\nimport (\n"
		imports += "\t\"gitee.com/orbit-w/orbit/app/proto/mme\"\n"
		if !hasDirtyFlag {
			imports += "\tdirtyflag \"gitee.com/orbit-w/orbit/lib/base/dirty_flag\"\n"
		}
		if !hasFieldMeta {
			imports += "\tfieldmeta \"gitee.com/orbit-w/orbit/lib/base/field_meta\"\n"
		}
		if !hasMgoBuilder {
			imports += "\t\"gitee.com/orbit-w/orbit/lib/module/db/mgo_builder\"\n"
		}
		if !hasMmeModel {
			imports += "\tmmemodel \"gitee.com/orbit-w/orbit/lib/module/mme_model\"\n"
		}
		if !hasProto {
			imports += "\t\"google.golang.org/protobuf/proto\"\n"
		}
		imports += ")\n"
		return content[:packageEnd+2] + imports + content[packageEnd+2:]
	}

	// 找到 import 块的结束位置
	importEnd := strings.Index(content[importStart:], ")\n")
	if importEnd == -1 {
		importEnd = strings.Index(content[importStart:], ")\r\n")
	}
	if importEnd == -1 {
		return content
	}
	importEnd += importStart + 2

	// 在 import 块中添加缺失的导入
	importBlock := content[importStart:importEnd]
	lines := strings.Split(importBlock, "\n")

	// 检查每一行，添加缺失的导入
	missingImports := []string{}
	if !hasDirtyFlag {
		missingImports = append(missingImports, "\tdirtyflag \"gitee.com/orbit-w/orbit/lib/base/dirty_flag\"")
	}
	if !hasFieldMeta {
		missingImports = append(missingImports, "\tfieldmeta \"gitee.com/orbit-w/orbit/lib/base/field_meta\"")
	}
	if !hasMgoBuilder {
		missingImports = append(missingImports, "\t\"gitee.com/orbit-w/orbit/lib/module/db/mgo_builder\"")
	}
	if !hasMmeModel {
		missingImports = append(missingImports, "\tmmemodel \"gitee.com/orbit-w/orbit/lib/module/mme_model\"")
	}
	if !hasProto {
		missingImports = append(missingImports, "\t\"google.golang.org/protobuf/proto\"")
	}

	if len(missingImports) == 0 {
		return content
	}

	// 在 import 块的倒数第二行（在 ")" 之前）插入缺失的导入
	newImportBlock := strings.Join(lines[:len(lines)-1], "\n")
	for _, imp := range missingImports {
		newImportBlock += "\n" + imp
	}
	newImportBlock += "\n" + lines[len(lines)-1]

	return content[:importStart] + newImportBlock + content[importEnd:]
}

// generateModuleWrappers 生成 Module Wrapper
func (g *GoWrapperGenerator) generateModuleWrappers(outputDir string) error {
	packageName := "mme"

	for _, module := range g.data.Modules {
		sb := strings.Builder{}

		// Wrapper 代码（不包含 package 和 import，因为这些会从已有文件中获取）

		// 生成 Wrapper 结构体
		wrapperName := module.Name + "Wrapper"
		sb.WriteString(fmt.Sprintf("type %s struct {\n", wrapperName))
		sb.WriteString(fmt.Sprintf("\tdata *%s\n", module.Name))
		sb.WriteString("\tdirtyflag.IDirtyFlag\n")
		sb.WriteString("\tfieldMetas *fieldmeta.FieldMetas\n\n")

		// 生成嵌套的 Mechanism Wrapper 字段
		for _, field := range module.Fields {
			if field.Type.IsMMEObjectType() {
				typeName := strings.TrimPrefix(ToGoBaseTypeFromFieldType(&field.Type), "*")
				wrapperTypeName := typeName + "Wrapper"
				sb.WriteString(fmt.Sprintf("\t%s *%s\n", field.Name+"Wrapper", wrapperTypeName))
			}
		}

		sb.WriteString("}\n\n")

		// 生成 New 构造函数
		sb.WriteString(fmt.Sprintf("func New%s(data *%s) *%s {\n", wrapperName, module.Name, wrapperName))
		sb.WriteString("\tif data == nil {\n")
		sb.WriteString("\t\tpanic(\"data is nil\")\n")
		sb.WriteString("\t}\n")
		sb.WriteString(fmt.Sprintf("\tw := &%s{\n", wrapperName))
		sb.WriteString("\t\tdata:       data,\n")
		sb.WriteString("\t\tIDirtyFlag: dirtyflag.NewDirtyFlag(),\n")
		sb.WriteString("\t\tfieldMetas: fieldmeta.NewFieldMetas(),\n")
		sb.WriteString("\t}\n\n")

		// 初始化嵌套的 Mechanism Wrapper
		for _, field := range module.Fields {
			if field.Type.IsMMEObjectType() {
				typeName := strings.TrimPrefix(ToGoBaseTypeFromFieldType(&field.Type), "*")
				wrapperFieldName := field.Name + "Wrapper"
				dirtyBitName := fmt.Sprintf("%sDirty%sBit", module.Name, field.Name)
				sb.WriteString(fmt.Sprintf("\t// 初始化嵌套的 %s Wrapper\n", typeName))
				sb.WriteString(fmt.Sprintf("\tif data.%s == nil {\n", field.Name))
				sb.WriteString(fmt.Sprintf("\t\tdata.%s = New%s()\n", field.Name, typeName))
				sb.WriteString("\t}\n")
				sb.WriteString(fmt.Sprintf("\tw.%s = New%sWrapper(data.%s)\n",
					wrapperFieldName, typeName, field.Name))
				sb.WriteString(fmt.Sprintf("\tw.%s.Link(w.GetDirtyTracker(), %s)\n",
					wrapperFieldName, dirtyBitName))
				sb.WriteString("\n")
			}
		}

		sb.WriteString("\treturn w\n")
		sb.WriteString("}\n\n")

		// 生成 InitFieldContext 方法
		sb.WriteString(fmt.Sprintf("func (w *%s) InitFieldContext() {\n", wrapperName))
		for _, field := range module.Fields {
			fieldIndexName := fmt.Sprintf("%sFieldIndex%s", module.Name, field.Name)
			// 根据 access 选项设置字段类型
			if field.Options.Access == "all" || field.Options.Access == "s" || field.Options.Access == "" {
				sb.WriteString(fmt.Sprintf("\tw.fieldMetas.SetFieldType(%s, fieldmeta.FieldTypeSync)\n",
					fieldIndexName))
			}
		}
		sb.WriteString("\n")
		// 初始化嵌套 Mechanism 的字段上下文
		for _, field := range module.Fields {
			if field.Type.IsMMEObjectType() {
				wrapperFieldName := field.Name + "Wrapper"
				sb.WriteString(fmt.Sprintf("\tif w.%s != nil {\n", wrapperFieldName))
				sb.WriteString(fmt.Sprintf("\t\tw.%s.InitFieldContext()\n", wrapperFieldName))
				sb.WriteString("\t}\n")
			}
		}
		sb.WriteString("}\n\n")

		// 生成 Name 方法
		sb.WriteString(fmt.Sprintf("func (w *%s) Name() string {\n", wrapperName))
		sb.WriteString(fmt.Sprintf("\treturn \"%s\"\n", module.Name))
		sb.WriteString("}\n\n")

		// 生成 MatchesAll 方法（实现 IFieldMetaContext 接口）
		sb.WriteString("// MatchesAll 判断字段是否匹配所有类型标记\n")
		sb.WriteString(fmt.Sprintf("func (w *%s) MatchesAll(fieldID uint8, fieldTypes ...fieldmeta.FieldType) bool {\n", wrapperName))
		sb.WriteString("\treturn w.fieldMetas.MatchesAll(fieldID, fieldTypes...)\n")
		sb.WriteString("}\n\n")

		// 生成 Getter 方法（获取嵌套的 Mechanism Wrapper）
		for _, field := range module.Fields {
			if field.Type.IsMMEObjectType() {
				typeName := strings.TrimPrefix(ToGoBaseTypeFromFieldType(&field.Type), "*")
				wrapperFieldName := field.Name + "Wrapper"
				methodName := strings.ToUpper(field.Name[0:1]) + field.Name[1:]
				sb.WriteString(fmt.Sprintf("func (w *%s) Get%s() *%sWrapper {\n",
					wrapperName, methodName, typeName))
				sb.WriteString(fmt.Sprintf("\treturn w.%s\n", wrapperFieldName))
				sb.WriteString("}\n\n")
			}
		}

		// 生成 ClearAllDirtyFlags 方法
		sb.WriteString("// ClearAllDirtyFlags 清除所有脏标记位\n")
		sb.WriteString(fmt.Sprintf("func (w *%s) ClearAllDirtyFlags() {\n", wrapperName))
		sb.WriteString("\tw.ClearAllDirty()\n\n")
		sb.WriteString("\t// 清除嵌套 Mechanism 的脏标记\n")
		for _, field := range module.Fields {
			if field.Type.IsMMEObjectType() {
				wrapperFieldName := field.Name + "Wrapper"
				sb.WriteString(fmt.Sprintf("\tif w.%s != nil {\n", wrapperFieldName))
				sb.WriteString(fmt.Sprintf("\t\tw.%s.ClearAllDirtyFlags()\n", wrapperFieldName))
				sb.WriteString("\t}\n")
			}
		}
		sb.WriteString("}\n\n")

		// 生成 BuildMongoUpdate 方法
		sb.WriteString("// BuildMongoUpdate 构建MongoDB更新操作\n")
		sb.WriteString(fmt.Sprintf("func (w *%s) BuildMongoUpdate(builder *mgo_builder.MongoUpdateBuilder, path *mgo_builder.NestedPath) {\n", wrapperName))
		sb.WriteString("\tif w == nil {\n")
		sb.WriteString("\t\treturn\n")
		sb.WriteString("\t}\n\n")
		sb.WriteString("\t// 构建嵌套字段路径\n")
		for _, field := range module.Fields {
			if field.Type.IsMMEObjectType() {
				wrapperFieldName := field.Name + "Wrapper"
				dirtyBitName := fmt.Sprintf("%sDirty%sBit", module.Name, field.Name)
				fieldNameSnake := CamelToSnake(field.Name)
				sb.WriteString(fmt.Sprintf("\tif w.%s != nil && w.IsDirty(%s) {\n",
					wrapperFieldName, dirtyBitName))
				sb.WriteString(fmt.Sprintf("\t\tw.%s.BuildMongoUpdate(builder, path.Field(\"%s\"))\n",
					wrapperFieldName, fieldNameSnake))
				sb.WriteString("\t}\n")
			}
		}
		sb.WriteString("}\n\n")

		// 生成 DeepCopy 方法
		sb.WriteString("// 包装器-深拷贝\n")
		sb.WriteString(fmt.Sprintf("func (w *%s) DeepCopy() *%s {\n", wrapperName, module.Name))
		sb.WriteString("\tif w == nil {\n")
		sb.WriteString("\t\treturn nil\n")
		sb.WriteString("\t}\n")
		sb.WriteString(fmt.Sprintf("\tcopy := &%s{}\n", module.Name))
		sb.WriteString("\tw.DeepCopyTo(copy)\n")
		sb.WriteString("\treturn copy\n")
		sb.WriteString("}\n\n")

		// 生成 DeepCopyTo 方法
		sb.WriteString("// 包装器-深拷贝\n")
		sb.WriteString(fmt.Sprintf("func (w *%s) DeepCopyTo(copy *%s) {\n", wrapperName, module.Name))
		sb.WriteString("\tif w == nil || w.data == nil || copy == nil {\n")
		sb.WriteString("\t\treturn\n")
		sb.WriteString("\t}\n\n")
		sb.WriteString("\t// 拷贝嵌套的 Mechanism 对象\n")
		for _, field := range module.Fields {
			if field.Type.IsMMEObjectType() {
				typeName := strings.TrimPrefix(ToGoBaseTypeFromFieldType(&field.Type), "*")
				wrapperFieldName := field.Name + "Wrapper"
				sb.WriteString(fmt.Sprintf("\tif w.data.%s != nil {\n", field.Name))
				sb.WriteString(fmt.Sprintf("\t\tif copy.%s == nil {\n", field.Name))
				sb.WriteString(fmt.Sprintf("\t\t\tcopy.%s = &%s{}\n", field.Name, typeName))
				sb.WriteString("\t\t}\n")
				sb.WriteString(fmt.Sprintf("\t\tw.%s.DeepCopyTo(copy.%s)\n", wrapperFieldName, field.Name))
				sb.WriteString("\t}\n")
			}
		}
		sb.WriteString("}\n\n")

		// 生成 ToProto 方法
		sb.WriteString("// ToProto 将 Module 数据转换为完整的 protobuf 结构体\n")
		sb.WriteString(fmt.Sprintf("func (w *%s) ToProto() *mme.%s {\n", wrapperName, module.Name))
		sb.WriteString("\tif w == nil || w.data == nil {\n")
		sb.WriteString("\t\treturn nil\n")
		sb.WriteString("\t}\n\n")
		sb.WriteString("\treturn w.data.ToProto()\n")
		sb.WriteString("}\n\n")

		// 生成 FromProto 方法
		sb.WriteString("// FromProto 从 protobuf 结构体加载数据到 Module\n")
		sb.WriteString(fmt.Sprintf("func (w *%s) FromProto(pb *mme.%s) {\n", wrapperName, module.Name))
		sb.WriteString("\tif w == nil || w.data == nil || pb == nil {\n")
		sb.WriteString("\t\treturn\n")
		sb.WriteString("\t}\n\n")
		for _, field := range module.Fields {
			if field.Type.IsMMEObjectType() {
				typeName := strings.TrimPrefix(ToGoBaseTypeFromFieldType(&field.Type), "*")
				wrapperFieldName := field.Name + "Wrapper"
				sb.WriteString(fmt.Sprintf("\tif pb.%s != nil {\n", field.Name))
				sb.WriteString(fmt.Sprintf("\t\tif w.data.%s == nil {\n", field.Name))
				sb.WriteString(fmt.Sprintf("\t\t\tw.data.%s = New%s()\n", field.Name, typeName))
				sb.WriteString("\t\t}\n")
				sb.WriteString(fmt.Sprintf("\t\tw.%s.FromProto(pb.%s)\n", wrapperFieldName, field.Name))
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
		sb.WriteString(fmt.Sprintf("\tincremental := &mme.%s{}\n\n", module.Name))

		for _, field := range module.Fields {
			fieldIndexName := fmt.Sprintf("%sFieldIndex%s", module.Name, field.Name)
			dirtyBitName := fmt.Sprintf("%sDirty%sBit", module.Name, field.Name)
			if field.Type.IsMMEObjectType() {
				wrapperFieldName := field.Name + "Wrapper"
				// 获取 proto 类型名称
				typeName := strings.TrimPrefix(ToGoBaseTypeFromFieldType(&field.Type), "*")
				protoTypeName := fmt.Sprintf("*mme.%s", typeName)
				sb.WriteString(fmt.Sprintf("\tif mmemodel.FieldCanBeIncrementalSynced(w, %s, %s, ctx) {\n",
					dirtyBitName, fieldIndexName))
				sb.WriteString(fmt.Sprintf("\t\tif w.%s != nil {\n", wrapperFieldName))
				sb.WriteString(fmt.Sprintf("\t\t\tpb := w.%s.ToIncrementalProtoWithContext(ctx)\n",
					wrapperFieldName))
				sb.WriteString("\t\t\tif pb != nil {\n")
				sb.WriteString(fmt.Sprintf("\t\t\t\tv, ok := pb.(%s)\n", protoTypeName))
				sb.WriteString("\t\t\t\tif ok {\n")
				sb.WriteString(fmt.Sprintf("\t\t\t\t\tincremental.%s = v\n", field.Name))
				sb.WriteString("\t\t\t\t}\n")
				sb.WriteString("\t\t\t}\n")
				sb.WriteString("\t\t}\n")
				sb.WriteString("\t}\n")
			}
		}

		sb.WriteString("\treturn incremental\n")
		sb.WriteString("}\n\n")

		// 写入文件（追加到模块文件）
		fileName := GetModuleFileName(module.Name)
		filePath := fmt.Sprintf("%s/%s", outputDir, fileName)

		// 读取现有文件内容（如果存在）
		existingContent := ""
		if FileExists(filePath) {
			if data, err := ReadFileContent(filePath); err == nil {
				existingContent = data
				// 检查并确保必要的导入存在
				existingContent = ensureModuleImports(existingContent)
			}
		}

		// 追加新内容（如果文件已存在，在末尾添加换行）
		var newContent string
		if existingContent != "" {
			// 确保现有内容以换行结尾
			existingContent = strings.TrimRight(existingContent, " \n\r\t")
			newContent = existingContent + "\n\n" + sb.String()
		} else {
			// 如果文件不存在，需要添加 package 声明和导入
			imports := "import (\n"
			imports += "\t\"gitee.com/orbit-w/orbit/app/proto/mme\"\n"
			imports += "\tdirtyflag \"gitee.com/orbit-w/orbit/lib/base/dirty_flag\"\n"
			imports += "\tfieldmeta \"gitee.com/orbit-w/orbit/lib/base/field_meta\"\n"
			imports += "\t\"gitee.com/orbit-w/orbit/lib/module/db/mgo_builder\"\n"
			imports += "\tmmemodel \"gitee.com/orbit-w/orbit/lib/module/mme_model\"\n"
			imports += "\t\"google.golang.org/protobuf/proto\"\n"
			imports += ")\n\n"
			newContent = fmt.Sprintf("package %s\n\n%s", packageName, imports) + sb.String()
		}
		if err := WriteFile(filePath, newContent); err != nil {
			return err
		}
	}

	return nil
}
