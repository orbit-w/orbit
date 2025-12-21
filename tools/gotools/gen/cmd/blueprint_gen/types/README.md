# Blueprint Generator 类型系统

`blueprint_gen` 类型系统和上下文结构的技术文档。

## 概览

本模块定义了 Orbit Blueprint Generator 使用的数据结构和类型系统。它主要由两部分组成：

1.  **底层类型系统** (`types/` 包)：处理 Protocol Buffers 和 MME 对象的字段定义、类型解析和验证。
2.  **生成上下文** (`blueprint_gen` 包中的 `types.go`)：定义生成过程的高层状态，包括实体、模块和网络定义的解析后 AST（抽象语法树）。

## 1. 底层类型系统 (`blueprint_types`)

位于 `orbit/tools/gotools/gen/cmd/blueprint_gen/types/`。

### 核心结构

#### `Field` (字段)
表示定义（Struct, Entity, Message）中的单个字段。
- **属性**：名称 (Name)、类型 (Type)、编号 (Number/ID)、选项 (Options)、注释 (Comments)。
- **元数据**：包含 `FieldMetadata` 用于深度检查（例如，解析 MME 对象引用）。

#### `FieldType` (字段类型)
描述字段的数据类型，支持递归类型。
- **支持类型**：
    - **基础类型**：`int32`, `string`, `bool` 等。
    - **复杂类型**：`map`, `xmap`, `repeated`。
    - **MME 对象**：对 `Entity` (实体), `Manager` (管理器), `Module` (模块) 等的引用。
- **解析**：`ParseTypeString()` 处理复杂的嵌套类型，如 `xmap<int32, map<string, HeroEntity>>`。

#### `FieldKind` (字段种类)
所有支持种类的枚举：
- **基础**：`FieldKindInt32`, `FieldKindString` 等。
- **容器**：`FieldKindMap`, `FieldKindRepeated`。
- **特殊**：`FieldKindMMEObject` (用于游戏引擎对象), `FieldKindXMap` (专用 Map)。

### 主要特性

- **类型推断**：自动推断类型是原始类型还是对已生成 MME 对象的引用。
- **深度解析**：正确处理嵌套泛型（例如 `map<K, V>`）。
- **元数据关联**：将字段与其更广泛的上下文关联（例如，识别字段是 `PlayerEntity` 引用）。

## 2. 生成上下文 (`blueprint_gen`)

定义在 `orbit/tools/gotools/gen/cmd/blueprint_gen/types.go`。

该层将解析后的数据聚合为准备进行代码生成的语义模型。

### `BlueprintContext`
保存项目完整解析状态的根对象。
- **Entities**：游戏实体列表 (`Entity`)。
- **Managers/Modules**：游戏逻辑容器。
- **NetWalls**：解析后的网络协议定义 (`.netwall` / `.yaml`)。
- **Namespace Management**：将类型名称映射到其 proto 包和命名空间。

### `NetWallFile`
表示解析后的 Network Wall 定义文件。
- 包含：`Requests` (请求), `Notifies` (通知), `DataStructs` (数据结构), `Enums` (枚举)。
- **功能**：桥接网络协议定义 (YAML/Proto) 和 Go 代码生成。

### `Enum` & `EnumValue`
表示从源文件解析的枚举类型，保留注释和选项以供生成使用。

## 用法示例

这些类型主要由 `parser` (解析器) 用于构建 AST，并由 `generator` (生成器) 用于生成最终的 Go 代码。

```go
// 示例：解析类型字符串
ft, err := blueprint_types.ParseTypeString("xmap<int32, PlayerEntity>")
if ft.IsXMapField() {
    keyType := ft.KeyType // int32
    valType := ft.ValueType // PlayerEntity (FieldKindMMEObject)
}
```
