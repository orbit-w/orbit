package game

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitee.com/orbit-w/meteor/modules/database/rdb"
	"gitee.com/orbit-w/orbit/lib/module/db/mongo"
	"gitee.com/orbit-w/orbit/lib/module/logger"
	"gitee.com/orbit-w/orbit/lib/module/persistence"
	netutils "gitee.com/orbit-w/orbit/lib/utils/net_utils"

	"gitee.com/orbit-w/orbit/config"
	"gitee.com/orbit-w/orbit/core/cluster"
	"gitee.com/orbit-w/orbit/core/network"
	stream "gitee.com/orbit-w/orbit/core/services/agent_stream"
	servicezone_behavior "gitee.com/orbit-w/orbit/core/services/service_zone/behavior"
	zone_meta "gitee.com/orbit-w/orbit/core/services/service_zone/meta"
	servicezone_mgr "gitee.com/orbit-w/orbit/core/services/service_zone/mgr"
	servicezone "gitee.com/orbit-w/orbit/core/services/service_zone/zone"
	"gitee.com/orbit-w/orbit/internal/game/modules/service"
	"gitee.com/orbit-w/orbit/internal/game/routers"

	_ "gitee.com/orbit-w/orbit/internal/game/controller_v2"
)

/*
   @Author: orbit-w
   @File: serve
   @2024 4月 周日 17:36
*/

func Serve(serverId int32) {
	config.SetServerId(serverId)
	servicezone_behavior.SetRouter(routers.GetRouter())
	stream.RegisterRequestHandler(requestHandler)

	// Register services
	services := RunServices()

	// 启动集群节点
	ClusterSrartNode(serverId)

	// 启动Zone
	StartZones(serverId)

	logger.GetLogger().Info("orbit service start complete")

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

func RunServices() *service.Services {
	// Init services

	services := service.NewServices()
	redisService := service.Wrapper("redis_service").WrapStart(func() error {
		rdb.Start(config.GetRedisOps())
		return nil
	}).WrapStop(func() error {
		rdb.Stop()
		return nil
	})

	mongoService := service.Wrapper("mongo_service").WrapStart(func() error {
		mongo.Start(config.GetMongoOps())
		return nil
	}).WrapStop(func() error {
		mongo.Stop()
		return nil
	})
	services.Reg(new(stream.AgentStream))                    //启动AgentStream服务
	services.Reg(redisService)                               //启动Redis服务
	services.Reg(mongoService)                               //启动MongoDB服务
	services.Reg(persistence.New())                          //启动持久化服务
	services.Reg(cluster.NewManager(config.GetServerName())) //启动集群管理服务
	services.Reg(zone_meta.NewZoneMetaService())             //启动ZoneMeta服务
	services.Reg(servicezone_mgr.NewZoneManager())           //启动ZoneManager服务
	err := services.Start()
	if err != nil {
		panic(err)
	}

	return services
}

func StartZones(serverId int32) {
	// 启动PlayerZone
	id := servicezone_mgr.GenLocalZoneId(servicezone.ZoneTypePlayer, serverId)
	currentNode := cluster.GetManager().GetCurrentNode()
	meta, err := zone_meta.SetZoneMeta(id, int32(servicezone.ZoneTypePlayer), &zone_meta.ZoneDispatcher{
		Type:     zone_meta.Zone_DispatcherType_ForDesignated,
		ServerId: serverId,
		NodeId:   currentNode.GetId(),
	})
	if err != nil {
		panic(err)
	}

	_, err = servicezone_mgr.StartZoneWithMeta(id, meta)
	if err != nil {
		panic(err)
	}
}

func ClusterSrartNode(serverId int32) {
	ip, err := netutils.GetPublicIPv4()
	if err != nil {
		panic(err)
	}
	nodeAddress := fmt.Sprintf("%s:%s", ip, config.GetServerPort())
	id := GenServerUniqueId(config.GetServerName(), serverId)
	if err := cluster.StartNode(config.GetNacosConfig(), config.GetServerStage(), id, nodeAddress); err != nil {
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
	id := servicezone_mgr.GenLocalZoneId(servicezone.ZoneTypePlayer, config.GetServerId())
	servicezone_mgr.ClientRequest(id, network.NewClientRequest(seq, pid, data, session))
	return nil
}

// 生成服务唯一Id工具函数
func GenServerUniqueId(serverName string, serverId int32) string {
	return fmt.Sprintf("%s-%d", serverName, serverId)
}
