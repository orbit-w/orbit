# Blueprint Generator 类型系统

Blueprint Generator 的类型系统和核心数据结构技术文档。

## 概览

本模块 (`blueprint_types`) 定义了 Orbit Blueprint Generator 的类型系统，用于解析和表示 YAML Blueprint 定义中的字段类型、MME 对象和数据结构。

**主要功能**：
- 解析复杂的类型字符串（包括嵌套的 `map`、`xmap`、`repeated`）
- 识别和分类 MME Object 类型（Entity、Manager、Module、Mechanism）
- 提供类型安全的字段定义和验证
- 支持 Protocol Buffers 兼容的类型系统

## 模块文件结构

```
types/
├── types.go          # 核心类型定义和类型解析器
├── field.go          # 字段定义和字段级别的辅助方法
├── label.go          # 字段标签（OPTIONAL, REQUIRED, REPEATED）
├── consts.go         # FieldKind 枚举常量和映射表
├── mme_obj/
│   └── mme.go       # MME Object 类型定义和识别逻辑
└── types_test.go     # 完整的单元测试套件
```

## 核心数据结构

### 1. `Field` - 字段定义

表示定义（Struct、Entity、Message）中的单个字段。

**核心属性**：
```go
type Field struct {
    Name    string        // 字段名称
    Type    FieldType     // 字段类型（支持递归类型）
    Number  int32         // 字段编号（Protocol Buffers 字段序号）
    Options FieldOption   // 字段选项（如 access=all/s/c）
    Comment string        // 字段注释
}
```

**关键方法**：
- `GetTypeName()` - 获取字段类型名称（如 `map`、`HeroManager`）
- `GetValueName()` - 获取 map/xmap/repeated 的值类型名称
- `IsMapField()` / `IsXMapField()` - 类型判断方法
- `IsMMEObjectType()` - 判断是否为 MME Object 类型

### 2. `FieldType` - 字段类型

描述字段的数据类型，**支持递归嵌套类型**。

**核心属性**：
```go
type FieldType struct {
    Kind         FieldKind   // 字段类型种类
    Label        FieldLabel  // 字段标签（OPTIONAL, REPEATED）
    Name         string      // 类型名称（去掉包名）
    TypeName     string      // 完整类型名称（带包名，如 "mme.HeroManager"）
    
    // 递归类型支持
    KeyType      *FieldType  // map/xmap 的 key 类型
    ValueType    *FieldType  // map/xmap/repeated 的值类型
    
    // 其他元数据
    Extendee     string      // 扩展字段的目标类型
    DefaultValue string      // 默认值
    JsonName     string      // JSON 名称
}
```

**支持的类型**：
- **基础类型**：`int32`, `int64`, `uint32`, `uint64`, `float`, `double`, `bool`, `string`, `bytes`, `fixed32`, `fixed64`
- **枚举类型**：`enum`
- **消息类型**：`message` - 普通 Protocol Buffers 消息
- **MME Object**：`Entity`, `Manager`, `Module`, `Mechanism` - 游戏引擎对象
- **容器类型**：`map`, `xmap`, `repeated` - 支持任意嵌套

**关键方法**：
- `GetKind()` - 获取类型种类
- `IsMapField()` / `IsXMapField()` / `IsMMEObjectType()` - 类型判断
- `IsMapValueMMEObject()` / `IsXMapValueMMEObject()` - 判断容器值类型
- `IsFieldBaseType()` - 判断是否为基础类型
- `String()` - 返回类型的字符串表示（用于调试）

### 3. `FieldKind` - 字段类型种类

枚举所有支持的字段类型。

**类型编号**（与 Protocol Buffers 兼容）：
```go
const (
    FieldKindUnknown   FieldKind = 0
    FieldKindDouble    FieldKind = 1   // double
    FieldKindFloat     FieldKind = 2   // float
    FieldKindInt64     FieldKind = 3   // int64
    FieldKindUInt64    FieldKind = 4   // uint64
    FieldKindInt32     FieldKind = 5   // int32
    FieldKindFixed64   FieldKind = 6   // fixed64
    FieldKindFixed32   FieldKind = 7   // fixed32
    FieldKindBool      FieldKind = 8   // bool
    FieldKindString    FieldKind = 9   // string
    FieldKindMessage   FieldKind = 11  // message
    FieldKindBytes     FieldKind = 12  // bytes
    FieldKindUInt32    FieldKind = 13  // uint32
    FieldKindEnum      FieldKind = 14  // enum
    FieldKindMMEObject FieldKind = 16  // MME 对象
    FieldKindMap       FieldKind = 18  // map
    FieldKindXMap      FieldKind = 19  // xmap（特殊 Map）
    FieldKindRepeated  FieldKind = 20  // repeated
)
```

**关键方法**：
- `IsBaseType()` - 是否为基础类型（包括 enum）
- `IsMMEObject()` / `IsMessage()` / `IsMap()` - 类型判断
- `String()` - 返回类型字符串表示

### 4. `FieldLabel` - 字段标签

定义字段的标签类型（参考 Protocol Buffers 设计）。

```go
const (
    FieldLabelUnknown  FieldLabel = 0  // 未知
    FieldLabelOptional FieldLabel = 1  // 可选字段（默认）
    FieldLabelRepeated FieldLabel = 2  // 重复字段（数组）
)
```

### 5. `FieldOption` - 字段选项

存储字段的扩展选项。

```go
type FieldOption struct {
    Access        string         // 访问控制：all/s/c（服务端/客户端）
    CustomOptions map[string]any // 自定义选项（可扩展）
}
```

## MME Object 类型系统

位于 `mme_obj/mme.go`，定义了 Orbit 游戏引擎的核心对象类型。

### MME Object 类型

```go
type ObjectType string

const (
    ObjectTypeEntity    ObjectType = "Entity"     // 实体（如 PlayerEntity, HeroEntity）
    ObjectTypeManager   ObjectType = "Manager"    // 管理器（如 HeroManager, ItemManager）
    ObjectTypeModule    ObjectType = "Module"     // 模块（如 HeroModule, EquipModule）
    ObjectTypeMechanism ObjectType = "Mechanism"  // 机制（如 LevelUpMechanism）
)
```

### 识别规则

**命名模式**：MME Object 类型通过后缀识别
- `PlayerEntity`, `HeroEntity` → Entity
- `HeroManager`, `ItemManager` → Manager
- `HeroModule`, `EquipModule` → Module
- `LevelUpMechanism`, `HeroMechanism` → Mechanism

**支持限定名称**：
- 简单名称：`HeroManager`
- 带包名：`mme.HeroManager`, `core.PlayerEntity`

**关键函数**：
- `IsMMEObjectType(typeName string) bool` - 判断类型名是否为 MME Object
- `ParseMMEObjectType(typeName string) (ObjectType, error)` - 解析 MME Object 类型

## 类型解析器

### `ParseTypeString` - 递归类型解析

核心函数，用于解析类型字符串，**支持任意嵌套的复杂类型**。

**签名**：
```go
func ParseTypeString(typeStr string) (*FieldType, error)
```

**支持的类型字符串**：

| 类型分类 | 示例 | 说明 |
|---------|------|------|
| **基础类型** | `int32`, `string`, `bool` | 标量类型 |
| **枚举** | `enum` | 枚举类型 |
| **消息** | `Book`, `core.Book` | Protocol Buffers 消息 |
| **MME Object** | `HeroManager`, `mme.PlayerEntity` | 游戏引擎对象 |
| **Map** | `map<int32, string>` | 键值对映射 |
| **XMap** | `xmap<int64, HeroManager>` | 特殊 Map（支持扩展特性） |
| **Repeated** | `repeated int32`, `repeated Book` | 数组/列表 |
| **嵌套 Map** | `map<int32, map<string, int64>>` | 多层嵌套 |
| **Repeated Map** | `repeated map<string, HeroManager>` | 数组包含 Map |
| **Map + Repeated** | `map<int32, repeated HeroManager>` | Map 的值为数组 |

**解析流程**：
1. **去除空格** - 处理首尾空格
2. **识别容器类型** - 检查 `repeated`, `xmap`, `map` 前缀
3. **递归解析** - 对 key/value 类型递归调用 `ParseTypeString`
4. **类型推断** - 识别基础类型 / MME Object / Message

**解析示例**：
```go
// 基础类型
ft, _ := ParseTypeString("int32")
// ft.Kind = FieldKindInt32, ft.Label = FieldLabelOptional

// MME Object
ft, _ := ParseTypeString("mme.HeroManager")
// ft.Kind = FieldKindMMEObject, ft.TypeName = "mme.HeroManager"

// 嵌套 Map
ft, _ := ParseTypeString("map<int32, map<string, HeroManager>>")
// ft.Kind = FieldKindMap
// ft.KeyType.Kind = FieldKindInt32
// ft.ValueType.Kind = FieldKindMap
// ft.ValueType.ValueType.Kind = FieldKindMMEObject
```

### 辅助解析函数

- `parseRepeatedType()` - 解析 `repeated T` 类型
- `parseMapTypeRecursive()` - 递归解析 `map<K,V>` 和 `xmap<K,V>`
- `parseBaseOrStructType()` - 解析基础类型或自定义类型
- `splitMapTypes()` - 分割 Map 的 Key 和 Value（处理嵌套 `<>`）
- `extractTypeName()` - 从完整类型名中提取简短名称
- `isMMEObjectType()` - 判断是否为 MME Object

## 常量和映射表

`consts.go` 定义了类型系统的常量和映射表。

**字符串常量**（用于 YAML/JSON 序列化）：
```go
FieldKindInt32String     = "int32"
FieldKindStringString    = "string"
FieldKindMMEObjectString = "MMEObject"
FieldKindMapString       = "map"
...
```

**映射表**：
- `FieldKindStringMap` - `FieldKind → string` 映射
- `FieldKindStringToKindMap` - `string → FieldKind` 映射（用于反向查找）

## 测试覆盖

`types_test.go` 提供了全面的单元测试（600+ 行）。

**测试分类**：
- **基础类型测试** - 所有标量类型
- **MME Object 测试** - Entity/Manager/Module/Mechanism 识别
- **消息类型测试** - 普通消息和带包名的消息
- **Map 类型测试** - 简单 Map、嵌套 Map、XMap
- **Repeated 测试** - 数组和复杂嵌套
- **边界情况** - 空字符串、格式错误、括号不匹配
- **性能测试** - Benchmark 测试（基础类型、Map、嵌套）

## 设计特点

### 1. 递归类型支持
通过 `KeyType` 和 `ValueType` 指针，`FieldType` 可以表示任意深度的嵌套类型：
```
map<int32, map<string, repeated HeroManager>>
     ^          ^              ^
     |          |              |
  KeyType    KeyType       ValueType
             (递归)        (递归)
```

### 2. 类型安全
- 使用强类型枚举（`FieldKind`、`FieldLabel`）
- 提供类型判断方法（`IsMapField()`, `IsMMEObjectType()`）
- Panic 保护（如 `IsXMapValueMMEObject()` 会检查前提条件）

### 3. Protocol Buffers 兼容
- `FieldKind` 编号与 `descriptorpb.FieldDescriptorProto.Type` 对齐
- 支持 `FieldLabel`（OPTIONAL, REQUIRED, REPEATED）
- 字段编号（`Number`）与 protobuf 字段序号一致

### 4. 扩展性
- `FieldOption.CustomOptions` 支持自定义扩展
- MME Object 类型通过命名约定识别，易于扩展新类型
- 类型解析器可处理未知类型（默认为 `FieldKindMessage`）

## 使用示例

### 解析字段类型
```go
import "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types"

// 解析简单类型
ft, err := blueprint_types.ParseTypeString("int32")
if err != nil {
    log.Fatal(err)
}
fmt.Println(ft.Kind)  // FieldKindInt32

// 解析复杂嵌套类型
ft, err = blueprint_types.ParseTypeString("xmap<int64, map<string, HeroManager>>")
if err != nil {
    log.Fatal(err)
}
fmt.Println(ft.Kind)                        // FieldKindXMap
fmt.Println(ft.ValueType.Kind)             // FieldKindMap
fmt.Println(ft.ValueType.ValueType.Kind)   // FieldKindMMEObject
```

### 构建字段定义
```go
field := &blueprint_types.Field{
    Name:    "heroes",
    Number:  1,
    Comment: "玩家拥有的英雄列表",
}

// 解析类型字符串
field.Type, _ = blueprint_types.ParseTypeString("map<int32, HeroManager>")

// 检查类型
if field.IsMapField() {
    fmt.Println("这是一个 Map 字段")
    if field.Type.IsMapValueMMEObject() {
        fmt.Println("Map 的值是 MME Object")
    }
}
```

### 识别 MME Object
```go
import "gitee.com/orbit-w/orbit/tools/gotools/gen/cmd/blueprint_gen/types/mme_obj"

// 判断类型名
if mmeobject.IsMMEObjectType("HeroManager") {
    objType, _ := mmeobject.ParseMMEObjectType("HeroManager")
    fmt.Println(objType.IsManager())  // true
}

// 支持带包名的类型
if mmeobject.IsMMEObjectType("mme.PlayerEntity") {
    objType, _ := mmeobject.ParseMMEObjectType("mme.PlayerEntity")
    fmt.Println(objType.IsEntity())  // true
}
```

## 相关资源

- **Blueprint YAML 定义**：`protocol/blueprint/mme/*.yaml`
- **代码生成工具**：`orbit/tools/gotools/gen/cmd/blueprint_gen`
- **生成的 Protobuf**：`orbit/pkg/proto/mme/`

## 注意事项

1. **类型字符串格式**：空格会被自动处理，但括号必须匹配
2. **MME Object 识别**：依赖命名约定（后缀为 Entity/Manager/Module/Mechanism）
3. **递归解析**：支持任意嵌套深度，但过深的嵌套可能影响性能
4. **类型推断优先级**：基础类型 → MME Object → Message
5. **XMap vs Map**：XMap 是特殊的 Map 类型，支持额外的引擎特性

## 特性规划：基于 AST 的类型引用系统 (RFC)

### 1. 设计理念：从“结构体”到“语法树”

当前系统仅将 YAML 解析为孤立的结构体（`BlueprintContext`），缺乏统一的语义模型。本方案引入 **AST（抽象语法树）** 概念，将所有的定义（Entity, Message）视为 **AST 节点**，将类型引用视为 **符号链接**。

目标是建立一个完整的 **Symbol Resolution (符号解析)** 机制，使得 `FieldType` 不仅包含类型的“名称”，还包含指向类型“定义”的直接引用。

### 2. 核心抽象：Definition (符号定义)

在 `types` 包中引入统一的 `Definition` 接口，用于抽象描述任何可被引用的用户自定义类型（User-Defined Types）。

```go
// DefinitionKind 定义节点的类型
type DefinitionKind int

const (
    DefKindEntity DefinitionKind = iota
    DefKindManager
    DefKindMessage
    DefKindEnum
    // ...
)

// Definition (符号定义接口)
// 所有的顶层定义（Entity, Manager, Message, Struct, Enum）都必须实现此接口
type Definition interface {
    // 基础元数据
    GetName() string           // 符号名称 (Symbol Name)
    GetKind() DefinitionKind   // 定义种类
    GetFile() string          // 定义所在源文件 (Source File)

    // 结构信息 (用于遍历 AST)
    GetFields() []*Field       // 获取成员字段列表 (Enum 返回空或 nil)
}
```

### 3. AST 节点增强：FieldType

将 `FieldType` 升级为具有语义链接能力的 AST 节点。它既包含词法信息（名称），也包含语义信息（引用）。

```go
type FieldType struct {
    // --- 词法信息 (Lexical Info) ---
    // 从源文件解析出的原始信息
    Kind         FieldKind
    Name         string      // 类型名称标识符 (Identifier)
    TypeName     string      // 完整限定名
    
    // --- 语义信息 (Semantic Info) ---
    // 符号解析 (Symbol Resolution) 后的结果
    // 指向该类型对应的 AST Definition 节点
    // 类似于编译器前端的 Symbol Table 查找结果
    ResolvedType Definition 
}
```

### 4. 编译流水线 (Compiler Pipeline)

将 Blueprint 的处理流程重构为标准的编译器 Pass：

#### Pass 1: Parsing (语法分析)
- **输入**：YAML 文件, Proto 文件
- **输出**：未链接的 AST 节点树
- **行为**：
    - 将源文件解析为 `Entity`, `Manager`, `NetMessage` 等对象。
    - 这些对象实现 `Definition` 接口。
    - 此时 `FieldType.ResolvedType` 均为 `nil`。

#### Pass 2: Symbol Table Construction (符号表构建)
- **输入**：AST 节点树
- **输出**：全局符号表 (`SymbolTable`)
- **行为**：
    - 遍历所有 AST 根节点。
    - 将 `Definition` 注册到全局符号表 `map[string]Definition`。
    - **语义检查**：检测重复定义 (Duplicate Declaration)。

#### Pass 3: Reference Resolution (引用解析/链接)
- **输入**：AST 节点树 + 符号表
- **输出**：完整的语义图 (Semantic Graph)
- **行为**：
    - 遍历所有 AST 中的 `Field`。
    - 对于 `FieldKindMMEObject` 或 `FieldKindMessage` 类型的字段：
        1. 获取其 `TypeName` (Identifier)。
        2. 在符号表中查找对应的 `Definition`。
        3. 将找到的 `Definition` 绑定到 `FieldType.ResolvedType`。
    - **语义检查**：检测未定义引用 (Undefined Reference)。

### 5. 优势

1.  **统一模型**：代码生成器通过 `Definition` 接口统一操作 Entity 和 Message，无需编写两套逻辑。
2.  **AST 遍历能力**：
    - 支持 Visitor 模式。
    - 可以轻松实现 **深度遍历**（例如：生成 DeepCopy 代码时，递归遍历所有引用的子对象）。
3.  **提前报错**：在解析阶段而非生成代码阶段就能发现“类型未定义”或“字段缺失”等错误。
4.  **IDE 友好**：这种结构非常接近 Language Server Protocol (LSP) 需要的数据结构，未来容易扩展出 IDE 插件支持（跳转到定义）。
