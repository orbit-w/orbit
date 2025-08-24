package gostruct

import "strings"

// isSystemProtoFile 判断是否为系统 proto 文件
// 在解析 FileDescriptorSet 时过滤掉 Google 的系统 proto（well-known types，如 timestamp.proto、descriptor.proto 等），
// 避免为这些系统消息生成 Go 结构体。
func isSystemProtoFile(filename string) bool {
	if filename == "" {
		return false
	}
	systemFiles := []string{
		"google/protobuf/descriptor.proto",
		"google/protobuf/any.proto",
		"google/protobuf/api.proto",
		"google/protobuf/duration.proto",
		"google/protobuf/empty.proto",
		"google/protobuf/field_mask.proto",
		"google/protobuf/source_context.proto",
		"google/protobuf/struct.proto",
		"google/protobuf/timestamp.proto",
		"google/protobuf/type.proto",
		"google/protobuf/wrappers.proto",
	}
	for _, systemFile := range systemFiles {
		if strings.HasSuffix(filename, systemFile) {
			return true
		}
	}
	return false
}
