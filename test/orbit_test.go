package orbit_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"gitee.com/orbit-w/meteor/modules/database/rdb"
	"gitee.com/orbit-w/orbit/app"

	mmeobj "gitee.com/orbit-w/orbit/app/mme"
	"gitee.com/orbit-w/orbit/app/modules/config_v2"
	"gitee.com/orbit-w/orbit/app/modules/service"
	"gitee.com/orbit-w/orbit/app/proto/core"
	"gitee.com/orbit-w/orbit/app/proto/pb"
	"gitee.com/orbit-w/orbit/app/routers"
	servicezone_behavior "gitee.com/orbit-w/orbit/core/services/service_zone/behavior"
	zone_meta "gitee.com/orbit-w/orbit/core/services/service_zone/meta"
	servicezone_mgr "gitee.com/orbit-w/orbit/core/services/service_zone/mgr"
	servicezone "gitee.com/orbit-w/orbit/core/services/service_zone/zone"
	"gitee.com/orbit-w/orbit/lib/module/db/mongo"
	"gitee.com/orbit-w/orbit/lib/module/persistence"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"
)

// 初始化服务
func Setup(nodeId string) *service.Services {
	config_v2.InitConfig("../configs/config_center.yaml")
	services := service.NewServices()
	redisService := service.Wrapper("redis_service").WrapStart(func() error {
		rdb.Start(config_v2.GetRedisOps())
		return nil
	}).WrapStop(func() error {
		rdb.Stop()
		return nil
	})

	mongoService := service.Wrapper("mongo_service").WrapStart(func() error {
		mongo.Start(config_v2.GetMongoOps())
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
	config_v2.InitConfig("../configs/config_center.yaml")

	app.Serve("1")
}

func Test_RedisDial(t *testing.T) {
	config_v2.InitConfig("../configs/config_center.yaml")
	rdb.Start(config_v2.GetRedisOps())
	cli := rdb.UniversalClient()
	result, err := cli.Get(context.TODO(), "test").Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		t.Fatalf("get test failed: %v", err)
	}
	t.Logf("get test result: %v", result)
}

func initRouter() {
	routers.RegisterHandler(pb.PID_Request_LoginRequest, func(ctx servicezone_behavior.IContext, msg proto.Message, entities ...mmeobj.IEntity) (proto.Message, string, error) {
		_ = msg.(*core.Request_LoginRequest)
		playerEntity, ok := entities[0].(*mmeobj.PlayerEntityWrapper)
		if !ok {
			return nil, "", fmt.Errorf("player entity not found in entities")
		}
		playerEntity.SetXXXId(1600000)

		var (
			heroId int64 = 100001
		)
		mgr := playerEntity.GetHeroManager()
		heroModule := mmeobj.NewHeroModule()
		wrapper := mgr.HeroMap_Set(heroId, heroModule)
		wrapper.GetLevelUp().SetCurLevel(10)

		return &core.OK{}, "OK", nil
	})
}

func Test_Persistence(t *testing.T) {
	serverId := "1"
	services := Setup(serverId)
	defer services.Stop()

	// 启动PlayerZone
	id := servicezone_mgr.GenLocalZoneId(servicezone.ZoneTypePlayer, serverId)
	meta, err := zone_meta.SetZoneMeta(id, int32(servicezone.ZoneTypePlayer), &zone_meta.ZoneDispatcher{
		Type:     zone_meta.Zone_DispatcherType_ForWorld,
		ServerId: serverId,
		NodeId:   serverId,
	})
	if err != nil {
		panic(err)
	}

	_, err = servicezone_mgr.StartZoneWithMeta(id, meta)
	if err != nil {
		panic(err)
	}
}
