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
				// 使用通用的字段生成方法，自动处理枚举类型引用
				fieldProto := g.generateFieldProto(field, field.Number, packageName)
				// 将缩进从 2 个空格改为 4 个空格（与 NetWall proto 格式一致）
				fieldProto = strings.ReplaceAll(fieldProto, "  ", "    ")
				sb.WriteString(fieldProto)
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

// collectNetWallImports 收集跨 NetWall 的引用（包括枚举类型）
func (g *ProtoGenerator) collectNetWallImports(wallFile *NetWallFile) []string {
	imports := make(map[string]bool)
	allMessages := make([]*NetMessage, 0)
	nameSpaces := g.data.GetNameSpace()

	allMessages = append(allMessages, wallFile.Requests...)
	allMessages = append(allMessages, wallFile.Notifies...)
	allMessages = append(allMessages, wallFile.DataStructs...)

	// 收集所有字段
	allFields := make([]*blueprint_types.Field, 0)
	for _, msg := range allMessages {
		allFields = append(allFields, msg.Fields...)
		if msg.Rsp != nil {
			allFields = append(allFields, msg.Rsp.Fields...)
		}
	}

	// 收集枚举类型的导入
	enumImports := g.collectEnumImportsFromFields(allFields, wallFile.PackageName)
	for _, imp := range enumImports {
		imports[imp] = true
	}

	// 遍历所有消息的字段，检测跨 NetWall 引用（消息类型）
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
		// 使用通用的字段生成方法，自动处理枚举类型引用
		fieldProto := g.generateFieldProto(field, field.Number, msg.PackageName)
		// 将缩进从 2 个空格改为 4 个空格，并添加消息的缩进
		fieldProto = strings.ReplaceAll(fieldProto, "  ", indent+"    ")
		sb.WriteString(fieldProto)
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
			// 使用通用的字段生成方法，自动处理枚举类型引用
			fieldProto := g.generateFieldProto(field, field.Number, msg.PackageName)
			// 将缩进从 2 个空格改为 8 个空格（Rsp 在 Request 内部），并添加消息的缩进
			fieldProto = strings.ReplaceAll(fieldProto, "  ", indent+"        ")
			sb.WriteString(fieldProto)
		}
		sb.WriteString(fmt.Sprintf("%s    }\n", indent))
	}

	sb.WriteString(fmt.Sprintf("%s}\n\n", indent))

	return sb.String()
}
