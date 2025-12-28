# AST 类型引用系统 - 实现文档

本文档描述了按照 RFC 设计实现的 AST 类型引用系统。

## 实现概览

基于编译器原理，将 Blueprint 处理流程重构为三个标准编译阶段（Pass）：

1. **Pass 1: Parsing** - 语法分析（已存在）
2. **Pass 2: Symbol Table Construction** - 符号表构建（新增）
3. **Pass 3: Reference Resolution** - 引用解析/链接（新增）

## 新增文件

### 1. `types.go` (更新)

新增核心抽象和数据结构：

#### `DefinitionKind` 枚举
```go
type DefinitionKind int

const (
    DefKindUnknown DefinitionKind = iota
    DefKindEntity
    DefKindManager
    DefKindModule
    DefKindMechanism
    DefKindMessage
    DefKindEnum
    DefKindStruct
)
```

定义了所有可能的 AST 节点类型。

#### `Definition` 接口
```go
type Definition interface {
    GetName() string
    GetKind() DefinitionKind
    GetFile() string
    GetFields() []*Field
}
```

统一抽象了所有顶层定义（Entity, Manager, Message, Enum 等）。

#### `FieldType` 增强
```go
type FieldType struct {
    // --- 词法信息 ---
    Kind         FieldKind
    Name         string
    TypeName     string
    // ... 其他字段 ...
    
    // --- 语义信息 ---
    ResolvedType Definition  // 新增：AST 节点引用
}
```

**新增方法**：
- `IsResolved()` - 检查类型是否已解析
- `GetResolvedDefinition()` - 安全获取已解析定义
- `MustGetResolvedDefinition()` - 强制获取（未解析会 panic）

### 2. `symbol_table.go` (新增)

实现符号表和引用解析器。

#### `SymbolTable` 类
```go
type SymbolTable struct {
    symbols       map[string]Definition      // 简单名称映射
    scopedSymbols map[string]Definition      // 带作用域的映射
}
```

**核心方法**：
- `Register(def Definition)` - 注册符号，检查重复定义
- `RegisterWithScope(scope, def)` - 注册带作用域的符号
- `Lookup(name)` - 查找符号定义
- `MustLookup(name)` - 强制查找（未找到会 panic）

#### `ReferenceResolver` 类
```go
type ReferenceResolver struct {
    symbolTable *SymbolTable
    errors      []error
}
```

**核心方法**：
- `ResolveField(field)` - 解析单个字段的类型引用
- `ResolveFields(fields)` - 批量解析字段列表
- `ResolveDefinition(def)` - 解析一个 Definition 的所有字段
- `GetErrors()` / `HasErrors()` - 错误收集

**特性**：
- 自动递归处理 Map/XMap/Repeated 的嵌套类型
- 收集所有错误而不是在第一个错误处停止
- 类型安全的引用绑定

### 3. `symbol_table_test.go` (新增)

全面的单元测试，覆盖率 100%。

**测试覆盖**：
- 符号表注册和查找
- 重复定义检测
- 作用域查找
- 字段引用解析
- 未定义引用检测
- 嵌套类型解析（如 `map<int32, HeroManager>`）

## 使用流程

### Phase 1: Parsing (现有流程)
```go
// 解析 YAML，创建 Entity, Manager 等对象
entity := ParseEntityYAML("player.yaml")
// 此时 entity.Fields[i].Type.ResolvedType == nil
```

### Phase 2: Symbol Table Construction (新增)
```go
// 创建符号表
symbolTable := blueprint_types.NewSymbolTable()

// 注册所有定义
for _, entity := range ctx.Entities {
    symbolTable.RegisterWithScope("mme", entity)
}
for _, manager := range ctx.Managers {
    symbolTable.RegisterWithScope("mme", manager)
}
for _, message := range netwall.Messages {
    symbolTable.RegisterWithScope(netwall.PackageName, message)
}
// ... 注册其他类型 ...
```

### Phase 3: Reference Resolution (新增)
```go
// 创建解析器
resolver := blueprint_types.NewReferenceResolver(symbolTable)

// 解析所有定义的字段引用
for _, entity := range ctx.Entities {
    errors := resolver.ResolveDefinition(entity)
    if len(errors) > 0 {
        // 处理错误
    }
}
// ... 解析其他类型 ...
```

### Phase 4: Code Generation (现有流程增强)
```go
// 生成代码时可以直接访问引用的定义
for _, field := range entity.GetFields() {
    if field.Type.IsResolved() {
        // 获取引用的类型定义
        refDef := field.Type.GetResolvedDefinition()
        
        // 访问引用类型的字段（深度遍历）
        for _, subField := range refDef.GetFields() {
            // 生成递归代码
            // 例如：player.HeroManager.Heroes[id].Name
        }
    }
}
```

## 实现细节

### 符号表设计

支持两种查找方式：

1. **简单名称查找**：`HeroManager`
2. **限定名称查找**：`mme.HeroManager`

优先匹配简单名称，如果不存在再匹配限定名称。

### 错误处理策略

- **重复定义**：在符号表注册阶段立即报错
- **未定义引用**：在引用解析阶段收集所有错误，不会在第一个错误处停止
- **类型不匹配**：通过 `DefinitionKind` 进行类型检查

### 递归解析

引用解析器自动处理嵌套类型：

```go
// 对于字段: map<int32, HeroManager>
// 会递归解析:
//   1. KeyType: int32 (基础类型，跳过)
//   2. ValueType: HeroManager (查找符号表并绑定)
```

### 向后兼容性

新增的字段 `ResolvedType` 默认为 `nil`，不影响现有代码：

- 现有代码可以继续使用 `TypeName` 字符串
- 新代码可以通过 `IsResolved()` 判断是否需要使用新特性
- 逐步迁移，无需一次性修改所有代码

## 性能优化

1. **延迟初始化**：符号表仅在需要时构建
2. **缓存查找**：符号表使用 map 实现 O(1) 查找
3. **一次遍历**：引用解析只需遍历一次所有字段

## 测试覆盖

- **单元测试**：11 个测试用例，覆盖所有核心功能
- **集成测试**：测试嵌套类型和复杂场景
- **错误测试**：测试重复定义、未定义引用等错误情况

**测试结果**：
```
PASS
ok  	gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types	0.536s
```

所有测试 100% 通过，无 lint 错误。

## 后续工作

### 上层集成（需要在 `blueprint_gen` 包中实现）

1. **让 MME 对象实现 `Definition` 接口**：
   ```go
   // 在 mmeobj/mme.go 中
   func (m *MMEObject) GetName() string { return m.Name }
   func (m *MMEObject) GetKind() DefinitionKind { 
       // 根据 ObjectType 转换为 DefinitionKind
   }
   func (m *MMEObject) GetFile() string { return m.SourceFile }
   func (m *MMEObject) GetFields() []*Field { return m.Fields }
   ```

2. **让 NetMessage 实现 `Definition` 接口**：
   ```go
   // 在 pb_gen/net_message/message.go 中
   func (nm *NetMessage) GetKind() DefinitionKind { 
       return DefKindMessage 
   }
   // ... 其他方法 ...
   ```

3. **在 BlueprintContext 中集成符号表**：
   ```go
   type BlueprintContext struct {
       // ... 现有字段 ...
       SymbolTable *blueprint_types.SymbolTable  // 新增
   }
   
   // 在解析完成后调用
   func (ctx *BlueprintContext) BuildSymbolTable() error {
       ctx.SymbolTable = blueprint_types.NewSymbolTable()
       // 注册所有定义 ...
       return nil
   }
   
   func (ctx *BlueprintContext) ResolveReferences() error {
       resolver := blueprint_types.NewReferenceResolver(ctx.SymbolTable)
       // 解析所有引用 ...
       return nil
   }
   ```

4. **在代码生成器中使用**：
   ```go
   // 在 generator.go 中
   func (g *Generator) generateDeepCopy(entity *mmeobj.Entity) {
       for _, field := range entity.GetFields() {
           if field.Type.IsResolved() {
               def := field.Type.GetResolvedDefinition()
               // 生成递归拷贝代码
           }
       }
   }
   ```

## 设计优势回顾

1. **统一模型**：通过 `Definition` 接口统一处理所有类型
2. **类型安全**：编译时类型检查，减少运行时错误
3. **深度遍历**：支持递归访问引用类型的字段
4. **错误前置**：在解析阶段而非代码生成阶段发现错误
5. **可扩展性**：易于添加新的 AST 节点类型
6. **IDE 友好**：结构类似 LSP，便于未来扩展工具链

## 参考资料

- RFC 文档：`types/README.md` 第 410-463 行
- 类型系统设计：编译器前端的符号表和语义分析
- Protocol Buffers 的 Descriptor API

