# Orbit

状态同步游戏服务框架。

## 项目概述

Orbit是一个基于Go语言开发的状态同步游戏服务框架，目前用于个人练习，并发模型使用Actor模型。

## 特性

- **高性能**: 优化的网络层设计，支持高并发连接
- **流式处理**: 支持长连接和流式数据传输
- **Actor模型**: 基于Actor模型实现的并发控制，提供良好的隔离性与可伸缩性
- **消息驱动**: 通过消息传递实现组件间通信，降低耦合度
- **状态同步**: 支持状态同步架构

## 核心组件

### Actor系统

基于Actor模型实现的轻量级并发处理框架，主要特点：

- 每个Actor维护自己的状态，通过消息传递与其他Actor通信
- 非共享内存模型，避免了锁竞争问题
- 支持监督树结构，实现故障隔离与恢复

### Actor定时器管理器 (TimerMgr)

提供高效的定时任务处理机制：

- **一次性定时器**: 到期后自动移除，适用于临时任务
- **系统定时器**: 支持自动续约，适用于周期性任务
- **高效调度**: 基于最小堆实现的优先级队列，保证O(log n)的操作复杂度

使用示例：

```go
// 创建定时器管理器
timerMgr := actor.NewTimerMgr(func() {
    // 定时器触发时的回调函数
})

// 添加一次性定时器
timer := timerMgr.AddTimerOnce("timer-id", 1*time.Second, "timer-message")

// 添加自动续约的系统定时器
systemTimer := timerMgr.AddSystemTimer("system-timer-id", 5*time.Second, "system-timer-message")

// 移除定时器
timerMgr.RemoveTimer("timer-id")

// 处理到期的定时器
timerMgr.Process(func(msg any) {
    // 处理定时器消息
    fmt.Println("定时器消息:", msg)
})
```

## 通信协议

### 网络连接

- 支持TCP/KCP协议（基于UDP的可靠传输协议）
- 支持长连接流式传输
- 服务器默认监听端口：8900

### 消息编码规则

#### 网络层编码

##### 上行消息（客户端 -> 服务器）

网络层的消息格式如下：

```
size (int32) | gzipped (bool) | type (byte) | [length (uint32) | body (bytes) ]...
```

- `size`: 消息总长度（4字节整数）
- `gzipped`: 消息体是否经过gzip压缩（1字节布尔值）
- `type`: 消息类型（1字节整数）
- `length`: 消息内容的长度（4字节整数）
- `body`: 实际的消息数据（变长字节数组）

##### 下行消息（服务器 -> 客户端）

网络层的消息格式如下：

```
size (int32) | gzipped (bool) | type (byte) | body (bytes)
```

- `size`: 消息总长度（4字节整数）
- `gzipped`: 消息体是否经过gzip压缩（1字节布尔值）
- `type`: 消息类型（1字节整数）
- `body`: 消息体（变长字节数组）

#### 业务层编码

业务层的消息体（body）格式如下：

```
[协议号（4byte）| seq（4byte，optional）| 消息长度（4byte）| 消息内容（bytes）]...
```

- `协议号`: 标识消息类型的ID（4字节整数）
- `seq`: 消息序列号，可选字段（4字节整数）
- `消息长度`: 消息内容的长度（4字节整数）
- `消息内容`: 实际的消息数据（变长字节数组）

## 安装

```bash
go get gitee.com/orbit-w/orbit
```

## 快速开始

### 创建服务器

```go
package main

import (
    "gitee.com/orbit-w/orbit/server"
)

func main() {
    // 创建服务器实例
    srv := server.NewServer()
    
    // 启动服务器
    if err := srv.Start(); err != nil {
        panic(err)
    }
    
    // 等待服务器关闭
    srv.Wait()
}
```

## 配置说明

配置文件位于 `configs/config.toml`，支持以下配置选项：

- 服务器配置
- 网络配置
- 日志配置
- 性能调优选项

## 性能优化

Orbit框架针对游戏服务器的特殊需求进行了多项性能优化：

- 对象池复用，减少GC压力
- 无锁化设计，减少上下文切换
- 批处理消息，提高吞吐量
- 高效的内存管理与缓存策略

## 贡献指南

欢迎贡献代码和提出问题！请查看[贡献指南](CONTRIBUTING.md)了解更多信息。

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。
