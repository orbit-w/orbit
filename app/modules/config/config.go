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
