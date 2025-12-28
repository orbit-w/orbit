# AST 类型引用系统 - 使用示例

本文档提供了使用新的 AST 类型引用系统的实际代码示例。

## 场景 1：基本使用流程

### Step 1: 创建符号表并注册定义

```go
package main

import (
    "fmt"
    blueprint_types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"
)

func main() {
    // 创建符号表
    symbolTable := blueprint_types.NewSymbolTable()
    
    // 假设我们已经解析了 YAML，得到了 Entity 和 Manager 对象
    // 这些对象需要实现 Definition 接口
    
    // 注册 Entity
    playerEntity := getPlayerEntity() // 返回 *mmeobj.Entity
    if err := symbolTable.RegisterWithScope("mme", playerEntity); err != nil {
        panic(fmt.Sprintf("Failed to register PlayerEntity: %v", err))
    }
    
    // 注册 Manager
    heroManager := getHeroManager() // 返回 *mmeobj.Manager
    if err := symbolTable.RegisterWithScope("mme", heroManager); err != nil {
        panic(fmt.Sprintf("Failed to register HeroManager: %v", err))
    }
    
    fmt.Printf("Symbol table size: %d\n", symbolTable.Size())
    fmt.Printf("Registered symbols: %v\n", symbolTable.AllSymbols())
}
```

### Step 2: 解析类型引用

```go
func resolveReferences(symbolTable *blueprint_types.SymbolTable, entity *mmeobj.Entity) error {
    // 创建引用解析器
    resolver := blueprint_types.NewReferenceResolver(symbolTable)
    
    // 解析 Entity 的所有字段引用
    errors := resolver.ResolveDefinition(entity)
    
    if resolver.HasErrors() {
        for _, err := range errors {
            fmt.Printf("Resolution error: %v\n", err)
        }
        return fmt.Errorf("failed to resolve %d references", len(errors))
    }
    
    fmt.Printf("Successfully resolved all references for %s\n", entity.GetName())
    return nil
}
```

### Step 3: 使用已解析的引用

```go
func generateCode(entity *mmeobj.Entity) {
    for _, field := range entity.GetFields() {
        fmt.Printf("Field: %s, Type: %s\n", field.Name, field.Type.TypeName)
        
        // 检查类型是否已解析
        if field.Type.IsResolved() {
            // 获取引用的定义
            refDef := field.Type.GetResolvedDefinition()
            
            fmt.Printf("  -> Resolved to: %s (kind: %s)\n", 
                refDef.GetName(), refDef.GetKind())
            
            // 遍历引用类型的字段（深度访问）
            for _, subField := range refDef.GetFields() {
                fmt.Printf("     - %s: %s\n", subField.Name, subField.Type.TypeName)
            }
        }
    }
}
```

## 场景 2：处理嵌套类型

### 解析 Map<int32, HeroManager>

```go
func resolveMapField() {
    symbolTable := blueprint_types.NewSymbolTable()
    
    // 注册 HeroManager
    heroManager := getHeroManager()
    symbolTable.Register(heroManager)
    
    // 创建字段: map<int32, HeroManager>
    field := &blueprint_types.Field{
        Name:   "heroMap",
        Number: 1,
        Type: blueprint_types.FieldType{
            Kind: blueprint_types.FieldKindMap,
            KeyType: &blueprint_types.FieldType{
                Kind: blueprint_types.FieldKindInt32,
                Name: "int32",
            },
            ValueType: &blueprint_types.FieldType{
                Kind:     blueprint_types.FieldKindMMEObject,
                TypeName: "HeroManager",
            },
        },
    }
    
    // 解析
    resolver := blueprint_types.NewReferenceResolver(symbolTable)
    if err := resolver.ResolveField(field); err != nil {
        panic(err)
    }
    
    // 访问解析后的嵌套类型
    if field.Type.ValueType.IsResolved() {
        heroMgrDef := field.Type.ValueType.GetResolvedDefinition()
        fmt.Printf("Map value type resolved to: %s\n", heroMgrDef.GetName())
        
        // 可以访问 HeroManager 的字段
        for _, heroField := range heroMgrDef.GetFields() {
            fmt.Printf("  Hero field: %s\n", heroField.Name)
        }
    }
}
```

## 场景 3：错误处理

### 处理未定义的类型引用

```go
func handleUndefinedReference() {
    symbolTable := blueprint_types.NewSymbolTable()
    // 故意不注册 ItemManager
    
    field := &blueprint_types.Field{
        Name:   "items",
        Number: 1,
        Type: blueprint_types.FieldType{
            Kind:     blueprint_types.FieldKindMessage,
            TypeName: "ItemManager", // 未定义
        },
    }
    
    resolver := blueprint_types.NewReferenceResolver(symbolTable)
    err := resolver.ResolveField(field)
    
    if err != nil {
        // 错误示例: "undefined type reference: field 'items' references unknown type 'ItemManager'"
        fmt.Printf("Error: %v\n", err)
        
        // 检查字段是否已解析
        if !field.Type.IsResolved() {
            fmt.Println("Field type is not resolved")
        }
    }
}
```

### 处理重复定义

```go
func handleDuplicateDefinition() {
    symbolTable := blueprint_types.NewSymbolTable()
    
    entity1 := newMockDefinition("PlayerEntity", blueprint_types.DefKindEntity, "player1.yaml")
    entity2 := newMockDefinition("PlayerEntity", blueprint_types.DefKindEntity, "player2.yaml")
    
    // 第一次注册成功
    if err := symbolTable.Register(entity1); err != nil {
        panic(err)
    }
    
    // 第二次注册失败
    if err := symbolTable.Register(entity2); err != nil {
        // 错误示例: "duplicate definition: symbol 'PlayerEntity' already defined in player1.yaml (kind: Entity)"
        fmt.Printf("Expected error: %v\n", err)
    }
}
```

## 场景 4：代码生成中的深度遍历

### 生成 DeepCopy 方法

```go
func generateDeepCopy(entity *mmeobj.Entity) string {
    var code strings.Builder
    
    code.WriteString(fmt.Sprintf("func (e *%s) DeepCopy() *%s {\n", 
        entity.GetName(), entity.GetName()))
    code.WriteString(fmt.Sprintf("    copy := &%s{}\n", entity.GetName()))
    
    for _, field := range entity.GetFields() {
        if field.Type.IsResolved() {
            // 如果字段引用了另一个对象，生成递归拷贝代码
            refDef := field.Type.GetResolvedDefinition()
            
            code.WriteString(fmt.Sprintf("    // Copy %s (type: %s)\n", 
                field.Name, refDef.GetName()))
            
            if refDef.GetKind() == blueprint_types.DefKindManager {
                // 对于 Manager，可能需要深度拷贝其中的 Module
                code.WriteString(fmt.Sprintf("    if e.%s != nil {\n", field.Name))
                code.WriteString(fmt.Sprintf("        copy.%s = e.%s.DeepCopy()\n", 
                    field.Name, field.Name))
                code.WriteString("    }\n")
            }
        } else {
            // 基础类型，直接拷贝
            code.WriteString(fmt.Sprintf("    copy.%s = e.%s\n", 
                field.Name, field.Name))
        }
    }
    
    code.WriteString("    return copy\n")
    code.WriteString("}\n")
    
    return code.String()
}
```

### 生成 Validate 方法

```go
func generateValidate(entity *mmeobj.Entity) string {
    var code strings.Builder
    
    code.WriteString(fmt.Sprintf("func (e *%s) Validate() error {\n", 
        entity.GetName()))
    
    for _, field := range entity.GetFields() {
        if field.Type.IsResolved() {
            refDef := field.Type.GetResolvedDefinition()
            
            // 生成递归验证
            code.WriteString(fmt.Sprintf("    // Validate %s\n", field.Name))
            code.WriteString(fmt.Sprintf("    if e.%s != nil {\n", field.Name))
            code.WriteString(fmt.Sprintf("        if err := e.%s.Validate(); err != nil {\n", 
                field.Name))
            code.WriteString(fmt.Sprintf("            return fmt.Errorf(\"%s.%s: %%w\", err)\n", 
                entity.GetName(), field.Name))
            code.WriteString("        }\n")
            code.WriteString("    }\n")
        }
    }
    
    code.WriteString("    return nil\n")
    code.WriteString("}\n")
    
    return code.String()
}
```

## 场景 5：符号表查询

### 按名称查找定义

```go
func lookupSymbols(symbolTable *blueprint_types.SymbolTable) {
    // 简单名称查找
    if def, exists := symbolTable.Lookup("HeroManager"); exists {
        fmt.Printf("Found: %s (kind: %s, file: %s)\n", 
            def.GetName(), def.GetKind(), def.GetFile())
    }
    
    // 限定名称查找
    if def, exists := symbolTable.Lookup("mme.HeroManager"); exists {
        fmt.Printf("Found with scope: %s\n", def.GetName())
    }
    
    // 检查是否存在
    if symbolTable.Has("ItemManager") {
        fmt.Println("ItemManager is registered")
    } else {
        fmt.Println("ItemManager is not registered")
    }
}
```

### 强制查找（MustLookup）

```go
func mustLookupExample(symbolTable *blueprint_types.SymbolTable) {
    // 如果不存在会 panic
    defer func() {
        if r := recover(); r != nil {
            fmt.Printf("Caught panic: %v\n", r)
        }
    }()
    
    // 这会 panic，因为 "UnknownType" 不存在
    def := symbolTable.MustLookup("UnknownType")
    fmt.Println(def.GetName()) // 不会执行到这里
}
```

## 场景 6：批量解析

### 解析整个 BlueprintContext

```go
func resolveBlueprintContext(ctx *BlueprintContext) error {
    // 1. 构建符号表
    symbolTable := blueprint_types.NewSymbolTable()
    
    // 注册所有 Entity
    for _, entity := range ctx.Entities {
        if err := symbolTable.RegisterWithScope("mme", entity); err != nil {
            return err
        }
    }
    
    // 注册所有 Manager
    for _, manager := range ctx.Managers {
        if err := symbolTable.RegisterWithScope("mme", manager); err != nil {
            return err
        }
    }
    
    // 注册所有 Module
    for _, module := range ctx.Modules {
        if err := symbolTable.RegisterWithScope("mme", module); err != nil {
            return err
        }
    }
    
    // 注册 NetWall 消息
    for _, netwall := range ctx.NetWalls {
        for _, message := range netwall.GetAllMessages() {
            if err := symbolTable.RegisterWithScope(netwall.PackageName, message); err != nil {
                return err
            }
        }
    }
    
    // 2. 创建解析器
    resolver := blueprint_types.NewReferenceResolver(symbolTable)
    
    // 3. 解析所有引用
    allErrors := make([]error, 0)
    
    for _, entity := range ctx.Entities {
        errors := resolver.ResolveDefinition(entity)
        allErrors = append(allErrors, errors...)
    }
    
    for _, manager := range ctx.Managers {
        errors := resolver.ResolveDefinition(manager)
        allErrors = append(allErrors, errors...)
    }
    
    // 检查错误
    if len(allErrors) > 0 {
        fmt.Printf("Found %d resolution errors:\n", len(allErrors))
        for _, err := range allErrors {
            fmt.Printf("  - %v\n", err)
        }
        return fmt.Errorf("reference resolution failed with %d errors", len(allErrors))
    }
    
    fmt.Println("All references resolved successfully!")
    return nil
}
```

## 场景 7：实现 Definition 接口

### 为 MMEObject 实现接口

```go
// 在 mmeobj/mme.go 中添加

import blueprint_types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"

// GetKind 实现 Definition 接口
func (m *MMEObject) GetKind() blueprint_types.DefinitionKind {
    switch m.ObjectType {
    case mmeobject.ObjectTypeEntity:
        return blueprint_types.DefKindEntity
    case mmeobject.ObjectTypeManager:
        return blueprint_types.DefKindManager
    case mmeobject.ObjectTypeModule:
        return blueprint_types.DefKindModule
    case mmeobject.ObjectTypeMechanism:
        return blueprint_types.DefKindMechanism
    default:
        return blueprint_types.DefKindUnknown
    }
}

// GetFile 实现 Definition 接口
func (m *MMEObject) GetFile() string {
    // 假设我们在解析时存储了源文件路径
    if file, ok := m.Settings["source_file"].(string); ok {
        return file
    }
    return ""
}

// MMEObject 已经有 GetName() 和 GetFields() 方法，无需额外实现
```

### 为 NetMessage 实现接口

```go
// 在 pb_gen/net_message/message.go 中添加

import blueprint_types "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"

// GetKind 实现 Definition 接口
func (nm *NetMessage) GetKind() blueprint_types.DefinitionKind {
    return blueprint_types.DefKindMessage
}

// GetFile 实现 Definition 接口
func (nm *NetMessage) GetFile() string {
    // 返回 NetWall 文件路径
    return nm.SourceFile
}

// NetMessage 需要实现 GetFields() 方法
func (nm *NetMessage) GetFields() []*blueprint_types.Field {
    return nm.Fields
}
```

## 最佳实践

### 1. 错误处理

```go
// ✅ 好的做法：收集所有错误后统一处理
errors := resolver.ResolveFields(fields)
if len(errors) > 0 {
    for _, err := range errors {
        log.Printf("Error: %v", err)
    }
    return fmt.Errorf("found %d errors", len(errors))
}

// ❌ 不好的做法：遇到第一个错误就返回
for _, field := range fields {
    if err := resolver.ResolveField(field); err != nil {
        return err // 用户看不到其他错误
    }
}
```

### 2. 类型安全访问

```go
// ✅ 好的做法：检查后再访问
if field.Type.IsResolved() {
    def := field.Type.GetResolvedDefinition()
    // 使用 def ...
}

// ⚠️ 风险做法：适用于确定已解析的场景
def := field.Type.MustGetResolvedDefinition() // 未解析会 panic
```

### 3. 符号表构建

```go
// ✅ 好的做法：按类型批量注册
for _, entity := range entities {
    symbolTable.RegisterWithScope("mme", entity)
}

// ✅ 也可以：分层注册
for _, entity := range entities {
    symbolTable.Register(entity) // 简单名称
}
for _, message := range messages {
    symbolTable.RegisterWithScope(packageName, message) // 带作用域
}
```

## 调试技巧

### 打印符号表内容

```go
func debugSymbolTable(st *blueprint_types.SymbolTable) {
    fmt.Printf("Symbol Table (%d symbols):\n", st.Size())
    for _, name := range st.AllSymbols() {
        def, _ := st.Lookup(name)
        fmt.Printf("  - %s (%s) from %s\n", 
            name, def.GetKind(), def.GetFile())
    }
}
```

### 打印解析状态

```go
func debugFieldResolution(field *blueprint_types.Field) {
    fmt.Printf("Field: %s\n", field.Name)
    fmt.Printf("  Type: %s\n", field.Type.TypeName)
    fmt.Printf("  Kind: %s\n", field.Type.Kind)
    fmt.Printf("  IsResolved: %v\n", field.Type.IsResolved())
    
    if field.Type.IsResolved() {
        def := field.Type.GetResolvedDefinition()
        fmt.Printf("  ResolvedTo: %s (%s)\n", def.GetName(), def.GetKind())
    }
}
```

## 性能考虑

### 避免重复解析

```go
// ✅ 好的做法：缓存解析结果
var resolvedOnce sync.Once
var globalResolver *blueprint_types.ReferenceResolver

func getResolver() *blueprint_types.ReferenceResolver {
    resolvedOnce.Do(func() {
        globalResolver = blueprint_types.NewReferenceResolver(globalSymbolTable)
        // 解析所有定义
    })
    return globalResolver
}
```

### 批量操作

```go
// ✅ 好的做法：批量解析
errors := resolver.ResolveFields(allFields)

// ❌ 低效做法：逐个解析
for _, field := range allFields {
    resolver.ResolveField(field)
}
```

