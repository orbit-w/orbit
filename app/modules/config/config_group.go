package config

import mongodbdriver "gitee.com/orbit-w/meteor/modules/database/no_sql/mongodb_driver"

type MainConfigGroup struct {
	Server *Server
	Redis  *GameMainRedis
	Mongo  *mongodbdriver.MongoDBConfig
}

func NewMainConfigGroup() *MainConfigGroup {
	return &MainConfigGroup{
		Server: &Server{},
		Redis:  &GameMainRedis{},
		Mongo:  &mongodbdriver.MongoDBConfig{},
	}
}

func (c *MainConfigGroup) GetRedisConfig() *GameMainRedis {
	return c.Redis
}

func (c *MainConfigGroup) GetServerConfig() *Server {
	return c.Server
}

func (c *MainConfigGroup) GetMongoConfig() *mongodbdriver.MongoDBConfig {
	return c.Mongo
}

func (c *MainConfigGroup) SetServerConfig(server *Server) {
	c.Server = server
}

func (c *MainConfigGroup) SetMongoConfig(mongo *mongodbdriver.MongoDBConfig) {
	c.Mongo = mongo
}

func (c *MainConfigGroup) SetRedisConfig(redis *GameMainRedis) {
	c.Redis = redis
}
