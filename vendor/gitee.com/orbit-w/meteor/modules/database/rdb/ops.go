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
}
