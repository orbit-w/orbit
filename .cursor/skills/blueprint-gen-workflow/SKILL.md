---
name: blueprint-gen-workflow
description: 指导 AI 操作 orbit 项目的 blueprint_gen 代码生成工具链，包括 YAML 蓝图文件编写规范、三阶段编译流程（解析→符号表→引用解析）、代码生成输出结构、常见错误排查。当用户询问 blueprint_gen、blueprintgen、MME YAML、蓝图生成、Proto 生成、Go Wrapper 生成、符号表、引用解析、FieldType、ResolvedType、Definition 接口 等相关问题时自动激活。
---

# Blueprint Gen Workflow

blueprint_gen 是 orbit 项目的代码生成工具，从 YAML 蓝图文件自动生成：
- Protocol Buffer 文件（`protocol/protocol/`）
- MME Go Wrapper 代码（`internal/game/mme/`）
- 协议 ID 文件（`pkg/proto/pb/`）
- Router 路由代码（`internal/game/routers/`）

**工具入口**：`tools/gotools/gen/cmd/blueprint_gen/`

---

## 快速上手

### 运行命令

```bash
go run ./tools/gotools/gen/main.go blueprintgen \
  --blueprint-dir ../protocol/blueprint \
  --proto-output ../protocol/protocol \
  --go-output internal/game/mme \
  --protocol-ids-output pkg/proto/pb \
  --controller-dir internal/game/controller_v2 \
  --router-output internal/game/routers/routers.go \
  --debug
```

### Blueprint 目录结构

```
protocol/blueprint/
├── mme/
│   ├── headfile.yaml      # 公共数据结构和枚举
│   ├── entities.yaml      # Entity 定义 + Register
│   ├── manager.yaml       # Manager 定义
│   ├── modules.yaml       # Module 定义
│   └── mechanisms.yaml    # Mechanism 定义 + Register
└── netwall/
    ├── hero.yaml          # NetWall 消息定义
    └── bag.yaml
```

---

## 三阶段编译流程

```
[Pass 1: Parsing]        解析所有 YAML 文件，生成未链接的 AST
        ↓
[Pass 2: Symbol Table]   BuildSymbolTable() — 注册所有定义到符号表
        ↓
[Pass 3: Reference]      ResolveReferences() — 将 TypeName 绑定到 Definition
        ↓
[Field Validation]       FiledChecker.Check() — 依赖 ResolvedType 验证字段组合规则
        ↓
[Code Generation]        生成 Proto / Go / Router 代码
```

> **关键约束**：`FiledChecker.Check()` 必须在 Pass 2+3 之后执行，否则 `ResolvedType` 为 nil 会导致 panic。

---

## YAML 文件编写规范

### 字段定义语法

```
<类型> <字段名>: <编号> [blueprint:"<选项>"]
```

**基础类型**：`int32` `int64` `uint32` `uint64` `float` `double` `bool` `string` `bytes`

**集合类型**：
```yaml
map<int32, HeroModule> Heroes: 1       # 普通 map（不持久化 Wrapper）
xmap<int32, HeroModule> Heroes: 1      # 扩展 map（持久化 Wrapper，支持脏标记）
repeated int32 ItemIds: 2
```

**字段编号约束**：每个对象的字段编号上限为 `64`（`ObjectFieldNumberMax`）。

**`access` 权限注解**（影响增量同步范围）：

| 值 | 含义 |
|----|------|
| `access=all` | 服务端和客户端均同步 |
| `access=s` | 仅服务端可见，不同步到客户端 |
| `access=c` | 仅客户端同步，服务端不写入 |

```yaml
int64 Id: 1 [blueprint:"access=all"]
int64 CreateTime: 2 [blueprint:"access=s"]   # 服务端内部字段
```

### headfile.yaml — 公共数据结构

```yaml
DataStruct:
  - ItemData:
      int32 ItemId: 1
      int32 Count: 2

Enums:
  - ServiceZoneType:
      ServiceZoneTypeUnknown: '0 [Content:"未知"]'
      ServiceZoneTypePlay: '1 [Content:"逻辑服"]'
```

### entities.yaml — Entity 定义

```yaml
Entity:
  - PlayerEntity:
      HeroManager HeroMgr: 1
      BagManager BagMgr: 2

Register:
  - PlayerEntity: 1      # 生成 EntityType 枚举，编号不能为 0
```

> Entity 的字段只能引用 **Manager** 类型。

### manager.yaml — Manager 定义

```yaml
Managers:
  - HeroManager:
      xmap<int32, HeroModule> Heroes: 1   # Value 只能是 Module
      HeroModule SingleHero: 2             # 直接引用 Module 也可
```

> Manager 字段规则：`map`/`xmap` 的 Value 必须是 Module；直接引用也必须是 Module。

### modules.yaml — Module 定义

```yaml
Modules:
  - HeroModule:
      - HeroMechanism HeroData: 1          # 只能引用 Mechanism
      - LevelUpMechanism LevelData: 2
```

> Module 字段只能引用 **Mechanism** 类型。

### mechanisms.yaml — Mechanism 定义

```yaml
Mechanisms:
  - HeroMechanism:
      HeroMechanism:
        int32 HeroId: 1
        string HeroName: 2
        int32 Level: 3

Register:
  - HeroMechanism: 1     # 生成 MechanismType 枚举
```

> Mechanism 字段只能是**基础类型**（不能引用其他 MME 对象）。

### netwall YAML — 网络消息

```yaml
NetWall:
  Name: Hero
  Requests:
    - GetHeroInfo:
        int32 HeroId: 1
        Response:
          HeroData HeroInfo: 1
  Notifies:
    - HeroLevelUp:
        int32 HeroId: 1
        int32 NewLevel: 2
```

---

## 字段演变约束（协议安全规则）

> **这是最容易破坏线上数据兼容性的规则，AI 必须严格遵守。**

1. **字段编号一旦上线，永久不可更改、不可复用**（即使删除了该字段，其编号也不能被新字段占用）
2. **只允许在末尾追加新字段**，禁止在中间插入
3. 删除字段：可以删除，但编号空洞保留，后续字段不填补该编号
4. YAML 与生成代码、Proto 文件保持一一对应，蓝图变更必须保持向后兼容

```yaml
# ✅ 正确：末尾追加
HeroMechanism:
  int64 Id: 1
  int32 ConfId: 2
  int32 Level: 3      # 新字段追加在末尾

# ❌ 错误：在中间插入（破坏已有编号的语义）
HeroMechanism:
  int64 Id: 1
  int32 NewField: 2   # 原来 ConfId=2 被挤走
  int32 ConfId: 3
```

---

## FieldIndex / DirtyBit 命名规则

生成的常量命名规范如下：

**命名格式**：
- `FieldIndex`：`{ObjectName}FieldIndex{FieldName}`
- `DirtyBit`：`{ObjectName}Dirty{FieldName}Bit`

**起始值规则**：

| 对象类型 | FieldIndex 起始 | 备注 |
|---------|----------------|------|
| **Entity** | Id 固定为 `0`，其他字段从 `1` 开始 | 自动生成的 `XXXId` 占用 index=0 |
| **Manager / Module / Mechanism** | 从 `0` 开始，按 YAML 字段顺序递增 | |

```go
// Mechanism（从 0 开始）
const (
    HeroMechanismFieldIndexId         = uint8(0)
    HeroMechanismFieldIndexConfId      // 1
    HeroMechanismFieldIndexLevel       // 2
)
const (
    HeroMechanismDirtyIdBit    int64 = 1 << HeroMechanismFieldIndexId
    HeroMechanismDirtyConfIdBit int64 = 1 << HeroMechanismFieldIndexConfId
)

// Entity（Id=0，其他字段从 1 开始）
const (
    PlayerEntityFieldIndexXXXId      = uint8(0)  // 自动生成的唯一 Id
    PlayerEntityFieldIndexHeroMgr    // 1（对应 YAML 第一个字段）
    PlayerEntityFieldIndexBagMgr     // 2
)
```

---

## Entity 自动生成字段

每个 Entity 会**自动生成**一个唯一 ID 字段，无需在 YAML 中手写：

```protobuf
// 自动插入，不在 YAML 中定义
int64 XXXId = 10000;   // 字段编号固定为 10000，FieldIndex = 0
```

- 变量名：`XXXId`（`XXX` 为 Entity 名前缀，如 `PlayerEntity` → `PlayerId`）
- 该字段**不使用 optional**
- FieldIndex 固定为 `0`，其他 YAML 字段从 `1` 开始编号

---

## MME 对象组合规则（核心约束）

```
Entity
 └── Manager (字段只能是 Manager)
      └── Module (直接引用 或 map/xmap<Key, Module>)
           └── Mechanism (字段只能是 Mechanism)
                └── 基础类型字段 (不能引用 MME 对象)
```

**命名约定**（用于 isMMEObjectType 自动识别）：
- Entity 后缀：`Entity`（如 `PlayerEntity`）
- Manager 后缀：`Manager`（如 `HeroManager`）
- Module 后缀：`Module`（如 `HeroModule`）
- Mechanism 后缀：`Mechanism`（如 `HeroMechanism`）

---

## 新增 MME 对象 Checklist

新增一个完整的 Mechanism（以 HeroMechanism 为例）：

1. **mechanisms.yaml** — 定义字段（基础类型），加入 Register
2. **modules.yaml** — 在对应 Module 中引用新 Mechanism
3. **（可选）manager.yaml** — 如需新 Module，在 Manager 中引用
4. **（可选）entities.yaml** — 如需新 Manager，在 Entity 中引用
5. **运行 blueprintgen** — 重新生成代码
6. **实现 Logic 层** — 在 `internal/game/mme_logic/` 实现 `MechanismLogicImpl`（必须实现 `OnLoad/OnSave/OnLogin/OnLogout`）和（可选）`ManagerLogicImpl`；`EntityImpl`/`ManagerAgentImpl`/`ModuleAgentImpl` 已由 blueprintgen 自动生成，无需手动编写

---

## 添加新代码生成器 Checklist

1. 实现 `Definition` 接口（`GetName`, `GetKind`, `GetFile`, `GetFields`）
2. 在 `BuildSymbolTable()` 中注册新类型（使用正确的 scope）
3. 先 `ResolveReferences()` 再 `FiledChecker.Check()`
4. 在 `DefinitionKind` 中添加对应枚举值

---

## 常见错误速查

| 错误信息 | 原因 | 解决方案 |
|---------|------|---------|
| `panic: manager XXX map field YYY value type is not mme object` | FiledChecker 在 Pass 2/3 之前执行 | 确保调用顺序：BuildSymbolTable → ResolveReferences → Check |
| `duplicate definition: symbol 'XXX' already defined` | 同一符号注册了两次 | 检查 Response 是否用了基础名称而非完整名称 |
| `undefined type reference: field 'YYY' references unknown type 'XXX'` | 类型 XXX 未注册到符号表 | 检查该类型是否在解析时加入了对应集合，并在 BuildSymbolTable 中注册 |
| `Register 中定义的 Entity 'XXX' 不存在` | entities.yaml 中 Register 引用了未定义的 Entity | 先在 Entity 部分定义，再加入 Register |
| `field number > 64` | 字段编号超出 ObjectFieldNumberMax | 将编号保持在 1-64 范围内 |

---

## 参考资料

- 详细 API 参考：[reference.md](reference.md)
- YAML 完整示例：[examples.md](examples.md)
- Agent 层代码生成器实现：skill `blueprint-gen-agent`
- Logic Layer 完整架构设计：`README.md § MME-Agent 逻辑开发架构 (Logic Layer)`
- 关键源码位置：`tools/gotools/gen/cmd/blueprint_gen/`
