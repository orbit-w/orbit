package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	stream "github.com/orbit-w/orbit/app/core/services/agent_stream"
	"github.com/orbit-w/orbit/app/modules/service"
	"github.com/orbit-w/orbit/lib/logger"
)

/*
   @Author: orbit-w
   @File: serve
   @2024 4月 周日 17:36
*/

func Serve(nodeId string) {
	//cfg := config.GetConfig()

	// Init services
	services := service.NewServices()

	// Register services
	RegServices(services)

	if err := services.Start(); err != nil {
		panic(fmt.Sprintf("services start error: %v", err))
	}

	gracefulShutdown(func(ctx context.Context) error {
		services.Stop()
		logger.GetLogger().Info("orbit service exit")
		logger.StopLogger()
		return nil
	})
}

func RegServices(services *service.Services) {
	services.Reg(new(stream.AgentStream))
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
