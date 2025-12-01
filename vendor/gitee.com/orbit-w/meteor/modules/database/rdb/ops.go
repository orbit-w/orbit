package rdb

type RedisClientOps struct {
	Addr           []string
	Cluster        bool
	Username       string
	Password       string
	DB             int
	MaxIdleConns   int
	MaxActiveConns int
}
