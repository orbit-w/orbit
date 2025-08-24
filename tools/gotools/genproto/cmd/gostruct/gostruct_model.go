package gostruct

import (
	"fmt"
	"sort"
	"strings"

	gogodesc "github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
)

// Map helpers for gogo descriptors
// isMapField performs a lightweight, field-level check to determine whether a
// FieldDescriptorProto likely represents a protobuf map.
//
// Important:
//   - In the protobuf descriptor model, a map<K,V> field is encoded as
//     "repeated <Message Entry>" where the referenced message has
//     option map_entry = true and two fields: key(1), value(2).
//   - Because this function does not have access to the referenced message
//     descriptor, it uses a conservative pre-check: TYPE_MESSAGE + LABEL_REPEATED
//     and an Entry suffix convention. This is only a hint used to short-circuit
//     obvious non-map fields.
//   - The authoritative detection happens when we resolve the referenced message
//     via the message index and call isMapEntryMessage on that DescriptorProto.
//     See getMapKeyValueTypes and convertFieldToGoFieldWithConfig for the
//     canonical path.
func isMapField(fieldDesc *gogodesc.FieldDescriptorProto) bool {
	if fieldDesc.GetType() != gogodesc.FieldDescriptorProto_TYPE_MESSAGE {
		return false
	}
	if fieldDesc.GetLabel() != gogodesc.FieldDescriptorProto_LABEL_REPEATED {
		return false
	}
	typeName := trimLeadingDot(fieldDesc.GetTypeName())
	short := shortTypeName(typeName)
	return strings.HasSuffix(short, "Entry")
}

// ConvertProtoField2GoFieldWithConfig 使用配置将proto字段转换为Go字段信息
func ConvertProtoField2GoFieldWithConfig(ctx ProtoContext, fieldDesc *gogodesc.FieldDescriptorProto, config *GenerationConfig) *GoField {
	protoName := fieldDesc.GetName()
	goName := toGoFieldName(protoName)
	protoType := getProtoFieldTypeWithIndex(ctx, fieldDesc)

	// 预先获取 map 字段的键值类型（避免重复调用）
	var keyField, valueField *gogodesc.FieldDescriptorProto
	isMap := isMapField(fieldDesc)
	if isMap {
		keyField, valueField = getMapKeyValueTypes(ctx, fieldDesc)
	}

	// 生成 Go 类型
	var goType string
	if isMap {
		if keyField != nil && valueField != nil {
			keyGoType := ConvertProtoField2GoType(keyField, config)
			valueGoType := ConvertProtoField2GoType(valueField, config)
			goType = fmt.Sprintf("map[%s]%s", keyGoType, valueGoType)
		} else {
			// 如果无法获取真实的map类型，回退到基础类型转换
			goType = getGoBaseTypeWithConfig(fieldDesc, config)
		}
	} else {
		// 对于非map字段，使用基础类型转换并处理repeated
		baseType := getGoBaseTypeWithConfig(fieldDesc, config)
		if fieldDesc.GetLabel() == gogodesc.FieldDescriptorProto_LABEL_REPEATED {
			goType = fmt.Sprintf("[]%s", baseType)
		} else {
			goType = baseType
		}
	}

	kind := determineFieldKind(fieldDesc)

	field := &GoField{
		Name:       goName,
		ProtoName:  protoName,
		Type:       goType,
		ProtoType:  protoType,
		Number:     fieldDesc.GetNumber(),
		Kind:       kind,
		Comment:    ExtractFieldComment(fieldDesc),
		IsOptional: false,
		IsOneof:    fieldDesc.OneofIndex != nil,
	}

	switch {
	case field.IsMapType():
		if keyField != nil && valueField != nil {
			field.MapInfo = &MapInfo{
				KeyType:   getElementType(keyField),
				KeyKind:   determineFieldKind(keyField),
				ValueType: getElementType(valueField),
				ValueKind: determineFieldKind(valueField),
			}
		} else {
			// 如果无法获取真实的map key-value类型，跳过MapInfo设置
			// 这种情况说明protobuf descriptor可能有问题，不应该使用错误的fallback
		}
	case field.IsRepeatedType():
		field.RepeatedInfo = &RepeatedInfo{
			ValueType: getElementType(fieldDesc),
			ValueKind: determineElementKind(fieldDesc), // 使用元素的kind
		}
	}

	if fieldDesc.OneofIndex != nil {
		field.OneofName = "oneof"
		field.OneofIndex = fieldDesc.GetOneofIndex()
	}

	field.Tags = generateFieldTags(field, fieldDesc)
	return field
}

// ConvertMessageToGoStructWithConfig 使用配置将proto消息转换为Go结构体信息
func ConvertMessageToGoStructWithConfig(ctx ProtoContext, msgDesc *gogodesc.DescriptorProto, config *GenerationConfig, isNested bool, parentName string) *GoStruct {
	protoName := msgDesc.GetName()
	fullName := protoName

	goStruct := &GoStruct{
		Name:        protoName,
		ProtoName:   protoName,
		PackageName: config.PackageName,
		Fields:      []*GoField{},
		IsNested:    isNested,
		ParentName:  parentName,
		FullName:    fullName,
		Comment:     GenerateStructComment(protoName, fullName, isNested, parentName),
	}

	fields := msgDesc.GetField()
	for _, fieldDesc := range fields {
		goField := ConvertProtoField2GoFieldWithConfig(ctx, fieldDesc, config)
		if config.FieldFilter != nil && !config.FieldFilter(goField) {
			continue
		}
		goStruct.Fields = append(goStruct.Fields, goField)
	}

	sort.Slice(goStruct.Fields, func(i, j int) bool {
		return goStruct.Fields[i].Number < goStruct.Fields[j].Number
	})
	return goStruct
}

// getMapKeyValueTypes 从消息索引中获取 map 字段的实际键值类型。
func getMapKeyValueTypes(ctx ProtoContext, fieldDesc *gogodesc.FieldDescriptorProto) (keyType, valueType *gogodesc.FieldDescriptorProto) {
	if !isMapField(fieldDesc) {
		return nil, nil
	}

	entryTypeName := trimLeadingDot(fieldDesc.GetTypeName())
	entryMsg, exists := ctx.GetMessageFromIndex(entryTypeName)
	if !exists {
		return nil, nil
	}

	var keyField, valueField *gogodesc.FieldDescriptorProto
	for _, field := range entryMsg.GetField() {
		if field.GetName() == "key" && field.GetNumber() == 1 {
			keyField = field
		} else if field.GetName() == "value" && field.GetNumber() == 2 {
			valueField = field
		}
	}
	return keyField, valueField
}
