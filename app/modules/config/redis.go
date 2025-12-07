package config

import (
	"fmt"

	"gitee.com/orbit-w/meteor/modules/database/rdb"
	"gitee.com/orbit-w/orbit/lib/module/logger"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type GameMainRedis struct {
	Addr           []string `yaml:"addr"`
	Username       string   `yaml:"username"`
	Password       string   `yaml:"password"`
	DB             int32    `yaml:"db"`
	Cluster        bool     `yaml:"cluster"`
	MaxIdleConns   int32    `yaml:"max_idle_conns"`
	MaxActiveConns int32    `yaml:"max_active_conns"`
}

func (r *GameMainRedis) GetGroupId() string {
	return "redis"
}

func (r *GameMainRedis) GetDataId() string {
	return "game.main"
}

func (r *GameMainRedis) Onload(cfg *Config, content string) error {
	redisConfig := new(GameMainRedis)
	if err := yaml.Unmarshal([]byte(content), redisConfig); err != nil {
		logger.GetLogger().Error("failed to unmarshal game main redis config", zap.Error(err), zap.String("group", r.GetGroupId()), zap.String("dataId", r.GetDataId()))
		return fmt.Errorf("failed to unmarshal game main redis config: %w", err)
	}

	mainConf := cfg.GetGameMainConfig()
	mainConf.SetRedisConfig(redisConfig)
	return nil
}

// GetRedisClientOps 获取Redis客户端操作配置
func (r *GameMainRedis) GetRedisClientOps() rdb.RedisClientOps {
	return rdb.RedisClientOps{
		Addr:           r.Addr,
		Username:       r.Username,
		Password:       r.Password,
		DB:             int(r.DB),
		Cluster:        r.Cluster,
		MaxIdleConns:   int(r.MaxIdleConns),
		MaxActiveConns: int(r.MaxActiveConns),
	}
}
