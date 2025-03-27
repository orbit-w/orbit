// Package main provides a tool for generating protocol IDs and glue code from proto message definitions
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

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

	// 收集所有包和消息
	var allMappings []ProtocolIDMapping

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
			mapping := generateProtocolIDs(string(content), packageName)
			if len(mapping.MessageIDs) > 0 {
				allMappings = append(allMappings, mapping)
			}
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
							fmt.Printf("  Request %d: Name=%s, FullName=%s\n", i+1, msg.Name, msg.FullName)
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
							fmt.Printf("  Notify %d: Name=%s, FullName=%s\n", i+1, msg.Name, msg.FullName)
						}
					}
				}
				generateNotifyGlueCode(notifyMessages, packageName, *outputDir)
			} else if !*quietMode {
				fmt.Printf("No notify messages found\n")
			}
		}
	}

	// 生成公共协议ID映射文件
	if *genProtoIDs && len(allMappings) > 0 {
		generateCommonProtocolMappings(allMappings, *outputDir)
	}

	if !*quietMode {
		fmt.Println("All generation completed!")
	}
}

// generateProtocolIDs 生成协议ID
func generateProtocolIDs(content, packageName string) ProtocolIDMapping {
	// 提取所有消息名称
	messageNames := extractMessageNames(content)
	if len(messageNames) == 0 {
		if *debugMode {
			fmt.Printf("No message definitions found for %s\n", packageName)
		}
		return ProtocolIDMapping{}
	}

	// 生成协议ID映射
	mapping := ProtocolIDMapping{
		PackageName: packageName,
	}

	// 按名称排序以得到一致的输出
	sort.Slice(messageNames, func(i, j int) bool {
		return messageNames[i].FullName < messageNames[j].FullName
	})

	// 为每个消息生成ID，只处理特定类型的消息
	for _, msg := range messageNames {
		// 跳过基本类型：Request, Notify
		if msg.Name == "Request" || msg.Name == "Notify" {
			if *debugMode {
				fmt.Printf("Skipping base message type: %s\n", msg.Name)
			}
			continue
		}

		// 处理响应消息
		if msg.Name == "Rsp" {
			// 从父消息名称中提取请求名称
			parentName := strings.TrimSuffix(msg.FullName, "_Rsp")
			if strings.HasPrefix(parentName, "Request_") {
				parentName = strings.TrimPrefix(parentName, "Request_")
				// 使用Core-Request_SearchBook_Rsp格式的消息名称计算PID
				fullName := fmt.Sprintf("%s-Request_%s_Rsp", packageName, parentName)
				pid := protoid.HashProtoMessage(fullName)
				mapping.MessageIDs = append(mapping.MessageIDs, MessageID{
					Name: fmt.Sprintf("Request_%s_Rsp", parentName),
					ID:   pid,
				})
				if *debugMode {
					fmt.Printf("Generated PID for response message: %s, ID: 0x%016x\n", fullName, pid)
				}
			}
			continue
		}

		// 只处理特定类型的消息
		isRequest := strings.HasPrefix(msg.FullName, "Request_")
		isNotify := strings.HasPrefix(msg.FullName, "Notify_")
		isOK := msg.Name == "OK"
		isFail := msg.Name == "Fail"

		if !isRequest && !isNotify && !isOK && !isFail {
			if *debugMode {
				fmt.Printf("Skipping non-special message: %s\n", msg.FullName)
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
			fmt.Printf("Generated PID for special message: %s, ID: 0x%016x\n", fullName, pid)
		}
	}

	return mapping
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
	// 构建输出文件名
	outputFile := filepath.Join(pbDir, strings.ToLower(packageName)+"_request_glue.go")
	file, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Failed to create output file: %v\n", err)
		return
	}
	defer file.Close()

	// 文件头
	fmt.Fprintf(file, "// Code generated by genproto. DO NOT EDIT.\n")
	fmt.Fprintf(file, "package pb\n\n")

	// 导入必要的包
	fmt.Fprintf(file, "import (\n")
	fmt.Fprintf(file, "\t\"fmt\"\n")
	fmt.Fprintf(file, "\t\"google.golang.org/protobuf/proto\"\n")
	fmt.Fprintf(file, "\t\"github.com/orbit-w/orbit/app/proto/pb/%s\"\n", strings.ToLower(packageName))
	fmt.Fprintf(file, ")\n\n")

	// 写入请求处理器接口
	fmt.Fprintf(file, "// %sRequestHandler 处理%s包的请求消息\n", packageName, packageName)
	fmt.Fprintf(file, "type %sRequestHandler interface {\n", packageName)
	for _, msg := range messages {
		if msg.Name == "Request" {
			continue
		}
		fmt.Fprintf(file, "\t// Handle%s 处理%s请求\n", msg.Name, msg.Name)
		if msg.Comment != "" {
			fmt.Fprintf(file, "\t// %s\n", msg.Comment)
		}
		fmt.Fprintf(file, "\tHandle%s(req *%s.%s) proto.Message\n", msg.Name, strings.ToLower(packageName), msg.FullName)
	}
	fmt.Fprintf(file, "}\n\n")

	// 生成分发函数
	fmt.Fprintf(file, "// Dispatch%sRequestByID 根据协议ID分发请求到对应处理函数\n", packageName)
	fmt.Fprintf(file, "func Dispatch%sRequestByID(handler %sRequestHandler, pid uint32, data []byte) (proto.Message, uint32, error) {\n", packageName, packageName)
	fmt.Fprintf(file, "\tvar response proto.Message\n")
	fmt.Fprintf(file, "\tswitch pid {\n")

	for _, msg := range messages {
		// 跳过不符合条件的消息
		if msg.Name == "Request" {
			continue
		}

		// 生成pid以供参考
		fullName := fmt.Sprintf("%s-%s", packageName, msg.FullName)
		_ = protoid.HashProtoMessage(fullName)

		fmt.Fprintf(file, "\tcase PID_%s_%s: // %s\n", packageName, msg.FullName, msg.FullName)
		fmt.Fprintf(file, "\t\treq := &%s.%s{}\n", strings.ToLower(packageName), msg.FullName)
		fmt.Fprintf(file, "\t\tif err := proto.Unmarshal(data, req); err != nil {\n")
		fmt.Fprintf(file, "\t\t\treturn nil, 0, fmt.Errorf(\"unmarshal %s failed: %%w\", err)\n", msg.FullName)
		fmt.Fprintf(file, "\t\t}\n\n")
		fmt.Fprintf(file, "\t\tresponse = handler.Handle%s(req)\n", msg.Name)
		fmt.Fprintf(file, "\t\n")
	}

	fmt.Fprintf(file, "\tdefault:\n")
	fmt.Fprintf(file, "\t\treturn nil, 0, fmt.Errorf(\"unknown request protocol ID: 0x%%08x\", pid)\n")
	fmt.Fprintf(file, "\t}\n\n")
	fmt.Fprintf(file, "\t// 使用公共映射文件获取响应ID\n")
	fmt.Fprintf(file, "\tresponsePid := GetResponsePID(response)\n")
	fmt.Fprintf(file, "\treturn response, responsePid, nil\n")
	fmt.Fprintf(file, "}\n")

	if !*quietMode {
		fmt.Printf("Generated request glue code in %s\n", outputFile)
	}
}

// 生成Notify消息的胶水代码
func generateNotifyGlueCode(messages []Message, packageName, pbDir string) {
	// 构建输出文件名
	outputFile := filepath.Join(pbDir, strings.ToLower(packageName)+"_notify_glue.go")
	file, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Failed to create output file: %v\n", err)
		return
	}
	defer file.Close()

	// 文件头
	fmt.Fprintf(file, "// Code generated by genproto. DO NOT EDIT.\n")
	fmt.Fprintf(file, "package pb\n\n")

	// 导入必要的包
	fmt.Fprintf(file, "import (\n")
	fmt.Fprintf(file, "\t\"fmt\"\n")
	fmt.Fprintf(file, "\t\"google.golang.org/protobuf/proto\"\n")
	fmt.Fprintf(file, "\t\"github.com/orbit-w/orbit/app/proto/pb/%s\"\n", strings.ToLower(packageName))
	fmt.Fprintf(file, ")\n\n")

	// 生成分发函数
	fmt.Fprintf(file, "// Parse%sNotifyByID 根据协议ID解析通知消息\n", packageName)
	fmt.Fprintf(file, "func Parse%sNotifyByID(pid uint32, data []byte) (proto.Message, uint32, error) {\n", packageName)
	fmt.Fprintf(file, "\tswitch pid {\n")

	for _, msg := range messages {
		// 跳过不符合条件的消息
		if msg.Name == "Notify" {
			continue
		}

		// 生成pid以供参考
		fullName := fmt.Sprintf("%s-%s", packageName, msg.FullName)
		_ = protoid.HashProtoMessage(fullName)

		fmt.Fprintf(file, "\tcase PID_%s_%s: // %s\n", packageName, msg.FullName, msg.FullName)
		fmt.Fprintf(file, "\t\tnotify := &%s.%s{}\n", strings.ToLower(packageName), msg.FullName)
		fmt.Fprintf(file, "\t\tif err := proto.Unmarshal(data, notify); err != nil {\n")
		fmt.Fprintf(file, "\t\t\treturn nil, 0, fmt.Errorf(\"unmarshal %s failed: %%w\", err)\n", msg.FullName)
		fmt.Fprintf(file, "\t\t}\n")
		fmt.Fprintf(file, "\t\treturn notify, pid, nil\n")
	}

	fmt.Fprintf(file, "\tdefault:\n")
	fmt.Fprintf(file, "\t\treturn nil, 0, fmt.Errorf(\"unknown notify protocol ID: 0x%%08x\", pid)\n")
	fmt.Fprintf(file, "\t}\n")
	fmt.Fprintf(file, "}\n\n")

	// 生成Marshal函数
	for _, msg := range messages {
		if msg.Name == "Notify" {
			continue
		}

		fmt.Fprintf(file, "// Marshal%s 序列化%s通知消息\n", msg.Name, msg.Name)
		if msg.Comment != "" {
			fmt.Fprintf(file, "// %s\n", msg.Comment)
		}
		fmt.Fprintf(file, "func Marshal%s(notify *%s.%s) ([]byte, uint32, error) {\n", msg.Name, strings.ToLower(packageName), msg.FullName)
		fmt.Fprintf(file, "\tdata, err := proto.Marshal(notify)\n")
		fmt.Fprintf(file, "\treturn data, PID_%s_%s, err\n", packageName, msg.FullName)
		fmt.Fprintf(file, "}\n\n")
	}

	if !*quietMode {
		fmt.Printf("Generated notify glue code in %s\n", outputFile)
	}
}

// generateCommonProtocolMappings 生成公共的协议ID映射文件
func generateCommonProtocolMappings(allMappings []ProtocolIDMapping, outputDir string) {
	outputFile := filepath.Join(outputDir, "protocol_ids.go")

	file, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Error creating protocol ID file: %v\n", err)
		return
	}
	defer file.Close()

	// 文件头
	fmt.Fprintf(file, "// Code generated by protocol ID generator. DO NOT EDIT.\n")
	fmt.Fprintf(file, "package pb\n\n")

	// 导入必要的包
	fmt.Fprintf(file, "import (\n")
	fmt.Fprintf(file, "\t\"fmt\"\n")
	fmt.Fprintf(file, "\t\"google.golang.org/protobuf/proto\"\n")
	fmt.Fprintf(file, "\t\"github.com/orbit-w/orbit/lib/utils/proto_utils\"\n")
	fmt.Fprintf(file, ")\n\n")

	// 生成所有协议ID常量
	fmt.Fprintf(file, "// 所有协议ID常量\n")
	fmt.Fprintf(file, "const (\n")
	for _, mapping := range allMappings {
		fmt.Fprintf(file, "\t// %s 包协议ID\n", mapping.PackageName)
		for _, msgID := range mapping.MessageIDs {
			fmt.Fprintf(file, "\tPID_%s_%s uint32 = 0x%08x // %s.%s\n",
				mapping.PackageName, msgID.Name, msgID.ID, mapping.PackageName, msgID.Name)
		}
		fmt.Fprintf(file, "\n")
	}
	fmt.Fprintf(file, ")\n\n")

	// 生成全局的 MessageNameToID 映射
	fmt.Fprintf(file, "// AllMessageNameToID 全局消息名称到ID的映射\n")
	fmt.Fprintf(file, "var AllMessageNameToID = map[string]uint32{\n")
	for _, mapping := range allMappings {
		for _, msgID := range mapping.MessageIDs {
			fmt.Fprintf(file, "\t\"%s-%s\": PID_%s_%s,\n",
				mapping.PackageName, msgID.Name, mapping.PackageName, msgID.Name)
		}
	}
	fmt.Fprintf(file, "}\n\n")

	// 生成全局的 IDToMessageName 映射
	fmt.Fprintf(file, "// AllIDToMessageName 全局ID到消息名称的映射\n")
	fmt.Fprintf(file, "var AllIDToMessageName = map[uint32]string{\n")
	for _, mapping := range allMappings {
		for _, msgID := range mapping.MessageIDs {
			fmt.Fprintf(file, "\tPID_%s_%s: \"%s-%s\",\n",
				mapping.PackageName, msgID.Name, mapping.PackageName, msgID.Name)
		}
	}
	fmt.Fprintf(file, "}\n\n")

	// 生成 MessagePackageMap 映射
	fmt.Fprintf(file, "// MessagePackageMap 消息名称到包名的映射\n")
	fmt.Fprintf(file, "var MessagePackageMap = map[string]string{\n")
	for _, mapping := range allMappings {
		for _, msgID := range mapping.MessageIDs {
			fmt.Fprintf(file, "\t\"%s\": \"%s\",\n",
				msgID.Name, mapping.PackageName)
		}
	}
	fmt.Fprintf(file, "}\n\n")

	// 生成全局 GetProtocolID 函数
	fmt.Fprintf(file, "// GetProtocolID 获取指定消息名称的协议ID\n")
	fmt.Fprintf(file, "func GetProtocolID(messageName string) (uint32, bool) {\n")
	fmt.Fprintf(file, "\tid, ok := AllMessageNameToID[messageName]\n")
	fmt.Fprintf(file, "\treturn id, ok\n")
	fmt.Fprintf(file, "}\n\n")

	// 生成全局 GetMessageName 函数
	fmt.Fprintf(file, "// GetMessageName 获取指定协议ID的消息名称\n")
	fmt.Fprintf(file, "func GetMessageName(pid uint32) (string, bool) {\n")
	fmt.Fprintf(file, "\tname, ok := AllIDToMessageName[pid]\n")
	fmt.Fprintf(file, "\treturn name, ok\n")
	fmt.Fprintf(file, "}\n\n")

	// 生成 GetResponsePID 函数
	fmt.Fprintf(file, "// GetResponsePID 获取响应消息的协议ID\n")
	fmt.Fprintf(file, "func GetResponsePID(response proto.Message) uint32 {\n")
	fmt.Fprintf(file, "\tif response == nil {\n")
	fmt.Fprintf(file, "\t\treturn 0\n")
	fmt.Fprintf(file, "\t}\n\n")
	fmt.Fprintf(file, "\tmessageName := proto_utils.ParseMessageName(response)\n")
	fmt.Fprintf(file, "\tif messageName == \"\" {\n")
	fmt.Fprintf(file, "\t\treturn 0\n")
	fmt.Fprintf(file, "\t}\n\n")
	fmt.Fprintf(file, "\t// 从映射表中查找包名\n")
	fmt.Fprintf(file, "\tpackageName, ok := MessagePackageMap[messageName]\n")
	fmt.Fprintf(file, "\tif !ok {\n")
	fmt.Fprintf(file, "\t\t// 找不到包名直接panic\n")
	fmt.Fprintf(file, "\t\tpanic(fmt.Sprintf(\"消息 %%s 未在映射表中找到对应的包名\", messageName))\n")
	fmt.Fprintf(file, "\t}\n\n")
	fmt.Fprintf(file, "\tfullName := packageName + \"-\" + messageName\n")
	fmt.Fprintf(file, "\tpid, _ := GetProtocolID(fullName)\n")
	fmt.Fprintf(file, "\treturn pid\n")
	fmt.Fprintf(file, "}\n")

	if !*quietMode {
		fmt.Printf("Generated common protocol ID mappings in %s\n", outputFile)
	}
}
