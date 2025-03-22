package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	protoDir  = flag.String("proto_dir", "app/proto", "Directory containing .proto files")
	debugMode = flag.Bool("debug", false, "Enable debug mode")
	quietMode = flag.Bool("quiet", true, "Quiet mode: only show errors")
)

// 消息结构，用于存储消息定义及其注释
type Message struct {
	Name     string
	Comment  string
	Fields   []Field
	Response string // 响应消息名称，如果有的话
}

// 字段结构，用于存储字段定义及其注释
type Field struct {
	Name    string
	Type    string
	Index   int
	Comment string
}

func main() {
	flag.Parse()

	// 确保proto目录存在
	if _, err := os.Stat(*protoDir); os.IsNotExist(err) {
		fmt.Printf("Proto directory %s does not exist\n", *protoDir)
		return
	}

	// 创建pb目录，如果不存在
	pbDir := filepath.Join(*protoDir, "pb")
	if _, err := os.Stat(pbDir); os.IsNotExist(err) {
		if err := os.MkdirAll(pbDir, 0755); err != nil {
			fmt.Printf("Failed to create pb directory: %v\n", err)
			return
		}
	}

	// 遍历proto目录中的所有.proto文件
	protoFiles, err := findProtoFiles(*protoDir)
	if err != nil {
		fmt.Printf("Error finding proto files: %v\n", err)
		return
	}

	for _, protoFile := range protoFiles {
		if !*quietMode {
			fmt.Printf("Processing %s...\n", protoFile)
		}
		processProtoFile(protoFile, pbDir)
	}

	if !*quietMode {
		fmt.Println("Glue code generation completed!")
	}
}

// 查找指定目录下的所有.proto文件
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

// 处理单个proto文件
func processProtoFile(protoFile, pbDir string) {
	// 读取proto文件内容
	content, err := os.ReadFile(protoFile)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", protoFile, err)
		return
	}

	// 解析包名
	packageName := parsePackageName(string(content))
	if packageName == "" {
		fmt.Printf("Package name not found in %s\n", protoFile)
		return
	}

	if !*quietMode {
		fmt.Printf("Found package name: %s\n", packageName)
	}

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
		generateRequestGlueCode(requestMessages, packageName, pbDir)
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
		generateNotifyGlueCode(notifyMessages, packageName, pbDir)
	} else if !*quietMode {
		fmt.Printf("No notify messages found\n")
	}
}

// 解析包名
func parsePackageName(content string) string {
	re := regexp.MustCompile(`package\s+([^;]+);`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
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
	writer.WriteString("import (\n\t\"fmt\"\n\t\"google.golang.org/protobuf/proto\"\n)\n\n")

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

	// 写入请求分发函数
	writer.WriteString(fmt.Sprintf("// Dispatch%sRequest 分发%s包的请求消息到对应处理函数\n", packageName, packageName))
	writer.WriteString(fmt.Sprintf("func Dispatch%sRequest(handler %sRequestHandler, msgName string, msgData []byte) (any, error) {\n", packageName, packageName))
	writer.WriteString("\tvar err error\n")
	writer.WriteString("\tvar response any\n\n")
	writer.WriteString("\tswitch msgName {\n")

	for _, msg := range messages {
		writer.WriteString(fmt.Sprintf("\tcase \"%s\":\n", msg.Name))
		writer.WriteString(fmt.Sprintf("\t\treq := &Request_%s{}\n", msg.Name))
		writer.WriteString("\t\tif err = proto.Unmarshal(msgData, req); err != nil {\n")
		writer.WriteString(fmt.Sprintf("\t\t\treturn nil, fmt.Errorf(\"unmarshal %s failed: %%v\", err)\n", msg.Name))
		writer.WriteString("\t\t}\n")
		writer.WriteString(fmt.Sprintf("\t\tresponse = handler.Handle%s(req)\n", msg.Name))
	}

	writer.WriteString("\tdefault:\n")
	writer.WriteString("\t\treturn nil, fmt.Errorf(\"unknown request message: %s\", msgName)\n")
	writer.WriteString("\t}\n\n")
	writer.WriteString("\treturn response, nil\n")
	writer.WriteString("}\n")

	writer.Flush()

	if !*quietMode {
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
	writer.WriteString("import (\n\t\"fmt\"\n\t\"google.golang.org/protobuf/proto\"\n)\n\n")

	// 写入通知序列化函数
	for _, msg := range messages {
		writer.WriteString(fmt.Sprintf("// Marshal%s 序列化%s通知消息\n", msg.Name, msg.Name))
		if msg.Comment != "" {
			writer.WriteString(fmt.Sprintf("// %s\n", msg.Comment))
		}
		writer.WriteString(fmt.Sprintf("func Marshal%s(notify *Notify_%s) ([]byte, error) {\n", msg.Name, msg.Name))
		writer.WriteString("\treturn proto.Marshal(notify)\n")
		writer.WriteString("}\n\n")
	}

	// 写入通知解析函数
	writer.WriteString(fmt.Sprintf("// Parse%sNotify 根据消息名称解析%s包的通知消息\n", packageName, packageName))
	writer.WriteString(fmt.Sprintf("func Parse%sNotify(msgName string, msgData []byte) (any, error) {\n", packageName))
	writer.WriteString("\tvar err error\n")
	writer.WriteString("\tvar notification any\n\n")
	writer.WriteString("\tswitch msgName {\n")

	for _, msg := range messages {
		writer.WriteString(fmt.Sprintf("\tcase \"%s\":\n", msg.Name))
		writer.WriteString(fmt.Sprintf("\t\tnotify := &Notify_%s{}\n", msg.Name))
		writer.WriteString("\t\tif err = proto.Unmarshal(msgData, notify); err != nil {\n")
		writer.WriteString(fmt.Sprintf("\t\t\treturn nil, fmt.Errorf(\"unmarshal %s notification failed: %%v\", err)\n", msg.Name))
		writer.WriteString("\t\t}\n")
		writer.WriteString("\t\tnotification = notify\n")
	}

	writer.WriteString("\tdefault:\n")
	writer.WriteString("\t\treturn nil, fmt.Errorf(\"unknown notification message: %s\", msgName)\n")
	writer.WriteString("\t}\n\n")
	writer.WriteString("\treturn notification, nil\n")
	writer.WriteString("}\n")

	writer.Flush()

	if !*quietMode {
		fmt.Printf("Generated notify glue code: %s\n", filename)
	}
}
