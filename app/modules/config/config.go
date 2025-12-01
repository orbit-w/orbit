package config

import (
	"os"

	"github.com/BurntSushi/toml"
	"github.com/spf13/viper"
)

var (
	cfg Config
)

type Config struct {
	Server  Server  `toml:"server"`
	Cluster Cluster `toml:"cluster"`
	Redis   Redis   `toml:"redis"`
}

type Cluster struct {
	Nacos NacosConfig `toml:"nacos"`
}

func (c *Config) GetNacosConfig() NacosConfig {
	return c.Cluster.Nacos
}

// NacosConfig Nacos 配置
type NacosConfig struct {
	ServerHosts  []string `toml:"server_hosts"`  // Nacos 服务器地址列表，格式: "127.0.0.1:8848"
	NamespaceID  string   `toml:"namespace_id"`  // 命名空间ID
	GroupName    string   `toml:"group_name"`    // 服务组名，默认: "DEFAULT_GROUP"
	Username     string   `toml:"username"`      // 用户名（可选）
	Password     string   `toml:"password"`      // 密码（可选）
	TimeoutMs    uint64   `toml:"timeout_ms"`    // 超时时间（毫秒），默认: 5000
	BeatInterval int64    `toml:"beat_interval"` // 心跳间隔（秒），默认: 5
}

type Server struct {
	Name  string `toml:"name"`  // 服务名称
	Stage string `toml:"stage"` // 环境
	Host  string `toml:"host"`  // 主机地址
	Port  string `toml:"port"`  // 端口
}

func GetConfig() *Config {
	return &cfg
}

func LoadConfig(filename string) {
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
}
