# Orbit Actor 系统

## 概述

Orbit Actor 系统是基于 Proto.Actor 框架的强大实现，提供了一个简化的 Actor 生命周期管理接口。该系统专为高并发场景设计，内置了弹性恢复机制。

## 核心特性

- **Actor 监督机制**：分层监督结构，实现强大的错误处理
- **基于级别的 Actor 管理**：为不同类型的 Actor 提供不同的优先级
- **自动恢复**：系统自动尝试重启失败的 Actor
- **状态管理**：高效跟踪 Actor 状态（启动中、停止中、已停止）

## 核心组件

### Actor Facade

一个简化的 Actor 系统交互接口：

```go
// ActorFacade 提供了管理 Actor 的简化接口
type ActorFacade struct {
    actorSystem *actor.ActorSystem
    supervisors []*actor.PID
}
```

### Actor 监督

负责管理 Actor 生命周期：

```go
// ActorSupervision 负责管理 Actor 生命周期
type ActorSupervision struct {
    state       atomic.Int32
    level       Level
    actorSystem *actor.ActorSystem
    starting    *Queue
    stopping    *Queue
}
```

## 可靠的 Actor 通信

### GetOrStartActor 方法

`GetOrStartActor` 方法是我们 Actor 系统的关键组件，通过确保获取到随时可用的 Actor 来保证通信可靠性：

```go
// GetOrStartActor 返回一个就绪可用的 Actor
func GetOrStartActor(actorName, pattern string) (*actor.PID, error)
```

#### 可靠性特性

- **保证就绪状态**：该方法确保返回的 Actor 处于就绪状态，可以立即处理消息。
- **自动重启机制**：如果检测到 Actor 正在停止中，系统会自动将请求加入队列，并在 Actor 完全停止后重新启动它。
- **停止状态处理**：当 Actor 正在停止过程中，系统不会立即返回错误，而是将请求加入队列，并在 Actor 完全停止后自动重启它。
- **消息投递保证**：通过确保 Actor 就绪后再返回给调用者，系统防止了由于向已停止或停止中的 Actor 发送消息而导致的消息投递失败。
- **透明恢复**：调用者无需处理 Actor 生命周期管理 - 系统透明地处理 Actor 重启，无需显式干预。
- **故障预防**：帮助避免常见问题，如消息超时或由于 Actor 状态不适当而导致的消息丢失。

这种方法显著减少了以下错误情况：
- 调用者向已停止的 Actor 发送消息
- 由于向正在停止过程中的 Actor 发送消息而导致消息丢失
- 由于 Actor 正在重启过程中而导致消息投递失败
- 由于 Actor 无法处理消息而导致超时
- 调用者遇到静默失败，无法检测到消息投递失败

使用 `GetOrStartActor` 的关键优势在于，客户端无需实现复杂的重试逻辑或处理 Actor 生命周期边缘情况 - 系统透明地处理这些问题，确保消息只发送给准备好接收它们的 Actor。

## 使用示例

```go
// 向 Actor 发送消息（消息投递）
func SendMessage(data string) error {
    return Send("my-actor", "worker-pattern", &MyMessage{Data: data})
}

// 与 Actor 进行请求-响应模式通信
func RequestData(id string) (Response, error) {
    result, err := Request("data-actor", "data-pattern", &DataRequest{ID: id})
    if err != nil {
        return nil, err
    }
    return result.(Response), nil
}
```

## 线程安全

Actor 系统的所有组件都设计为线程安全的，允许多个 goroutine 并发访问而无需额外的同步机制。