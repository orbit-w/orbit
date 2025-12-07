package config_v2

import (
	"github.com/spf13/viper"
)

var (
	manager *ConfigManager
)

// InitConfig 初始化配置管理器
func InitConfig(filename string) error {
	if manager == nil {
		manager = NewConfigManager()
	}
	return manager.Start(filename)
}

// StopConfig 停止配置管理器
func StopConfig() error {
	if manager != nil {
		return manager.Stop()
	}
	return nil
}

// GetViper 获取指定 DataId 和 Group 的 viper 实例
func GetViper(dataId, group string) *viper.Viper {
	if manager == nil {
		return nil
	}
	return manager.GetViper(dataId, group)
}

// Get 获取配置值
func Get(dataId, group, key string) any {
	if manager == nil {
		return nil
	}
	return manager.Get(dataId, group, key)
}

// GetString 获取字符串配置
func GetString(dataId, group, key string) string {
	if manager == nil {
		return ""
	}
	return manager.GetString(dataId, group, key)
}

// GetInt 获取整数配置
func GetInt(dataId, group, key string) int {
	if manager == nil {
		return 0
	}
	return manager.GetInt(dataId, group, key)
}

// GetBool 获取布尔配置
func GetBool(dataId, group, key string) bool {
	if manager == nil {
		return false
	}
	return manager.GetBool(dataId, group, key)
}

// GetStringSlice 获取字符串数组配置
func GetStringSlice(dataId, group, key string) []string {
	if manager == nil {
		return nil
	}
	return manager.GetStringSlice(dataId, group, key)
}

// GetStringMap 获取字符串映射配置
func GetStringMap(dataId, group, key string) map[string]any {
	if manager == nil {
		return nil
	}
	return manager.GetStringMap(dataId, group, key)
}

// Unmarshal 将配置反序列化到结构体
func Unmarshal(dataId, group string, rawVal any) error {
	if manager == nil {
		return nil
	}
	return manager.Unmarshal(dataId, group, rawVal)
}

// UnmarshalKey 将配置的某个 key 反序列化到结构体
func UnmarshalKey(dataId, group, key string, rawVal any) error {
	if manager == nil {
		return nil
	}
	return manager.UnmarshalKey(dataId, group, key, rawVal)
}

// OnConfigChange 注册配置变更回调
func OnConfigChange(dataId, group string, callback func()) {
	if manager != nil {
		manager.OnConfigChange(dataId, group, callback)
	}
}
