# Zone Meta Service

Zone Meta Service 是 Orbit 框架中用于管理游戏区服（Zone）元数据的核心组件。它负责维护区服的基础信息、调度策略以及与物理节点的绑定关系。

## 功能特性

- **元数据管理**：管理区服 ID、模式（Pattern）等基础信息。
- **自动节点绑定**：在创建元数据时，自动获取并绑定当前运行节点的身份信息（NodeID）。
- **分布式缓存**：基于 Redis 的分布式缓存机制，支持发布/订阅模式的缓存失效通知。
- **调度器集成**：集成 `ZoneDispatcher`，支持多种调度策略（世界服、指定节点、随机等）。

## 核心 API

### NewZoneMeta

创建一个新的 `ZoneMeta` 实例。该函数会自动检测当前运行环境，并通过 `dispatcher` 绑定当前节点的 ID。

```go
func NewZoneMeta(id string, pattern int32, dispatcher *ZoneDispatcher) *ZoneMeta
```

- **参数**:
    - `id`: 区服 ID。
    - `pattern`: 区服模式。
    - `dispatcher`: 调度器配置。
- **行为**:
    - 如果 `dispatcher` 不为空，且当前通过 `cluster.Manager` 能获取到节点信息，会自动将 `dispatcher.NodeId` 设置为当前节点的 ID。

### SetZoneMeta

创建并缓存 `ZoneMeta`。这是注册区服元数据的主要入口。

```go
func SetZoneMeta(id string, pattern int32, dispatcher *ZoneDispatcher) (*ZoneMeta, error)
```

- **流程**:
    1. 调用 `NewZoneMeta` 创建实例（包含自动节点绑定）。
    2. 将实例序列化并存储到 Redis 缓存中。

### GetZoneMeta

从缓存中获取区服元数据。

```go
func GetZoneMeta(id string) (*ZoneMeta, error)
```

## 数据结构

### ZoneMeta (Proto Definition)

```protobuf
message ZoneMeta {
  string Id = 1;              // 区服 ID
  int32 Pattern = 2;          // 模式
  ZoneDispatcher Dispatcher = 3; // 调度器信息
}
```

### ZoneDispatcher (Proto Definition)

```protobuf
message ZoneDispatcher {
  Zone_DispatcherType Type = 1; // 调度类型 (ForWorld, ForDesignated, ForRandom)
  int32 ServerId = 2;           // 逻辑服 ID
  string NodeId = 3;            // 节点 ID (自动绑定)
}
```

## 调度策略详解 (Dispatcher Policy)

`Zone_DispatcherType` 定义了 Zone 在集群中的分配和调度逻辑：

### 1. ForWorld (0) - 世界分组
- **含义**: 全局唯一或大世界分组模式。
- **适用场景**: 主城、公共区域、全服活动地图等需要全局统一或特定世界服务器承载的场景。
- **分配逻辑**: 
    - 优先使用 `Dispatcher.NodeId` 指定的节点（如果已设置）。
    - 如果未指定或指定节点不可用，则选择负载最低的在线节点。
    - 适合需要稳定性但允许故障转移的场景。

### 2. ForDesignated (1) - 指定节点
- **含义**: 强绑定模式，指定特定的物理节点或逻辑服。
- **适用场景**: 
    - 开发者调试（强制连某节点）。
    - 有状态服务的恢复（需回到上次所在的节点）。
    - 某些必须在特定硬件或配置节点上运行的 Zone。
- **分配逻辑**: 
    - 严格匹配 `Dispatcher.NodeId`。
    - 如果 `NodeId` 为空或节点不在线，返回错误，不会进行故障转移。
    - 在 `NewZoneMeta` 中，该模式常配合自动获取当前节点 ID 使用，确立"谁创建，谁负责"的关系。
    - `ZoneManager` 在加载 Zone 时会验证当前节点是否匹配，不匹配则拒绝加载。

### 3. ForRandom (2) - 随机/负载均衡
- **含义**: 动态分配模式。
- **适用场景**: 副本、战斗房间、临时活动地图等可以动态创建且无特定节点依赖的场景。
- **分配逻辑**: 
    - 系统从所有在线节点池中选择一个节点。
    - **默认策略**：负载均衡，根据节点负载选择负载最低的节点。
    - **负载计算公式**：`Load = ZoneCount * 1000 + EntityCount / 100 + CCU`
        - 优先考虑 ZoneCount（区服数量）
        - 其次考虑 EntityCount（实体数量）
        - 最后考虑 CCU（并发用户数）
    - 如果只有一个在线节点，直接返回该节点。

## 调度器实现 (Dispatcher Implementation)

### ZoneDispatcherService

调度器服务负责根据调度策略选择合适的节点。

```go
type ZoneDispatcherService struct {
    clusterMgr *cluster.Manager
}
```

#### 核心方法

**SelectNode**: 根据调度策略选择节点

```go
func (s *ZoneDispatcherService) SelectNode(dispatcher *ZoneDispatcher) (*cluster.Node, error)
```

**DispatchZone**: 为 Zone 分配节点并更新 dispatcher 的 NodeId

```go
func (s *ZoneDispatcherService) DispatchZone(zoneId string, dispatcher *ZoneDispatcher) (*cluster.Node, error)
```

#### 错误类型

- `ErrNoAvailableNode`: 没有可用节点
- `ErrNodeNotFound`: 指定节点不存在
- `ErrInvalidDispatcher`: 无效的调度器配置

### 集成到 ZoneManager

`ZoneManager` 在加载 Zone 时会自动进行节点验证：

1. 对于 `ForDesignated` 策略，验证当前节点是否是指定节点。
2. 如果验证失败，拒绝加载 Zone 并返回错误。
3. 这确保了 Zone 只在正确的节点上运行。

```go
// ZoneManager 结构
type ZoneManager struct {
    system     *actor.ActorSystem
    cache      cmap.ConcurrentMap[string, *actor.PID]
    exec       *unipue_task_exec.UniqueTaskExecutor
    dispatcher *zone_meta.ZoneDispatcherService  // 调度器服务
}
```

## 使用示例

### 示例 1: 创建随机分配的 Zone

```go
import (
    zone_meta "gitee.com/orbit-w/orbit/core/services/service_zone/meta"
)

func CreateRandomZone() {
    // 1. 准备调度器配置（随机/负载均衡模式）
    dispatcher := &zone_meta.ZoneDispatcher{
        Type: zone_meta.Zone_DispatcherType_ForRandom,
    }

    // 2. 注册元数据
    // NodeId 会在 NewZoneMeta 中自动填充为当前节点 ID
    zoneMeta, err := zone_meta.SetZoneMeta("dungeon_1001", 1, dispatcher)
    if err != nil {
        panic(err)
    }
    
    // zoneMeta.Dispatcher.NodeId 现在等于当前节点 ID
}
```

### 示例 2: 创建指定节点的 Zone

```go
func CreateDesignatedZone() {
    // 1. 准备调度器配置（指定节点模式）
    dispatcher := &zone_meta.ZoneDispatcher{
        Type:   zone_meta.Zone_DispatcherType_ForDesignated,
        NodeId: "node-001", // 明确指定节点
    }

    // 2. 注册元数据
    zoneMeta, err := zone_meta.SetZoneMeta("main_city", 1, dispatcher)
    if err != nil {
        panic(err)
    }
    
    // 该 Zone 只能在 node-001 上运行
}
```

### 示例 3: 使用调度器服务选择节点

```go
func DispatchZoneToNode() {
    // 1. 创建调度器服务
    dispatcherSvc := zone_meta.NewZoneDispatcherService()
    
    // 2. 准备调度器配置
    dispatcher := &zone_meta.ZoneDispatcher{
        Type: zone_meta.Zone_DispatcherType_ForRandom,
    }
    
    // 3. 为 Zone 分配节点
    node, err := dispatcherSvc.DispatchZone("battle_room_1001", dispatcher)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Zone assigned to node: %s (Address: %s)\n", node.ID, node.Address)
    
    // 4. 创建元数据（dispatcher.NodeId 已被更新）
    zoneMeta, err := zone_meta.SetZoneMeta("battle_room_1001", 1, dispatcher)
    if err != nil {
        panic(err)
    }
}
```

### 示例 4: 世界服模式

```go
func CreateWorldZone() {
    // 1. 准备调度器配置（世界服模式）
    dispatcher := &zone_meta.ZoneDispatcher{
        Type:   zone_meta.Zone_DispatcherType_ForWorld,
        NodeId: "world-node-001", // 可选：指定首选节点
    }

    // 2. 注册元数据
    // 如果 world-node-001 在线，优先使用它
    // 如果不在线，自动选择负载最低的节点（故障转移）
    zoneMeta, err := zone_meta.SetZoneMeta("world_map", 1, dispatcher)
    if err != nil {
        panic(err)
    }
}
```

## 最佳实践

1. **ForWorld**: 用于需要高可用性的全局服务，允许故障转移。
2. **ForDesignated**: 用于有状态服务或调试场景，确保 Zone 在特定节点运行。
3. **ForRandom**: 用于无状态、可动态创建的服务，实现负载均衡。
4. **节点验证**: `ZoneManager` 会自动验证节点匹配性，无需手动检查。
5. **缓存同步**: 元数据存储在 Redis 中，支持分布式环境下的数据共享。

## 架构图

```
┌─────────────────────────────────────────────────────────────┐
│                      Client Request                         │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                      ZoneManager                            │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  1. GetZone(zoneId)                                  │   │
│  │  2. Load ZoneMeta from Redis                         │   │
│  │  3. validateNodeForZone(zoneMeta)  ◄─────────────┐   │   │
│  │  4. Create Zone Actor                            │   │   │
│  └──────────────────────────────────────────────────┘   │   │
└────────────────────────────┬────────────────────────────┼───┘
                             │                            │
                             ▼                            │
┌─────────────────────────────────────────────────────────┼───┐
│              ZoneDispatcherService                      │   │
│  ┌──────────────────────────────────────────────────┐   │   │
│  │  SelectNode(dispatcher)                          │   │   │
│  │    ├─ ForWorld: selectWorldNode()                │   │   │
│  │    ├─ ForDesignated: selectDesignatedNode() ─────┼───┘   │
│  │    └─ ForRandom: selectRandomNode()              │       │
│  │         └─ selectLeastLoadedNode()                │       │
│  └──────────────────────────────────────────────────┘       │
└────────────────────────────┬────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────┐
│                   Cluster Manager                           │
│  - GetOnlineNodes()                                         │
│  - GetNode(nodeId)                                          │
│  - GetCurrentNode()                                         │
│                                                             │
│  Node Metrics:                                              │
│    - ZoneCount, EntityCount, CCU                            │
│    - Load = ZoneCount*1000 + EntityCount/100 + CCU          │
└─────────────────────────────────────────────────────────────┘
```

