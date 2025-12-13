package config_v2

// NacosConfig Nacos配置中心配置
type NacosConfig struct {
	ServerHosts  []string `yaml:"server_hosts"`  // Nacos 服务器地址列表，格式: "127.0.0.1:8848" 或 "127.0.0.1"
	NamespaceID  string   `yaml:"namespace_id"`  // 命名空间ID
	Username     string   `yaml:"username"`      // 用户名（可选）
	Password     string   `yaml:"password"`      // 密码（可选）
	TimeoutMs    uint64   `yaml:"timeout_ms"`    // 超时时间（毫秒），默认: 5000
	BeatInterval int64    `yaml:"beat_interval"` // 心跳间隔（秒），默认: 5
}

// ConfigSource 配置源定义，支持从 Nacos 加载多个配置
type ConfigSource struct {
	DataID     string `yaml:"data_id"`     // Nacos DataID
	Group      string `yaml:"group"`       // Nacos Group
	Format     string `yaml:"format"`      // 配置格式: yaml, json, toml, properties
	NeedListen bool   `yaml:"need_listen"` // 是否需要监听配置变化
}

// CenterConfig 配置中心配置
type CenterConfig struct {
	Type    string          `yaml:"type"`    // 配置中心类型，支持: "nacos"
	Nacos   *NacosConfig    `yaml:"nacos"`   // Nacos 配置
	Sources []*ConfigSource `yaml:"sources"` // 配置源列表
}
