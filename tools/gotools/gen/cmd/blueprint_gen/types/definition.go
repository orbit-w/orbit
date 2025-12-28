package blueprint_types

// Definition (符号定义接口)
// 所有顶层定义（Entity, Manager, Message, Struct, Enum）都必须实现此接口
// 这是 AST 节点的抽象表示，用于类型引用解析和语义分析
type Definition interface {
	// GetName 获取符号名称 (Symbol Name)
	GetName() string

	// GetKind 获取定义种类
	GetKind() DefinitionKind

	// GetFile 获取定义所在源文件 (Source File)
	// 返回 YAML 文件路径或 Proto 文件路径
	GetFile() string

	// GetFields 获取成员字段列表（用于遍历 AST）
	// 对于 Enum 类型，返回 nil 或空切片
	// 对于有字段的类型（Entity, Manager, Message 等），返回字段列表
	GetFields() []*Field
}
