package config_v2

import (
	"time"

	mongodbdriver "gitee.com/orbit-w/meteor/modules/database/no_sql/mongodb_driver"
	"gitee.com/orbit-w/meteor/modules/database/rdb"
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

func GetTimeDuration(dataId, group, key string) time.Duration {
	if manager == nil {
		return 0
	}
	return manager.GetTimeDuration(dataId, group, key)
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

// GetServerName 获取服务器名称
func GetServerName() string {
	return GetString(DtaIDGameMain, GroupServer, "name")
}

// GetServerStage 获取服务器阶段
func GetServerStage() string {
	return GetString(DtaIDGameMain, GroupServer, "stage")
}

// GetServerPort 获取服务器端口
func GetServerPort() string {
	return GetString(DtaIDGameMain, GroupServer, "port")
}

// GetRedisOps 获取 Redis 配置
func GetRedisOps() rdb.RedisClientOps {
	return rdb.RedisClientOps{
		Addr:           GetStringSlice(DtaIDGameMain, GroupRedis, "addr"),
		Cluster:        GetBool(DtaIDGameMain, GroupRedis, "cluster"),
		Username:       GetString(DtaIDGameMain, GroupRedis, "username"),
		Password:       GetString(DtaIDGameMain, GroupRedis, "password"),
		DB:             int(GetInt(DtaIDGameMain, GroupRedis, "db")),
		MaxIdleConns:   int(GetInt(DtaIDGameMain, GroupRedis, "max_idle_conns")),
		MaxActiveConns: int(GetInt(DtaIDGameMain, GroupRedis, "max_active_conns")),
	}
}

// GetMongoOps 获取 MongoDB 配置
func GetMongoOps() *mongodbdriver.MongoDBConfig {
	return &mongodbdriver.MongoDBConfig{
		URI:             GetString(DtaIDGameMain, GroupMongo, "uri"),
		ConnectTimeout:  GetTimeDuration(DtaIDGameMain, GroupMongo, "connect_timeout"),
		MaxPoolSize:     uint64(GetInt(DtaIDGameMain, GroupMongo, "max_pool_size")),
		MinPoolSize:     uint64(GetInt(DtaIDGameMain, GroupMongo, "min_pool_size")),
		MaxConnIdleTime: GetTimeDuration(DtaIDGameMain, GroupMongo, "max_conn_idle_time"),
		MaxConnecting:   uint64(GetInt(DtaIDGameMain, GroupMongo, "max_connecting")),
		WriteTimeout:    GetTimeDuration(DtaIDGameMain, GroupMongo, "write_timeout"),
		ReadTimeout:     GetTimeDuration(DtaIDGameMain, GroupMongo, "read_timeout"),
		RetryWrites:     GetBool(DtaIDGameMain, GroupMongo, "retry_writes"),
		RetryReads:      GetBool(DtaIDGameMain, GroupMongo, "retry_reads"),
	}
}

// GetNacosConfig 获取 Nacos 配置
func GetNacosConfig() *NacosConfig {
	if manager == nil {
		return nil
	}
	return manager.GetNacosConfig()
}
