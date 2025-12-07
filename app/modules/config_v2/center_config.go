package config_v2

// NacosConfig Nacos配置中心配置
type NacosConfig struct {
	ServerHosts []string `toml:"server_hosts"` // Nacos 服务器地址列表，格式: "127.0.0.1:8848" 或 "127.0.0.1"
	NamespaceID string   `toml:"namespace_id"` // 命名空间ID
	Username    string   `toml:"username"`     // 用户名（可选）
	Password    string   `toml:"password"`     // 密码（可选）
	TimeoutMs   uint64   `toml:"timeout_ms"`   // 超时时间（毫秒），默认: 5000
}

// ConfigSource 配置源定义，支持从 Nacos 加载多个配置
type ConfigSource struct {
	DataID string `toml:"data_id"` // Nacos DataID
	Group  string `toml:"group"`   // Nacos Group
	Format string `toml:"format"`  // 配置格式: yaml, json, toml, properties
}

// CenterConfig 配置中心配置
type CenterConfig struct {
	Type    string          `toml:"type"`    // 配置中心类型，支持: "nacos"
	Nacos   *NacosConfig    `toml:"nacos"`   // Nacos 配置
	Sources []*ConfigSource `toml:"sources"` // 配置源列表
}
