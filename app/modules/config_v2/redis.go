package config

import (
	"fmt"

	"gitee.com/orbit-w/meteor/modules/database/rdb"
	"gitee.com/orbit-w/orbit/lib/module/logger"
	"github.com/BurntSushi/toml"
	"go.uber.org/zap"
)

type GameMainRedis struct {
	Addr           []string `ymal:"addr"`
	Username       string   `ymal:"username"`
	Password       string   `ymal:"password"`
	DB             int      `ymal:"db"`
	Cluster        bool     `ymal:"cluster"`
	MaxIdleConns   int      `ymal:"max_idle_conns"`
	MaxActiveConns int      `ymal:"max_active_conns"`
}

func (r *GameMainRedis) GetGroupId() string {
	return "redis"
}

func (r *GameMainRedis) GetDataId() string {
	return "game-main.yaml"
}

func (r *GameMainRedis) Onload(cfg *Config, content string) error {
	redisConfig := new(GameMainRedis)
	if err := toml.Unmarshal([]byte(content), redisConfig); err != nil {
		logger.GetLogger().Error("failed to unmarshal game main redis config", zap.Error(err), zap.String("group", r.GetGroupId()), zap.String("dataId", r.GetDataId()))
		return fmt.Errorf("failed to unmarshal game main redis config: %w", err)
	}

	cfg.GameMain.Redis = redisConfig
	return nil
}

// GetRedisClientOps 获取Redis客户端操作配置
func (r *GameMainRedis) GetRedisClientOps() rdb.RedisClientOps {
	return rdb.RedisClientOps{
		Addr:           r.Addr,
		Username:       r.Username,
		Password:       r.Password,
		DB:             r.DB,
		Cluster:        r.Cluster,
		MaxIdleConns:   r.MaxIdleConns,
		MaxActiveConns: r.MaxActiveConns,
	}
}
