# Blueprint Gen — 详细参考文档

## 一、核心数据结构

### BlueprintContext（编译上下文）

位于 `context.go`，贯穿所有编译阶段的全局数据容器。

```go
type BlueprintContext struct {
    HeadFile            *HeadFileConfig          // headfile.yaml 内容
    Entities            []*mmeobj.Entity
    Managers            []*mmeobj.Manager
    Modules             []*mmeobj.Module
    Mechanisms          []*mmeobj.Mechanism
    NetWalls            []*NetWallFile
    Enums               []*Enum
    ObjectTypeMap       map[string]mmeobject.ObjectType  // 名称→对象类型 映射
    NameSpaces          map[string]string  // 名称→包名 映射（懒加载）
    SymbolTable         *types.SymbolTable // Pass 2 后填充
}
```

**关键方法**：
- `BuildSymbolTable() error` — Pass 2，注册所有符号
- `ResolveReferences() error` — Pass 3，绑定 TypeName → Definition
- `AddEntity/Manager/Module/Mechanism/NetWallFile(...)` — 注册对象
- `GetObjectType(name) (ObjectType, bool)` — 查询对象类型
- `GetNameSpace() map[string]string` — 获取名称→包名映射（用于 import 生成）

---

### FieldType（字段类型描述）

位于 `types/types.go`，描述一个字段的完整类型信息。

```go
type FieldType struct {
    // --- 词法信息 ---
    Kind         FieldKind   // 类型种类枚举
    Label        FieldLabel  // OPTIONAL / REQUIRED / REPEATED
    Name         string      // 不带包名的类型名，如 "HeroManager"
    TypeName     string      // 完整类型名，如 "mme.HeroManager"
    KeyType      *FieldType  // map/xmap 的 Key 类型
    ValueType    *FieldType  // map/xmap 的 Value 类型 / repeated 元素类型

    // --- 语义信息（Pass 3 后填充）---
    ResolvedType Definition  // 指向 AST Definition 节点
}
```

**常用检查方法**：
```go
ft.IsResolved()                  // 是否已完成类型解析（Pass 3 后才为 true）
ft.GetResolvedDefinition()       // 安全获取，未解析返回 nil
ft.MustGetResolvedDefinition()   // 强制获取，未解析 panic（代码生成器中使用）
ft.IsMMEObjectType()             // 是否是 MME 对象（Entity/Manager/Module/Mechanism）
ft.IsMessage()                   // 是否是 NetWall 消息或 DataStruct
ft.IsMapField()                  // 是否是 map<>
ft.IsXMapField()                 // 是否是 xmap<>
ft.IsFieldBaseType()             // 是否是基础类型（int32/string/bool 等）
ft.IsMapValueMMEObject()         // map 的 Value 是否是 MME 对象
ft.IsXMapValueMMEObject()        // xmap 的 Value 是否是 MME 对象
```

---

### FieldKind 枚举

```go
FieldKindInt32     FieldKind = 5
FieldKindInt64     FieldKind = 3
FieldKindUInt32    FieldKind = 13
FieldKindUInt64    FieldKind = 4
FieldKindFloat     FieldKind = 2
FieldKindDouble    FieldKind = 1
FieldKindBool      FieldKind = 8
FieldKindString    FieldKind = 9
FieldKindBytes     FieldKind = 12
FieldKindEnum      FieldKind = 14
FieldKindMessage   FieldKind = 11  // NetWall 消息 / Common DataStruct
FieldKindMMEObject FieldKind = 16  // Entity/Manager/Module/Mechanism
FieldKindMap       FieldKind = 18
FieldKindXMap      FieldKind = 19
FieldKindRepeated  FieldKind = 20
```

---

### Definition 接口

位于 `types/definition.go`，所有顶层定义（MME 对象、消息、DataStruct）均需实现。

```go
type Definition interface {
    GetName()   string
    GetKind()   DefinitionKind
    GetFile()   string          // 源文件路径（SourceFile 字段）
    GetFields() []*Field
}
```

**DefinitionKind 枚举**：
```go
DefKindEntity    // 1 - Entity
DefKindManager   // 2 - Manager
DefKindModule    // 3 - Module
DefKindMechanism // 4 - Mechanism
DefKindMessage   // 5 - NetWall 消息 / DataStruct
DefKindEnum      // 6 - 枚举
DefKindStruct    // 7 - Common DataStruct
```

---

### SymbolTable（符号表）

位于 `types/symbol_table.go`。

```go
// 创建
st := types.NewSymbolTable()

// 注册（带 scope 命名空间）
st.RegisterWithScope("mme", entity)           // 简单名 + scope.简单名 两种查找
st.RegisterWithScope("Hero", netwallMessage)  // scope = NetWall 包名

// 查找
def, ok := st.Lookup("HeroManager")           // 简单名查找
def, ok = st.Lookup("mme.HeroManager")        // 限定名查找
def    = st.MustLookup("HeroManager")         // 未找到则 panic

// 调试
allSymbols := st.AllSymbols()
```

**注册顺序（必须严格遵守）**：
1. MME 对象（Entity → Manager → Module → Mechanism），scope = `"mme"`
2. NetWall 消息（Request → Response → Notify → DataStruct），scope = 包名
3. Common DataStructs（来自 headfile.yaml），scope = `"MME"`

---

### ReferenceResolver（引用解析器）

位于 `types/symbol_table.go`。

```go
resolver := types.NewReferenceResolver(symbolTable)

// 解析单个字段列表（收集错误，不提前中断）
errors := resolver.ResolveFields(fields)

// 错误访问
errors := resolver.GetErrors()
hasErr := resolver.HasErrors()
```

自动递归处理嵌套类型：
- `map<int32, HeroManager>` → 解析 HeroManager 并填充 `ValueType.ResolvedType`
- `xmap<int64, repeated HeroModule>` → 递归解析 HeroModule

---

## 二、MME 对象模型

### MMEObject（基类）

位于 `mmeobj/mme.go`，所有 MME 对象内嵌此结构。

```go
type MMEObject struct {
    Name       string
    ObjectType ObjectType
    Fields     []*types.Field
    Settings   map[string]any  // YAML 中的 Settings 块
    SourceFile string          // 源 YAML 文件路径（必须在解析时设置）
}
```

**字段遍历**：
```go
obj.RangeFields(func(field *types.Field) bool {
    // return false 中断遍历
    return true
})
obj.RemoveField(field)
obj.HasMapField()
obj.HasXMapField()
```

### 继承结构

```
MMEObject (base)
├── Entity     { *MMEObject }
├── Manager    { *MMEObject }
├── Module     { *MMEObject; Settings }
└── Mechanism  { *MMEObject; Requests []*NetMessage; Notifies []*NetMessage }
```

---

## 三、代码生成输出结构

### Proto 输出（`protocol/protocol/`）

```
protocol/
├── entities.proto       # Entity 对应的 Proto 消息
├── managers.proto
├── modules.proto
├── mechanisms.proto
├── common.proto         # 公共枚举 + DataStruct
└── hero.proto           # 各 NetWall 包的消息定义
```

### Go 输出（`internal/game/mme/`）

```
internal/game/mme/
├── entities/
│   └── player/
│       ├── player_entity_wrapper.go   # EntityWrapper
│       └── player_entity.go           # EntityImpl（Agent 层，懒加载 Manager）
├── managers/
│   └── hero_manager/
│       └── hero_manager_wrapper.go
├── modules/
│   └── hero_module/
│       └── hero_module_wrapper.go
└── mechanisms/
    └── hero_mechanism/
        └── hero_mechanism_wrapper.go
```

### Protocol IDs 输出（`pkg/proto/pb/`）

```
pkg/proto/pb/
├── protocol_ids.pb.go    # 协议 ID 常量（每个 NetWall 包一个文件）
├── pb_factories.go       # Proto 消息工厂（根据 ID 构造对象）
└── ref_factories.go      # Ref 工厂
```

---

## 四、字段访问模式（代码生成器中）

### 安全遍历字段并访问类型

```go
for _, field := range mechanism.GetFields() {
    switch {
    case field.Type.IsFieldBaseType():
        // 基础类型字段，直接处理
        typeName := field.Type.Kind.String() // "int32", "string" 等
        
    case field.Type.IsMMEObjectType() && field.Type.IsResolved():
        def := field.Type.GetResolvedDefinition()
        switch def.GetKind() {
        case types.DefKindModule:
            // 字段引用的是 Module
        case types.DefKindMechanism:
            // 字段引用的是 Mechanism
        }
        
    case field.Type.IsMapField():
        valueType := field.Type.ValueType
        if valueType.IsMMEObjectType() && valueType.IsResolved() {
            def := valueType.MustGetResolvedDefinition()
            // map<Key, SomeModule>
        }
        
    case field.Type.IsXMapField():
        // xmap 与 map 同理，但生成 XMap Wrapper
    }
}
```

### 获取对象的 Proto 包名（用于 import）

```go
protoFile := ctx.GetProtoImportByObjectName("HeroManager") // "managers.proto"
pkgName   := ctx.GetPackageProtoName("HeroManager")        // "MME"
```

---

## 五、FiledChecker 验证规则

`FiledChecker.Check()` 在 Pass 3 之后执行，验证以下规则：

| 对象类型 | 字段约束 |
|---------|---------|
| **Entity** | 字段类型只能是 Manager |
| **Manager** | 字段只能是直接 Module 引用 或 map/xmap<Key, Module> |
| **Module** | 字段只能是 Mechanism（直接引用，不支持 map）|
| **Mechanism** | 字段只能是基础类型（不能引用 MME 对象）|
| **所有对象** | 字段编号 ≤ 64 |

---

## 六、CLI 参数说明

| 参数 | 默认值 | 说明 |
|-----|-------|-----|
| `--blueprint-dir` | `../protocol/blueprint` | YAML 蓝图目录 |
| `--proto-output` | `../protocol/protocol` | Proto 文件输出目录 |
| `--go-output` | `internal/game/mme` | Go 文件输出目录 |
| `--protocol-ids-output` | `pkg/proto/pb` | 协议 ID 文件输出目录 |
| `--controller-dir` | `internal/game/controller_v2` | Controller 目录（用于 Router 生成）|
| `--router-output` | `internal/game/routers/routers.go` | Router 输出文件 |
| `--controller-path` | `<controller-dir>/controller.go` | Controller 文件路径 |
| `--debug` | `false` | 输出调试信息 |

> 省略 `--router-output` 或 `--controller-dir` 时跳过 Router 生成，不影响其他输出。

---

## 七、Wrapper 层架构

### Wrapper 核心功能

每层（Entity/Manager/Module/Mechanism）自动生成对应 Wrapper，提供：

| 功能 | 说明 |
|-----|-----|
| 脏标记追踪 | 位标记系统（`IDirtyFlag`），追踪字段变更 |
| 字段元数据 | `FieldMetas` 管理字段同步类型和权限 |
| 增量同步 | `ToIncrementalProtoWithContext(ctx)` 按脏标记生成增量 proto |
| 全量序列化 | `ToProto()` / `FromProto()` |
| 持久化支持 | `BuildMongoUpdate()` 生成 MongoDB 增量更新 |
| 深拷贝 | `DeepCopy()` / `DeepCopyTo()` |
| 脏标记清除 | `ClearAllDirty()` |

### 脏标记链接机制

Wrapper 通过 `Link()` 建立层级间的脏标记传递链：

```go
// Module → Mechanism 链接
hm.BaseWrapper.Link(hm.GetDirtyTracker(), HeroModuleDirtyBaseBit)

// Manager → Module（通过 XMapWrapper 自动管理）
heroMapLink = xmapwrapper.NewXMapWrapperWithParent(
    &data.HeroMap,
    hm.GetDirtyTracker(),
    HeroManagerDirtyHeroMapBit,
    NewHeroModuleWrapper,
)
```

传递方向：`Mechanism` → `Module` → `Manager` → `Entity`（子对象脏 → 父对象自动标脏）

### SyncContext 双模式

```go
SyncContextServer  // 同步所有字段（access=all + access=s）
SyncContextClient  // 仅同步 access=all 和 access=c 的字段
```

`FieldCanBeIncrementalSynced(wrapper, dirtyBit, fieldIndex, ctx)` 判断字段在当前上下文是否可同步。

### xmap 的 Proto 增量生成规则

YAML 中每个 `xmap<KeyType, ValueType> FieldName: N` 字段，自动生成：

```protobuf
// 1. 全量字段（标准 map，编号 N）
map<KeyType, ValueType> FieldName = N;

// 2. 增量变更列表（编号 1000+N，保留区间 1000-9999）
repeated FieldName_XXXMapChangeRecord FieldName_XXXChangeList = 1000+N;

// 3. 变更记录消息
message FieldName_XXXMapChangeRecord {
    KeyType  Key      = 1;
    ValueType Value   = 2;
    bool     IsDelete = 3;  // true=删除，false=新增/更新
}
```

### map 字段 Accessor 选择规则

| Value 类型 | 生成的 Accessor | 说明 |
|-----------|----------------|-----|
| 基础类型 / struct | `*xmap.MapAccessor` | 追踪 Set/Delete，支持增量 ChangeList |
| MME Object（Module/Mechanism） | `*xmapwrapper.XMapWrapper` | 管理子 Wrapper 生命周期，递归传递脏标记 |
| 其他指针类型 | **生成报错** | 需调整 YAML 声明 |

### 增量同步完整流程

```
1. Setter 修改数据  →  自动标记字段 DirtyBit
2. DirtyBit 通过 Link 向上传递（子 → 父）
3. 调用 ToIncrementalProtoWithContext(ctx)
4. FieldCanBeIncrementalSynced() 过滤字段
5. 网络传输增量 proto
6. 接收端 FromProto() 或增量合并
7. ClearAllDirty() 清除脏标记
```

### 路由分发系统（Location）

三层分发链，根据 `mme.MMELocation` 逐层定位到 Mechanism：

```
Entity.Location(loc)
  → loc.GetManagerIndex() → ManagerWrapper.Location(loc)
       → loc.GetModuleIndex() → ModuleWrapper.Location(loc)
            → loc.GetMechanismIndex() → (MechanismWrapper, MechanismType)
```

```go
// Entity 层
func (w *PlayerEntityWrapper) Location(loc *mme.MMELocation) (any, mme.MechanismType) {
    switch loc.GetManagerIndex() {
    case int32(PlayerEntityFieldIndexHeroMgr):
        return w.HeroManagerWrapper.Location(loc)
    }
    return nil, mme.MechanismType_Unknown
}

// Manager 层（含 map 分支）
func (w *HeroManagerWrapper) Location(loc *mme.MMELocation) (any, mme.MechanismType) {
    switch loc.GetModuleIndex() {
    case int32(HeroManagerFieldIndexHeroMap):
        moduleWrapper, exists := w.heroMapLink.Get(loc.GetKey())
        if exists { return moduleWrapper.Location(loc) }
    }
    return nil, mme.MechanismType_Unknown
}

// Module 层（终点，返回 Mechanism）
func (w *HeroModuleWrapper) Location(loc *mme.MMELocation) (any, mme.MechanismType) {
    switch loc.GetMechanismIndex() {
    case int32(HeroModuleFieldIndexBase):
        return w.BaseWrapper, mme.MechanismType_Hero
    }
    return nil, mme.MechanismType_Unknown
}
```

### Wrapper 层级结构示例

```
PlayerEntityWrapper
├── IDirtyFlag / FieldMetas
└── HeroManagerWrapper
    ├── heroMapLink: *XMapWrapper[int64, *HeroModule, *HeroModuleWrapper]
    └── HeroModuleWrapper（通过 map key 访问）
        ├── BaseWrapper:   *HeroMechanismWrapper
        │   └── skillsAccessor: *MapAccessor[int32, int32]
        └── LevelUpWrapper: *LevelUpMechanismWrapper
```

---

## 八、关键源码位置

| 功能 | 文件 |
|-----|-----|
| CLI 入口 | `blueprint_gen/cmd.go` |
| 解析编排 | `blueprint_gen/parser.go`, `mme_parser.go`, `netwall_parser.go` |
| 编译上下文 | `blueprint_gen/context.go` |
| 字段类型系统 | `blueprint_gen/types/types.go` |
| 符号表 | `blueprint_gen/types/symbol_table.go` |
| Definition 接口 | `blueprint_gen/types/definition.go` |
| MME 对象模型 | `blueprint_gen/mmeobj/mme.go` |
| 字段验证 | `blueprint_gen/field_checker.go` |
| Proto 生成 | `blueprint_gen/proto_gen.go`, `proto_mme_gen.go`, `proto_netwall_gen.go` |
| Go 生成 | `blueprint_gen/go_wrapper_gen.go`, `go_struct_gen.go` |
| Router 生成 | `blueprint_gen/router_gen_adapter.go` |
| Agent 生成 | `blueprint_gen/agent_gen/` |
