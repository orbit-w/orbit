// Package main provides a tool for generating protocol IDs and glue code from proto message definitions
package protocolgen

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"gitee.com/orbit-w/orbit/lib/base/protoid"
	descriptor "github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
)

var (
	genGlueCmd = &cobra.Command{
		Use:   "protocolgen",
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
	genGlueCmd.Flags().Bool("gen-pb-go", true, "Generate pb.go files using protoc")
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
	genPbGo, _ := cmd.Flags().GetBool("gen-pb-go")
	protoFile, _ := cmd.Flags().GetString("proto-file")

	// 确定运行模式
	var mode Mode
	if debugMode {
		mode = ModeDebug
	} else if quietMode {
		mode = ModeQuiet
	} else {
		mode = ModeNormal
	}

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

	// 首先生成 pb.go 文件
	if genPbGo {
		if err := generatePbGoFiles(protoFiles, outputDir, mode); err != nil {
			fmt.Printf("Error generating pb.go files: %v\n", err)
			return
		}
	}

	// 创建Context并配置选项
	ctx := NewContext(protoFiles, outputDir, protoDir)
	ctx.SetMode(mode)
	ctx.SetGenProtoIDs(genProtoIDs)
	ctx.SetGenProtoCode(genProtoCode)

	// 解析proto文件, 并生成请求和通知的胶水代码
	err = processProtoFiles(ctx)
	if err != nil {
		fmt.Printf("Error processing proto files: %v\n", err)
		return
	}

	// 生成协议ID
	if ctx.GetGenProtoIDs() {
		protoMapping := generateProtocolIDs(ctx)
		if err := generateCommonProtocolMappings(protoMapping, outputDir, ctx.GetMode()); err != nil {
			fmt.Printf("Error generating protocol mappings: %v\n", err)
			return
		}
	}

	if ctx.GetMode().ShouldPrint() {
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

func processProtoFiles(ctx *Context) error {
	fds, err := ctx.ParseProtoFiles()
	if err != nil {
		return fmt.Errorf("parsing proto files: %w", err)
	}

	// Create a set of expected proto file basenames
	expectedFiles := make(map[string]struct{})
	for _, p := range ctx.GetProtoFiles() {
		expectedFiles[filepath.Base(p)] = struct{}{}
	}

	for i := range fds {
		fd := fds[i]
		fileName := filepath.Base(fd.GetName())
		if _, ok := expectedFiles[fileName]; !ok {
			// Skip imported/dependency files
			continue
		}

		err := parseFileDescriptor(fd, ctx)
		if err != nil {
			if ctx.GetMode().ShouldPrint() {
				fmt.Printf("Error processing %s: %v\n", fd.GetName(), err)
			}
			continue
		}
	}

	// 添加通用Response 消息结构
	// 添加Rsp_OK消息
	ctx.AddRspMessage(Message{
		Type:        MessageTypeRsp,
		Name:        CommonRspOK,
		FullName:    CommonRspOK,
		PackageName: "pb",
		PidName:     CommonRspOK,
	})

	// 添加Rsp_Fail消息
	ctx.AddRspMessage(Message{
		Type:        MessageTypeRsp,
		Name:        CommonRspFail,
		FullName:    CommonRspFail,
		PackageName: "pb",
		PidName:     CommonRspFail,
	})

	// 生成Request消息的胶水代码
	if err := generateRequestGlueCode(ctx); err != nil {
		if ctx.GetMode().ShouldPrint() {
			fmt.Printf("Error generating request glue code: %v\n", err)
		}
		return err
	}

	// 生成Notify消息的胶水代码
	if err := generateNotifyGlueCode(ctx); err != nil {
		if ctx.GetMode().ShouldPrint() {
			fmt.Printf("Error generating notify glue code: %v\n", err)
		}
		return err
	}

	return nil
}

// parseFileDescriptor 处理单个文件描述符
// 返回生成的协议ID映射和可能的错误
func parseFileDescriptor(fd *descriptor.FileDescriptorProto, ctx *Context) error {
	if ctx.GetMode().ShouldPrint() {
		fmt.Printf("Processing %s...\n", fd.GetName())
	}

	packageName := fd.GetPackage()
	if packageName == "" {
		if ctx.GetMode().ShouldPrint() {
			fmt.Printf("Package name not found in %s\n", fd.GetName())
		}
		return fmt.Errorf("package name not found in %s", fd.GetName())
	}

	// 解析协议消息
	parseRequestMessages(ctx, fd, packageName)

	parseNotifyMessages(ctx, fd, ctx.GetMode(), packageName)

	return nil
}

// generateProtocolIDs 生成协议ID
func generateProtocolIDs(ctx *Context) ProtocolIDMapping {
	// 生成协议ID映射
	mapping := ProtocolIDMapping{}

	messages := ctx.AllNetMessages()

	// 按名称排序以得到一致的输出
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].FullName < messages[j].FullName
	})

	for _, msg := range messages {
		pid := protoid.HashProtoMessage(msg.PidName)
		mapping.MessageIDs = append(mapping.MessageIDs, MessageID{
			Name: msg.FullName,
			ID:   pid,
		})
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

func genRequestMessageNamesFromDescriptor(dp, father *descriptor.DescriptorProto, packageName string) MessageName {
	msgInfo := MessageName{
		Type: MessageTypeRequest,
	}
	ReqfullName := father.GetName() + "_" + dp.GetName()
	msgInfo.Name = dp.GetName()
	msgInfo.ReqFullName = ReqfullName
	msgInfo.MsgWall = father.GetName()
	msgInfo.SetPackageName(packageName)
	msgInfo.DP = dp

	// 解析NestedType，如果NestedType中存在Rsp，则设置RspFullName
	for _, nested := range dp.NestedType {
		if nested.GetName() == SuffixRsp || nested.GetName() == SuffixResponse {
			msgInfo.RspFullName = ReqfullName + "_" + SuffixRsp
		}
	}
	return msgInfo
}

func genNotifyMessageNamesFromDescriptor(dp, father *descriptor.DescriptorProto, packageName string) MessageName {
	msgInfo := MessageName{
		Type: MessageTypeNotify,
	}
	NotifyfullName := father.GetName() + "_" + dp.GetName()
	msgInfo.NotifyFullName = NotifyfullName
	msgInfo.SetPackageName(packageName)
	msgInfo.Name = dp.GetName()
	msgInfo.MsgWall = father.GetName()
	msgInfo.DP = dp
	return msgInfo
}

func extractMessageNamesFromDescriptorByMsgWall(messages []*descriptor.DescriptorProto, msgWall string, packageName string) []MessageName {
	var names []MessageName
	for i := range messages {
		msg := messages[i]
		name := msg.GetName()
		if name != msgWall {
			continue
		}

		switch msgWall {
		case MsgWallReq:
			for i := range msg.NestedType {
				msgInfo := genRequestMessageNamesFromDescriptor(msg.NestedType[i], msg, packageName)
				names = append(names, msgInfo)
			}
		case MsgWallNotify:
			for i := range msg.NestedType {
				msgInfo := genNotifyMessageNamesFromDescriptor(msg.NestedType[i], msg, packageName)
				names = append(names, msgInfo)
			}
		default:
			panic(fmt.Sprintf("unknown msgWall: %s", msgWall))
		}
	}
	return names
}

// 解析Request消息
func parseRequestMessages(ctx *Context, fd *descriptor.FileDescriptorProto, packageName string) {
	// Extract nested messages from Request
	allMessageNames := extractMessageNamesFromDescriptorByMsgWall(fd.MessageType, MsgWallReq, packageName)
	for i := range allMessageNames {
		msgInfo := allMessageNames[i]
		req := Message{
			Type:        msgInfo.Type,
			Name:        msgInfo.Name,
			FullName:    msgInfo.ReqFullName,
			Response:    msgInfo.RspFullName,
			PackageName: packageName,
			PidName:     msgInfo.GetReqPidName(),
		}

		// Parse fields
		for i, f := range msgInfo.DP.Field {
			field := Field{
				Type:  f.GetTypeName(), // Or resolve properly
				Name:  f.GetName(),
				Index: i + 1,
			}
			req.Fields = append(req.Fields, field)
		}

		ctx.AddReqMessage(req)

		if msgInfo.RspFullName != "" {
			// 生成Rsp消息
			rsp := Message{
				Type:        MessageTypeRsp,
				Name:        msgInfo.Name,
				FullName:    msgInfo.RspFullName,
				PackageName: packageName,
				PidName:     msgInfo.GetRspPidName(),
			}
			ctx.AddRspMessage(rsp)
		}
	}
}

// 解析Notify消息
func parseNotifyMessages(ctx *Context, fd *descriptor.FileDescriptorProto, mode Mode, packageName string) {
	// Extract nested messages from Notify
	allMessageNames := extractMessageNamesFromDescriptorByMsgWall(fd.MessageType, MsgWallNotify, packageName)
	// 过滤出Notify_前缀的消息
	for _, msgInfo := range allMessageNames {
		if mode.IsDebug() {
			fmt.Printf("DEBUG: Processing notify message '%s'\n", msgInfo.NotifyFullName)
		}

		message := Message{
			Type:        msgInfo.Type,
			Name:        msgInfo.Name,
			FullName:    msgInfo.NotifyFullName,
			PackageName: packageName,
			PidName:     msgInfo.GetNotifyPidName(),
		}

		// Parse fields
		for i, f := range msgInfo.DP.Field {
			field := Field{
				Type:  f.GetTypeName(),
				Name:  f.GetName(),
				Index: i + 1,
			}
			message.Fields = append(message.Fields, field)
		}

		ctx.AddNotifyMessage(message)
	}
}

// 生成Request消息的胶水代码
func generateRequestGlueCode(ctx *Context) error {
	// 构建输出文件名
	outputFile := filepath.Join(ctx.GetOutputDir(), "request_glue.go")
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
	fmt.Fprintf(file, "\t\"fmt\"\n\n")
	fmt.Fprintf(file, "\t\"github.com/gogo/protobuf/proto\"\n")
	fmt.Fprintf(file, ")\n\n")

	messages := ctx.GetReqMessage()

	// 生成UnmarshalRequest函数
	fmt.Fprintf(file, "// UnmarshalRequest 根据协议ID分发请求到对应处理函数\n")
	fmt.Fprintf(file, "func UnmarshalRequest(pid uint32, data []byte) (proto.Message, error) {\n")
	fmt.Fprintf(file, "\tvar req proto.Message\n")
	fmt.Fprintf(file, "\tswitch pid {\n")

	for i := range messages {
		msg := messages[i]
		// 生成pid以供参考
		_ = protoid.HashProtoMessage(msg.FullName)

		fmt.Fprintf(file, "\tcase PID_%s: // %s\n", msg.FullName, msg.FullName)
		fmt.Fprintf(file, "\t\treq = &%s{}\n", msg.FullName)
		fmt.Fprintf(file, "\t\tif err := proto.Unmarshal(data, req); err != nil {\n")
		fmt.Fprintf(file, "\t\t\treturn nil, fmt.Errorf(\"unmarshal %s failed: %%w\", err)\n", msg.FullName)
		fmt.Fprintf(file, "\t\t}\n")
	}

	fmt.Fprintf(file, "\tdefault:\n")
	fmt.Fprintf(file, "\t\treturn nil, fmt.Errorf(\"unknown request protocol ID: 0x%%08x\", pid)\n")
	fmt.Fprintf(file, "\t}\n")
	fmt.Fprintf(file, "\treturn req, nil\n")
	fmt.Fprintf(file, "}\n\n")

	if ctx.GetMode().ShouldPrint() {
		fmt.Printf("Generated request glue code in %s\n", outputFile)
	}
	return nil
}

// 生成Notify消息的胶水代码
func generateNotifyGlueCode(ctx *Context) error {
	// 构建输出文件名
	outputFile := filepath.Join(ctx.GetOutputDir(), "notify_glue.go")
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

	fmt.Fprintf(file, ")\n\n")

	// 生成分发函数
	fmt.Fprintf(file, "// ParseNotifyByID 根据协议ID解析通知消息\n")
	fmt.Fprintf(file, "func ParseNotifyByID(pid uint32, data []byte) (proto.Message, uint32, error) {\n")
	fmt.Fprintf(file, "\tswitch pid {\n")
	messages := ctx.GetNotifyMessage()

	for _, msg := range messages {
		// 跳过不符合条件的消息
		if msg.Name == "Notify" {
			continue
		}

		// 生成pid以供参考
		_ = protoid.HashProtoMessage(msg.FullName)

		fmt.Fprintf(file, "\tcase PID_%s: // %s\n", msg.FullName, msg.FullName)
		fmt.Fprintf(file, "\t\tnotify := &%s{}\n", msg.FullName)
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
		fmt.Fprintf(file, "func Marshal%s(notify *%s) ([]byte, uint32, error) {\n", msg.Name, msg.FullName)
		fmt.Fprintf(file, "\tdata, err := proto.Marshal(notify)\n")
		fmt.Fprintf(file, "\treturn data, PID_%s, err\n", msg.FullName)
		fmt.Fprintf(file, "}\n\n")
	}

	if ctx.GetMode().ShouldPrint() {
		fmt.Printf("Generated notify glue code in %s\n", outputFile)
	}
	return nil
}

// generateCommonProtocolMappings 生成公共的协议ID映射文件
func generateCommonProtocolMappings(allMappings ProtocolIDMapping, outputDir string, mode Mode) error {
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
	for _, msgID := range allMappings.MessageIDs {
		fmt.Fprintf(file, "\tPID_%s uint32 = 0x%08x // %s\n",
			msgID.Name, msgID.ID, msgID.Name)
	}
	fmt.Fprintf(file, "\n")
	fmt.Fprintf(file, ")\n\n")

	// 生成全局的 MessageNameToID 映射
	fmt.Fprintf(file, "// AllMessageNameToID 全局消息名称到ID的映射\n")
	fmt.Fprintf(file, "var AllMessageNameToID = map[string]uint32{\n")
	for _, msgID := range allMappings.MessageIDs {
		fmt.Fprintf(file, "\t\"%s\": PID_%s,\n",
			msgID.Name, msgID.Name)
	}
	fmt.Fprintf(file, "}\n\n")

	// 生成全局的 IDToMessageName 映射
	fmt.Fprintf(file, "// AllIDToMessageName 全局ID到消息名称的映射\n")
	fmt.Fprintf(file, "var AllIDToMessageName = map[uint32]string{\n")
	for _, msgID := range allMappings.MessageIDs {
		fmt.Fprintf(file, "\tPID_%s: \"%s\",\n",
			msgID.Name, msgID.Name)
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
	fmt.Fprintf(file, "\tpid, ok := GetProtocolID(messageName)\n")
	fmt.Fprintf(file, "\tif !ok {\n")
	fmt.Fprintf(file, "\t\tpanic(fmt.Sprintf(\"消息 %s 未找到协议ID\", messageName))\n")
	fmt.Fprintf(file, "\t}\n")
	fmt.Fprintf(file, "\treturn pid\n")
	fmt.Fprintf(file, "}\n")

	if mode.ShouldPrint() {
		fmt.Printf("Generated common protocol ID mappings in %s\n", outputFile)
	}
	return nil
}

// generatePbGoFiles 使用 protoc 生成 pb.go 文件
func generatePbGoFiles(protoFiles []string, outputDir string, mode Mode) error {
	if len(protoFiles) == 0 {
		return fmt.Errorf("no proto files to process")
	}

	// 获取第一个proto文件的目录作为基础目录
	protoDir := filepath.Dir(protoFiles[0])

	// 确保输出目录存在
	if err := ensureOutputDir(outputDir); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", outputDir, err)
	}

	// 为每个proto文件生成pb.go文件到同一个目录
	for _, protoFile := range protoFiles {
		if mode.ShouldPrint() {
			fmt.Printf("Generating pb.go for %s...\n", protoFile)
		}

		// 构建protoc命令参数
		var cmdArgs []string
		cmdArgs = append(cmdArgs,
			"--proto_path=.",
			"--proto_path="+protoDir,
			"--proto_path=vendor/github.com/asynkron/protoactor-go/actor",
			"--proto_path=$GOPATH/pkg/mod",
			"--go_out="+outputDir,
			"--go_opt=paths=source_relative",
		)
		cmdArgs = append(cmdArgs, protoFile)

		// 执行protoc命令
		cmd := exec.Command("protoc", cmdArgs...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("protoc failed for %s: %w\nOutput: %s", protoFile, err, output)
		}

		if mode.ShouldPrint() {
			fmt.Printf("Generated pb.go file for %s in %s\n", filepath.Base(protoFile), outputDir)
		}
	}

	return nil
}
