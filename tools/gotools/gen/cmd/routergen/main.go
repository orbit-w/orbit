// Package main provides a tool for generating protocol IDs and glue code from proto message definitions
package routergen

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"gitee.com/orbit-w/orbit/lib/base/protoid"
	descriptor "github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
)

var (
	genGlueCmd = &cobra.Command{
		Use:   "routergen",
		Short: "Generate protocol IDs and glue code from proto files",
		Run:   runRouterGluegen,
	}
)

func InitCmd(father *cobra.Command) {
	genGlueCmd.Flags().String("proto-dir", "app/proto/pb", "Directory containing proto files")
	genGlueCmd.Flags().String("output-dir", "app/proto/pb", "Output directory for generated files")
	genGlueCmd.Flags().Bool("debug", false, "Enable debug mode")
	genGlueCmd.Flags().Bool("quiet", false, "Enable quiet mode")
	genGlueCmd.Flags().Bool("gen-proto-code", true, "Generate glue code")
	genGlueCmd.Flags().Bool("gen-proto-ids", true, "Generate protocol IDs")
	genGlueCmd.Flags().String("proto-file", "", "Specific proto file to process (or directory)")

	father.AddCommand(genGlueCmd)
}

func runRouterGluegen(cmd *cobra.Command, args []string) {
	protoDir, _ := cmd.Flags().GetString("proto-dir")
	outputDir, _ := cmd.Flags().GetString("output-dir")
	debugMode, _ := cmd.Flags().GetBool("debug")
	quietMode, _ := cmd.Flags().GetBool("quiet")
	genProtoCode, _ := cmd.Flags().GetBool("gen-proto-code")
	genProtoIDs, _ := cmd.Flags().GetBool("gen-proto-ids")
	protoFile, _ := cmd.Flags().GetString("proto-file")

	// 确保output目录存在
	if err := ensureOutputDir(outputDir); err != nil {
		fmt.Printf("Failed to create output directory: %v\n", err)
		return
	}

	protoFiles, err := getProtoFiles(protoFile, protoDir)
	if err != nil {
		fmt.Printf("Error getting proto files: %v\n", err)
		return
	}

	// 创建Context并解析proto文件
	ctx := NewContext(protoFiles, outputDir, protoDir)
	allMappings, err := processProtoFiles(ctx, quietMode, debugMode, genProtoIDs, genProtoCode)
	if err != nil {
		fmt.Printf("Error processing proto files: %v\n", err)
		return
	}

	if genProtoIDs && len(allMappings) > 0 {
		if err := generateCommonProtocolMappings(allMappings, outputDir, quietMode); err != nil {
			fmt.Printf("Error generating protocol mappings: %v\n", err)
			return
		}
	}

	if !quietMode {
		fmt.Println("All generation completed!")
	}
}

func ensureOutputDir(outputDir string) error {
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		return os.MkdirAll(outputDir, 0755)
	}
	return nil
}

func getProtoFiles(protoFile, protoDir string) ([]string, error) {
	if protoFile != "" {
		var err error
		protoFile, err = filepath.Abs(protoFile)
		if err != nil {
			return nil, fmt.Errorf("resolving proto file path: %w", err)
		}
		fileInfo, err := os.Stat(protoFile)
		if err != nil {
			return nil, fmt.Errorf("accessing path: %w", err)
		}
		if fileInfo.IsDir() {
			return findProtoFiles(protoFile)
		}
		return []string{protoFile}, nil
	}
	return findProtoFiles(protoDir)
}

func processProtoFiles(ctx *Context, quietMode, debugMode, genProtoIDs, genProtoCode bool) ([]ProtocolIDMapping, error) {
	var allMappings []ProtocolIDMapping

	fds, err := ctx.ParseProtoFiles()
	if err != nil {
		return nil, fmt.Errorf("parsing proto files: %w", err)
	}

	// Create a set of expected proto file basenames
	expectedFiles := make(map[string]struct{})
	for _, p := range ctx.GetProtoFiles() {
		expectedFiles[filepath.Base(p)] = struct{}{}
	}

	for _, fd := range fds {
		fileName := filepath.Base(fd.GetName())
		if _, ok := expectedFiles[fileName]; !ok {
			// Skip imported/dependency files
			continue
		}

		mapping, err := processFileDescriptor(fd, ctx, quietMode, debugMode, genProtoIDs, genProtoCode)
		if err != nil {
			if !quietMode {
				fmt.Printf("Error processing %s: %v\n", fd.GetName(), err)
			}
			continue
		}

		if mapping != nil && len(mapping.MessageIDs) > 0 {
			allMappings = append(allMappings, *mapping)
		}
	}

	return allMappings, nil
}

// processFileDescriptor 处理单个文件描述符
// 返回生成的协议ID映射和可能的错误
func processFileDescriptor(fd *descriptor.FileDescriptorProto, ctx *Context, quietMode, debugMode, genProtoIDs, genProtoCode bool) (*ProtocolIDMapping, error) {
	if !quietMode || debugMode {
		fmt.Printf("Processing %s...\n", fd.GetName())
	}

	packageName := fd.GetPackage()
	if packageName == "" {
		if !quietMode {
			fmt.Printf("Package name not found in %s\n", fd.GetName())
		}
		return nil, fmt.Errorf("package name not found in %s", fd.GetName())
	}

	if !quietMode {
		fmt.Printf("Found package name: %s\n", packageName)
	}

	var mapping *ProtocolIDMapping

	// 生成协议ID
	if genProtoIDs {
		protoMapping := generateProtocolIDs(fd, packageName, debugMode)
		if len(protoMapping.MessageIDs) > 0 {
			mapping = &protoMapping
		}
	}

	// 生成协议代码
	if genProtoCode {
		if err := processRequestMessages(fd, packageName, ctx, quietMode, debugMode); err != nil {
			return mapping, fmt.Errorf("processing request messages: %w", err)
		}

		if err := processNotifyMessages(fd, packageName, ctx, quietMode, debugMode); err != nil {
			return mapping, fmt.Errorf("processing notify messages: %w", err)
		}
	}

	return mapping, nil
}

// processRequestMessages 处理请求消息
func processRequestMessages(fd *descriptor.FileDescriptorProto, packageName string, ctx *Context, quietMode, debugMode bool) error {
	requestMessages := parseRequestMessages(fd, debugMode)
	if len(requestMessages) > 0 {
		if !quietMode {
			fmt.Printf("Found %d request messages\n", len(requestMessages))
			if debugMode {
				for i, msg := range requestMessages {
					fmt.Printf("  Request %d: Name=%s, FullName=%s\n", i+1, msg.Name, msg.FullName)
				}
			}
		}
		if err := generateRequestGlueCode(requestMessages, packageName, ctx, quietMode); err != nil {
			if !quietMode {
				fmt.Printf("Error generating request glue code: %v\n", err)
			}
			return err
		}
	} else if !quietMode {
		fmt.Printf("No request messages found\n")
	}
	return nil
}

// processNotifyMessages 处理通知消息
func processNotifyMessages(fd *descriptor.FileDescriptorProto, packageName string, ctx *Context, quietMode, debugMode bool) error {
	notifyMessages := parseNotifyMessages(fd, debugMode)
	if len(notifyMessages) > 0 {
		if !quietMode {
			fmt.Printf("Found %d notify messages\n", len(notifyMessages))
			if debugMode {
				for i, msg := range notifyMessages {
					fmt.Printf("  Notify %d: Name=%s, FullName=%s\n", i+1, msg.Name, msg.FullName)
				}
			}
		}
		if err := generateNotifyGlueCode(notifyMessages, packageName, ctx, quietMode); err != nil {
			if !quietMode {
				fmt.Printf("Error generating notify glue code: %v\n", err)
			}
			return err
		}
	} else if !quietMode {
		fmt.Printf("No notify messages found\n")
	}
	return nil
}

// generateProtocolIDs 生成协议ID
func generateProtocolIDs(fd *descriptor.FileDescriptorProto, packageName string, debugMode bool) ProtocolIDMapping {
	// 提取所有消息名称，包括嵌套
	messageNames := extractMessageNamesFromDescriptor(fd.MessageType, "")
	if len(messageNames) == 0 {
		if debugMode {
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
			if debugMode {
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
				if debugMode {
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
			if debugMode {
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
		if debugMode {
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

// extractMessageNamesFromDescriptor recursively extracts message names from descriptor
func extractMessageNamesFromDescriptor(messages []*descriptor.DescriptorProto, parent string) []MessageName {
	var names []MessageName
	for _, msg := range messages {
		fullName := msg.GetName()
		if parent != "" {
			fullName = parent + "_" + fullName
		}
		names = append(names, MessageName{
			Name:     msg.GetName(),
			FullName: fullName,
		})
		names = append(names, extractMessageNamesFromDescriptor(msg.NestedType, fullName)...)
	}
	return names
}

// 解析Request消息
func parseRequestMessages(fd *descriptor.FileDescriptorProto, debugMode bool) []Message {
	var messages []Message

	// Find the top-level Request message
	var requestMsg *descriptor.DescriptorProto
	for _, msg := range fd.MessageType {
		if msg.GetName() == "Request" {
			requestMsg = msg
			break
		}
	}
	if requestMsg == nil {
		if debugMode {
			fmt.Println("DEBUG: Request message not found")
		}
		return messages
	}

	if debugMode {
		fmt.Println("DEBUG: Found Request message")
	}

	// Extract nested messages from Request
	allMessageNames := extractMessageNamesFromDescriptor(requestMsg.NestedType, "Request")
	if debugMode {
		fmt.Printf("DEBUG: Found %d nested messages in Request block\n", len(allMessageNames))
		for i, msg := range allMessageNames {
			fmt.Printf("DEBUG:   Message %d: Name=%s, FullName=%s\n", i+1, msg.Name, msg.FullName)
		}
	}

	// 过滤出Request_前缀的消息
	for _, msgInfo := range allMessageNames {
		// 忽略Rsp消息和通用的Request类型本身
		if msgInfo.Name == "Rsp" || msgInfo.Name == "Request" {
			if debugMode {
				fmt.Printf("DEBUG: Skipping base message type: %s\n", msgInfo.Name)
			}
			continue
		}

		fullName := msgInfo.FullName
		if !strings.HasPrefix(fullName, "Request_") {
			// 确保消息名称有正确的前缀
			fullName = "Request_" + fullName
		}

		if debugMode {
			fmt.Printf("DEBUG: Processing request message '%s'\n", fullName)
		}

		// Find the corresponding DescriptorProto
		msgDesc := findMessageDescriptor(requestMsg.NestedType, msgInfo.Name)
		if msgDesc == nil {
			continue
		}

		message := Message{
			Name:     msgInfo.Name,
			Comment:  extractMessageCommentFromDescriptor(fd, msgDesc), // Implement this if needed
			FullName: fullName,
		}

		// Parse fields
		for i, f := range msgDesc.Field {
			field := Field{
				Type:    f.GetTypeName(), // Or resolve properly
				Name:    f.GetName(),
				Index:   i + 1,
				Comment: extractFieldCommentFromDescriptor(fd, f), // Implement if needed
			}
			message.Fields = append(message.Fields, field)
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
			if debugMode {
				fmt.Printf("DEBUG: Found Rsp message for %s: %s\n", message.Name, message.Response)
			}
		} else {
			message.Response = "OK" // 默认使用通用成功OK
		}

		messages = append(messages, message)
	}

	return messages
}

// findMessageDescriptor finds a message descriptor by name in nested types
func findMessageDescriptor(messages []*descriptor.DescriptorProto, name string) *descriptor.DescriptorProto {
	for _, msg := range messages {
		if msg.GetName() == name {
			return msg
		}
		if nested := findMessageDescriptor(msg.NestedType, name); nested != nil {
			return nested
		}
	}
	return nil
}

// extractMessageCommentFromDescriptor extracts comment for a message (stub, implement if needed)
func extractMessageCommentFromDescriptor(fd *descriptor.FileDescriptorProto, msg *descriptor.DescriptorProto) string {
	// Use fd.SourceCodeInfo to get comments
	// This is a stub; implement proper extraction if comments are crucial
	return ""
}

// extractFieldCommentFromDescriptor extracts comment for a field (stub)
func extractFieldCommentFromDescriptor(fd *descriptor.FileDescriptorProto, field *descriptor.FieldDescriptorProto) string {
	return ""
}

// 解析Notify消息
func parseNotifyMessages(fd *descriptor.FileDescriptorProto, debugMode bool) []Message {
	var messages []Message

	// Find the top-level Notify message
	var notifyMsg *descriptor.DescriptorProto
	for _, msg := range fd.MessageType {
		if msg.GetName() == "Notify" {
			notifyMsg = msg
			break
		}
	}
	if notifyMsg == nil {
		if debugMode {
			fmt.Println("DEBUG: Notify message not found")
		}
		return messages
	}

	if debugMode {
		fmt.Println("DEBUG: Found Notify message")
	}

	// Extract nested messages from Notify
	allMessageNames := extractMessageNamesFromDescriptor(notifyMsg.NestedType, "Notify")
	if debugMode {
		fmt.Printf("DEBUG: Found %d nested messages in Notify block\n", len(allMessageNames))
		for i, msg := range allMessageNames {
			fmt.Printf("DEBUG:   Message %d: Name=%s, FullName=%s\n", i+1, msg.Name, msg.FullName)
		}
	}

	// 过滤出Notify_前缀的消息
	for _, msgInfo := range allMessageNames {
		// 忽略通用的Notify类型本身
		if msgInfo.Name == "Notify" {
			if debugMode {
				fmt.Printf("DEBUG: Skipping base message type: %s\n", msgInfo.Name)
			}
			continue
		}

		fullName := msgInfo.FullName
		if !strings.HasPrefix(fullName, "Notify_") {
			// 确保消息名称有正确的前缀
			fullName = "Notify_" + fullName
		}

		if debugMode {
			fmt.Printf("DEBUG: Processing notify message '%s'\n", fullName)
		}

		// Find the corresponding DescriptorProto
		msgDesc := findMessageDescriptor(notifyMsg.NestedType, msgInfo.Name)
		if msgDesc == nil {
			continue
		}

		message := Message{
			Name:     msgInfo.Name,
			Comment:  extractMessageCommentFromDescriptor(fd, msgDesc),
			FullName: fullName,
		}

		// Parse fields
		for i, f := range msgDesc.Field {
			field := Field{
				Type:    f.GetTypeName(),
				Name:    f.GetName(),
				Index:   i + 1,
				Comment: extractFieldCommentFromDescriptor(fd, f),
			}
			message.Fields = append(message.Fields, field)
		}

		messages = append(messages, message)
	}

	return messages
}

// 提取go_package值
func extractGoPackage(fd *descriptor.FileDescriptorProto) string {
	goPkg := fd.Options.GetGoPackage()
	if goPkg == "" {
		return ""
	}
	parts := strings.Split(goPkg, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return goPkg
}

// 生成Request消息的胶水代码
func generateRequestGlueCode(messages []Message, packageName string, ctx *Context, quietMode bool) error {
	// 构建输出文件名
	outputFile := filepath.Join(ctx.GetOutputDir(), strings.ToLower(packageName)+"_request_glue.go")
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("creating request glue file: %w", err)
	}
	defer file.Close()

	// 文件头
	fmt.Fprintf(file, "// Code generated by genproto. DO NOT EDIT.\n")
	fmt.Fprintf(file, "package pb\n\n")

	// 导入必要的包
	fmt.Fprintf(file, "import (\n")
	fmt.Fprintf(file, "\t\"fmt\"\n")
	fmt.Fprintf(file, "\t\"github.com/gogo/protobuf/proto\"\n")

	// 使用Context中缓存的fds数据
	var goPackage string
	fds, err := ctx.ParseProtoFiles()
	if err != nil {
		return fmt.Errorf("getting cached fds for glue code: %w", err)
	}
	for _, fd := range fds {
		extractedPackage := fd.GetPackage()
		if extractedPackage == packageName {
			goPackage = extractGoPackage(fd)
			break
		}
	}

	// 如果找不到go_package，使用默认的包名
	if goPackage == "" {
		goPackage = strings.ToLower(packageName)
		fmt.Fprintf(file, "\t\"gitee.com/orbit-w/orbit/app/proto/pb/%s\"\n", goPackage)
	} else {
		fmt.Fprintf(file, "\t\"gitee.com/orbit-w/orbit/app/proto/pb/%s\"\n", goPackage)
	}

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
		fmt.Fprintf(file, "\tHandle%s(req *%s.%s) proto.Message\n", msg.Name, goPackage, msg.FullName)
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
		fmt.Fprintf(file, "\t\treq := &%s.%s{}\n", goPackage, msg.FullName)
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

	if !quietMode {
		fmt.Printf("Generated request glue code in %s\n", outputFile)
	}
	return nil
}

// 生成Notify消息的胶水代码
func generateNotifyGlueCode(messages []Message, packageName string, ctx *Context, quietMode bool) error {
	// 构建输出文件名
	outputFile := filepath.Join(ctx.GetOutputDir(), strings.ToLower(packageName)+"_notify_glue.go")
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("creating notify glue file: %w", err)
	}
	defer file.Close()

	// 文件头
	fmt.Fprintf(file, "// Code generated by genproto. DO NOT EDIT.\n")
	fmt.Fprintf(file, "package pb\n\n")

	// 导入必要的包
	fmt.Fprintf(file, "import (\n")
	fmt.Fprintf(file, "\t\"fmt\"\n")
	fmt.Fprintf(file, "\t\"github.com/gogo/protobuf/proto\"\n")

	// 使用Context中缓存的fds数据
	var goPackage string
	fds, err := ctx.ParseProtoFiles()
	if err != nil {
		return fmt.Errorf("getting cached fds for glue code: %w", err)
	}
	for _, fd := range fds {
		extractedPackage := fd.GetPackage()
		if extractedPackage == packageName {
			goPackage = extractGoPackage(fd)
			break
		}
	}

	// 如果找不到go_package，使用默认的包名
	if goPackage == "" {
		goPackage = strings.ToLower(packageName)
		fmt.Fprintf(file, "\t\"gitee.com/orbit-w/orbit/app/proto/pb/%s\"\n", goPackage)
	} else {
		fmt.Fprintf(file, "\t\"gitee.com/orbit-w/orbit/app/proto/pb/%s\"\n", goPackage)
	}

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
		fmt.Fprintf(file, "\t\tnotify := &%s.%s{}\n", goPackage, msg.FullName)
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
		fmt.Fprintf(file, "func Marshal%s(notify *%s.%s) ([]byte, uint32, error) {\n", msg.Name, goPackage, msg.FullName)
		fmt.Fprintf(file, "\tdata, err := proto.Marshal(notify)\n")
		fmt.Fprintf(file, "\treturn data, PID_%s_%s, err\n", packageName, msg.FullName)
		fmt.Fprintf(file, "}\n\n")
	}

	if !quietMode {
		fmt.Printf("Generated notify glue code in %s\n", outputFile)
	}
	return nil
}

// generateCommonProtocolMappings 生成公共的协议ID映射文件
func generateCommonProtocolMappings(allMappings []ProtocolIDMapping, outputDir string, quietMode bool) error {
	outputFile := filepath.Join(outputDir, "protocol_ids.go")

	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("creating protocol ID file: %w", err)
	}
	defer file.Close()

	// 文件头
	fmt.Fprintf(file, "// Code generated by protocol ID generator. DO NOT EDIT.\n")
	fmt.Fprintf(file, "package pb\n\n")

	// 导入必要的包
	fmt.Fprintf(file, "import (\n")
	fmt.Fprintf(file, "\t\"fmt\"\n")
	fmt.Fprintf(file, "\t\"github.com/gogo/protobuf/proto\"\n")
	fmt.Fprintf(file, "\t\"gitee.com/orbit-w/orbit/lib/utils/proto_utils\"\n")
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

	if !quietMode {
		fmt.Printf("Generated common protocol ID mappings in %s\n", outputFile)
	}
	return nil
}
