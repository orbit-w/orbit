package config

import mongodbdriver "gitee.com/orbit-w/meteor/modules/database/no_sql/mongodb_driver"

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
	Mongo  *mongodbdriver.MongoDBConfig
}

func (c *GameMainConfig) GetRedisConfig() *GameMainRedis {
	return c.Redis
}

func (c *GameMainConfig) GetServerConfig() *Server {
	return &c.Server
}

func (c *GameMainConfig) GetMongoConfig() *mongodbdriver.MongoDBConfig {
	return c.Mongo
}

func (c *GameMainConfig) SetMongoConfig(mongo *mongodbdriver.MongoDBConfig) {
	c.Mongo = mongo
}
