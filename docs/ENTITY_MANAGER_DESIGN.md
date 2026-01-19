# Entity 生命周期管理与异步加载设计方案

## 1. 背景与目标

本方案是对 [Orbit 并发调度与锁机制设计方案](./并发调度与锁机制设计方案.md) 的补充实现。

在 Orbit 的高并发调度模型中，Worker 需要在执行业务逻辑前，按照 "Canonical Ordering"（规范序）对涉及的所有 Entity ID 进行加锁。这就引出了一个核心矛盾：**Worker 必须在访问数据之前先获取锁，但此时 Entity 的数据可能根本还没有加载到内存中。**

因此，`EntityManager` 的设计目标如下：
1.  **锁与数据解耦**：支持对尚未加载（甚至尚未创建）的 Entity ID 获取锁对象。
2.  **异步非阻塞**：加载过程必须是异步的，严禁阻塞 Worker 线程。
3.  **时序保障**：在 Entity 加载期间到达的消息，必须在 Worker 内部暂存并保持严格时序，不能直接丢弃或乱序重试。
4.  **无空转 (No Busy Waiting)**：避免使用轮询（Polling）方式等待 IO，采用事件驱动（Event Driven）机制。

---

## 2. 核心架构：EntityRecord

为了实现"锁"与"数据"分离，我们引入 `EntityContainer` 作为 Entity 在内存中的容器。它独立于业务数据存在，承载了**并发控制**和**生命周期状态**。

### 2.1 数据结构

```go
package manager

import (
    "sync"
    "gitee.com/orbit-w/meteor/bases/sync/reentrantlock"
    "gitee.com/orbit-w/orbit/internal/game/mme_agent"
)

// EntityStatus 定义实体的加载状态
type EntityStatus int32

const (
    StatusNone      EntityStatus = 0 // 未加载
    StatusLoading   EntityStatus = 1 // 加载中
    StatusLoaded    EntityStatus = 2 // 已加载
    StatusFailed    EntityStatus = 3 // 加载失败
    StatusDestroyed EntityStatus = 4 // 已销毁
)

// EntityContainer 是 Entity 在内存中的容器
type EntityContainer struct {
    EntityID int64

    // Lock 是业务逻辑锁 (ReentrantLock)
    // 关键点：允许在 Entity 数据未加载时就获取此锁，用于防重入和并发控制。
    // 该锁对象的生命周期长于 Entity 数据本身。
    Lock *reentrantlock.ReentrantLock

    // mu 保护 Record 自身的内部状态 (status, entity, waiters)
    // 注意：这是一个极其短临界的互斥锁，绝不包含 IO 操作。
    mu sync.Mutex

    // Status 标记当前加载状态
    Status EntityStatus

    // Entity 实际的业务数据对象 (MME Agent)
    // 仅在 Status == StatusLoaded 时可用
    Entity mme_agent.IEntity

    // waiters 等待加载完成的回调列表
    waiters []func(mme_agent.IEntity, error)
}
```

### 2.2 EntityManager 接口设计

`EntityManager` 对外提供三个正交的 API，分别对应锁获取、数据获取和加载触发。

```go
type EntityManager struct {
    // 存储所有 Record。Key: EntityID, Value: *EntityRecord
    records sync.Map
    loader  IEntityLoader
}

// GetLock 获取指定 Entity 的业务锁
// 即使 Entity 未加载，也会返回一个有效的锁对象。这是实现 "排序 TryLock" 的基础。
func (m *EntityManager) GetLock(entityID int64) *reentrantlock.ReentrantLock

// GetEntity 尝试获取已加载的 Entity
// 这是一个非阻塞调用。如果数据未就绪，直接返回 nil。
func (m *EntityManager) GetEntity(entityID int64) mme_agent.IEntity

// AsyncLoad 触发异步加载
// 如果 Entity 已经在加载中，则只注册回调；否则启动新的加载任务。
// cb 会在加载完成后被调用（通常在单独的 IO goroutine 中）。
func (m *EntityManager) AsyncLoad(entityID int64, cb func(mme_agent.IEntity, error))
```

---

## 3. 调度模型集成：挂起-唤醒机制 (Suspend & Wake-up)

为了解决"数据未就绪"场景下的调度问题，我们放弃了传统的"重试队列轮询"方案，采用了更高效的 **"挂起-唤醒"** 机制。

### 3.1 为什么不使用 RetryList 轮询？

如果将等待加载的消息放入 `RetryList`，Worker 会周期性地唤醒并执行：`TryLock` -> `GetEntity(nil)` -> `AsyncLoad` -> `Unlock`。
*   **资源浪费**：Worker 在 IO 期间（可能几十毫秒）会进行成百上千次无意义的 CPU 空转。
*   **竞争干扰**：反复的加锁解锁可能干扰其他正常业务对该锁的获取。

### 3.2 优化后的 Worker 结构

在 Worker 内部增加状态追踪，用于识别哪些 Anchor 处于"加载挂起"状态。

```go
type Worker struct {
    // ... 原有字段 (InputChan, StagingArea, RetryList) ...

    // LoadingAnchors 记录当前正在加载的 Anchor ID
    // 作用：作为断路器，防止对正在加载的 Entity 进行无效的调度尝试
    LoadingAnchors map[int64]bool 
}

// 内部事件：用于通知 Worker 加载完成
type EntityLoadedEvent struct {
    AnchorID int64
    Entity   mme_agent.IEntity
    Err      error
}
```

### 3.3 详细执行流程

#### 阶段一：消息分发 (Fetch & Dispatch)
*(保持原方案不变)*
1.  Worker 收到消息 `M`，提取涉及的 ID 列表，排序确定 `AnchorID`。
2.  检查 `StagingArea[AnchorID]`：
    *   **非空**：说明该 Entity 正忙（被锁住或正在加载）。为保时序，将 `M` 追加到队尾。
    *   **为空**：进入阶段二。

#### 阶段二：尝试执行 (Try & Process)
1.  **加载断路检查 (Fast Fail)**：
    *   检查 `Worker.LoadingAnchors[AnchorID]`。
    *   若为 `true`，说明正在加载中。直接将 `M` 放入 `StagingArea`（如果还没放），**不进行任何处理，也不加入 RetryList**。当前处理结束。

2.  **获取锁 (TryLock)**：
    *   按序尝试获取所有涉及 ID 的锁。
    *   **失败**：进入**锁竞争分支** -> 放入 `StagingArea`，加入 `RetryList`（走原有退避重试逻辑）。
    *   **成功**：持有所有锁，进入下一步。

3.  **数据状态检查 (GetEntity)**：
    *   调用 `EntityManager.GetEntity(AnchorID)`。
    
    *   **情况 A：数据就绪 (Entity != nil)**
        *   执行业务逻辑。
        *   解锁。
        *   检查 `StagingArea` 是否还有堆积消息，循环处理。

    *   **情况 B：数据未就绪 (Entity == nil)** -> **触发挂起流程**
        1.  **发起加载**：调用 `EntityManager.AsyncLoad(AnchorID, cb)`。
            回调逻辑：`cb` 执行时，向 Worker 的 `InputChan` 发送 `EntityLoadedEvent`。
        2.  **保存消息**：确保当前消息 `M` 位于 `StagingArea[AnchorID]` 的 **头部 (Head)**。
        3.  **标记挂起**：设置 `Worker.LoadingAnchors[AnchorID] = true`。
        4.  **释放锁**：立即调用 `Unlock` 释放持有的所有锁（避免阻塞其他已有数据的业务）。
        5.  **静默**：不将此 ID 加入 `RetryList`。

#### 阶段三：唤醒 (Wake-up)
1.  IO 线程完成加载，触发回调，`EntityLoadedEvent` 进入 Worker 的 `InputChan`。
2.  Worker 处理该事件：
    *   **清除标记**：`delete(Worker.LoadingAnchors, event.AnchorID)`。
    *   **触发处理**：如果加载成功，立即针对该 `AnchorID` 发起一次处理流程（提取 `StagingArea` 队头消息 -> TryLock -> GetEntity ...）。

---

## 4. 方案总结与状态流转

本方案通过三种状态的流转，完美覆盖了并发调度中的不同场景：

| 场景 | 状态特征 | 消息位置 | 驱动机制 | 行为策略 |
| :--- | :--- | :--- | :--- | :--- |
| **空闲** | 无锁，已加载 | 此时无消息 | 消息到达 | 直接执行 |
| **锁竞争** | 锁被占用 | `StagingArea` | `RetryList` | **轮询 + 退避 (Spin/Backoff)** |
| **加载中** | 数据未就绪 | `StagingArea` | `LoadingAnchors` | **挂起 + 事件唤醒 (Suspend/Wakeup)** |

### 优势
1.  **完全死锁免疫**：严格遵循 "Canonical Ordering" 先锁 ID 后操作的原则。
2.  **IO 零阻塞**：Worker 线程从不进行阻塞式 IO，也不在 IO 期间空转等待。
3.  **严格时序**：`StagingArea` 作为统一的缓冲池，无论是等锁还是等数据，都保证了同一 Entity 的消息 FIFO 执行。
4.  **高效能**：区分了"短等待"（锁竞争）和"长等待"（IO），分别采用最适合的调度策略。
