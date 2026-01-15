# Orbit 并发调度与锁机制设计方案

## 1. 核心挑战

在 Orbit 消息调度模型中，虽然通过 Dispatcher 实现了基于 Entity 的分片路由，但在处理跨 Entity 事务（如交易、组队、战斗结算）时，不可避免地会遇到以下矛盾：

1.  **资源竞争**：多个 Worker 可能同时尝试锁定同一个 Entity。
2.  **死锁风险**：若 Worker A 锁了 Entity 1 等 Entity 2，而 Worker B 锁了 Entity 2 等 Entity 1，将导致死锁。
3.  **时序破坏**：传统的 "锁失败即 Requeue（放回队尾）" 策略会导致同一 Entity 的消息乱序。
4.  **阻塞问题**：Actor 模型要求 Worker 必须是高吞吐的，严禁使用操作系统级阻塞锁（如 `Mutex.Lock()`）。

本方案通过 **"排序 TryLock + 本地分级队列"** 的组合策略解决上述所有问题。

---

## 2. 锁机制选型

### 2.1 锁的选择
使用项目中已有的 `ReentrantLock` (`meteor/bases/sync/reentrantlock`)。
*   **特性**：支持同一 Goroutine 重入，支持自旋（Spin），支持非阻塞尝试。
*   **适用性**：适应业务逻辑中可能存在的递归调用（Method A -> Method B -> Method A）。

### 2.2 使用原则
*   **强制非阻塞**：严禁使用 `Lock()` 方法。必须使用 `TryLock()` 或 `TryLockWithSpin()`。
*   **快速失败**：如果无法获取所有必要的锁，立即放弃当前操作，释放已持有的所有锁。

---

## 3. 死锁规避策略：规范序（Canonical Ordering）

在任何涉及多个 Entity 的原子操作中，必须严格遵守**"先排序，后加锁"**的规则。

### 规则定义
1.  提取消息涉及的所有 `EntityRef` ID。
2.  对 ID 列表进行**字典序（Lexicographical）**或**数值序**排序。
3.  Worker 必须严格按照排序后的顺序依次尝试加锁。

### 效果
如果 Worker 1 需要锁 [A, B]，Worker 2 需要锁 [B, A]：
*   双方都会先尝试锁 A。
*   谁先拿到 A，谁才有资格去尝试锁 B。
*   **结果**：永远不会出现环路等待，从数学上根除了死锁可能。

---

## 4. 时序保障策略：Worker 内部暂存区 (Staging Area)

为了解决锁竞争失败后的重试问题，同时不破坏消息时序，Worker 内部必须实现**基于 Entity 的二级调度**，而不是简单地将消息退回全局队列。

### 4.1 数据结构

在 Worker 内部维护一个暂存区：

```go
type Worker struct {
    // 1. 全局输入通道 (来自 Dispatcher)
    InputChan <-chan Message 

    // 2. 本地暂存区 (Staging Area)
    // Key: AnchorKey (主导 Entity ID), Value: 待处理消息队列 (FIFO)
    // 作用：当主导 Entity 处于"忙碌/竞争"状态时，后续消息在此排队，确保时序。
    StagingArea map[string]*Queue 

    // 3. 重试列表
    // 记录哪些 AnchorKey 当前有消息积压，需要重试
    RetryList *UniqueQueue 
}
```

### 4.2 执行流程 (两阶段调度)

#### 阶段一：消息分发 (Fetch & Dispatch)
当 Worker 从 `InputChan` 收到消息 `M` (涉及 Entity [A, B], Anchor 为 A)：

1.  **检查 StagingArea[A]**：
    *   **非空**：说明 A 正处于繁忙或等待重试状态。为保证 M 晚于之前的消息执行，**必须**将 M 追加到 `StagingArea[A]` 的**队尾**。当前处理结束。
    *   **为空**：说明 A 当前空闲。直接进入阶段二。

#### 阶段二：尝试执行 (Try & Lock)
尝试执行消息 `M`：

1.  **排序**：对涉及的 ID [A, B] 进行排序 -> [A, B]。
2.  **顺序 TryLock**：
    *   `TryLock(A)` -> 成功。
    *   `TryLock(B)` -> 失败 (被其他 Worker 占用)。
3.  **处理结果**：
    *   **成功**：执行业务逻辑 -> 释放所有锁 -> 返回。
    *   **失败 (锁竞争)**：
        1.  **回滚**：立即释放已持有的锁 (A)。
        2.  **入列**：将 `M` 放入 `StagingArea[A]` 的**队头 (Head)** (如果是刚从 InputChan 来的消息则新建队列)。
        3.  **标记**：将 A 加入 `RetryList`。
        4.  **避让**：设置短暂的 Backoff (可选)。

#### 阶段三：重试循环 (Retry Loop)
Worker 在处理主循环空闲时，或每隔固定 Tick：

1.  遍历 `RetryList` 中的 AnchorKey (如 A)。
2.  取出 `StagingArea[A]` 的**队头**消息。
3.  再次尝试 **阶段二**。
4.  **如果成功**：
    *   从队列移除该消息。
    *   **继续处理**该队列的下一条消息（Batch Process），直到队列为空或再次遇到锁竞争。

---

## 5. 方案总结

| 问题 | 解决方案 | 核心机制 |
| :--- | :--- | :--- |
| **死锁** | 规范序加锁 | `Sort(IDs) -> TryLock in order` |
| **Worker 阻塞** | 乐观锁 + 自旋 | `TryLockWithSpin()` |
| **消息乱序** | 本地暂存区 | `StagingArea` (Head-of-Line Blocking per Entity) |
| **饥饿** | 重试机制 | `RetryList` + Backoff |

该方案将 Orbit 的消息调度分为两层：
1.  **全局层 (Dispatcher)**：负责负载均衡，将消息路由到特定 Worker。
2.  **本地层 (Worker)**：负责时序保障和竞争处理，通过内部队列消化瞬时的锁冲突。
