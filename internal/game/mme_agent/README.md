# MME Logic Layer 使用指南

## 概述

MME Logic 层实现了 MME 架构文档中定义的 Agent 代理模式，提供了数据与逻辑分离、依赖注入和强类型访问能力。

## 架构层次

```
Controller (业务入口)
    ↓
Agent (代理层 - 自动生成)
    ↓
Manager Logic (管理器逻辑 - 用户实现)
    ↓
Module Logic (模块逻辑 - 用户实现)
    ↓
Mechanism Logic (机制逻辑 - 用户实现)
    ↓
Wrapper (数据层 - 自动生成)
```

## 核心组件

### 1. 接口定义 (`interfaces.go`)

- `IAgent`: Agent 代理接口，供 Logic 层跨模块访问
- `IMechanismLogic`: Mechanism Logic 基础接口
- `IModuleLogic`: Module Logic 基础接口
- `IManagerLogic`: Manager Logic 基础接口

### 2. 自动生成的接口 (`interfaces_gen.go`)

为每个 Mechanism、Module、Manager 自动生成对应的 Logic 接口，包含：
- Mechanism: Getter/Setter 方法
- Module: 访问内部 Mechanism Logic 的方法
- Manager: 访问 Module Logic 的方法（支持 xmap 和单例）

### 3. Factory 注册表 (`factory_registry.go`)

全局工厂注册表，用于创建 Logic 实例：
- `RegisterMechanismLogicFactory(name, factory)`
- `RegisterModuleLogicFactory(name, factory)`
- `RegisterManagerLogicFactory(name, factory)`

### 4. Agent (`/internal/game/agent/`)

为每个 Entity 自动生成的 Agent 类，提供：
- 强类型的 Manager Logic 访问方法
- 实现 `IAgent` 接口供 Logic 层跨模块访问

## 使用示例

### 1. 创建 Agent

```go
// 加载或创建 Entity
entity := mme.NewPlayerEntity()

// 序列化并加载到 Wrapper
data, _ := bson.Marshal(entity)
entityWrapper := mme.NewPlayerEntityWrapper()
entityWrapper.Load(data)
entityWrapper.InitFieldContext()

// 创建 Agent
playerAgent := agent.NewPlayerEntityAgent(entityWrapper)
```

### 2. 访问 Logic 层

```go
// 获取 Manager Logic
heroManager := playerAgent.GetHeroManagerLogic()

// 访问 Module Logic（xmap 字段）
heroLogic := heroManager.GetHeroMap(heroId)

// 访问 Mechanism Logic
baseLogic := heroLogic.GetBase()
levelUpLogic := heroLogic.GetLevelUp()

// 读写数据
baseLogic.SetConfId(100)
level := levelUpLogic.GetCurLevel()
```

### 3. 实现 Logic 类

#### Mechanism Logic 实现

```go
package mechanisms

import (
	"gitee.com/orbit-w/orbit/internal/game/mme"
	"gitee.com/orbit-w/orbit/internal/game/mme_logic"
)

func init() {
	// 注册工厂
	mme_logic.RegisterMechanismLogicFactory("HeroMechanism", func(wrapper any, agent mme_logic.IAgent) any {
		return NewHeroMechanismLogic(wrapper.(*mme.HeroMechanismWrapper), agent)
	})
}

type HeroMechanismLogicImpl struct {
	wrapper *mme.HeroMechanismWrapper
	agent   mme_logic.IAgent
}

func NewHeroMechanismLogic(wrapper *mme.HeroMechanismWrapper, agent mme_logic.IAgent) mme_logic.IHeroMechanismLogic {
	return &HeroMechanismLogicImpl{
		wrapper: wrapper,
		agent:   agent,
	}
}

// 实现接口方法
func (m *HeroMechanismLogicImpl) GetWrapper() any {
	return m.wrapper
}

func (m *HeroMechanismLogicImpl) GetAgent() mme_logic.IAgent {
	return m.agent
}

func (m *HeroMechanismLogicImpl) GetId() int64 {
	return m.wrapper.GetId()
}

func (m *HeroMechanismLogicImpl) SetId(id int64) {
	m.wrapper.SetId(id)
}

// 业务方法
func (m *HeroMechanismLogicImpl) AddSkill(skillId int32, level int32) {
	m.wrapper.GetSkillsAccessor().Set(skillId, level)
}
```

#### Module Logic 实现

```go
package modules

import (
	"gitee.com/orbit-w/orbit/internal/game/mme"
	"gitee.com/orbit-w/orbit/internal/game/mme_logic"
	"gitee.com/orbit-w/orbit/internal/game/mme_logic/mechanisms"
)

func init() {
	mme_logic.RegisterModuleLogicFactory("HeroModule", func(wrapper any, agent mme_logic.IAgent) any {
		return NewHeroModuleLogic(wrapper.(*mme.HeroModuleWrapper), agent)
	})
}

type HeroModuleLogicImpl struct {
	wrapper *mme.HeroModuleWrapper
	agent   mme_logic.IAgent

	// 直接持有所有 Mechanism Logic
	baseLogic         mme_logic.IHeroMechanismLogic
	levelUpLogic      mme_logic.ILevelUpMechanismLogic
	talentUnlockLogic mme_logic.IManualUnlockMechanismLogic
	skinWearLogic     mme_logic.IWearMechanismLogic
}

func NewHeroModuleLogic(wrapper *mme.HeroModuleWrapper, agent mme_logic.IAgent) mme_logic.IHeroModuleLogic {
	impl := &HeroModuleLogicImpl{
		wrapper: wrapper,
		agent:   agent,
	}

	// 初始化所有 Mechanism Logic
	impl.baseLogic = mechanisms.NewHeroMechanismLogic(wrapper.GetBase(), agent)
	impl.levelUpLogic = mechanisms.NewLevelUpMechanismLogic(wrapper.GetLevelUp(), agent)
	impl.talentUnlockLogic = mechanisms.NewManualUnlockMechanismLogic(wrapper.GetTalentUnlock(), agent)
	impl.skinWearLogic = mechanisms.NewWearMechanismLogic(wrapper.GetSkinWear(), agent)

	return impl
}

func (m *HeroModuleLogicImpl) GetBase() mme_logic.IHeroMechanismLogic {
	return m.baseLogic
}

// 业务方法
func (m *HeroModuleLogicImpl) LevelUp(exp int32) {
	if impl, ok := m.levelUpLogic.(*mechanisms.LevelUpMechanismLogicImpl); ok {
		impl.AddExp(exp)
	}
}
```

#### Manager Logic 实现

```go
package managers

import (
	"gitee.com/orbit-w/orbit/internal/game/mme"
	"gitee.com/orbit-w/orbit/internal/game/mme_logic"
)

func init() {
	mme_logic.RegisterManagerLogicFactory("HeroManager", func(wrapper any, agent mme_logic.IAgent) any {
		return NewHeroManagerLogic(wrapper.(*mme.HeroManagerWrapper), agent)
	})
}

type HeroManagerLogicImpl struct {
	wrapper *mme.HeroManagerWrapper
	agent   mme_logic.IAgent

	// xmap 字段：map 缓存（懒加载）
	heroModuleLogics map[int64]mme_logic.IHeroModuleLogic

	// 单例字段：直接持有
	singleHeroModuleLogic mme_logic.IHeroModuleLogic
}

func NewHeroManagerLogic(wrapper *mme.HeroManagerWrapper, agent mme_logic.IAgent) mme_logic.IHeroManagerLogic {
	impl := &HeroManagerLogicImpl{
		wrapper:          wrapper,
		agent:            agent,
		heroModuleLogics: make(map[int64]mme_logic.IHeroModuleLogic),
	}

	// 初始化单例字段
	if wrapper.GetSingleHeroModule() != nil {
		factory := mme_logic.GetModuleLogicFactory("HeroModule")
		if factory != nil {
			impl.singleHeroModuleLogic = factory(wrapper.GetSingleHeroModule(), agent).(mme_logic.IHeroModuleLogic)
		}
	}

	return impl
}

// xmap 字段访问（懒加载）
func (m *HeroManagerLogicImpl) GetHeroMap(heroId int64) mme_logic.IHeroModuleLogic {
	if logic, exists := m.heroModuleLogics[heroId]; exists {
		return logic
	}

	// 从 Wrapper 获取并缓存
	var moduleWrapper *mme.HeroModuleWrapper
	m.wrapper.HeroMap_Range(func(key int64, wrapper *mme.HeroModuleWrapper) bool {
		if key == heroId {
			moduleWrapper = wrapper
			return false
		}
		return true
	})

	if moduleWrapper == nil {
		return nil
	}

	factory := mme_logic.GetModuleLogicFactory("HeroModule")
	if factory == nil {
		return nil
	}

	logic := factory(moduleWrapper, m.agent).(mme_logic.IHeroModuleLogic)
	m.heroModuleLogics[heroId] = logic
	return logic
}

// 单例字段访问（直接返回）
func (m *HeroManagerLogicImpl) GetSingleHeroModule() mme_logic.IHeroModuleLogic {
	return m.singleHeroModuleLogic
}
```

## 设计原则

1. **数据与逻辑分离**：数据存储在 Wrapper，逻辑实现在 Logic
2. **强类型访问**：Agent 直接持有强类型 Manager Logic 字段，零运行时开销
3. **结构对应**：Logic 层成员变量严格对应 Manager/Module 字段
   - xmap 字段 → map 缓存（懒加载）
   - 单例字段 → 直接持有（构造时创建）
4. **无锁设计**：单线程模型，无需考虑线程安全
5. **依赖注入**：通过 Agent 接口实现跨模块访问

## 测试

参见 `agent_test.go` 中的测试用例。

## 生成命令

```bash
go run tools/gotools/gen/main.go blueprintgen \
  --logic-output=internal/game/mme_logic \
  --agent-output=internal/game/agent
```

