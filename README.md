# NetBluePrint 设计文档

## 概述

NetBluePrint 是一个基于 YAML 的网络/实体结构协议设计与自动生成系统，用于定义各类业务中的复杂实体分层数据结构及网络消息。设计核心是用抽象的蓝图（Blueprint）实现 Entity → Manager → Module → Mechanism 的多层自动化生成，并天然支持结构演变、增量同步和分布式场景。

MME 架构定义了典型的四层结构，不局限于任何具体领域模型，可复用到绝大多数服务端数据建模需求中。

---

## MME 分层模型与自动生成（抽象规范）

### 四层核心结构

- **Entity（实体）**：业务载体，生命周期管理，不直接管理内部逻辑。包含一个或多个 Manager，作为顶层数据结构。不能使用map/slice等容器定义Field。
- **Manager（管理器）**：以单例模式或Map聚合编排Module单元，例如以map，唯一 key 组织和管理多个 Module，负责查找、增删、全/增量同步。
- **Module（模块）**：功能单元，自由组合一组 Mechanism，定义完整业务/属性能力。一个 Module 可包含多个 Mechanism，实现功能组合。不能使用map/slice等容器定义Field。
- **Mechanism（机制）**：最小的可扩展数据和逻辑单元，专注每个方向的数据与行为。包含具体的业务字段、请求和通知定义。

### 实际代码结构示例

```go
// Entity 层：PlayerEntity
type PlayerEntity struct {
    XXXId       int64
    HeroManager *HeroManager
}

// Manager 层：HeroManager
type HeroManager struct {
    HeroMap map[int64]*HeroModule  // map 组织多个 Module
}

// Module 层：HeroModule
type HeroModule struct {
    Base    *HeroMechanism      // 组合 Mechanism
    LevelUp *LevelUpMechanism   // 组合 Mechanism
}

// Mechanism 层：HeroMechanism
type HeroMechanism struct {
    Id         int64
    ConfId     int32
    CreateTime int64
    UseTimes   int32
    Skills     map[int32]int32  // 值类型 map
}
```

### 自动生成规则

- 所有结构均由 YAML → Proto → go代码自动生成，保持多端一致性。
- MME中Object定义和proto文件和go结构文件的定义是一一对应的。
- 字段一旦上线，其 FieldIndex **不可更改、不可复用**，可删除但严禁插入。
- 允许在结构“末尾”顺序追加字段。
- Manager 层推荐 map/dictionary 扩展模式，横向分片。
- Module 层采用组合机制，灵活拼装与插拔 Mechanism。
- Mechanism 支持标量、数组、map 等多种基础和复合类型。
- 从 YAML 自动生成对应 protobuf、go/typescript 等多端代码。
- 每层自动生成脏数据追踪、全量与增量同步接口。
- 生成proto文件，默认生成optional字段。
- 自动生成DirtyBit和FieldIndex 常量，DirtyBit命名规则：{ObjName}Dirty{FieldName}Bit, FieldIndex的命名规则：{ObjName}FieldIndex{FieldName}
- FieldIndex 范围是 [0,63]
- 对于Entity对象，默认自动生成唯一Id字段,字段名称是XXXId，类型是int64： `int64 XXXId = 10000;`, 不可使用optional。
```protobuf
syntax = "proto3";

package MME;
option go_package = "./mme";

import "protocol/common.proto";
// XXXId 是根据PlayerEntity结构体中Id字段自动化生成的，结构模式固定，不要修改。
message PlayerEntity {
  managers.HeroManager HeroManager = 1;

  int64 XXXId = 10000;
}
```

---

## 生成示例（非领域化通用范式）

```yaml
# blueprint/mme/manager.yaml
SomeManager:
  keyField: int64
  module: SomeModule

# blueprint/mme/modules.yaml
SomeModule:
  mechanisms:
    - AMechanism
    - BMechanism

# blueprint/mme/mechanisms.yaml
AMechanism:
  fields:
    Attribute1: int32
    Scores: map<int32, int32>
    # 脏位、增量同步、字段索引等自动编码
```

---

## 协议、数据同步与进化约束

### 增量同步自动生成原则： 
- 对 map/字典类字段，系统生成 ChangeList 与 ChangeType 枚举，实现高效分段同步。
- Wrapper 层自动集成 ToProto/FromProto/ToIncrementalProto 等接口，支持只同步变动部分。

### 字段演变和约束：
- 字段顺序一经上线不可变，仅允许 **末尾追加字段或删除字段**，FieldIndex 禁止覆盖。
- 各结构层级组合可扩展，但原有结构、层级关系不可变。
- YAML 蓝图变更应始终保持向后兼容，所有代码和协议由蓝图自动化生成。

---

## 增量 map 字段自动生成规则

在 blueprint YAML 中声明 xmap 字段（如 xmap<int64, SomeModule> Items: 1），代码生成工具会自动为每个 xmap 字段生成高效的全量和增量同步支持。

### 1. 自动 proto 字段生成与命名范式

- 除 NetWall 外，MME 相关结构体的普通字段均生成为 `optional`（`proto3` 语义下的可选）以提升演进兼容性；`map` 与 `repeated` 不使用 `optional`。
- Yaml文件中，每个 xmap<KeyType, ValueType> 字段 `FieldName` ，自动生成：
  - `map<KeyType, ValueType> FieldName = N;` // 全量同步（xmap 在 YAML 中声明，proto 中生成为标准 map）
  - `repeated <FieldName>_XXXMapChangeRecord FieldName_XXXChangeList = 1000+N;` // 增量同步（编号范围 1000-9999 为保留编号区间）
  - 对应 change record message：

```protobuf
message <FieldName>_XXXMapChangeRecord {
  <KeyType> Key = 1;
  <ValueType> Value = 2;
  bool IsDelete = 3; // 若 true 为删除，false 为新增/变更；删除操作时 Value 字段可忽略
}
```
注意：`{FieldName}` 会被替换为实际的字段名，如 `Skills`、`HeroMap` 等。

### 2. 典型同步流程

- **全量同步**：首次下发时，直接同步 map 字段。
- **增量模式**：高频变更场景仅需同步 `FieldName_XXXChangeList`，用 `IsDelete` 区分 set/delete：
  - `IsDelete = false`：新增或更新操作，需要提供 `Value` 字段
  - `IsDelete = true`：删除操作，`Value` 字段可忽略（但 proto 定义中仍需包含该字段）
- **客户端/服务端和数据库均可只消费变更记录，恢复 map 状态而无需全表覆盖**。

### 3. 规则摘要

- blueprint 层每声明一个 map 字段，均自动拥有全量及变更双通道能力。
- 支持所有 key/value 基础类型及 Message 类型，保持多端与协议一致。
- 变更记录只需增/删，无需携带历史快照。
- 可与脏字段/bit 标记、ToIncrementalProto 等增量同步机制无缝协同。
- （后续可扩展 message 字段，实现更丰富的数据变更表达）。

---

## Go代码自动生成规则：

### Go结构体FieldIndex/FieldDirty 常量自动生成规则：
- 对于Entity对象：
    1. FieldIndex的命名规范是{Object}FieldIndex{FieldName}
    2. Id的FieldIndex默认是0.
    3. 其他定义Field的FieldIndex从1开始，按YAML中字段定义顺序递增
    4. DirtyBit生成规则：Dirty{ObjName}{FieldName}Bit int64 = 1 << FieldIndex
- 对于Module/Manager/Mechanism对象：
    1. FieldIndex的命名规范是{Object}FieldIndex{FieldName}
    2. 定义的Field的FieldIndex从0开始，按YAML中字段定义顺序递增
    3. DirtyBit生成规则：Dirty{ObjName}{FieldName}Bit int64 = 1 << FieldIndex

示例:
```yaml
---
Mechanisms:
  #英雄机制
  - HeroMechanism:
    #英雄实例唯一Id
    int64 Id: 1 [blueprint:"access=all"]
    #英雄配置ID
    int32 ConfId: 2 [blueprint:"access=all"]
    #玩家获得英雄的时间
    int64 CreateTime: 3 [blueprint:"access=s"]
    #英雄被使用次数
    int32 UseTimes: 4 [blueprint:"access=all"]
    #技能
    xmap<int32, int32> Skills: 5 [blueprint:"access=all"]
---
```
```go
const (
	HeroMechanismFieldIndexId = uint8(0)
	HeroMechanismFieldIndexConfId
	HeroMechanismFieldIndexCreateTime
	HeroMechanismFieldIndexUseTimes
	HeroMechanismFieldIndexSkills
)

// Dirty bits for Mechanism fields
const (
	HeroMechanismDirtyIdBit         int64 = 1 << HeroMechanismFieldIndexId
	HeroMechanismDirtyConfIdBit     int64 = 1 << HeroMechanismFieldIndexConfId
	HeroMechanismDirtyCreateTimeBit int64 = 1 << HeroMechanismFieldIndexCreateTime
	HeroMechanismDirtyUseTimesBit   int64 = 1 << HeroMechanismFieldIndexUseTimes
	HeroMechanismDirtySkillsBit     int64 = 1 << HeroMechanismFieldIndexSkills
)
```

### map字段Accessor规则

在Go代码自动生成阶段，map字段会根据Value类型自动选择访问器模式：

1. **Value为值类型（如int、string、struct等）**：
   - 自动生成 `*xmap.MapAccessor` 作为map的读写访问入口，支持KV对的批量与增量操作。
   - 支持脏标记、变更追踪、配合增量同步协议字段无缝衔接。

2. **Value为MME Object引用类型（如 *Module、*Mechanism 等）**：
   - 自动生成 `*xmapwrapper.XMapWrapper`，管理map内的MME对象的包装与生命周期关系，递归支持子对象的脏数据同步与变更。
   - 支持多层嵌套脏位递推，一致性强。

3. **Value为其他指针/引用类型**：
   - 代码生成器将抛出异常并阻止生成，需用户显式声明对应适配器或调整YAML声明。

该自动访问器规则确保：map字段不论值类型与对象复杂度如何，都能自动适配协议增量同步和本地脏标记递归运算，提升迭代安全性。

---

## Go Wrapper 层架构

MME 架构为每一层（Entity、Manager、Module、Mechanism）自动生成对应的 Wrapper 包装器，提供统一的脏标记追踪、增量同步、数据转换和持久化能力。

### Wrapper 核心功能

每个 Wrapper 包含以下核心组件：

1. **数据层（Data）**：原始结构体，存储实际业务数据
2. **脏标记追踪（IDirtyFlag）**：位标记系统，追踪字段变更
3. **字段元数据（FieldMetas）**：管理字段的同步类型和访问权限
4. **嵌套 Wrapper 链接**：建立父子层级的脏标记传递关系

### 脏标记链接机制

Wrapper 层通过 `Link` 机制建立层级间的脏标记传递链：

- **Entity → Manager**：Entity 的脏标记变化会自动触发 Manager 层的脏标记
- **Manager → Module**：通过 `XMapWrapper` 管理 map 中的 Module，Module 脏标记向上传递
- **Module → Mechanism**：Module 通过 `Link` 方法将 Mechanism 的脏标记关联到自身

示例：
```go
// Module 层链接 Mechanism
hm.BaseWrapper.Link(hm.GetDirtyTracker(), HeroModuleDirtyBaseBit)

// Manager 层通过 XMapWrapper 自动链接 Module
heroMapLink = xmapwrapper.NewXMapWrapperWithParent(
    &data.HeroMap,
    hm.GetDirtyTracker(),
    HeroManagerDirtyHeroMapBit,
    NewHeroModuleWrapper,
)
```

### 标准接口方法

每个 Wrapper 自动生成以下标准接口：

- `InitFieldContext()`：初始化字段元数据上下文，设置字段同步类型
- `ToProto()`：将数据转换为完整的 protobuf 结构体
- `FromProto()`：从 protobuf 结构体加载数据
- `ToIncrementalProtoWithContext(ctx SyncContext)`：根据脏标记生成增量同步数据
- `BuildMongoUpdate()`：构建 MongoDB 增量更新操作
- `DeepCopy()` / `DeepCopyTo()`：深拷贝数据
- `ClearAllDirty()`：清除所有脏标记

### 增量同步上下文（SyncContext）

系统支持两种同步上下文：

- **SyncContextServer**：服务端同步，所有字段都可同步
- **SyncContextClient**：客户端同步，仅同步标记为 `FieldTypeSync` 的字段

通过 `FieldCanBeIncrementalSynced()` 方法判断字段是否可以在指定上下文中增量同步，结合字段的 `access` 权限（all/s/c）和字段元数据，实现细粒度的同步控制。

### 字段元数据（FieldMetas）

`FieldMetas` 管理系统字段的元信息：

- **FieldType**：字段类型标记，如 `FieldTypeSync`（需要同步）、`FieldTypeLogin`（登录时加载）等
- **MatchesAll()**：判断字段是否匹配所有指定的类型标记，用于同步控制

示例：
```go
// 初始化字段上下文
m.fieldMetas.SetFieldType(HeroMechanismFieldIndexId, fieldmeta.FieldTypeSync)
m.fieldMetas.SetFieldType(HeroMechanismFieldIndexCreateTime, fieldmeta.FieldTypeSync)

// 判断字段是否可同步
if mmemodel.FieldCanBeIncrementalSynced(m, HeroMechanismDirtyIdBit, HeroMechanismFieldIndexId, ctx) {
    // 生成增量数据
}
```

### Wrapper 层级结构示例

```
PlayerEntityWrapper
├── data: *PlayerEntity
├── IDirtyFlag: 脏标记追踪器
├── FieldMetas: 字段元数据
└── HeroManagerWrapper
    ├── data: *HeroManager
    ├── heroMapLink: *XMapWrapper[int64, *HeroModule, *HeroModuleWrapper]
    └── HeroModuleWrapper (通过 map 访问)
        ├── data: *HeroModule
        ├── BaseWrapper: *HeroMechanismWrapper
        └── LevelUpWrapper: *LevelUpMechanismWrapper
            ├── data: *LevelUpMechanism
            └── skillsAccessor: *MapAccessor[int32, int32] (如果存在 map 字段)
```

### 数据持久化支持

Wrapper 层自动生成 `BuildMongoUpdate()` 方法，支持 MongoDB 的增量更新：

- 仅更新标记为脏的字段
- 支持嵌套字段路径
- 对于 map 字段，全量覆盖（MongoDB map 字段更新限制）

### Map 操作追踪机制

对于 map 字段，系统通过 `MapAccessor`（值类型）或 `XMapWrapper`（MME Object 类型）追踪所有操作：

**MapAccessor（值类型 map）**：
- 记录 Set、Delete 操作
- 通过 `RangeOperations()` 遍历所有变更操作
- 支持增量同步时生成 `ChangeList`

**XMapWrapper（MME Object map）**：
- 管理 map 中每个 Object 的 Wrapper 生命周期
- 追踪 Object 的增删改操作
- 递归支持子对象的脏标记传递
- 支持增量同步时，嵌套 Object 的增量数据生成

示例（值类型 map 增量同步）：
```go
if mmemodel.FieldCanBeIncrementalSynced(m, HeroMechanismDirtySkillsBit, HeroMechanismFieldIndexSkills, ctx) {
    incremental.Skills_XXXChangeList = make([]*mme.HeroMechanism_Skills_XXXMapChangeRecord, 0)
    m.skillsAccessor.RangeOperations(func(key int32, operation xmap.MapOperation[int32]) bool {
        switch operation.Type {
        case xmap.SetOperation:
            v, _ := m.skillsAccessor.Get(key)
            incremental.Skills_XXXChangeList = append(incremental.Skills_XXXChangeList, &mme.HeroMechanism_Skills_XXXMapChangeRecord{
                Key:   key,
                Value: v,
            })
        case xmap.DeleteOperation:
            incremental.Skills_XXXChangeList = append(incremental.Skills_XXXChangeList, &mme.HeroMechanism_Skills_XXXMapChangeRecord{
                Key:      key,
                IsDelete: true,
            })
        }
        return true
    })
}
```

### 增量同步完整流程

1. **数据变更**：通过 Wrapper 的 Setter 方法修改数据，自动标记脏位
2. **脏标记传递**：子对象的脏标记通过 Link 机制向上传递
3. **生成增量数据**：调用 `ToIncrementalProtoWithContext(ctx)` 生成增量 protobuf
4. **同步判断**：根据 `SyncContext` 和字段元数据判断字段是否可同步
5. **网络传输**：将增量数据序列化并发送到客户端/服务端
6. **数据合并**：接收端通过 `FromProto()` 或增量合并逻辑更新本地数据
7. **清除脏标记**：同步完成后调用 `ClearAllDirty()` 清除脏标记

### Wrapper 使用示例

```go
// 创建 Entity Wrapper
entity := NewPlayerEntity()
wrapper := NewPlayerEntityWrapper(entity)

// 初始化字段上下文（设置同步类型）
wrapper.InitFieldContext()

// 修改数据（自动标记脏位）
heroWrapper := wrapper.GetHeroManager().HeroMap_Set(1, NewHeroModule())
heroWrapper.GetBase().SetConfId(100)

// 获取增量同步数据（仅包含变更字段）
incremental := wrapper.ToIncrementalProto(mmemodel.SyncContextClient)

// 持久化到 MongoDB（仅更新脏字段）
builder := mgo_builder.NewMongoUpdateBuilder()
wrapper.BuildMongoUpdate(builder, mgo_builder.NewNestedPath("player"))

// 同步完成后清除脏标记
wrapper.ClearAllDirty()
```

---

## YAML 文件组织规范及目录

保持原有 YAML 命名和组织方式，具体内容可参考下述模式：

- headfile.yaml: 通用全局字段/配置
- entities.yaml: 实体（Entity）定义，引用一个或多个 Manager
- manager.yaml: Manager 及其含有的 Module 定义
- mechanisms.yaml: Mechanism 原子机制定义
- modules.yaml: Module 机制拼装与配置

## 目录结构

```
blueprint/
├── mme/                    # MME (Entity → Manager → Module → Mechanism)
│   ├── headfile.yaml       # 通用字段、选项、公共数据结构
│   ├── entities.yaml       # 实体定义（引用 Manager）
│   ├── manager.yaml        # 管理器定义（容纳多个 Module）
│   ├── mechanisms.yaml     # 机制定义（数据/请求/通知）
│   └── modules.yaml        # 模块定义（组合 Mechanism）
├── netwall/                # 网络消息墙
│   ├── Core.yaml           # 核心网络消息墙
│   ├── mme.yaml            # MME 网络墙（对外暴露的 MME 接口）
│   └── Sample.yaml         # 示例网络消息墙
└── README.md               # 本文档
```

## YAML 文件组织规则

### 1. 文件命名规范

- `headfile.yaml`: 定义通用字段、选项、公共数据结构
- `entities.yaml`: 定义实体（Entity），引用 Manager
- `manager.yaml`: 定义管理器（Manager），组织多个 Module
- `mechanisms.yaml`: 定义机制（Mechanism）与数据结构
- `modules.yaml`: 定义模块（Module）对机制的组合与配置
- `Core.yaml`: 定义核心网络消息墙
- `mme.yaml`: 定义 MME 网络消息墙
- `Sample.yaml`: 定义示例网络消息墙

### 2. 文件结构规范（MME）

MME 采用四层结构：Entity → Manager → Module → Mechanism。

- headfile.yaml - 通用定义文件
```yaml
# 字段/选项/通用数据结构
---
- EntityFields:

- ModuleFields:
  bool includeInLoginData: false
- ModuleStorageOption:
    Choices:
      - Single
      - Dictionary

- MechanismFields:
  bool readOnly: false
  bool activeSync: false

- MechanismDataFieldOption:
  - access:
      Default: all
      Description: 数据访问权限（all/s/c）
      Choices: [all, s, c]

- DataStruct:
  - Fish:
      int32  FishId: 1
      map<int32, int32> BodyMap: 2
```

- entities.yaml - 实体定义文件
```yaml
# 实体定义
---
Entity:
  - PlayerEntity:
      # 引用 Manager（字段名自定义）
      - HeroManager HeroManager: 1
```

- manager.yaml - 管理器定义文件
```yaml
# 管理器是组织多个 Module 的容器
---
Managers:
  - HeroManager:
      map<int64, HeroModule> HeroMap: 1
```

- modules.yaml - 模块定义文件
```yaml
# 模块组合多个机制
---
Modules:
  - HeroModule:
      - HeroMechanism Base: 1
        Settings:
          activeSync: true
      - LevelUpMechanism LevelUp: 2
        Settings:
          readOnly: true
          AutoLevelUp: true
      - ManualUnlockMechanism TalentUnlock: 3
      - WearMechanism SkinWear: 4
```

- mechanisms.yaml - 机制与数据结构定义文件
```yaml
# 机制定义
---
Mechanisms:
  - LevelUpMechanism:
      int32 CurLevel: 1 [blueprint:"access=all"]
      int32 CurExp:   2 [blueprint:"access=all"]
      int32 ConfId:   3 [blueprint:"access=all", orbit:"Access=all", mme:"Access=all"]
    Settings:
      bool AutoLevelUp: true
    Requests:
      - AskLevelUp:
          int32 UpNum: 1
        Rsp:
          string Result: 1
    Notifies:
      - ExpChange:
          int32 Exp: 1

  - HeroMechanism:
      int32 ConfId:     1 [blueprint:"access=all"]
      int64 CreateTime: 2
      int32 UseTimes:   3

  - ManualUnlockMechanism:
      map<int32, bool>  UnlockMap: 1 [blueprint:"access=all"]

  - WearMechanism:
      map<int32, int32> WearMap: 1 [blueprint:"access=all"]
      int32             ConfId:  2
---
# 数据结构
DataStruct:
  - Chicken:
      int32             EggCnt: 1 [blueprint:"access=all"]
      map<int32, int32> EggMap: 2
```

### 3. 文件结构规范（NetWall）

每个 NetWall YAML 文件包含以下部分：
- `NetWall`：定义网络消息墙的名称、请求和通知
  - `Name`：NetWall 名称，用于生成 proto 文件的 package 名称和文件名
  - `Requests`：请求消息列表，每个请求可以包含字段和可选的 `Rsp` 响应
  - `Notifies`：通知消息列表，每个通知包含字段
- `DataStructs`：数据结构定义，可以在当前 NetWall 或其他 NetWall 中使用

**生成规则**：
- 每个 NetWall YAML 文件生成一个独立的 proto 文件
- **YAML 文件名**：使用小写驼峰命名方式（如 `core.yaml`、`mme.yaml`、`sample.yaml`）
- **生成的 Proto 文件名**：`{NetWall.Name}.proto`（首字母大写，如 `core.proto`、`mme.proto`、`sample.proto`）
- **package 名称**：`package {NetWall.Name};`（首字母大写，如 `package Core;`、`package MME;`、`package Sample;`）
- 如果引用了其他 NetWall 的数据结构（如 `Core.MMELocation`），会自动添加 `import` 语句

示例（core.yaml）：
```yaml
#核心网络消息墙
---
NetWall:
  Name: Core
  Requests:
    - SearchBook:
        #位置
        string Query: 1
        #要第几页
        int32 PageNumber: 2
      Rsp:
        Book Result: 1
    - HeartBeat:
    
  Notifies:
    - BeAttacked:
        int32 CurHp: 1
---
#数据结构,Core中定义的数据结构可以在其他墙中使用例如Core.OK
DataStructs:
  - Book:
      string Content: 1
  ##通用成功
  - OK
  ##通用失败
  - Error:
      string Reason: 1
  ##MME位置
  - MMELocation:
      int64 EntityId: 1
      int32 ModuleId: 2
      int32 MechanismIndex: 3
```

MME 网络墙（对外暴露的 MME 接口，引用 Core 的数据结构）：
```yaml
#MME网络消息墙
---
NetWall:
  Name: MME
  Requests:
    - AskLevelUp:
        int32 UpNum: 1
        Core.MMELocation Loc: 1000
      Rsp:
        string Result: 1

  Notifies:
    - ExpChange:
        int32 Exp: 1
        Core.MMELocation Loc: 1000
```

## 字段与编号规范

- 支持基础类型与 `proto3` 对齐：`int32`, `int64`, `string`, `bool`, `float`, `double`, `map<k,v>`, `repeated`，以及自定义结构。
- 每个字段必须有唯一编号；从 1 递增；不重复，不随意修改。
- 字段选项必须在 `headfile.yaml` 的 `MechanismDataFieldOption` 中事先声明，使用格式 `[blueprint:"optionName=value"]`。支持多个选项，以逗号分隔。
- `ModuleStorageOption` 定义模块存储形态（Single/Dictionary），由生成器或上层使用方诠释具体意义。

示例：
```yaml
- MechanismDataFieldOption:
  - access:
      Default: all
      Description: 数据访问权限
      Choices: [all, s, c]

# 使用
int32 FieldName: 1 [blueprint:"access=all"]
```

## 层级结构说明（MME）

MME 四层：Entity → Manager → Module → Mechanism。

- Entity（实体层）
  - 定义游戏实体（如 `PlayerEntity`），包含若干 Manager 字段。
  - 定义于 `entities.yaml`。
  - 示例：
```yaml
Entity:
  - PlayerEntity:
      - HeroManager HeroManager: 1
```

- Manager（管理器层）
  - 作为模块容器，通常维护 `map<id, Module>`。
  - 定义于 `manager.yaml`。
  - 示例：
```yaml
Managers:
  - HeroManager:
      map<int64, HeroModule> HeroMap: 1
```

- Module（模块层）
  - 功能模块，组合多个 Mechanism，并可在 `Settings` 中声明默认行为（例如 `activeSync`、`readOnly` 等）。
  - 定义于 `modules.yaml`。
  - 示例：
```yaml
Modules:
  - HeroModule:
      - HeroMechanism Base: 1
        Settings:
          activeSync: true
      - LevelUpMechanism LevelUp: 2
        Settings:
          readOnly: true
          AutoLevelUp: true
```

- Mechanism（机制层）
  - 定义具体数据、请求与通知。可使用 `MechanismFields` 的通用字段与 HeadFile 的字段选项。
  - 定义于 `mechanisms.yaml`。
  - 示例：
```yaml
Mechanisms:
  - LevelUpMechanism:
      int32 CurLevel: 1 [blueprint:"access=all"]
      int32 CurExp:   2 [blueprint:"access=all"]
    Requests:
      - AskLevelUp:
          int32 UpNum: 1
        Rsp:
          string Result: 1
    Notifies:
      - ExpChange:
          int32 Exp: 1
```

## 转换到 Proto 文件规则

### 1. 文件映射关系

| YAML 文件 | Proto 文件 | 说明 |
|-----------|------------|------|
| `blueprint/mme/headfile.yaml` | `protocol/common.proto` | 通用数据结构与字段选项 |
| `blueprint/mme/mechanisms.yaml` | `protocol/mechanisms.proto` | 机制数据结构 |
| `blueprint/mme/modules.yaml` | `protocol/modules.proto` | 模块组合结构 |
| `blueprint/mme/manager.yaml` | `protocol/managers.proto` | 管理器结构（map 到 Module） |
| `blueprint/mme/entities.yaml` | `protocol/entities.proto` | 实体结构（引用 Manager） |
| `blueprint/netwall/core.yaml` | `protocol/core.proto` | 核心网络墙（包含 Request、Notify 和 DataStructs） |
| `blueprint/netwall/mme.yaml` | `protocol/mme.proto` | MME 网络墙（包含 Request 和 Notify） |
| `blueprint/netwall/sample.yaml` | `protocol/sample.proto` | 示例网络墙（包含 Request 和 Notify） |

### 2. 包名与 go_package

- **MME 相关 Proto**：统一使用 `package MME;`，统一 `option go_package = "./mme";`（使用相对路径，实际生成路径为 `app/proto/mme`）
- **NetWall Proto**：每个 NetWall 生成独立的 proto 文件，package 名称使用 NetWall 的 `Name` 字段（首字母大写），如 `package Core;`、`package MME;`、`package Sample;`
- 所有 NetWall proto 文件统一使用 `option go_package = "./mme";`

### 3. NetWall 文件生成规则

每个 NetWall YAML 文件生成一个独立的 proto 文件，文件命名规则：
- **YAML 文件名**：使用小写驼峰命名方式，如 `core.yaml`、`mme.yaml`、`sample.yaml`
- **生成的 Proto 文件名**：使用 NetWall 的 `Name` 字段（首字母大写），如 `core.yaml` → `core.proto`，`mme.yaml` → `mme.proto`，`sample.yaml` → `sample.proto`
- **文件内容**：包含该 NetWall 的所有 `Requests`、`Notifies` 和 `DataStructs`

### 4. 网络墙消息转换

- NetWall 的 `Requests` 全部生成在 `message Request` 内部；`Notifies` 生成在 `message Notify` 内部。
- 简单请求允许无参数；有 `Rsp` 的请求在该请求的内部定义 `message Rsp`。
- `DataStructs` 生成在同一个 proto 文件中，作为独立的 `message` 定义。

示例（来自 `core.yaml`）：
```protobuf
syntax = "proto3";

package Core;

message Request {
    message SearchBook {
        // 位置
        string Query = 1;
        // 要第几页
        int32 PageNumber = 2;
        message Rsp {
            Book Result = 1;
        }
    }

    message HeartBeat {
    }
}

message Notify {
    message BeAttacked {
        int32 CurHp = 1;
    }
}

message Book {
    string Content = 1;
}

message OK {
}

message Error {
    string Reason = 1;
}

message MMELocation {
    optional int64 EntityId = 1;
    optional int32 ManagerId = 2;
    optional int64 Key = 3;
    optional int32 ModuleId = 4;
    optional int32 MechanismIndex = 5;
}
```

### 5. NetWall 之间的引用规则

- 如果 NetWall 中引用了其他 NetWall 的数据结构（如 `Core.MMELocation`），需要在 proto 文件头部添加 `import` 语句
- 引用格式：`{NetWallName}.{DataType}`，如 `Core.MMELocation`
- 生成器会自动检测跨 NetWall 的引用，并添加相应的 `import` 语句

示例（来自 `mme.yaml`）：
```protobuf
syntax = "proto3";

import "core.proto";

package MME;

message Request {
    message AskLevelUp {
        int32 UpNum = 1;
        Core.MMELocation Loc = 1000;
        message Rsp {
            string Result = 1;
        }
    }
}

message Notify {
    message ExpChange {
        int32 Exp = 1;
        Core.MMELocation Loc = 1000;
    }
}
```

### 6. MME 请求/通知转换与 Loc 规则

- 从 Mechanism 的 `Requests/Notifies` 派生出的对外接口，最终体现在 `netwall/mme.yaml` 中。
- 为保持一致性，网络层接口中的 `Core.MMELocation Loc` 使用保留编号 `1000`；若 YAML 已显式编写 Loc（如本仓库的 `netwall/mme.yaml`），应保证其为 `1000`。若未显式声明，生成器会自动注入。
- **保留编号范围**：
  - `1000`：Request 和 Notify 中，保留用于 `Loc` 字段

### 7. 导入语句

- **NetWall proto 文件**：如果引用了其他 NetWall 的数据结构，需要导入对应的 proto 文件（如 `import "core.proto";`）
- **MME proto 文件**：根据依赖关系自动导入，如 `import "common.proto";`、`import "mechanisms.proto";` 等
- `common.proto` 无需导入其他文件

### 8. 注释保留

- YAML 中 `#` 注释转换为 Proto 中 `//` 注释，并尽量保留层级。
- 字段注释会保留在对应字段的上方。

### 9. 字段访问权限（access）的实际应用：
- 1. access=all: 客户端和服务端都可以读写。
- 2. access=s: 仅服务端可读写，客户端只读。
- 3. access=c: 仅客户端可读写，服务端只读。
- 4. 生成器应在Go代码中添加相应的访问控制逻辑。

## 最佳实践

### 1. 命名规范

- 类型与消息使用 PascalCase，字段使用 camelCase
- 使用有意义的名称，避免难懂的缩写

### 2. 版本管理

- 字段编号一旦确定不要修改
- 新增字段使用新的编号
- 废弃字段保留编号并注释说明

### 3. 文档与生成同步

- 为关键字段添加注释，保持 YAML 与 Proto 同步
- 定期检查生成结果；验证语法、编译与序列化/反序列化

## 示例对照（节选）

- 来自 `blueprint/netwall/core.yaml` 的完整文件（YAML 文件名使用小写驼峰）：
```yaml
#核心网络消息墙
---
NetWall:
  Name: Core
  Requests:
    - SearchBook:
        #位置
        string Query: 1
        #要第几页
        int32 PageNumber: 2
      Rsp:
        Book Result: 1
    - HeartBeat:
    
  Notifies:
    - BeAttacked:
        int32 CurHp: 1
---
#数据结构,Core中定义的数据结构可以在其他墙中使用例如Core.OK
DataStructs:
  - Book:
      string Content: 1
  ##通用成功
  - OK
  ##通用失败
  - Error:
      string Reason: 1
  ##MME位置
  - MMELocation:
      int64 EntityId: 1
      int32 ModuleId: 2
      int32 MechanismIndex: 3
```
对应生成的 `core.proto`（文件名首字母大写，package 名称首字母大写）：
```protobuf
syntax = "proto3";

package Core;

message Request {

    message SearchBook {
        // 位置
        string Query = 1;
        // 要第几页
        int32 PageNumber = 2;
        message Rsp {
            Book Result = 1;
        }
    }


    message HeartBeat {
    }


}

message Notify {

    message BeAttacked {
        int32 CurHp = 1;
    }


}

message Book {
    string Content = 1;
}


// #通用成功
message OK {
}


// #通用失败
message Error {
    string Reason = 1;
}


// #MME位置
message MMELocation {
    optional int64 EntityId = 1;
    optional int32 ManagerId = 2;
    optional int64 Key = 3;
    optional int32 ModuleId = 4;
    optional int32 MechanismIndex = 5;
}
```

- 来自 `blueprint/mme/mechanisms.yaml` 的机制：
```yaml
Mechanisms:
  - LevelUpMechanism:
      int32 CurLevel: 1 [blueprint:"access=all"]
      int32 CurExp:   2 [blueprint:"access=all"]
      int32 ConfId:   3 [blueprint:"access=all", orbit:"Access=all", mme:"Access=all"]
    Requests:
      - AskLevelUp:
          int32 UpNum: 1
        Rsp:
          string Result: 1
    Notifies:
      - ExpChange:
          int32 Exp: 1
```
对应对外网络墙（`blueprint/netwall/mme.yaml`）以及生成的消息，包含保留位 `Loc=1000`：
```yaml
#MME网络消息墙
---
NetWall:
  Name: MME
  Requests:
    - AskLevelUp:
        int32 UpNum: 1
        Core.MMELocation Loc: 1000
      Rsp:
        string Result: 1

  Notifies:
    - ExpChange:
        int32 Exp: 1
        Core.MMELocation Loc: 1000
```
对应生成的 `mme.proto`（文件名首字母大写，package 名称首字母大写）：
```protobuf
syntax = "proto3";

import "Core.proto";

package MME;

message Request {

    message AskLevelUp {
        int32 UpNum = 1;
        Core.MMELocation Loc = 1000;
        message Rsp {
            string Result = 1;
        }
    }


}

message Notify {

    message ExpChange {
        int32 Exp = 1;
        Core.MMELocation Loc = 1000;
    }


}
```

以上内容已与当前 `mme` 目录下的 YAML 定义对齐，并引入 Manager 层级，确保从 YAML 到 Proto 的生成规则清晰一致。

---

## 架构优势与总结

### MME 架构核心优势

1. **分层清晰，职责明确**
   - Entity 负责生命周期管理
   - Manager 负责集合管理（map 模式）
   - Module 负责功能组合（多个 Mechanism）
   - Mechanism 负责原子能力（数据+行为）

2. **自动化程度高**
   - 从 YAML 蓝图自动生成 Proto、Go 代码
   - 自动生成脏标记系统、增量同步、持久化接口
   - 减少手写代码，降低出错概率

3. **支持高效增量同步**
   - 位标记脏追踪系统，精确追踪字段变更
   - 支持 map 字段的增量变更列表
   - 支持多种同步上下文（Server/Client）
   - 结合字段元数据和访问权限，实现细粒度同步控制

4. **类型安全与扩展性**
   - 强类型约束，编译期检查
   - 字段演变有严格规则，保证向后兼容
   - 支持灵活的机制组合，易于扩展新功能

5. **数据一致性保证**
   - 统一的 Wrapper 层提供一致的操作接口
   - 脏标记链接机制确保层级间数据一致性
   - 支持深拷贝、全量/增量转换

### 典型应用场景

- **游戏玩家数据管理**：Entity(PlayerEntity) → Manager(HeroManager/ItemManager) → Module(HeroModule/ItemModule) → Mechanism(LevelUp/Equip)
- **配置管理系统**：Entity(ConfigEntity) → Manager(ConfigManager) → Module(ConfigModule) → Mechanism(Version/Auth)
- **多端数据同步**：服务端与客户端通过增量同步协议，仅传输变更字段，节省带宽

### 设计原则总结

1. **字段编号不可变**：一旦上线，FieldIndex 和字段编号不可修改，只能追加或删除
2. **向后兼容优先**：所有结构演变必须保持向后兼容
3. **自动化优先**：尽可能通过代码生成减少手写代码
4. **增量同步优先**：优先使用增量同步，减少网络传输和数据库更新开销
5. **类型安全优先**：通过强类型和编译期检查，减少运行时错误

以上架构设计已在生产环境中验证，提供了高效、安全、可维护的数据建模和同步解决方案。
