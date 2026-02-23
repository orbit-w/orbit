---
name: blueprint-gen-agent
description: 指导 AI 维护和扩展 blueprint_gen 的 Agent 层代码生成器（agent_gen/ 包）。生成目标包括：EntityImpl、ManagerAgentImpl、ModuleAgentImpl，以及 mme_agent/imodels.go 接口文件和静态框架文件。mme_logic 代码（Logic 接口 + 实现脚手架）不在自动生成范围内，由用户手写或后续单独工具负责。当用户询问 mme_agent_gen、agent_gen、agent 代码生成、EntityImpl 生成、ManagerAgentImpl 生成、ModuleAgentImpl 生成 等相关问题时自动激活。
---

# Blueprint Gen — Agent 层代码生成器

## 实现状态

`agent_gen/` 包**已实现**，位于：

```
tools/gotools/gen/cmd/blueprint_gen/agent_gen/
```

生成架构三层代码：

```
[已有] go_wrapper_gen.go  →  internal/game/mme/          (数据层 Wrapper)
[已有] agent_gen/         →  internal/game/mme_agent/    (代理层，自动生成)
```

---

## 源文件结构

```
tools/gotools/gen/cmd/blueprint_gen/agent_gen/
├── agent_gen.go        ← AgentGenerator 主协调器
├── entity_gen.go       ← EntityImpl 生成逻辑
├── manager_gen.go      ← ManagerAgentImpl 生成逻辑
├── module_gen.go       ← ModuleAgentImpl 生成逻辑
├── imodels_gen.go      ← Agent 接口文件生成（imodels.go）
├── static_gen.go       ← 静态框架文件（ent_factory.go、life_cycle_ext.go）
└── utils.go            ← 工具函数与导入路径常量
```

> **注意**：`logic_iface_gen.go`（曾用于生成 `mme_logic/` 下的 Logic 接口与实现脚手架）已从自动生成范围移除。`mme_logic/` 目录下的所有代码由用户手写或后续单独工具负责，**不在本生成器职责范围内**。

---

## 生成目标：五类文件

| 源文件 | 输入 | 输出文件 | 输出内容 |
|--------|------|----------|---------|
| `static_gen.go` | — | `mme_agent/ent_factory.go` | `IEntity` 接口 + 工厂注册表 |
| `static_gen.go` | — | `mme_agent/life_cycle_ext.go` | `CallOnLoad/Save/Login/Logout` 鸭子类型帮助函数 |
| `imodels_gen.go` | `Managers`, `Modules` | `mme_agent/imodels.go` | `IBaseAgent` + `I<Manager>Agent` + `I<Module>Agent` |
| `entity_gen.go` | `Entities` | `mme_agent/<entity>_impl.go` | `<Entity>Impl` 结构体 + 懒加载 + 生命周期分发 |
| `manager_gen.go` | `Managers` | `mme_agent/<manager>_agent_impl.go` | `<Manager>AgentImpl` 结构体 + 容器 + Logic 委托 |
| `module_gen.go` | `Modules` | `mme_agent/<module>_agent_impl.go` | `<Module>AgentImpl` 结构体 + `Get<Mech>Logic()` |

> `mme_logic/` 目录（Logic 接口 + 实现脚手架）**不在自动生成范围**，由用户手写或后续单独工具负责。

---

## 入口：AgentGenerator

```go
// tools/gotools/gen/cmd/blueprint_gen/agent_gen/agent_gen.go

type AgentGenerator struct {
    entities    []*mmeobj.Entity
    managers    []*mmeobj.Manager
    modules     []*mmeobj.Module
    mechanisms  []*mmeobj.Mechanism
    symbolTable *types.SymbolTable
}

func NewAgentGenerator(
    entities []*mmeobj.Entity,
    managers []*mmeobj.Manager,
    modules  []*mmeobj.Module,
    mechanisms []*mmeobj.Mechanism,
    symbolTable *types.SymbolTable,
) *AgentGenerator

// agentOutput: mme_agent 目录
func (g *AgentGenerator) Generate(agentOutput string) error
```

---

## cmd.go 集成

`tools/gotools/gen/cmd/blueprint_gen/cmd.go` 已添加：

**新 CLI Flags（`InitCmd`）**：
```go
blueprintGenCmd.Flags().String("agent-output", "", "Output directory for agent files (e.g. internal/game/mme_agent)")
```

> `--logic-output` 标志已移除，`mme_logic/` 不再由本工具生成。

**调用时机**（在 Go 文件生成之后、Router 代码之前）：
```go
agentOutput, _ := cmd.Flags().GetString("agent-output")
if agentOutput != "" {
    agentGen := agent_gen.NewAgentGenerator(
        data.Entities, data.Managers, data.Modules, data.Mechanisms, data.SymbolTable,
    )
    if err := agentGen.Generate(agentOutput); err != nil {
        cmd.PrintErrln("Failed to generate agent files:", err)
        return
    }
    FormatGoFiles(agentOutput)
}
```

---

## 各层生成规则

### 1. entity_gen.go — `<Entity>Impl`

**结构体**（通过 `field.Type.IsMMEObjectType() && def.GetKind() == types.DefKindManager` 筛选 Manager 字段）：

```go
type <Entity>Impl struct {
    entityWrapper *mmeobj.<Entity>Wrapper

    <managerField>Agent I<Manager>Agent  // 每个 Manager 字段生成一个缓存
}
```

**懒加载 Getter**：
```go
func (a *<Entity>Impl) Get<Field>Agent() I<Manager>Agent {
    if a.<managerField>Agent == nil {
        a.<managerField>Agent = New<Manager>Agent(a.entityWrapper.Get<Field>())
    }
    return a.<managerField>Agent
}
```

**生命周期**（OnLoad 调用 `entityWrapper.Load(raw)` 然后逐个 `CallOnLoad(Get<Field>Agent(), new)`）：
```go
func (a *<Entity>Impl) OnLoad(raw bson.Raw, new bool) error {
    if err := a.entityWrapper.Load(raw); err != nil { return err }
    if err := CallOnLoad(a.Get<Field>Agent(), new); err != nil { return err }
    return nil
}
// OnSave / OnLogin / OnLogout 模式相同（无 raw / new 参数）
```

**输出文件名**：`camelToSnake(<EntityName>) + "_impl.go"`，例：`player_entity_impl.go`

---

### 2. manager_gen.go — `<Manager>AgentImpl`

**字段分类**：
- `moduleFieldMap`：`field.Type.IsXMapField() || field.Type.IsMapField()` 且 value 是 Module
- `moduleFieldDirect`：`field.Type.IsMMEObjectType()` 且解析为 Module

**结构体**：
```go
type <Manager>AgentImpl struct {
    imodels.I<Manager>Logic               // 嵌入 Logic 接口（委托）
    wrapper *mme.<Manager>Wrapper

    // Map 字段 → ModuleMapContainer
    <field>Container *container.ModuleMapContainer[<Key>, *mme.<Module>, *mme.<Module>Wrapper, I<Module>Agent]

    // 直接字段 → 懒加载缓存
    <field>Agent I<Module>Agent
}
```

**构造函数**（同时初始化所有 Map 容器）：
```go
func New<Manager>Agent(wrapper *mme.<Manager>Wrapper) I<Manager>Agent {
    ins := &<Manager>AgentImpl{
        wrapper:        wrapper,
        I<Manager>Logic: managers.New<Manager>Logic(wrapper),
        <field>Container: container.NewModuleMapContainer(wrapper.Get<Field>(), New<Module>Logic),
    }
    return ins
}
```

**Map 查找方法**：
```go
func (m *<Manager>AgentImpl) <Field>_GetModule(key <KeyType>) (I<Module>Agent, bool) {
    return m.<field>Container.Get(key)
}
```

**直接 Module 懒加载 Getter**：
```go
func (m *<Manager>AgentImpl) Get<Field>() I<Module>Agent {
    if m.<field>Agent == nil {
        wrapper := m.wrapper.Get<Field>()
        if wrapper == nil { return nil }
        m.<field>Agent = New<Module>Logic(wrapper)
    }
    return m.<field>Agent
}
```

**生命周期**（Map 容器先 Range，直接 Module 再 Call）：
```go
func (m *<Manager>AgentImpl) OnLoad(new bool) error {
    var err error
    m.<field>Container.Range(func(key <Key>, agent I<Module>Agent) bool {
        if err = CallOnLoad(agent, new); err != nil { return false }
        return true
    })
    if err != nil { return err }
    if err := CallOnLoad(m.Get<DirectField>(), new); err != nil { return err }
    return nil
}
```

**输出文件名**：`camelToSnake(<Manager>) + "_agent_impl.go"`

---

### 3. module_gen.go — `<Module>AgentImpl`

**字段筛选**：`field.Type.IsMMEObjectType() && def.GetKind() == types.DefKindMechanism`

**结构体**：
```go
type <Module>AgentImpl struct {
    wrapper *mme.<Module>Wrapper

    <mechField>Logic imodels.I<Mechanism>Logic  // 每个 Mechanism 字段一个缓存
}
```

**构造函数**（无参初始化，懒加载）：
```go
func New<Module>Logic(wrapper *mme.<Module>Wrapper) I<Module>Agent {
    return &<Module>AgentImpl{wrapper: wrapper}
}
```

**Mechanism Logic 懒加载 Getter**：
```go
func (m *<Module>AgentImpl) Get<Mechanism>Logic() imodels.I<Mechanism>Logic {
    if m.<mechField>Logic == nil {
        m.<mechField>Logic = mechanisms.New<Mechanism>Logic(m.wrapper.Get<Field>())
    }
    return m.<mechField>Logic
}
```

**输出文件名**：`camelToSnake(<Module>) + "_agent_impl.go"`

---

### 4. imodels_gen.go — `mme_agent/imodels.go`

每次全量覆盖，包含：
- `IBaseAgent { GetWrapper() any }`
- 每个 Manager → `I<Manager>Agent { IBaseAgent; imodels.I<Manager>Logic }`
- 每个 Module → `I<Module>Agent { IBaseAgent; Get<Mech>Logic() imodels.I<Mech>Logic; ... }`

---

### 5. static_gen.go — 静态框架文件（每次覆盖）

**`ent_factory.go`**：
```go
type IEntity interface {
    GetId() int64
    GetEntityType() mme.EntityType
    GetEntityWrapper() mmeobj.IEntityWrapper
    OnLoad(raw bson.Raw, new bool) error
    OnSave() error; OnLogin() error; OnLogout() error
}
type EntityFactory func() IEntity
var mapEntityFactories = make(map[mme.EntityType]EntityFactory)
func RegisterEntityFactory(...)
func GetEntityFactory(...) EntityFactory
```

**`life_cycle_ext.go`**：鸭子类型生命周期帮助函数，通过接口断言调用：
```go
func CallOnLoad(agent IBaseAgent, new bool) error  // 断言 interface{ OnLoad(bool) error }
func CallOnSave(agent IBaseAgent) error
func CallOnLogin(agent IBaseAgent) error
func CallOnLogout(agent IBaseAgent) error
```

---

## 工具函数（utils.go）

```go
// 导入路径常量
const (
    mmeImportPath             = "gitee.com/orbit-w/orbit/internal/game/mme"
    protoMmeImportPath        = "gitee.com/orbit-w/orbit/pkg/proto/mme"
    imodelsImportPath         = "gitee.com/orbit-w/orbit/internal/game/mme_logic/imodels"
    managersLogicImportPath   = "gitee.com/orbit-w/orbit/internal/game/mme_logic/managers"
    mechanismsLogicImportPath = "gitee.com/orbit-w/orbit/internal/game/mme_logic/mechanisms"
    containerImportPath       = "gitee.com/orbit-w/orbit/lib/container"
    bsonImportPath            = "go.mongodb.org/mongo-driver/v2/bson"
)

func camelToSnake(s string) string   // 驼峰→蛇形（文件名）
func firstLower(s string) string     // 首字母小写（字段名）
func firstUpper(s string) string     // 首字母大写
func writeFile(path, content string) error  // 创建目录+写文件
func fileExists(path string) bool    // 检查文件是否存在（用于不覆盖逻辑）
```

---

## 字段类型判断速查

```go
// 筛选 Manager 字段（用于 Entity 生成）
field.Type.IsMMEObjectType() && field.Type.IsResolved() &&
    field.Type.GetResolvedDefinition().GetKind() == types.DefKindManager

// 筛选 Map/XMap Module 字段（用于 Manager 生成）
(field.Type.IsXMapField() || field.Type.IsMapField()) &&
    vt.IsMMEObjectType() && vt.IsResolved() &&
    vt.GetResolvedDefinition().GetKind() == types.DefKindModule

// 筛选直接 Module 字段（用于 Manager 生成）
field.Type.IsMMEObjectType() && field.Type.IsResolved() &&
    field.Type.GetResolvedDefinition().GetKind() == types.DefKindModule

// 筛选 Mechanism 字段（用于 Module 生成）
field.Type.IsMMEObjectType() && field.Type.IsResolved() &&
    field.Type.GetResolvedDefinition().GetKind() == types.DefKindMechanism

// Map 字段的 Key/Value 类型
field.Type.KeyKind().String()        // key 的 Go 类型字符串，如 "int64"
field.Type.GetValueType().GetName()  // value 的类型名，如 "HeroModule"
```

---

## 命名规则

| 输入（YAML 名） | 用途 | 示例 |
|----------------|------|------|
| `HeroManager` | 结构体/接口名前缀 | `HeroManagerAgentImpl`, `IHeroManagerAgent`, `IHeroManagerLogic` |
| `camelToSnake(name)` | 文件名 | `hero_manager_agent_impl.go` |
| `firstLower(name) + "Agent"` | Entity 中的 Manager 缓存字段 | `heroManagerAgent` |
| `firstLower(fieldName) + "Container"` | Manager 中的 Map 容器字段 | `heroMapContainer` |
| `firstLower(fieldName) + "Agent"` | Manager 中的直接 Module 缓存字段 | `singleHeroModuleAgent` |
| `firstLower(fieldName) + "Logic"` | Module 中的 Mechanism 缓存字段 | `baseLogic` |

---

## 实际输出目录结构

```
internal/game/
├── mme_agent/                           ← 本工具全量生成（每次覆盖）
│   ├── ent_factory.go               ← static_gen
│   ├── life_cycle_ext.go            ← static_gen
│   ├── imodels.go                   ← imodels_gen
│   ├── player_entity_impl.go        ← entity_gen
│   ├── hero_manager_agent_impl.go   ← manager_gen
│   └── hero_module_agent_impl.go    ← module_gen
└── mme_logic/                           ← 用户手写 / 后续单独工具生成，不在本工具范围
    ├── imodels/
    │   ├── base.go
    │   ├── i_hero_manager_logic.go
    │   └── i_hero_mechanism_logic.go
    ├── managers/
    │   └── hero_manager_logic_impl.go
    └── mechanisms/
        └── hero_mechanism_logic_impl.go
```

---

## 参考资料

- 已实现的生成器代码：`tools/gotools/gen/cmd/blueprint_gen/agent_gen/`
- 生成目标参考（手写版本）：`internal/game/mme_agent_v2/`
- Logic 层参考（用户手写）：`internal/game/mme_logic/`
- Wrapper 生成器（生成模式参考）：`tools/gotools/gen/cmd/blueprint_gen/go_wrapper_gen.go`
- Generator 接口与 CodeBuilder：`tools/gotools/gen/cmd/blueprint_gen/generator.go`
- BlueprintContext 定义：`tools/gotools/gen/cmd/blueprint_gen/context.go`
- 命令入口：`tools/gotools/gen/cmd/blueprint_gen/cmd.go`
