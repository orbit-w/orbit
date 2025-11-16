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
	Server Server
	Cluster Cluster `toml:"cluster"`
}

type Cluster struct {
	Nacos NacosConfig `toml:"nacos"`
}

// NacosConfig Nacos 配置
type NacosConfig struct {
	ServerHosts []string `toml:"server_hosts"` // Nacos 服务器地址列表，格式: "127.0.0.1:8848"
	NamespaceID string   `toml:"namespace_id"`  // 命名空间ID
	GroupName   string   `toml:"group_name"`   // 服务组名，默认: "DEFAULT_GROUP"
	Username    string   `toml:"username"`     // 用户名（可选）
	Password    string   `toml:"password"`     // 密码（可选）
	LogDir      string   `toml:"log_dir"`      // 日志目录
	CacheDir    string   `toml:"cache_dir"`    // 缓存目录
	TimeoutMs   uint64   `toml:"timeout_ms"`   // 超时时间（毫秒），默认: 5000
	BeatInterval int64   `toml:"beat_interval"` // 心跳间隔（秒），默认: 5
}

type Server struct {
	Stage string `toml:"stage"`
	Host  string `toml:"host"`
	Port  string `toml:"port"`
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
