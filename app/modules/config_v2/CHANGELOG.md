# Config V2 更新日志

## [2.0.0] - 2025-12-07

### 💥 破坏性变更 (Breaking Changes)

#### 所有 API 方法现在需要 `group` 参数

为了确保配置的唯一性，所有配置访问方法现在都需要同时提供 `dataId` 和 `group` 参数。

在 Nacos 中，`dataId` 和 `group` 的组合才能唯一标识一个配置。

#### API 变更列表

**配置访问方法：**

| 旧 API | 新 API |
|--------|--------|
| `Get(dataId, key)` | `Get(dataId, group, key)` |
| `GetString(dataId, key)` | `GetString(dataId, group, key)` |
| `GetInt(dataId, key)` | `GetInt(dataId, group, key)` |
| `GetBool(dataId, key)` | `GetBool(dataId, group, key)` |
| `GetStringSlice(dataId, key)` | `GetStringSlice(dataId, group, key)` |
| `GetStringMap(dataId, key)` | `GetStringMap(dataId, group, key)` |

**配置反序列化：**

| 旧 API | 新 API |
|--------|--------|
| `Unmarshal(dataId, rawVal)` | `Unmarshal(dataId, group, rawVal)` |
| `UnmarshalKey(dataId, key, rawVal)` | `UnmarshalKey(dataId, group, key, rawVal)` |

**其他方法：**

| 旧 API | 新 API |
|--------|--------|
| `GetViper(dataId)` | `GetViper(dataId, group)` |
| `OnConfigChange(dataId, callback)` | `OnConfigChange(dataId, group, callback)` |

### ✨ 改进

- 使用 `any` 替代 `interface{}` (Go 1.18+)
- 改进了内部键生成机制，使用 `dataId.group` 格式
- 更准确的错误消息，包含 `dataId` 和 `group` 信息

### 📝 迁移指南

#### 步骤 1：更新所有配置访问调用

**旧代码：**
```go
name := config_v2.GetString("game.main", "name")
port := config_v2.GetInt("game.main", "port")
```

**新代码：**
```go
name := config_v2.GetString("game.main", "DEFAULT_GROUP", "name")
port := config_v2.GetInt("game.main", "DEFAULT_GROUP", "port")
```

#### 步骤 2：更新配置反序列化调用

**旧代码：**
```go
var config ServerConfig
err := config_v2.Unmarshal("game.main", &config)
```

**新代码：**
```go
var config ServerConfig
err := config_v2.Unmarshal("game.main", "DEFAULT_GROUP", &config)
```

#### 步骤 3：更新配置变更回调

**旧代码：**
```go
config_v2.OnConfigChange("game.main", func() {
    newName := config_v2.GetString("game.main", "name")
})
```

**新代码：**
```go
config_v2.OnConfigChange("game.main", "DEFAULT_GROUP", func() {
    newName := config_v2.GetString("game.main", "DEFAULT_GROUP", "name")
})
```

### 📌 注意事项

1. **默认 Group**: 如果在 Nacos 中没有指定 group，通常使用 `"DEFAULT_GROUP"`
2. **配置唯一性**: `dataId + group` 的组合对应唯一的 viper 实例
3. **向后兼容性**: 此版本不向后兼容 1.x 版本，需要更新所有代码

### 🔍 为什么做这个变更？

**问题：**
旧版本只使用 `dataId` 作为键，但在 Nacos 中：
- 相同的 `dataId` 可以存在于不同的 `group` 中
- 这会导致配置冲突和覆盖问题

**解决方案：**
使用 `dataId + group` 组合作为唯一标识符，确保：
- 不同 group 中的相同 dataId 不会冲突
- 配置访问更加明确和可预测
- 与 Nacos 的设计理念保持一致

---

**文档版本**: 2.0.0  
**发布日期**: 2025-12-07
