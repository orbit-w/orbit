package gostruct

import (
	"fmt"

	gogodesc "github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
)

// generateStructComment 生成Go结构体的标准化注释文档
// 该函数为生成的Go结构体创建描述性注释，说明其与原始protobuf消息的对应关系
//
// 参数说明:
//   - protoName: proto消息的名称，如 "PlayerEntity"
//   - fullName: proto消息的完整名称，通常与protoName相同，如 "PlayerEntity"
//   - isNested: 是否为嵌套消息。true表示该消息定义在另一个消息内部
//   - parentName: 父消息名称，仅当isNested=true时有效，如 "GameData"
//
// 返回值:
//
//	返回格式化的注释字符串，用于Go结构体声明前的文档注释
//
// 生成规则:
//   - 顶级消息: "{StructName} represents the proto message {MessageName}"
//   - 嵌套消息: "{StructName} is a nested message from {ParentName} (proto: {MessageName})"
//
// 使用示例:
//   - generateStructComment("PlayerEntity", "PlayerEntity", false, "")
//     → "PlayerEntity represents the proto message PlayerEntity"
//   - generateStructComment("Config", "Config", true, "PlayerEntity")
//     → "Config is a nested message from PlayerEntity (proto: Config)"
func generateStructComment(protoName, fullName string, isNested bool, parentName string) string {
	if isNested {
		return fmt.Sprintf("%s is a nested message from %s (proto: %s)", protoName, parentName, fullName)
	}
	return fmt.Sprintf("%s represents the proto message %s", protoName, fullName)
}

// extractFieldComment 生成Go字段的标准化注释文档
// 该函数为生成的Go字段创建描述性注释，包含字段名和proto字段编号信息
//
// 参数说明:
//   - fieldDesc: protobuf字段描述符，包含字段的所有元信息
//
// 返回值:
//
//	返回格式化的字段注释字符串，格式为: "Field {FieldName} (number: {FieldNumber})"
//
// 使用示例:
//   - 对于proto字段 "int64 id = 1;"
//   - extractFieldComment(fieldDesc) → "Field id (number: 1)"
func extractFieldComment(fieldDesc *gogodesc.FieldDescriptorProto) string {
	return fmt.Sprintf("Field %s (number: %d)", fieldDesc.GetName(), fieldDesc.GetNumber())
}

// GenerateStructComment 导出版本的generateStructComment，供外部包使用
// 生成Go结构体的标准化注释文档
func GenerateStructComment(protoName, fullName string, isNested bool, parentName string) string {
	return generateStructComment(protoName, fullName, isNested, parentName)
}

// ExtractFieldComment 导出版本的extractFieldComment，供外部包使用
// 生成Go字段的标准化注释文档
func ExtractFieldComment(fieldDesc *gogodesc.FieldDescriptorProto) string {
	return extractFieldComment(fieldDesc)
}
