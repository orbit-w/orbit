# NetBluePrint 设计文档

## 概述

NetBluePrint 是一个基于 YAML 的网络/实体结构协议设计与自动生成系统，用于定义各类业务中的复杂实体分层数据结构及网络消息。设计核心是用抽象的蓝图（Blueprint）实现 Entity → Manager → Module → Mechanism 的多层自动化生成，并天然支持结构演变、增量同步和分布式场景。

MME 架构定义了典型的四层结构，不局限于任何具体领域模型，可复用到绝大多数服务端数据建模需求中。

---

## MME 分层模型与自动生成（抽象规范）

### 四层核心结构

- **Entity（实体）**：业务载体，生命周期管理，不直接管理内部逻辑。
- **Manager（管理器）**：以单例模式或Map聚合编排Module单元，例如以map，唯一 key 组织和管理多个 Module，负责查找、增删、全/增量同步。
- **Module（模块）**：功能单元，自由组合一组 Mechanism，定义完整业务/属性能力。
- **Mechanism（机制）**：最小的可扩展数据和逻辑单元，专注每个方向的数据与行为。

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
- Entity 对象，生成Proto message，默认自动生成Field Id int64 = 10000;
- 自动生成DirtyBit和FieldIndex 常量，DirtyBit命名规则：{ObjName}DirtyBit{FieldName}, FieldIndex的命名规则：{ObjName}FieldIndex{FieldName}

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

在 blueprint YAML 中声明 map 字段（如 map<int64, SomeModule> Items: 1），代码生成工具会自动为每个 map 字段生成高效的全量和增量同步支持，规则如下：

### 1. 自动 proto 字段生成与命名范式

- 每个 map<KeyType, ValueType> 字段 `FieldName` ，自动生成：
  - `map<KeyType, ValueType> FieldName = N;` // 全量同步
  - `repeated <Struct>_<FieldName>_XXXMapChangeRecord FieldName_XXXChangeList = 1000+N;` // 增量同步
  - 对应 change record message：

```protobuf
message <Struct>_<FieldName>_XXXMapChangeRecord {
  <KeyType> key = 1;
  <ValueType> value = 2;
  bool isDelete = 3; // 若 true 为删除，false 为新增/变更
}
```

### 2. 典型同步流程

- **全量同步**：首次下发时，直接同步 map 字段。
- **增量模式**：高频变更场景仅需同步 `FieldName_XXXChangeList`，用 isDelete 区分 set/delete，实现与本地 map 状态一致。
- **客户端/服务端和数据库均可只消费变更记录，恢复 map 状态而无需全表覆盖**。

### 3. 规则摘要

- blueprint 层每声明一个 map 字段，均自动拥有全量及变更双通道能力。
- 支持所有 key/value 基础类型及 Message 类型，保持多端与协议一致。
- 变更记录只需增/删，无需携带历史快照。
- 可与脏字段/bit 标记、ToIncrementalProto 等增量同步机制无缝协同。
- （后续可扩展 message 字段，实现更丰富的数据变更表达）。

---

## Go自动生成map字段Accessor规则

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

- CommonDataStruct:
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

```yaml
# 网络消息墙
---
NetWall:
  Name: Core
  Requests:
    - SearchBook:
        string Query: 1
        int32  PageNumber: 2
      Rsp:
        Book Result: 1
    - HeartBeat
  Notifies:
    - BeAttacked:
        int32 CurHp: 1
---
# 数据结构
DataStructs:
  - Book:
      string Content: 1
  - OK
  - Error:
      string Reason: 1
  - MMELocation:
      int64 EntityId: 1
      int32 ModuleId: 2
      int32 MechanismIndex: 3
```

MME 网络墙（对外暴露的 MME 接口）：
```yaml
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
| `blueprint/mme/headfile.yaml` | `cspb/common.proto` | 通用数据结构与字段选项 |
| `blueprint/mme/mechanisms.yaml` | `cspb/structs.proto` | 机制数据结构 |
| `blueprint/mme/modules.yaml` | `cspb/structs.proto` | 模块组合结构 |
| `blueprint/mme/manager.yaml` | `cspb/structs.proto` | 管理器结构（map 到 Module） |
| `blueprint/mme/entities.yaml` | `cspb/structs.proto` | 实体结构（引用 Manager） |
| `blueprint/netwall/Core.yaml` | `cspb/request.proto`, `cspb/notify.proto`, `cspb/structs.proto` | 核心网络墙 |
| `blueprint/netwall/mme.yaml` | `cspb/request.proto`, `cspb/notify.proto` | MME 网络墙 |
| `blueprint/netwall/Sample.yaml` | `cspb/request.proto`, `cspb/notify.proto` | 示例网络墙 |

### 2. 包名与 go_package

- 所有 Proto 统一使用 `package pb;`
- 统一 `option go_package = "gitee.com/orbit-w/orbit/app/proto/pb";`

### 3. 网络墙消息转换

- NetWall 的 `Requests` 全部生成在 `message Request` 内部；`Notifies` 生成在 `message Notify` 内部。
- 简单请求允许无参数；有 `Rsp` 的请求在该请求的内部定义 `message Rsp`。

示例（来自 `Core.yaml`）：
```protobuf
message Request {
  message SearchBook {
    string Query = 1;
    int32  PageNumber = 2;
    message Rsp { Book Result = 1; }
  }
  message HeartBeat {}
}
```

### 4. MME 请求/通知转换与 Loc 规则

- 从 Mechanism 的 `Requests/Notifies` 派生出的对外接口，最终体现在 `netwall/mme.yaml` 中。
- 为保持一致性，网络层接口中的 `Core.MMELocation Loc` 使用保留编号 `1000`；若 YAML 已显式编写 Loc（如本仓库的 `netwall/mme.yaml`），应保证其为 `1000`。若未显式声明，生成器会自动注入。

示例（来自当前 YAML 的 MME 网络墙）：
```protobuf
message Request {
  message AskLevelUp {
    int32 UpNum = 1;
    Core.MMELocation Loc = 1000;
    message Rsp { string Result = 1; }
  }
}
message Notify  { message ExpChange { int32 Exp = 1; Core.MMELocation Loc = 1000; } }
```

### 5. 数据结构转换（MME）

- 除 NetWall 外，MME 相关结构体的普通字段均生成为 `optional`（`proto3` 语义下的可选）以提升演进兼容性；`map` 与 `repeated` 不使用 `optional`。
- 对于 `xmap<key, value> Field = id`，具体操作步骤：
-   1.按照范式XXXChange_{Field}，生成MessageName
-   1.生成 message MessageName , 包含三个字段：
-       1) common.ChangeType ChangeType = 1;
-       2) 字段名称是Key，类型根据xmap中指定的key类型设置。
-       3) 字段名称是Value，类型根据xmap中指定的value类型设置。
-   3.在结构体末尾生成对应的变化记录字段：`repeated MessageName MessageName = 1000 + id;`。
- 对于Entity对象，自动生成唯一Id字段，类型是int64： `int64 XXXId = 10000;`
- 对于Module/Manager/Mechanism对象，自动生成唯一Id字段，类型是int64： `int64 XXXId = 10000;`
- 生成MME中Entity/Module/Manager/Mechanism四种结构体时，DirtyBit 标记位，默认 DirtyXXXIdBit int64 = 1 << 0. 其他字段根据MME yaml中定义的FieldIndex进行左移。

示例1：
```yaml
#Player实例
Entity:
  - PlayerEntity:
      #英雄模块
      HeroManager HeroManager: 1
```
```protobuf
syntax = "proto3";

package entities;
option go_package = "./mme";

import "protocol/managers.proto";
// XXXId 是根据PlayerEntity结构体中Id字段自动化生成的，结构模式固定，不要修改。
message PlayerEntity {
  managers.HeroManager HeroManager = 1;

  int64 XXXId = 10000;
}
```
示例2:
```yaml
---
Mechanisms:
  #英雄机制
  - HeroMechanism:
      #英雄配置ID
      int32 ConfId: 1 [blueprint:"access=all"]
      #玩家获得英雄的时间
      int64 CreateTime: 2
      #英雄被使用次数
      int32 UseTimes: 3
      #技能
      xmap<int32, int32> Skills: 4 [blueprint:"access=all"]
---
```
```protobuf
syntax = "proto3";

package mechanisms;
option go_package = "./mme";

import "protocol/common.proto";

// 英雄机制
// XXXId 是根据HeroMechanism结构体中Id字段自动化生成的。结构模式固定，不要修改。
message HeroMechanism {
  optional int64 Id = 1;         // 英雄实例唯一Id
  optional int32 ConfId = 2;     // 英雄配置ID
  optional int64 CreateTime = 3; // 玩家获得英雄的时间
  optional int32 UseTimes = 4;   // 英雄被使用次数
  map<int32, int32> Skills = 5; // 技能

  repeated XXXChange_Skills XXXChange_Skills = 1005; // 技能变化

  int64 XXXId = 10000; // 自动化生成MechanismId，范式，不可修改。
}

//XXXChange_{Skills} 是根据HeroMechanism结构体中Skills字段自动化生成的。结构模式固定，不要修改。
message XXXChange_Skills {
  common.ChangeType ChangeType = 1;
  int32 Key = 2;
  int32 Value = 3; 
}
```


### 6. 导入语句

- `request.proto`、`notify.proto` 需要导入 `structs.proto` 与 `common.proto`
- `structs.proto` 需要导入 `actor.proto`（如使用 PID 类型）
- `common.proto` 无需导入其他文件

### 7. 注释保留

- YAML 中 `#` 注释转换为 Proto 中 `//` 注释，并尽量保留层级。

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

- 来自 `blueprint/netwall/Core.yaml` 的请求：
```yaml
NetWall:
  Name: Core
  Requests:
    - SearchBook:
        string Query: 1
        int32  PageNumber: 2
      Rsp:
        Book Result: 1
    - HeartBeat
```
对应 Proto：
```protobuf
message Request {
  message SearchBook {
    string Query = 1;
    int32  PageNumber = 2;
    message Rsp { Book Result = 1; }
  }
  message HeartBeat {}
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
```protobuf
message Request { message AskLevelUp { int32 UpNum = 1; Core.MMELocation Loc = 1000; message Rsp { string Result = 1; } } }
message Notify  { message ExpChange { int32 Exp = 1; Core.MMELocation Loc = 1000; } }
```

以上内容已与当前 `mme` 目录下的 YAML 定义对齐，并引入 Manager 层级，确保从 YAML 到 Proto 的生成规则清晰一致。
