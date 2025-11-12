# 集群服务发现

基于 etcd 实现的分布式服务发现和注册系统。

## 功能特性

- ✅ 服务注册与注销
- ✅ 自动心跳保活（租约机制）
- ✅ 服务发现（实时监听节点变化）
- ✅ 节点信息更新
- ✅ 获取在线节点列表
- ✅ 节点缓存（提升性能）

## 快速开始

### 1. 初始化服务发现

```go
import "gitee.com/orbit-w/orbit/core/cluster"

config := cluster.DiscoveryConfig{
    Endpoints:        []string{"127.0.0.1:2379"}, // etcd 端点
    KeyPrefix:        "/orbit/nodes/",            // key 前缀
    LeaseTTL:         10,                          // 租约 TTL (秒)
    HeartbeatInterval: 3 * time.Second,           // 心跳间隔
    DialTimeout:      5 * time.Second,            // 连接超时
}

discovery, err := cluster.NewDiscovery(config)
if err != nil {
    log.Fatalf("Failed to create discovery: %v", err)
}
defer discovery.Close()
```

### 2. 注册节点

```go
node := &cluster.Node{
    ID:      "node-001",
    Address: "192.168.1.100:8080",
    State:   cluster.NodeStateOnline,
}

err := discovery.Register(node)
if err != nil {
    log.Fatalf("Failed to register node: %v", err)
}
```

### 3. 更新节点信息

```go
node.AddZoneCount(5)
node.AddEntityCount(100)
node.AddCCU(50)
err := discovery.UpdateNode(node)
```

### 4. 获取节点列表

```go
// 从 etcd 获取所有在线节点
nodes, err := discovery.GetNodes()
if err != nil {
    log.Printf("Failed to get nodes: %v", err)
}

// 从缓存获取在线节点（更快）
onlineNodes := discovery.GetOnlineNodes()

// 获取指定节点
node, err := discovery.GetNode("node-001")

// 获取节点数量
count := discovery.GetNodeCount()
```

### 5. 注销节点

```go
err := discovery.Unregister()
if err != nil {
    log.Printf("Failed to unregister: %v", err)
}
```

## API 说明

### DiscoveryConfig

服务发现配置结构：

- `Endpoints`: etcd 端点列表（必需）
- `KeyPrefix`: etcd key 前缀，默认 `/orbit/nodes/`
- `LeaseTTL`: 租约 TTL（秒），默认 10
- `HeartbeatInterval`: 心跳间隔，默认 3 秒
- `DialTimeout`: 连接超时，默认 5 秒
- `Username`: etcd 用户名（可选）
- `Password`: etcd 密码（可选）

### Discovery 方法

- `Register(node *Node) error`: 注册节点
- `Unregister() error`: 注销节点
- `UpdateNode(node *Node) error`: 更新节点信息
- `GetNode(nodeID string) (*Node, error)`: 获取指定节点
- `GetNodes() ([]*Node, error)`: 获取所有在线节点（从 etcd）
- `GetOnlineNodes() []*Node`: 获取所有在线节点（从缓存）
- `GetNodeCount() int`: 获取节点数量
- `Close() error`: 关闭服务发现管理器

## 工作原理

1. **服务注册**: 节点注册时创建 etcd 租约，将节点信息写入 etcd
2. **心跳保活**: 定期续约并更新节点信息，保持节点在线状态
3. **服务发现**: 通过 etcd watch 机制实时监听节点变化，更新本地缓存
4. **自动清理**: 节点离线时，租约过期自动从 etcd 删除

## 注意事项

- 确保 etcd 服务正常运行
- 心跳间隔应小于租约 TTL（建议心跳间隔 = TTL / 3）
- 节点信息更新会同步到 etcd，其他节点可通过 watch 实时获取
- 使用 `GetOnlineNodes()` 获取缓存数据，性能更好
- 使用 `GetNodes()` 从 etcd 获取最新数据，但性能较慢

