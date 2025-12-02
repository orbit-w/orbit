package config

type NacosConfig struct {
	ServerHosts  []string `toml:"server_hosts"`  // Nacos 服务器地址列表，格式: "127.0.0.1:8848"
	NamespaceID  string   `toml:"namespace_id"`  // 命名空间ID
	Username     string   `toml:"username"`      // 用户名（可选）
	Password     string   `toml:"password"`      // 密码（可选）
	TimeoutMs    uint64   `toml:"timeout_ms"`    // 超时时间（毫秒），默认: 5000
	BeatInterval int64    `toml:"beat_interval"` // 心跳间隔（秒），默认: 5
}
