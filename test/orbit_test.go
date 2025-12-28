package orbit_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"gitee.com/orbit-w/meteor/modules/database/rdb"
	"gitee.com/orbit-w/orbit/internal/game"
	"gitee.com/orbit-w/orbit/pkg/proto/pb"
	"gitee.com/orbit-w/orbit/pkg/proto/play"
	"go.mongodb.org/mongo-driver/v2/bson"

	"gitee.com/orbit-w/orbit/config"
	"gitee.com/orbit-w/orbit/core/network"
	servicezone_behavior "gitee.com/orbit-w/orbit/core/services/service_zone/behavior"
	zone_meta "gitee.com/orbit-w/orbit/core/services/service_zone/meta"
	servicezone_mgr "gitee.com/orbit-w/orbit/core/services/service_zone/mgr"
	servicezone "gitee.com/orbit-w/orbit/core/services/service_zone/zone"
	mmeobj "gitee.com/orbit-w/orbit/internal/game/mme"
	"gitee.com/orbit-w/orbit/internal/game/modules/service"
	"gitee.com/orbit-w/orbit/internal/game/routers"
	"gitee.com/orbit-w/orbit/lib/module/db/mgo_builder"
	"gitee.com/orbit-w/orbit/lib/module/db/mongo"
	"gitee.com/orbit-w/orbit/lib/module/persistence"
	"gitee.com/orbit-w/orbit/pkg/proto/core"
	"gitee.com/orbit-w/orbit/pkg/proto/mme"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"
)

// 初始化服务
func Setup() *service.Services {
	// 初始化路由器
	servicezone_behavior.SetRouter(routers.GetRouter())

	if err := config.InitConfig("../configs/config_center.yaml"); err != nil {
		panic(err)
	}
	services := service.NewServices()
	redisService := service.Wrapper("redis_service").WrapStart(func() error {
		rdb.Start(config.GetRedisOps())
		return nil
	}).WrapStop(func() error {
		rdb.Stop()
		return nil
	})

	mongoService := service.Wrapper("mongo_service").WrapStart(func() error {
		return mongo.Start(config.GetMongoOps())
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
	config.InitConfig("../configs/config_center.yaml")

	game.Serve(1)
}

func Test_RedisDial(t *testing.T) {
	config.InitConfig("../configs/config_center.yaml")
	rdb.Start(config.GetRedisOps())
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
		_ = msg.(*play.Request_LoginRequest)
		playerEntity, ok := entities[0].(*mmeobj.PlayerEntityWrapper)
		if !ok {
			return nil, "", fmt.Errorf("player entity not found in entities")
		}

		var (
			heroId int64 = 100001
		)
		mgr := playerEntity.GetHeroManager()
		mgr.HeroMap_Range(func(id int64, heroModule *mmeobj.HeroModuleWrapper) bool {
			if id == heroId {
				heroModule.GetLevelUp().SetCurLevel(888)
				return false
			}
			return true
		})
		return &core.OK{}, "OK", nil
	})
}

// 测试设置PlayerEntity并持久化
func Test_SetPlayerEntityAndPersist(t *testing.T) {
	serverId := int32(1)
	services := Setup()
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

// 测试路由请求处理逻辑
func Test_RequestLogin(t *testing.T) {
	serverId := int32(1)
	initRouter()
	services := Setup()
	defer services.Stop()

	// 启动PlayerZone
	id := servicezone_mgr.GenLocalZoneId(servicezone.ZoneTypePlayer, serverId)
	meta, err := zone_meta.SetZoneMeta(id, int32(servicezone.ZoneTypePlayer), &zone_meta.ZoneDispatcher{
		Type:     zone_meta.Zone_DispatcherType_ForDesignated,
		ServerId: serverId,
		NodeId:   fmt.Sprintf("%d", serverId), // 当前节点ID是服务器ID
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

	req := &play.Request_LoginRequest{
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

func Test_ProtoToWrapper(t *testing.T) {
	serverId := int32(1)
	services := Setup()
	defer services.Stop()

	// 启动PlayerZone
	id := servicezone_mgr.GenLocalZoneId(servicezone.ZoneTypePlayer, serverId)
	meta, err := zone_meta.SetZoneMeta(id, int32(servicezone.ZoneTypePlayer), &zone_meta.ZoneDispatcher{
		Type:     zone_meta.Zone_DispatcherType_ForDesignated,
		ServerId: serverId,
		NodeId:   fmt.Sprintf("%d", serverId), // 当前节点ID是服务器ID
	})
	if err != nil {
		panic(err)
	}

	_, err = servicezone_mgr.StartZoneWithMeta(id, meta)
	if err != nil {
		panic(err)
	}

	playerId := int64(1600007)

	playerEntity := mmeobj.NewPlayerEntity()
	playerEntity.XXXId = playerId
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
	acc := heroWrapper.GetSkinWear().GetWearMapAccessor()
	acc.Set(100001, 100001)
	acc.Set(100002, 100002)
	acc.Set(100003, 100003)
	acc.Set(100004, 100004)
	acc.Set(100005, 100005)
	acc.Set(100006, 100006)
	acc.Set(100007, 100007)
	acc.Set(100008, 100008)
	acc.Set(100009, 100009)
	pbMsg := wrapper.ToProto()
	playerData, err := proto.Marshal(pbMsg)
	if err != nil {
		panic(err)
	}
	req := &core.Request_SetEntityRequest{
		EntityRef: &mme.EntityRef{
			EntityId:   proto.Int64(playerId),
			EntityType: mme.EntityType_PlayerEntityType.Enum(),
		},
		Data: playerData,
	}
	data, err := proto.Marshal(req)
	if err != nil {
		panic(err)
	}

	zoneId := servicezone_mgr.GenLocalZoneId(servicezone.ZoneTypePlayer, serverId)
	servicezone_mgr.ClientRequest(zoneId, network.NewClientRequest(1, pb.PID_Request_SetEntityRequest, data, nil))
	if err != nil {
		panic(err)
	}
	time.Sleep(5 * time.Minute)
}
