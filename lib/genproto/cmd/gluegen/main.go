// Package main provides a tool for generating protocol IDs and glue code from proto message definitions
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/orbit-w/orbit/lib/utils/protoid"
)

var (
	protoDir     = flag.String("proto_dir", "app/proto", "Directory containing .proto files")
	outputDir    = flag.String("output_dir", "app/proto/pb", "Directory for generated files")
	debugMode    = flag.Bool("debug", false, "Enable debug mode")
	quietMode    = flag.Bool("quiet", true, "Quiet mode: only show errors")
	genProtoCode = flag.Bool("gen_proto_code", true, "Generate glue code for proto messages")
	genProtoIDs  = flag.Bool("gen_proto_ids", true, "Generate protocol IDs for proto messages")
)

// ProtocolIDMapping 用于存储协议ID映射
type ProtocolIDMapping struct {
	PackageName string
	MessageIDs  []MessageID
}

// MessageID 用于存储消息ID
type MessageID struct {
	Name string
	ID   uint32
}

// Message 消息结构，用于存储消息定义及其注释
type Message struct {
	Name     string
	FullName string // 包含父消息路径的完整名称
	Comment  string
	Fields   []Field
	Response string // 响应消息名称，如果有的话
}

// Field 字段结构，用于存储字段定义及其注释
type Field struct {
	Name    string
	Type    string
	Index   int
	Comment string
}

// MessageName 用于存储消息名称和完整路径
type MessageName struct {
	Name     string
	FullName string
}

func main() {
	flag.Parse()

	// 确保output目录存在
	if _, err := os.Stat(*outputDir); os.IsNotExist(err) {
		if err := os.MkdirAll(*outputDir, 0755); err != nil {
			fmt.Printf("Failed to create output directory: %v\n", err)
			return
		}
	}

	// 查找所有proto文件
	protoFiles, err := findProtoFiles(*protoDir)
	if err != nil {
		fmt.Printf("Error finding proto files: %v\n", err)
		return
	}

	// 处理每个proto文件
	for _, protoFile := range protoFiles {
		if !*quietMode || *debugMode {
			fmt.Printf("Processing %s...\n", protoFile)
		}

		// 读取proto文件内容
		content, err := os.ReadFile(protoFile)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", protoFile, err)
			continue
		}

		// 解析包名
		packageName := extractPackageName(string(content))
		if packageName == "" {
			fmt.Printf("Package name not found in %s\n", protoFile)
			continue
		}

		if !*quietMode {
			fmt.Printf("Found package name: %s\n", packageName)
		}

		// 生成协议ID（如果需要）
		if *genProtoIDs {
			generateProtocolIDs(string(content), packageName)
		}

		// 生成胶水代码（如果需要）
		if *genProtoCode {
			// 解析Request消息
			requestMessages := parseRequestMessages(string(content))
			if len(requestMessages) > 0 {
				if !*quietMode {
					fmt.Printf("Found %d request messages\n", len(requestMessages))
					if *debugMode {
						for i, msg := range requestMessages {
							fmt.Printf("  Request %d: %s, Response: %s\n", i+1, msg, msg.Response)
						}
					}
				}
				generateRequestGlueCode(requestMessages, packageName, *outputDir)
			} else if !*quietMode {
				fmt.Printf("No request messages found\n")
			}

			// 解析Notify消息
			notifyMessages := parseNotifyMessages(string(content))
			if len(notifyMessages) > 0 {
				if !*quietMode {
					fmt.Printf("Found %d notify messages\n", len(notifyMessages))
					if *debugMode {
						for i, msg := range notifyMessages {
							fmt.Printf("  Notify %d: %s\n", i+1, msg.Name)
						}
					}
				}
				generateNotifyGlueCode(notifyMessages, packageName, *outputDir)
			} else if !*quietMode {
				fmt.Printf("No notify messages found\n")
			}
		}
	}

	if !*quietMode {
		fmt.Println("All generation completed!")
	}
}

// generateProtocolIDs 生成协议ID
func generateProtocolIDs(content, packageName string) {
	// 提取所有消息名称
	messageNames := extractMessageNames(content)
	if len(messageNames) == 0 {
		if *debugMode {
			fmt.Printf("No message definitions found for %s\n", packageName)
		}
		return
	}

	// 生成协议ID映射
	mapping := ProtocolIDMapping{
		PackageName: packageName,
	}

	// 按名称排序以得到一致的输出
	sort.Slice(messageNames, func(i, j int) bool {
		return messageNames[i].FullName < messageNames[j].FullName
	})

	// 为每个消息生成ID，过滤掉基本类型和衍生类型
	for _, msg := range messageNames {
		// 跳过基本类型：Request, Rsp, Notify
		if msg.Name == "Request" || msg.Name == "Rsp" || msg.Name == "Notify" {
			if *debugMode {
				fmt.Printf("Skipping base message type: %s\n", msg.Name)
			}
			continue
		}

		// 跳过所有单纯的容器类型消息（不需要协议ID）
		// 例如：Request、Notify 及其可能的衍生结构
		if msg.FullName == "Request" || msg.FullName == "Notify" || msg.FullName == "Rsp" ||
			(strings.HasPrefix(msg.FullName, "Request_") && strings.HasSuffix(msg.FullName, "_Rsp")) {
			if *debugMode {
				fmt.Printf("Skipping container message: %s\n", msg.FullName)
			}
			continue
		}

		fullName := fmt.Sprintf("%s-%s", packageName, msg.FullName)
		pid := protoid.HashProtoMessage(fullName)
		mapping.MessageIDs = append(mapping.MessageIDs, MessageID{
			Name: msg.FullName,
			ID:   pid,
		})

		if *debugMode {
			fmt.Printf("Generated PID for message: %s, ID: 0x%08x\n", fullName, pid)
		}
	}

	if len(mapping.MessageIDs) == 0 {
		if *debugMode {
			fmt.Printf("No protocol IDs generated for package %s after filtering\n", packageName)
		}
		return
	}

	// 生成协议ID文件
	generateProtocolIDFile(mapping)
}

// findProtoFiles 查找指定目录下的所有.proto文件
func findProtoFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".proto") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// extractPackageName 提取包名
func extractPackageName(content string) string {
	re := regexp.MustCompile(`package\s+([^;]+);`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// extractMessageNames 提取所有消息名称，包括嵌套消息，并添加适当的父消息前缀
func extractMessageNames(content string) []MessageName {
	var messageNames []MessageName
	var parentStack []string

	// 使用更复杂的正则表达式匹配消息定义，包括嵌套情况
	messageRegex := regexp.MustCompile(`(?m)^(\s*)message\s+(\w+)\s*\{`)
	closeBraceRegex := regexp.MustCompile(`(?m)^(\s*)\}`)

	lines := strings.Split(content, "\n")
	lineIndex := 0

	for lineIndex < len(lines) {
		line := lines[lineIndex]

		// 检查是否有消息定义
		matches := messageRegex.FindStringSubmatch(line)
		if len(matches) > 0 {
			indentation := len(matches[1])
			messageName := matches[2]

			// 根据缩进调整父消息堆栈
			for len(parentStack) > 0 && indentationLevel(parentStack[len(parentStack)-1]) >= indentation {
				parentStack = parentStack[:len(parentStack)-1]
			}

			// 构建完整的消息名称
			fullName := messageName
			if len(parentStack) > 0 {
				// 如果有父消息，添加父消息前缀
				parent := stripIndentation(parentStack[len(parentStack)-1])

				// 特殊处理Rsp和响应相关消息
				if messageName == "Rsp" {
					// 针对各种响应情况的统一处理
					if strings.HasPrefix(parent, "Request_") {
						// 常规情况：嵌套在Request_XXX下的Rsp
						fullName = parent + "_Rsp"
					} else if strings.HasPrefix(parent, "Notify_") {
						// 通知的响应情况
						fullName = parent + "_Rsp"
					} else {
						// 对于其他情况，尝试添加适当的前缀
						prefix := detectMessageTypePrefix(parent)
						if prefix != "" {
							fullName = prefix + "_" + parent + "_Rsp"
						} else {
							// 默认情况
							fullName = parent + "_" + messageName
						}
					}
				} else {
					// 处理普通嵌套消息
					fullName = parent + "_" + messageName
				}
			}

			// 添加消息名称到结果列表
			messageNames = append(messageNames, MessageName{
				Name:     messageName,
				FullName: fullName,
			})

			// 将当前消息及其缩进添加到堆栈
			parentStack = append(parentStack, fmt.Sprintf("%s%s", matches[1], messageName))
		}

		// 检查是否有闭合括号，用于调整父消息堆栈
		closeMatches := closeBraceRegex.FindStringSubmatch(line)
		if len(closeMatches) > 0 {
			indentation := len(closeMatches[1])

			// 移除缩进小于或等于当前闭合括号的所有父消息
			for len(parentStack) > 0 && indentationLevel(parentStack[len(parentStack)-1]) >= indentation {
				parentStack = parentStack[:len(parentStack)-1]
			}
		}

		lineIndex++
	}

	return messageNames
}

// 辅助函数：获取消息定义的缩进级别
func indentationLevel(s string) int {
	return len(s) - len(strings.TrimLeft(s, " \t"))
}

// 辅助函数：去除消息名称中的缩进
func stripIndentation(s string) string {
	return strings.TrimLeft(s, " \t")
}

// 检测消息类型前缀 (Request_/Notify_)
func detectMessageTypePrefix(msgName string) string {
	// 检查消息名称中是否包含常见关键字，用于推断其类型
	switch {
	case strings.Contains(strings.ToLower(msgName), "search") ||
		strings.Contains(strings.ToLower(msgName), "query") ||
		strings.Contains(strings.ToLower(msgName), "get") ||
		strings.Contains(strings.ToLower(msgName), "find"):
		return "Request"
	case strings.Contains(strings.ToLower(msgName), "update") ||
		strings.Contains(strings.ToLower(msgName), "notify") ||
		strings.Contains(strings.ToLower(msgName), "push") ||
		strings.Contains(strings.ToLower(msgName), "event"):
		return "Notify"
	default:
		// 无法确定，默认不添加前缀
		return ""
	}
}

// generateProtocolIDFile 生成协议ID文件
func generateProtocolIDFile(mapping ProtocolIDMapping) {
	// 基于包名的输出文件名
	packageNameLower := strings.ToLower(mapping.PackageName)
	outputFile := filepath.Join(*outputDir, fmt.Sprintf("%s_protocol_ids.go", packageNameLower))

	// 按名称排序以获得一致的输出
	sort.Slice(mapping.MessageIDs, func(i, j int) bool {
		return mapping.MessageIDs[i].Name < mapping.MessageIDs[j].Name
	})

	// 协议ID常量文件的模板
	const tmpl = `// Code generated by protocol ID generator. DO NOT EDIT.
package pb

// Protocol IDs for {{.PackageName}} package messages
const (
{{- range .MessageIDs}}
	PID_{{$.PackageName}}_{{.Name}} uint32 = 0x{{printf "%08x" .ID}} // {{$.PackageName}}.{{.Name}}
{{- end}}
)

// MessageNameToID maps message names to their protocol IDs
var {{.PackageName}}MessageNameToID = map[string]uint32{
{{- range .MessageIDs}}
	"{{.Name}}": PID_{{$.PackageName}}_{{.Name}},
{{- end}}
}

// IDToMessageName maps protocol IDs to their message names
var {{.PackageName}}IDToMessageName = map[uint32]string{
{{- range .MessageIDs}}
	PID_{{$.PackageName}}_{{.Name}}: "{{.Name}}",
{{- end}}
}

// GetProtocolID returns the protocol ID for the given message name
func Get{{.PackageName}}ProtocolID(messageName string) (uint32, bool) {
	id, ok := {{.PackageName}}MessageNameToID[messageName]
	return id, ok
}

// GetMessageName returns the message name for the given protocol ID
func Get{{.PackageName}}MessageName(pid uint32) (string, bool) {
	name, ok := {{.PackageName}}IDToMessageName[pid]
	return name, ok
}
`

	// 解析模板
	t, err := template.New("protocolIDs").Parse(tmpl)
	if err != nil {
		fmt.Printf("Error parsing template: %v\n", err)
		return
	}

	// 创建输出文件
	file, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Error creating file %s: %v\n", outputFile, err)
		return
	}
	defer file.Close()

	// 执行模板
	err = t.Execute(file, mapping)
	if err != nil {
		fmt.Printf("Error executing template: %v\n", err)
		return
	}

	if *debugMode {
		fmt.Printf("Generated protocol ID file: %s\n", outputFile)
	}
}

// 解析Request消息
func parseRequestMessages(content string) []Message {
	var messages []Message

	// 提取完整的Request消息块（包括所有嵌套内容）
	requestRegex := regexp.MustCompile(`(?s)message\s+Request\s+\{(.*?)\n\}`)
	requestMatches := requestRegex.FindStringSubmatch(content)
	if len(requestMatches) < 2 {
		if *debugMode {
			fmt.Println("DEBUG: Request block not found")
		}
		return messages
	}

	requestBody := requestMatches[1]
	if *debugMode {
		fmt.Println("DEBUG: Found Request block")
	}

	// 获取所有消息定义，包括嵌套
	allMessageNames := extractMessageNames(requestBody)
	if *debugMode {
		fmt.Printf("DEBUG: Found %d nested messages in Request block\n", len(allMessageNames))
		for i, msg := range allMessageNames {
			fmt.Printf("DEBUG:   Message %d: Name=%s, FullName=%s\n", i+1, msg.Name, msg.FullName)
		}
	}

	// 过滤出Request_前缀的消息
	for _, msgInfo := range allMessageNames {
		// 忽略Rsp消息和通用的Request类型本身
		if msgInfo.Name == "Rsp" || msgInfo.Name == "Request" {
			if *debugMode {
				fmt.Printf("DEBUG: Skipping base message type: %s\n", msgInfo.Name)
			}
			continue
		}

		fullName := msgInfo.FullName
		if !strings.HasPrefix(fullName, "Request_") {
			// 确保消息名称有正确的前缀
			fullName = "Request_" + fullName
		}

		if *debugMode {
			fmt.Printf("DEBUG: Processing request message '%s'\n", fullName)
		}

		// 直接从原始内容中获取该消息的定义，以确保注释匹配
		msgRegex := regexp.MustCompile(fmt.Sprintf(`(?s)(message\s+%s\s+\{[^{}]*(?:\{[^{}]*\}[^{}]*)*\})`, regexp.QuoteMeta(msgInfo.Name)))
		msgMatches := msgRegex.FindStringSubmatch(requestBody)

		if len(msgMatches) > 1 {
			msgContent := msgMatches[1]

			// 解析消息体获取字段
			fieldRegex := regexp.MustCompile(`(?m)^\s*([^{}\/]+)\s+([^{}\/]+)\s*=\s*(\d+);(?:\s*//(.*))?`)
			fieldMatches := fieldRegex.FindAllStringSubmatch(msgContent, -1)

			message := Message{
				Name:     msgInfo.Name,
				Comment:  extractMessageComment(requestBody, msgInfo.Name),
				FullName: fullName,
			}

			// 解析字段
			for _, fieldMatch := range fieldMatches {
				if len(fieldMatch) >= 4 {
					field := Field{
						Type:    strings.TrimSpace(fieldMatch[1]),
						Name:    strings.TrimSpace(fieldMatch[2]),
						Index:   len(message.Fields) + 1,
						Comment: "",
					}

					if len(fieldMatch) > 4 && fieldMatch[4] != "" {
						field.Comment = strings.TrimSpace(fieldMatch[4])
					}

					message.Fields = append(message.Fields, field)
				}
			}

			// 检查是否有对应的Rsp消息
			rspFullName := fullName + "_Rsp"
			rspExists := false

			// 查找是否存在Rsp消息
			for _, rspInfo := range allMessageNames {
				if rspInfo.FullName == rspFullName ||
					(rspInfo.Name == "Rsp" && strings.HasPrefix(rspInfo.FullName, fullName)) {
					rspExists = true
					break
				}
			}

			if rspExists {
				message.Response = rspFullName
				if *debugMode {
					fmt.Printf("DEBUG: Found Rsp message for %s: %s\n", message.Name, message.Response)
				}
			} else {
				message.Response = "OK" // 默认使用通用成功OK
			}

			messages = append(messages, message)
		}
	}

	return messages
}

// 解析Notify消息
func parseNotifyMessages(content string) []Message {
	var messages []Message

	// 提取完整的Notify消息块（包括所有嵌套内容）
	notifyRegex := regexp.MustCompile(`(?s)message\s+Notify\s+\{(.*?)\n\}`)
	notifyMatches := notifyRegex.FindStringSubmatch(content)
	if len(notifyMatches) < 2 {
		if *debugMode {
			fmt.Println("DEBUG: Notify block not found")
		}
		return messages
	}

	notifyBody := notifyMatches[1]
	if *debugMode {
		fmt.Println("DEBUG: Found Notify block")
	}

	// 获取所有消息定义，包括嵌套
	allMessageNames := extractMessageNames(notifyBody)
	if *debugMode {
		fmt.Printf("DEBUG: Found %d nested messages in Notify block\n", len(allMessageNames))
		for i, msg := range allMessageNames {
			fmt.Printf("DEBUG:   Message %d: Name=%s, FullName=%s\n", i+1, msg.Name, msg.FullName)
		}
	}

	// 过滤出Notify_前缀的消息
	for _, msgInfo := range allMessageNames {
		// 忽略通用的Notify类型本身
		if msgInfo.Name == "Notify" {
			if *debugMode {
				fmt.Printf("DEBUG: Skipping base message type: %s\n", msgInfo.Name)
			}
			continue
		}

		fullName := msgInfo.FullName
		if !strings.HasPrefix(fullName, "Notify_") {
			// 确保消息名称有正确的前缀
			fullName = "Notify_" + fullName
		}

		if *debugMode {
			fmt.Printf("DEBUG: Processing notify message '%s'\n", fullName)
		}

		// 直接从原始内容中获取该消息的定义，以确保注释匹配
		msgRegex := regexp.MustCompile(fmt.Sprintf(`(?s)(message\s+%s\s+\{[^{}]*(?:\{[^{}]*\}[^{}]*)*\})`, regexp.QuoteMeta(msgInfo.Name)))
		msgMatches := msgRegex.FindStringSubmatch(notifyBody)

		if len(msgMatches) > 1 {
			msgContent := msgMatches[1]

			// 解析消息体获取字段
			fieldRegex := regexp.MustCompile(`(?m)^\s*([^{}\/]+)\s+([^{}\/]+)\s*=\s*(\d+);(?:\s*//(.*))?`)
			fieldMatches := fieldRegex.FindAllStringSubmatch(msgContent, -1)

			message := Message{
				Name:     msgInfo.Name,
				Comment:  extractMessageComment(notifyBody, msgInfo.Name),
				FullName: fullName,
			}

			// 解析字段
			for _, fieldMatch := range fieldMatches {
				if len(fieldMatch) >= 4 {
					field := Field{
						Type:    strings.TrimSpace(fieldMatch[1]),
						Name:    strings.TrimSpace(fieldMatch[2]),
						Index:   len(message.Fields) + 1,
						Comment: "",
					}

					if len(fieldMatch) > 4 && fieldMatch[4] != "" {
						field.Comment = strings.TrimSpace(fieldMatch[4])
					}

					message.Fields = append(message.Fields, field)
				}
			}

			messages = append(messages, message)
		}
	}

	return messages
}

// 提取消息注释
func extractMessageComment(content, messageName string) string {
	// 查找消息定义前的注释 - 使用更精确的匹配方式
	// 先获取消息定义的行
	msgDefRegex := regexp.MustCompile(fmt.Sprintf(`(?m)^(\s*)message\s+%s\s*\{`, regexp.QuoteMeta(messageName)))
	msgMatches := msgDefRegex.FindStringIndex(content)
	if len(msgMatches) < 2 {
		return ""
	}

	// 从消息定义行向上查找注释块
	contentBeforeMsg := content[:msgMatches[0]]
	lines := strings.Split(contentBeforeMsg, "\n")

	// 从下往上遍历寻找注释
	var commentLines []string
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])

		// 如果是空行，则停止收集注释
		if line == "" {
			break
		}

		// 检查是否是注释行
		if strings.HasPrefix(line, "//") {
			// 提取注释内容（去掉 // 前缀）
			commentContent := strings.TrimSpace(strings.TrimPrefix(line, "//"))
			commentLines = append([]string{commentContent}, commentLines...) // 保持注释顺序
		} else if strings.HasPrefix(line, "/*") && strings.HasSuffix(line, "*/") {
			// 处理单行 /* ... */ 注释
			commentContent := strings.TrimSpace(line[2 : len(line)-2])
			commentLines = append([]string{commentContent}, commentLines...)
		} else {
			// 不是注释行，停止收集
			break
		}
	}

	// 合并注释行
	if len(commentLines) > 0 {
		return strings.Join(commentLines, " ")
	}

	return ""
}

// 解析消息字段
func parseMessageFields(message *Message, msgBody string) {
	// 查找字段和注释
	fieldRegex := regexp.MustCompile(`(\w+)\s+(\w+)\s*=\s*(\d+);(?:\s*//(.*))?`)
	fieldMatches := fieldRegex.FindAllStringSubmatch(msgBody, -1)

	for _, fieldMatch := range fieldMatches {
		field := Field{
			Type:    fieldMatch[1],
			Name:    fieldMatch[2],
			Index:   len(message.Fields) + 1,
			Comment: "",
		}

		if len(fieldMatch) > 4 && fieldMatch[4] != "" {
			field.Comment = strings.TrimSpace(fieldMatch[4])
		}

		message.Fields = append(message.Fields, field)
	}
}

// 生成Request消息的胶水代码
func generateRequestGlueCode(messages []Message, packageName, pbDir string) {
	filename := filepath.Join(pbDir, strings.ToLower(packageName)+"_request_glue.go")

	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error creating file %s: %v\n", filename, err)
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	// 写入文件头
	writer.WriteString("// 自动生成的代码 - 请勿手动修改\n")
	writer.WriteString("package pb\n\n")

	// 写入导入
	writer.WriteString("import (\n")
	writer.WriteString("\t\"fmt\"\n")
	writer.WriteString("\t\"google.golang.org/protobuf/proto\"\n")
	writer.WriteString("\t\"github.com/orbit-w/orbit/lib/utils/proto_utils\"\n")
	writer.WriteString(")\n\n")

	// 添加基础消息常量 - 无需协议ID
	writer.WriteString("// Base message types that don't need protocol IDs\n")
	writer.WriteString("const (\n")
	writer.WriteString(fmt.Sprintf("\tPID_%s_Request uint32 = 0\n", packageName))
	writer.WriteString(fmt.Sprintf("\tPID_%s_Response uint32 = 0\n", packageName))
	writer.WriteString(")\n\n")

	// 写入接口定义
	writer.WriteString(fmt.Sprintf("// %sRequestHandler 定义处理%s包Request消息的接口\n", packageName, packageName))
	writer.WriteString(fmt.Sprintf("type %sRequestHandler interface {\n", packageName))
	for _, msg := range messages {
		writer.WriteString(fmt.Sprintf("\t// Handle%s 处理%s请求\n", msg.Name, msg.Name))
		if msg.Comment != "" {
			writer.WriteString(fmt.Sprintf("\t// %s\n", msg.Comment))
		}
		writer.WriteString(fmt.Sprintf("\tHandle%s(req *Request_%s) any\n", msg.Name, msg.Name))
	}
	writer.WriteString("}\n\n")

	// 写入Dispatch函数
	writer.WriteString(fmt.Sprintf("// Dispatch%sRequest 分发%s包的请求消息到对应处理函数\n", packageName, packageName))
	writer.WriteString(fmt.Sprintf("func Dispatch%sRequest(handler %sRequestHandler, msgName string, data []byte) (any, uint32, error) {\n", packageName, packageName))
	writer.WriteString("\tvar err error\n")
	writer.WriteString("\tvar response any\n")
	writer.WriteString("\tvar responsePid uint32\n\n")

	writer.WriteString("\tswitch msgName {\n")
	for _, msg := range messages {
		// 对于消息名，我们使用完整路径
		fullNameForCase := msg.FullName

		// 如果FullName为空，则使用传统的名称
		if fullNameForCase == "" {
			fullNameForCase = "Request_" + msg.Name
		}

		writer.WriteString(fmt.Sprintf("\tcase \"%s\":\n", fullNameForCase))
		writer.WriteString(fmt.Sprintf("\t\treq := &%s{}\n", fullNameForCase))
		writer.WriteString("\t\tif err = proto.Unmarshal(data, req); err != nil {\n")
		writer.WriteString(fmt.Sprintf("\t\t\treturn nil, 0, fmt.Errorf(\"unmarshal %s failed: %%v\", err)\n", msg.Name))
		writer.WriteString("\t\t}\n")
		writer.WriteString(fmt.Sprintf("\t\tresponse = handler.Handle%s(req)\n", msg.Name))
		writer.WriteString(fmt.Sprintf("\t\tresponsePid = get%sResponsePID(response)\n", packageName))
	}
	writer.WriteString("\tdefault:\n")
	writer.WriteString(fmt.Sprintf("\t\treturn nil, 0, fmt.Errorf(\"unknown request message: %%s\", msgName)\n"))
	writer.WriteString("\t}\n\n")

	writer.WriteString("\treturn response, responsePid, nil\n")
	writer.WriteString("}\n\n")

	// 写入DispatchByID函数
	writer.WriteString(fmt.Sprintf("// Dispatch%sRequestByID 根据协议ID分发%s包的请求消息到对应处理函数\n", packageName, packageName))
	writer.WriteString(fmt.Sprintf("func Dispatch%sRequestByID(handler %sRequestHandler, pid uint32, data []byte) (any, uint32, error) {\n", packageName, packageName))
	writer.WriteString("\tvar err error\n")
	writer.WriteString("\tvar response any\n")
	writer.WriteString("\tvar responsePid uint32\n\n")

	writer.WriteString("\tswitch pid {\n")
	for _, msg := range messages {
		writer.WriteString(fmt.Sprintf("\tcase PID_%s_%s:\n", packageName, msg.FullName))
		writer.WriteString(fmt.Sprintf("\t\treq := &%s{}\n", msg.FullName))
		writer.WriteString("\t\tif err = proto.Unmarshal(data, req); err != nil {\n")
		writer.WriteString(fmt.Sprintf("\t\t\treturn nil, 0, fmt.Errorf(\"unmarshal %s failed: %%v\", err)\n", msg.Name))
		writer.WriteString("\t\t}\n")
		writer.WriteString(fmt.Sprintf("\t\tresponse = handler.Handle%s(req)\n", msg.Name))
		writer.WriteString(fmt.Sprintf("\t\tresponsePid = get%sResponsePID(response)\n", packageName))
	}
	writer.WriteString("\tdefault:\n")
	writer.WriteString(fmt.Sprintf("\t\treturn nil, 0, fmt.Errorf(\"unknown request protocol ID: 0x%%x\", pid)\n"))
	writer.WriteString("\t}\n\n")

	writer.WriteString("\treturn response, responsePid, nil\n")
	writer.WriteString("}\n\n")

	// 最后写入getResponsePID辅助函数，添加基础消息类型处理
	writer.WriteString(fmt.Sprintf("// get%sResponsePID 通过反射获取响应消息的协议ID\n", packageName))
	writer.WriteString(fmt.Sprintf("func get%sResponsePID(response any) uint32 {\n", packageName))
	writer.WriteString("\t// 对基础消息类型特殊处理\n")
	writer.WriteString("\tif response == nil {\n")
	writer.WriteString("\t\treturn 0\n")
	writer.WriteString("\t}\n\n")

	writer.WriteString("\t// 获取消息名称\n")
	writer.WriteString("\ttypeName := proto_utils.ParseMessageName(response)\n")
	writer.WriteString("\tif typeName == \"\" {\n")
	writer.WriteString("\t\treturn 0\n")
	writer.WriteString("\t}\n\n")

	// 特殊处理基础消息类型
	writer.WriteString("\t// 基础消息类型特殊处理\n")
	writer.WriteString("\tif typeName == \"Response\" || typeName == \"OK\" {\n")
	writer.WriteString(fmt.Sprintf("\t\treturn PID_%s_Response\n", packageName))
	writer.WriteString("\t}\n\n")

	// 通过名称查找PID - 通用方式处理所有类型
	writer.WriteString("\t// 查找类型对应的协议ID\n")
	writer.WriteString(fmt.Sprintf("\tpid, ok := Get%sProtocolID(typeName)\n", packageName))
	writer.WriteString("\tif ok {\n")
	writer.WriteString("\t\treturn pid\n")
	writer.WriteString("\t}\n\n")

	writer.WriteString("\t// 未找到对应的协议ID\n")
	writer.WriteString("\treturn 0\n")
	writer.WriteString("}\n")

	writer.Flush()

	if *debugMode {
		fmt.Printf("Generated request glue code: %s\n", filename)
	}
}

// 生成Notify消息的胶水代码
func generateNotifyGlueCode(messages []Message, packageName, pbDir string) {
	filename := filepath.Join(pbDir, strings.ToLower(packageName)+"_notify_glue.go")

	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error creating file %s: %v\n", filename, err)
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	// 写入文件头
	writer.WriteString("// 自动生成的代码 - 请勿手动修改\n")
	writer.WriteString("package pb\n\n")

	// 写入导入
	writer.WriteString("import (\n")
	writer.WriteString("\t\"fmt\"\n")
	writer.WriteString("\t\"google.golang.org/protobuf/proto\"\n")
	writer.WriteString("\t\"github.com/orbit-w/orbit/lib/utils/proto_utils\"\n")
	writer.WriteString(")\n\n")

	// 添加基础消息常量 - 无需协议ID
	writer.WriteString("// Base notification message type that doesn't need protocol ID\n")
	writer.WriteString("const (\n")
	writer.WriteString(fmt.Sprintf("\tPID_%s_Notify uint32 = 0\n", packageName))
	writer.WriteString(")\n\n")

	// 为每个包创建唯一的getNotificationPID函数名
	notificationPIDFuncName := fmt.Sprintf("get%sNotificationPID", packageName)

	// 添加getNotificationPID辅助函数，添加包名前缀确保唯一性
	writer.WriteString(fmt.Sprintf("// %s 通过反射获取通知消息的协议ID\n", notificationPIDFuncName))
	writer.WriteString(fmt.Sprintf("func %s(notification any) uint32 {\n", notificationPIDFuncName))
	writer.WriteString("\t// 对基础消息类型特殊处理\n")
	writer.WriteString("\tif notification == nil {\n")
	writer.WriteString("\t\treturn 0\n")
	writer.WriteString("\t}\n\n")

	writer.WriteString("\t// 获取消息名称\n")
	writer.WriteString("\ttypeName := proto_utils.ParseMessageName(notification)\n")
	writer.WriteString("\tif typeName == \"\" {\n")
	writer.WriteString("\t\treturn 0\n")
	writer.WriteString("\t}\n\n")

	// 特殊处理基础消息类型
	writer.WriteString("\t// 基础消息类型特殊处理\n")
	writer.WriteString("\tif typeName == \"Notify\" {\n")
	writer.WriteString(fmt.Sprintf("\t\treturn PID_%s_Notify\n", packageName))
	writer.WriteString("\t}\n\n")

	// 通过名称查找PID - 保持简洁
	writer.WriteString("\t// 查找类型对应的协议ID\n")
	writer.WriteString(fmt.Sprintf("\tpid, ok := Get%sProtocolID(typeName)\n", packageName))
	writer.WriteString("\tif ok {\n")
	writer.WriteString("\t\treturn pid\n")
	writer.WriteString("\t}\n\n")

	writer.WriteString("\t// 未找到对应的协议ID\n")
	writer.WriteString("\treturn 0\n")
	writer.WriteString("}\n\n")

	// 写入通知序列化函数 - 修改为使用getNotificationPID函数
	for _, msg := range messages {
		writer.WriteString(fmt.Sprintf("// Marshal%s 序列化%s通知消息\n", msg.Name, msg.Name))
		if msg.Comment != "" {
			writer.WriteString(fmt.Sprintf("// %s\n", msg.Comment))
		}
		writer.WriteString(fmt.Sprintf("func Marshal%s(notify *%s) ([]byte, uint32, error) {\n", msg.Name, msg.FullName))
		writer.WriteString("\tdata, err := proto.Marshal(notify)\n")
		// 对于已知消息类型，我们可以直接返回其固定的协议ID
		writer.WriteString(fmt.Sprintf("\treturn data, PID_%s_%s, err\n", packageName, msg.FullName))
		writer.WriteString("}\n\n")
	}

	// 写入基于消息名称的通知解析函数
	writer.WriteString(fmt.Sprintf("// Parse%sNotify 根据消息名称解析%s包的通知消息\n", packageName, packageName))
	writer.WriteString(fmt.Sprintf("func Parse%sNotify(msgName string, data []byte) (any, uint32, error) {\n", packageName))
	writer.WriteString("\tvar err error\n")
	writer.WriteString("\tvar notification any\n")
	writer.WriteString("\tvar notificationPid uint32\n\n")
	writer.WriteString("\tswitch msgName {\n")

	for _, msg := range messages {
		// 对于消息名，我们使用完整路径
		fullNameForCase := msg.FullName

		// 如果FullName为空，则使用传统的名称
		if fullNameForCase == "" {
			fullNameForCase = "Notify_" + msg.Name
		}

		writer.WriteString(fmt.Sprintf("\tcase \"%s\":\n", fullNameForCase))
		writer.WriteString(fmt.Sprintf("\t\tnotify := &%s{}\n", fullNameForCase))
		writer.WriteString("\t\tif err = proto.Unmarshal(data, notify); err != nil {\n")
		writer.WriteString(fmt.Sprintf("\t\t\treturn nil, 0, fmt.Errorf(\"unmarshal %s notification failed: %%v\", err)\n", msg.Name))
		writer.WriteString("\t\t}\n")
		writer.WriteString("\t\tnotification = notify\n")
		writer.WriteString(fmt.Sprintf("\t\tnotificationPid = %s(notification)\n", notificationPIDFuncName))
	}

	writer.WriteString("\tdefault:\n")
	writer.WriteString("\t\treturn nil, 0, fmt.Errorf(\"unknown notification message: %s\", msgName)\n")
	writer.WriteString("\t}\n\n")
	writer.WriteString("\treturn notification, notificationPid, nil\n")
	writer.WriteString("}\n\n")

	// 写入基于协议ID的通知解析函数
	writer.WriteString(fmt.Sprintf("// Parse%sNotifyByID 根据协议ID解析%s包的通知消息\n", packageName, packageName))
	writer.WriteString(fmt.Sprintf("func Parse%sNotifyByID(pid uint32, data []byte) (any, uint32, error) {\n", packageName))
	writer.WriteString("\tvar err error\n")
	writer.WriteString("\tvar notification any\n")
	writer.WriteString("\tvar notificationPid uint32\n\n")
	writer.WriteString("\tswitch pid {\n")

	// 特殊处理基础Notify类型
	writer.WriteString(fmt.Sprintf("\tcase PID_%s_Notify:\n", packageName))
	writer.WriteString("\t\t// 基础通知类型特殊处理\n")
	writer.WriteString("\t\treturn nil, 0, fmt.Errorf(\"cannot unmarshal base Notify type directly\")\n\n")

	for _, msg := range messages {
		writer.WriteString(fmt.Sprintf("\tcase PID_%s_%s:\n", packageName, msg.FullName))
		writer.WriteString(fmt.Sprintf("\t\tnotify := &%s{}\n", msg.FullName))
		writer.WriteString("\t\tif err = proto.Unmarshal(data, notify); err != nil {\n")
		writer.WriteString(fmt.Sprintf("\t\t\treturn nil, 0, fmt.Errorf(\"unmarshal notification with ID 0x%%08x failed: %%v\", pid, err)\n"))
		writer.WriteString("\t\t}\n")
		writer.WriteString("\t\tnotification = notify\n")
		// 使用添加了包名前缀的函数
		writer.WriteString(fmt.Sprintf("\t\tnotificationPid = %s(notification)\n", notificationPIDFuncName))
	}

	writer.WriteString("\tdefault:\n")
	writer.WriteString("\t\treturn nil, 0, fmt.Errorf(\"unknown notification protocol ID: 0x%08x\", pid)\n")
	writer.WriteString("\t}\n\n")
	writer.WriteString("\treturn notification, notificationPid, nil\n")
	writer.WriteString("}\n")

	writer.Flush()

	if *debugMode {
		fmt.Printf("Generated notify glue code: %s\n", filename)
	}
}
