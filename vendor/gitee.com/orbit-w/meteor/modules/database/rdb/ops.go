package rdb

import "time"

type RedisClientOps struct {
	Addr           []string
	Cluster        bool
	Username       string
	Password       string
	DB             int
	MaxIdleConns   int
	MaxActiveConns int
	DialTimeout    time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	PoolTimeout    time.Duration
}

func ParseOps(ops *RedisClientOps) {
	if ops.DialTimeout == 0 {
		ops.DialTimeout = DefaultDialTimeout
	}
	if ops.ReadTimeout == 0 {
		ops.ReadTimeout = DefaultReadTimeout
	}
	if ops.WriteTimeout == 0 {
		ops.WriteTimeout = DefaultWriteTimeout
	}
	if ops.PoolTimeout == 0 {
		ops.PoolTimeout = DefaultPoolTimeout
	}
}
