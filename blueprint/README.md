# NetBluePrint 设计文档

## 概述

NetBluePrint 是一个基于 YAML 的网络协议设计系统，用于定义游戏中的网络消息与数据结构。通过 YAML 蓝图文件，可自动生成 Protocol Buffers (.proto) 文件，实现网络协议的标准化与自动化管理。

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
