package rdb

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/redis/go-redis/v9"
)

var (
	universalClient redis.UniversalClient
	once            sync.Once
	state           atomic.Int32

	ErrAddressInvalid = errors.New("err_redis_address_invalid")
)

const (
	stateInit = iota
	stateRunning
	stateStopping
	stateStopped
)

// UniversalClient 获取原始的redis 虚拟连接实例
func UniversalClient() redis.UniversalClient {
	return universalClient
}

// 全局单例模式启动redis连接
func Start(ops RedisClientOps) {
	once.Do(func() {
		var err error
		universalClient, err = NewClient(ops)
		if err != nil {
			panic(err)
		}
		state.Store(stateRunning)
	})
}

func Stop() {
	if state.CompareAndSwap(stateRunning, stateStopping) {
		_ = universalClient.Close()
		state.Store(stateStopped)
	}
}

func NewClient(ops RedisClientOps) (redis.UniversalClient, error) {
	if len(ops.Addr) == 0 {
		return nil, ErrAddressInvalid
	}

	var client redis.UniversalClient
	switch {
	case ops.Cluster:
		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:          ops.Addr,
			Username:       ops.Username,
			Password:       ops.Password,
			MaxIdleConns:   ops.MaxIdleConns,
			MaxActiveConns: ops.MaxActiveConns,
		})
	default:
		client = redis.NewClient(&redis.Options{
			Addr:           ops.Addr[0],
			Username:       ops.Username,
			Password:       ops.Password, // no password set
			DB:             ops.DB,       // use default db
			MaxIdleConns:   ops.MaxIdleConns,
			MaxActiveConns: ops.MaxActiveConns,
		})
	}

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return client, nil
}
