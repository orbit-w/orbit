package gostruct

import (
	"fmt"
	"go/format"
	"regexp"
	"sort"
	"strings"

	gogodesc "github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
)

// ProtoContext is the minimal context contract required by this package.
// It is intentionally small so callers can adapt existing contexts easily.
type ProtoContext interface {
	GetFileDescriptorSet() *gogodesc.FileDescriptorSet
	AddMessageToIndex(name string, msg *gogodesc.DescriptorProto)
	GetMessageFromIndex(name string) (*gogodesc.DescriptorProto, bool)
}

// GenerationOptions controls optional codegen features.
type GenerationOptions struct {
	// EmitEntityOpsForwarders controls whether to generate Entity-level forwarders
	// for component map Ops helpers. When false, users should prefer:
	//   e.GetXxxComponent().YyyOps()
	// over auto-forwarded e.XxxComponentYyyOps().
	EmitEntityOpsForwarders bool
}

var genOptions = GenerationOptions{
	EmitEntityOpsForwarders: false,
}

// SetGenerationOptions allows callers (CLI) to customize generator behavior.
func SetGenerationOptions(opts GenerationOptions) { genOptions = opts }

// GoStruct 表示从proto message生成的Go结构体信息
type GoStruct struct {
	MessageDescriptor gogodesc.DescriptorProto
	Name              string
	ProtoName         string
	PackageName       string
	Fields            []*GoField
	IsNested          bool
	ParentName        string
	Comment           string
	FullName          string
	SourceFile        string
}

// FieldKind 表示字段的种类
type FieldKind int

// GenerationConfig 生成配置
type GenerationConfig struct {
	PackageName     string
	IncludeComments bool
	CustomTags      map[string]string
	FieldFilter     func(*GoField) bool
	TypeMapping     map[string]string
}

// DefaultGenerationConfig 返回默认生成配置
func DefaultGenerationConfig(packageName string) *GenerationConfig {
	return &GenerationConfig{
		PackageName:     packageName,
		IncludeComments: true,
		CustomTags:      make(map[string]string),
		FieldFilter:     nil,
		TypeMapping:     make(map[string]string),
	}
}

// ParseProtoToGoStructs 解析proto文件，生成Go结构体信息
func ParseProtoToGoStructs(ctx ProtoContext, packageName string) ([]*GoStruct, error) {
	config := DefaultGenerationConfig(packageName)
	return ParseProtoToGoStructsWithConfig(ctx, config)
}

// ParseProtoToGoStructsWithConfig 使用配置解析proto文件，生成Go结构体信息
func ParseProtoToGoStructsWithConfig(ctx ProtoContext, config *GenerationConfig) ([]*GoStruct, error) {
	fds := ctx.GetFileDescriptorSet()
	if fds == nil {
		return nil, fmt.Errorf("descriptor set is nil")
	}

	var structs []*GoStruct
	buildMessageIndex(ctx)

	for _, file := range fds.File {
		if isSystemProtoFile(file.GetName()) {
			continue
		}
		for _, msg := range file.GetMessageType() {
			if isMapEntryMessage(msg) {
				continue
			}
			sourceName := file.GetName()
			goStruct := ConvertMessageToGoStructWithConfig(ctx, msg, config, false, "")
			if goStruct != nil {
				goStruct.SourceFile = sourceName
				structs = append(structs, goStruct)
			}
			nestedStructs := extractNestedGoStructsWithConfig(ctx, msg, config, msg.GetName())
			for _, ns := range nestedStructs {
				ns.SourceFile = sourceName
			}
			structs = append(structs, nestedStructs...)
		}
	}
	return structs, nil
}

// IsSystemProtoFile exported wrapper.
func IsSystemProtoFile(filename string) bool { return isSystemProtoFile(filename) }

// isMapEntryMessage 判断给定消息是否为 protobuf 的 map entry 消息。
func isMapEntryMessage(msgDesc *gogodesc.DescriptorProto) bool {
	if msgDesc == nil {
		return false
	}
	if options := msgDesc.GetOptions(); options != nil && options.GetMapEntry() {
		return true
	}
	return false
}

// IsMapEntryMessage exported wrapper.
func IsMapEntryMessage(msgDesc *gogodesc.DescriptorProto) bool { return isMapEntryMessage(msgDesc) }

// extractNestedGoStructsWithConfig 使用配置递归提取嵌套的Go结构体
func extractNestedGoStructsWithConfig(ctx ProtoContext, msgDesc *gogodesc.DescriptorProto, config *GenerationConfig, parentName string) []*GoStruct {
	var structs []*GoStruct
	nested := msgDesc.GetNestedType()
	for _, nestedMsg := range nested {
		if isMapEntryMessage(nestedMsg) {
			continue
		}
		goStruct := ConvertMessageToGoStructWithConfig(ctx, nestedMsg, config, true, parentName)
		if goStruct != nil {
			structs = append(structs, goStruct)
		}
		nestedStructs := extractNestedGoStructsWithConfig(ctx, nestedMsg, config, parentName+"_"+nestedMsg.GetName())
		structs = append(structs, nestedStructs...)
	}
	return structs
}

// getProtoFieldTypeWithIndex 获取proto字段类型字符串（感知 map entry，基于上下文索引准确还原）
func getProtoFieldTypeWithIndex(ctx ProtoContext, fieldDesc *gogodesc.FieldDescriptorProto) string {
	if isMapField(fieldDesc) {
		keyField, valueField := getMapKeyValueTypes(ctx, fieldDesc)
		if keyField != nil && valueField != nil {
			keyType := getElementType(keyField)
			valueType := getElementType(valueField)
			return fmt.Sprintf("map<%s, %s>", keyType, valueType)
		}
	}
	if fieldDesc.GetLabel() == gogodesc.FieldDescriptorProto_LABEL_REPEATED {
		elementType := getElementType(fieldDesc)
		return fmt.Sprintf("repeated %s", elementType)
	}
	return getElementType(fieldDesc)
}

func determineFieldKind(fieldDesc *gogodesc.FieldDescriptorProto) FieldKind {
	if isMapField(fieldDesc) {
		return FieldKindMap
	}
	if fieldDesc.GetLabel() == gogodesc.FieldDescriptorProto_LABEL_REPEATED {
		return FieldKindRepeated
	}
	if fieldDesc.GetType() == gogodesc.FieldDescriptorProto_TYPE_MESSAGE {
		return FieldKindMessage
	}
	if fieldDesc.GetType() == gogodesc.FieldDescriptorProto_TYPE_ENUM {
		return FieldKindEnum
	}
	return FieldKindScalar
}

// determineElementKind 确定repeated字段元素的kind
func determineElementKind(fieldDesc *gogodesc.FieldDescriptorProto) FieldKind {
	if fieldDesc.GetType() == gogodesc.FieldDescriptorProto_TYPE_MESSAGE {
		return FieldKindMessage
	}
	if fieldDesc.GetType() == gogodesc.FieldDescriptorProto_TYPE_ENUM {
		return FieldKindEnum
	}
	return FieldKindScalar
}

func generateFieldTags(field *GoField, fieldDesc *gogodesc.FieldDescriptorProto) *FieldTags {
	tags := &FieldTags{
		JSON:     generateJSONTag(field),
		Protobuf: generateProtobufTag(field, fieldDesc),
		BSON:     generateBSONTag(field),
		Custom:   []CustomTag{},
	}
	return tags
}

func generateJSONTag(field *GoField) *JSONTag {
	return &JSONTag{
		Name:      field.ProtoName,
		OmitEmpty: field.IsOptional || field.IsRepeatedType(),
		Ignore:    false,
	}
}

func generateBSONTag(field *GoField) *BsonTag {
	return &BsonTag{
		Name:      field.ProtoName,
		OmitEmpty: field.IsOptional || field.IsRepeatedType(),
		Ignore:    false,
	}
}

func generateProtobufTag(field *GoField, fieldDesc *gogodesc.FieldDescriptorProto) *ProtobufTag {
	wireType := getWireType(fieldDesc)
	rule := "opt"
	if field.IsRepeatedType() {
		rule = "rep"
	}
	return &ProtobufTag{
		Type:   wireType,
		Number: field.Number,
		Name:   field.ProtoName,
		Rule:   rule,
	}
}

func getWireType(fieldDesc *gogodesc.FieldDescriptorProto) string {
	switch fieldDesc.GetType() {
	case gogodesc.FieldDescriptorProto_TYPE_BOOL,
		gogodesc.FieldDescriptorProto_TYPE_ENUM,
		gogodesc.FieldDescriptorProto_TYPE_INT32, gogodesc.FieldDescriptorProto_TYPE_SINT32, gogodesc.FieldDescriptorProto_TYPE_UINT32,
		gogodesc.FieldDescriptorProto_TYPE_INT64, gogodesc.FieldDescriptorProto_TYPE_SINT64, gogodesc.FieldDescriptorProto_TYPE_UINT64:
		return "varint"
	case gogodesc.FieldDescriptorProto_TYPE_FIXED32, gogodesc.FieldDescriptorProto_TYPE_SFIXED32, gogodesc.FieldDescriptorProto_TYPE_FLOAT:
		return "fixed32"
	case gogodesc.FieldDescriptorProto_TYPE_FIXED64, gogodesc.FieldDescriptorProto_TYPE_SFIXED64, gogodesc.FieldDescriptorProto_TYPE_DOUBLE:
		return "fixed64"
	case gogodesc.FieldDescriptorProto_TYPE_STRING, gogodesc.FieldDescriptorProto_TYPE_BYTES, gogodesc.FieldDescriptorProto_TYPE_MESSAGE:
		return "bytes"
	default:
		return "varint"
	}
}

func toGoFieldName(protoName string) string {
	parts := strings.Split(protoName, "_")
	var result strings.Builder
	for _, part := range parts {
		if part != "" {
			result.WriteString(strings.Title(part))
		}
	}
	return result.String()
}

// ValidateGoStruct 验证Go结构体的有效性
func ValidateGoStruct(goStruct *GoStruct) error {
	if goStruct == nil {
		return fmt.Errorf("GoStruct is nil")
	}
	if goStruct.Name == "" {
		return fmt.Errorf("GoStruct name is empty")
	}
	if goStruct.PackageName == "" {
		return fmt.Errorf("GoStruct package name is empty")
	}
	fieldNames := make(map[string]bool)
	for _, field := range goStruct.Fields {
		if err := ValidateGoField(field); err != nil {
			return fmt.Errorf("invalid field %s: %v", field.Name, err)
		}
		if fieldNames[field.Name] {
			return fmt.Errorf("duplicate field name: %s", field.Name)
		}
		fieldNames[field.Name] = true
	}
	return nil
}

// ValidateGoField 验证Go字段的有效性
func ValidateGoField(field *GoField) error {
	if field == nil {
		return fmt.Errorf("GoField is nil")
	}
	if field.Name == "" {
		return fmt.Errorf("field name is empty")
	}
	if field.Type == "" {
		return fmt.Errorf("field type is empty")
	}
	if field.Number <= 0 {
		return fmt.Errorf("invalid field number: %d", field.Number)
	}
	if field.IsMapType() {
		if field.MapInfo.KeyType == "" {
			return fmt.Errorf("map field missing key type")
		}
		if field.MapInfo.ValueType == "" {
			return fmt.Errorf("map field missing value type")
		}
	}
	return nil
}

// OptimizeGoStructs 优化Go结构体列表
func OptimizeGoStructs(structs []*GoStruct) []*GoStruct {
	seen := make(map[string]bool)
	var optimized []*GoStruct
	for _, s := range structs {
		key := fmt.Sprintf("%s.%s", s.PackageName, s.Name)
		if !seen[key] {
			seen[key] = true
			optimized = append(optimized, s)
		}
	}
	sort.Slice(optimized, func(i, j int) bool { return optimized[i].Name < optimized[j].Name })
	return optimized
}

// renderFieldTags 渲染字段标签为字符串
func renderFieldTags(tags *FieldTags) string {
	if tags == nil {
		return ""
	}
	// Index custom tags by key for lookup/override
	customByKey := make(map[string]string, len(tags.Custom))
	for _, ct := range tags.Custom {
		customByKey[ct.Key] = ct.Value
	}

	var tagParts []string
	// json
	if tags.JSON != nil && !tags.JSON.Ignore {
		jsonTag := tags.JSON.Name
		if tags.JSON.OmitEmpty {
			jsonTag += ",omitempty"
		}
		tagParts = append(tagParts, fmt.Sprintf(`json:"%s"`, jsonTag))
	}

	// bson: prefer explicit BSON tag; then custom override; else mirror json
	if tags.BSON != nil && !tags.BSON.Ignore {
		bsonTag := tags.BSON.Name
		if tags.BSON.OmitEmpty {
			bsonTag += ",omitempty"
		}
		tagParts = append(tagParts, fmt.Sprintf(`bson:"%s"`, bsonTag))
	}

	// protobuf
	if tags.Protobuf != nil {
		protoTag := fmt.Sprintf("%s,%d,%s,name=%s",
			tags.Protobuf.Type, tags.Protobuf.Number, tags.Protobuf.Rule, tags.Protobuf.Name)
		tagParts = append(tagParts, fmt.Sprintf(`protobuf:"%s"`, protoTag))
	}
	// remaining custom tags
	for k, v := range customByKey {
		tagParts = append(tagParts, fmt.Sprintf(`%s:"%s"`, k, v))
	}
	return strings.Join(tagParts, " ")
}

// getElementType 获取字段的元素类型
func getElementType(fieldDesc *gogodesc.FieldDescriptorProto) string {
	if fieldDesc.GetType() == gogodesc.FieldDescriptorProto_TYPE_MESSAGE || fieldDesc.GetType() == gogodesc.FieldDescriptorProto_TYPE_ENUM {
		return shortTypeName(trimLeadingDot(fieldDesc.GetTypeName()))
	}
	switch fieldDesc.GetType() {
	case gogodesc.FieldDescriptorProto_TYPE_BOOL:
		return "bool"
	case gogodesc.FieldDescriptorProto_TYPE_INT32:
		return "int32"
	case gogodesc.FieldDescriptorProto_TYPE_INT64:
		return "int64"
	case gogodesc.FieldDescriptorProto_TYPE_UINT32:
		return "uint32"
	case gogodesc.FieldDescriptorProto_TYPE_UINT64:
		return "uint64"
	case gogodesc.FieldDescriptorProto_TYPE_SINT32:
		return "sint32"
	case gogodesc.FieldDescriptorProto_TYPE_SINT64:
		return "sint64"
	case gogodesc.FieldDescriptorProto_TYPE_FIXED32:
		return "fixed32"
	case gogodesc.FieldDescriptorProto_TYPE_FIXED64:
		return "fixed64"
	case gogodesc.FieldDescriptorProto_TYPE_SFIXED32:
		return "sfixed32"
	case gogodesc.FieldDescriptorProto_TYPE_SFIXED64:
		return "sfixed64"
	case gogodesc.FieldDescriptorProto_TYPE_FLOAT:
		return "float"
	case gogodesc.FieldDescriptorProto_TYPE_DOUBLE:
		return "double"
	case gogodesc.FieldDescriptorProto_TYPE_STRING:
		return "string"
	case gogodesc.FieldDescriptorProto_TYPE_BYTES:
		return "bytes"
	default:
		return fieldDesc.GetType().String()
	}
}

// getGoBaseTypeWithConfig 使用配置获取Go基础类型
func getGoBaseTypeWithConfig(fieldDesc *gogodesc.FieldDescriptorProto, config *GenerationConfig) string {
	switch fieldDesc.GetType() {
	case gogodesc.FieldDescriptorProto_TYPE_MESSAGE:
		typeName := shortTypeName(trimLeadingDot(fieldDesc.GetTypeName()))
		if customType, exists := config.TypeMapping[typeName]; exists {
			return fmt.Sprintf("*%s", customType)
		}
		return fmt.Sprintf("*%s", typeName)
	case gogodesc.FieldDescriptorProto_TYPE_ENUM:
		typeName := shortTypeName(trimLeadingDot(fieldDesc.GetTypeName()))
		if customType, exists := config.TypeMapping[typeName]; exists {
			return customType
		}
		return typeName
	case gogodesc.FieldDescriptorProto_TYPE_BOOL:
		return "bool"
	case gogodesc.FieldDescriptorProto_TYPE_INT32:
		return "int32"
	case gogodesc.FieldDescriptorProto_TYPE_INT64:
		return "int64"
	case gogodesc.FieldDescriptorProto_TYPE_UINT32:
		return "uint32"
	case gogodesc.FieldDescriptorProto_TYPE_UINT64:
		return "uint64"
	case gogodesc.FieldDescriptorProto_TYPE_SINT32:
		return "int32"
	case gogodesc.FieldDescriptorProto_TYPE_SINT64:
		return "int64"
	case gogodesc.FieldDescriptorProto_TYPE_FIXED32:
		return "uint32"
	case gogodesc.FieldDescriptorProto_TYPE_FIXED64:
		return "uint64"
	case gogodesc.FieldDescriptorProto_TYPE_SFIXED32:
		return "int32"
	case gogodesc.FieldDescriptorProto_TYPE_SFIXED64:
		return "int64"
	case gogodesc.FieldDescriptorProto_TYPE_FLOAT:
		return "float32"
	case gogodesc.FieldDescriptorProto_TYPE_DOUBLE:
		return "float64"
	case gogodesc.FieldDescriptorProto_TYPE_STRING:
		return "string"
	case gogodesc.FieldDescriptorProto_TYPE_BYTES:
		return "[]byte"
	default:
		return "any"
	}
}

// GenerateGoStructCode 根据GoStruct列表生成Go代码
func GenerateGoStructCode(structs []*GoStruct) string {
	var (
		builder              strings.Builder
		mongoUpdateGenerator = MongoUpdateGenerator{}
	)
	if len(structs) > 0 {
		builder.WriteString(fmt.Sprintf(`// Code generated from proto files. DO NOT EDIT.

package %s

`, structs[0].PackageName))
		// Imports: dirty, xmap, and mgo_builder are required for components
		needDirty := requiresDirtyImport(structs)
		needXmap := false
		needMgoBuilder := false
		// naive scan: if any Component has a map field, we'll emit OpsV2 using xmap
		// if any Entity has component fields, we'll emit BuildMongoUpdate using mgo_builder
		for _, s := range structs {
			if shouldEmbedDirtyTracker(s.Name) && isComponent(s.Name) {
				needMgoBuilder = true // Components need mgo_builder for BuildMongoUpdate
				for _, f := range s.Fields {
					if strings.HasPrefix(f.Type, "map[") && strings.Contains(f.Type, "]") {
						needXmap = true
						break
					}
				}
			}
			// Check if Entity has component fields
			if isEntity(s.Name) && hasComponentFields(s) {
				needMgoBuilder = true // Entities with components need mgo_builder
			}
		}
		if needDirty || needXmap || needMgoBuilder {
			builder.WriteString("import (\n")
			if needMgoBuilder {
				builder.WriteString("\t\"gitee.com/orbit-w/orbit/lib/module/db/mgo_builder\"\n")
			}
			if needDirty {

				builder.WriteString("\tdirty\"gitee.com/orbit-w/orbit/lib/module/dirty/dirty_tracker\"\n")
			}
			if needXmap {
				builder.WriteString("\t\"gitee.com/orbit-w/orbit/lib/module/dirty/xmap\"\n")
			}
			builder.WriteString(")\n\n")
		}
	}
	// Build index for cross-struct lookups (e.g., Entity -> Component)
	structIndex := make(map[string]*GoStruct, len(structs))
	for _, s := range structs {
		structIndex[s.Name] = s
	}

	for _, goStruct := range structs {
		builder.WriteString(generateSingleGoStruct(goStruct))
		builder.WriteString("\n\n")

		// If this is a Component, place constructor immediately after struct declaration
		if shouldEmbedDirtyTracker(goStruct.Name) && isComponent(goStruct.Name) {
			ctor := generateComponentConstructor(goStruct)
			if ctor != "" {
				builder.WriteString(ctor)
				builder.WriteString("\n\n")
			}

			// Then generate dirty bits and accessors
			bits := generateComponentDirtyBits(goStruct)
			if bits != "" {
				builder.WriteString(bits)
				builder.WriteString("\n\n")
			}
			accessors := generateComponentAccessors(goStruct)
			if accessors != "" {
				builder.WriteString(accessors)
				builder.WriteString("\n\n")
			}

			// Mongo update and dirty clearing helpers
			mongoUpdate := mongoUpdateGenerator.GenerateMongoUpdateMethod(goStruct)
			if mongoUpdate != "" {
				builder.WriteString(mongoUpdate)
				builder.WriteString("\n\n")
			}

			clearRec := generateComponentClearDirtyRecursiveMethod(goStruct)
			if clearRec != "" {
				builder.WriteString(clearRec)
				builder.WriteString("\n\n")
			}
		}

		// If this is an Entity, place safe getters and optional forwarders below struct
		if isEntity(goStruct.Name) {
			// Safe Component getters (lazy init via New<Component>())
			getters := generateEntitySafeComponentGetters(goStruct)
			if getters != "" {
				builder.WriteString(getters)
				builder.WriteString("\n\n")
			}
			// Aggregated BuildMongoUpdate across all Components
			agg := generateEntityMongoUpdateAggregator(goStruct)
			if agg != "" {
				builder.WriteString(agg)
				builder.WriteString("\n\n")
			}

			if genOptions.EmitEntityOpsForwarders {
				fwd := generateEntityOpsForwarders(goStruct, structIndex)
				if fwd != "" {
					builder.WriteString(fwd)
					builder.WriteString("\n\n")
				}
			}
		}

		// Generate DeepCopy method for all structures
		deepCopy := generateDeepCopyMethod(goStruct)
		if deepCopy != "" {
			builder.WriteString(deepCopy)
			builder.WriteString("\n\n")
		}
	}
	return beautifyGeneratedGo(builder.String())
}

// GroupStructsBySourceFile 将结构体按来源 proto 文件分组
func GroupStructsBySourceFile(structs []*GoStruct) map[string][]*GoStruct {
	groups := make(map[string][]*GoStruct)
	for _, s := range structs {
		file := s.SourceFile
		if file == "" {
			file = "_unknown.proto"
		}
		groups[file] = append(groups[file], s)
	}
	return groups
}

// generateSingleGoStruct 生成单个Go结构体代码
func generateSingleGoStruct(goStruct *GoStruct) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("// %s\n", goStruct.Comment))
	builder.WriteString(fmt.Sprintf("type %s struct {\n", goStruct.Name))
	if shouldEmbedDirtyTracker(goStruct.Name) {
		builder.WriteString("\tdirty.DirtyTracker\n")
	}
	for _, field := range goStruct.Fields {
		// Business rule: for Entity top-level id, override bson tag to _id
		if isEntity(goStruct.Name) && strings.EqualFold(field.Name, "Id") {
			if field.Tags == nil {
				field.Tags = &FieldTags{}
			}
			// ensure JSON tag exists
			if field.Tags.JSON == nil {
				field.Tags.JSON = &JSONTag{Name: field.ProtoName}
			}
			// inject/override bson custom tag
			if field.Tags.BSON == nil {
				field.Tags.BSON = &BsonTag{Name: "_id"}
			}
		}
		builder.WriteString(fmt.Sprintf("\t%s %s", field.Name, field.Type))
		if field.Tags != nil {
			tagStr := renderFieldTags(field.Tags)
			if tagStr != "" {
				builder.WriteString(fmt.Sprintf(" `%s`", tagStr))
			}
		}
		if field.ProtoType != "" && field.ProtoType != field.Type {
			builder.WriteString(fmt.Sprintf(" // proto: %s", field.ProtoType))
		}
		builder.WriteString("\n")
	}
	builder.WriteString("}")
	return builder.String()
}

// GenerateSingleGoStruct exported wrapper.
func GenerateSingleGoStruct(s *GoStruct) string { return generateSingleGoStruct(s) }

func shouldEmbedDirtyTracker(structName string) bool {
	// Only Components embed dirty tracker; Entities are passive containers
	return isComponent(structName)
}

func requiresDirtyImport(structs []*GoStruct) bool {
	for _, s := range structs {
		if shouldEmbedDirtyTracker(s.Name) {
			return true
		}
	}
	return false
}

// ShouldEmbedDirtyTracker exported wrapper.
func ShouldEmbedDirtyTracker(structName string) bool { return shouldEmbedDirtyTracker(structName) }

// RequiresDirtyImport exported wrapper.
func RequiresDirtyImport(structs []*GoStruct) bool { return requiresDirtyImport(structs) }

// Entity/Component type helpers

// isEntity checks if a struct name represents an Entity
func isEntity(structName string) bool {
	return strings.HasSuffix(structName, "Entity")
}

// isComponent checks if a struct name represents a Component
func isComponent(structName string) bool {
	return strings.HasSuffix(structName, "Component")
}

// isComponentType checks if a Go type represents a Component (handles pointer types)
func isComponentType(goType string) bool {
	typeName := extractTypeName(goType)
	return isComponent(typeName)
}

// extractTypeName extracts the base type name from a Go type string (removes * prefix)
func extractTypeName(goType string) string {
	return strings.TrimPrefix(goType, "*")
}

// hasComponentFields checks if a struct has any component-typed fields
func hasComponentFields(s *GoStruct) bool {
	if s == nil {
		return false
	}
	for _, f := range s.Fields {
		if isComponentType(f.Type) {
			return true
		}
	}
	return false
}

// Exported wrappers for external use
func IsEntity(structName string) bool      { return isEntity(structName) }
func IsComponent(structName string) bool   { return isComponent(structName) }
func IsComponentType(goType string) bool   { return isComponentType(goType) }
func ExtractTypeName(goType string) string { return extractTypeName(goType) }
func HasComponentFields(s *GoStruct) bool  { return hasComponentFields(s) }

// beautifyGeneratedGo formats the generated source using go/format and
// compacts excessive blank lines to a single blank line between top-level blocks.
func beautifyGeneratedGo(src string) string {
	// Compact 3+ consecutive newlines into exactly two (one blank line)
	re := regexp.MustCompile("\n{3,}")
	compact := re.ReplaceAllString(src, "\n\n")
	// Format with gofmt
	formatted, err := format.Source([]byte(compact))
	if err == nil {
		return string(formatted)
	}
	// Fallback to compact if formatting fails
	return compact
}

// local helpers
func trimLeadingDot(s string) string {
	if len(s) > 0 && s[0] == '.' {
		return s[1:]
	}
	return s
}

func shortTypeName(full string) string {
	if i := strings.LastIndex(full, "."); i >= 0 {
		return full[i+1:]
	}
	return full
}

// buildMessageIndex builds a fully qualified name → descriptor index for all messages (top-level and nested)
func buildMessageIndex(ctx ProtoContext) {
	fds := ctx.GetFileDescriptorSet()
	if fds == nil {
		return
	}
	for _, file := range fds.File {
		pkg := file.GetPackage()
		for _, m := range file.GetMessageType() {
			addMessageToIndex(ctx, m, pkg, "")
		}
	}
}

func addMessageToIndex(ctx ProtoContext, msg *gogodesc.DescriptorProto, pkg string, parentPrefix string) {
	msgName := msg.GetName()
	var fullName string
	if parentPrefix != "" {
		fullName = parentPrefix + "." + msgName
	} else {
		fullName = msgName
	}
	if pkg != "" {
		fullName = pkg + "." + fullName
	}
	ctx.AddMessageToIndex(fullName, msg)
	for _, nested := range msg.GetNestedType() {
		var newPrefix string
		if parentPrefix != "" {
			newPrefix = parentPrefix + "." + msgName
		} else {
			newPrefix = msgName
		}
		addMessageToIndex(ctx, nested, pkg, newPrefix)
	}
}

// ParseNestedMessage parses one nested message into GoStructs including deeper nested children.
func ParseNestedMessage(ctx ProtoContext, nestedMsg *gogodesc.DescriptorProto, config *GenerationConfig, parentName string) []*GoStruct {
	var structs []*GoStruct
	if IsMapEntryMessage(nestedMsg) {
		return nil
	}
	s := ConvertMessageToGoStructWithConfig(ctx, nestedMsg, config, true, parentName)
	if s != nil {
		structs = append(structs, s)
	}
	nestedStructs := extractNestedGoStructsWithConfig(ctx, nestedMsg, config, parentName+"_"+nestedMsg.GetName())
	structs = append(structs, nestedStructs...)
	return structs
}
