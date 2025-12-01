package config

import "gitee.com/orbit-w/meteor/modules/database/rdb"

type Redis struct {
	Addr           []string `toml:"addr"`
	Username       string   `toml:"username"`
	Password       string   `toml:"password"`
	DB             int      `toml:"db"`
	Cluster        bool     `toml:"cluster"`
	MaxIdleConns   int      `toml:"max_idle_conns"`
	MaxActiveConns int      `toml:"max_active_conns"`
}

// GetRedisClientOps 获取Redis客户端操作配置
func (r *Redis) GetRedisClientOps() rdb.RedisClientOps {
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
