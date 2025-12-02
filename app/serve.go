package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitee.com/orbit-w/meteor/modules/database/rdb"
	"gitee.com/orbit-w/orbit/lib/module/logger"
	"gitee.com/orbit-w/orbit/lib/module/persistence"
	netutils "gitee.com/orbit-w/orbit/lib/utils/net_utils"

	"gitee.com/orbit-w/orbit/app/modules/config"
	"gitee.com/orbit-w/orbit/app/modules/service"
	"gitee.com/orbit-w/orbit/app/routers"
	"gitee.com/orbit-w/orbit/core/cluster"
	"gitee.com/orbit-w/orbit/core/network"
	stream "gitee.com/orbit-w/orbit/core/services/agent_stream"
	servicezone_behavior "gitee.com/orbit-w/orbit/core/services/service_zone/behavior"
	zone_meta "gitee.com/orbit-w/orbit/core/services/service_zone/meta"
	servicezone_mgr "gitee.com/orbit-w/orbit/core/services/service_zone/mgr"
	servicezone "gitee.com/orbit-w/orbit/core/services/service_zone/zone"

	_ "gitee.com/orbit-w/orbit/app/controller_v2"
)

/*
   @Author: orbit-w
   @File: serve
   @2024 4月 周日 17:36
*/

func Serve(nodeId string) {
	cfg := config.GetConfig()
	// 初始化 routers
	routers.Init()

	servicezone_behavior.SetRouter(routers.GetRouter())
	stream.RegisterRequestHandler(requestHandler)

	// Register services
	services := RunServices(cfg)

	// 启动集群节点
	ClusterSrartNode(nodeId)

	// 启动Zone
	StartZones(nodeId)

	gracefulShutdown(func(ctx context.Context) error {
		// 停止服务
		services.Stop()

		// 停止配置管理器
		config.StopConfig()
		logger.GetLogger().Info("orbit service exit")
		logger.StopLogger()
		return nil
	})
}

func RunServices(cfg *config.Config) *service.Services {
	// Init services
	services := service.NewServices()
	redisService := service.Wrapper("redis_service").WrapStart(func() error {
		rdb.Start(cfg.GetRedisConfig().GetRedisClientOps())
		return nil
	}).WrapStop(func() error {
		rdb.Stop()
		return nil
	})

	services.Reg(persistence.New("configs/mongodb.toml"))             //启动持久化服务
	services.Reg(new(stream.AgentStream))                             //启动AgentStream服务
	services.Reg(redisService)                                        //启动Redis服务
	services.Reg(cluster.NewManager(cfg.GetServerName()))             //启动集群管理服务
	services.Reg(servicezone_mgr.NewZoneManager())                    //启动ZoneManager服务
	services.Reg(zone_meta.NewZoneMetaService(rdb.UniversalClient())) //启动ZoneMeta服务

	return services
}

func StartZones(nodeId string) {
	// 启动PlayerZone
	id := servicezone_mgr.GenLocalZoneId(servicezone.ZoneTypePlayer, nodeId)
	meta, err := zone_meta.SetZoneMeta(id, int32(servicezone.ZoneTypePlayer), &zone_meta.ZoneDispatcher{
		Type:     zone_meta.Zone_DispatcherType_ForWorld,
		ServerId: nodeId,
		NodeId:   nodeId,
	})
	if err != nil {
		panic(err)
	}

	_, err = servicezone_mgr.StartZoneWithMeta(id, meta)
	if err != nil {
		panic(err)
	}
}

func ClusterSrartNode(nodeId string) {
	nacosCfg := config.GetNacosConfig()
	ip, err := netutils.GetLocalIPv4()
	if err != nil {
		panic(err)
	}
	serverCfg := config.GetGameMainConfig().Server
	nodeAddress := fmt.Sprintf("%s:%s", ip, serverCfg.Port)
	if err := cluster.StartNode(nacosCfg, serverCfg.Stage, nodeId, nodeAddress); err != nil {
		panic(err)
	}
}

// gracefulShutdown 优雅关闭服务
func gracefulShutdown(stopper func(ctx context.Context) error) {
	// 等待中断信号
	quit := make(chan os.Signal, 1)
	// 监听 SIGINT（Ctrl+C）和 SIGTERM 信号
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// 创建一个5分钟超时的context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if stopper != nil {
		if err := stopper(ctx); err != nil {
			log.Printf("Error stopping stopper: %v", err)
		}
	}

	log.Println("Server exiting")
}

var requestHandler = func(session *network.Session, data []byte, seq, pid uint32) error {
	servicezone_mgr.ClientRequest("play", network.NewClientRequest(seq, pid, data, session))
	return nil
}
