package blueprint_gen

import (
	"fmt"
	"strings"

	blueprint_types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"
)

// generateNetWallProto 生成 NetWall Proto 文件
func (g *ProtoGenerator) generateNetWallProto(outputDir string) error {
	for _, netwallFile := range g.data.NetWalls {
		err := g.generateNetWallProtoFile(outputDir, netwallFile)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *ProtoGenerator) generateNetWallProtoFile(outputDir string, netwallFile *NetWallFile) error {
	sb := strings.Builder{}

	// 检测跨 NetWall 的引用
	imports := g.collectNetWallImports(netwallFile)

	// 生成文件头部
	packageName := netwallFile.PackageName
	if packageName == "" {
		return fmt.Errorf("NetWall package name is empty")
	}
	sb.WriteString(g.generateProtoHeader(packageName, imports))

	// 生成 Request message
	if len(netwallFile.Requests) > 0 {
		sb.WriteString("message Request {\n\n")
		for _, req := range netwallFile.Requests {
			sb.WriteString(g.generateNetMessageProto(req, "    "))
		}
		sb.WriteString("}\n\n")
	}

	// 生成 Notify message
	if len(netwallFile.Notifies) > 0 {
		sb.WriteString("message Notify {\n\n")
		for _, notify := range netwallFile.Notifies {
			sb.WriteString(g.generateNetMessageProto(notify, "    "))
		}
		sb.WriteString("}\n\n")
	}

	// 生成 DataStructs
	for _, ds := range netwallFile.DataStructs {
		// 添加注释
		if ds.Comment != "" {
			sb.WriteString(fmt.Sprintf("// %s\n", ds.Comment))
		}
		sb.WriteString(fmt.Sprintf("message %s {\n", ds.Name))
		if len(ds.Fields) > 0 {
			// 按编号排序字段
			fields := make([]*blueprint_types.Field, len(ds.Fields))
			copy(fields, ds.Fields)
			for i := 0; i < len(fields)-1; i++ {
				for j := i + 1; j < len(fields); j++ {
					if fields[i].Number > fields[j].Number {
						fields[i], fields[j] = fields[j], fields[i]
					}
				}
			}

			for _, field := range fields {
				// 添加注释
				if field.Comment != "" {
					sb.WriteString(fmt.Sprintf("    // %s\n", field.Comment))
				}

				// 生成字段类型
				protoType := g.typeConverter.ToProtoType(&field.Type)
				isOptional := g.typeConverter.IsOptionalInProto(&field.Type)

				// 如果是Message类型，则需要添加包名
				if field.Type.IsMessage() {
					protoPackageName, ok := g.data.NameSpaces[field.Type.Name]
					if ok && protoPackageName != packageName {
						protoType = fmt.Sprintf("%s.%s", protoPackageName, field.Type.Name)
					}
				}

				// 如果是Enum类型，则需要添加包名（Enum包）
				if field.Type.Kind == blueprint_types.FieldKindEnum {
					nameSpaces := g.data.GetNameSpace()
					protoPackageName, ok := nameSpaces[field.Type.Name]
					if ok && protoPackageName != packageName {
						protoType = fmt.Sprintf("%s.%s", protoPackageName, field.Type.Name)
					}
				}

				// 确保字段名不包含冒号
				fieldName := strings.TrimSuffix(field.Name, ":")
				fieldName = strings.TrimSpace(fieldName)

				// 构建字段定义（使用 4 个空格缩进）
				if isOptional {
					sb.WriteString(fmt.Sprintf("    optional %s %s = %d;\n", protoType, fieldName, field.Number))
				} else {
					sb.WriteString(fmt.Sprintf("    %s %s = %d;\n", protoType, fieldName, field.Number))
				}
			}
		}
		sb.WriteString("}\n\n")
	}

	// 生成来自对应 NetWall 文件的枚举
	for _, enum := range netwallFile.Enums {
		sb.WriteString(generateEnumProto(enum))
	}

	// 生成文件名（首字母小写）
	fileName := strings.ToLower(packageName) + ".proto"
	content := sb.String()
	return WriteFile(outputDir+"/"+fileName, content)
}

// collectNetWallImports 收集跨 NetWall 的引用
func (g *ProtoGenerator) collectNetWallImports(wallFile *NetWallFile) []string {
	imports := make(map[string]bool)
	allMessages := make([]*NetMessage, 0)
	nameSpaces := g.data.GetNameSpace()

	allMessages = append(allMessages, wallFile.Requests...)
	allMessages = append(allMessages, wallFile.Notifies...)
	allMessages = append(allMessages, wallFile.DataStructs...)

	// 遍历所有消息的字段，检测跨 NetWall 引用
	for _, msg := range allMessages {
		for _, field := range msg.Fields {
			name := field.Type.Name
			if name == "" {
				panic(fmt.Sprintf("NetWall %s 的消息 %s 的字段 %s 没有名称", msg.PackageName, msg.Name, field.Name))
			}

			packageName, ok := nameSpaces[name]
			if ok && packageName != msg.PackageName {
				fileName := strings.ToLower(packageName)
				path := fileName + ".proto"
				imports[path] = true
			}
		}
	}

	// 转换为切片并排序
	result := make([]string, 0, len(imports))
	for imp := range imports {
		result = append(result, imp)
	}
	// 简单排序（按字母顺序）
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i] > result[j] {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}

// generateNetMessageProto 生成 NetMessage 的 Proto 定义
func (g *ProtoGenerator) generateNetMessageProto(msg *NetMessage, indent string) string {
	sb := strings.Builder{}

	// 添加注释
	if msg.Comment != "" {
		sb.WriteString(fmt.Sprintf("%s// %s\n", indent, msg.Comment))
	}

	// 生成 message 定义
	sb.WriteString(fmt.Sprintf("%smessage %s {\n", indent, msg.Name))

	// 按编号排序字段
	fields := make([]*blueprint_types.Field, len(msg.Fields))
	copy(fields, msg.Fields)
	for i := 0; i < len(fields)-1; i++ {
		for j := i + 1; j < len(fields); j++ {
			if fields[i].Number > fields[j].Number {
				fields[i], fields[j] = fields[j], fields[i]
			}
		}
	}

	// 生成字段定义
	for _, field := range fields {
		// 添加注释
		if field.Comment != "" {
			sb.WriteString(fmt.Sprintf("%s    // %s\n", indent, field.Comment))
		}

		// 生成字段类型
		protoType := g.typeConverter.ToProtoType(&field.Type)
		isOptional := g.typeConverter.IsOptionalInProto(&field.Type)

		// 如果是Message类型，则需要添加包名
		if field.Type.IsMessage() {
			protoPackageName, ok := g.data.NameSpaces[field.Type.Name]
			if ok && protoPackageName != msg.PackageName {
				protoType = fmt.Sprintf("%s.%s", protoPackageName, field.Type.Name)
			}
		}

		// 如果是Enum类型，则需要添加包名（Enum包）
		if field.Type.Kind == blueprint_types.FieldKindEnum {
			nameSpaces := g.data.GetNameSpace()
			protoPackageName, ok := nameSpaces[field.Type.Name]
			if ok && protoPackageName != msg.PackageName {
				protoType = fmt.Sprintf("%s.%s", protoPackageName, field.Type.Name)
			}
		}

		// 确保字段名不包含冒号
		fieldName := strings.TrimSuffix(field.Name, ":")
		fieldName = strings.TrimSpace(fieldName)

		// 构建字段定义（使用 4 个空格缩进）
		if isOptional {
			sb.WriteString(fmt.Sprintf("%s    optional %s %s = %d;\n", indent, protoType, fieldName, field.Number))
		} else {
			sb.WriteString(fmt.Sprintf("%s    %s %s = %d;\n", indent, protoType, fieldName, field.Number))
		}
	}

	// 生成 Rsp（如果有）
	if msg.Rsp != nil {
		sb.WriteString(fmt.Sprintf("%s    message Rsp {\n", indent))
		// 按编号排序 Rsp 字段
		rspFields := make([]*blueprint_types.Field, len(msg.Rsp.Fields))
		copy(rspFields, msg.Rsp.Fields)
		for i := 0; i < len(rspFields)-1; i++ {
			for j := i + 1; j < len(rspFields); j++ {
				if rspFields[i].Number > rspFields[j].Number {
					rspFields[i], rspFields[j] = rspFields[j], rspFields[i]
				}
			}
		}

		for _, field := range rspFields {
			// 添加注释
			if field.Comment != "" {
				sb.WriteString(fmt.Sprintf("%s        // %s\n", indent, field.Comment))
			}

			// 生成字段类型
			protoType := g.typeConverter.ToProtoType(&field.Type)
			isOptional := g.typeConverter.IsOptionalInProto(&field.Type)

			// 如果是Message类型，则需要添加包名
			if field.Type.IsMessage() {
				protoPackageName, ok := g.data.NameSpaces[field.Type.Name]
				if ok && protoPackageName != msg.PackageName {
					protoType = fmt.Sprintf("%s.%s", protoPackageName, field.Type.Name)
				}
			}

			// 如果是Enum类型，则需要添加包名（Enum包）
			if field.Type.Kind == blueprint_types.FieldKindEnum {
				nameSpaces := g.data.GetNameSpace()
				protoPackageName, ok := nameSpaces[field.Type.Name]
				if ok && protoPackageName != msg.PackageName {
					protoType = fmt.Sprintf("%s.%s", protoPackageName, field.Type.Name)
				}
			}

			// 确保字段名不包含冒号
			fieldName := strings.TrimSuffix(field.Name, ":")
			fieldName = strings.TrimSpace(fieldName)

			// 构建字段定义（使用 8 个空格缩进，因为 Rsp 在 Request 内部）
			if isOptional {
				sb.WriteString(fmt.Sprintf("%s        optional %s %s = %d;\n", indent, protoType, fieldName, field.Number))
			} else {
				sb.WriteString(fmt.Sprintf("%s        %s %s = %d;\n", indent, protoType, fieldName, field.Number))
			}
		}
		sb.WriteString(fmt.Sprintf("%s    }\n", indent))
	}

	sb.WriteString(fmt.Sprintf("%s}\n\n", indent))

	return sb.String()
}
