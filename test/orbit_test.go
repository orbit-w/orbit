package orbit_test

import (
	"context"
	"errors"
	"testing"

	"gitee.com/orbit-w/meteor/modules/database/rdb"
	"gitee.com/orbit-w/orbit/app/modules/config"
	"github.com/redis/go-redis/v9"
)

func Setup(nodeId string) {
}

func Test_orbit(t *testing.T) {
	Setup("game_nd00")
}

func Test_RedisDial(t *testing.T) {
	config.InitConfig("../configs/config.toml")
	conf := config.GetConfig()
	rdb.Start(conf.GetRedisConfig().GetRedisClientOps())
	cli := rdb.UniversalClient()
	result, err := cli.Get(context.TODO(), "test").Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		t.Fatalf("get test failed: %v", err)
	}
	t.Logf("get test result: %v", result)
}
