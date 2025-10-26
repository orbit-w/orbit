# XMapLink - 通用的xmap管理类

## 简介

`XMapLink` 是一个通用的管理类，用于自动处理 protobuf map 和包装对象之间的同步、脏标记传递以及 Link/Unlink 管理。

### 解决的问题

在使用 xmap 管理 Mechanism/Module/Manager 时，需要手动处理大量重复的逻辑：

1. ❌ 手动维护两个 map（protobuf map 和包装对象 map）
2. ❌ 手动创建和初始化 MapAccessor
3. ❌ 手动 Link/Unlink 每个包装对象
4. ❌ 手动传递脏标记
5. ❌ Set 时需要记得先 Unlink 旧对象
6. ❌ 手动调用 TrackSetWithDelete（引用类型）

### 提供的解决方案

✅ **自动化管理**：一行代码完成所有初始化  
✅ **自动同步**：protobuf map 和包装对象 map 自动保持一致  
✅ **自动 Link/Unlink**：对象生命周期自动管理  
✅ **自动脏标记传递**：修改子对象时自动标记父对象为脏  
✅ **类型安全**：使用泛型确保编译时类型检查  
✅ **通用性强**：适用于所有 Manager/Module/Mechanism 场景  

## 核心功能

### 1. 自动管理包装对象的生命周期

```go
// XMapLink 自动处理：
// - 创建包装对象
// - Link 到父 DirtyTracker
// - LinkFactsAccessor 到 xmap
// - 在对象不再需要时 Unlink
link := xmaplink.NewXMapLinkWithParent(...)
```

### 2. 脏标记自动传递

```go
// 获取包装对象
hero, _ := link.Get(heroId)

// 修改包装对象，脏标记自动传递给父节点
hero.LevelUp.SetCurLevel(10)
// 此时 HeroModule、HeroManager 都被自动标记为脏
```

### 3. 智能 Set 操作

```go
// Set 时自动处理：
// - Unlink 旧对象（如果存在）
// - 创建新对象并 Link
// - 调用 TrackSetWithDelete（引用类型）
// - 更新两个 map
// - 标记脏位
hero := link.Set(heroId, heroModulePb)
```

## 快速开始

### 第一步：确保包装对象实现 Linkable 接口

所有基于 `dirtyflag.IDirtyFlag` 的对象自动实现此接口：

```go
type HeroModule[FactsAccessorKey comparable] struct {
    mme *mme.HeroModule
    dirtyflag.IDirtyFlag[FactsAccessorKey]
    // ... 其他字段
}
```

### 第二步：在 Manager 中使用 XMapLink

```go
type HeroManager[FactsAccessorKey comparable] struct {
    heroManager *mme.HeroManager
    dirtyflag.IDirtyFlag[FactsAccessorKey]
    
    // 使用 XMapLink 管理 Hero map
    heroMapLink *xmaplink.XMapLink[int64, *mme.HeroModule, *HeroModule[int64]]
}
```

### 第三步：初始化 XMapLink

```go
func NewHeroManager[FactsAccessorKey comparable](pt *mme.HeroManager) *HeroManager[FactsAccessorKey] {
    m := &HeroManager[FactsAccessorKey]{
        heroManager: pt,
        IDirtyFlag:  dirtyflag.NewDirtyFlag[FactsAccessorKey](),
    }
    
    // 一行代码完成所有初始化
    m.heroMapLink = xmaplink.NewXMapLinkWithParent(
        &m.heroManager.HeroMap,     // protobuf map 引用
        m.GetDirtyTracker(),         // 父 DirtyTracker
        HeroManagerDirtyHeroMapBit,  // 脏标记位
        NewHeroModule[int64],        // 包装器工厂函数
        true,                        // 引用类型（使用 TrackSetWithDelete）
    )
    
    return m
}
```

### 第四步：使用简化的 API

```go
// Get - 获取包装对象
hero, ok := m.heroMapLink.Get(heroId)

// Set - 设置/添加对象
hero := m.heroMapLink.Set(heroId, heroModulePb)

// Delete - 删除对象
deleted := m.heroMapLink.Delete(heroId)

// Range - 遍历所有对象
m.heroMapLink.Range(func(id int64, hero *HeroModule[int64]) bool {
    // 处理每个 hero
    return true // 返回 false 停止遍历
})

// Has - 检查是否存在
exists := m.heroMapLink.Has(heroId)

// Len - 获取数量
count := m.heroMapLink.Len()

// Clear - 清空所有对象
m.heroMapLink.Clear()
```

## 代码对比

### 重构前（手动管理）

```go
type HeroManager[FactsAccessorKey comparable] struct {
    heroManager     *mme.HeroManager
    IDirtyFlag      dirtyflag.IDirtyFlag[FactsAccessorKey]
    heroMapAccessor xmap.MapAccessor[int64, *HeroModule[int64]]
    heroMapLinked   map[int64]*HeroModule[int64]
}

func NewHeroManager[...](pt *mme.HeroManager) *HeroManager[...] {
    m := &HeroManager[...]{
        heroManager:   pt,
        IDirtyFlag:    dirtyflag.NewDirtyFlag[...](),
        heroMapLinked: make(map[int64]*HeroModule[int64]),
    }
    
    // 手动初始化 MapAccessor
    heroMap := m.heroManager.HeroMap
    m.heroMapAccessor = xmap.NewMapAccessorWithMarkerForRef[...](
        &heroMap, m.GetDirtyTracker(), HeroManagerDirtyHeroMapBit)
    
    // 手动初始化并 Link 每个对象
    if m.heroManager.HeroMap != nil {
        for key, heroModulePb := range m.heroManager.HeroMap {
            heroModule := NewHeroModule[int64](heroModulePb)
            heroModule.Link(m.GetDirtyTracker(), HeroManagerDirtyHeroMapBit)
            heroModule.LinkFactsAccessor(key)
            m.heroMapLinked[key] = heroModule
        }
    }
    return m
}

func (m *HeroManager[...]) SetHero(id int64, heroModulePb *mme.HeroModule) *HeroModule[int64] {
    // 手动处理旧对象的 Unlink
    if oldHero, exists := m.heroMapLinked[id]; exists {
        oldHero.Unlink()
    }
    
    // 手动创建新对象并 Link
    heroModule := NewHeroModule[int64](heroModulePb)
    heroModule.Link(m.GetDirtyTracker(), HeroManagerDirtyHeroMapBit)
    heroModule.LinkFactsAccessor(id)
    
    // 手动更新两个 map
    m.heroManager.HeroMap[id] = heroModulePb
    m.heroMapLinked[id] = heroModule
    
    // 手动标记脏位
    m.heroMapAccessor.Set(id, heroModulePb)
    
    return heroModule
}
```

### 重构后（使用 XMapLink）

```go
type HeroManager[FactsAccessorKey comparable] struct {
    heroManager *mme.HeroManager
    IDirtyFlag  dirtyflag.IDirtyFlag[FactsAccessorKey]
    heroMapLink *xmaplink.XMapLink[int64, *mme.HeroModule, *HeroModule[int64]]
}

func NewHeroManager[...](pt *mme.HeroManager) *HeroManager[...] {
    m := &HeroManager[...]{
        heroManager: pt,
        IDirtyFlag:  dirtyflag.NewDirtyFlag[...](),
    }
    
    // 一行代码完成所有初始化
    m.heroMapLink = xmaplink.NewXMapLinkWithParent(
        &m.heroManager.HeroMap,
        m.GetDirtyTracker(),
        HeroManagerDirtyHeroMapBit,
        NewHeroModule[int64],
        true,
    )
    
    return m
}

func (m *HeroManager[...]) SetHero(id int64, heroModulePb *mme.HeroModule) *HeroModule[int64] {
    // 一行代码自动完成所有操作
    return m.heroMapLink.Set(id, heroModulePb)
}
```

### 效果

- ✅ **代码量减少 60%**
- ✅ **复杂度降低**：消除所有手动 Link/Unlink 逻辑
- ✅ **安全性提升**：消除人为错误的可能性
- ✅ **可维护性提升**：统一的管理逻辑

## API 参考

### 构造函数

#### NewXMapLinkWithParent（推荐）

```go
func NewXMapLinkWithParent[K comparable, PbValue any, WrapperValue Linkable[K]](
    pbMap *map[K]PbValue,           // protobuf map 引用
    parentTracker *dt.DirtyTracker, // 父 DirtyTracker
    parentBit int64,                // 脏标记位
    wrapperFactory WrapperFactory[PbValue, WrapperValue], // 包装器工厂
    isRefType bool,                 // 是否引用类型
) *XMapLink[K, PbValue, WrapperValue]
```

最常用的构造方式，自动设置父节点。

#### NewXMapLinkForRef

```go
func NewXMapLinkForRef[K comparable, PbValue any, WrapperValue Linkable[K]](
    pbMap *map[K]PbValue,
    marker xmap.DirtyMarker,
    dirtyBit int64,
    wrapperFactory WrapperFactory[PbValue, WrapperValue],
) *XMapLink[K, PbValue, WrapperValue]
```

专门用于引用类型（指针）的快捷方式，自动设置 `isRefType=true`。

#### NewXMapLinkForValue

```go
func NewXMapLinkForValue[K comparable, PbValue any, WrapperValue Linkable[K]](
    pbMap *map[K]PbValue,
    marker xmap.DirtyMarker,
    dirtyBit int64,
    wrapperFactory WrapperFactory[PbValue, WrapperValue],
) *XMapLink[K, PbValue, WrapperValue]
```

专门用于值类型的快捷方式，自动设置 `isRefType=false`。

### 核心方法

| 方法 | 说明 | 自动处理 |
|------|------|---------|
| `Get(key K) (WrapperValue, bool)` | 获取包装对象 | 返回已 Link 的对象 |
| `Set(key K, pbValue PbValue) WrapperValue` | 设置/添加对象 | Unlink 旧对象、创建新对象、Link、更新 map、标记脏位 |
| `Delete(key K) bool` | 删除对象 | Unlink、删除、标记脏位 |
| `Has(key K) bool` | 检查是否存在 | - |
| `Len() int` | 返回长度 | - |
| `Range(f func(K, WrapperValue) bool)` | 遍历所有对象 | - |
| `Clear()` | 清空所有对象 | Unlink 所有对象、清空 map、标记脏位 |
| `Keys() []K` | 返回所有 key | - |
| `Values() []WrapperValue` | 返回所有包装对象 | - |

## 实现要求

### 包装对象必须实现 Linkable 接口

```go
type Linkable[K comparable] interface {
    Link(parent *dt.DirtyTracker, parentBit int64)
    LinkFactsAccessor(parent *dt.DirtyTracker, parentBit int64, 
                      factsAccessor xmap.FactsAccessor[K], key K)
    Unlink()
    GetDirtyTracker() *dt.DirtyTracker
}
```

**注意**：所有基于 `dirtyflag.IDirtyFlag` 的类型自动满足此接口。

### 包装器工厂函数

```go
type WrapperFactory[PbValue any, WrapperValue any] func(pb PbValue) WrapperValue
```

工厂函数应该：
- ✅ 只做对象创建，不要有副作用
- ✅ 返回正确类型的包装对象
- ✅ 保持简单和高效

示例：

```go
func NewHeroModule[FactsAccessorKey comparable](pt *mme.HeroModule) *HeroModule[FactsAccessorKey] {
    return &HeroModule[FactsAccessorKey]{
        mme:        pt,
        IDirtyFlag: dirtyflag.NewDirtyFlag[FactsAccessorKey](),
        // ... 初始化其他字段
    }
}
```

## 最佳实践

### 1. 优先使用 NewXMapLinkWithParent

除非有特殊需求，否则使用这个最便利的构造函数。

### 2. 明确指定引用类型标志

- **指针类型**（如 `*mme.HeroModule`）：`isRefType=true`
- **值类型**（如 `int32`, `string`）：`isRefType=false`

### 3. 工厂函数要简单

包装器工厂函数应该只做对象创建，不要有副作用。

### 4. 通过 Get 获取再修改

```go
// ✅ 正确
hero, _ := m.heroMapLink.Get(heroId)
hero.LevelUp.SetCurLevel(10)

// ❌ 错误：不要直接操作 protobuf 对象
m.heroManager.HeroMap[heroId].LevelUp.CurLevel = &level
```

### 5. 利用自动脏标记传递

通过 Get 获取的包装对象，修改时会自动传递脏标记，无需手动调用 `MarkDirty`。

## 性能考虑

1. **零额外分配**：使用接口避免闭包分配
2. **延迟初始化**：map 在首次使用时才创建
3. **批量操作支持**：提供 Range、Keys、Values 等批量操作方法
4. **类型安全**：编译时检查类型，运行时无反射开销

## 适用场景

✅ **适用于**：
- Manager 管理 Module map
- Module 管理 Mechanism map
- 任意层级的嵌套管理
- Value 是 Mechanism/Module/Manager 对象指针

❌ **不适用于**：
- Value 是简单值类型（如 `map[int32]int32`）
  - 简单值类型应直接使用 `xmap.MapAccessor`
- 不需要包装对象的场景

## 文件说明

- `xmap_link.go` - 核心实现
- `factory.go` - 便利构造函数
- `example_usage.md` - 详细使用指南
- `hero_manager_refactored_example.go` - 重构示例
- `README.md` - 本文档

## 许可证

继承项目许可证。

