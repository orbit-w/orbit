# Config V2 模块

基于 Viper + Nacos 的灵活配置管理模块，**无需为每个配置项定义结构体**。

## 🎯 设计理念

### 为什么重新设计？

原 config 模块的问题：
1. ❌ **不够灵活**：每个 Nacos DataId/Group 都需要定义结构体和实现接口
2. ❌ **代码冗余**：添加新配置项需要写大量重复代码
3. ❌ **维护成本高**：配置结构变更需要修改多个文件

### Config V2 的优势

1. ✅ **极致灵活**：使用 Viper 直接访问配置，无需定义结构体
2. ✅ **零样板代码**：添加新配置只需在 toml 中配置，不需要写代码
3. ✅ **Viper 原生支持**：利用 Viper 的远程配置能力集成 Nacos
4. ✅ **动态访问**：通过 key 路径动态访问任何配置值
5. ✅ **多数据源**：轻松管理多个 DataId/Group 配置

## 核心特性

- ✅ 使用 Viper 的 Remote Provider 支持 Nacos
- ✅ 无需为每个配置项定义结构体
- ✅ 支持动态配置访问（GetString, GetInt, GetBool 等）
- ✅ 支持配置热更新和变更回调
- ✅ 支持多配置源（多个 DataId/Group）
- ✅ 支持配置反序列化（可选，按需使用）

## 快速开始

### 1. 依赖说明

Config V2 使用以下依赖：
- `github.com/spf13/viper` - 配置管理
- `github.com/nacos-group/nacos-sdk-go/v2` - Nacos SDK（与原 config 模块相同）

**注意**：无需手动安装，`go mod` 会自动管理依赖。

### 2. 配置文件

创建 `config_center.yaml`:

```yaml
type: nacos

nacos:
  server_hosts:
    - your-nacos-server.com:8848  # 支持多个服务器
  namespace_id: your-namespace-id
  username: ""
  password: ""
  timeout_ms: 5000

# 配置源列表 - 可以添加任意多个
sources:
  - data_id: game.main
    group: server
    format: yaml

  - data_id: game.main
    group: redis
    format: yaml

  - data_id: custom.config
    group: DEFAULT_GROUP
    format: json
```

### 3. Nacos 中的配置示例

在 Nacos 中创建配置（Group: `server`, DataId: `game.main`）:

```yaml
name: "game-server"
stage: "dev"
host: "0.0.0.0"
port: 8080
features:
  enable_debug: true
  max_players: 1000
database:
  host: "localhost"
  port: 3306
  name: "gamedb"
```

### 4. 代码使用

```go
package main

import (
    "fmt"
    config_v2 "gitee.com/orbit-w/orbit/app/modules/config_v2"
)

func main() {
    // 1. 初始化配置
    if err := config_v2.InitConfig("config_center.yaml"); err != nil {
        panic(err)
    }
    defer config_v2.StopConfig()

    // 2. 直接访问配置，无需定义结构体！
    // 参数：dataId, group, key
    serverName := config_v2.GetString("game.main", "server", "name")
    serverPort := config_v2.GetInt("game.main", "server", "port")
    enableDebug := config_v2.GetBool("game.main", "server", "features.enable_debug")
    
    fmt.Printf("Server: %s:%d, Debug: %v\n", serverName, serverPort, enableDebug)

    // 3. 访问嵌套配置
    dbHost := config_v2.GetString("game.main", "server", "database.host")
    dbPort := config_v2.GetInt("game.main", "server", "database.port")
    
    fmt.Printf("Database: %s:%d\n", dbHost, dbPort)

    // 4. 获取数组配置
    redisAddrs := config_v2.GetStringSlice("game.main", "redis", "redis.addrs")
    fmt.Printf("Redis addrs: %v\n", redisAddrs)

    // 5. 获取 Map 配置
    features := config_v2.GetStringMap("game.main", "server", "features")
    fmt.Printf("Features: %+v\n", features)
}
```

## API 文档

### 初始化和管理

```go
// 初始化配置管理器
func InitConfig(filename string) error

// 停止配置管理器
func StopConfig() error
```

### 动态访问配置

```go
// 获取任意类型配置
func Get(dataId, group, key string) any

// 获取字符串配置
func GetString(dataId, group, key string) string

// 获取整数配置
func GetInt(dataId, group, key string) int

// 获取布尔配置
func GetBool(dataId, group, key string) bool

// 获取字符串数组
func GetStringSlice(dataId, group, key string) []string

// 获取 Map 配置
func GetStringMap(dataId, group, key string) map[string]any
```

### 配置反序列化（可选）

如果你仍然想使用结构体，也支持：

```go
// 反序列化整个配置到结构体
func Unmarshal(dataId, group string, rawVal any) error

// 反序列化配置的某个 key 到结构体
func UnmarshalKey(dataId, group, key string, rawVal any) error
```

示例：

```go
type ServerConfig struct {
    Name  string `yaml:"name"`
    Stage string `yaml:"stage"`
    Host  string `yaml:"host"`
    Port  int    `yaml:"port"`
}

var config ServerConfig
err := config_v2.Unmarshal("game.main", "server", &config)
if err != nil {
    panic(err)
}

fmt.Printf("Server: %+v\n", config)
```

### 配置变更回调

```go
// 注册配置变更回调
func OnConfigChange(dataId, group string, callback func())
```

示例：

```go
config_v2.OnConfigChange("game.main", "server", func() {
    fmt.Println("配置已更新!")
    newName := config_v2.GetString("game.main", "server", "name")
    fmt.Printf("新的服务器名称: %s\n", newName)
})
```

### 获取 Viper 实例（高级）

```go
// 获取指定 DataId 和 Group 的 viper 实例
func GetViper(dataId, group string) *viper.Viper
```

这允许你使用 Viper 的所有高级功能。

**注意**：在 Nacos 中，`dataId` 和 `group` 的组合才能唯一标识一个配置，因此所有 API 都需要同时提供这两个参数。

## 使用示例

### 示例 1：访问简单配置

```go
// Nacos 配置 (Group: "server", DataId: "game.main"):
// name: "my-server"
// port: 8080

serverName := config_v2.GetString("game.main", "server", "name")
serverPort := config_v2.GetInt("game.main", "server", "port")
```

### 示例 2：访问嵌套配置

```go
// Nacos 配置 (Group: "server", DataId: "game.main"):
// database:
//   host: "localhost"
//   port: 3306
//   credentials:
//     username: "admin"
//     password: "secret"

dbHost := config_v2.GetString("game.main", "server", "database.host")
dbPort := config_v2.GetInt("game.main", "server", "database.port")
username := config_v2.GetString("game.main", "server", "database.credentials.username")
```

### 示例 3：访问数组配置

```go
// Nacos 配置 (Group: "redis", DataId: "game.main"):
// redis:
//   addrs:
//     - "127.0.0.1:6379"
//     - "127.0.0.1:6380"

redisAddrs := config_v2.GetStringSlice("game.main", "redis", "redis.addrs")
// redisAddrs = ["127.0.0.1:6379", "127.0.0.1:6380"]
```

### 示例 4：使用结构体（可选）

```go
type RedisConfig struct {
    Addrs    []string `yaml:"addrs"`
    Password string   `yaml:"password"`
    DB       int      `yaml:"db"`
}

var redisConfig RedisConfig
err := config_v2.UnmarshalKey("game.main", "redis", "redis", &redisConfig)
if err != nil {
    panic(err)
}

fmt.Printf("Redis: %+v\n", redisConfig)
```

### 示例 5：多配置源

```yaml
# config_center.yaml
sources:
  - data_id: game.main
    group: server
    format: yaml

  - data_id: game.main
    group: redis
    format: yaml

  - data_id: feature.flags
    group: feature
    format: json
```

```go
// 访问不同的配置源
serverName := config_v2.GetString("game.main", "server", "name")  // 从 server group
redisAddr := config_v2.GetString("game.main", "redis", "addr")   // 从 redis group
featureX := config_v2.GetBool("feature.flags", "feature", "feature_x_enabled")
```

### 示例 6：配置热更新

```go
// 注册回调，配置变更时自动执行
config_v2.OnConfigChange("game.main", "server", func() {
    // 配置已更新，重新读取
    newMaxPlayers := config_v2.GetInt("game.main", "server", "max_players")
    fmt.Printf("最大玩家数已更新为: %d\n", newMaxPlayers)
    
    // 通知其他模块配置已变更
    notifyGameLogic(newMaxPlayers)
})
```

## 配置格式支持

支持的配置格式（在 `config_center.yaml` 的 `format` 字段指定）：

- `yaml` - YAML 格式（推荐）
- `json` - JSON 格式
- `toml` - TOML 格式
- `properties` - Properties 格式

## 架构设计

### 传统方式 vs Config V2

#### 传统方式（原 config 模块）

```go
// 1. 定义结构体
type ServerConfig struct { ... }

// 2. 实现接口
func (s *ServerConfig) GetGroupId() string { ... }
func (s *ServerConfig) GetDataId() string { ... }
func (s *ServerConfig) Onload(cfg *Config, content string) error { ... }

// 3. 注册配置
manager.ListenConfigItem(&ServerConfig{})

// 4. 使用配置
config.GetServerName()
```

每添加一个配置项需要：定义结构体 + 实现接口 + 注册 = **大量样板代码**

#### Config V2 方式

```yaml
# 1. 在 config_center.yaml 添加配置源
sources:
  - data_id: new.config
    group: new_group
    format: yaml
```

```go
// 2. 直接使用，无需任何代码！
value := config_v2.GetString("new.config", "new_group", "some.key")
```

添加新配置项只需：**修改 yaml 文件** = **零代码**

## 最佳实践

### 1. 配置命名规范

使用点号分隔的层级结构：

```yaml
server:
  name: "game-server"
  network:
    host: "0.0.0.0"
    port: 8080
```

访问：`GetString("game.main", "server", "server.network.host")`

### 2. 环境隔离

使用不同的 namespace 隔离不同环境：

```yaml
# 开发环境
nacos:
  namespace_id: dev-namespace

# 生产环境
nacos:
  namespace_id: prod-namespace
```

### 3. 配置热更新处理

```go
config_v2.OnConfigChange("game.main", "server", func() {
    // 重新加载配置
    reloadConfiguration()
    
    // 通知相关模块
    notifyModules()
    
    // 记录日志
    logger.Info("Configuration reloaded")
})
```

### 4. 错误处理

```go
if err := config_v2.InitConfig("config_center.yaml"); err != nil {
    logger.Fatal("Failed to init config", zap.Error(err))
    // 根据错误类型做不同处理
    // 1. 文件不存在 -> 检查路径
    // 2. 连接 Nacos 失败 -> 检查网络/地址
    // 3. 配置格式错误 -> 检查 Nacos 中的配置
}
```

### 5. 默认值处理

Viper 支持设置默认值：

```go
v := config_v2.GetViper("game.main", "server")
if v != nil {
    v.SetDefault("server.port", 8080)
    v.SetDefault("server.host", "0.0.0.0")
}

// 或者在代码中处理
port := config_v2.GetInt("game.main", "server", "server.port")
if port == 0 {
    port = 8080  // 默认值
}
```

## 故障排查

### 问题 1：无法连接 Nacos

**检查**：
- Nacos 服务器地址和端口是否正确
- 网络是否可达
- 认证信息是否正确

**解决**：
```bash
# 测试连接
curl http://your-nacos-server:8848/nacos/v1/console/health/liveness
```

### 问题 2：配置读取为空

**检查**：
- DataId 和 Group 是否正确
- namespace_id 是否正确
- 配置格式（format）是否正确
- Nacos 中配置是否存在

### 问题 3：配置热更新不生效

**原因**：配置监听默认 30 秒检查一次

**解决**：修改 `manager.go` 中的 `time.NewTicker(30 * time.Second)` 调整检查间隔

## 迁移指南

### 从原 config 模块迁移

#### 步骤 1：更新配置文件

```yaml
# 旧格式（TOML）
[nacos]
server_hosts = ["host1", "host2"]

# 新格式（YAML）
nacos:
  server_hosts:
    - host1
    - host2

sources:
  - data_id: game.main
    group: server
    format: yaml
```

#### 步骤 2：替换代码

```go
// 旧代码
config.GetServerName()
config.GetRedisConfig().Addr

// 新代码
config_v2.GetString("game.main", "server", "name")
config_v2.GetStringSlice("game.main", "redis", "redis.addr")
```

#### 步骤 3：删除结构体定义（可选）

不再需要 server.go, redis.go, mongodb.go 等文件。

## 性能考虑

- ✅ **内存占用**：每个 DataId 一个 Viper 实例，按需加载
- ✅ **配置缓存**：Viper 自动缓存配置值
- ✅ **并发安全**：使用 sync.RWMutex 保护并发访问
- ✅ **热更新开销**：30秒检查一次，可调整

## 技术架构

### 核心技术栈

- `github.com/spf13/viper` - 配置管理和动态访问
- `github.com/nacos-group/nacos-sdk-go/v2` - Nacos SDK（直接集成）
- `go.uber.org/zap` - 日志

### 架构说明

Config V2 采用 **Nacos SDK + Viper** 的组合架构：

1. **Nacos SDK** 负责：
   - 连接 Nacos 服务器
   - 获取配置内容
   - 监听配置变更
   - 实时推送更新

2. **Viper** 负责：
   - 解析配置内容（支持 YAML/JSON/TOML 等）
   - 提供动态访问 API
   - 配置缓存管理

3. **ConfigManager** 作为桥梁：
   - 从 Nacos SDK 获取配置字符串
   - 将配置注入到 Viper 实例
   - 管理多个配置源
   - 协调配置更新

这种架构的优势：
- ✅ 使用官方 Nacos SDK，稳定可靠
- ✅ 充分利用 Viper 的灵活性
- ✅ 与原 config 模块使用相同的 Nacos SDK
- ✅ 无需第三方桥接库

## 总结

Config V2 的核心优势：

1. **灵活性**：无需为每个配置定义结构体
2. **简洁性**：零样板代码，配置即用
3. **强大性**：基于 Viper，功能完整
4. **可扩展性**：轻松添加新配置源

适用场景：
- ✅ 需要频繁添加新配置
- ✅ 配置结构变化频繁
- ✅ 需要灵活访问配置
- ✅ 多个配置源管理

---

**文档版本**: V2.0  
**最后更新**: 2025-12-07
