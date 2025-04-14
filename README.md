# Orbit

个人练习用游戏服务框架，支持长连接流式传输和高并发处理。

## 项目概述

Orbit是一个基于Go语言开发的分布式服务框架，目前用于个人练习

## 特性

- **高性能**: 优化的网络层设计，支持高并发连接
- **可扩展**: 灵活的插件系统，易于扩展功能
- **流式处理**: 支持长连接和流式数据传输

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

## 贡献指南

欢迎贡献代码和提出问题！请查看[贡献指南](CONTRIBUTING.md)了解更多信息。

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。
