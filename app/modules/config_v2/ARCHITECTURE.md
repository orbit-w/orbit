# Config V2 架构设计

## 技术架构

### 核心组件

```
┌─────────────────────────────────────────────────┐
│           Application Layer                     │
│  使用 GetString(), GetInt() 等 API 访问配置      │
└─────────────────────────────────────────────────┘
                       ↓
┌─────────────────────────────────────────────────┐
│          Global API Layer (configs.go)          │
│  提供全局访问函数，简化 API 调用                 │
└─────────────────────────────────────────────────┘
                       ↓
┌─────────────────────────────────────────────────┐
│      ConfigManager (manager.go)                 │
│  ┌─────────────────────────────────────────┐   │
│  │  Nacos SDK                               │   │
│  │  - 连接 Nacos                            │   │
│  │  - 获取配置内容（字符串）                │   │
│  │  - 监听配置变更                          │   │
│  └──────────────┬──────────────────────────┘   │
│                 ↓                                │
│  ┌─────────────────────────────────────────┐   │
│  │  Configuration Injection                 │   │
│  │  将配置字符串注入到 Viper 实例           │   │
│  └──────────────┬──────────────────────────┘   │
│                 ↓                                │
│  ┌─────────────────────────────────────────┐   │
│  │  Viper Instances Map                     │   │
│  │  map[dataId]*viper.Viper                │   │
│  │  - 解析配置（YAML/JSON/TOML）           │   │
│  │  - 提供动态访问 API                      │   │
│  │  - 配置缓存                              │   │
│  └─────────────────────────────────────────┘   │
└─────────────────────────────────────────────────┘
                       ↓
┌─────────────────────────────────────────────────┐
│              Nacos Server                       │
│  配置存储和推送                                  │
└─────────────────────────────────────────────────┘
```

## 工作流程

### 初始化流程

```
1. 启动 ConfigManager
   ↓
2. 读取 config_center.toml
   - Nacos 连接配置
   - 配置源列表 (sources)
   ↓
3. 初始化 Nacos SDK Client
   - 连接 Nacos 服务器
   - 验证连接
   ↓
4. 遍历 sources 列表
   ↓
   ┌─ 对每个 source:
   │  4.1 通过 Nacos SDK 获取配置内容（字符串）
   │  4.2 创建 Viper 实例
   │  4.3 将配置内容注入 Viper
   │  4.4 存储到 map[dataId]*viper.Viper
   │  4.5 注册 Nacos 配置监听器
   └─
   ↓
5. 初始化完成，可以访问配置
```

### 配置访问流程

```
应用调用 GetString("game.main", "server", "server.name")
   ↓
configs.go: GetString(dataId, group, key)
   ↓
manager.GetString(dataId, group, key)
   ↓
manager.GetViper(dataId, group)  // 从 map 获取 viper 实例（使用 dataId.group 作为键）
   ↓
viper.GetString(key)             // Viper 动态访问
   ↓
返回配置值
```

### 配置更新流程

```
1. Nacos 服务器配置变更
   ↓
2. Nacos SDK 监听器收到推送通知
   OnChange(namespace, group, dataId, newContent)
   ↓
3. ConfigManager 处理更新
   3.1 创建新的 Viper 实例
   3.2 将新配置内容注入 Viper
   3.3 替换 map 中的旧实例
   ↓
4. 执行用户注册的回调函数
   callbacks[dataId]()
   ↓
5. 应用自动使用新配置
   下次调用 GetString() 自动返回新值
```

## 关键设计决策

### 1. 为什么使用 Nacos SDK 而不是 nacos-viper-remote？

**决策**：直接使用 `nacos-sdk-go/v2` + 手动注入 Viper

**原因**：
- ✅ **官方支持**：nacos-sdk-go 是 Nacos 官方维护的 SDK
- ✅ **稳定可靠**：经过大量生产环境验证
- ✅ **一致性**：与原 config 模块使用相同的 SDK
- ✅ **可控性**：完全控制配置加载和更新流程
- ✅ **无中间层**：减少第三方依赖的风险

**nacos-viper-remote 的问题**：
- ❌ 第三方库，维护不活跃
- ❌ 可能存在版本兼容性问题
- ❌ 增加额外的依赖复杂度
- ❌ 更新不及时

### 2. 为什么使用 Map 管理多个 Viper 实例？

**决策**：`map[dataId]*viper.Viper`

**原因**：
- ✅ **隔离性**：每个配置源独立的 Viper 实例，互不干扰
- ✅ **灵活性**：支持不同格式（YAML、JSON、TOML）
- ✅ **简洁性**：通过 dataId 直接访问对应配置
- ✅ **扩展性**：轻松添加新配置源

**替代方案对比**：
- ❌ 单一 Viper 实例：不同配置源可能冲突
- ❌ 合并配置：失去配置源边界，难以管理

### 3. 为什么不缓存配置值？

**决策**：每次访问都通过 Viper 获取

**原因**：
- ✅ **实时性**：配置更新后立即生效
- ✅ **简洁性**：无需管理缓存失效逻辑
- ✅ **可靠性**：Viper 内部已有缓存机制
- ✅ **低开销**：Viper 的 Get 操作很快（map 查找）

### 4. 配置更新策略

**决策**：整体替换 Viper 实例

**实现**：
```go
// 收到配置更新
v := viper.New()
v.ReadConfig(bytes.NewBufferString(newContent))

m.mu.Lock()
m.vipers[dataId] = v  // 原子替换
m.mu.Unlock()
```

**原因**：
- ✅ **原子性**：一次性替换，无中间状态
- ✅ **一致性**：确保配置的完整性
- ✅ **简洁性**：无需处理部分更新

**替代方案**：
- ❌ 增量更新：复杂，容易出错
- ❌ 锁定整个更新过程：影响并发访问

## 并发安全

### 读写锁机制

```go
type ConfigManager struct {
    vipers map[string]*viper.Viper
    mu     sync.RWMutex
}

// 读操作（高频）
func (m *ConfigManager) GetViper(dataId, group string) *viper.Viper {
    key := m.genConfigNameSpaceId(dataId, group)
    m.mu.RLock()         // 读锁
    defer m.mu.RUnlock()
    return m.vipers[dataId]
}

// 写操作（低频）
func (m *ConfigManager) updateViper(dataId string, v *viper.Viper) {
    m.mu.Lock()          // 写锁
    defer m.mu.Unlock()
    m.vipers[dataId] = v
}
```

**设计理由**：
- 读操作（GetViper）频率极高，使用 RLock 允许并发读
- 写操作（配置更新）频率很低，使用 Lock 独占
- Viper 实例本身是并发安全的

## 性能考虑

### 1. 配置访问性能

- **Viper 查找**：O(1) map 查找
- **无网络调用**：所有配置缓存在内存
- **并发读取**：读锁，支持高并发

### 2. 配置更新性能

- **推送机制**：Nacos SDK 使用长连接，实时推送
- **异步更新**：不阻塞配置访问
- **原子替换**：写锁时间极短

### 3. 内存占用

- **每个 DataId 一个 Viper 实例**：通常只有几个到十几个
- **配置内容**：已经在内存中，Viper 不增加额外占用
- **连接开销**：单一 Nacos SDK 客户端

## 错误处理

### 初始化错误

```go
if err := InitConfig("config_center.toml"); err != nil {
    // 1. 文件不存在
    // 2. 配置格式错误
    // 3. Nacos 连接失败
    // 4. 配置内容为空
    log.Fatal(err)
}
```

### 运行时错误

```go
// 配置不存在
v := GetViper("non-exist", "DEFAULT_GROUP")  // 返回 nil
value := GetString("non-exist", "DEFAULT_GROUP", "key")  // 返回空字符串

// 配置更新失败
// - 记录日志
// - 保持旧配置
// - 不影响服务运行
```

## 与原 config 模块对比

| 特性 | 原 config | Config V2 |
|------|-----------|-----------|
| Nacos SDK | nacos-sdk-go/v2 | nacos-sdk-go/v2 |
| 配置定义 | 必须定义结构体 | 无需定义 |
| 配置访问 | 通过结构体 | 通过动态 API |
| 灵活性 | 低 | 极高 |
| 代码量 | 高 | 低 |
| 学习成本 | 中 | 低 |
| Nacos 集成 | 直接调用 | 直接调用 |
| 配置更新 | 实时推送 | 实时推送 |

**共同点**：
- 都使用官方 nacos-sdk-go/v2
- 都支持配置热更新
- 都使用实时推送机制

**差异点**：
- Config V2 引入 Viper 提供动态访问
- Config V2 无需定义配置结构体
- Config V2 更灵活，代码量更少

## 最佳实践

### 1. 配置源组织

```toml
# 按功能模块划分
[[sources]]
data_id = "game.server"
group = "server"
format = "yaml"

[[sources]]
data_id = "game.redis"
group = "redis"
format = "yaml"
```

### 2. 错误处理

```go
// 初始化时检查
if err := config_v2.InitConfig("config.toml"); err != nil {
    log.Fatalf("Config init failed: %v", err)
}

// 使用时提供默认值
port := config_v2.GetInt("game.server", "DEFAULT_GROUP", "port")
if port == 0 {
    port = 8080  // 默认值
}
```

### 3. 配置变更处理

```go
config_v2.OnConfigChange("game.server", "DEFAULT_GROUP", func() {
    // 验证新配置
    newPort := config_v2.GetInt("game.server", "DEFAULT_GROUP", "port")
    if newPort < 1024 || newPort > 65535 {
        log.Error("Invalid port in new config")
        return
    }
    
    // 应用新配置
    applyNewConfig(newPort)
})
```

## 总结

Config V2 采用 **Nacos SDK + Viper** 的组合架构，既保持了 Nacos SDK 的稳定性和官方支持，又充分利用了 Viper 的灵活性和动态访问能力。

这种设计实现了：
- ✅ 零样板代码
- ✅ 动态配置访问
- ✅ 配置热更新
- ✅ 高性能
- ✅ 并发安全
- ✅ 易于使用和维护

---

**文档版本**: V2.0  
**最后更新**: 2025-12-07
