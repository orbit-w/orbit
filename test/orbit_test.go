package orbit_test

import (
	"context"
	"errors"
	"testing"

	"gitee.com/orbit-w/meteor/modules/database/rdb"
	"gitee.com/orbit-w/orbit/app"
	"gitee.com/orbit-w/orbit/app/modules/config"
	"gitee.com/orbit-w/orbit/app/modules/service"
	zone_meta "gitee.com/orbit-w/orbit/core/services/service_zone/meta"
	servicezone_mgr "gitee.com/orbit-w/orbit/core/services/service_zone/mgr"
	"gitee.com/orbit-w/orbit/lib/module/db/mongo"
	"gitee.com/orbit-w/orbit/lib/module/persistence"
	"github.com/redis/go-redis/v9"
)

// 初始化服务
func Setup(nodeId string) *service.Services {
	config.InitConfig("../configs/config.toml")
	cfg := config.GetConfig()
	services := service.NewServices()
	redisService := service.Wrapper("redis_service").WrapStart(func() error {
		rdb.Start(cfg.GetRedisConfig().GetRedisClientOps())
		return nil
	}).WrapStop(func() error {
		rdb.Stop()
		return nil
	})

	mongoService := service.Wrapper("mongo_service").WrapStart(func() error {
		mongo.Start(cfg.GetGameMainConfig().GetMongoConfig())
		return nil
	}).WrapStop(func() error {
		mongo.Stop()
		return nil
	})

	services.Reg(redisService)
	services.Reg(mongoService)
	services.Reg(persistence.New())
	services.Reg(zone_meta.NewZoneMetaService())   //启动ZoneMeta服务
	services.Reg(servicezone_mgr.NewZoneManager()) //启动ZoneManager服务

	services.Start()
	return services
}

func Test_orbit(t *testing.T) {
	config.InitConfig("../configs/config_center.toml")

	app.Serve("1")
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

func Test_Persistence(t *testing.T) {
	services := Setup("1")
	defer services.Stop()

}
