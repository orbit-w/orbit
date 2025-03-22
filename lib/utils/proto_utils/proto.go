package proto_utils

import (
	"fmt"
	"reflect"
	"strings"
)

func ParseMessageName(resp any) string {
	// 检查nil响应
	if resp == nil {
		return ""
	}

	// 使用反射获取类型信息
	v := reflect.ValueOf(resp)
	// 处理指针类型
	if v.Kind() == reflect.Ptr {
		// 检查空指针
		if v.IsNil() {
			return ""
		}
		// 获取指针指向的元素类型
		v = v.Elem()
	}

	// 获取类型名称
	typeName := v.Type().Name()
	// 如果类型名为空（例如匿名结构体），使用完整的类型字符串
	if typeName == "" {
		typeName = fmt.Sprintf("%T", resp)
		// 提取短类型名（去除包名和指针前缀）
		parts := strings.Split(typeName, ".")
		typeName = parts[len(parts)-1]
		// 去除可能的指针前缀(*)
		typeName = strings.TrimPrefix(typeName, "*")
	}

	return typeName
}
