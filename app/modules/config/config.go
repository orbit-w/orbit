package config

import (
	"fmt"
	"github.com/spf13/viper"
)

var (
	cfg Config
)

type Config struct {
	Server struct {
		Stage string //环境
		Host  string
		Port  string
	}
}

func GetConfig() *Config {
	return &cfg
}

func LoadConfig(name, path string) {
	viper.SetConfigName(name) // 配置文件的名字（没有扩展名）
	viper.AddConfigPath(path) // 查找配置文件的路径

	// 尝试读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Fatal error config file, path: %s, err: %s \n", path, err.Error()))
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		panic(fmt.Errorf("viper unmarshal failed, path: %s, err: %s \n", path, err.Error()))
	}
}
