package blueprint_types

import (
	"fmt"
)

// SymbolTable 全局符号表
// 存储所有已定义的类型（Entity, Manager, Message, Enum 等）
// 用于 Pass 3: Reference Resolution 阶段的符号查找
type SymbolTable struct {
	// symbols 符号名称到定义的映射
	// Key: 符号名称（如 "HeroManager", "PlayerEntity"）
	// Value: Definition 接口实例
	symbols map[string]Definition

	// scopedSymbols 带作用域的符号映射（可选）
	// Key: 完整限定名（如 "mme.HeroManager", "core.Book"）
	// Value: Definition 接口实例
	scopedSymbols map[string]Definition
}

// NewSymbolTable 创建新的符号表
func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		symbols:       make(map[string]Definition),
		scopedSymbols: make(map[string]Definition),
	}
}

// Register 注册一个符号定义到符号表
// 如果符号已存在，返回错误（重复定义检查）
func (st *SymbolTable) Register(def Definition) error {
	name := def.GetName()
	if name == "" {
		return fmt.Errorf("cannot register definition with empty name")
	}

	// 检查重复定义
	if existing, exists := st.symbols[name]; exists {
		return fmt.Errorf("duplicate definition: symbol '%s' already defined in %s (kind: %s)",
			name, existing.GetFile(), existing.GetKind())
	}

	st.symbols[name] = def
	return nil
}

// RegisterWithScope 注册带作用域的符号定义
// scope: 作用域名称（如包名 "mme", "core"）
// def: 符号定义
func (st *SymbolTable) RegisterWithScope(scope string, def Definition) error {
	// 先注册不带作用域的符号
	if err := st.Register(def); err != nil {
		return err
	}

	// 再注册带作用域的符号
	if scope != "" {
		qualifiedName := scope + "." + def.GetName()
		st.scopedSymbols[qualifiedName] = def
	}

	return nil
}

// Lookup 根据符号名称查找定义
// 先查找简单名称，再查找带作用域的名称
func (st *SymbolTable) Lookup(name string) (Definition, bool) {
	// 先尝试简单名称查找
	if def, exists := st.symbols[name]; exists {
		return def, true
	}

	// 再尝试带作用域的名称查找
	if def, exists := st.scopedSymbols[name]; exists {
		return def, true
	}

	return nil, false
}

// MustLookup 查找符号，如果不存在则 panic
func (st *SymbolTable) MustLookup(name string) Definition {
	def, exists := st.Lookup(name)
	if !exists {
		panic(fmt.Sprintf("undefined symbol: %s", name))
	}
	return def
}

// Has 检查符号是否已注册
func (st *SymbolTable) Has(name string) bool {
	_, exists := st.Lookup(name)
	return exists
}

// Size 返回符号表中的符号数量
func (st *SymbolTable) Size() int {
	return len(st.symbols)
}

// AllSymbols 返回所有符号名称列表
func (st *SymbolTable) AllSymbols() []string {
	names := make([]string, 0, len(st.symbols))
	for name := range st.symbols {
		names = append(names, name)
	}
	return names
}

// ReferenceResolver 引用解析器
// 用于 Pass 3: Reference Resolution 阶段
// 遍历所有字段的类型引用，将类型名称解析为 Definition 引用
type ReferenceResolver struct {
	symbolTable *SymbolTable
	errors      []error // 收集解析过程中的错误
}

// NewReferenceResolver 创建新的引用解析器
func NewReferenceResolver(symbolTable *SymbolTable) *ReferenceResolver {
	return &ReferenceResolver{
		symbolTable: symbolTable,
		errors:      make([]error, 0),
	}
}

// ResolveField 解析单个字段的类型引用
// 递归处理 Map/XMap/Repeated 的嵌套类型
func (rr *ReferenceResolver) ResolveField(field *Field) error {
	return rr.resolveFieldType(&field.Type, field.Name)
}

// resolveFieldType 递归解析 FieldType 的引用
func (rr *ReferenceResolver) resolveFieldType(ft *FieldType, fieldName string) error {
	// 只处理 MMEObject 和 Message 类型
	if ft.Kind != FieldKindMMEObject && ft.Kind != FieldKindMessage {
		// 对于容器类型，递归处理其元素类型
		if ft.KeyType != nil {
			if err := rr.resolveFieldType(ft.KeyType, fieldName); err != nil {
				return err
			}
		}
		if ft.ValueType != nil {
			if err := rr.resolveFieldType(ft.ValueType, fieldName); err != nil {
				return err
			}
		}
		return nil
	}

	// 获取类型名称
	typeName := ft.TypeName
	if typeName == "" {
		typeName = ft.Name
	}
	if typeName == "" {
		return fmt.Errorf("field %s has empty type name", fieldName)
	}

	// 在符号表中查找定义
	def, exists := rr.symbolTable.Lookup(typeName)
	if !exists {
		err := fmt.Errorf("undefined type reference: field '%s' references unknown type '%s'", fieldName, typeName)
		rr.errors = append(rr.errors, err)
		return err
	}

	// 绑定引用
	ft.ResolvedType = def
	return nil
}

// ResolveFields 批量解析字段列表
func (rr *ReferenceResolver) ResolveFields(fields []*Field) []error {
	for _, field := range fields {
		if err := rr.ResolveField(field); err != nil {
			// 继续解析其他字段，收集所有错误
			continue
		}
	}
	return rr.errors
}

// ResolveDefinition 解析一个 Definition 中的所有字段引用
func (rr *ReferenceResolver) ResolveDefinition(def Definition) []error {
	fields := def.GetFields()
	if len(fields) == 0 {
		return nil
	}
	return rr.ResolveFields(fields)
}

// GetErrors 获取解析过程中收集的所有错误
func (rr *ReferenceResolver) GetErrors() []error {
	return rr.errors
}

// HasErrors 检查是否有错误
func (rr *ReferenceResolver) HasErrors() bool {
	return len(rr.errors) > 0
}

// ClearErrors 清空错误列表
func (rr *ReferenceResolver) ClearErrors() {
	rr.errors = make([]error, 0)
}
