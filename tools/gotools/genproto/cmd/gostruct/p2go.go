package gostruct

import (
	"fmt"

	gogodesc "github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
)

// ConvertProtoField2GoType 将单个 proto 字段转换为 Go 数据类型
func ConvertProtoField2GoType(fieldDesc *gogodesc.FieldDescriptorProto, config *GenerationConfig) string {
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
