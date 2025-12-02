package config

import (
	"fmt"
)

var (
	manager *ConfigManager
)

// 同一命名空间下，所有Data Id 的配置的集合，用于管理所有配置
type Config struct {
	GameMain *GameMainConfig `toml:"game_main"`
}

func (c *Config) GetGameMainConfig() *GameMainConfig {
	return c.GameMain
}

func (c *Config) GetServerName() string {
	return c.GameMain.Server.Name
}

func (c *Config) GetServerStage() string {
	return c.GameMain.Server.Stage
}

func (c *Config) GetRedisConfig() *GameMainRedis {
	return c.GameMain.GetRedisConfig()
}

type GameMainConfig struct {
	Server Server         `toml:"server"`
	Redis  *GameMainRedis `toml:"redis"`
}

func (c *GameMainConfig) GetRedisConfig() *GameMainRedis {
	return c.Redis
}

func (c *GameMainConfig) GetServerConfig() *Server {
	return &c.Server
}

func GetConfig() *Config {
	return manager.cfg
}

// 获取Nacos集群配置
func GetNacosConfig() *NacosConfig {
	return manager.centerConfig.Nacos
}

func GetGameMainConfig() *GameMainConfig {
	return manager.cfg.GameMain
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
