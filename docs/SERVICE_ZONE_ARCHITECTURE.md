# 服务区架构技术文档

## 概述

服务区架构（Service Zone Architecture）是基于 MME（Entity → Manager → Module → Mechanism）四层数据模型和 ECS（Entity-Component-System）前端架构设计的服务端区域管理系统。该架构将游戏世界划分为多个服务区（Service Zone），每个服务区独立管理其内部的实体（Entity）生命周期、数据同步和消息路由，实现高效的分区管理和水平扩展。

---

## 核心概念

### 1. 服务区（Service Zone）

**定义**：服务区是游戏世界中的逻辑分区单元，每个服务区代表一个独立的游戏场景或区域（如副本、主城、野外地图等）。

**特性**：
- **独立性**：每个服务区拥有独立的 Actor 系统，管理其内部的 Entity
- **可扩展性**：支持动态创建和销毁，实现水平扩展
- **隔离性**：服务区之间的数据相互隔离，通过消息路由进行通信
- **生命周期管理**：完整的创建、运行、销毁生命周期

### 2. 服务区与 MME 架构的映射关系

```
服务区（Service Zone）
  └── Entity（MME Entity）
        └── Manager（MME Manager）
              └── Module（MME Module）
                    └── Mechanism（MME Mechanism）
```

**映射规则**：
- **服务区**：管理多个 Entity，每个 Entity 对应一个 Actor
- **Entity**：MME 顶层实体，包含一个或多个 Manager
- **Manager**：对应 ECS 中的 Component 集合管理器
- **Module**：对应 ECS 中的功能 Component
- **Mechanism**：对应 ECS 中的原子 Component

### 3. 服务区与 ECS 架构的对应关系

| ECS 架构 | 服务区架构 | 说明 |
|---------|-----------|------|
| Entity | Entity（MME） | 实体标识，唯一 ID |
| Component | Manager/Module/Mechanism | 数据组件，通过 MME 分层组织 |
| System | Service Zone System | 逻辑系统，处理服务区内的业务逻辑 |

**关键差异**：
- **ECS Component**：扁平化的数据组件
- **MME Component**：分层的 Manager → Module → Mechanism 结构，提供更细粒度的数据组织和同步能力

---

## 架构设计

### 1. 服务区层级结构

```
ServiceZoneManager（服务区管理器）
  ├── ServiceZone（服务区实例）
  │     ├── ZoneId（服务区ID）
  │     ├── ZoneType（服务区类型：副本/主城/野外等）
  │     ├── ActorSystem（Actor系统，使用区域分组）
  │     └── EntityMap（Entity集合）
  │           ├── Entity1（PlayerEntity）
  │           │     └── MME Wrapper（数据+脏标记）
  │           ├── Entity2（NPCEntity）
  │           └── ...
  └── ServiceZone（服务区实例）
        └── ...
```

### 2. 服务区 Actor 分组策略

基于现有的 Actor 系统，服务区使用 **区域分组（DISPATCHER_TYPE_IN_REGION）** 策略：

```go
// Actor 元数据配置
type ZoneActorMeta struct {
    ActorName string
    Pattern   string
    Dispatcher Dispatcher {
        Type: DISPATCHER_TYPE_IN_REGION
        ServerId: zoneId  // 服务区ID作为ServerId
        NodeId: nodeId   // 节点ID
    }
}
```

**分组规则**：
- 同一服务区的所有 Entity Actor 使用相同的 `ServerId`（服务区ID）
- 通过 `DISPATCHER_TYPE_IN_REGION` 确保消息路由到正确的服务区
- 支持服务区在不同节点间的迁移和负载均衡

### 3. 服务区生命周期管理

#### 3.1 创建服务区

```go
// 服务区创建流程
func CreateServiceZone(zoneId string, zoneType ZoneType) (*ServiceZone, error) {
    // 1. 创建服务区实例
    zone := NewServiceZone(zoneId, zoneType)
    
    // 2. 初始化 Actor 系统（使用区域分组）
    actorMeta := &actor.Meta{
        ActorName: fmt.Sprintf("zone-%s", zoneId),
        Pattern:   "zone-pattern",
        Dispatcher: actor.Dispatcher{
            Type:     actor.DISPATCHER_TYPE_IN_REGION,
            ServerId: zoneId,
            NodeId:   getCurrentNodeId(),
        },
    }
    
    // 3. 注册到服务区管理器
    ServiceZoneManager.Register(zone)
    
    // 4. 启动服务区
    return zone.Start()
}
```

#### 3.2 销毁服务区

```go
// 服务区销毁流程
func (zone *ServiceZone) Destroy(ctx context.Context) error {
    // 1. 停止所有 Entity Actor
    for _, entity := range zone.EntityMap {
        entity.Stop(ctx)
    }
    
    // 2. 持久化所有 Entity 数据
    for _, entity := range zone.EntityMap {
        entity.Persist()
    }
    
    // 3. 从服务区管理器注销
    ServiceZoneManager.Unregister(zone.ZoneId)
    
    // 4. 清理资源
    return zone.Stop(ctx)
}
```

### 4. Entity 在服务区中的管理

#### 4.1 Entity 注册到服务区

```go
// Entity 注册流程
func (zone *ServiceZone) RegisterEntity(entityId int64, entityType EntityType) error {
    // 1. 创建 Entity Wrapper（MME）
    entityWrapper := NewEntityWrapper(entityId, entityType)
    
    // 2. 创建 Entity Actor
    actorName := fmt.Sprintf("entity-%d", entityId)
    actorPID, err := actor.GetOrStartActor(
        actorName,
        "zone-pattern",
        &actor.Props{
            Meta: &actor.Meta{
                Dispatcher: actor.Dispatcher{
                    Type:     actor.DISPATCHER_TYPE_IN_REGION,
                    ServerId: zone.ZoneId,
                },
            },
        },
    )
    
    // 3. 关联 Entity 和 Actor
    entityWrapper.SetActorPID(actorPID)
    
    // 4. 注册到服务区
    zone.EntityMap[entityId] = entityWrapper
    
    return nil
}
```

#### 4.2 Entity 数据同步

服务区内的 Entity 通过 MME Wrapper 实现增量同步：

```go
// Entity 数据同步流程
func (zone *ServiceZone) SyncEntityToClient(entityId int64, session *network.Session) error {
    entity := zone.EntityMap[entityId]
    if entity == nil {
        return ErrEntityNotFound
    }
    
    // 1. 获取增量同步数据（仅同步变更字段）
    incremental := entity.ToIncrementalProto(mmemodel.SyncContextClient)
    
    // 2. 序列化并发送到客户端
    data, err := proto.Marshal(incremental)
    if err != nil {
        return err
    }
    
    // 3. 发送到客户端
    return session.SendData(data, seq, pid)
}
```

### 5. 服务区消息路由

#### 5.1 服务区内消息路由

服务区内的消息通过 Actor 系统自动路由到对应的 Entity：

```go
// 服务区内消息处理
func (zone *ServiceZone) HandleMessage(msg *Message) error {
    // 1. 解析消息目标 Entity
    entityId := msg.GetTargetEntityId()
    
    // 2. 获取 Entity Actor
    entity := zone.EntityMap[entityId]
    if entity == nil {
        return ErrEntityNotFound
    }
    
    // 3. 通过 Actor 系统发送消息
    return entity.GetActorPID().Request(msg, timeout)
}
```

#### 5.2 跨服务区消息路由

跨服务区的消息通过服务区管理器路由：

```go
// 跨服务区消息路由
func (manager *ServiceZoneManager) RouteMessage(msg *Message) error {
    // 1. 解析目标服务区
    targetZoneId := msg.GetTargetZoneId()
    
    // 2. 获取目标服务区
    targetZone := manager.GetZone(targetZoneId)
    if targetZone == nil {
        return ErrZoneNotFound
    }
    
    // 3. 转发消息到目标服务区
    return targetZone.HandleMessage(msg)
}
```

---

## 数据模型设计

### 1. 服务区数据结构

```go
// ServiceZone 服务区结构
type ServiceZone struct {
    ZoneId      string                    // 服务区ID
    ZoneType    ZoneType                  // 服务区类型
    EntityMap   map[int64]*EntityWrapper // Entity集合
    ActorSystem *actor.ActorSystem        // Actor系统
    State      ZoneState                  // 服务区状态
    CreatedAt  int64                      // 创建时间
    UpdatedAt  int64                      // 更新时间
}

// ZoneType 服务区类型
type ZoneType int32

const (
    ZoneTypeDungeon ZoneType = iota // 副本
    ZoneTypeCity                     // 主城
    ZoneTypeWild                     // 野外
    ZoneTypeArena                    // 竞技场
)
```

### 2. Entity 在服务区中的扩展

```go
// ZoneEntityWrapper 服务区 Entity 包装器
type ZoneEntityWrapper struct {
    *EntityWrapper                    // MME Entity Wrapper
    ZoneId        string              // 所属服务区ID
    ActorPID      *actor.PID         // Actor PID
    LastSyncTime  int64               // 最后同步时间
    DirtyFields   int64               // 脏字段标记（继承自MME）
}
```

### 3. 服务区配置

```yaml
# 服务区配置示例
ServiceZones:
  - ZoneId: "dungeon-001"
    ZoneType: ZoneTypeDungeon
    MaxEntities: 100
    AutoDestroy: true
    DestroyTimeout: 3600  # 1小时后自动销毁
    PersistenceInterval: 60  # 60秒持久化一次
```

---

## 与 ECS 前端架构的对接

### 1. Entity ID 映射

**服务端 Entity ID** ↔ **前端 ECS Entity ID**

```go
// Entity ID 映射规则
// 服务端：int64 类型，全局唯一
// 前端：uint32 类型，在服务区内唯一（可复用）
type EntityIdMapper struct {
    ServerToClient map[int64]uint32  // 服务端ID → 前端ID
    ClientToServer map[uint32]int64  // 前端ID → 服务端ID
}
```

### 2. Component 数据同步

**MME Mechanism** → **ECS Component**

```go
// Mechanism 到 Component 的转换
func (mechanism *MechanismWrapper) ToECSComponent() *ECSComponent {
    return &ECSComponent{
        Type:  mechanism.GetType(),
        Data:  mechanism.ToProto(),
        Dirty: mechanism.GetDirtyBits(),
    }
}
```

**同步策略**：
- **全量同步**：Entity 首次进入服务区时，同步所有 Component
- **增量同步**：Entity 数据变更时，仅同步变更的 Component（通过 MME 脏标记）

### 3. System 逻辑处理

**服务区 System** 对应 **ECS System**

```go
// 服务区 System 示例：移动系统
type MovementSystem struct {
    zone *ServiceZone
}

func (sys *MovementSystem) Update(deltaTime float32) {
    // 1. 遍历服务区内所有 Entity
    for _, entity := range sys.zone.EntityMap {
        // 2. 检查 Entity 是否有移动 Component
        if entity.HasComponent(ComponentTypeMovement) {
            // 3. 处理移动逻辑
            sys.ProcessMovement(entity)
        }
    }
}
```

---

## 服务区管理器（ServiceZoneManager）

### 1. 核心功能

```go
// ServiceZoneManager 服务区管理器
type ServiceZoneManager struct {
    zones      map[string]*ServiceZone  // 服务区集合
    zoneRouter *ZoneRouter              // 服务区路由
    mutex      sync.RWMutex             // 读写锁
}

// 核心方法
func (mgr *ServiceZoneManager) CreateZone(zoneId string, zoneType ZoneType) (*ServiceZone, error)
func (mgr *ServiceZoneManager) GetZone(zoneId string) (*ServiceZone, error)
func (mgr *ServiceZoneManager) DestroyZone(zoneId string) error
func (mgr *ServiceZoneManager) RouteMessage(msg *Message) error
func (mgr *ServiceZoneManager) GetZoneByEntity(entityId int64) (*ServiceZone, error)
```

### 2. 服务区路由策略

```go
// ZoneRouter 服务区路由器
type ZoneRouter struct {
    // 基于 Entity ID 的路由
    entityToZone map[int64]string
    
    // 基于服务区类型的路由
    typeRouter map[ZoneType]func(*Message) string
    
    // 负载均衡路由
    loadBalancer *LoadBalancer
}
```

---

## 数据持久化

### 1. 服务区级持久化

```go
// 服务区数据持久化
func (zone *ServiceZone) Persist(ctx context.Context) error {
    // 1. 持久化所有 Entity 数据
    for _, entity := range zone.EntityMap {
        if entity.IsDirty() {
            // 2. 构建 MongoDB 增量更新
            builder := mgo_builder.NewMongoUpdateBuilder()
            entity.BuildMongoUpdate(builder, mgo_builder.NewNestedPath("entities"))
            
            // 3. 执行持久化
            if err := persistence.Persist("entities", entity.GetId(), builder); err != nil {
                return err
            }
            
            // 4. 清除脏标记
            entity.ClearAllDirty()
        }
    }
    
    return nil
}
```

### 2. 服务区元数据持久化

```go
// 服务区元数据
type ZoneMetadata struct {
    ZoneId      string
    ZoneType    ZoneType
    EntityCount int
    CreatedAt   int64
    UpdatedAt   int64
}
```

---

## 性能优化

### 1. 服务区隔离

- **数据隔离**：不同服务区的 Entity 数据相互隔离，减少锁竞争
- **计算隔离**：每个服务区独立运行，支持并行处理

### 2. 增量同步优化

- **脏标记追踪**：通过 MME Wrapper 的脏标记系统，仅同步变更字段
- **批量同步**：将多个 Entity 的增量数据合并批量发送

### 3. Actor 系统优化

- **区域分组**：使用 `DISPATCHER_TYPE_IN_REGION` 确保消息路由效率
- **Actor 池化**：复用 Actor 实例，减少创建销毁开销

---

## 最佳实践

### 1. 服务区设计原则

- **单一职责**：每个服务区专注于一个游戏场景或功能
- **适度规模**：控制服务区内的 Entity 数量（建议 50-200 个）
- **生命周期管理**：及时销毁空闲服务区，释放资源

### 2. Entity 管理原则

- **及时注册**：Entity 进入服务区时立即注册
- **及时注销**：Entity 离开服务区时立即注销
- **数据同步**：定期同步 Entity 数据到客户端和数据库

### 3. 消息路由原则

- **服务区内优先**：优先在服务区内路由消息
- **跨区最小化**：尽量减少跨服务区的消息路由
- **异步处理**：使用 Actor 系统异步处理消息

---

## 扩展性设计

### 1. 水平扩展

- **服务区迁移**：支持服务区在不同节点间迁移
- **负载均衡**：根据服务区负载动态分配 Entity

### 2. 功能扩展

- **自定义服务区类型**：支持定义新的服务区类型
- **自定义 System**：支持在服务区内添加自定义 System

### 3. 监控与调试

- **服务区监控**：监控服务区的 Entity 数量、消息处理量等
- **性能分析**：分析服务区的性能瓶颈，优化热点路径

---

## 总结

服务区架构通过将游戏世界划分为多个独立的服务区，实现了：

1. **高效的分区管理**：每个服务区独立管理 Entity，减少全局锁竞争
2. **灵活的扩展能力**：支持动态创建和销毁服务区，实现水平扩展
3. **统一的数据模型**：基于 MME 架构，提供统一的数据组织和同步能力
4. **无缝的 ECS 对接**：通过 Entity ID 映射和 Component 同步，实现与前端 ECS 架构的无缝对接

该架构已在生产环境中验证，为大型多人在线游戏提供了高效、可扩展的服务端解决方案。

