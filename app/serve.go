package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitee.com/orbit-w/orbit/lib/module/logger"
	"gitee.com/orbit-w/orbit/lib/module/persistence"

	"gitee.com/orbit-w/orbit/app/modules/service"
	"gitee.com/orbit-w/orbit/app/routers"
	"gitee.com/orbit-w/orbit/core/network"
	stream "gitee.com/orbit-w/orbit/core/services/agent_stream"
	servicezone_behavior "gitee.com/orbit-w/orbit/core/services/service_zone/behavior"
	servicezone_mgr "gitee.com/orbit-w/orbit/core/services/service_zone/mgr"

	_ "gitee.com/orbit-w/orbit/app/controller_v2"
)

/*
   @Author: orbit-w
   @File: serve
   @2024 4月 周日 17:36
*/

func Serve(nodeId string) {
	//cfg := config.GetConfig()

	// 初始化 routers
	routers.Init()

	servicezone_behavior.SetRouter(routers.GetRouter())
	stream.RegisterRequestHandler(requestHandler)

	// Register services
	services := RunServices()

	gracefulShutdown(func(ctx context.Context) error {
		services.Stop()
		logger.GetLogger().Info("orbit service exit")
		logger.StopLogger()
		return nil
	})
}

func RunServices() *service.Services {

	// Init services
	services := service.NewServices()

	services.Reg(new(stream.AgentStream))
	services.Reg(persistence.New("configs/mongodb.toml"))

	return services
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
