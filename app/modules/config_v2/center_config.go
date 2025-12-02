package config

import (
	"os"

	"github.com/BurntSushi/toml"
	"github.com/spf13/viper"
)

var (
	centerConfig CenterConfig
)

type CenterConfig struct {
	Type  string             `toml:"type"` // 配置中心类型，支持: "nacos"
	Nacos *CenterNacosConfig `toml:"nacos"`
}

type CenterNacosConfig struct {
	ServerHosts  []string `toml:"server_hosts"`  // Nacos 服务器地址列表，格式: "127.0.0.1:8848"
	NamespaceID  string   `toml:"namespace_id"`  // 命名空间ID
	Username     string   `toml:"username"`      // 用户名（可选）
	Password     string   `toml:"password"`      // 密码（可选）
	TimeoutMs    uint64   `toml:"timeout_ms"`    // 超时时间（毫秒），默认: 5000
	BeatInterval int64    `toml:"beat_interval"` // 心跳间隔（秒），默认: 5
}

func (m *ConfigManager) LoadConfig(filename string) {
	cfg := new(CenterConfig)
	viper.SetConfigFile(filename)
	viper.SetConfigType("toml")

	// 尝试读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		panic("viper read config failed")
	}

	// 读取配置文件
	content, err := os.ReadFile(filename)
	if err != nil {
		panic("read config failed")
	}

	if err := toml.Unmarshal(content, &cfg); err != nil {
		panic("unmarshal config failed")
	}

	m.centerConfig = *cfg
}
