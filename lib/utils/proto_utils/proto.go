package proto_utils

import (
	"fmt"
	"reflect"
	"strings"
)

// ParseMessageName 提取消息的名称
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

// ParsePackageName 提取消息的包名
func ParsePackageName(resp any) string {
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

	// 获取完整的类型字符串，例如 core.Request_SearchBook
	fullTypeName := fmt.Sprintf("%T", resp)

	// 去除可能的指针前缀(*)
	fullTypeName = strings.TrimPrefix(fullTypeName, "*")

	// 分割包路径与类型名
	parts := strings.Split(fullTypeName, ".")

	// 如果长度小于2，无法提取包名
	if len(parts) < 2 {
		return ""
	}

	// 获取类型名用于在映射表中查找
	typeName := parts[len(parts)-1]

	// 先从消息名到包名的映射表中查找
	if pkg, ok := MessageToPackageMap[typeName]; ok {
		return pkg
	}

	// 获取直接包名
	directPkg := parts[len(parts)-2]

	// 根据直接包名映射到协议包名（如 core -> Core）
	pkgName := NormalizePackageName(directPkg)

	// 将消息名和包名添加到映射表中以便后续使用
	RegisterMessagePackage(typeName, pkgName)

	return pkgName
}

// NormalizePackageName 标准化包名，将首字母大写
func NormalizePackageName(pkgName string) string {
	switch strings.ToLower(pkgName) {
	case "core":
		return "Core"
	case "season":
		return "Season"
	// 添加其他包名映射...
	default:
		// 首字母大写作为默认处理
		if len(pkgName) > 0 {
			return strings.ToUpper(pkgName[:1]) + pkgName[1:]
		}
		return pkgName
	}
}

// MessageToPackageMap 消息名称到包名的映射
var MessageToPackageMap = make(map[string]string)

// RegisterMessagePackage 注册消息名和包名的映射关系
// 导出此函数以供其他包使用
func RegisterMessagePackage(messageName, packageName string) {
	MessageToPackageMap[messageName] = packageName
}

// InitMessageToPackageMap 初始化消息名到包名的映射表
// 在应用启动时调用，可以预先加载已知的映射关系
func InitMessageToPackageMap() {
	// 注册Core包的消息
	RegisterMessagePackage("OK", "Core")
	RegisterMessagePackage("Fail", "Core")
	RegisterMessagePackage("Request_SearchBook", "Core")
	RegisterMessagePackage("Request_SearchBook_Rsp", "Core")
	RegisterMessagePackage("Request_HeartBeat", "Core")
	RegisterMessagePackage("Notify_BeAttacked", "Core")

	// 注册Season包的消息
	RegisterMessagePackage("Request_SeasonInfo", "Season")

	// 可以根据需要添加更多映射关系
}
