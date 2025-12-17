package config

import (
	mongodbdriver "gitee.com/orbit-w/meteor/modules/database/no_sql/mongodb_driver"
	"gitee.com/orbit-w/meteor/modules/database/rdb"
)

// GetServerName 获取服务器名称
func GetServerName() string {
	return GetString(DtaIDGameMain, GameMainGroupServer, "name")
}

// GetServerStage 获取服务器环境
func GetServerStage() string {
	return GetString(DtaIDGameMain, GameMainGroupServer, "stage")
}

// GetServerPort 获取服务器端口
func GetServerPort() string {
	return GetString(DtaIDGameMain, GameMainGroupServer, "port")
}

// GetRedisOps 获取 Redis 配置
func GetRedisOps() rdb.RedisClientOps {
	return rdb.RedisClientOps{
		Addr:           GetStringSlice(DtaIDGameMain, GameMainGroupRedis, "addr"),
		Cluster:        GetBool(DtaIDGameMain, GameMainGroupRedis, "cluster"),
		Username:       GetString(DtaIDGameMain, GameMainGroupRedis, "username"),
		Password:       GetString(DtaIDGameMain, GameMainGroupRedis, "password"),
		DB:             int(GetInt(DtaIDGameMain, GameMainGroupRedis, "db")),
		MaxIdleConns:   int(GetInt(DtaIDGameMain, GameMainGroupRedis, "max_idle_conns")),
		MaxActiveConns: int(GetInt(DtaIDGameMain, GameMainGroupRedis, "max_active_conns")),
	}
}

// GetMongoOps 获取 MongoDB 配置
func GetMongoOps() *mongodbdriver.MongoDBConfig {
	return &mongodbdriver.MongoDBConfig{
		URI:                    GetString(DtaIDGameMain, GameMainGroupMongo, "uri"),
		ConnectTimeout:         GetTimeDuration(DtaIDGameMain, GameMainGroupMongo, "connect_timeout"),
		MaxPoolSize:            uint64(GetInt(DtaIDGameMain, GameMainGroupMongo, "max_pool_size")),
		MinPoolSize:            uint64(GetInt(DtaIDGameMain, GameMainGroupMongo, "min_pool_size")),
		MaxConnIdleTime:        GetTimeDuration(DtaIDGameMain, GameMainGroupMongo, "max_conn_idle_time"),
		MaxConnecting:          uint64(GetInt(DtaIDGameMain, GameMainGroupMongo, "max_connecting")),
		WriteTimeout:           GetTimeDuration(DtaIDGameMain, GameMainGroupMongo, "write_timeout"),
		ReadTimeout:            GetTimeDuration(DtaIDGameMain, GameMainGroupMongo, "read_timeout"),
		RetryWrites:            GetBool(DtaIDGameMain, GameMainGroupMongo, "retry_writes"),
		RetryReads:             GetBool(DtaIDGameMain, GameMainGroupMongo, "retry_reads"),
		PingTimeout:            GetTimeDuration(DtaIDGameMain, GameMainGroupMongo, "ping_timeout"),
		DisconnectTimeout:      GetTimeDuration(DtaIDGameMain, GameMainGroupMongo, "disconnect_timeout"),
		ServerSelectionTimeout: GetTimeDuration(DtaIDGameMain, GameMainGroupMongo, "server_selection_timeout"),
	}
}

// Protocol 获取客户端跟服务器的通信协议
func GateProtocol() string {
	return GetString(DtaIDGateMain, GateGroupServer, TagProtocol)
}
