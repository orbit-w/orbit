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
	Type  string       `toml:"type"` // 配置中心类型，支持: "nacos"
	Nacos *NacosConfig `toml:"nacos"`
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
