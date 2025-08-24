package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	gogoproto "github.com/gogo/protobuf/proto"
	gogodesc "github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/spf13/cobra"
)

var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate extension methods and utilities for protocol buffers",
	Long: `Generate extension methods for DeltaSyncMap and DeltaSyncList containers.
This command analyzes proto files and creates type-safe helper methods.
Optionally generates serialization utilities for skip_serialization fields.
Can also generate plain Go structs mirroring proto messages when --go-structs is set.`,
	Run: func(cmd *cobra.Command, args []string) {
		protoDir, _ := cmd.Flags().GetString("proto-dir")
		outputDir, _ := cmd.Flags().GetString("output-dir")
		protobufInclude, _ := cmd.Flags().GetString("protobuf-include")
		includeSerialize, _ := cmd.Flags().GetBool("serialize")
		genGoStructs, _ := cmd.Flags().GetBool("go-structs")
		structsPkg, _ := cmd.Flags().GetString("structs-package")

		if err := generateAll(protoDir, outputDir, protobufInclude, includeSerialize, genGoStructs, structsPkg); err != nil {
			fmt.Printf("Error generating code: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Successfully generated all code!")
	},
}

// generateAll 生成所有代码（扩展方法和序列化工具）
func generateAll(protoDir, outputDir, protobufInclude string, includeSerialize bool, includeGoStructs bool, structsPackage string) error {
	// 生成扩展方法
	if err := generateExtensions(protoDir, outputDir, protobufInclude); err != nil {
		return fmt.Errorf("failed to generate extensions: %v", err)
	}
	fmt.Println("Container extensions generated successfully!")

	// 可选生成序列化工具
	if includeSerialize {
		if err := generateSerializeCode(protoDir, outputDir, protobufInclude); err != nil {
			return fmt.Errorf("failed to generate serialize code: %v", err)
		}
		fmt.Println("Serialization utilities generated successfully!")
	}

	// 可选生成 Go structs
	if includeGoStructs {
		if err := generateGoStructsFromProto(protoDir, outputDir, protobufInclude, structsPackage); err != nil {
			return fmt.Errorf("failed to generate Go structs: %v", err)
		}
		fmt.Println("Go structs generated successfully!")
	}

	return nil
}

// generateExtensions 生成扩展方法
func generateExtensions(protoDir, outputDir, protobufInclude string) error {

	ctx := NewContext(protoDir, outputDir, protobufInclude)
	// 解析 proto 文件
	if err := parseProtoFiles(ctx); err != nil {
		return fmt.Errorf("failed to parse proto files: %v", err)
	}

	// 分析 proto 文件
	err := analyzeProtoFiles(ctx)
	if err != nil {
		return fmt.Errorf("failed to analyze proto files: %v", err)
	}

	// 解析 MapKeyType 枚举
	err = parseMapKeyTypeEnum(ctx)
	if err != nil {
		return fmt.Errorf("failed to parse MapKeyType enum: %v", err)
	}

	defer ctx.Clear()

	// 生成扩展代码
	code := generateExtensionCode(ctx)

	// 写入文件
	outputFile := filepath.Join(outputDir, "container_ext.pb.go")
	return writeToFile(outputFile, code)
}

// helpers moved to utils.go

func keyEnumByMapType(goKey string) string {
	switch goKey {
	case "string":
		return "MapKeyType_KEY_STRING"
	case "int32":
		return "MapKeyType_KEY_INT32"
	case "int64":
		return "MapKeyType_KEY_INT64"
	case "uint32":
		return "MapKeyType_KEY_UINT32"
	case "uint64":
		return "MapKeyType_KEY_UINT64"
	default:
		return ""
	}
}

func elementTypeName(f *gogodesc.FieldDescriptorProto) string {
	switch f.GetType() {
	case gogodesc.FieldDescriptorProto_TYPE_MESSAGE:
		return shortTypeName(trimLeadingDot(f.GetTypeName()))
	case gogodesc.FieldDescriptorProto_TYPE_ENUM:
		return shortTypeName(trimLeadingDot(f.GetTypeName()))
	case gogodesc.FieldDescriptorProto_TYPE_STRING:
		return "string"
	case gogodesc.FieldDescriptorProto_TYPE_INT64, gogodesc.FieldDescriptorProto_TYPE_SINT64, gogodesc.FieldDescriptorProto_TYPE_SFIXED64:
		return "int64"
	case gogodesc.FieldDescriptorProto_TYPE_INT32, gogodesc.FieldDescriptorProto_TYPE_SINT32, gogodesc.FieldDescriptorProto_TYPE_SFIXED32:
		return "int32"
	case gogodesc.FieldDescriptorProto_TYPE_UINT64, gogodesc.FieldDescriptorProto_TYPE_FIXED64:
		return "uint64"
	case gogodesc.FieldDescriptorProto_TYPE_UINT32, gogodesc.FieldDescriptorProto_TYPE_FIXED32:
		return "uint32"
	case gogodesc.FieldDescriptorProto_TYPE_BOOL:
		return "bool"
	case gogodesc.FieldDescriptorProto_TYPE_BYTES:
		return "[]byte"
	case gogodesc.FieldDescriptorProto_TYPE_FLOAT:
		return "float32"
	case gogodesc.FieldDescriptorProto_TYPE_DOUBLE:
		return "float64"
	default:
		return f.GetType().String()
	}
}

func parseProtoFiles(ctx *Context) error {
	var protoFiles []string
	err := filepath.Walk(ctx.GetProtoDir(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".proto") {
			protoFiles = append(protoFiles, path)
		}
		return nil
	})
	if err != nil {
		return err
	}
	// 编译 proto 文件
	descriptorSet, err := compileProtoFiles(ctx.GetProtoDir(), protoFiles, ctx.GetProtobufInclude())
	if err != nil {
		return err
	}
	ctx.SetFileDescriptorSet(descriptorSet)
	return nil
}

// parseMapKeyTypeEnum 解析proto文件中的MapKeyType枚举
func parseMapKeyTypeEnum(ctx *Context) error {
	// 使用 gogo descriptor 扫描 FullMap 的 oneof hash_map 中的包装消息名，推断 MapKeyType → 包装类型 → Go基础键类型
	fds := ctx.GetFileDescriptorSet()
	if fds == nil {
		return fmt.Errorf("descriptor set is nil")
	}

	// 记录所有消息，便于快速查找
	for _, file := range fds.File {
		for _, msg := range file.GetMessageType() {
			if msg.GetName() == "FullMap" {
				// 扫描字段，查找属于 oneof hash_map 的字段
				for _, field := range msg.GetField() {
					// 判断是否属于名为 hash_map 的 oneof
					if field.OneofIndex != nil && int(*field.OneofIndex) < len(msg.GetOneofDecl()) {
						if msg.GetOneofDecl()[field.GetOneofIndex()].GetName() == "hash_map" {
							typeName := trimLeadingDot(field.GetTypeName()) // e.g. pb.StringKeyMap
							short := shortTypeName(typeName)                // e.g. StringKeyMap
							goKey := inferGoTypeFromMapType(short)
							if goKey != "" {
								enumVal := keyEnumByMapType(goKey) // e.g. MapKeyType_KEY_STRING
								if enumVal != "" {
									ctx.AddMapKeyType(MapKeyTypeInfo{EnumValue: enumVal, TypeName: short, MapType: goKey})
								}
							}
						}
					}
				}
			}
		}
	}
	return nil
}

// buildDynamicMapping 动态构建 MapKeyType 和 FullMap 字段的映射关系
// 保留占位，实际逻辑由 parseMapKeyTypeEnum(gogo) 完成
func buildDynamicMapping(_ *Context, _ interface{}, _ interface{}) {}

// parseDynamicKeyType 动态解析单个键类型
// 旧逻辑移除

// inferGoTypeFromMapType 从 Map 类型名称推断对应的 Go 类型
func inferGoTypeFromMapType(mapTypeName string) string {
	switch mapTypeName {
	case "StringKeyMap":
		return "string"
	case "Int64KeyMap":
		return "int64"
	case "Int32KeyMap":
		return "int32"
	case "Uint64KeyMap":
		return "uint64"
	case "Uint32KeyMap":
		return "uint32"
	default:
		// 尝试从名称中推断
		lower := strings.ToLower(mapTypeName)
		if strings.Contains(lower, "string") {
			return "string"
		} else if strings.Contains(lower, "int64") {
			return "int64"
		} else if strings.Contains(lower, "int32") {
			return "int32"
		} else if strings.Contains(lower, "uint64") {
			return "uint64"
		} else if strings.Contains(lower, "uint32") {
			return "uint32"
		} else if strings.Contains(lower, "bool") {
			return "bool"
		} else if strings.Contains(lower, "bytes") {
			return "[]byte"
		}
		return ""
	}
}

// analyzeProtoFiles 分析proto文件夹中的所有proto文件
func analyzeProtoFiles(ctx *Context) error {
	// 基于 gogo descriptor 的简单扫描：
	// - 找出包含 DeltaSyncMap 或 DeltaSyncList 字段的消息
	// - 将这些消息及其字段信息收集到 Context
	fds := ctx.GetFileDescriptorSet()
	if fds == nil {
		return fmt.Errorf("descriptor set is nil")
	}

	// 建立消息名索引，解析 map entry 需要
	// not needed in simplified analyzer

	for _, file := range fds.File {
		for _, msg := range file.GetMessageType() {
			pm := ProtoMessage{
				Name:   msg.GetName(),
				Fields: []ProtoField{},
			}

			hasMap := false
			hasList := false

			for _, f := range msg.GetField() {
				fieldInfo := ProtoField{
					Name: f.GetName(),
					Type: elementTypeName(f),
					Tags: map[string]string{},
				}
				pm.Fields = append(pm.Fields, fieldInfo)

				// 检测使用 DeltaSyncMap / DeltaSyncList
				if f.GetType() == gogodesc.FieldDescriptorProto_TYPE_MESSAGE {
					switch tname := shortTypeName(trimLeadingDot(f.GetTypeName())); tname {
					case "DeltaSyncMap":
						hasMap = true
						// 记录 MapFieldInfo
						pm.MapFields = append(pm.MapFields, MapFieldInfo{Name: f.GetName(), SyncTags: []string{"asset"}, AllTags: map[string]string{"sync": "asset"}})
					case "DeltaSyncList":
						hasList = true
					}
				}
			}

			pm.UsesMap = hasMap
			pm.UsesList = hasList
			if hasMap || hasList {
				ctx.AddProtoMessage(pm)
			}
		}
	}
	return nil
}

// compileProtoFiles 使用 protoc 编译 proto 文件
func compileProtoFiles(protoDir string, protoFiles []string, protobufInclude string) (*gogodesc.FileDescriptorSet, error) {
	// 创建临时文件存储描述符
	tmpFile, err := os.CreateTemp("", "proto_descriptors_*.pb")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// 构建 protoc 命令
	args := []string{
		"--descriptor_set_out=" + tmpFile.Name(),
		"--include_source_info",
		"--proto_path=" + protoDir,
	}

	// 添加 Google protobuf 标准库路径
	var protobufIncludePath string
	if protobufInclude != "" {
		// 使用传入的路径
		protobufIncludePath = protobufInclude
		fmt.Printf("Using specified protobuf include path: %s\n", protobufIncludePath)
	} else {
		// 动态检测protobuf include路径
		protobufIncludePath = findDescriptorProto()
		if protobufIncludePath == "" {
			panic("protobuf well-known types not found")
		}
	}

	args = append(args, "--proto_path="+protobufIncludePath)
	args = append(args, protoFiles...)
	descriptorProtoPath := filepath.Join(protobufIncludePath, "google/protobuf/descriptor.proto")
	args = append(args, descriptorProtoPath)

	cmd := exec.Command("protoc", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("protoc failed: %v\nOutput: %s", err, string(output))
	}

	// 读取描述符
	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return nil, err
	}

	var descriptorSet gogodesc.FileDescriptorSet
	if err := gogoproto.Unmarshal(data, &descriptorSet); err != nil {
		return nil, err
	}

	return &descriptorSet, nil
}

// extractMessagesFromDescriptor 从文件描述符中提取消息信息
func extractMessagesFromDescriptor(_ interface{}) []ProtoMessage {
	return nil
}

// extractNestedMessages 递归提取嵌套消息
func extractNestedMessages(_ interface{}) []ProtoMessage { return nil }

// extractMessageInfo 从消息描述符中提取信息
func extractMessageInfo(_ interface{}) *ProtoMessage {
	return nil
}

// extractFieldInfo 从字段描述符中提取字段信息
func extractFieldInfo(_ interface{}) ProtoField {
	return ProtoField{}
}

// extractMapFieldInfo 提取 DeltaSyncMap 字段的详细信息
func extractMapFieldInfo(_ interface{}, _ map[string]string) MapFieldInfo {
	return MapFieldInfo{}
}

// getFieldTypeName 获取字段类型名称
func getFieldTypeName(_ interface{}) string {
	return ""
}

// extractTagsFromComments 从注释中提取标签信息
func extractTagsFromComments(_ interface{}) map[string]string {
	return map[string]string{}
}

// generateExtensionCode 生成扩展代码
func generateExtensionCode(ctx *Context) string {
	var builder strings.Builder
	messages := ctx.GetProtoMessages()
	mapKeyTypes := ctx.GetMapKeyTypes()

	// 文件头
	builder.WriteString(`package pb

import (
    "github.com/gogo/protobuf/proto"
)

// SyncMode 定义同步模式
type SyncMode int

const (
	SyncModeUnknown SyncMode = iota
	SyncModeFullMap          // 全量模式
	SyncModeChanges          // 增量模式
)

// deepCopyGenericData 深拷贝 GenericData 对象
func deepCopyGenericData(original *GenericData) *GenericData {
	if original == nil {
		return nil
	}
	
	// 使用 protobuf 的 Clone 方法进行深拷贝
	cloned := proto.Clone(original).(*GenericData)
	return cloned
}

// FilterForSerialization 过滤消息，移除标记为skip_serialization的字段（gogo版，无反射）
func FilterForSerialization(msg proto.Message) proto.Message {
    cloned := proto.Clone(msg)
    switch m := cloned.(type) {
    case *DeltaSyncMap:
        // container.proto 中标记了 storage = 3 [(skip_serialization) = true]
        m.Storage = nil
        return m
    default:
        return cloned
    }
}
`)

	builder.WriteString(generateAreGenericKeysEqualMethod(mapKeyTypes))

	// 生成DeltaSyncMap的基础方法
	builder.WriteString(generateDeltaSyncMapMethods(messages, mapKeyTypes))

	// 为使用DeltaSyncMap的消息生成扩展方法
	for _, msg := range messages {
		if msg.UsesMap {
			builder.WriteString(generateComponentExtensions(msg))
		}
	}

	return builder.String()
}

// generateAreGenericKeysEqualMethod 动态生成键比较方法
func generateAreGenericKeysEqualMethod(mapKeyTypes []MapKeyTypeInfo) string {
	var result strings.Builder

	result.WriteString("// areGenericKeysEqual 比较两个 GenericKey 是否相等\n")
	result.WriteString("// 根据键类型动态生成，避免类型断言，性能更高\n")
	result.WriteString("func areGenericKeysEqual(et MapKeyType, key1, key2 *GenericKey) bool {\n")
	result.WriteString("\tswitch et {\n")

	for _, keyType := range mapKeyTypes {
		// 根据 MapType 获取对应的 GenericKey getter 方法名
		getterMethod := fmt.Sprintf("Get%s", getGenericKeyField(keyType.MapType))
		result.WriteString(fmt.Sprintf("\tcase %s:\n", keyType.EnumValue))
		result.WriteString(fmt.Sprintf("\t\treturn key1.%s() == key2.%s()\n", getterMethod, getterMethod))
	}

	result.WriteString("\t}\n")
	result.WriteString("\treturn proto.Equal(key1, key2)\n")
	result.WriteString("}\n\n")

	return result.String()
}

// generateInitializeMethod 动态生成Initialize方法
func generateInitializeMethod() string {
	return `// Initialize initializes the DeltaSyncMap with a key type and sets up storage.
func (m *DeltaSyncMap) Initialize(kt MapKeyType) {
	m.MapKeyType = kt
	m.Version = 1
	m.initializeStorage()
}

`
}

// generateMapKeyTypeCases 动态生成MapKeyType的case语句
func generateMapKeyTypeCases(mapKeyTypes []MapKeyTypeInfo) string {
	var cases strings.Builder
	for _, keyType := range mapKeyTypes {
		cases.WriteString(fmt.Sprintf(`
	case %s:
		storage.HashMap = &FullMap_%s{%s: &%s{Map: make(map[%s]*GenericData)}}`,
			keyType.EnumValue, keyType.TypeName, keyType.TypeName, keyType.TypeName, keyType.MapType))
	}

	return cases.String()
}

// generateDeltaSyncMapMethods 生成DeltaSyncMap的基础方法
func generateDeltaSyncMapMethods(messages []ProtoMessage, mapKeyTypes []MapKeyTypeInfo) string {
	var builder strings.Builder

	// 收集所有使用的 tag 类型
	tagTypes := make(map[string]bool)
	for _, msg := range messages {
		for _, field := range msg.MapFields {
			for _, tag := range field.SyncTags {
				if tag != "" {
					tagTypes[tag] = true
				}
			}
		}
	}

	// 如果没有 tag，使用默认的 asset
	if len(tagTypes) == 0 {
		tagTypes["asset"] = true
	}

	// 动态生成Initialize方法
	initializeMethod := generateInitializeMethod()
	builder.WriteString(initializeMethod)

	builder.WriteString(`
// GetSyncMode 获取当前同步模式
func (m *DeltaSyncMap) GetSyncMode() SyncMode {
	if m.SyncData == nil {
		return SyncModeUnknown
	}
	switch m.SyncData.(type) {
	case *DeltaSyncMap_FullMap:
		return SyncModeFullMap
	case *DeltaSyncMap_Changes:
		return SyncModeChanges
	default:
		return SyncModeUnknown
	}
}

// initializeStorage 初始化服务器端数据存储
func (m *DeltaSyncMap) initializeStorage() {
	storage := &FullMap{}
	switch m.MapKeyType {` + generateMapKeyTypeCases(mapKeyTypes) + `
	default:
		panic("unsupported key type: " + m.MapKeyType.String())
	}
	
	m.Storage = storage
}

// DeepCopyToSyncData 将服务器数据深度复制到同步数据
func (m *DeltaSyncMap) DeepCopyToSyncData() *DeltaSyncMap {
	if m.Storage == nil {
		m.initializeStorage()
	}

	transport := &DeltaSyncMap{
		MapKeyType: m.MapKeyType,
		Version: m.Version,
	}
	
	// 深度复制 storage 到 sync_data
	syncData := &FullMap{}
	switch m.MapKeyType {` + generateDeepCopyDataCases(mapKeyTypes) + `
	default:
		panic("unsupported key type: " + m.MapKeyType.String())
	}
	
	transport.SyncData = &DeltaSyncMap_FullMap{FullMap: syncData}
	return transport
}

// DeepCopyToSyncDataFiltered 将服务器数据深度复制到同步数据，并过滤不需要序列化的字段
func (m *DeltaSyncMap) DeepCopyToSyncDataFiltered() *DeltaSyncMap {
	// 先获取深拷贝的数据
	transport := m.DeepCopyToSyncData()
	
	// 过滤掉标记为 skip_serialization 的字段
	filtered := FilterForSerialization(transport).(*DeltaSyncMap)
	return filtered
}

// CopyToSyncData 将服务器数据浅复制到同步数据
func (m *DeltaSyncMap) CopyToSyncData() *DeltaSyncMap {
	if m.Storage == nil {
		m.initializeStorage()
	}

	transport := &DeltaSyncMap{
		MapKeyType: m.MapKeyType,
		Version: m.Version,
	}
	
	syncData := &FullMap{}
	switch m.MapKeyType {` + generateCopyDataCases(mapKeyTypes) + `
	default:
		panic("unsupported key type: " + m.MapKeyType.String())
	}
	
	transport.SyncData = &DeltaSyncMap_FullMap{FullMap: syncData}
	return transport
}

// CopyToSyncDataFiltered 将服务器数据浅复制到同步数据，并过滤不需要序列化的字段
func (m *DeltaSyncMap) CopyToSyncDataFiltered() *DeltaSyncMap {
	// 先获取浅拷贝的数据
	transport := m.CopyToSyncData()
	
	// 过滤掉标记为 skip_serialization 的字段
	filtered := FilterForSerialization(transport).(*DeltaSyncMap)
	return filtered
}

// Copy 浅拷贝HashMap元素
func (m *DeltaSyncMap) Copy() *FullMap {
	if m.Storage == nil {
		m.initializeStorage()
	}
	
	syncData := &FullMap{}
	switch m.MapKeyType {` + generateCopyDataCases(mapKeyTypes) + `
	default:
		panic("unsupported key type: " + m.MapKeyType.String())
	}
	
	return syncData
}

// DeepCopy 深拷贝HashMap元素
func (m *DeltaSyncMap) DeepCopy() *FullMap {
	if m.Storage == nil {
		m.initializeStorage()
	}
	
	syncData := &FullMap{}
	switch m.MapKeyType {` + generateDeepCopyDataCases(mapKeyTypes) + `
	default:
		panic("unsupported key type: " + m.MapKeyType.String())
	}
	
	return syncData
}

// addChange 添加变更记录（仅在增量模式下）
// 实现变更压缩：对同一个key的多次变更，只保留最新的一次
// 智能优化：ADD+DELETE=无操作，DELETE+ADD/UPDATE=UPDATE
func (m *DeltaSyncMap) addChange(changeType ChangeType, key *GenericKey, data *GenericData) {
	if m.GetSyncMode() != SyncModeChanges {
		return
	}
	
	changes := m.GetChanges()
	if changes == nil {
		return
	}
	
	// 查找是否已存在相同key的变更记录
	existingIndex := -1
	var existingChange *GenericMapChange
	for i, change := range changes.ChangeList {
		if areGenericKeysEqual(m.MapKeyType, change.Key, key) {
			existingIndex = i
			existingChange = change
			break
		}
	}
	
	if existingIndex >= 0 {
		// 存在相同key的变更，需要智能合并
		existingType := existingChange.Type
		
		// 优化规则：
		// ADD + DELETE = 无操作（删除该变更记录）
		// UPDATE + DELETE = DELETE
		// DELETE + ADD = UPDATE  
		// DELETE + UPDATE = UPDATE
		// ADD + UPDATE = ADD (使用新data)
		// UPDATE + UPDATE = UPDATE (使用新data)
		
		if existingType == ChangeType_ADD && changeType == ChangeType_DELETE {
			// ADD + DELETE = 无操作，删除变更记录
			changes.ChangeList = append(changes.ChangeList[:existingIndex], changes.ChangeList[existingIndex+1:]...)
			return
		} else if existingType == ChangeType_DELETE && (changeType == ChangeType_ADD || changeType == ChangeType_UPDATE) {
			// DELETE + ADD/UPDATE = UPDATE
			changes.ChangeList[existingIndex] = &GenericMapChange{
				Type: ChangeType_UPDATE,
				Key:  key,
				Data: data,
			}
			return
		} else if existingType == ChangeType_ADD && changeType == ChangeType_UPDATE {
			// ADD + UPDATE = ADD (使用新数据)
			changes.ChangeList[existingIndex] = &GenericMapChange{
				Type: ChangeType_ADD,
				Key:  key,
				Data: data,
			}
			return
		} else {
			// 其他情况直接替换
			changes.ChangeList[existingIndex] = &GenericMapChange{
				Type: changeType,
				Key:  key,
				Data: data,
			}
			return
		}
	}
	
	// 没有找到相同key的变更，添加新的变更记录
	newChange := &GenericMapChange{
		Type: changeType,
		Key:  key,
		Data: data,
	}
	changes.ChangeList = append(changes.ChangeList, newChange)
}

`)

	// 为每种 tag 类型生成对应的方法
	for tagType := range tagTypes {
		capitalizedTag := strings.Title(tagType)
		builder.WriteString(fmt.Sprintf(`// Put%s 向 map 中添加一个 %s (仅支持 string 键)
func (m *DeltaSyncMap) Put%s(key string, %s *%s) {
	data := &GenericData{
		Data: &GenericData_%s{%s: %s},
	}
	m.PutString(key, data)
}

// Get%s 从 map 中获取 %s (仅支持 string 键)
func (m *DeltaSyncMap) Get%s(key string) (*%s, bool) {
	if data, ok := m.GetString(key); ok {
		if %s := data.Get%s(); %s != nil {
			return %s, true
		}
	}
	return nil, false
}

// Delete%s 从 map 中删除 %s (仅支持 string 键)
func (m *DeltaSyncMap) Delete%s(key string) bool {
	return m.DeleteString(key)
}

`,
			// Put method
			capitalizedTag, tagType, capitalizedTag, tagType, capitalizedTag,
			capitalizedTag, capitalizedTag, tagType,

			// Get method
			capitalizedTag, tagType, capitalizedTag, capitalizedTag,
			tagType, capitalizedTag, tagType, tagType,

			// Delete method
			capitalizedTag, tagType, capitalizedTag,
		))
	}

	// 为每种键类型生成对应的基础 Put/Get/Delete 方法
	for _, keyType := range mapKeyTypes {
		keyTypeName := getKeyTypeMethodName(keyType.MapType)
		builder.WriteString(generateKeyTypeMethods(keyType, keyTypeName))
	}

	builder.WriteString(`
// HasChanges 检查是否有待同步的变更
func (m *DeltaSyncMap) HasChanges() bool {
	if m.GetSyncMode() != SyncModeChanges {
		return false
	}
	changes := m.GetChanges()
	return changes != nil && len(changes.ChangeList) > 0
}

// ClearChanges 清空变更记录（仅在增量模式下有效）
func (m *DeltaSyncMap) ClearChanges() {
	if m.GetSyncMode() == SyncModeChanges {
		changes := m.GetChanges()
		if changes != nil {
			changes.ChangeList = changes.ChangeList[:0]
		}
	}
}

// GetStorageSize 获取服务器端数据的大小
func (m *DeltaSyncMap) GetStorageSize() int {
	if m.Storage == nil {
		return 0
	}
	
	switch m.MapKeyType {` + generateGetSizeCases(mapKeyTypes) + `
	default:
		return 0
	}
}

`)

	return builder.String()
}

// generateCopyDataCases 生成复制数据的 case 语句
func generateCopyDataCases(mapKeyTypes []MapKeyTypeInfo) string {
	var cases strings.Builder
	for _, keyType := range mapKeyTypes {
		cases.WriteString(fmt.Sprintf(`
	case %s:
		serverMap := m.Storage.%s()
		if serverMap != nil {
			newMap := make(map[%s]*GenericData)
			for k, v := range serverMap.Map {
				newMap[k] = v
			}
			syncData.HashMap = &FullMap_%s{%s: &%s{Map: newMap}}
		}`,
			keyType.EnumValue, getFullMapGetter(keyType.TypeName), keyType.MapType,
			keyType.TypeName, keyType.TypeName, keyType.TypeName))
	}
	return cases.String()
}

func generateDeepCopyDataCases(mapKeyTypes []MapKeyTypeInfo) string {
	var cases strings.Builder
	for _, keyType := range mapKeyTypes {
		cases.WriteString(fmt.Sprintf(`
	case %s:
		serverMap := m.Storage.%s()
		if serverMap != nil {
			newMap := make(map[%s]*GenericData)
			for k, v := range serverMap.Map {
				if v != nil {
					// 深拷贝 GenericData
					newMap[k] = deepCopyGenericData(v)
				}
			}
			syncData.HashMap = &FullMap_%s{%s: &%s{Map: newMap}}
		}`,
			keyType.EnumValue, getFullMapGetter(keyType.TypeName), keyType.MapType,
			keyType.TypeName, keyType.TypeName, keyType.TypeName))
	}
	return cases.String()
}

// generateGetSizeCases 生成获取大小的 case 语句
func generateGetSizeCases(mapKeyTypes []MapKeyTypeInfo) string {
	var cases strings.Builder
	for _, keyType := range mapKeyTypes {
		cases.WriteString(fmt.Sprintf(`
	case %s:
		keyMap := m.Storage.%s()
		if keyMap != nil {
			return len(keyMap.Map)
		}
		return 0`,
			keyType.EnumValue, getFullMapGetter(keyType.TypeName)))
	}
	return cases.String()
}

// generateComponentExtensions 为组件生成扩展方法
func generateComponentExtensions(msg ProtoMessage) string {
	var builder strings.Builder

	// 生成组件的扩展方法
	builder.WriteString(fmt.Sprintf(`// ===== %s 扩展方法 =====

`, msg.Name))

	// 跟踪已生成的方法，避免重复
	generatedMethods := make(map[string]bool)

	// 为每个DeltaSyncMap字段生成方法
	for _, fieldInfo := range msg.MapFields {
		fieldName := fieldInfo.Name
		syncTags := fieldInfo.SyncTags

		// 如果没有 sync tags，使用默认的 asset
		if len(syncTags) == 0 {
			syncTags = []string{"asset"}
		}

		methodPrefix := strings.Title(strings.TrimSuffix(fieldName, "s")) // items -> Item

		// 生成Initialize方法（每个字段只生成一次）
		initMethodKey := fmt.Sprintf("Initialize%s", methodPrefix)
		if !generatedMethods[initMethodKey] {
			builder.WriteString(fmt.Sprintf(`// Initialize%s 初始化 %s 的 %s 字段
func (c *%s) Initialize%s() {
	if c.%s == nil {
		c.%s = &DeltaSyncMap{}
	}
	c.%s.Initialize(MapKeyType_KEY_STRING)
}

`,
				methodPrefix, msg.Name, fieldName,
				msg.Name, methodPrefix,
				strings.Title(fieldName), strings.Title(fieldName),
				strings.Title(fieldName)))
			generatedMethods[initMethodKey] = true
		}

		// 为每个sync tag生成对应的数据操作方法
		for _, tag := range syncTags {
			capitalizedTag := strings.Title(tag)

			// 检查是否已经生成过这个类型的方法
			addMethodKey := fmt.Sprintf("Add%s", capitalizedTag)
			if !generatedMethods[addMethodKey] {
				builder.WriteString(fmt.Sprintf(`// Add%s 添加%s到%s
func (c *%s) Add%s(itemId string, %s *%s) {
	if c.%s == nil {
		c.Initialize%s()
	}
	c.%s.Put%s(itemId, %s)
	c.DirtyFlag = 1 // 标记为脏数据
}

// Get%s 获取指定ID的%s
func (c *%s) Get%s(itemId string) (*%s, bool) {
	if c.%s == nil {
		return nil, false
	}
	return c.%s.Get%s(itemId)
}

// Remove%s 移除指定ID的%s
func (c *%s) Remove%s(itemId string) bool {
	if c.%s == nil {
		return false
	}
	success := c.%s.Delete%s(itemId)
	if success {
		c.DirtyFlag = 1 // 标记为脏数据
	}
	return success
}

`,
					// Add method
					capitalizedTag, strings.ToLower(tag), fieldName,
					msg.Name, capitalizedTag, strings.ToLower(tag), capitalizedTag,
					strings.Title(fieldName), methodPrefix,
					strings.Title(fieldName), capitalizedTag, strings.ToLower(tag),

					// Get method
					capitalizedTag, strings.ToLower(tag),
					msg.Name, capitalizedTag, capitalizedTag,
					strings.Title(fieldName),
					strings.Title(fieldName), capitalizedTag,

					// Remove method
					capitalizedTag, strings.ToLower(tag),
					msg.Name, capitalizedTag,
					strings.Title(fieldName),
					strings.Title(fieldName), capitalizedTag))

				generatedMethods[addMethodKey] = true
				generatedMethods[fmt.Sprintf("Get%s", capitalizedTag)] = true
				generatedMethods[fmt.Sprintf("Remove%s", capitalizedTag)] = true
			}
		}

		// 生成同步相关方法（每个字段只生成一次）
		syncMethodKey := fmt.Sprintf("Start%sSync", methodPrefix)
		if !generatedMethods[syncMethodKey] {
			builder.WriteString(fmt.Sprintf(`
// Get%sChanges 获取%s变更数据
func (c *%s) Get%sChanges() *DeltaSyncMap {
	if c.%s == nil {
		return nil
	}
	return c.%s
}

// Clear%sChanges 清除%s变更记录
func (c *%s) Clear%sChanges() {
	if c.%s != nil {
		c.%s.ClearChanges()
	}
}

`,
				methodPrefix, strings.ToLower(fieldName),
				msg.Name, methodPrefix,
				strings.Title(fieldName),
				strings.Title(fieldName),

				methodPrefix, strings.ToLower(fieldName),
				msg.Name, methodPrefix,
				strings.Title(fieldName), strings.Title(fieldName)))

			generatedMethods[syncMethodKey] = true
			generatedMethods[fmt.Sprintf("Get%sChanges", methodPrefix)] = true
			generatedMethods[fmt.Sprintf("Clear%sChanges", methodPrefix)] = true
		}
	}

	return builder.String()
}

// getKeyTypeMethodName 根据键类型获取方法名
func getKeyTypeMethodName(keyType string) string {
	switch keyType {
	case "string":
		return "String"
	case "int32":
		return "Int32"
	case "int64":
		return "Int64"
	case "uint32":
		return "Uint32"
	case "uint64":
		return "Uint64"
	default:
		return strings.Title(keyType)
	}
}

// generateKeyTypeMethods 为指定键类型生成 Put/Get/Delete 方法
func generateKeyTypeMethods(keyType MapKeyTypeInfo, methodSuffix string) string {
	var builder strings.Builder

	// 获取对应的 GenericKey 字段名和 FullMap 方法名
	genericKeyField := getGenericKeyField(keyType.MapType)
	fullMapGetter := getFullMapGetter(keyType.TypeName)

	builder.WriteString(fmt.Sprintf(`// Get%s retrieves the value for a %s key from server data.
func (m *DeltaSyncMap) Get%s(key %s) (*GenericData, bool) {
	if m.MapKeyType != %s {
		return nil, false
	}
	
	if m.Storage == nil {
		return nil, false
	}
	
	keyMap := m.Storage.%s()
	if keyMap == nil {
		return nil, false
	}
	v, ok := keyMap.Map[key]
	return v, ok
}

// Put%s adds or updates the value for a %s key in server data, increments version and records change.
func (m *DeltaSyncMap) Put%s(key %s, value *GenericData) {
	if m.MapKeyType != %s {
		panic("key type mismatch")
	}
	
	if m.Storage == nil {
		m.initializeStorage()
	}
	
	// 更新服务器端数据（深拷贝输入数据以避免外部修改）
	deepCopiedValue := deepCopyGenericData(value)
	keyMap := m.Storage.%s()
	if keyMap == nil {
		panic("server data not initialized properly")
	}
	keyMap.Map[key] = deepCopiedValue
	
	// 记录变更（在增量模式下，重用深拷贝的值）
	genericKey := &GenericKey{Key: &GenericKey_%s{%s: key}}
	m.addChange(ChangeType_UPDATE, genericKey, deepCopiedValue)
	
	m.Version++
}

// Delete%s removes a %s key from server data, increments version and records change.
func (m *DeltaSyncMap) Delete%s(key %s) bool {
	if m.MapKeyType != %s {
		return false
	}
	
	if m.Storage == nil {
		return false
	}
	
	// 从服务器端数据中删除
	keyMap := m.Storage.%s()
	if keyMap == nil {
		return false
	}
	
	_, existed := keyMap.Map[key]
	if existed {
		delete(keyMap.Map, key)
		
		// 记录变更（在增量模式下）
		genericKey := &GenericKey{Key: &GenericKey_%s{%s: key}}
		m.addChange(ChangeType_DELETE, genericKey, nil)
		
		m.Version++
	}
	
	return existed
}

`,
		// Get method
		methodSuffix, keyType.MapType, methodSuffix, keyType.MapType, keyType.EnumValue, fullMapGetter,

		// Put method
		methodSuffix, keyType.MapType, methodSuffix, keyType.MapType, keyType.EnumValue,
		fullMapGetter, genericKeyField, genericKeyField,

		// Delete method
		methodSuffix, keyType.MapType, methodSuffix, keyType.MapType, keyType.EnumValue,
		fullMapGetter, genericKeyField, genericKeyField,
	))

	return builder.String()
}

// generateSerializeCode 生成序列化代码
func generateSerializeCode(protoDir, outputDir, protobufInclude string) error {
	ctx := NewContext(protoDir, outputDir, protobufInclude)

	// 解析 proto 文件
	if err := parseProtoFiles(ctx); err != nil {
		return fmt.Errorf("failed to parse proto files: %v", err)
	}

	// 分析扩展字段
	extensionNumber, err := findSkipSerializationExtension(ctx)
	if err != nil {
		return fmt.Errorf("failed to find skip_serialization extension: %v", err)
	}

	defer ctx.Clear()

	// 生成序列化代码
	code := generateSerializationUtilities(extensionNumber)

	// 写入文件
	outputFile := filepath.Join(outputDir, "serialize.pb.go")
	return writeToFile(outputFile, code)
}

// findSkipSerializationExtension 查找 skip_serialization 扩展定义
func findSkipSerializationExtension(ctx *Context) (int32, error) {
	// 解析描述符
	var extensionNumber int32 = 50000
	fds := ctx.GetFileDescriptorSet()
	for _, file := range fds.File {
		for _, ext := range file.GetExtension() {
			if ext.GetName() == "skip_serialization" {
				extensionNumber = ext.GetNumber()
				return extensionNumber, nil
			}
		}
	}
	return extensionNumber, nil
}

// generateSerializationUtilities 生成序列化工具代码
func generateSerializationUtilities(extensionNumber int32) string {
	var builder strings.Builder

	// 文件头
	builder.WriteString(`package pb

import (
    "github.com/gogo/protobuf/proto"
)

`)

	// 注释：E_SkipSerialization 扩展已在 container.pb.go 中定义
	builder.WriteString("// E_SkipSerialization 扩展定义位于 container.pb.go 中\n\n")

	// 生成序列化工具函数
	builder.WriteString(`// 自定义序列化函数：过滤带有 skip_serialize 标记的字段
func MarshalWithSkip(msg proto.Message) ([]byte, error) {
	// 克隆消息，避免修改原对象
	cloned := proto.Clone(msg)
	// 递归清除带标记的字段
	ClearSkipFields(cloned.ProtoReflect())
	// 调用原生 Marshal 序列化处理后的消息
    return proto.Marshal(cloned)
}

// gogo 简化版：无需反射清理，直接复用 FilterForSerialization
// IsFieldSkipped/GetSkippedFields 在 gogo 环境下不提供基于反射的实现

// CompareSerializationSize 比较普通序列化和跳过序列化的大小差异
func CompareSerializationSize(msg proto.Message) (normalSize, skipSize int, err error) {
	// 普通序列化
	normalData, err := proto.Marshal(msg)
	if err != nil {
		return 0, 0, err
	}

	// 跳过序列化
	skipData, err := MarshalWithSkip(msg)
	if err != nil {
		return 0, 0, err
	}

	return len(normalData), len(skipData), nil
}
`)

	return builder.String()
}
