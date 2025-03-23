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
							fmt.Printf("  Request %d: %s, Response: %s\n", i+1, msg.Name, msg.Response)
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
	sort.Strings(messageNames)

	// 为每个消息生成ID
	for _, msgName := range messageNames {
		fullName := fmt.Sprintf("%s.%s", packageName, msgName)
		pid := protoid.HashProtoMessage(fullName)
		mapping.MessageIDs = append(mapping.MessageIDs, MessageID{
			Name: msgName,
			ID:   pid,
		})

		if *debugMode {
			fmt.Printf("Message: %s, ID: 0x%08x\n", fullName, pid)
		}
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

// extractMessageNames 提取所有顶级消息名称
func extractMessageNames(content string) []string {
	var messageNames []string
	re := regexp.MustCompile(`message\s+(\w+)\s*\{`)
	matches := re.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			messageNames = append(messageNames, match[1])
		}
	}
	return messageNames
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

	// 使用更灵活的正则表达式来查找嵌套的消息定义
	// 先找到注释，然后找到消息定义
	commentRegex := regexp.MustCompile(`(?s)//([^\n]*(?:\n\s*//[^\n]*)*)`)
	messageRegex := regexp.MustCompile(`(?s)message\s+(\w+)\s+\{([^{}]*(?:\{[^{}]*\}[^{}]*)*)\}`)

	// 在requestBody中查找所有消息定义
	pos := 0
	for pos < len(requestBody) {
		// 查找注释
		comment := ""
		commentMatches := commentRegex.FindStringSubmatchIndex(requestBody[pos:])
		if len(commentMatches) > 0 && commentMatches[0] == 0 {
			comment = requestBody[pos+commentMatches[2] : pos+commentMatches[3]]
			pos += commentMatches[1]
		}

		// 查找消息定义
		messageMatches := messageRegex.FindStringSubmatchIndex(requestBody[pos:])
		if len(messageMatches) == 0 {
			break
		}

		msgNameStart := pos + messageMatches[2]
		msgNameEnd := pos + messageMatches[3]
		msgBodyStart := pos + messageMatches[4]
		msgBodyEnd := pos + messageMatches[5]

		msgName := requestBody[msgNameStart:msgNameEnd]
		msgBody := requestBody[msgBodyStart:msgBodyEnd]

		// 忽略Rsp消息
		if msgName != "Rsp" {
			if *debugMode {
				fmt.Printf("DEBUG: Found message '%s'\n", msgName)
			}

			message := Message{
				Name:    msgName,
				Comment: comment,
			}

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

			// 查找Rsp子消息
			rspRegex := regexp.MustCompile(`message\s+Rsp\s*\{`)
			if rspRegex.MatchString(msgBody) {
				message.Response = "Request_" + msgName + "_Rsp"
				if *debugMode {
					fmt.Printf("DEBUG: Found Rsp message for %s\n", msgName)
				}
			} else {
				message.Response = "OK" // 默认使用通用成功OK
			}

			messages = append(messages, message)
		}

		pos += messageMatches[1]
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

	// 使用更灵活的正则表达式来查找嵌套的消息定义
	// 先找到注释，然后找到消息定义
	commentRegex := regexp.MustCompile(`(?s)//([^\n]*(?:\n\s*//[^\n]*)*)`)
	messageRegex := regexp.MustCompile(`(?s)message\s+(\w+)\s+\{([^{}]*(?:\{[^{}]*\}[^{}]*)*)\}`)

	// 在notifyBody中查找所有消息定义
	pos := 0
	for pos < len(notifyBody) {
		// 查找注释
		comment := ""
		commentMatches := commentRegex.FindStringSubmatchIndex(notifyBody[pos:])
		if len(commentMatches) > 0 && commentMatches[0] == 0 {
			comment = notifyBody[pos+commentMatches[2] : pos+commentMatches[3]]
			pos += commentMatches[1]
		}

		// 查找消息定义
		messageMatches := messageRegex.FindStringSubmatchIndex(notifyBody[pos:])
		if len(messageMatches) == 0 {
			break
		}

		msgNameStart := pos + messageMatches[2]
		msgNameEnd := pos + messageMatches[3]
		msgBodyStart := pos + messageMatches[4]
		msgBodyEnd := pos + messageMatches[5]

		msgName := notifyBody[msgNameStart:msgNameEnd]
		msgBody := notifyBody[msgBodyStart:msgBodyEnd]

		if *debugMode {
			fmt.Printf("DEBUG: Found message '%s'\n", msgName)
		}

		message := Message{
			Name:    msgName,
			Comment: comment,
		}

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

		messages = append(messages, message)

		pos += messageMatches[1]
	}

	return messages
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
	writer.WriteString("import (\n")
	writer.WriteString("\t\"fmt\"\n")
	writer.WriteString("\t\"google.golang.org/protobuf/proto\"\n")
	writer.WriteString("\t\"github.com/orbit-w/orbit/lib/utils/proto_utils\"\n")
	writer.WriteString(")\n\n")

	// 写入请求处理接口
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

	// 为每个包创建唯一的getResponsePID函数名
	responsePIDFuncName := fmt.Sprintf("get%sResponsePID", packageName)

	// 写入getResponsePID辅助函数，添加包名前缀确保唯一性
	writer.WriteString(fmt.Sprintf("// %s 通过反射获取响应消息的协议ID\n", responsePIDFuncName))
	writer.WriteString(fmt.Sprintf("func %s(response any) uint32 {\n", responsePIDFuncName))
	writer.WriteString("\t// 获取消息名称\n")
	writer.WriteString("\ttypeName := proto_utils.ParseMessageName(response)\n")
	writer.WriteString("\tif typeName == \"\" {\n")
	writer.WriteString("\t\treturn 0\n")
	writer.WriteString("\t}\n\n")

	// 通过名称查找PID - 通用方式处理所有类型
	writer.WriteString("\t// 查找类型对应的协议ID\n")
	writer.WriteString(fmt.Sprintf("\tpid, ok := Get%sProtocolID(typeName)\n", packageName))
	writer.WriteString("\tif ok {\n")
	writer.WriteString("\t\treturn pid\n")
	writer.WriteString("\t}\n\n")

	writer.WriteString("\t// 未找到对应的协议ID\n")
	writer.WriteString("\treturn 0\n")
	writer.WriteString("}\n\n")

	// 写入请求分发函数 - 基于消息名称
	writer.WriteString(fmt.Sprintf("// Dispatch%sRequest 分发%s包的请求消息到对应处理函数\n", packageName, packageName))
	writer.WriteString(fmt.Sprintf("func Dispatch%sRequest(handler %sRequestHandler, msgName string, data []byte) (any, uint32, error) {\n", packageName, packageName))
	writer.WriteString("\tvar err error\n")
	writer.WriteString("\tvar response any\n")
	writer.WriteString("\tvar responsePid uint32\n\n")
	writer.WriteString("\tswitch msgName {\n")

	for _, msg := range messages {
		writer.WriteString(fmt.Sprintf("\tcase \"%s\":\n", msg.Name))
		writer.WriteString(fmt.Sprintf("\t\treq := &Request_%s{}\n", msg.Name))
		writer.WriteString("\t\tif err = proto.Unmarshal(data, req); err != nil {\n")
		writer.WriteString(fmt.Sprintf("\t\t\treturn nil, 0, fmt.Errorf(\"unmarshal %s failed: %%v\", err)\n", msg.Name))
		writer.WriteString("\t\t}\n")
		writer.WriteString(fmt.Sprintf("\t\tresponse = handler.Handle%s(req)\n", msg.Name))
		// 使用添加了包名前缀的函数
		writer.WriteString(fmt.Sprintf("\t\tresponsePid = %s(response)\n", responsePIDFuncName))
	}

	writer.WriteString("\tdefault:\n")
	writer.WriteString("\t\treturn nil, 0, fmt.Errorf(\"unknown request message: %s\", msgName)\n")
	writer.WriteString("\t}\n\n")
	writer.WriteString("\treturn response, responsePid, nil\n")
	writer.WriteString("}\n\n")

	// 写入基于协议ID的请求分发函数
	writer.WriteString(fmt.Sprintf("// Dispatch%sRequestByID 通过协议ID分发%s包的请求消息到对应处理函数\n", packageName, packageName))
	writer.WriteString(fmt.Sprintf("func Dispatch%sRequestByID(handler %sRequestHandler, pid uint32, data []byte) (any, uint32, error) {\n", packageName, packageName))
	writer.WriteString("\tvar err error\n")
	writer.WriteString("\tvar response any\n")
	writer.WriteString("\tvar responsePid uint32\n\n")
	writer.WriteString("\tswitch pid {\n")

	for _, msg := range messages {
		writer.WriteString(fmt.Sprintf("\tcase PID_%s_%s:\n", packageName, msg.Name))
		writer.WriteString(fmt.Sprintf("\t\treq := &Request_%s{}\n", msg.Name))
		writer.WriteString("\t\tif err = proto.Unmarshal(data, req); err != nil {\n")
		writer.WriteString(fmt.Sprintf("\t\t\treturn nil, 0, fmt.Errorf(\"unmarshal message with ID 0x%%08x failed: %%v\", pid, err)\n"))
		writer.WriteString("\t\t}\n")
		writer.WriteString(fmt.Sprintf("\t\tresponse = handler.Handle%s(req)\n", msg.Name))
		// 使用添加了包名前缀的函数
		writer.WriteString(fmt.Sprintf("\t\tresponsePid = %s(response)\n", responsePIDFuncName))
	}

	writer.WriteString("\tdefault:\n")
	writer.WriteString("\t\treturn nil, 0, fmt.Errorf(\"unknown protocol ID: 0x%08x\", pid)\n")
	writer.WriteString("\t}\n\n")
	writer.WriteString("\treturn response, responsePid, nil\n")
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
	writer.WriteString("import (\n")
	writer.WriteString("\t\"fmt\"\n")
	writer.WriteString("\t\"google.golang.org/protobuf/proto\"\n")
	writer.WriteString("\t\"github.com/orbit-w/orbit/lib/utils/proto_utils\"\n")
	writer.WriteString(")\n\n")

	// 为每个包创建唯一的getNotificationPID函数名
	notificationPIDFuncName := fmt.Sprintf("get%sNotificationPID", packageName)

	// 添加getNotificationPID辅助函数，添加包名前缀确保唯一性
	writer.WriteString(fmt.Sprintf("// %s 通过反射获取通知消息的协议ID\n", notificationPIDFuncName))
	writer.WriteString(fmt.Sprintf("func %s(notification any) uint32 {\n", notificationPIDFuncName))
	writer.WriteString("\t// 获取消息名称\n")
	writer.WriteString("\ttypeName := proto_utils.ParseMessageName(notification)\n")
	writer.WriteString("\tif typeName == \"\" {\n")
	writer.WriteString("\t\treturn 0\n")
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
		writer.WriteString(fmt.Sprintf("func Marshal%s(notify *Notify_%s) ([]byte, uint32, error) {\n", msg.Name, msg.Name))
		writer.WriteString("\tdata, err := proto.Marshal(notify)\n")
		// 对于已知消息类型，我们可以直接返回其固定的协议ID
		writer.WriteString(fmt.Sprintf("\treturn data, PID_%s_%s, err\n", packageName, msg.Name))
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
		writer.WriteString(fmt.Sprintf("\tcase \"%s\":\n", msg.Name))
		writer.WriteString(fmt.Sprintf("\t\tnotify := &Notify_%s{}\n", msg.Name))
		writer.WriteString("\t\tif err = proto.Unmarshal(data, notify); err != nil {\n")
		writer.WriteString(fmt.Sprintf("\t\t\treturn nil, 0, fmt.Errorf(\"unmarshal %s notification failed: %%v\", err)\n", msg.Name))
		writer.WriteString("\t\t}\n")
		writer.WriteString("\t\tnotification = notify\n")
		// 使用添加了包名前缀的函数
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

	for _, msg := range messages {
		writer.WriteString(fmt.Sprintf("\tcase PID_%s_%s:\n", packageName, msg.Name))
		writer.WriteString(fmt.Sprintf("\t\tnotify := &Notify_%s{}\n", msg.Name))
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
