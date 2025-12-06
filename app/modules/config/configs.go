package config

import "fmt"

// 静态方法，用于获取配置
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
