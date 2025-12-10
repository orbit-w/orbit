package orbit_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"gitee.com/orbit-w/meteor/modules/database/rdb"
	"gitee.com/orbit-w/orbit/app"
	"go.mongodb.org/mongo-driver/v2/bson"

	mmeobj "gitee.com/orbit-w/orbit/app/mme"
	"gitee.com/orbit-w/orbit/app/modules/config_v2"
	"gitee.com/orbit-w/orbit/app/modules/service"
	"gitee.com/orbit-w/orbit/app/proto/core"
	"gitee.com/orbit-w/orbit/app/proto/mme"
	"gitee.com/orbit-w/orbit/app/proto/pb"
	"gitee.com/orbit-w/orbit/app/routers"
	"gitee.com/orbit-w/orbit/core/network"
	servicezone_behavior "gitee.com/orbit-w/orbit/core/services/service_zone/behavior"
	zone_meta "gitee.com/orbit-w/orbit/core/services/service_zone/meta"
	servicezone_mgr "gitee.com/orbit-w/orbit/core/services/service_zone/mgr"
	servicezone "gitee.com/orbit-w/orbit/core/services/service_zone/zone"
	"gitee.com/orbit-w/orbit/lib/module/db/mgo_builder"
	"gitee.com/orbit-w/orbit/lib/module/db/mongo"
	"gitee.com/orbit-w/orbit/lib/module/persistence"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"
)

// 初始化服务
func Setup(nodeId string) *service.Services {
	// 初始化路由器
	servicezone_behavior.SetRouter(routers.GetRouter())

	if err := config_v2.InitConfig("../configs/config_center.yaml"); err != nil {
		panic(err)
	}
	services := service.NewServices()
	redisService := service.Wrapper("redis_service").WrapStart(func() error {
		rdb.Start(config_v2.GetRedisOps())
		return nil
	}).WrapStop(func() error {
		rdb.Stop()
		return nil
	})

	mongoService := service.Wrapper("mongo_service").WrapStart(func() error {
		return mongo.Start(config_v2.GetMongoOps())
	}).WrapStop(func() error {
		mongo.Stop()
		return nil
	})

	services.Reg(redisService)
	services.Reg(mongoService)
	services.Reg(persistence.New())
	services.Reg(zone_meta.NewZoneMetaService())   //启动ZoneMeta服务
	services.Reg(servicezone_mgr.NewZoneManager()) //启动ZoneManager服务

	if err := services.Start(); err != nil {
		tr := err.Error()
		fmt.Println(tr)
		panic(err)
	}
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
	servicezone_behavior.SetRouter(routers.GetRouter())
	routers.RegisterFackRouter(pb.PID_Request_LoginRequest, func(ctx servicezone_behavior.IContext, msg proto.Message, entities ...mmeobj.IEntity) (proto.Message, string, error) {
		_ = msg.(*core.Request_LoginRequest)
		playerEntity, ok := entities[0].(*mmeobj.PlayerEntityWrapper)
		if !ok {
			return nil, "", fmt.Errorf("player entity not found in entities")
		}

		var (
			heroId int64 = 100001
		)
		mgr := playerEntity.GetHeroManager()
		heroModule := mmeobj.NewHeroModule()
		wrapper := mgr.HeroMap_Set(heroId, heroModule)
		wrapper.GetLevelUp().SetCurLevel(888)

		return &core.OK{}, "OK", nil
	})
}

// 测试设置PlayerEntity并持久化
func Test_SetPlayerEntityAndPersist(t *testing.T) {
	serverId := "1"
	services := Setup(serverId)
	defer services.Stop()

	playerEntity := mmeobj.NewPlayerEntity()
	playerEntity.XXXId = 1600000
	wrapper := mmeobj.NewPlayerEntityWrapper()
	raw, err := bson.Marshal(playerEntity)
	if err != nil {
		panic(err)
	}
	wrapper.Load(raw)
	heroWrapper := wrapper.GetHeroManager().HeroMap_Set(100001, mmeobj.NewHeroModule())
	heroWrapper.GetBase().SetId(100001)
	heroWrapper.GetBase().SetConfId(100001)
	heroWrapper.GetBase().SetCreateTime(time.Now().Unix())
	heroWrapper.GetBase().SetUseTimes(10)
	heroWrapper.GetLevelUp().SetCurExp(100001)
	heroWrapper.GetLevelUp().SetConfId(100001)
	heroWrapper.GetLevelUp().SetCurLevel(10)

	heroWrapper2 := wrapper.GetHeroManager().HeroMap_Set(100002, mmeobj.NewHeroModule())
	heroWrapper2.GetBase().SetId(100002)
	heroWrapper2.GetBase().SetConfId(100002)
	heroWrapper2.GetBase().SetCreateTime(time.Now().Unix())
	heroWrapper2.GetBase().SetUseTimes(10)
	heroWrapper2.GetLevelUp().SetCurExp(100001)
	heroWrapper2.GetLevelUp().SetConfId(100001)
	heroWrapper2.GetLevelUp().SetCurLevel(10)
	builder := mgo_builder.NewMongoUpdateBuilder()
	wrapper.BuildMongoUpdate(builder)
	update := builder.Build()
	zoneId := servicezone_mgr.GenLocalZoneId(servicezone.ZoneTypePlayer, serverId)
	database := servicezone.ZoneIdToDatabase(zoneId)
	resp, err := persistence.PersistSync(context.TODO(), database, wrapper.Collection(), wrapper.GetXXXId(), update)
	if err != nil {
		panic(err)
	}
	fmt.Println(resp.Success)
	fmt.Println(resp.MatchedCount)
	fmt.Println(resp.ModifiedCount)
	fmt.Println(resp.Collection)
	fmt.Println(resp.DocumentID)
	fmt.Println(resp.Error)
}

func Test_RequestLogin(t *testing.T) {
	serverId := "1"
	initRouter()
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

	ref := &mme.EntityRef{}
	ref.EntityId = proto.Int64(1600000)
	ref.EntityType = mme.EntityType_PlayerEntityType.Enum()

	req := &core.Request_LoginRequest{
		PlayerEntityRef: ref,
	}
	data, err := proto.Marshal(req)
	if err != nil {
		panic(err)
	}
	err = servicezone_mgr.ClientRequest(id, network.NewClientRequest(1, pb.PID_Request_LoginRequest, data, nil))
	if err != nil {
		panic(err)
	}

	time.Sleep(5 * time.Minute)
}
