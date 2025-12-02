package config

import (
	"fmt"
)

var (
	manager *ConfigManager
)

// 同一命名空间下，所有Data Id 的配置的集合，用于管理所有配置
type Config struct {
	GameMain GameMainConfig `toml:"game_main"`
}

type GameMainConfig struct {
	Server Server         `toml:"server"`
	Redis  *GameMainRedis `toml:"redis"`
}

func GetConfig() *Config {
	return manager.cfg
}

func InitConfig(filename string) {
	if manager == nil {
		manager = NewConfigManager()
	}
	manager.Start(filename)
}

func StopConfig() error {
	if manager != nil {
		return manager.Stop()
	}
	return nil
}

func GenDataId(serviceName, stage, packageName string) string {
	return fmt.Sprintf("%s-%s-%s.yaml", serviceName, stage, packageName)
}
